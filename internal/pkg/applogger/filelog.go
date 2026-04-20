package applogger

import (
	"context"
	"log/slog"
	"os"

	"apigateway/internal/utils"
)

const _logFilePath = "app.log"

type fileWriter struct {
	logger *slog.Logger
}

func newFileWriter(level slog.Level) (*fileWriter, error) {
	f, err := os.OpenFile(_logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{Level: level})
	return &fileWriter{logger: slog.New(handler)}, nil
}

func (w *fileWriter) debugContext(ctx context.Context, msg string, attrs ...any) {
	attrs = append(attrs, utils.ExtractAttrs(ctx)...)
	w.logger.DebugContext(ctx, msg, attrs...)
}

func (w *fileWriter) infoContext(ctx context.Context, msg string, attrs ...any) {
	attrs = append(attrs, utils.ExtractAttrs(ctx)...)
	w.logger.InfoContext(ctx, msg, attrs...)
}

func (w *fileWriter) errorContext(ctx context.Context, msg string, attrs ...any) {
	attrs = append(attrs, utils.ExtractAttrs(ctx)...)
	w.logger.ErrorContext(ctx, msg, attrs...)
}
