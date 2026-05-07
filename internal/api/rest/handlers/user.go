package handlers

import (
	"encoding/json"
	"net/http"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/controllers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// grpcStatusCode converts a gRPC error into a WithStatusCode option.
// Unauthenticated → 401, NotFound → 404, AlreadyExists → 409, etc.
func grpcStatusCode(err error) WriteErrorOption {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return WithStatusCode(http.StatusUnauthorized)
		case codes.NotFound:
			return WithStatusCode(http.StatusNotFound)
		case codes.AlreadyExists:
			return WithStatusCode(http.StatusConflict)
		case codes.InvalidArgument:
			return WithStatusCode(http.StatusBadRequest)
		case codes.PermissionDenied:
			return WithStatusCode(http.StatusForbidden)
		}
	}
	return WithStatusCode(http.StatusInternalServerError)
}

func (h *UserHandlers) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.responder.WriteError(ctx, w, err, WithStatusCode(http.StatusBadRequest))
		return
	}

	if err := h.userCtrl.Register(ctx, data); err != nil {
		h.responder.WriteError(ctx, w, err, grpcStatusCode(err))
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
		h.responder.WriteError(ctx, w, err, grpcStatusCode(err))
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
		h.responder.WriteError(ctx, w, err, grpcStatusCode(err))
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
		h.responder.WriteError(ctx, w, err, grpcStatusCode(err))
		return
	}

	h.responder.WriteJSON(ctx, w, tokens)
}
