package uhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"syscall"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// wrapTransientNetworkError mirrors Baton HTTP retry classification for callers
// that use the transport directly, such as oauth2-backed SDK clients.
func wrapTransientNetworkError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return WrapErrors(codes.Unavailable, "unexpected EOF", err)
	}
	if errors.Is(err, syscall.ECONNRESET) {
		return WrapErrors(codes.Unavailable, "connection reset", err)
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return WrapErrors(codes.DeadlineExceeded, fmt.Sprintf("request timeout: %v", urlErr.URL), urlErr)
		}
		if urlErr.Temporary() {
			return WrapErrors(codes.Unavailable, fmt.Sprintf("temporary error: %v", urlErr.URL), urlErr)
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request timeout")
	}

	return err
}
