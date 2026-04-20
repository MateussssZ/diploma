package app

import (
	"apigateway/config"
	"apigateway/internal/api/rest"
	"apigateway/internal/api/rest/handlers"
	"apigateway/internal/api/rest/middlewares"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"
)

type RestDep struct {
	Version     string               `validate:"required,min=6"`
	Controllers *Controllers         `validate:"required"`
	Config      *config.RESTServer   `validate:"required"`
	Logger      applogger.IAppLogger `validate:"required"`
	Metrics     metrics.IMetrics     `validate:"required"`
	JWTSecret   string               `validate:"required"`
}

type Rest struct {
	restServer *rest.Server
}

func NewRest(_ context.Context, dep RestDep) (*Rest, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewRest", err)
	}

	// инициализация универсального обработчика ответов клиенту
	responder, err := handlers.NewResponder(handlers.ResponderDep{
		Logger:  dep.Logger,
		Metrics: dep.Metrics,
	})
	if err != nil {
		return nil, err
	}

	// инициализация всех middlewares
	requestID, err := middlewares.NewRequestIDMiddleware()
	if err != nil {
		return nil, err
	}

	recovery, err := middlewares.NewRecoveryMiddleware(middlewares.RecoveryMiddlewareDep{
		Responder: responder,
	})
	if err != nil {
		return nil, err
	}

	metrics, err := middlewares.NewMetricsMiddleware(middlewares.MetricsMiddlewareDep{
		Logger:  dep.Logger,
		Metrics: dep.Metrics,
	})
	if err != nil {
		return nil, err
	}

	auth, err := middlewares.NewAuthMiddleware(dep.JWTSecret)
	if err != nil {
		return nil, err
	}

	// инициализация всех handlers
	bases, err := handlers.NewBaseHandlers(handlers.Dep{
		AppVersion: dep.Version,
		Logger:     dep.Logger,
	})
	if err != nil {
		return nil, err
	}

	userHandlers, err := handlers.NewUserHandlers(handlers.UserHandlersDep{
		Responder: responder,
		UserCtrl:  dep.Controllers.User,
	})
	if err != nil {
		return nil, err
	}

	auctionHandlers, err := handlers.NewAuctionHandlers(handlers.AuctionHandlersDep{
		Responder:   responder,
		AuctionCtrl: dep.Controllers.Auction,
	})
	if err != nil {
		return nil, err
	}

	restRoutes, err := rest.NewRoute(rest.RouteDep{
		Metrics:             dep.Metrics,
		RequestIDMiddleware: requestID.Middleware,
		RecoveryMiddleware:  recovery.Middleware,
		MetricsMiddleware:   metrics.Middleware,
		AuthMiddleware:      auth.Middleware,
		BaseHandlers:        bases,
		UserHandlers:        userHandlers,
		AuctionHandlers:     auctionHandlers,
	})
	if err != nil {
		return nil, err
	}

	restServer, err := rest.NewServer(rest.Options{
		ListenPort:        dep.Config.ListenPort,
		ReadTimeout:       dep.Config.ReadTimeout,
		WriteTimeout:      dep.Config.WriteTimeout,
		ShutdownTimeout:   dep.Config.ShutdownTimeout,
		ReadHeaderTimeout: dep.Config.ReadHeaderTimeout,
		IdleTimeout:       dep.Config.IdleTimeout,
		MaxHeaderBytes:    dep.Config.MaxHeaderBytes,
	}, rest.Dependencies{
		Handler: restRoutes,
	})
	if err != nil {
		return nil, err
	}

	return &Rest{
		restServer: restServer,
	}, nil
}

func (r *Rest) Start(ctx context.Context) error {
	return r.restServer.Start(ctx)
}
