package app

import (
	"apigateway/config"
	"apigateway/internal/clients"
	"apigateway/internal/integrations/cache"
	"apigateway/internal/integrations/kafka"
	"apigateway/internal/integrations/wsmanager"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
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
	cacheManager  *cache.CacheManager
	auctionConn   *grpc.ClientConn
	authConn      *grpc.ClientConn
	logger        applogger.IAppLogger
}

func NewApp(ctx context.Context, dep Dep) (*App, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewApp", err)
	}

	promMetrics := metrics.NewPrometheus()

	redisClient, err := cache.NewRedisClient(
		dep.Config.Redis.Address,
		dep.Config.Redis.DB,
		dep.Config.Redis.Password,
		dep.Config.Redis.MaxRetries,
		dep.Config.Redis.PoolSize,
		dep.Config.Redis.DialTimeout,
		dep.Config.Redis.ReadTimeout,
		dep.Config.Redis.WriteTimeout,
	)
	if err != nil {
		dep.Logger.Error(context.Background(),
			fmt.Errorf("failed to connect to Redis, cache disabled: %w", err))
		// Cache is optional - continue without it
		redisClient = nil
	}

	cacheManager := cache.NewCacheManager(
		redisClient,
		dep.Logger,
		promMetrics,
		cache.CacheTTLs{
			AuctionDetailTTL: dep.Config.CacheConfig.AuctionDetailTTL,
		},
	)

	auctionConn, err := grpc.NewClient(
		dep.Config.AuctionService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             1 * time.Second,
			PermitWithoutStream: true,
		}),
		// Built-in gRPC retry: up to 3 attempts with exponential backoff for
		// transient errors (Unavailable, ResourceExhausted).
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{}],
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.05s",
					"maxBackoff": "1s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
				},
				"timeout": "5s"
			}]
		}`),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AuctionService at %s: %w", dep.Config.AuctionService.Address, err)
	}
	auctionConn.Connect()

	usecases, err := NewUsecases(UsecasesDep{
		AuctionClient: clients.NewAuctionServiceClient(auctionConn),
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
		WSCfg:          dep.Config.WS,
		CacheManager:   cacheManager,
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
		Version:      dep.Version,
		Controllers:  controllers,
		Config:       &dep.Config.RESTServer,
		Logger:       dep.Logger,
		Metrics:      promMetrics,
		JWTSecret:    dep.Credentials.Auth.JWTSecret,
		WSManager:    integrations.WSManager,
		CacheManager: integrations.CacheManager,
	})
	if err != nil {
		return nil, err
	}

	return &App{
		grpc:          grpcSrv,
		rest:          restSrv,
		wsManager:     integrations.WSManager,
		kafkaConsumer: integrations.KafkaConsumer,
		cacheManager:  cacheManager,
		auctionConn:   auctionConn,
		authConn:      authConn,
		logger:        dep.Logger,
	}, nil
}

// Stop shuts down all resources in reverse order of creation.
// Must be called after Start's errgroup has returned.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info(ctx, "stopping application components")

	// 1. Close all active WebSocket connections gracefully.
	//    Uses a fresh timeout so it isn't affected by the already-cancelled signal ctx.
	wsCtx, wsCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer wsCancel()
	a.wsManager.Shutdown(wsCtx)

	// 2. Close outbound gRPC client connections.
	if a.auctionConn != nil {
		if err := a.auctionConn.Close(); err != nil {
			a.logger.Error(ctx, fmt.Errorf("error closing AuctionService gRPC conn: %w", err))
		}
	}
	if a.authConn != nil {
		if err := a.authConn.Close(); err != nil {
			a.logger.Error(ctx, fmt.Errorf("error closing AuthService gRPC conn: %w", err))
		}
	}

	// 3. Close Redis connection pool.
	if a.cacheManager != nil {
		if err := a.cacheManager.Close(); err != nil {
			a.logger.Error(ctx, fmt.Errorf("error closing cache manager: %w", err))
		}
	}

	a.logger.Info(ctx, "all components stopped")
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
