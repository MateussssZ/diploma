// handlers.go — HTTP handlers for the polling baseline service.
//
// Architecture difference from the optimised service:
//   - NO WebSocket
//   - NO Redis cache   (every GET hits gRPC directly)
//   - NO Kafka consumer (auction state is fetched on demand)
//
// Clients that want "real-time" updates must poll GET /auctions/:id repeatedly.
// This is the Chapter 4.2 baseline.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	auctionpb "apigateway/api/grpc/auctionservice"
	authpb "apigateway/api/grpc/authservice"
	"pollingsvc/internal/client"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type ctxKey string

const ctxUserID ctxKey = "user_id"

// ─── Handler container ───────────────────────────────────────────────────────

type Handlers struct {
	auction   *client.AuctionClient
	auth      *client.AuthClient
	jwtSecret []byte
}

func New(auctionCli *client.AuctionClient, authCli *client.AuthClient, jwtSecret string) *Handlers {
	return &Handlers{
		auction:   auctionCli,
		auth:      authCli,
		jwtSecret: []byte(jwtSecret),
	}
}

// ─── Middleware ───────────────────────────────────────────────────────────────

// AuthMiddleware validates the Bearer JWT and puts userID into context.
func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Authorization")
		if len(raw) < 8 || raw[:7] != "Bearer " {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		tok, err := jwt.Parse(raw[7:], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return h.jwtSecret, nil
		})
		if err != nil || !tok.Valid {
			writeErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		claims, _ := tok.Claims.(jwt.MapClaims)
		sub, _ := claims.GetSubject()
		ctx := context.WithValue(r.Context(), ctxUserID, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ─── Health ───────────────────────────────────────────────────────────────────

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "mode": "polling"})
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.Register(r.Context(), req.Login, req.Password, req.Email); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "User registered successfully")
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req authpb.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "logged out")
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// ─── Auctions ─────────────────────────────────────────────────────────────────

// GetAuctions returns a paginated list — hits gRPC every time (no cache).
func (h *Handlers) GetAuctions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if size <= 0 {
		size = 20
	}
	resp, err := h.auction.GetLots(r.Context(), int32(page), int32(size))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, lotsListToJSON(resp))
}

// GetAuctionByID returns a single auction — hits gRPC every time (no cache).
// Polling clients call this endpoint repeatedly to detect price changes.
func (h *Handlers) GetAuctionByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["auction_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("invalid auction_id: %v", err))
		return
	}
	lot, err := h.auction.GetLot(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, lotFullToJSON(lot))
}

// CreateAuction creates a new auction.
func (h *Handlers) CreateAuction(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxUserID).(string)
	sellerID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid user_id in token")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartPrice  int64  `json:"start_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	lot, err := h.auction.CreateLot(r.Context(), &auctionpb.CreateLotRequest{
		Title:         req.Title,
		Description:   req.Description,
		StartingPrice: strconv.FormatInt(req.StartPrice, 10),
		Status:        auctionpb.LotStatus_ACTIVE,
		SellerId:      sellerID,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"auction_id": strconv.FormatInt(lot.Id, 10)})
}

// PlaceBid places a bid. No cache invalidation needed (no cache).
func (h *Handlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxUserID).(string)
	bidderID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid user_id in token")
		return
	}

	auctionID := mux.Vars(r)["auction_id"]
	lotID, err := strconv.ParseInt(auctionID, 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid auction_id")
		return
	}

	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Amount <= 0 {
		writeErr(w, http.StatusBadRequest, "amount must be > 0")
		return
	}

	_, err = h.auction.PlaceBid(r.Context(), lotID, bidderID, strconv.FormatInt(req.Amount, 10))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ─── JSON helpers ─────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func lotFullToJSON(l *auctionpb.LotFull) map[string]interface{} {
	return map[string]interface{}{
		"auction_id":    strconv.FormatInt(l.Id, 10),
		"title":         l.Title,
		"description":   l.Description,
		"start_price":   client.ParsePrice(l.StartingPrice),
		"current_price": client.ParsePrice(l.CurrentPrice),
		"status":        l.Status.String(),
		"creator_id":    strconv.FormatInt(l.SellerId, 10),
	}
}

func lotsListToJSON(resp *auctionpb.GetLotsResponse) map[string]interface{} {
	auctions := make([]map[string]interface{}, 0, len(resp.Content))
	for _, lot := range resp.Content {
		auctions = append(auctions, map[string]interface{}{
			"auction_id":    strconv.FormatInt(lot.Id, 10),
			"title":         lot.Title,
			"current_price": client.ParsePrice(lot.CurrentPrice),
			"status":        lot.Status.String(),
		})
	}
	return map[string]interface{}{
		"auctions":    auctions,
		"page":        resp.Page,
		"page_size":   resp.Size,
		"total_pages": resp.TotalPages,
		"total_items": resp.TotalElements,
	}
}
