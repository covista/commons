package logging

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logkey struct{}
var logger *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	_logger, err := config.Build()
	if err != nil {
		log.Fatal("Could not build logger", err)
		logger = zap.NewNop().Sugar()
	} else {
		logger = _logger.Named("commons").Sugar()
	}
}

func NewContextWithLogger() context.Context {
	return WithLogger(context.Background())
}

func WithLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, logkey, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(logkey).(*zap.SugaredLogger); ok {
		return logger
	}
	log.Fatal("bad")
	return zap.L().Sugar()
}
