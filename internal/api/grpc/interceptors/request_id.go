package interceptors

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"apigateway/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryRequestID() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		var requestID string
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if vals := md.Get(string(utils.CtxRequestID)); len(vals) > 0 {
				requestID = vals[0]
			}
		}

		if requestID == "" {
			requestID = uuid.New().String()
		}

		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("grpc_method", info.FullMethod)
		})

		ctx = sentry.SetHubOnContext(ctx, hub)
		ctx = context.WithValue(ctx, utils.CtxRequestID, requestID)
		ctx = context.WithValue(ctx, utils.CtxRequestMethod, info.FullMethod)

		return handler(ctx, req)
	}
}
