package rest

import (
	"apigateway/internal/api/rest/handlers"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouteDep struct {
	Metrics             metrics.IMetrics          `validate:"required"`
	RequestIDMiddleware mux.MiddlewareFunc        `validate:"required"`
	RecoveryMiddleware  mux.MiddlewareFunc        `validate:"required"`
	MetricsMiddleware   mux.MiddlewareFunc        `validate:"required"`
	AuthMiddleware      mux.MiddlewareFunc        `validate:"required"`
	BaseHandlers        handlers.IBaseHandlers    `validate:"required"`
	UserHandlers        handlers.IUserHandlers    `validate:"required"`
	AuctionHandlers     handlers.IAuctionHandlers `validate:"required"`
}

func NewRoute(dep RouteDep) (http.Handler, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewRoute", err)
	}

	r := mux.NewRouter()

	r.Use(dep.RequestIDMiddleware.Middleware)
	r.Use(dep.RecoveryMiddleware.Middleware)
	r.Use(dep.MetricsMiddleware.Middleware)

	r.HandleFunc("/health", dep.BaseHandlers.Health).Methods(http.MethodGet)
	r.HandleFunc("/version", dep.BaseHandlers.Version).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.HandlerFor(
		dep.Metrics.GetPrometheusRegistry(),
		promhttp.HandlerOpts{},
	)).Methods(http.MethodGet)

	// User and authentication routes
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", dep.UserHandlers.Register).Methods(http.MethodPost)
	auth.HandleFunc("/login", dep.UserHandlers.Login).Methods(http.MethodPost)
	auth.HandleFunc("/logout", dep.UserHandlers.Logout).Methods(http.MethodPost)
	auth.HandleFunc("/refresh", dep.UserHandlers.Refresh).Methods(http.MethodPost)

	// Auction routes (protected)
	auction := r.PathPrefix("/auctions").Subrouter()
	auction.Use(dep.AuthMiddleware)
	auction.HandleFunc("", dep.AuctionHandlers.GetAuctions).Methods(http.MethodGet)
	auction.HandleFunc("", dep.AuctionHandlers.CreateAuction).Methods(http.MethodPost)
	auction.HandleFunc("/user", dep.AuctionHandlers.GetUserAuctions).Methods(http.MethodGet)
	auction.HandleFunc("/subscribed", dep.AuctionHandlers.GetSubscribedAuctions).Methods(http.MethodGet)
	auction.HandleFunc("/{auction_id}", dep.AuctionHandlers.GetAuctionByID).Methods(http.MethodGet)

	// WebSocket endpoint (auth через query param ?token=...)
	ws := r.PathPrefix("/ws").Subrouter()
	ws.HandleFunc("/auctions", dep.AuctionHandlers.ConnectAuction).Methods(http.MethodGet)

	return r, nil
}
