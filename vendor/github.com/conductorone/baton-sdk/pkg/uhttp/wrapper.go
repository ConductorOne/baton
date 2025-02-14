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
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	uRateLimit "go.uber.org/ratelimit"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
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

type WrapperOption interface {
	Apply(*BaseHttpClient)
}

// Keep a handle on all caches so we can clear them later.
var caches []icache

func ClearCaches(ctx context.Context) error {
	l := ctxzap.Extract(ctx)
	l.Debug("clearing caches")
	var err error
	for _, cache := range caches {
		l.Debug("clearing cache", zap.String("cache", fmt.Sprintf("%T", cache)), zap.Any("stats", cache.Stats(ctx)))
		err = cache.Clear(ctx)
		if err != nil {
			err = errors.Join(err, err)
		}
	}
	return err
}

type (
	HttpClient interface {
		HttpClient() *http.Client
		Do(req *http.Request, options ...DoOption) (*http.Response, error)
		NewRequest(ctx context.Context, method string, url *url.URL, options ...RequestOption) (*http.Request, error)
	}
	BaseHttpClient struct {
		HttpClient    *http.Client
		rateLimiter   uRateLimit.Limiter
		baseHttpCache icache
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

	caches = append(caches, cache)

	for _, opt := range opts {
		opt.Apply(cli)
	}

	return cli, nil
}

// WithJSONResponse is a wrapper that marshals the returned response body into
// the provided shape. If the API should return an empty JSON body (i.e. HTTP
// status code 204 No Content), then pass a `nil` to `response`.
func WithJSONResponse(response interface{}) DoOption {
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
func WithAlwaysJSONResponse(response interface{}) DoOption {
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

type ErrorResponse interface {
	Message() string
}

func WithErrorResponse(resource ErrorResponse) DoOption {
	return func(resp *WrapperResponse) error {
		if resp.StatusCode < 300 {
			return nil
		}

		contentHeader := resp.Header.Get(ContentType)

		if !IsJSONContentType(contentHeader) {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return fmt.Errorf("unexpected content type for JSON error response: %s. status code: %d", contentHeader, resp.StatusCode)
		}

		// Decode the JSON response body into the ErrorResponse
		if err := json.Unmarshal(resp.Body, &resource); err != nil {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return fmt.Errorf("failed to unmarshal JSON error response: %w. status code: %d", err, resp.StatusCode)
		}

		// Construct a more detailed error message
		errMsg := fmt.Sprintf("Request failed with status %d: %s", resp.StatusCode, resource.Message())

		return status.Error(codes.Unknown, errMsg)
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

		resource.Limit = rl.Limit
		resource.Remaining = rl.Remaining
		resource.ResetAt = rl.ResetAt
		resource.Status = rl.Status

		return nil
	}
}

func WithXMLResponse(response interface{}) DoOption {
	return func(resp *WrapperResponse) error {
		if !IsXMLContentType(resp.Header.Get(ContentType)) {
			return fmt.Errorf("unexpected content type for xml response: %s", resp.Header.Get(ContentType))
		}
		if response == nil && len(resp.Body) == 0 {
			return nil
		}
		err := xml.Unmarshal(resp.Body, response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal xml response: %w. body %v", err, resp.Body)
		}
		return nil
	}
}

func WithResponse(response interface{}) DoOption {
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

	if req.Method == http.MethodGet {
		resp, err = c.baseHttpCache.Get(req)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			l.Debug("http cache miss", zap.String("url", req.URL.String()))
		} else {
			l.Debug("http cache hit", zap.String("url", req.URL.String()))
		}
	}

	if resp == nil {
		resp, err = c.HttpClient.Do(req)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				if urlErr.Timeout() {
					return nil, WrapErrors(codes.DeadlineExceeded, fmt.Sprintf("request timeout: %v", urlErr.URL), urlErr)
				}
				if urlErr.Temporary() {
					return nil, WrapErrors(codes.Unavailable, fmt.Sprintf("temporary error: %v", urlErr.URL), urlErr)
				}
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, status.Error(codes.DeadlineExceeded, "request timeout")
			}
			return nil, err
		}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if len(body) > 0 {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		// Turn certain body read errors into grpc statuses so we retry
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return resp, WrapErrors(codes.Unavailable, "unexpected EOF", err)
		}
		if errors.Is(err, syscall.ECONNRESET) {
			return resp, WrapErrors(codes.Unavailable, "connection reset", err)
		}
		return resp, err
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

	switch resp.StatusCode {
	case http.StatusRequestTimeout:
		return resp, WrapErrorsWithRateLimitInfo(codes.DeadlineExceeded, resp, optErrs...)
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable:
		return resp, WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, optErrs...)
	case http.StatusNotFound:
		return resp, WrapErrorsWithRateLimitInfo(codes.NotFound, resp, optErrs...)
	case http.StatusUnauthorized:
		return resp, WrapErrorsWithRateLimitInfo(codes.Unauthenticated, resp, optErrs...)
	case http.StatusForbidden:
		return resp, WrapErrorsWithRateLimitInfo(codes.PermissionDenied, resp, optErrs...)
	case http.StatusConflict:
		return resp, WrapErrorsWithRateLimitInfo(codes.AlreadyExists, resp, optErrs...)
	case http.StatusNotImplemented:
		return resp, WrapErrorsWithRateLimitInfo(codes.Unimplemented, resp, optErrs...)
	}

	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return resp, WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, optErrs...)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, WrapErrorsWithRateLimitInfo(codes.Unknown, resp, append(optErrs, fmt.Errorf("unexpected status code: %d", resp.StatusCode))...)
	}

	if req.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
		cacheErr := c.baseHttpCache.Set(req, resp)
		if cacheErr != nil {
			l.Warn("error setting cache", zap.String("url", req.URL.String()), zap.Error(cacheErr))
		}
	}

	return resp, errors.Join(optErrs...)
}

func WithHeader(key, value string) RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		return nil, map[string]string{
			key: value,
		}, nil
	}
}

func WithJSONBody(body interface{}) RequestOption {
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

func WithAcceptJSONHeader() RequestOption {
	return WithAccept(applicationJSON)
}

func WithContentTypeJSONHeader() RequestOption {
	return WithContentType(applicationJSON)
}

func WithAcceptXMLHeader() RequestOption {
	return WithAccept(applicationXML)
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
	var headers map[string]string = make(map[string]string)
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
