package logging

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	loggerKey    contextKey = "logger"
)

var globalLogger *zap.Logger

// InitLogger initializes the global logger
func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	globalLogger = logger
}

// GetLogger returns a logger from context or global logger
func GetLogger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return globalLogger
	}
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return globalLogger
}

// WithRequestID adds request ID to logger
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := GetLogger(ctx).With(zap.String("request_id", requestID))
	return context.WithValue(ctx, loggerKey, logger)
}

// LogRequest logs HTTP request details
func LogRequest(ctx context.Context, method, path string, duration time.Duration, statusCode int) {
	GetLogger(ctx).Info("http_request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	)
}

// LogError logs error with context
func LogError(ctx context.Context, msg string, err error, fields ...zapcore.Field) {
	logger := GetLogger(ctx)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	logger.Error(msg, fields...)
}

// LogOperation logs operation metrics
func LogOperation(ctx context.Context, operation string, duration time.Duration, success bool) {
	GetLogger(ctx).Info("operation_metrics",
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
	)
}
