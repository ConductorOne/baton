package helpers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SplitFullName(name string) (string, string) {
	names := strings.SplitN(name, " ", 2)
	var firstName, lastName string

	switch len(names) {
	case 1:
		firstName = names[0]
	case 2:
		firstName = names[0]
		lastName = names[1]
	}

	return firstName, lastName
}

func ExtractRateLimitData(statusCode int, header *http.Header) (*v2.RateLimitDescription, error) {
	if header == nil {
		return nil, nil
	}

	var rlstatus v2.RateLimitDescription_Status

	var limit int64
	var err error
	limitStr := header.Get("X-Ratelimit-Limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var remaining int64
	remainingStr := header.Get("X-Ratelimit-Remaining")
	if remainingStr != "" {
		remaining, err = strconv.ParseInt(remainingStr, 10, 64)
		if err != nil {
			return nil, err
		}
		if remaining > 0 {
			rlstatus = v2.RateLimitDescription_STATUS_OK
		}
	}

	var resetAt time.Time
	reset := header.Get("X-Ratelimit-Reset")
	if reset != "" {
		res, err := strconv.ParseInt(reset, 10, 64)
		if err != nil {
			return nil, err
		}

		resetAt = time.Now().Add(time.Second * time.Duration(res))
	}

	// If we didn't get any rate limit headers and status code is 429, return some sane defaults
	if limit == 0 && remaining == 0 && resetAt.IsZero() && statusCode == http.StatusTooManyRequests {
		limit = 1
		remaining = 0
		resetAt = time.Now().Add(time.Second * 60)
		rlstatus = v2.RateLimitDescription_STATUS_OVERLIMIT
	}

	return &v2.RateLimitDescription{
		Status:    rlstatus,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   timestamppb.New(resetAt),
	}, nil
}

func IsJSONContentType(contentType string) bool {
	if !strings.HasPrefix(contentType, "application") {
		return false
	}

	if !strings.Contains(contentType, "json") {
		return false
	}

	return true
}

var xmlContentTypes []string = []string{
	"text/xml",
	"application/xml",
}

func IsXMLContentType(contentType string) bool {
	// there are some janky APIs out there
	normalizedContentType := strings.TrimSpace(strings.ToLower(contentType))

	for _, xmlContentType := range xmlContentTypes {
		if normalizedContentType == xmlContentType {
			return true
		}
	}
	return false
}
