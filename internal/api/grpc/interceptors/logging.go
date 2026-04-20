package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"apigateway/internal/pkg/applogger"
	"google.golang.org/grpc"
	"time"
)

func UnaryLogging(logger applogger.IAppLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		logger.Debug(ctx, "gRPC call started")

		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		if err != nil {
			jsonBody, errMarshal := json.MarshalIndent(req, "", "  ")
			if errMarshal != nil {
				jsonBody = []byte("cannot marshal request")
			}

			logger.Error(ctx, fmt.Errorf("gRPC call failed: %w", err),
				"duration", duration,
				"body", string(jsonBody),
			)

			return nil, err
		}

		logger.Debug(ctx, "gRPC call succeeded",
			"duration", duration)

		return resp, err
	}
}
