package metrics //nolint:revive,nolintlint // we can't change the package name for backwards compatibility

import (
	"context"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

var (
	_ Handler        = (*otelHandler)(nil)
	_ Int64Counter   = (*otelInt64Counter)(nil)
	_ Int64Gauge     = (*otelInt64Gauge)(nil)
	_ Int64Histogram = (*otelInt64Histogram)(nil)
)

type otelHandler struct {
	name         string
	meter        otelmetric.Meter
	provider     otelmetric.MeterProvider
	defaultAttrs *[]attribute.KeyValue

	int64CountersMtx sync.Mutex
	int64Counters    map[string]*otelInt64Counter
	int64HistosMtx   sync.Mutex
	int64Histos      map[string]*otelInt64Histogram
	int64GaugesMtx   sync.Mutex
	int64Gauges      map[string]*otelInt64Gauge
}

type baseAttrs struct {
	defaultAttrs *[]attribute.KeyValue
}

func (a *baseAttrs) getAttributes(tags map[string]string) []attribute.KeyValue {
	attrs := makeAttrs(tags)
	if a.defaultAttrs != nil {
		attrs = append(attrs, *a.defaultAttrs...)
	}

	return attrs
}

func (a *baseAttrs) setDefaultAttrs(attrs *[]attribute.KeyValue) {
	a.defaultAttrs = attrs
}

type otelInt64Counter struct {
	*baseAttrs
	counter otelmetric.Int64Counter
}

func (c *otelInt64Counter) Add(ctx context.Context, value int64, tags map[string]string) {
	attrs := c.getAttributes(tags)

	c.counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

type otelInt64Histogram struct {
	*baseAttrs
	histo otelmetric.Int64Histogram
}

func (h *otelInt64Histogram) Record(ctx context.Context, value int64, tags map[string]string) {
	attrs := h.getAttributes(tags)

	h.histo.Record(ctx, value, otelmetric.WithAttributes(attrs...))
}

type otelInt64Gauge struct {
	*baseAttrs
	value int64
	attrs []attribute.KeyValue
	gauge otelmetric.Int64ObservableGauge
}

func (g *otelInt64Gauge) Observe(_ context.Context, value int64, tags map[string]string) {
	g.attrs = g.getAttributes(tags)
	g.value = value
}

func (h *otelHandler) Int64Histogram(name string, description string, unit Unit) Int64Histogram {
	h.int64HistosMtx.Lock()
	defer h.int64HistosMtx.Unlock()

	name = strings.ToLower(name)

	c, ok := h.int64Histos[name]
	if !ok {
		histo, err := h.meter.Int64Histogram(name, otelmetric.WithDescription(description), otelmetric.WithUnit(string(unit)))
		if err != nil {
			panic(err)
		}
		c = &otelInt64Histogram{histo: histo, baseAttrs: &baseAttrs{}}
		h.int64Histos[name] = c
	}

	c.setDefaultAttrs(h.defaultAttrs)

	return c
}

func (h *otelHandler) Int64Counter(name string, description string, unit Unit) Int64Counter {
	h.int64CountersMtx.Lock()
	defer h.int64CountersMtx.Unlock()

	name = strings.ToLower(name)

	c, ok := h.int64Counters[name]
	if !ok {
		counter, err := h.meter.Int64Counter(name, otelmetric.WithDescription(description), otelmetric.WithUnit(string(unit)))
		if err != nil {
			panic(err)
		}
		c = &otelInt64Counter{counter: counter, baseAttrs: &baseAttrs{}}
		h.int64Counters[name] = c
	}

	c.setDefaultAttrs(h.defaultAttrs)

	return c
}

func (h *otelHandler) Int64Gauge(name string, description string, unit Unit) Int64Gauge {
	h.int64GaugesMtx.Lock()
	defer h.int64GaugesMtx.Unlock()

	name = strings.ToLower(name)

	c, ok := h.int64Gauges[name]
	if !ok {
		gauge, err := h.meter.Int64ObservableGauge(name, otelmetric.WithDescription(description), otelmetric.WithUnit(string(unit)))
		if err != nil {
			panic(err)
		}

		c = &otelInt64Gauge{gauge: gauge, baseAttrs: &baseAttrs{}}

		_, err = h.meter.RegisterCallback(func(ctx context.Context, observer otelmetric.Observer) error {
			observer.ObserveInt64(c.gauge, c.value, otelmetric.WithAttributes(c.attrs...))
			return nil
		}, c.gauge)
		if err != nil {
			panic(err)
		}

		h.int64Gauges[name] = c
	}

	c.setDefaultAttrs(h.defaultAttrs)

	return c
}

func (h *otelHandler) WithTags(tags map[string]string) Handler {
	attrs := makeAttrs(tags)

	h.defaultAttrs = &attrs

	return h
}

func makeAttrs(tags map[string]string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(tags))
	for k, v := range tags {
		attrs = append(attrs, attribute.String(k, v))
	}

	return attrs
}

func NewOtelHandler(_ context.Context, provider otelmetric.MeterProvider, name string) Handler {
	return &otelHandler{
		name:          name,
		meter:         provider.Meter(name),
		provider:      provider,
		int64Counters: make(map[string]*otelInt64Counter),
		int64Histos:   make(map[string]*otelInt64Histogram),
		int64Gauges:   make(map[string]*otelInt64Gauge),
	}
}
