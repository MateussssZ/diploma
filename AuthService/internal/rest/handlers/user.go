package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"authservice/internal/pkg/errorspkg"
	"authservice/internal/usecases"
)

type IUserHandlers interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
}

type UserHandlersDep struct {
	Responder IResponder
	Usecase   usecases.IUserUsecase
}

type UserHandlers struct {
	UserHandlersDep
}

func NewUserHandlers(dep UserHandlersDep) *UserHandlers {
	return &UserHandlers{dep}
}

type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tokenPairResponse struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	AccessTokenExpiry  int64  `json:"access_token_expiry"`
	RefreshTokenExpiry int64  `json:"refresh_token_expiry"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func errStatusCode(err error) WriteErrorOption {
	var h errorspkg.ErrorHandler
	if errors.As(err, &h) {
		return WithStatusCode(h.HTTPStatus())
	}
	return WithStatusCode(http.StatusInternalServerError)
}

func (h *UserHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.Responder.WriteError(r.Context(), w, errors.New("invalid request body"),
			WithStatusCode(http.StatusBadRequest))
		return
	}
	err := h.Usecase.Register(r.Context(), body.Login, body.Password, body.Email)
	if err != nil {
		h.Responder.WriteError(r.Context(), w, err, errStatusCode(err))
		return
	}
	h.Responder.WriteJSON(r.Context(), w, map[string]string{"status": "registered"})
}

func (h *UserHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.Responder.WriteError(r.Context(), w, errors.New("invalid request body"),
			WithStatusCode(http.StatusBadRequest))
		return
	}
	pair, err := h.Usecase.Login(r.Context(), body.Login, body.Password)
	if err != nil {
		h.Responder.WriteError(r.Context(), w, err, errStatusCode(err))
		return
	}
	h.Responder.WriteJSON(r.Context(), w, tokenPairResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessTokenExpiry:  pair.AccessTokenExpiry,
		RefreshTokenExpiry: pair.RefreshTokenExpiry,
	})
}

func (h *UserHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var body logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.Responder.WriteError(r.Context(), w, errors.New("invalid request body"),
			WithStatusCode(http.StatusBadRequest))
		return
	}
	if err := h.Usecase.Logout(r.Context(), body.RefreshToken); err != nil {
		h.Responder.WriteError(r.Context(), w, err, errStatusCode(err))
		return
	}
	h.Responder.WriteJSON(r.Context(), w, map[string]string{"status": "logged out"})
}

func (h *UserHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.Responder.WriteError(r.Context(), w, errors.New("invalid request body"),
			WithStatusCode(http.StatusBadRequest))
		return
	}
	pair, err := h.Usecase.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		h.Responder.WriteError(r.Context(), w, err, errStatusCode(err))
		return
	}
	h.Responder.WriteJSON(r.Context(), w, tokenPairResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessTokenExpiry:  pair.AccessTokenExpiry,
		RefreshTokenExpiry: pair.RefreshTokenExpiry,
	})
}
