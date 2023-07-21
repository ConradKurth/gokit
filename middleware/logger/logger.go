package logger

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ConradKurth/gokit/logger"
	"github.com/ConradKurth/gokit/middleware/userinfo"
	"github.com/go-chi/chi/v5"
)

// Logger defines the interface functions that need to be implemented by
// the passed logger for the middleware.
type Logger interface {
	InfoCtx(ctx context.Context, msg string, fields ...logger.Field)
	DebugCtx(ctx context.Context, msg string, fields ...logger.Field)
	ErrorCtx(ctx context.Context, msg string, fields ...logger.Field)
	WarnCtx(ctx context.Context, msg string, fields ...logger.Field)

	Close() error
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// NewMiddleware returns a new logger middleware that logs
// the result of a processed HTTP request.
func NewMiddleware(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := logger.SetLogger(r.Context(), log)
			r = r.WithContext(ctx)

			rec := &statusRecorder{
				ResponseWriter: w,
			}

			next.ServeHTTP(rec, r)

			userId, err := userinfo.GetUserID(r.Context())
			if err != nil {
				log.DebugCtx(ctx, "No userID found")
			}

			rctx := chi.RouteContext(r.Context())
			routePattern := strings.Join(rctx.RoutePatterns, "")

			fields := []logger.Field{
				logger.Any("type", "request"),
				logger.Any("status", rec.statusCode),
				logger.Any("method", r.Method),
				logger.Any("host", r.Host),
				logger.Any("path", r.URL.Path),
				logger.Any("target", routePattern),
				logger.Any("query", r.URL.RawQuery),
				logger.Any("ip", r.RemoteAddr),
				logger.Any("duration_ms", time.Since(start).Milliseconds()),
				logger.Any("user_agent", r.UserAgent()),
				logger.Any("referer", r.Referer()),
				logger.Any("user_id", userId),
			}

			switch {
			case rec.statusCode >= http.StatusBadRequest && rec.statusCode < http.StatusInternalServerError:
				// client error
				log.WarnCtx(ctx, "", fields...)
			case rec.statusCode >= http.StatusInternalServerError:
				// server error
				log.ErrorCtx(ctx, "", fields...)
			default:
				log.InfoCtx(ctx, "", fields...)
			}
		})
	}
}
