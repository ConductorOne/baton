// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: c1/connector/v2/annotation_etag.proto

package v2

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on ETag with the rules defined in the proto
// definition for this message. If any rules are violated, the first error
// encountered is returned, or nil if there are no violations.
func (m *ETag) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ETag with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ETagMultiError, or nil if none found.
func (m *ETag) ValidateAll() error {
	return m.validate(true)
}

func (m *ETag) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Value

	if len(errors) > 0 {
		return ETagMultiError(errors)
	}

	return nil
}

// ETagMultiError is an error wrapping multiple validation errors returned by
// ETag.ValidateAll() if the designated constraints aren't met.
type ETagMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ETagMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ETagMultiError) AllErrors() []error { return m }

// ETagValidationError is the validation error returned by ETag.Validate if the
// designated constraints aren't met.
type ETagValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ETagValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ETagValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ETagValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ETagValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ETagValidationError) ErrorName() string { return "ETagValidationError" }

// Error satisfies the builtin error interface
func (e ETagValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sETag.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ETagValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ETagValidationError{}
