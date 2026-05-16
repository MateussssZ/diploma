package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/controllers"
	"apigateway/internal/integrations/cache"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"apigateway/internal/utils"
)

// IWSManager is the interface implemented by wsmanager.WSManager.
type IWSManager interface {
	HandleConnection(ctx context.Context, conn *websocket.Conn, userID string)
	ConnectionCount() int
	MaxConnectionsLimit() int
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type IAuctionHandlers interface {
	GetAuctions(w http.ResponseWriter, r *http.Request)
	GetAuctionByID(w http.ResponseWriter, r *http.Request)
	GetUserAuctions(w http.ResponseWriter, r *http.Request)
	GetSubscribedAuctions(w http.ResponseWriter, r *http.Request)
	CreateAuction(w http.ResponseWriter, r *http.Request)
	ConnectAuction(w http.ResponseWriter, r *http.Request)
}

type AuctionHandlersDep struct {
	Responder    IResponder               `validate:"required"`
	AuctionCtrl  controllers.IAuctionCtrl `validate:"required"`
	WSManager    IWSManager               `validate:"required"`
	JWTSecret    string                   `validate:"required"`
	CacheManager *cache.CacheManager      `validate:"required"`
}

type AuctionHandlers struct {
	responder    IResponder
	auctionCtrl  controllers.IAuctionCtrl
	wsManager    IWSManager
	jwtSecret    []byte
	cacheManager *cache.CacheManager
}

func NewAuctionHandlers(dep AuctionHandlersDep) (*AuctionHandlers, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewAuctionHandlers", err)
	}
	return &AuctionHandlers{
		responder:    dep.Responder,
		auctionCtrl:  dep.AuctionCtrl,
		wsManager:    dep.WSManager,
		jwtSecret:    []byte(dep.JWTSecret),
		cacheManager: dep.CacheManager,
	}, nil
}

func (h *AuctionHandlers) GetAuctions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	auctions, err := h.auctionCtrl.GetAuctions(ctx, page, pageSize)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, auctions)
}

func (h *AuctionHandlers) GetAuctionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auctionID := mux.Vars(r)["auction_id"]
	if auctionID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("auction_id"), WithStatusCode(http.StatusBadRequest))
		return
	}
	if _, err := strconv.ParseInt(auctionID, 10, 64); err != nil {
		h.responder.WriteError(ctx, w, errorspkg.NewValidationError("GetAuctionByID", err), WithStatusCode(http.StatusBadRequest))
		return
	}

	auction, err := h.cacheManager.GetAuctionDetail(ctx, auctionID, func() (*models.AuctionDetail, error) {
		return h.auctionCtrl.GetAuctionByID(ctx, auctionID)
	})
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, auction)
}

func (h *AuctionHandlers) GetUserAuctions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(utils.CtxUserID).(string)
	if !ok || userID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("user_id"), WithStatusCode(http.StatusUnauthorized))
		return
	}

	auctions, err := h.auctionCtrl.GetUserAuctions(ctx, userID)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, auctions)
}

func (h *AuctionHandlers) GetSubscribedAuctions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(utils.CtxUserID).(string)
	if !ok || userID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("user_id"), WithStatusCode(http.StatusUnauthorized))
		return
	}

	auctions, err := h.auctionCtrl.GetSubscribedAuctions(ctx, userID)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, auctions)
}

func (h *AuctionHandlers) CreateAuction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	if err := validate.Struct(data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	userID, ok := ctx.Value(utils.CtxUserID).(string)
	if !ok || userID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("user_id"), WithStatusCode(http.StatusUnauthorized))
		return
	}

	auctionID, err := h.auctionCtrl.CreateAuction(ctx, data, userID)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, models.CreateAuctionResponse{AuctionID: auctionID})
}

// ConnectAuction upgrades the HTTP connection to WebSocket and hands it off to WSManager.
// Auth token is passed via the ?token= query parameter (standard practice for browser WS clients).
func (h *AuctionHandlers) ConnectAuction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token query parameter", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	userID, err := claims.GetSubject()
	if err != nil || userID == "" {
		http.Error(w, "invalid token subject", http.StatusUnauthorized)
		return
	}

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.responder.WriteError(ctx, w, fmt.Errorf("ws upgrade: %w", err))
		return
	}
	defer ws.Close()

	if h.wsManager.ConnectionCount() >= h.wsManager.MaxConnectionsLimit() {
		ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "too many connections"))
		return
	}

	h.wsManager.HandleConnection(ctx, ws, userID)
}
