package applogger

import (
	"context"
	"fmt"
	"log/slog"
)

type IAppLogger interface {
	Debug(ctx context.Context, msg string, attrs ...any)
	Info(ctx context.Context, msg string, attrs ...any)
	Error(ctx context.Context, err error, attrs ...any)
	ErrorWithEventID(ctx context.Context, err error, attrs ...any) string
}

type AppLogger struct {
	stdout *stdoutWriter
	file   *fileWriter
}

// NewAppLogger creates a logger that writes to stdout (all levels) and
// to "app.log" on disk (errors only).
func NewAppLogger(level slog.Level) *AppLogger {
	fw, err := newFileWriter(level)
	if err != nil {
		slog.Warn(fmt.Sprintf("error creating fileWriter, file logging disabled: %v", err))
	}

	return &AppLogger{
		stdout: newStdoutWriter(level),
		file:   fw,
	}
}

func (l *AppLogger) Debug(ctx context.Context, msg string, attrs ...any) {
	l.stdout.debugContext(ctx, msg, attrs...)
	if l.file != nil {
		l.file.debugContext(ctx, msg, attrs...)
	}
}

func (l *AppLogger) Info(ctx context.Context, msg string, attrs ...any) {
	l.stdout.infoContext(ctx, msg, attrs...)
	if l.file != nil {
		l.file.infoContext(ctx, msg, attrs...)
	}
}

func (l *AppLogger) Error(ctx context.Context, err error, attrs ...any) {
	_ = l.ErrorWithEventID(ctx, err, attrs...)
}

func (l *AppLogger) ErrorWithEventID(ctx context.Context, err error, attrs ...any) string {
	l.stdout.errorContext(ctx, err.Error(), attrs...)
	if l.file != nil {
		l.file.errorContext(ctx, err.Error(), attrs...)
	}
	return ""
}
