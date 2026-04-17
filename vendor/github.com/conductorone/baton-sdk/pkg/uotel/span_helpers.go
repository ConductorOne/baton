package uotel

import (
	"context"
	"errors"

	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsExpectedError returns true for errors that are expected during normal
// operation and should not be recorded as span errors.
func IsExpectedError(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case grpccodes.Canceled, grpccodes.DeadlineExceeded:
			return true
		default:
		}
	}
	return false
}

// EndSpanWithError ends a span and records the error if it is non-nil and not
// an expected error. Expected errors (context cancellation, deadline exceeded)
// are recorded with an Unset status to avoid polluting error dashboards.
func EndSpanWithError(span trace.Span, err error) {
	if err != nil {
		if IsExpectedError(err) {
			span.SetStatus(otelcodes.Unset, err.Error())
		} else {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
	}
	span.End()
}
