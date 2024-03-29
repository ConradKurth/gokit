package instrument

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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

		pemBlock, _ := pem.Decode([]byte(c.GetString("temporal.tls.cert")))

		parseCert, err := x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing cert")
		}

		keyBlock, _ := pem.Decode([]byte(c.GetString("temporal.tls.key")))
		key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing key")
		}

		tlsCert := tls.Certificate{
			Certificate: [][]byte{parseCert.Raw},
			PrivateKey:  key,
			Leaf:        parseCert,
		}

		opts.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{tlsCert},
			},
		}
	}

	opts.Interceptors = append(opts.Interceptors, tracingInterceptor)

	temporalClient, err := client.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("could not dial the client: %w", err)
	}

	return temporalClient, nil
}
