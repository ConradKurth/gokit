package instrument

import (
	"context"
	"fmt"
	"sync"

	"github.com/ConradKurth/gokit/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc/credentials"
)

var metricOnce sync.Once
var traceOnce sync.Once

var metricAccum *sdkmetric.Accumulator
var traceProvider *sdktrace.TracerProvider

// GetMetricAccumulator will create a otlp accumulator which can be used to gather metrics
func GetMetricAccumulator(ctx context.Context, c *config.Config) (*sdkmetric.Accumulator, error) {
	var oErr error
	metricOnce.Do(func() {
		exp, err := newMetricExporter(ctx, c)
		if err != nil {
			oErr = err
			return
		}
		a := basic.New(simple.NewWithHistogramDistribution(), exp)
		metricAccum = metric.NewAccumulator(a)
	})
	return metricAccum, oErr
}

// GetTracingProvider will create a new trading provider
func GetTracingProvider(ctx context.Context, c *config.Config, serviceName string) (*sdktrace.TracerProvider, error) {
	var oErr error
	traceOnce.Do(func() {
		exp, err := getTraceExporter(ctx, c)
		if err != nil {
			oErr = err
			return
		}

		// Create a new tracer provider with a batch span processor and the otlp exporter.
		traceProvider = newTraceProvider(exp, serviceName)

		// Set the Tracer Provider and the W3C Trace Context propagator as globals
		otel.SetTracerProvider(traceProvider)

		// Register the trace context and baggage propagators so data is propagated across services/processes.
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)
	})

	return traceProvider, oErr
}

func newMetricExporter(ctx context.Context, c *config.Config) (*otlpmetric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint("api.honeycomb.io:443"),
		otlpmetricgrpc.WithHeaders(map[string]string{
			"x-honeycomb-team":    c.GetString("honeycomb.writeKey"),
			"x-honeycomb-dataset": fmt.Sprintf("%s-%s", config.GetEnvironment().String(), c.GetString("honeycomb.metricDataSet")),
		}),
		otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	}
	return otlpmetricgrpc.New(ctx, opts...)
}

func getTraceExporter(ctx context.Context, c *config.Config) (*otlptrace.Exporter, error) {

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint("api.honeycomb.io:443"),
		otlptracegrpc.WithHeaders(map[string]string{
			"x-honeycomb-team":    c.GetString("honeycomb.writeKey"),
			"x-honeycomb-dataset": fmt.Sprintf("%s-%s", config.GetEnvironment().String(), c.GetString("honeycomb.traceDataSet")),
		}),
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	}

	return otlptracegrpc.New(ctx, opts...)
}

func newTraceProvider(exp *otlptrace.Exporter, serviceName string) *sdktrace.TracerProvider {
	// The service.name attribute is required.
	sampleRate := 0.1
	resource :=
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.Key("SampleRate").Float64(sampleRate), // additional resource attribute
		)

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))), // sampler
	)
}
