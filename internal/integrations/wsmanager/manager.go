package wsmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
)

const (
	pingInterval        = 30 * time.Second
	pongWait            = 60 * time.Second
	writeWait           = 10 * time.Second
	sendBufSize         = 64
	broadcastBufSize    = 1024
	numBroadcastWorkers = 8
)

// WSMessage is an incoming message from the client.
type WSMessage struct {
	Action    string          `json:"action"`
	AuctionID string          `json:"auction_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// WSEvent is an outgoing event sent to the client.
type WSEvent struct {
	Event     string      `json:"event"`
	AuctionID string      `json:"auction_id,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type broadcastEnvelope struct {
	auctionID string
	event     WSEvent
}

type placeBidPayload struct {
	Amount int64 `json:"amount"`
}

// IAuctionActions is the subset of auction business logic used by the manager.
type IAuctionActions interface {
	CreateAuction(ctx context.Context, req models.CreateAuctionRequest, userID string) (string, error)
	PlaceBid(ctx context.Context, auctionID string, amount int64, userID string) (*models.PlaceBidResponse, error)
}

type connection struct {
	id     string
	userID string
	conn   *websocket.Conn
	send   chan WSEvent
}

// WSManager tracks active WebSocket connections and routes events between them.
// Broadcasts are funnelled through a shared channel consumed by N worker goroutines.
type WSManager struct {
	mu           sync.RWMutex
	connections  map[string]*connection
	subscribers  map[string]map[string]struct{} // auctionID → connID set
	connAuctions map[string]map[string]struct{} // connID → auctionID set

	broadcastCh chan broadcastEnvelope

	actions IAuctionActions
	metrics metrics.IMetrics
	logger  applogger.IAppLogger
}

func NewWSManager(actions IAuctionActions, m metrics.IMetrics, logger applogger.IAppLogger) *WSManager {
	return &WSManager{
		connections:  make(map[string]*connection),
		subscribers:  make(map[string]map[string]struct{}),
		connAuctions: make(map[string]map[string]struct{}),
		broadcastCh:  make(chan broadcastEnvelope, broadcastBufSize),
		actions:      actions,
		metrics:      m,
		logger:       logger,
	}
}

// Start launches background broadcast workers. Blocks until ctx is cancelled.
func (m *WSManager) Start(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < numBroadcastWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.broadcastWorker(ctx)
		}()
	}
	wg.Wait()
}

func (m *WSManager) broadcastWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case env, ok := <-m.broadcastCh:
			if !ok {
				return
			}
			m.deliver(env.auctionID, env.event)
		}
	}
}

// Broadcast enqueues an event for all subscribers of auctionID.
// Non-blocking: drops the event if the channel is full.
func (m *WSManager) Broadcast(auctionID string, event WSEvent) {
	select {
	case m.broadcastCh <- broadcastEnvelope{auctionID: auctionID, event: event}:
	default:
		m.metrics.WSBroadcastDroppedInc()
		m.logger.Info(context.Background(), "broadcast channel full, event dropped",
			"auctionID", auctionID, "event", event.Event)
	}
}

func (m *WSManager) deliver(auctionID string, event WSEvent) {
	m.mu.RLock()
	subs := make([]string, 0, len(m.subscribers[auctionID]))
	for connID := range m.subscribers[auctionID] {
		subs = append(subs, connID)
	}
	m.mu.RUnlock()

	for _, connID := range subs {
		m.mu.RLock()
		conn, ok := m.connections[connID]
		m.mu.RUnlock()
		if !ok {
			continue
		}
		select {
		case conn.send <- event:
			m.metrics.WSMessagesSentInc(event.Event)
		default:
			m.metrics.WSBroadcastDroppedInc()
			m.logger.Info(context.Background(), "connection send buffer full, event dropped",
				"connID", connID, "event", event.Event)
		}
	}
}

// HandleConnection registers the connection and runs readPump + writePump goroutines.
// Blocks until the connection is closed or ctx is cancelled.
func (m *WSManager) HandleConnection(ctx context.Context, ws *websocket.Conn, userID string) {
	conn := &connection{
		id:     uuid.New().String(),
		userID: userID,
		conn:   ws,
		send:   make(chan WSEvent, sendBufSize),
	}

	m.register(conn)
	defer m.unregister(conn)
	m.metrics.WSConnectionsInc()
	defer m.metrics.WSConnectionsDec()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		m.writePump(ctx, conn)
	}()
	go func() {
		defer wg.Done()
		m.readPump(ctx, conn)
	}()

	conn.send <- WSEvent{Event: "connected"}

	wg.Wait()
}

// readPump reads incoming WS frames and dispatches actions.
// Closes conn.send when the read loop exits so writePump terminates.
func (m *WSManager) readPump(ctx context.Context, conn *connection) {
	defer close(conn.send)

	conn.conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.conn.SetPongHandler(func(string) error {
		conn.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := conn.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			select {
			case conn.send <- WSEvent{Event: "error", Error: "invalid JSON"}:
			default:
			}
			continue
		}

		m.dispatch(ctx, conn, msg)
	}
}

