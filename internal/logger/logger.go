package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Logger = *logrus.Entry

type ctxKey string

const loggerCtxKey ctxKey = "logger"

func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

func FromContext(ctx context.Context) Logger {
	log, found := ctx.Value(loggerCtxKey).(Logger)
	if !found {
		return logrus.NewEntry(logrus.New())
	}

	return log
}
