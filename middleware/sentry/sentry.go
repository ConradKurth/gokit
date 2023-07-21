package msentry

import (
	"net/http"
	"time"

	"github.com/ConradKurth/gokit/config"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// NewMiddleware returns a new sentry middleware.
func NewMiddleware(c *config.Config) func(http.Handler) http.Handler {
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         c.GetBoolDefault("sentry.repanic", false),
		Timeout:         time.Second * time.Duration(c.GetInt("sentry.timeoutSecs")),
		WaitForDelivery: c.GetBoolDefault("sentry.waitForDelivery", false),
	})

	return sentryHandler.Handle
}
