package handlers

import (
	"fmt"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"net/http"
)

type IBaseHandlers interface {
	Health(w http.ResponseWriter, r *http.Request)
	Version(w http.ResponseWriter, r *http.Request)
}

type Dep struct {
	AppVersion string               `validate:"required,min=6"`
	Logger     applogger.IAppLogger `validate:"required"`
}

type BaseHandlers struct {
	appVersion string
	logger     applogger.IAppLogger
}

func NewBaseHandlers(dep Dep) (*BaseHandlers, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewBaseHandlers", err)
	}

	return &BaseHandlers{
		appVersion: dep.AppVersion,
		logger:     dep.Logger,
	}, nil
}

func (h *BaseHandlers) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *BaseHandlers) Version(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("{\"version\": \"%s\"}", h.appVersion)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(response))
	if err != nil {
		h.logger.Error(r.Context(), fmt.Errorf("rest handler Version - error: %w", err))
	}
}
