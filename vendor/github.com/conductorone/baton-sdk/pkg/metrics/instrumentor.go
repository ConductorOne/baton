package metrics //nolint:revive,nolintlint // we can't change the package name for backwards compatibility

import (
	"context"
	"strconv"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/tasks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	taskSuccessCounterName = "baton_sdk.task_success"
	taskFailureCounterName = "baton_sdk.task_failure"
	taskDurationHistoName  = "baton_sdk.task_latency"
	taskSuccessCounterDesc = "number of successful tasks by task type"
	taskFailureCounterDesc = "number of failed tasks by task type, grpc code, and rate limit status"
	taskDurationHistoDesc  = "duration of all tasks by task type and status"
)

// FailureReason contains extracted information about why a task failed.
type FailureReason struct {
	GrpcCode    codes.Code
	IsRateLimit bool
}

// extractFailureReason extracts the gRPC status code and rate limit information from an error.
func extractFailureReason(err error) FailureReason {
	if err == nil {
		return FailureReason{GrpcCode: codes.Unknown}
	}

	reason := FailureReason{
		GrpcCode: status.Code(err),
	}

	// Check for rate limit details in gRPC status
	if st, ok := status.FromError(err); ok {
		for _, detail := range st.Details() {
			if rl, ok := detail.(*v2.RateLimitDescription); ok {
				if rl.GetStatus() == v2.RateLimitDescription_STATUS_OVERLIMIT {
					reason.IsRateLimit = true
					break
				}
			}
		}
	}

	return reason
}

type M struct {
	underlying Handler
}

func (m *M) RecordTaskSuccess(ctx context.Context, task tasks.TaskType, dur time.Duration) {
	c := m.underlying.Int64Counter(taskSuccessCounterName, taskSuccessCounterDesc, Dimensionless)
	h := m.underlying.Int64Histogram(taskDurationHistoName, taskDurationHistoDesc, Milliseconds)
	c.Add(ctx, 1, map[string]string{"task_type": task.String()})
	h.Record(ctx, dur.Milliseconds(), map[string]string{"task_type": task.String(), "task_status": "success"})
}

func (m *M) RecordTaskFailure(ctx context.Context, task tasks.TaskType, dur time.Duration, err error) {
	reason := extractFailureReason(err)

	c := m.underlying.Int64Counter(taskFailureCounterName, taskFailureCounterDesc, Dimensionless)
	h := m.underlying.Int64Histogram(taskDurationHistoName, taskDurationHistoDesc, Milliseconds)

	counterAttrs := map[string]string{
		"task_type":     task.String(),
		"grpc_code":     reason.GrpcCode.String(),
		"is_rate_limit": strconv.FormatBool(reason.IsRateLimit),
	}
	histoAttrs := map[string]string{
		"task_type":     task.String(),
		"task_status":   "failure",
		"grpc_code":     reason.GrpcCode.String(),
		"is_rate_limit": strconv.FormatBool(reason.IsRateLimit),
	}

	c.Add(ctx, 1, counterAttrs)
	h.Record(ctx, dur.Milliseconds(), histoAttrs)
}

func New(handler Handler) *M {
	return &M{underlying: handler}
}
