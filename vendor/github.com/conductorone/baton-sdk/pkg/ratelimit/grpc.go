package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"time"

	connectorV2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	ratelimitV1 "github.com/conductorone/baton-sdk/pb/c1/ratelimit/v1"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// wait calculates the duration from 'now' to 'until', clamped to MAX/MIN_RATELIMIT_WAIT.
func wait(start, now, until time.Time) (time.Duration, bool) {
	maxWaitTime := start.Add(maxRatelimitWait)

	// we've waited past our max wait time
	if now.After(maxWaitTime) {
		return 0, false
	}
	// clamp
	d := until.Sub(now)
	if d < minRatelimitWait {
		d = minRatelimitWait
	} else if d > maxRatelimitWait {
		d = maxRatelimitWait
	}
	// check if our wait time would exceed our maximum wait time
	if now.Add(d).After(maxWaitTime) {
		return 0, false
	}
	return d, true
}

type hasAnnos interface {
	GetAnnotations() []*anypb.Any
}

type hasResource interface {
	GetResource() *connectorV2.Resource
}

type hasResourceType interface {
	GetResourceTypeId() string
}

func getRatelimitDescriptors(ctx context.Context, method string, in interface{}, descriptors ...*ratelimitV1.RateLimitDescriptors_Entry) *ratelimitV1.RateLimitDescriptors {
	ret := &ratelimitV1.RateLimitDescriptors{
		Entries: descriptors,
	}

	ret.Entries = append(ret.Entries, &ratelimitV1.RateLimitDescriptors_Entry{
		Key:   descriptorKeyConnectorMethod,
		Value: method,
	})

	// ListEntitlements, ListGrants
	if req, ok := in.(hasResource); ok {
		ret.Entries = append(ret.Entries, &ratelimitV1.RateLimitDescriptors_Entry{
			Key:   descriptorKeyConnectorResourceType,
			Value: req.GetResource().Id.ResourceType,
		})
		return ret
	}

	// ListResources
	if req, ok := in.(hasResourceType); ok {
		ret.Entries = append(ret.Entries, &ratelimitV1.RateLimitDescriptors_Entry{
			Key:   descriptorKeyConnectorResourceType,
			Value: req.GetResourceTypeId(),
		})
		return ret
	}

	return ret
}

// UnaryInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryInterceptor(now func() time.Time, descriptors ...*ratelimitV1.RateLimitDescriptors_Entry) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// If this is a call to the rate limit service, skip it
		if strings.HasPrefix(method, "/c1.ratelimit.v1.RateLimiterService/") {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		l := ctxzap.Extract(ctx)

		rlClient := ratelimitV1.NewRateLimiterServiceClient(cc)

		start := now().UTC()
		token := ""

		l.Debug("beginning ratelimit check for request", zap.String("method", method))

		rlDescriptors := getRatelimitDescriptors(ctx, method, req, descriptors...)

		for {
			rlReq := &ratelimitV1.DoRequest{
				RequestToken: token,
				Service:      connectorServiceKey,
				Descriptors:  rlDescriptors,
			}
			resp, err := rlClient.Do(ctx, rlReq)
			if err != nil {
				l.Error("ratelimit: error", zap.Error(err))
				return status.Error(codes.Unknown, err.Error())
			}
			token = resp.RequestToken

			switch resp.Description.Status {
			case ratelimitV1.RateLimitDescription_STATUS_OK, ratelimitV1.RateLimitDescription_STATUS_EMPTY:
				l.Debug("ratelimit ok - calling method", zap.String("method", method))
				err = invoker(ctx, method, req, reply, cc, opts...)
				if err != nil {
					rlErr := reportRatelimit(
						ctx,
						rlClient,
						rlReq.RequestToken,
						ratelimitV1.RateLimitDescription_STATUS_ERROR,
						rlDescriptors,
						nil,
					)
					if rlErr != nil {
						return fmt.Errorf("ratelimit: error reporting ratelimit after request error: %w", err)
					}

					l.Error("ratelimit: error running client request", zap.Error(err))
					return err
				}

				if reply != nil {
					if resp, ok := req.(hasAnnos); ok {
						err = reportRatelimit(ctx, rlClient, rlReq.RequestToken, ratelimitV1.RateLimitDescription_STATUS_OK, rlDescriptors, resp.GetAnnotations())
						if err != nil {
							l.Error("ratelimit: error reporting rate limit", zap.Error(err))
							return nil // Explicitly not failing the request as it has already been run successfully.
						}
					}
				}

				return nil

			case ratelimitV1.RateLimitDescription_STATUS_OVERLIMIT:
				resetAt := resp.Description.ResetAt.AsTime()
				d, ok := wait(start, now().UTC(), resetAt)
				if !ok {
					l.Error("ratelimit: timeout")
					return status.Error(codes.ResourceExhausted, "overlimit")
				}

				l.Info("ratelimit overlimit - waiting", zap.String("method", method), zap.Duration("wait_period", d))

				// Overlimit -- wait up to maxRatelimitWait before trying the request again or the request is cancelled.
				select {
				case <-time.After(d):
					continue
				case <-ctx.Done():
					return status.FromContextError(ctx.Err()).Err()
				}

			default:
				l.Warn("ratelimit: unspecified status")
				return nil
			}
		}
	}
}

func reportRatelimit(
	ctx context.Context,
	rlClient ratelimitV1.RateLimiterServiceClient,
	token string,
	status ratelimitV1.RateLimitDescription_Status,
	descriptors *ratelimitV1.RateLimitDescriptors,
	anys []*any.Any,
) error {
	l := ctxzap.Extract(ctx)
	annos := annotations.Annotations(anys)

	rlAnnotation := &ratelimitV1.RateLimitDescription{
		Status: status,
	}

	_, err := annos.Pick(rlAnnotation)
	if err != nil {
		return err
	}

	_, err = rlClient.Report(ctx, &ratelimitV1.ReportRequest{
		RequestToken: token,
		Description:  rlAnnotation,
		Descriptors:  descriptors,
		Service:      "connector",
	})
	if err != nil {
		l.Error("ratelimit: report failed", zap.Error(err))
		return err
	}

	return nil
}
