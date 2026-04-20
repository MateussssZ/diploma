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

	"golang.org/x/sync/errgroup"
)

const version = "v0.0.0"

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
	logger.Info(ctx, ">>>>>>>> Version: ", "version", version)

	application, err := app.NewApp(ctx, app.Dep{
		Version:     version,
		Credentials: cred,
		Config:      conf,
		Logger:      logger,
	})
	if err != nil {
		logger.Error(ctx, err, "error", "application initialization failed")
		return
	}

	eg := &errgroup.Group{}
	application.Start(ctx, eg)
	logger.Info(ctx, "Application has been started!")

	go func() {
		<-ctx.Done()
		logger.Info(ctx, "Please wait, services are stopping...Chill around 30 seconds")
	}()

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, err, "error", "application start failed")
	}

	logger.Info(ctx, "Application is stopped correctly. The force will be with you")
}
