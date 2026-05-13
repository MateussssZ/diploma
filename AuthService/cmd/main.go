package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	authpb "authservice/api/auth"
	"authservice/config"
	"authservice/internal/repo/postgres"
	"authservice/internal/rest"
	"authservice/internal/rest/handlers"
	"authservice/internal/server"
	"authservice/internal/usecases"

	_ "github.com/jackc/pgx/v5/stdlib" // register "pgx" driver for database/sql (used by goose)
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// ── database ──────────────────────────────────────────────────────────────
	pool, err := postgres.NewPool(ctx, cfg.Postgres.ToDSN())
	if err != nil {
		logger.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// ── migrations ────────────────────────────────────────────────────────────
	migDB, err := goose.OpenDBWithDriver("postgres", cfg.Postgres.ToDSN())
	if err != nil {
		logger.Error("goose open db", "err", err)
		os.Exit(1)
	}
	if err := goose.UpContext(ctx, migDB, "migrations"); err != nil {
		logger.Error("goose up", "err", err)
		os.Exit(1)
	}
	_ = migDB.Close()

	// ── wiring ────────────────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepo(pool)

	userUsecase, err := usecases.NewUserUsecase(usecases.UserUsecaseDep{
		UserRepo:   userRepo,
		JWTSecret:  cfg.Auth.JWTSecret,
		AccessTTL:  cfg.Auth.AccessTTL,
		RefreshTTL: cfg.Auth.RefreshTTL,
	})
	if err != nil {
		logger.Error("create usecase", "err", err)
		os.Exit(1)
	}

	authServer := server.NewAuthServer(userUsecase)

	// ── gRPC server (unix socket) ─────────────────────────────────────────────
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.GRPCServer.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.GRPCServer.MaxSendMsgSize),
	)
	authpb.RegisterAuthServiceServer(grpcServer, authServer)

	// Remove stale socket file if it exists from a previous run.
	_ = os.Remove(cfg.GRPCServer.SocketPath)

	lis, err := net.Listen("unix", cfg.GRPCServer.SocketPath)
	if err != nil {
		logger.Error("listen unix socket", "path", cfg.GRPCServer.SocketPath, "err", err)
		os.Exit(1)
	}
	defer os.Remove(cfg.GRPCServer.SocketPath)

	go func() {
		logger.Info("AuthService gRPC listening", "socket", cfg.GRPCServer.SocketPath)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("grpc serve", "err", err)
		}
	}()

	responder := handlers.NewResponder(logger)
	userHandlers := handlers.NewUserHandlers(handlers.UserHandlersDep{
		Responder: responder,
		Usecase:   userUsecase,
	})
	router := rest.NewRouter(rest.RouterDep{
		Responder:   responder,
		UserHandler: userHandlers,
	})
	httpSrv := rest.NewServer(rest.Options{
		Address: cfg.HTTPServer.Address,
	}, router, logger)

	go func() {
		if err := httpSrv.Start(ctx); err != nil {
			logger.Error("http server error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down AuthService")
	grpcServer.GracefulStop()
	_ = httpSrv.Stop(context.Background())
}
