package grpc

import (
	"context"
	"errors"
	"net"

	"apigateway/config"
	"apigateway/internal/api/grpc/interceptors"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ServerDep struct {
	Config  *config.GRPCServer   `validate:"required"`
	Logger  applogger.IAppLogger `validate:"required"`
	Metrics metrics.IMetrics     `validate:"required"`
}

type Server struct {
	srv      *grpc.Server
	listener net.Listener
	logger   applogger.IAppLogger
}

func NewServer(ctx context.Context, dep ServerDep) (*Server, error) {
	if dep.Config == nil {
		return nil, errorspkg.NewValidationError("NewGRPCServer", errors.New("Config is required"))
	}
	if dep.Logger == nil {
		return nil, errorspkg.NewValidationError("NewGRPCServer", errors.New("Logger is required"))
	}
	if dep.Metrics == nil {
		return nil, errorspkg.NewValidationError("NewGRPCServer", errors.New("Metrics is required"))
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(dep.Config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(dep.Config.MaxSendMsgSize),
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryRequestID(),
			interceptors.UnaryRecovery(dep.Logger),
			interceptors.UnaryLogging(dep.Logger),
			interceptors.UnaryMetrics(dep.Metrics, dep.Logger),
		),
	}
	srv := grpc.NewServer(opts...)
	reflection.Register(srv)

	lc := &net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", dep.Config.Address)
	if err != nil {
		// обернуть ошибку
		return nil, err
	}
	lis.Addr()

	return &Server{
		srv:      srv,
		listener: lis,
		logger:   dep.Logger,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error)

	go func() {
		s.logger.Info(ctx, "Listening: ", "address", s.listener.Addr().String())
		if err := s.srv.Serve(s.listener); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop(ctx)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.srv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.srv.Stop()
		return ctx.Err()
	}
}
