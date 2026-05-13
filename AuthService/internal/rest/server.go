package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Options struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	srv  *http.Server
	log  *slog.Logger
	opts Options
}

func NewServer(opts Options, handler http.Handler, log *slog.Logger) *Server {
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 15 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 15 * time.Second
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = 10 * time.Second
	}
	return &Server{
		log:  log,
		opts: opts,
		srv: &http.Server{
			Addr:         opts.Address,
			Handler:      handler,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		},
	}
}

func (s *Server) Start(_ context.Context) error {
	s.log.Info("http server starting", "address", s.opts.Address)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	shutCtx, cancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer cancel()
	s.log.Info("http server shutting down")
	return s.srv.Shutdown(shutCtx)
}
