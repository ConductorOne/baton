package uhttp

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type tlsClientConfigOption struct {
	config *tls.Config
}

func (o tlsClientConfigOption) Apply(c *Transport) {
	c.tlsClientConfig = o.config
}

// WithTLSClientConfig returns an Option that sets the TLS client configuration.
// `tlsConfig` is a structure that is used to configure a TLS client or server.
func WithTLSClientConfig(tlsConfig *tls.Config) Option {
	return tlsClientConfigOption{config: tlsConfig}
}

type loggerOption struct {
	log    bool
	logger *zap.Logger
}

func (o loggerOption) Apply(c *Transport) {
	c.log = o.log
	c.logger = o.logger
}

// WithLogger sets a logger options to the transport layer.
func WithLogger(log bool, logger *zap.Logger) Option {
	return loggerOption{
		log:    log,
		logger: logger,
	}
}

type userAgentOption struct {
	userAgent string
}

func (o userAgentOption) Apply(c *Transport) {
	c.userAgent = o.userAgent
}

// WithUserAgent sets a user agent option to the transport layer.
func WithUserAgent(userAgent string) Option {
	return userAgentOption{
		userAgent: userAgent,
	}
}

type Option interface {
	Apply(*Transport)
}

// NewClient creates a new HTTP client that uses the given context and options to create a new transport layer.
func NewClient(ctx context.Context, options ...Option) (*http.Client, error) {
	httpClient := &http.Client{
		Timeout: 300 * time.Second, // 5 minutes
	}
	t, err := NewTransport(ctx, options...)
	if err != nil {
		return nil, err
	}
	httpClient.Transport = t
	return httpClient, nil
}
