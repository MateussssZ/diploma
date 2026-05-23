// main.go — entry point for the polling baseline service.
//
// This service intentionally omits every optimisation that the main apigateway adds:
//   - No WebSocket (clients must poll GET /auctions/:id to see updates)
//   - No Redis cache (every request goes to AuctionService via gRPC)
//   - No Kafka consumer (no push notifications)
//
// Purpose: Chapter 4.2 performance baseline.
// Run side-by-side with the optimised service on a different port (default :8082).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"pollingsvc/config"
	"pollingsvc/internal/client"
	"pollingsvc/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	// gRPC client — AuctionService
	auctionCli, auctionConn, err := client.NewAuctionClient(cfg.AuctionService.Address)
	if err != nil {
		slog.Error("failed to connect to AuctionService", "err", err)
		os.Exit(1)
	}
	defer auctionConn.Close()

	// gRPC client — AuthService (unix socket)
	authCli, authConn, err := client.NewAuthClient(cfg.AuthService.SocketPath)
	if err != nil {
		slog.Error("failed to connect to AuthService", "err", err)
		os.Exit(1)
	}
	defer authConn.Close()

	h := handlers.New(auctionCli, authCli, cfg.AuthService.JWTSecret)

	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)

	// Auth
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	auth.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	auth.HandleFunc("/logout", h.Logout).Methods(http.MethodPost)
	auth.HandleFunc("/refresh", h.Refresh).Methods(http.MethodPost)

	// Auctions (all protected)
	auction := r.PathPrefix("/auctions").Subrouter()
	auction.Use(h.AuthMiddleware)
	auction.HandleFunc("", h.GetAuctions).Methods(http.MethodGet)
	auction.HandleFunc("", h.CreateAuction).Methods(http.MethodPost)
	auction.HandleFunc("/{auction_id}", h.GetAuctionByID).Methods(http.MethodGet)
	auction.HandleFunc("/{auction_id}/bid", h.PlaceBid).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.RESTServer.ListenPort),
		Handler:      r,
		ReadTimeout:  cfg.RESTServer.ReadTimeout,
		WriteTimeout: cfg.RESTServer.WriteTimeout,
	}

	go func() {
		slog.Info("polling_svc starting", "addr", srv.Addr, "mode", "http-polling (no WS, no Redis, no Kafka)")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down polling_svc")

	shutCtx, cancel := context.WithTimeout(context.Background(), cfg.RESTServer.ShutdownTimeout)
	defer cancel()
	srv.Shutdown(shutCtx)

	slog.Info("polling_svc stopped")
}
