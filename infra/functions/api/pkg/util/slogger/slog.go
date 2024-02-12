package slogger

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/o11y"
)

type Logger struct {
	*slog.Logger
}

var (
	logger *Logger
)

func Init(isProd bool) *Logger {
	var once sync.Once
	once.Do(func() {
		if isProd {
			logger = &Logger{slog.New(slog.NewJSONHandler(os.Stdout, nil))}
		} else {
			logger = &Logger{slog.New(slog.NewTextHandler(os.Stdout, nil))}
		}
	})
	logger.Info("Logger initialized", "isProd", isProd)

	return logger
}

func (l *Logger) WithArgs(args ...any) *Logger {
	return &Logger{
		l.With(args...),
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	traceID := o11y.GetTraceID(ctx)
	return l.WithArgs(
		"traceId", traceID,
	)
}

func New() *Logger {
	return logger
}
