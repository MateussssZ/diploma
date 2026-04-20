package applogger

import (
	"context"
	"log/slog"
	"os"

	"apigateway/internal/utils"
)

type stdoutWriter struct {
	logger *slog.Logger
}

func newStdoutWriter(level slog.Level) *stdoutWriter {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return &stdoutWriter{logger: slog.New(handler)}
}

func (w *stdoutWriter) debugContext(ctx context.Context, msg string, attrs ...any) {
	attrs = append(attrs, utils.ExtractAttrs(ctx)...)
	w.logger.DebugContext(ctx, msg, attrs...)
}

func (w *stdoutWriter) infoContext(ctx context.Context, msg string, attrs ...any) {
	attrs = append(attrs, utils.ExtractAttrs(ctx)...)
	w.logger.InfoContext(ctx, msg, attrs...)
}

func (w *stdoutWriter) errorContext(ctx context.Context, msg string, attrs ...any) {
	attrs = append(attrs, utils.ExtractAttrs(ctx)...)
	w.logger.ErrorContext(ctx, msg, attrs...)
}
