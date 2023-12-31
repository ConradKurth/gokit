package instrument

import (
	"context"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/logger"
	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
)

// NewTemporalClient will create a new temporal client with interceptors added
func NewTemporalClient(ctx context.Context, l logger.Logger, c *config.Config, serviceName string) (client.Client, error) {

	_, err := GetTracingProvider(ctx, c, serviceName)
	if err != nil {
		return nil, err
	}

	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{Tracer: otel.Tracer("temporal")})
	if err != nil {
		return nil, err
	}

	opts := client.Options{
		HostPort: c.GetString("temporal.hostPort"),
		Logger:   logger.NewLoggerAdapter(l),
	}

	opts.Interceptors = append(opts.Interceptors, tracingInterceptor)

	temporalClient, err := client.Dial(opts)
	if err != nil {
		return nil, err
	}

	return temporalClient, nil
}
