package instrument

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
)

// NewTemporalClient will create a new temporal client with interceptors added
func NewTemporalClient(ctx context.Context, l logger.Logger, c *config.Config, serviceName string) (client.Client, error) {

	_, err := GetTracingProvider(ctx, c, serviceName)
	if err != nil {
		return nil, err
	}

	accum, err := GetMetricAccumulator(ctx, c)
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

	opts.MetricsHandler = newMetricHandler(metric.WrapMeterImpl(accum))

	temporalClient, err := client.NewClient(opts)
	if err != nil {
		return nil, err
	}

	return temporalClient, nil
}

var registryLock sync.Mutex
var handlerRegistry = map[string]client.MetricsHandler{}

type metricHandler struct {
	counts   sync.Mutex
	counters map[string]client.MetricsCounter

	gauge  sync.Mutex
	gauges map[string]client.MetricsGauge

	timer  sync.Mutex
	timers map[string]client.MetricsTimer

	m metric.Meter
}

func newMetricHandler(m metric.Meter) *metricHandler {
	return &metricHandler{
		m:        m,
		counters: make(map[string]client.MetricsCounter),
		gauges:   make(map[string]client.MetricsGauge),
		timers:   make(map[string]client.MetricsTimer),
	}
}

func (m *metricHandler) WithTags(tags map[string]string) client.MetricsHandler {
	registryLock.Lock()
	defer registryLock.Unlock()

	key := keFromMap(make([]byte, 0, 256), tags)

	if v, ok := handlerRegistry[string(key)]; ok {
		return v
	}

	h := newMetricHandler(m.m)
	handlerRegistry[string(key)] = h
	return h
}

func (m *metricHandler) Counter(name string) client.MetricsCounter {
	m.counts.Lock()
	defer m.counts.Unlock()

	if v, ok := m.counters[name]; ok {
		return v
	}

	c, err := m.m.NewInt64Counter(name)
	if err != nil {
		panic(err)
	}

	v := &wrappedInt64Counter{
		m: c,
	}
	m.counters[name] = v
	return v
}

func (m *metricHandler) Gauge(name string) client.MetricsGauge {
	m.gauge.Lock()
	defer m.gauge.Unlock()
	if v, ok := m.gauges[name]; ok {
		return v
	}
	c, err := m.m.NewFloat64GaugeObserver(name, nil)
	if err != nil {
		panic(err)
	}
	v := &wrappedFloat64Gauge{
		m: c,
	}
	m.gauges[name] = v
	return v
}

func (m *metricHandler) Timer(name string) client.MetricsTimer {
	m.timer.Lock()
	defer m.timer.Unlock()

	if v, ok := m.timers[name]; ok {
		return v
	}
	c, err := m.m.NewInt64Histogram(name)
	if err != nil {
		panic(err)
	}
	v := &wrappedInt64Histogram{
		m: c,
	}
	m.timers[name] = v

	return v
}

type wrappedInt64Counter struct {
	m metric.Int64Counter
}

func (w *wrappedInt64Counter) Inc(v int64) {
	w.m.Add(context.Background(), v)
}

type wrappedFloat64Gauge struct {
	m metric.Float64GaugeObserver
}

func (w *wrappedFloat64Gauge) Update(v float64) {
	w.m.Observation(v)
}

type wrappedInt64Histogram struct {
	m metric.Int64Histogram
}

func (w *wrappedInt64Histogram) Record(v time.Duration) {
	w.m.Record(context.Background(), int64(v))
}

const (
	keyPairSplitter = ','
	keyNameSplitter = '='
)

// keFromMap will essentially take a map, order the keys and then make a key based of the order items in the map
func keFromMap(buf []byte, maps ...map[string]string) []byte {
	// stack allocated
	keys := make([]string, 0, 32)
	for _, m := range maps {
		for k := range m {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	var lastKey string // last key written to the buffer
	for _, k := range keys {
		if len(lastKey) > 0 {
			if k == lastKey {
				// Already wrote this key.
				continue
			}
			buf = append(buf, keyPairSplitter)
		}
		lastKey = k

		buf = append(buf, k...)
		buf = append(buf, keyNameSplitter)

		// Find and write the value for this key. Rightmost map takes
		// precedence.
		for j := len(maps) - 1; j >= 0; j-- {
			if v, ok := maps[j][k]; ok {
				buf = append(buf, v...)
				break
			}
		}
	}

	return buf
}
