package uotel

import (
	"context"
)

// InitOtel initializes OpenTelemetry with the given configuration.
// It returns a function that can be called to shutdown OpenTelemetry.
func InitOtel(ctx context.Context, opts ...Option) (context.Context, func(context.Context) error, error) {
	config := newConfig(opts...)

	ctx, err := config.init(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ctx, config.Close, nil
}
