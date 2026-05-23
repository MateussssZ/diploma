package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"net/http"
)

// IResponder is interface for building response (error) output to the client
type IResponder interface {
	WriteError(ctx context.Context, w http.ResponseWriter, err error, opts ...WriteErrorOption)
	WriteJSON(ctx context.Context, w http.ResponseWriter, body any, opts ...WriteErrorOption)
}

type ResponderDep struct {
	Logger  applogger.IAppLogger `validate:"required"`
	Metrics metrics.IMetrics     `validate:"required"`
}

type Responder struct {
	logger  applogger.IAppLogger
	metrics metrics.IMetrics
}

func NewResponder(dep ResponderDep) (*Responder, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewResponder", err)
	}

	return &Responder{
		logger:  dep.Logger,
		metrics: dep.Metrics,
	}, nil
}

func (r *Responder) write(ctx context.Context, w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		r.logger.Error(ctx, fmt.Errorf("json.encode body - error: %w", err))
	}
}

func (r *Responder) WriteError(ctx context.Context, w http.ResponseWriter, err error, opts ...WriteErrorOption) {
	r.metrics.OperationsErrorsTotalInc()

	// инициализация конфига по умолчанию
	cfg := &writeErrorCfg{}
	// применение всех переданных опций (при их наличии), изменяющих конфиг
	for _, opt := range opts {
		opt(cfg)
	}

	// сбор атрибутов логирования
	attrs := cfg.getAttrs()

	var eventID string
	if cfg.logAsInfo {
		r.logger.Info(ctx, err.Error(), attrs...)
	} else {
		eventID = r.logger.ErrorWithEventID(ctx, fmt.Errorf("http request failed: %w", err), attrs...)
	}

	responseErr := struct {
		ErrorID string `json:"ErrorID"`
		Message string `json:"Message"`
	}{
		ErrorID: eventID,
		Message: err.Error(),
	}

	var statusCode = http.StatusInternalServerError
	var handledErr errorspkg.ErrorHandler

	if errors.As(err, &handledErr) {
		statusCode = handledErr.HTTPStatus()
	}
	if cfg.statusCode != 0 {
		statusCode = cfg.statusCode
	}

	r.write(ctx, w, statusCode, responseErr)
}

func (r *Responder) WriteJSON(ctx context.Context, w http.ResponseWriter, body any, opts ...WriteErrorOption) {
	r.logger.Debug(ctx, "http request succeeded")

	// инициализация конфига по умолчанию
	cfg := &writeErrorCfg{}
	// применение всех переданных опций (при их наличии), изменяющих конфиг
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.err != nil {
		r.logger.Error(ctx, cfg.err, cfg.getAttrs()...)
	}

	r.write(ctx, w, http.StatusOK, body)
}

type WriteErrorOption func(*writeErrorCfg)

type writeErrorCfg struct {
	requestBody any
	statusCode  int
	logAsInfo   bool
	tags        []any
	err         error
}

func (cfg *writeErrorCfg) getAttrs() []any {
	var attrs []any
	if cfg.requestBody != nil {
		attrs = append(attrs, "body", cfg.requestBody)
	}
	if len(cfg.tags) != 0 {
		attrs = append(attrs, cfg.tags...)
	}

	return attrs
}

// WithRequestBody опция логирования тела запроса
func WithRequestBody(body any) WriteErrorOption {
	return func(c *writeErrorCfg) {
		c.requestBody = body
	}
}

// WithStatusCode опция изменения статус кода
func WithStatusCode(code int) WriteErrorOption {
	return func(c *writeErrorCfg) {
		c.statusCode = code
	}
}

// WithTags опция установки дополнительных атрибутов для логирования
func WithTags(key string, val any) WriteErrorOption {
	return func(c *writeErrorCfg) {
		if c.tags == nil {
			c.tags = make([]any, 0, 2)
		}
		c.tags = append(c.tags, key, val)
	}
}