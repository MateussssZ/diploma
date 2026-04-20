package interceptors

import (
	"context"
	"encoding/json"
	"fmt"

	"apigateway/internal/pkg/applogger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryRecovery(logger applogger.IAppLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		var (
			resp any
			err  error
		)

		defer func() {
			if r := recover(); r != nil {
				jsonBody, errMarshal := json.MarshalIndent(req, "", "  ")
				if errMarshal != nil {
					jsonBody = []byte("cannot marshal request")
				}

				logger.Error(ctx, fmt.Errorf("gRPC call panicked"),
					"method", info.FullMethod,
					"panic", r,
					"body", string(jsonBody),
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		resp, err = handler(ctx, req)

		return resp, err
	}
}
