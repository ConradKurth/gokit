package instrument

import (
	"context"

	"github.com/ConradKurth/gokit/config"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// var traceOnce sync.Once

var traceProvider *sdktrace.TracerProvider

// GetTracingProvider will create a new trading provider
func GetTracingProvider(ctx context.Context, c *config.Config, serviceName string) (*sdktrace.TracerProvider, error) {
	// var oErr error
	// traceOnce.Do(func() {
	// 	exp, err := getTraceExporter(ctx, c)
	// 	if err != nil {
	// 		oErr = err
	// 		return
	// 	}

	// 	// Create a new tracer provider with a batch span processor and the otlp exporter.
	// 	traceProvider = newTraceProvider(exp, serviceName)

	// 	// Set the Tracer Provider and the W3C Trace Context propagator as globals
	// 	otel.SetTracerProvider(traceProvider)

	// 	// Register the trace context and baggage propagators so data is propagated across services/processes.
	// 	otel.SetTextMapPropagator(
	// 		propagation.NewCompositeTextMapPropagator(
	// 			propagation.TraceContext{},
	// 			propagation.Baggage{},
	// 		),
	// 	)
	// })

	// return traceProvider, oErr
	return traceProvider, nil
}

// func getTraceExporter(ctx context.Context, c *config.Config) (*otlptrace.Exporter, error) {
// 	// opts := []otlptracegrpc.Option{
// 	// 	otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
// 	// }
// 	// return otlptracegrpc.New(ctx, opts...)
// 	return nil, nil
// }

// func newTraceProvider(exp *otlptrace.Exporter, serviceName string) *sdktrace.TracerProvider {
// 	return nil
// 	// The service.name attribute is required.
// 	// sampleRate := 0.1
// 	// resource :=
// 	// 	resource.NewWithAttributes(
// 	// 		semconv.SchemaURL,
// 	// 		semconv.ServiceNameKey.String(serviceName),
// 	// 		attribute.Key("SampleRate").Float64(sampleRate), // additional resource attribute
// 	// 	)

// 	// return sdktrace.NewTracerProvider(
// 	// 	sdktrace.WithBatcher(exp),
// 	// 	sdktrace.WithResource(resource),
// 	// 	sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))), // sampler
// 	// )
// }
