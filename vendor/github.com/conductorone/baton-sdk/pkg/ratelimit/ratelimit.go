package ratelimit

import (
	"context"
	"fmt"
	"time"

	ratelimitV1 "github.com/conductorone/baton-sdk/pb/c1/ratelimit/v1"
)

const (
	connectorServiceKey                = "connector"
	descriptorKeyConnectorMethod       = "connector_method"
	descriptorKeyConnectorResourceType = "connector_resource_type"

	minRatelimitWait = 1 * time.Second                     // Minimum time to wait after a request was ratelimited before trying again
	maxRatelimitWait = (1 * time.Hour) + (5 * time.Minute) // Maximum time to wait after a request was ratelimited before erroring
)

// NewLimiter configures a RateLimitServer server.
func NewLimiter(ctx context.Context, now func() time.Time, cfg *ratelimitV1.RateLimiterConfig) (ratelimitV1.RateLimiterServiceServer, error) {
	if cfg == nil {
		return &NoOpRateLimiter{}, nil
	}

	if c := cfg.GetDisabled(); c != nil {
		return &NoOpRateLimiter{}, nil
	}

	if c := cfg.GetSlidingMem(); c != nil {
		return NewSlidingMemoryRateLimiter(ctx, now, c.UsePercent), nil
	}

	if c := cfg.GetFixedMem(); c != nil {
		return NewFixedMemoryRateLimiter(ctx, now, c.Rate, c.Period.AsDuration()), nil
	}

	if c := cfg.GetExternal(); c != nil {
		return nil, fmt.Errorf("external rate limiters are not implemented yet")
	}

	return nil, fmt.Errorf("invalid ratelimiter config")
}
