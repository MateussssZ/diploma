package interceptors

import (
	"context"
	"fmt"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"regexp"
	"time"
)

func UnaryMetrics(m metrics.IMetrics, logger applogger.IAppLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		start := time.Now()

		m.OperationsTotalInc()

		m.GRPCServerRequestsInFlightInc()
		defer m.GRPCServerRequestsInFlightDec()

		resp, handlerErr := handler(ctx, req)
		if handlerErr != nil {
			m.OperationsErrorsTotalInc()
		}

		service, method, err := extractServiceMethodRegex(info.FullMethod)
		if err != nil {
			logger.Error(ctx, fmt.Errorf("UnaryMetricsInterceptor.extractServiceMethodRegex - error: %w", err))
		}

		statusCode := status.Code(handlerErr).String()
		m.GRPCServerRequestsTotalInc(service, method, statusCode)

		duration := time.Since(start).Seconds()
		m.GRPCServerRequestDurationInc(service, method, duration)

		return resp, handlerErr
	}
}

func extractServiceMethodRegex(path string) (service, method string, err error) {
	// Регулярное выражение для извлечения сервиса и метода
	// Матчится на: /package.Service/Method
	re := regexp.MustCompile(`^/(?:[^/]+\.)?([^./]+)/([^/]+)$`)

	matches := re.FindStringSubmatch(path)
	if matches == nil {
		return "", "", fmt.Errorf("invalid format: %s", path)
	}

	return matches[1], matches[2], nil
}
