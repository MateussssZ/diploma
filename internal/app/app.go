package app

import (
	"apigateway/config"
	"apigateway/internal/clients"
	"apigateway/internal/integrations/kafka"
	"apigateway/internal/integrations/wsmanager"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"
	"fmt"
	"net"

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
	grpc          *Grpc
	rest          *Rest
	wsManager     *wsmanager.WSManager
	kafkaConsumer *kafka.Consumer
}

func NewApp(ctx context.Context, dep Dep) (*App, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewApp", err)
	}

	promMetrics := metrics.NewPrometheus()

	usecases, err := NewUsecases(UsecasesDep{
		AuctionClient: func() *clients.AuctionServiceClient {
			conn, _ := grpc.NewClient(dep.Config.AuctionService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			conn.Connect()
			return clients.NewAuctionServiceClient(conn)
		}(),
	})
	if err != nil {
		return nil, err
	}

	authConn, err := grpc.NewClient(
		"passthrough:///authservice",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", dep.Config.AuthService.SocketPath)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("dial AuthService unix socket: %w", err)
	}
	authClient := clients.NewAuthServiceClient(authConn)

	integrations, err := NewIntegrations(IntegrationsDep{
		AuctionActions: usecases.Auction,
		Metrics:        promMetrics,
		Logger:         dep.Logger,
		KafkaCfg:       dep.Config.Kafka,
	})
	if err != nil {
		return nil, err
	}

	controllers, err := NewControllers(ControllersDep{
		Usecases:   usecases,
		AuthClient: authClient,
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
		WSManager:   integrations.WSManager,
	})
	if err != nil {
		return nil, err
	}

	return &App{
		grpc:          grpcSrv,
		rest:          restSrv,
		wsManager:     integrations.WSManager,
		kafkaConsumer: integrations.KafkaConsumer,
	}, nil
}

func (a *App) Start(ctx context.Context, eg *errgroup.Group) {
	eg.Go(func() error {
		return a.grpc.Start(ctx)
	})

	eg.Go(func() error {
		return a.rest.Start(ctx)
	})

	eg.Go(func() error {
		a.wsManager.Start(ctx)
		return nil
	})

	eg.Go(func() error {
		return a.kafkaConsumer.Start(ctx)
	})
}
