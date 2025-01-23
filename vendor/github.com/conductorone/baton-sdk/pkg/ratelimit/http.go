package ratelimit

import (
	"net/http"
	"strconv"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var limitHeaders = []string{
	"X-Ratelimit-Limit",
	"Ratelimit-Limit",
	"X-RateLimit-Requests-Limit", // Linear uses a non-standard header
	"X-Rate-Limit-Limit",         // Okta uses a non-standard header
}

var remainingHeaders = []string{
	"X-Ratelimit-Remaining",
	"Ratelimit-Remaining",
	"X-RateLimit-Requests-Remaining", // Linear uses a non-standard header
	"X-Rate-Limit-Remaining",         // Okta uses a non-standard header
}

var resetAtHeaders = []string{
	"X-Ratelimit-Reset",
	"Ratelimit-Reset",
	"X-RateLimit-Requests-Reset", // Linear uses a non-standard header
	"X-Rate-Limit-Reset",         // Okta uses a non-standard header
	"Retry-After",                // Often returned with 429
}

const thirtyYears = 60 * 60 * 24 * 365 * (2000 - 1970)

// Many APIs don't follow standards and return incorrect datetimes. This function tries to handle those cases.
func parseTime(timeStr string) (time.Time, error) {
	var t time.Time
	res, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		t, err = time.Parse(time.RFC850, timeStr)
		if err != nil {
			// Datetimes should be RFC850 but some APIs return RFC3339
			t, err = time.Parse(time.RFC3339, timeStr)
		}
		return t, err
	}

	// Times are supposed to be in seconds, but some APIs return milliseconds
	if res > thirtyYears*1000 {
		res /= 1000
	}

	// Times are supposed to be offsets, but some return absolute seconds since 1970.
	if res > thirtyYears {
		// If more than 30 years, it's probably an absolute timestamp
		t = time.Unix(res, 0)
	} else {
		// Otherwise, it's a relative timestamp
		t = time.Now().Add(time.Second * time.Duration(res))
	}

	return t, nil
}

func ExtractRateLimitData(statusCode int, header *http.Header) (*v2.RateLimitDescription, error) {
	if header == nil {
		return nil, nil
	}

	var rlstatus v2.RateLimitDescription_Status

	var limit int64
	var err error
	for _, limitHeader := range limitHeaders {
		limitStr := header.Get(limitHeader)
		if limitStr != "" {
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	var remaining int64
	for _, remainingHeader := range remainingHeaders {
		remainingStr := header.Get(remainingHeader)
		if remainingStr != "" {
			remaining, err = strconv.ParseInt(remainingStr, 10, 64)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if remaining > 0 {
		rlstatus = v2.RateLimitDescription_STATUS_OK
	}

	var resetAt time.Time
	for _, resetAtHeader := range resetAtHeaders {
		resetAtStr := header.Get(resetAtHeader)
		if resetAtStr != "" {
			resetAt, err = parseTime(resetAtStr)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if statusCode == http.StatusTooManyRequests {
		rlstatus = v2.RateLimitDescription_STATUS_OVERLIMIT
		remaining = 0
	}

	// If we didn't get any rate limit headers and status code is 429, return some sane defaults
	if remaining == 0 && resetAt.IsZero() && rlstatus == v2.RateLimitDescription_STATUS_OVERLIMIT {
		limit = 1
		resetAt = time.Now().Add(time.Second * 60)
	}

	return &v2.RateLimitDescription{
		Status:    rlstatus,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   timestamppb.New(resetAt),
	}, nil
}
