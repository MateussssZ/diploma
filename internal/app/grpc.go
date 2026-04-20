package app

import (
	"apigateway/config"
	"apigateway/internal/api/grpc"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"
)

type GrpcDep struct {
	Config  *config.GRPCServer   `validate:"required"`
	Logger  applogger.IAppLogger `validate:"required"`
	Metrics metrics.IMetrics     `validate:"required"`
}

type Grpc struct {
	grpcServer *grpc.Server
}

func NewGrpc(ctx context.Context, dep GrpcDep) (*Grpc, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewGrpc", err)
	}

	_, err := grpc.NewApigatewayService()
	if err != nil {
		return nil, err
	}

	grpcServer, err := grpc.NewServer(ctx, grpc.ServerDep{
		Config:  dep.Config,
		Logger:  dep.Logger,
		Metrics: dep.Metrics,
	})
	if err != nil {
		return nil, err
	}

	return &Grpc{grpcServer: grpcServer}, nil
}

func (g *Grpc) Start(ctx context.Context) error {
	return g.grpcServer.Start(ctx)
}
