package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/controllers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

type IAuctionHandlers interface {
	GetAuctions(w http.ResponseWriter, r *http.Request)
	GetAuctionByID(w http.ResponseWriter, r *http.Request)
	GetUserAuctions(w http.ResponseWriter, r *http.Request)
	GetSubscribedAuctions(w http.ResponseWriter, r *http.Request)
	CreateAuction(w http.ResponseWriter, r *http.Request)
}

type AuctionHandlersDep struct {
	Responder   IResponder               `validate:"required"`
	AuctionCtrl controllers.IAuctionCtrl `validate:"required"`
}

type AuctionHandlers struct {
	responder   IResponder
	auctionCtrl controllers.IAuctionCtrl
}

func NewAuctionHandlers(dep AuctionHandlersDep) (*AuctionHandlers, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewAuctionHandlers", err)
	}
	return &AuctionHandlers{
		responder:   dep.Responder,
		auctionCtrl: dep.AuctionCtrl,
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

	auction, err := h.auctionCtrl.GetAuctionByID(ctx, auctionID)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, auction)
}

func (h *AuctionHandlers) GetUserAuctions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value("user_id").(string)
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

	userID, ok := ctx.Value("user_id").(string)
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

	userID, ok := ctx.Value("user_id").(string)
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
