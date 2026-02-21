package metrics //nolint:revive,nolintlint // we can't change the package name for backwards compatibility

import (
	"context"
)

type Handler interface {
	Int64Counter(name string, description string, unit Unit) Int64Counter
	Int64Gauge(name string, description string, unit Unit) Int64Gauge
	Int64Histogram(name string, description string, unit Unit) Int64Histogram
	WithTags(tags map[string]string) Handler
}

type Int64Counter interface {
	Add(ctx context.Context, value int64, tags map[string]string)
}

type Int64Histogram interface {
	Record(ctx context.Context, value int64, tags map[string]string)
}

type Int64Gauge interface {
	Observe(ctx context.Context, value int64, tags map[string]string)
}

type Unit string

const (
	Dimensionless Unit = "1"
	Bytes         Unit = "By"
	Milliseconds  Unit = "ms"
)
