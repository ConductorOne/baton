package uhttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/conductorone/baton-sdk/pkg/sdk"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

var loggedResponseHeaders = []string{
	// Limit headers
	"X-Ratelimit-Limit",
	"Ratelimit-Limit",
	"X-RateLimit-Requests-Limit", // Linear uses a non-standard header
	"X-Rate-Limit-Limit",         // Okta uses a non-standard header

	// Remaining headers
	"X-Ratelimit-Remaining",
	"Ratelimit-Remaining",
	"X-RateLimit-Requests-Remaining", // Linear uses a non-standard header
	"X-Rate-Limit-Remaining",         // Okta uses a non-standard header

	// Reset headers
	"X-Ratelimit-Reset",
	"Ratelimit-Reset",
	"X-RateLimit-Requests-Reset", // Linear uses a non-standard header
	"X-Rate-Limit-Reset",         // Okta uses a non-standard header
	"Retry-After",                // Often returned with 429
}

// NewTransport creates a new Transport, applies the options, and then cycles the transport.
func NewTransport(ctx context.Context, options ...Option) (*Transport, error) {
	t := newTransport()
	for _, opt := range options {
		opt.Apply(t)
	}
	t.userAgent = t.userAgent + " baton-sdk/" + sdk.Version

	_, err := t.cycle(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

type Transport struct {
	userAgent       string
	tlsClientConfig *tls.Config
	roundTripper    http.RoundTripper
	logger          *zap.Logger
	log             bool
	nextCycle       time.Time
	mtx             sync.RWMutex
}

func newTransport() *Transport {
	return &Transport{
		tlsClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

func (t *Transport) cycle(ctx context.Context) (http.RoundTripper, error) {
	n := time.Now()
	t.mtx.RLock()
	if t.nextCycle.After(n) && t.roundTripper != nil {
		defer t.mtx.RUnlock()
		return t.roundTripper, nil
	}
	t.mtx.RUnlock()

	t.mtx.Lock()
	defer t.mtx.Unlock()
	n = time.Now()
	// other goroutine changed it under us
	if t.nextCycle.After(n) && t.roundTripper != nil {
		return t.roundTripper, nil
	}
	var err error
	t.roundTripper, err = t.make(ctx)
	if err != nil {
		return nil, err
	}
	t.nextCycle = n.Add(time.Minute * 5)
	return t.roundTripper, nil
}

type userAgentTripper struct {
	next      http.RoundTripper
	userAgent string
}

func (uat *userAgentTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", uat.userAgent)
	}
	return uat.next.RoundTrip(req)
}

func (t *Transport) make(_ context.Context) (http.RoundTripper, error) {
	// based on http.DefaultTransport
	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       t.tlsClientConfig,
	}
	err := http2.ConfigureTransport(baseTransport)
	if err != nil {
		return nil, err
	}
	var rv http.RoundTripper = baseTransport
	rv = &userAgentTripper{next: rv, userAgent: t.userAgent}
	return rv, nil
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	rt, err := t.cycle(ctx)
	if err != nil {
		return nil, fmt.Errorf("uhttp: cycle failed: %w", err)
	}
	if t.log {
		t.l(ctx).Debug("Request started",
			zap.String("http.method", req.Method),
			zap.String("http.url_details.host", req.URL.Host),
			zap.String("http.url_details.path", req.URL.Path),
			zap.String("http.url_details.query", req.URL.RawQuery),
		)
	}
	resp, err := rt.RoundTrip(req)
	if t.log {
		fields := []zap.Field{zap.String("http.method", req.Method),
			zap.String("http.url_details.host", req.URL.Host),
			zap.String("http.url_details.path", req.URL.Path),
			zap.String("http.url_details.query", req.URL.RawQuery),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		if resp != nil {
			fields = append(fields, zap.Int("http.status_code", resp.StatusCode))

			headers := make(map[string][]string, len(resp.Header))
			for _, header := range loggedResponseHeaders {
				if v := resp.Header.Values(header); len(v) > 0 {
					headers[header] = v
				}
			}

			fields = append(fields, zap.Any("http.headers", headers))
		}

		t.l(ctx).Debug("Request complete", fields...)
	}
	return resp, err
}

func (t *Transport) l(ctx context.Context) *zap.Logger {
	if t.logger != nil {
		return t.logger
	}
	return ctxzap.Extract(ctx)
}
