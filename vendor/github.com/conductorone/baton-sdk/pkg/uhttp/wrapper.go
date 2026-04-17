package uhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	uRateLimit "go.uber.org/ratelimit"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/metrics"
	"github.com/conductorone/baton-sdk/pkg/ratelimit"
)

const (
	ContentType               = "Content-Type"
	applicationJSON           = "application/json"
	applicationXML            = "application/xml"
	applicationFormUrlencoded = "application/x-www-form-urlencoded"
	applicationVndApiJSON     = "application/vnd.api+json"
	acceptHeader              = "Accept"
	authorizationHeader       = "Authorization"

	httpCacheHitCounterName  = "baton_sdk.http_cache_hit"
	httpCacheMissCounterName = "baton_sdk.http_cache_miss"
	httpCacheHitCounterDesc  = "number of HTTP cache hits"
	httpCacheMissCounterDesc = "number of HTTP cache misses"
)

type WrapperResponse struct {
	Header     http.Header
	Body       []byte
	Status     string
	StatusCode int
}

type rateLimiterOption struct {
	rate int
	per  time.Duration
}

func (o rateLimiterOption) Apply(c *BaseHttpClient) {
	opts := []uRateLimit.Option{}
	if o.per > 0 {
		opts = append(opts, uRateLimit.Per(o.per))
	}
	c.rateLimiter = uRateLimit.New(o.rate, opts...)
}

// WithRateLimiter returns a WrapperOption that sets the rate limiter for the http client.
// `rate` is the number of requests allowed per `per` duration.
// `per` is the duration in which the rate limit is enforced.
// Example: WithRateLimiter(10, time.Second) will allow 10 requests per second.
func WithRateLimiter(rate int, per time.Duration) WrapperOption {
	return rateLimiterOption{rate: rate, per: per}
}

type metricsHandlerOption struct {
	handler metrics.Handler
}

func (o metricsHandlerOption) Apply(c *BaseHttpClient) {
	c.metricsHandler = o.handler
}

// WithMetricsHandler returns a WrapperOption that sets the metrics handler for the http client.
// When set, cache hits and misses will be recorded as metrics.
func WithMetricsHandler(handler metrics.Handler) WrapperOption {
	return metricsHandlerOption{handler: handler}
}

type WrapperOption interface {
	Apply(*BaseHttpClient)
}

// Keep a handle on all caches so we can clear them later.
var (
	caches    []icache
	cachesMtx sync.RWMutex
)

