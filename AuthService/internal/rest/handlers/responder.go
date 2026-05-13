package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"log/slog"
	"net/http"
)

// IResponder — интерфейс вывода ответов клиенту (идентичен основному проекту).
type IResponder interface {
	WriteError(ctx context.Context, w http.ResponseWriter, err error, opts ...WriteErrorOption)
	WriteJSON(ctx context.Context, w http.ResponseWriter, body any, opts ...WriteErrorOption)
}

type Responder struct {
	logger *slog.Logger
}

func NewResponder(logger *slog.Logger) *Responder {
	return &Responder{logger: logger}
}

func (r *Responder) write(ctx context.Context, w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		r.logger.ErrorContext(ctx, "json.encode body", "err", err)
	}
}

func (r *Responder) WriteError(ctx context.Context, w http.ResponseWriter, err error, opts ...WriteErrorOption) {
	cfg := &writeErrorCfg{}
	for _, opt := range opts {
		opt(cfg)
	}

	attrs := cfg.getAttrs()
	r.logger.ErrorContext(ctx, fmt.Sprintf("http request failed: %v", err), attrs...)

	responseErr := struct {
		Message string `json:"Message"`
	}{
		Message: err.Error(),
	}

	statusCode := http.StatusInternalServerError
	var handledErr HTTPStatusError
	if errors.As(err, &handledErr) {
		statusCode = handledErr.HTTPStatus()
	}
	if cfg.statusCode != 0 {
		statusCode = cfg.statusCode
	}

	r.write(ctx, w, statusCode, responseErr)
}

func (r *Responder) WriteJSON(ctx context.Context, w http.ResponseWriter, body any, opts ...WriteErrorOption) {
	r.logger.DebugContext(ctx, "http request succeeded")
	r.write(ctx, w, http.StatusOK, body)
}

// HTTPStatusError — интерфейс для ошибок, несущих HTTP-статус.
type HTTPStatusError interface {
	error
	HTTPStatus() int
}

// WriteErrorOption — функция изменения настроек ответа об ошибке.
type WriteErrorOption func(*writeErrorCfg)

type writeErrorCfg struct {
	requestBody any
	statusCode  int
	logAsInfo   bool
	tags        []any
}

func (cfg *writeErrorCfg) getAttrs() []any {
	var attrs []any
	if cfg.requestBody != nil {
		attrs = append(attrs, "body", cfg.requestBody)
	}
	attrs = append(attrs, cfg.tags...)
	return attrs
}

func WithStatusCode(code int) WriteErrorOption {
	return func(c *writeErrorCfg) { c.statusCode = code }
}

func WithRequestBody(body any) WriteErrorOption {
	return func(c *writeErrorCfg) { c.requestBody = body }
}

func WithLogAsInfo() WriteErrorOption {
	return func(c *writeErrorCfg) { c.logAsInfo = true }
}

func WithTags(key string, val any) WriteErrorOption {
	return func(c *writeErrorCfg) {
		c.tags = append(c.tags, key, val)
	}
}
