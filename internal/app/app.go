package app

import (
	"apigateway/config"
	"apigateway/internal/clients"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Dep struct {
	Version     string               `validate:"required,min=6"`
	Credentials *config.Credentials  `validate:"required"`
	Config      *config.Config       `validate:"required"`
	Logger      applogger.IAppLogger `validate:"required"`
}

type App struct {
	grpc *Grpc
	rest *Rest
}

func NewApp(ctx context.Context, dep Dep) (*App, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewApp", err)
	}

	promMetrics := metrics.NewPrometheus()

	registries, err := NewRepo(ctx, RepoDep{
		PostgresCfg: dep.Credentials.Postgres,
	})
	if err != nil {
		return nil, err
	}

	/* // Интерфейсы интеграций, например с s3
	integrations, err := NewIntegrations(IntegrationsDep{})
	if err != nil {
		return nil, err
	}
	*/

	usecases, err := NewUsecases(UsecasesDep{
		Repo: registries,
		AuctionClient: func() *clients.AuctionServiceClient {
			conn, _ := grpc.NewClient(dep.Config.AuctionService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			return clients.NewAuctionServiceClient(conn)
		}(),
		JWTSecret: dep.Credentials.Auth.JWTSecret,
	})
	if err != nil {
		return nil, err
	}

	controllers, err := NewControllers(ControllersDep{
		Usecases: usecases,
	})
	if err != nil {
		return nil, err
	}

	grpcSrv, err := NewGrpc(ctx, GrpcDep{
		Config:  &dep.Config.GRPCServer,
		Logger:  dep.Logger,
		Metrics: promMetrics,
	})
	if err != nil {
		return nil, err
	}

	restSrv, err := NewRest(ctx, RestDep{
		Version:     dep.Version,
		Controllers: controllers,
		Config:      &dep.Config.RESTServer,
		Logger:      dep.Logger,
		Metrics:     promMetrics,
		JWTSecret:   dep.Credentials.Auth.JWTSecret,
	})
	if err != nil {
		return nil, err
	}

	return &App{
		grpc: grpcSrv,
		rest: restSrv,
	}, nil
}

func (a *App) Start(ctx context.Context, eg *errgroup.Group) {
	eg.Go(func() error {
		return a.grpc.Start(ctx)
	})

	eg.Go(func() error {
		return a.rest.Start(ctx)
	})
}