// writePump writes outgoing events to the WS connection and sends periodic pings.
func (m *WSManager) writePump(ctx context.Context, conn *connection) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	var closeOnce sync.Once
	closeConn := func() {
		closeOnce.Do(func() {
			conn.conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(writeWait),
			)
			conn.conn.Close()
		})
	}
	defer closeConn()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-conn.send:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			conn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
			m.metrics.WSMessagesSentInc(event.Event)

		case <-ticker.C:
			conn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}

func (m *WSManager) dispatch(ctx context.Context, conn *connection, msg WSMessage) {
	m.metrics.WSMessagesReceivedInc(msg.Action)
	switch msg.Action {
	case "subscribe":
		m.handleSubscribe(conn, msg.AuctionID)
	case "unsubscribe":
		m.handleUnsubscribe(conn, msg.AuctionID)
	case "create_auction":
		m.handleCreateAuction(ctx, conn, msg.Payload)
	case "place_bid":
		m.handlePlaceBid(ctx, conn, msg.AuctionID, msg.Payload)
	default:
		trySend(conn.send, WSEvent{Event: "error", Error: fmt.Sprintf("unknown action: %s", msg.Action)})
	}
}

func (m *WSManager) handleSubscribe(conn *connection, auctionID string) {
	if auctionID == "" {
		trySend(conn.send, WSEvent{Event: "error", Error: "auction_id is required for subscribe"})
		return
	}

	m.mu.Lock()
	if m.subscribers[auctionID] == nil {
		m.subscribers[auctionID] = make(map[string]struct{})
	}
	m.subscribers[auctionID][conn.id] = struct{}{}
	if m.connAuctions[conn.id] == nil {
		m.connAuctions[conn.id] = make(map[string]struct{})
	}
	m.connAuctions[conn.id][auctionID] = struct{}{}
	m.mu.Unlock()

	trySend(conn.send, WSEvent{Event: "subscribed", AuctionID: auctionID})
}

func (m *WSManager) handleUnsubscribe(conn *connection, auctionID string) {
	if auctionID == "" {
		trySend(conn.send, WSEvent{Event: "error", Error: "auction_id is required for unsubscribe"})
		return
	}

	m.mu.Lock()
	delete(m.subscribers[auctionID], conn.id)
	delete(m.connAuctions[conn.id], auctionID)
	m.mu.Unlock()

	trySend(conn.send, WSEvent{Event: "unsubscribed", AuctionID: auctionID})
}

func (m *WSManager) handleCreateAuction(ctx context.Context, conn *connection, raw json.RawMessage) {
	var req models.CreateAuctionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		trySend(conn.send, WSEvent{Event: "error", Error: "invalid create_auction payload"})
		return
	}

	auctionID, err := m.actions.CreateAuction(ctx, req, conn.userID)
	if err != nil {
		m.logger.Error(ctx, err, "action", "create_auction", "userID", conn.userID)
		trySend(conn.send, WSEvent{Event: "error", Error: "failed to create auction"})
		return
	}

	trySend(conn.send, WSEvent{
		Event:     "auction_created",
		AuctionID: auctionID,
		Payload:   map[string]string{"auction_id": auctionID},
	})
}

func (m *WSManager) handlePlaceBid(ctx context.Context, conn *connection, auctionID string, raw json.RawMessage) {
	if auctionID == "" {
		trySend(conn.send, WSEvent{Event: "error", Error: "auction_id is required for place_bid"})
		return
	}

	var p placeBidPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		trySend(conn.send, WSEvent{Event: "error", Error: "invalid place_bid payload"})
		return
	}

	result, err := m.actions.PlaceBid(ctx, auctionID, p.Amount, conn.userID)
	if err != nil {
		m.logger.Error(ctx, err, "action", "place_bid", "auctionID", auctionID, "userID", conn.userID)
		trySend(conn.send, WSEvent{Event: "error", Error: "failed to place bid"})
		return
	}

	m.Broadcast(auctionID, WSEvent{
		Event:     "bid_placed",
		AuctionID: auctionID,
		Payload: map[string]interface{}{
			"bidder_id": conn.userID,
			"amount":    p.Amount,
			"success":   result.Success,
		},
	})
}

// trySend attempts a non-blocking send on ch.
func trySend(ch chan WSEvent, event WSEvent) {
	select {
	case ch <- event:
	default:
	}
}

func (m *WSManager) register(conn *connection) {
	m.mu.Lock()
	m.connections[conn.id] = conn
	m.connAuctions[conn.id] = make(map[string]struct{})
	m.mu.Unlock()
}

func (m *WSManager) unregister(conn *connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for auctionID := range m.connAuctions[conn.id] {
		delete(m.subscribers[auctionID], conn.id)
	}
	delete(m.connAuctions, conn.id)
	delete(m.connections, conn.id)
}
