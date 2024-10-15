package ratelimit

import (
	"context"
	"sync"
	"time"

	ratelimitV1 "github.com/conductorone/baton-sdk/pb/c1/ratelimit/v1"
	v1 "github.com/conductorone/baton-sdk/pb/c1/ratelimit/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	rl "go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type MemRateLimiter struct {
	sync.Mutex
	limiter    rl.Limiter
	now        func() time.Time
	usePercent float64
}

// TODO
func (m *MemRateLimiter) Do(ctx context.Context, req *ratelimitV1.DoRequest) (*ratelimitV1.DoResponse, error) {
	if m.limiter == nil {
		return &v1.DoResponse{
			RequestToken: req.RequestToken,
			Description: &v1.RateLimitDescription{
				Status: v1.RateLimitDescription_STATUS_EMPTY,
			},
		}, nil
	}

	m.limiter.Take()

	return &v1.DoResponse{
		RequestToken: req.RequestToken,
		Description: &v1.RateLimitDescription{
			Status: v1.RateLimitDescription_STATUS_EMPTY,
		},
	}, nil
}

// Report updates the rate limiter with relevant information.
func (m *MemRateLimiter) Report(ctx context.Context, req *ratelimitV1.ReportRequest) (*ratelimitV1.ReportResponse, error) {
	m.Lock()
	defer m.Unlock()

	if m.usePercent == 0 {
		return &ratelimitV1.ReportResponse{}, nil
	}

	if req.GetDescription() == nil {
		return &ratelimitV1.ReportResponse{}, nil
	}
	desc := req.GetDescription()

	if desc.ResetAt == nil {
		return &ratelimitV1.ReportResponse{}, nil
	}

	if desc.Remaining == 0 {
		return &ratelimitV1.ReportResponse{}, nil
	}

	resetAt := desc.ResetAt.AsTime().UTC()
	windowDuration := resetAt.Sub(m.now())
	if windowDuration > 5*time.Minute {
		windowDuration = 5 * time.Minute
	}
	remaining := int64(m.usePercent * float64(desc.Remaining))
	if remaining < 1 {
		remaining = 1
	}

	limiterSize := remaining / int64(windowDuration/time.Second)
	ctxzap.Extract(ctx).Debug(
		"updating rate limiter",
		zap.Int64("calculated_remaining", remaining),
		zap.Int64("remaining", desc.Remaining),
		zap.Int64("rate", limiterSize),
		zap.Time("reset_at", resetAt),
	)
	m.limiter = rl.New(int(limiterSize))

	return &ratelimitV1.ReportResponse{}, nil
}

// NewSlidingMemoryRateLimiter returns an in-memory limiter that attempts to use rate limiting reports to define a
// window size and set the limits to a fair amount given the `usePercent` argument.
func NewSlidingMemoryRateLimiter(ctx context.Context, now func() time.Time, usePercent float64) *MemRateLimiter {
	return &MemRateLimiter{
		limiter:    rl.New(100, rl.Per(time.Second)),
		usePercent: usePercent,
		now:        now,
	}
}

// NewFixedMemoryRateLimiter returns an in-memory limiter that allows rate/period requests e.g. 100/second, 1000/minute.
func NewFixedMemoryRateLimiter(ctx context.Context, now func() time.Time, rate int64, period time.Duration) *MemRateLimiter {
	return &MemRateLimiter{
		limiter: rl.New(int(rate), rl.Per(period)),
		now:     now,
	}
}
