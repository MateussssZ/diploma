package middlewares

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"apigateway/internal/utils"
	"net/http"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() (*RequestIDMiddleware, error) {
	return &RequestIDMiddleware{}, nil
}

func (mw *RequestIDMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get(string(utils.CtxRequestID))
		if requestID == "" {
			requestID = "G-" + uuid.New().String()
		}

		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetRequest(r)
			scope.SetTag("http.method", r.Method)
			scope.SetTag("http.url", r.URL.String())
		})

		ctx := sentry.SetHubOnContext(r.Context(), hub)
		ctx = context.WithValue(ctx, utils.CtxRequestID, requestID)
		ctx = context.WithValue(ctx, utils.CtxRequestMethod, r.URL.String())

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
