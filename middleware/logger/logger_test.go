package logger

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ConradKurth/gokit/logger"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type testLogger struct {
	logger   logger.Logger
	observed *observer.ObservedLogs
}

func newTestLogger() *testLogger {
	core, observedLogs := observer.New(zap.DebugLevel)
	z := zap.New(core)
	log := logger.NewWithLogger(z)

	return &testLogger{
		logger:   log,
		observed: observedLogs,
	}
}

func (t testLogger) InfoCtx(ctx context.Context, msg string, fields ...logger.Field) {
	t.logger.InfoCtx(ctx, msg, fields...)
}

func (t testLogger) DebugCtx(ctx context.Context, msg string, fields ...logger.Field) {
	t.logger.DebugCtx(ctx, msg, fields...)
}

func (t testLogger) ErrorCtx(ctx context.Context, msg string, fields ...logger.Field) {
	t.logger.ErrorCtx(ctx, msg, fields...)
}

func (t testLogger) WarnCtx(ctx context.Context, msg string, fields ...logger.Field) {
	t.logger.WarnCtx(ctx, msg, fields...)
}

func (t testLogger) Close() error {
	return nil
}

// TestLoggerMiddlewareClientError tests that client errors
// are logged on warning level.
func TestLoggerMiddlewareClientError(t *testing.T) {
	log := newTestLogger()
	r := chi.NewRouter()
	r.Use(NewMiddleware(log))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("test-key", "value")
		w.WriteHeader(http.StatusForbidden)
	})
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, "value", resp.Header.Get("test-key"))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	logs := log.observed.TakeAll()
	require.Equal(t, 2, len(logs))
	assert.Equal(t, zap.DebugLevel, logs[0].Entry.Level) // no userid found
	assert.Equal(t, zap.WarnLevel, logs[1].Entry.Level)
}

// TestLoggerMiddlewareServerError tests that server errors
// are logged on error level.
func TestLoggerMiddlewareServerError(t *testing.T) {
	log := newTestLogger()
	r := chi.NewRouter()
	r.Use(NewMiddleware(log))

	body := []byte("test")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write(body)
		require.NoError(t, err)
	})
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, b)

	logs := log.observed.TakeAll()
	require.Equal(t, 2, len(logs))
	assert.Equal(t, zap.DebugLevel, logs[0].Entry.Level) // no userid found
	assert.Equal(t, zap.ErrorLevel, logs[1].Entry.Level)
}