func ClearCaches(ctx context.Context) error {
	l := ctxzap.Extract(ctx)
	l.Debug("clearing caches")
	var errs []error
	cachesMtx.RLock()
	defer cachesMtx.RUnlock()
	for _, cache := range caches {
		l.Debug("clearing cache", zap.String("cache", fmt.Sprintf("%T", cache)), zap.Any("stats", cache.Stats(ctx)))
		err := cache.Clear(ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type (
	HttpClient interface {
		HttpClient() *http.Client
		Do(req *http.Request, options ...DoOption) (*http.Response, error)
		NewRequest(ctx context.Context, method string, url *url.URL, options ...RequestOption) (*http.Request, error)
	}
	BaseHttpClient struct {
		HttpClient     *http.Client
		rateLimiter    uRateLimit.Limiter
		baseHttpCache  icache
		metricsHandler metrics.Handler
	}

	DoOption      func(resp *WrapperResponse) error
	RequestOption func() (io.ReadWriter, map[string]string, error)
)

func NewBaseHttpClient(httpClient *http.Client, opts ...WrapperOption) *BaseHttpClient {
	ctx := context.TODO()
	client, err := NewBaseHttpClientWithContext(ctx, httpClient, opts...)
	if err != nil {
		return nil
	}
	return client
}

func NewBaseHttpClientWithContext(ctx context.Context, httpClient *http.Client, opts ...WrapperOption) (*BaseHttpClient, error) {
	l := ctxzap.Extract(ctx)

	cache, err := NewHttpCache(ctx, nil)
	if err != nil {
		l.Error("error creating http cache", zap.Error(err))
		return nil, err
	}
	cli := &BaseHttpClient{
		HttpClient:    httpClient,
		baseHttpCache: cache,
	}

	cachesMtx.Lock()
	caches = append(caches, cache)
	cachesMtx.Unlock()

	for _, opt := range opts {
		opt.Apply(cli)
	}

	return cli, nil
}

// WithJSONResponse is a wrapper that marshals the returned response body into
// the provided shape. If the API should return an empty JSON body (i.e. HTTP
// status code 204 No Content), then pass a `nil` to `response`.
func WithJSONResponse(response any) DoOption {
	return func(resp *WrapperResponse) error {
		contentHeader := resp.Header.Get(ContentType)

		if !IsJSONContentType(contentHeader) {
			if len(resp.Body) != 0 {
				// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
				return fmt.Errorf("unexpected content type for JSON response: %s. status code: %d", contentHeader, resp.StatusCode)
			}
			return fmt.Errorf("unexpected content type for JSON response: %s. status code: %d", contentHeader, resp.StatusCode)
		}
		if response == nil && len(resp.Body) == 0 {
			return nil
		}
		err := json.Unmarshal(resp.Body, response)
		if err != nil {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return fmt.Errorf("failed to unmarshal json response: %w. status code: %d", err, resp.StatusCode)
		}
		return nil
	}
}

// Ignore content type header and always try to parse the response as JSON.
func WithAlwaysJSONResponse(response any) DoOption {
	return func(resp *WrapperResponse) error {
		if response == nil && len(resp.Body) == 0 {
			return nil
		}
		err := json.Unmarshal(resp.Body, response)
		if err != nil {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return fmt.Errorf("failed to unmarshal json response: %w. status code: %d", err, resp.StatusCode)
		}
		return nil
	}
}

func WithXMLResponse(response any) DoOption {
	return func(resp *WrapperResponse) error {
		if !IsXMLContentType(resp.Header.Get(ContentType)) {
			return fmt.Errorf("unexpected content type for xml response: %s", resp.Header.Get(ContentType))
		}
		if response == nil && len(resp.Body) == 0 {
			return nil
		}
		err := xml.Unmarshal(resp.Body, response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal xml response: %w. status code: %d", err, resp.StatusCode)
		}
		return nil
	}
}

// Ignore content type header and always try to parse the response as XML.
func WithAlwaysXMLResponse(response any) DoOption {
	return func(resp *WrapperResponse) error {
		if response == nil && len(resp.Body) == 0 {
			return nil
		}
		err := xml.Unmarshal(resp.Body, response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal xml response: %w. status code: %d", err, resp.StatusCode)
		}
		return nil
	}
}

type ErrorResponse interface {
	Message() string
}

// GrpcCodeFromHTTPStatus maps an HTTP status code to the appropriate gRPC status code.
func GrpcCodeFromHTTPStatus(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusRequestTimeout:
		return codes.DeadlineExceeded
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return codes.Unavailable
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusNotImplemented:
		return codes.Unimplemented
	}
	switch {
	case httpStatus >= 500 && httpStatus <= 599:
		return codes.Unavailable
	case httpStatus >= 400 && httpStatus <= 499:
		return codes.InvalidArgument
	default:
		return codes.Unknown
	}
}

func WithErrorResponse(resource ErrorResponse) DoOption {
	return func(resp *WrapperResponse) error {
		if resp.StatusCode < 300 {
			return nil
		}

		contentHeader := resp.Header.Get(ContentType)

		grpcCode := GrpcCodeFromHTTPStatus(resp.StatusCode)

		if !IsJSONContentType(contentHeader) {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return status.Errorf(grpcCode,
				"unexpected content type for JSON error response: %s. status code: %d. body: %s",
				contentHeader, resp.StatusCode, string(resp.Body))
		}

		// Decode the JSON response body into the ErrorResponse
		if err := json.Unmarshal(resp.Body, &resource); err != nil {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return status.Errorf(grpcCode,
				"failed to unmarshal JSON error response: %v. status code: %d. body: %s",
				err, resp.StatusCode, string(resp.Body))
		}

		// Construct a more detailed error message
		errMsg := fmt.Sprintf("Request failed with status %d: %s", resp.StatusCode, resource.Message())

		return status.Error(grpcCode, errMsg)
	}
}

func WithRatelimitData(resource *v2.RateLimitDescription) DoOption {
	return func(resp *WrapperResponse) error {
		if resource == nil {
			return fmt.Errorf("WithRatelimitData: rate limit description is nil")
		}
		rl, err := ratelimit.ExtractRateLimitData(resp.StatusCode, &resp.Header)
		if err != nil {
			return err
		}

		resource.SetLimit(rl.GetLimit())
		resource.SetRemaining(rl.GetRemaining())
		resource.SetResetAt(rl.GetResetAt())
		resource.SetStatus(rl.GetStatus())

		return nil
	}
}

func WithResponse(response any) DoOption {
	return func(resp *WrapperResponse) error {
		if IsJSONContentType(resp.Header.Get(ContentType)) {
			return WithJSONResponse(response)(resp)
		}
		if IsXMLContentType(resp.Header.Get(ContentType)) {
			return WithXMLResponse(response)(resp)
		}

		return status.Error(codes.Unknown, "unsupported content type")
	}
}

// Handle anything that can be marshaled into JSON or XML.
// If the response is a list, its values will be put into the "items" field.
// If the response is a single value (int, string, bool, etc), it will be put into the "value" field.
// A response of `null` results in the "value" field being set to `nil`.
func WithGenericResponse(response *map[string]any) DoOption {
	return func(resp *WrapperResponse) error {
		if response == nil {
			return status.Error(codes.InvalidArgument, "response is nil")
		}

		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 && len(resp.Body) == 0 {
			return nil
		}

		var v any
		var err error

		if IsJSONContentType(resp.Header.Get(ContentType)) {
			err = WithJSONResponse(&v)(resp)
			if err != nil {
				return err
			}
			if list, ok := v.([]any); ok {
				(*response)["items"] = list
			} else if vMap, ok := v.(map[string]any); ok {
				*response = vMap
			} else if boolValue, ok := v.(bool); ok {
				(*response) = map[string]any{"value": boolValue}
			} else if floatValue, ok := v.(float64); ok {
				(*response) = map[string]any{"value": floatValue}
			} else if stringValue, ok := v.(string); ok {
				(*response) = map[string]any{"value": stringValue}
			} else if v == nil {
				// JSON response is literally `null`.
				(*response) = map[string]any{"value": nil}
			} else {
				return status.Errorf(codes.Internal, "unsupported value type: %T", v)
			}
			return nil
		}

		if IsXMLContentType(resp.Header.Get(ContentType)) {
			var xm xmlMap
			err = WithXMLResponse(&xm)(resp)
			if err != nil {
				return err
			}
			if vMap, ok := xm.data.(map[string]any); ok {
				*response = vMap
			} else {
				return status.Errorf(codes.Internal, "unsupported XML structure: %T", xm.data)
			}
			return nil
		}

		return status.Error(codes.Unknown, fmt.Sprintf("unsupported content type: %s", resp.Header.Get(ContentType)))
	}
}

func WrapErrors(preferredCode codes.Code, statusMsg string, errs ...error) error {
	st := status.New(preferredCode, statusMsg)

	if len(errs) == 0 {
		return st.Err()
	}

	allErrs := append([]error{st.Err()}, errs...)
	return errors.Join(allErrs...)
}

func WrapErrorsWithRateLimitInfo(preferredCode codes.Code, resp *http.Response, errs ...error) error {
	st := status.New(preferredCode, resp.Status)

	description, err := ratelimit.ExtractRateLimitData(resp.StatusCode, &resp.Header)
	// Ignore any error extracting rate limit data
	if err == nil {
		st, _ = st.WithDetails(description)
	}

	if len(errs) == 0 {
		return st.Err()
	}

	allErrs := append([]error{st.Err()}, errs...)
	return errors.Join(allErrs...)
}

func (c *BaseHttpClient) recordCacheHit(ctx context.Context) {
	if c.metricsHandler == nil {
		return
	}
	counter := c.metricsHandler.Int64Counter(httpCacheHitCounterName, httpCacheHitCounterDesc, metrics.Dimensionless)
	counter.Add(ctx, 1, nil)
}

func (c *BaseHttpClient) recordCacheMiss(ctx context.Context) {
	if c.metricsHandler == nil {
		return
	}
	counter := c.metricsHandler.Int64Counter(httpCacheMissCounterName, httpCacheMissCounterDesc, metrics.Dimensionless)
	counter.Add(ctx, 1, nil)
}

func (c *BaseHttpClient) Do(req *http.Request, options ...DoOption) (*http.Response, error) {
	var (
		err  error
		resp *http.Response
	)
	l := ctxzap.Extract(req.Context())

	// If a rate limiter is defined, take a token before making the request.
	if c.rateLimiter != nil {
		c.rateLimiter.Take()
	}

	if req.Method == http.MethodGet && req.Header.Get("Cache-Control") != "no-cache" {
		resp, err = c.baseHttpCache.Get(req)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			c.recordCacheMiss(req.Context())
		} else {
			c.recordCacheHit(req.Context())
		}
	}

	if resp == nil {
		//nolint:gosec // this HTTP wrapper intentionally supports arbitrary connector-defined endpoints.
		resp, err = c.HttpClient.Do(req)
		if err != nil {
			l.Error("base-http-client: HTTP error response", zap.Error(err))
			return resp, wrapTransientNetworkError(err)
		}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if len(body) > 0 {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		return resp, wrapTransientNetworkError(err)
	}

	// Replace resp.Body with a no-op closer so nobody has to worry about closing the reader.
	shouldPrint := os.Getenv("BATON_DEBUG_PRINT_RESPONSE_BODY")
	if shouldPrint != "" {
		resp.Body = io.NopCloser(wrapPrintBody(bytes.NewBuffer(body)))
	} else {
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	wresp := WrapperResponse{
		Header:     resp.Header,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
	}

	var optErrs []error
	for _, option := range options {
		optErr := option(&wresp)
		if optErr != nil {
			optErrs = append(optErrs, optErr)
		}
	}

	// Log response headers directly for certain errors
	if resp.StatusCode >= 400 {
		redactedHeaders := RedactSensitiveHeaders(resp.Header)
		logFields := []zap.Field{
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.Any("headers", redactedHeaders),
		}
		if resp.StatusCode >= 500 {
			l.Error("base-http-client: HTTP error status", logFields...)
		} else {
			l.Warn("base-http-client: HTTP error status", logFields...)
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		grpcCode := GrpcCodeFromHTTPStatus(resp.StatusCode)
		return resp, WrapErrorsWithRateLimitInfo(grpcCode, resp, optErrs...)
	}

	if req.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
		cacheErr := c.baseHttpCache.Set(req, resp)
		if cacheErr != nil {
			l.Warn("error setting cache", zap.String("url", req.URL.String()), zap.Error(cacheErr))
		}
	}

	return resp, errors.Join(optErrs...)
}

var sensitiveStrings = []string{
	"api-key",
	"auth",
	"cookie",
	"proxy-authorization",
	"set-cookie",
	"x-forwarded-for",
	"x-forwarded-proto",
}

func RedactSensitiveHeaders(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	safe := make(http.Header, len(h))
	for k, v := range h {
		sensitive := false
		headerKey := strings.ToLower(k)
		for _, sensitiveString := range sensitiveStrings {
			if strings.Contains(headerKey, sensitiveString) {
				sensitive = true
				break
			}
		}

		if sensitive {
			safe[k] = []string{"REDACTED"}
		} else {
			safe[k] = v
		}
	}
	return safe
}

func WithHeader(key, value string) RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		return nil, map[string]string{
			key: value,
		}, nil
	}
}

func WithNoCache() RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		return nil, map[string]string{
			"Cache-Control": "no-cache",
		}, nil
	}
}

