package instrument

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/ConradKurth/gokit/config"
	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
)

// NewTemporalClient will create a new temporal client with interceptors added
func NewTemporalClient(ctx context.Context, c *config.Config, serviceName string) (client.Client, error) {

	_, err := GetTracingProvider(ctx, c, serviceName)
	if err != nil {
		return nil, err
	}

	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{Tracer: otel.Tracer("temporal")})
	if err != nil {
		return nil, err
	}

	opts := client.Options{
		HostPort:  c.GetString("temporal.hostPort"),
		Namespace: c.GetString("temporal.namespace"),
	}

	if c.GetBool("temporal.tls.enabled") {
		cert, err := tls.X509KeyPair(
			[]byte(c.GetString("temporal.tls.ca")),
			[]byte(c.GetString("temporal.tls.key")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed loading client cert and key: %w", err)
		}

		opts.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
	}

	opts.Interceptors = append(opts.Interceptors, tracingInterceptor)

	temporalClient, err := client.Dial(opts)
	if err != nil {
		return nil, err
	}

	return temporalClient, nil
}
