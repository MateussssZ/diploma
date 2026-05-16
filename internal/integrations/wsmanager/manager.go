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
	pingInterval = 30 * time.Second
	pongWait     = 60 * time.Second
	writeWait    = 10 * time.Second
	// actionCallTimeout caps every gRPC action invoked from a WS message.
	actionCallTimeout = 5 * time.Second
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
// Slow/blocking actions (PlaceBid, CreateAuction) are dispatched to a per-connection
// worker pool so that readPump is never stalled on a gRPC round-trip.
type WSManager struct {
	mu           sync.RWMutex
	connections  map[string]*connection
	subscribers  map[string]map[string]struct{} // auctionID → connID set
	connAuctions map[string]map[string]struct{} // connID → auctionID set

	connWg sync.WaitGroup // tracks active HandleConnection calls for clean shutdown

	broadcastCh   chan broadcastEnvelope
	actionWorkers chan func() // shared pool for blocking WS actions

	// tunable config (from config.yaml WS section)
	numBroadcastWorkers int
	numActionWorkers    int
	maxConnections      int
	sendBufSize         int

	actions IAuctionActions
	metrics metrics.IMetrics
	logger  applogger.IAppLogger
}

// WSManagerConfig holds runtime-tunable parameters for WSManager.
// All values come from config.yaml WS section so they can be adjusted
// without recompiling — profile first with pprof, then tune.
type WSManagerConfig struct {
	NumActionWorkers    int // goroutines executing blocking gRPC calls
	NumBroadcastWorkers int // goroutines draining broadcastCh
	BroadcastBufSize    int // broadcastCh capacity (events)
	ActionQueueSize     int // actionWorkers channel capacity (jobs)
	MaxConnections      int // hard cap on simultaneous WS connections
}

func NewWSManager(cfg WSManagerConfig, actions IAuctionActions, m metrics.IMetrics, logger applogger.IAppLogger) *WSManager {
	return &WSManager{
		connections:         make(map[string]*connection),
		subscribers:         make(map[string]map[string]struct{}),
		connAuctions:        make(map[string]map[string]struct{}),
		broadcastCh:         make(chan broadcastEnvelope, cfg.BroadcastBufSize),
		actionWorkers:       make(chan func(), cfg.ActionQueueSize),
		numBroadcastWorkers: cfg.NumBroadcastWorkers,
		numActionWorkers:    cfg.NumActionWorkers,
		maxConnections:      cfg.MaxConnections,
		sendBufSize:         128, // per-connection outbound; small (auction events are ~200B)
		actions:             actions,
		metrics:             m,
		logger:              logger,
	}
}

// Start launches background broadcast workers and action workers. Blocks until ctx is cancelled.
func (m *WSManager) Start(ctx context.Context) {
	var wg sync.WaitGroup

	// Broadcast workers: deliver events to subscribers.
	for i := 0; i < m.numBroadcastWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.runBroadcastWorker(ctx)
		}()
	}

	// Action workers: execute blocking WS actions (PlaceBid, CreateAuction)
	// off the readPump goroutine so it stays responsive.
	for i := 0; i < m.numActionWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case fn, ok := <-m.actionWorkers:
					if !ok {
						return
					}
					fn()
				}
			}
		}()
	}

	wg.Wait()
}

