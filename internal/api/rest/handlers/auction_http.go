package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/controllers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/utils"
)

// IHTTPAuctionHandlers — HTTP-only auction handlers for benchmarking
type IHTTPAuctionHandlers interface {
	GetAuctions(w http.ResponseWriter, r *http.Request)
	GetAuctionByID(w http.ResponseWriter, r *http.Request)
	CreateAuction(w http.ResponseWriter, r *http.Request)
	PlaceBid(w http.ResponseWriter, r *http.Request)
}

type HTTPAuctionHandlersDep struct {
	Responder                IResponder               `validate:"required"`
	AuctionCtrl              controllers.IAuctionCtrl `validate:"required"`
	AuctionServiceHTTPClient interface{}              `validate:"required"` // HTTPAuctionServiceClient
}

type HTTPAuctionHandlers struct {
	responder                IResponder
	auctionCtrl              controllers.IAuctionCtrl
	auctionServiceHTTPClient interface{}
}

func NewHTTPAuctionHandlers(dep HTTPAuctionHandlersDep) (*HTTPAuctionHandlers, error) {
	return &HTTPAuctionHandlers{
		responder:                dep.Responder,
		auctionCtrl:              dep.AuctionCtrl,
		auctionServiceHTTPClient: dep.AuctionServiceHTTPClient,
	}, nil
}

func (h *HTTPAuctionHandlers) GetAuctions(w http.ResponseWriter, r *http.Request) {
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

func (h *HTTPAuctionHandlers) GetAuctionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auctionID := mux.Vars(r)["auction_id"]
	if auctionID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("auction_id"), WithStatusCode(http.StatusBadRequest))
		return
	}

	auction, err := h.auctionCtrl.GetAuctionByID(ctx, auctionID)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, auction)
}

func (h *HTTPAuctionHandlers) CreateAuction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
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

func (h *HTTPAuctionHandlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auctionID := mux.Vars(r)["auction_id"]
	if auctionID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("auction_id"), WithStatusCode(http.StatusBadRequest))
		return
	}

	var data struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	userID, ok := ctx.Value(utils.CtxUserID).(string)
	if !ok || userID == "" {
		h.responder.WriteError(ctx, w, errorspkg.NewUnitIsMissedError("user_id"), WithStatusCode(http.StatusUnauthorized))
		return
	}

	result, err := h.auctionCtrl.PlaceBid(ctx, auctionID, data.Amount, userID)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, result)
}
