package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ConradKurth/gokit/config"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

func (c contextKey) String() string {
	return "ctx-key-" + string(c)
}

const loggerKey = contextKey("logger")

func GetContextKey() contextKey {
	return loggerKey
}

// SetLogger will set a logger on the context
func SetLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func GetLogger(ctx context.Context) Logger {
	v, ok := ctx.Value(GetContextKey()).(Logger)
	if !ok {
		return NewNoop()
	}
	return v
}

func GetLoggerReq(r *http.Request) Logger {
	return GetLogger(r.Context())
}

// TODO remove interface here, use it at the caller where needed and
// return type instead here.
type Logger interface {
	InfoCtx(ctx context.Context, msg string, fields ...Field)
	DebugCtx(ctx context.Context, msg string, fields ...Field)
	ErrorCtx(ctx context.Context, msg string, fields ...Field)
	WarnCtx(ctx context.Context, msg string, fields ...Field)

	Close() error
}

type logger struct {
	logger *otelzap.Logger
}

type logtailWriter struct {
	token string
	data  []string
}

// Write will write data to our buffer
func (l *logtailWriter) Write(p []byte) (int, error) {
	l.data = append(l.data, string(p))
	return len(p), nil
}

func (l *logtailWriter) sendLoggingError(step string, logErr error, body []byte) error {

	type loggingErr struct {
		Error string `json:"error"`
		Step  string `json:"step"`
		Body  string `json:"body"`
	}

	b, err := json.Marshal(loggingErr{Error: logErr.Error(), Step: step, Body: string(body)})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://in.logtail.com", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", l.token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("did not get a valid status code: %v", resp.StatusCode)
	}
	return nil
}

// Sync will sync the data to logtail
func (l *logtailWriter) Sync() error {

	defer func() {
		l.data = nil
	}()

	values := make([]interface{}, 0, len(l.data))
	for _, d := range l.data {
		var value interface{}

		for _, line := range strings.Split(d, "\n") {
			if line == "" {
				continue
			}
			if err := json.Unmarshal([]byte(line), &value); err != nil {
				return l.sendLoggingError("unmarhsalling", err, nil)
			}
			values = append(values, value)
		}
	}

	if len(values) == 0 {
		return nil
	}

	body, err := json.Marshal(values)
	if err != nil {
		return l.sendLoggingError("marhsalling", err, nil)
	}

	req, err := http.NewRequest(http.MethodPost, "https://in.logtail.com", bytes.NewBuffer(body))
	if err != nil {
		return l.sendLoggingError("making request", err, body)
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", l.token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return l.sendLoggingError("do request", err, body)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return l.sendLoggingError("response", fmt.Errorf("did not get a valid status code: %v", resp.StatusCode), body)
	}
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

// NewV2 creates a new logger based on the current environment config and will pipe logs to logtail itself
func NewV2(c *config.Config) Logger {
	var l *zap.Logger
	var err error
	writer := &logtailWriter{
		token: c.GetString("logtail.token"),
	}

	buffer := &zapcore.BufferedWriteSyncer{
		WS: writer,
	}

	if config.IsDevelopment() {

		writers := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(buffer))

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			writers,
			zap.NewAtomicLevelAt(zap.DebugLevel),
		)

		l, err = zap.NewDevelopment(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return core
		}))
	} else {
		writers := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(zapcore.Lock(buffer)))
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			writers,
			zap.InfoLevel,
		)

		l, err = zap.NewProduction(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return core
		}))
	}
	if err != nil {
		panic(err)
	}

	log := &logger{
		logger: otelzap.New(l, otelzap.WithStackTrace(true)),
	}

	// defer log.Close()

	return log
}

// New creates a new logger based on the current environment config.
func New() Logger {
	var l *zap.Logger
	var err error
	if config.IsDevelopment() {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}

	return &logger{
		logger: otelzap.New(l, otelzap.WithStackTrace(true)),
	}
}

func (l *logger) Close() error {
	return l.logger.Sync()
}

// NewWithLogger returns a new logger based on the passed already initialized
// logger instance. Useful for tests where the log output is observed used a
// custom log observer.
func NewWithLogger(log *zap.Logger) Logger {
	return &logger{
		logger: otelzap.New(log),
	}
}

func (l *logger) getFields(ctx context.Context, fields ...Field) []zap.Field {
	z := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		z = append(z, zap.Any(f.Key(), f.Value()))
	}

	if reqID := middleware.GetReqID(ctx); reqID != "" {
		z = append(z, zap.String("local_request_id", reqID))
	}

	return z
}

func (l *logger) InfoCtx(ctx context.Context, msg string, fields ...Field) {
	f := l.getFields(ctx, fields...)
	l.logger.InfoContext(ctx, msg, f...)
}

func (l *logger) DebugCtx(ctx context.Context, msg string, fields ...Field) {
	f := l.getFields(ctx, fields...)
	l.logger.DebugContext(ctx, msg, f...)
}

func (l *logger) ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	sentryItems := map[string]interface{}{
		"message": msg,
	}
	err := errors.New(msg)
	for _, f := range fields {
		if e, ok := f.Value().(error); ok {
			err = e
		} else {
			sentryItems[f.Key()] = f.Value()
		}
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetContext("fields", sentryItems)
		sentry.CaptureException(err)
	})

	f := l.getFields(ctx, fields...)
	l.logger.Error(msg, f...)
}

func (l *logger) WarnCtx(ctx context.Context, msg string, fields ...Field) {
	f := l.getFields(ctx, fields...)
	l.logger.WarnContext(ctx, msg, f...)
}