// runBroadcastWorker wraps broadcastWorker with panic recovery and auto-restart.
func (m *WSManager) runBroadcastWorker(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error(ctx, fmt.Errorf("broadcast worker panic: %v", r))
			// Restart the worker if the context is still active.
			if ctx.Err() == nil {
				go m.runBroadcastWorker(ctx)
			}
		}
	}()
	m.broadcastWorker(ctx)
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
	// Acquire lock ONCE, collect all target connections, then release.
	// Avoids N separate RLock/RUnlock in the subscriber loop (critical hot path).
	m.mu.RLock()
	subIDs := m.subscribers[auctionID]
	conns := make([]*connection, 0, len(subIDs))
	for connID := range subIDs {
		if conn, ok := m.connections[connID]; ok {
			conns = append(conns, conn)
		}
	}
	m.mu.RUnlock()

	for _, conn := range conns {
		select {
		case conn.send <- event:
			// WSMessagesSentInc is counted in writePump after actual write.
		default:
			m.metrics.WSBroadcastDroppedInc()
			m.logger.Info(context.Background(), "connection send buffer full, event dropped",
				"connID", conn.id, "event", event.Event)
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
		send:   make(chan WSEvent, m.sendBufSize),
	}

	m.register(conn)
	defer m.unregister(conn)
	m.metrics.WSConnectionsInc()
	defer m.metrics.WSConnectionsDec()

	// Track this connection for graceful shutdown.
	m.connWg.Add(1)
	defer m.connWg.Done()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error(ctx, fmt.Errorf("ws writePump panic: %v", r), "conn_id", conn.id)
				conn.conn.Close() // force readPump to exit too
			}
		}()
		m.writePump(ctx, conn)
	}()
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error(ctx, fmt.Errorf("ws readPump panic: %v", r), "conn_id", conn.id)
				conn.conn.Close() // force writePump to exit too
			}
		}()
		m.readPump(ctx, conn)
	}()

	conn.send <- WSEvent{Event: "connected"}

	wg.Wait()
}

// Shutdown gracefully closes all active WebSocket connections and waits for
// their handlers to finish. ctx controls the maximum wait time.
func (m *WSManager) Shutdown(ctx context.Context) {
	// Send a "going away" close frame to every active connection.
	// writePump may have already done this via ctx.Done(), but we do it
	// explicitly here so connections that are still mid-read also get closed.
	m.mu.RLock()
	for _, c := range m.connections {
		c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
			time.Now().Add(writeWait),
		)
		c.conn.Close()
	}
	m.mu.RUnlock()

	// Wait for all HandleConnection goroutine pairs to finish.
	done := make(chan struct{})
	go func() {
		m.connWg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		// Timeout: force-close any stragglers.
		m.mu.Lock()
		for _, c := range m.connections {
			c.conn.Close()
		}
		m.mu.Unlock()
	}
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
		// Async: gRPC round-trip must not block readPump.
		payload := msg.Payload
		m.submitAction(ctx, func() {
			m.handleCreateAuction(ctx, conn, payload)
		})
	case "place_bid":
		// Async: gRPC round-trip must not block readPump.
		auctionID := msg.AuctionID
		payload := msg.Payload
		m.submitAction(ctx, func() {
			m.handlePlaceBid(ctx, conn, auctionID, payload)
		})
	default:
		trySend(conn.send, WSEvent{Event: "error", Error: fmt.Sprintf("unknown action: %s", msg.Action)})
	}
}

// submitAction enqueues fn into the shared action worker pool.
// If the pool is full the action is dropped and an error is sent to the client.
func (m *WSManager) submitAction(ctx context.Context, fn func()) {
	select {
	case m.actionWorkers <- fn:
	default:
		m.logger.Info(ctx, "action worker pool full, dropping action")
		m.metrics.WSBroadcastDroppedInc()
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

	callCtx, cancel := context.WithTimeout(ctx, actionCallTimeout)
	defer cancel()

	auctionID, err := m.actions.CreateAuction(callCtx, req, conn.userID)
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

	callCtx, cancel := context.WithTimeout(ctx, actionCallTimeout)
	defer cancel()

	result, err := m.actions.PlaceBid(callCtx, auctionID, p.Amount, conn.userID)
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

// ConnectionCount returns the current number of active WebSocket connections.
func (m *WSManager) ConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// MaxConnectionsLimit returns the configured maximum simultaneous WS connections.
func (m *WSManager) MaxConnectionsLimit() int {
	return m.maxConnections
}