func WithBody(body []byte) RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		return bytes.NewBuffer(body), nil, nil
	}
}

func WithJSONBody(body any) RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		buffer := new(bytes.Buffer)
		err := json.NewEncoder(buffer).Encode(body)
		if err != nil {
			return nil, nil, err
		}

		_, headers, err := WithContentTypeJSONHeader()()
		if err != nil {
			return nil, nil, err
		}

		return buffer, headers, nil
	}
}

func WithFormBody(body string) RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		var buffer bytes.Buffer
		_, err := buffer.WriteString(body)
		if err != nil {
			return nil, nil, err
		}

		_, headers, err := WithContentTypeFormHeader()()
		if err != nil {
			return nil, nil, err
		}

		return &buffer, headers, nil
	}
}

func WithXMLBody(body any) RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		var buffer bytes.Buffer

		err := xml.NewEncoder(&buffer).Encode(body)
		if err != nil {
			return nil, nil, err
		}

		_, headers, err := WithContentTypeXMLHeader()()
		if err != nil {
			return nil, nil, err
		}

		return &buffer, headers, nil
	}
}

func WithAcceptJSONHeader() RequestOption {
	return WithAccept(applicationJSON)
}

func WithContentTypeJSONHeader() RequestOption {
	return WithContentType(applicationJSON)
}

