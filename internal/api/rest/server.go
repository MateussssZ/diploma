package rest

import (
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"context"
	"fmt"
	"net"
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
	listenAddr      string
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
			// Addr is intentionally left empty — we bind via explicit net.Listener
			// in Start() so we can control TCP keep-alive and SO_REUSEADDR tuning.
			Handler:           dep.Handler,
			ReadTimeout:       opt.ReadTimeout,
			ReadHeaderTimeout: opt.ReadHeaderTimeout,
			WriteTimeout:      opt.WriteTimeout,
			IdleTimeout:       opt.IdleTimeout,
			MaxHeaderBytes:    opt.MaxHeaderBytes,
		},
		listenAddr:      fmt.Sprintf(":%s", opt.ListenPort),
		shutdownTimeout: opt.ShutdownTimeout,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	lc := net.ListenConfig{
		KeepAlive: 60 * time.Second, // detect dead connections faster than default 15m
	}
	ln, err := lc.Listen(ctx, "tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("rest server listen %s: %w", s.listenAddr, err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.srvHTTP.Serve(ln)
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
