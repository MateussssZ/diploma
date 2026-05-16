package rest

import (
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"
	"fmt"
	"net/http"
	"time"
)

type Dependencies struct {
	Handler http.Handler `validate:"required"`
}

type Options struct {
	ListenPort        string        `validate:"required,min=1"`
	ReadTimeout       time.Duration `validate:"gt=0"`
	WriteTimeout      time.Duration `validate:"gt=0"`
	ShutdownTimeout   time.Duration `validate:"gt=0"`
	ReadHeaderTimeout time.Duration `validate:"gt=0"`
	IdleTimeout       time.Duration `validate:"gt=0"`
	MaxHeaderBytes    int           `validate:"gt=0"`
}

type Server struct {
	srvHTTP         *http.Server
	shutdownTimeout time.Duration
}

func NewServer(opt Options, dep Dependencies) (*Server, error) {
	if err := validate.Struct(opt); err != nil {
		return nil, errorspkg.NewValidationError("NewServer", err)
	}
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewServer", err)
	}

	return &Server{
		srvHTTP: &http.Server{
			Addr:              fmt.Sprintf(":%s", opt.ListenPort),
			Handler:           dep.Handler,
			ReadTimeout:       opt.ReadTimeout,
			ReadHeaderTimeout: opt.ReadHeaderTimeout,
			WriteTimeout:      opt.WriteTimeout,
			IdleTimeout:       opt.IdleTimeout,
			MaxHeaderBytes:    opt.MaxHeaderBytes,
		},
		shutdownTimeout: opt.ShutdownTimeout,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error)

	go func() {
		errChan <- s.srvHTTP.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop(ctx)
	}
}

func (s *Server) Stop(_ context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	return s.srvHTTP.Shutdown(ctx)
}
