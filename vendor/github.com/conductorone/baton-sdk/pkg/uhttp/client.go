package uhttp

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net/http"
	"sort"
	"strings"
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

type icache interface {
	Get(req *http.Request) (*http.Response, error)
	Set(req *http.Request, value *http.Response) error
	Clear(ctx context.Context) error
	Stats(ctx context.Context) CacheStats
}

// CreateCacheKey generates a cache key based on the request URL, query parameters, and headers.
func CreateCacheKey(req *http.Request) (string, error) {
	if req == nil {
		return "", fmt.Errorf("request is nil")
	}
	var sortedParams []string
	// Normalize the URL path
	path := strings.ToLower(req.URL.Path)
	// Combine the path with sorted query parameters
	queryParams := req.URL.Query()
	for k, v := range queryParams {
		for _, value := range v {
			sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", k, value))
		}
	}

	sort.Strings(sortedParams)
	queryString := strings.Join(sortedParams, "&")
	// Include relevant headers in the cache key
	var headerParts []string
	for key, values := range req.Header {
		for _, value := range values {
			if key == "Accept" || key == "Content-Type" || key == "Cookie" || key == "Range" {
				headerParts = append(headerParts, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	sort.Strings(headerParts)
	headersString := strings.Join(headerParts, "&")
	// Create a unique string for the cache key
	cacheString := fmt.Sprintf("%s?%s&headers=%s", path, queryString, headersString)

	// Hash the cache string to create a key
	hash := sha256.New()
	_, err := hash.Write([]byte(cacheString))
	if err != nil {
		return "", err
	}

	cacheKey := fmt.Sprintf("%x", hash.Sum(nil))
	return cacheKey, nil
}