func WithAcceptXMLHeader() RequestOption {
	return WithAccept(applicationXML)
}

func WithContentTypeXMLHeader() RequestOption {
	return WithContentType(applicationXML)
}

func WithContentTypeFormHeader() RequestOption {
	return WithContentType(applicationFormUrlencoded)
}

func WithContentTypeVndHeader() RequestOption {
	return WithContentType(applicationVndApiJSON)
}

func WithAcceptVndJSONHeader() RequestOption {
	return WithAccept(applicationVndApiJSON)
}

func WithContentType(ctype string) RequestOption {
	return WithHeader(ContentType, ctype)
}

func WithAccept(value string) RequestOption {
	return WithHeader(acceptHeader, value)
}

func WithBearerToken(token string) RequestOption {
	return WithHeader(authorizationHeader, fmt.Sprintf("Bearer %s", token))
}

func (c *BaseHttpClient) NewRequest(ctx context.Context, method string, url *url.URL, options ...RequestOption) (*http.Request, error) {
	var buffer io.ReadWriter
	var headers = make(map[string]string)
	for _, option := range options {
		buf, h, err := option()
		if err != nil {
			return nil, err
		}

		if buf != nil {
			buffer = buf
		}

		for k, v := range h {
			headers[k] = v
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), buffer)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
