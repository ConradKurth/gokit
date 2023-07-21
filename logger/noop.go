package logger

import "context"

type dummyLogger struct {
}

func (l *dummyLogger) Info(msg string, fields ...Field) {
}

func (l *dummyLogger) InfoCtx(ctx context.Context, msg string, fields ...Field) {
}

func (l *dummyLogger) Debug(msg string, fields ...Field) {
}

func (l *dummyLogger) DebugCtx(ctx context.Context, msg string, fields ...Field) {
}

func (l *dummyLogger) Error(msg string, fields ...Field) {
}

func (l *dummyLogger) ErrorCtx(ctx context.Context, msg string, fields ...Field) {
}

func (l *dummyLogger) Warn(msg string, fields ...Field) {
}

func (l *dummyLogger) WarnCtx(ctx context.Context, msg string, fields ...Field) {
}

func (l *dummyLogger) Close() error {
	return nil
}

func NewNoop() Logger {
	return &dummyLogger{}
}
