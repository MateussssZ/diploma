package main

import (
	"apigateway/config"
	"apigateway/internal/app"
	"apigateway/internal/pkg/applogger"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const version = "v1.0.0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf, err := config.NewConfig()
	if err != nil {
		slog.Error(fmt.Sprintf("configurations initialization failed - error: %v", err))
		return
	}

	cred, err := config.NewCredentials()
	if err != nil {
		slog.Error(fmt.Sprintf("credentials initialization failed - error: %v", err))
		return
	}

	logger := applogger.NewAppLogger(conf.Logger.Level)
	logger.Info(ctx, "starting", "version", version)

	application, err := app.NewApp(ctx, app.Dep{
		Version:     version,
		Credentials: cred,
		Config:      conf,
		Logger:      logger,
	})
	if err != nil {
		logger.Error(ctx, err)
		return
	}

	eg := &errgroup.Group{}
	application.Start(ctx, eg)
	logger.Info(ctx, "application started")

	go func() {
		<-ctx.Done()
		logger.Info(ctx, "shutting down")
	}()

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, err)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	application.Stop(stopCtx)

	logger.Info(ctx, "application stopped")
}
