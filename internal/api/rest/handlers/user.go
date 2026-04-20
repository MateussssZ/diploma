package handlers

import (
	"encoding/json"
	"net/http"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/controllers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

type IUserHandlers interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
}

type UserHandlersDep struct {
	Responder IResponder            `validate:"required"`
	UserCtrl  controllers.IUserCtrl `validate:"required"`
}

type UserHandlers struct {
	responder IResponder
	userCtrl  controllers.IUserCtrl
}

func NewUserHandlers(dep UserHandlersDep) (*UserHandlers, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewUserHandlers", err)
	}
	return &UserHandlers{
		responder: dep.Responder,
		userCtrl:  dep.UserCtrl,
	}, nil
}

func (h *UserHandlers) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	if err := h.userCtrl.Register(ctx, data); err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, "User registered successfully")
}

func (h *UserHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	tokens, err := h.userCtrl.Login(ctx, data)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, tokens)
}

func (h *UserHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	if err := h.userCtrl.Logout(ctx, data.RefreshToken); err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, "User logged out successfully")
}

func (h *UserHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	tokens, err := h.userCtrl.Refresh(ctx, data.RefreshToken)
	if err != nil {
		h.responder.WriteError(ctx, w, err)
		return
	}

	h.responder.WriteJSON(ctx, w, tokens)
}
