package logger

import (
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

type LoggerAdapter struct {
	zl *otelzap.Logger
}

func NewLoggerAdapter(zapLogger Logger) *LoggerAdapter {
	return &LoggerAdapter{
		// Skip one call frame to exclude zap_adapter itself.
		// Or it can be configured when logger is created (not always possible).
		zl: zapLogger.(*logger).logger.WithOptions(zap.AddCallerSkip(1)),
	}
}

func (log *LoggerAdapter) fields(keyvals []interface{}) []zap.Field {
	if len(keyvals)%2 != 0 {
		return []zap.Field{zap.Error(fmt.Errorf("odd number of keyvals pairs: %v", keyvals))}
	}

	var fields []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	return fields
}

func (log *LoggerAdapter) Debug(msg string, keyvals ...interface{}) {
	log.zl.Debug(msg, log.fields(keyvals)...)
}

func (log *LoggerAdapter) Info(msg string, keyvals ...interface{}) {
	log.zl.Info(msg, log.fields(keyvals)...)
}

func (log *LoggerAdapter) Warn(msg string, keyvals ...interface{}) {
	log.zl.Warn(msg, log.fields(keyvals)...)
}

func (log *LoggerAdapter) Error(msg string, keyvals ...interface{}) {
	log.zl.Error(msg, log.fields(keyvals)...)
}

func (log *LoggerAdapter) With(keyvals ...interface{}) log.Logger {
	newLogger := otelzap.New(log.zl.With(log.fields(keyvals)...), otelzap.WithStackTrace(true))
	return &LoggerAdapter{zl: newLogger}
}
