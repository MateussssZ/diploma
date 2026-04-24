package middlewares

import (
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type MetricsMiddlewareDep struct {
	Logger  applogger.IAppLogger `validate:"required"`
	Metrics metrics.IMetrics     `validate:"required"`
}

type MetricsMiddleware struct {
	logger  applogger.IAppLogger
	metrics metrics.IMetrics
}

func NewMetricsMiddleware(dep MetricsMiddlewareDep) (*MetricsMiddleware, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewMetricsMiddleware", err)
	}

	return &MetricsMiddleware{
		logger:  dep.Logger,
		metrics: dep.Metrics,
	}, nil
}

func (mw *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		mw.logger.Debug(r.Context(), "http request started")

		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// WebSocket upgrade requires http.Hijacker — skip wrapping for WS routes
		if strings.HasPrefix(r.URL.Path, "/ws") {
			next.ServeHTTP(w, r)
			return
		}

		route := mux.CurrentRoute(r)
		path, err := route.GetPathTemplate()
		if err != nil {
			mw.logger.Error(r.Context(), fmt.Errorf("route.GetPathTemplate - error: %w", err))
			path = r.URL.Path
		}

		mw.metrics.OperationsTotalInc()

		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		mw.metrics.HTTPServerRequestDurationInc(path, r.Method, duration)

		if rw.status >= 400 {
			mw.metrics.HTTPServerRequestsErrorsTotalInc(path, r.Method)
			mw.metrics.OperationsErrorsTotalInc()
		}
	})
}

// statusWriter wraps http.ResponseWriter to capture the written status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
