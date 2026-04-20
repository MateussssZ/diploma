package middlewares

import (
	"fmt"
	"github.com/gorilla/mux"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"net/http"
	"time"
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

		start := time.Now()

		mw.metrics.OperationsTotalInc()

		route := mux.CurrentRoute(r)
		path, err := route.GetPathTemplate()
		if err != nil {
			mw.logger.Error(r.Context(), fmt.Errorf("route.GetPathTemplate - error: %w", err))
		}

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		mw.metrics.HTTPServerRequestDurationInc(path, r.Method, duration)
	})
}
