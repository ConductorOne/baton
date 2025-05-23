// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: c1/connector/v2/resource.proto

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

// Validate checks the field values on ResourceType with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ResourceType) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ResourceType with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ResourceTypeMultiError, or
// nil if none found.
func (m *ResourceType) ValidateAll() error {
	return m.validate(true)
}

func (m *ResourceType) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := len(m.GetId()); l < 1 || l > 1024 {
		err := ResourceTypeValidationError{
			field:  "Id",
			reason: "value length must be between 1 and 1024 bytes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if m.GetDisplayName() != "" {

		if l := len(m.GetDisplayName()); l < 1 || l > 1024 {
			err := ResourceTypeValidationError{
				field:  "DisplayName",
				reason: "value length must be between 1 and 1024 bytes, inclusive",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	_ResourceType_Traits_Unique := make(map[ResourceType_Trait]struct{}, len(m.GetTraits()))

	for idx, item := range m.GetTraits() {
		_, _ = idx, item

		if _, exists := _ResourceType_Traits_Unique[item]; exists {
			err := ResourceTypeValidationError{
				field:  fmt.Sprintf("Traits[%v]", idx),
				reason: "repeated value must contain unique items",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		} else {
			_ResourceType_Traits_Unique[item] = struct{}{}
		}

		if _, ok := ResourceType_Trait_name[int32(item)]; !ok {
			err := ResourceTypeValidationError{
				field:  fmt.Sprintf("Traits[%v]", idx),
				reason: "value must be one of the defined enum values",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	for idx, item := range m.GetAnnotations() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourceTypeValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourceTypeValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourceTypeValidationError{
					field:  fmt.Sprintf("Annotations[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ResourceTypeMultiError(errors)
	}

	return nil
}

// ResourceTypeMultiError is an error wrapping multiple validation errors
// returned by ResourceType.ValidateAll() if the designated constraints aren't met.
type ResourceTypeMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourceTypeMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourceTypeMultiError) AllErrors() []error { return m }

// ResourceTypeValidationError is the validation error returned by
// ResourceType.Validate if the designated constraints aren't met.
type ResourceTypeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourceTypeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourceTypeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ResourceTypeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourceTypeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourceTypeValidationError) ErrorName() string { return "ResourceTypeValidationError" }

// Error satisfies the builtin error interface
func (e ResourceTypeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResourceType.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourceTypeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourceTypeValidationError{}

// Validate checks the field values on
// ResourceTypesServiceListResourceTypesRequest with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ResourceTypesServiceListResourceTypesRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on
// ResourceTypesServiceListResourceTypesRequest with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in
// ResourceTypesServiceListResourceTypesRequestMultiError, or nil if none found.
func (m *ResourceTypesServiceListResourceTypesRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ResourceTypesServiceListResourceTypesRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetParent()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ResourceTypesServiceListResourceTypesRequestValidationError{
					field:  "Parent",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ResourceTypesServiceListResourceTypesRequestValidationError{
					field:  "Parent",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetParent()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ResourceTypesServiceListResourceTypesRequestValidationError{
				field:  "Parent",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetPageSize() != 0 {

		if m.GetPageSize() > 250 {
			err := ResourceTypesServiceListResourceTypesRequestValidationError{
				field:  "PageSize",
				reason: "value must be less than or equal to 250",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	if m.GetPageToken() != "" {

		if l := len(m.GetPageToken()); l < 1 || l > 2048 {
			err := ResourceTypesServiceListResourceTypesRequestValidationError{
				field:  "PageToken",
				reason: "value length must be between 1 and 2048 bytes, inclusive",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	for idx, item := range m.GetAnnotations() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourceTypesServiceListResourceTypesRequestValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourceTypesServiceListResourceTypesRequestValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourceTypesServiceListResourceTypesRequestValidationError{
					field:  fmt.Sprintf("Annotations[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ResourceTypesServiceListResourceTypesRequestMultiError(errors)
	}

	return nil
}

// ResourceTypesServiceListResourceTypesRequestMultiError is an error wrapping
// multiple validation errors returned by
// ResourceTypesServiceListResourceTypesRequest.ValidateAll() if the
// designated constraints aren't met.
type ResourceTypesServiceListResourceTypesRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourceTypesServiceListResourceTypesRequestMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourceTypesServiceListResourceTypesRequestMultiError) AllErrors() []error { return m }

// ResourceTypesServiceListResourceTypesRequestValidationError is the
// validation error returned by
// ResourceTypesServiceListResourceTypesRequest.Validate if the designated
// constraints aren't met.
type ResourceTypesServiceListResourceTypesRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourceTypesServiceListResourceTypesRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourceTypesServiceListResourceTypesRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ResourceTypesServiceListResourceTypesRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourceTypesServiceListResourceTypesRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourceTypesServiceListResourceTypesRequestValidationError) ErrorName() string {
	return "ResourceTypesServiceListResourceTypesRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ResourceTypesServiceListResourceTypesRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResourceTypesServiceListResourceTypesRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourceTypesServiceListResourceTypesRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourceTypesServiceListResourceTypesRequestValidationError{}

// Validate checks the field values on
// ResourceTypesServiceListResourceTypesResponse with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ResourceTypesServiceListResourceTypesResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on
// ResourceTypesServiceListResourceTypesResponse with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in
// ResourceTypesServiceListResourceTypesResponseMultiError, or nil if none found.
func (m *ResourceTypesServiceListResourceTypesResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ResourceTypesServiceListResourceTypesResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetList() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourceTypesServiceListResourceTypesResponseValidationError{
						field:  fmt.Sprintf("List[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourceTypesServiceListResourceTypesResponseValidationError{
						field:  fmt.Sprintf("List[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourceTypesServiceListResourceTypesResponseValidationError{
					field:  fmt.Sprintf("List[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if m.GetNextPageToken() != "" {

		if l := len(m.GetNextPageToken()); l < 1 || l > 2048 {
			err := ResourceTypesServiceListResourceTypesResponseValidationError{
				field:  "NextPageToken",
				reason: "value length must be between 1 and 2048 bytes, inclusive",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	for idx, item := range m.GetAnnotations() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourceTypesServiceListResourceTypesResponseValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourceTypesServiceListResourceTypesResponseValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourceTypesServiceListResourceTypesResponseValidationError{
					field:  fmt.Sprintf("Annotations[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ResourceTypesServiceListResourceTypesResponseMultiError(errors)
	}

	return nil
}

// ResourceTypesServiceListResourceTypesResponseMultiError is an error wrapping
// multiple validation errors returned by
// ResourceTypesServiceListResourceTypesResponse.ValidateAll() if the
// designated constraints aren't met.
type ResourceTypesServiceListResourceTypesResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourceTypesServiceListResourceTypesResponseMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourceTypesServiceListResourceTypesResponseMultiError) AllErrors() []error { return m }

// ResourceTypesServiceListResourceTypesResponseValidationError is the
// validation error returned by
// ResourceTypesServiceListResourceTypesResponse.Validate if the designated
// constraints aren't met.
type ResourceTypesServiceListResourceTypesResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourceTypesServiceListResourceTypesResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourceTypesServiceListResourceTypesResponseValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e ResourceTypesServiceListResourceTypesResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourceTypesServiceListResourceTypesResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourceTypesServiceListResourceTypesResponseValidationError) ErrorName() string {
	return "ResourceTypesServiceListResourceTypesResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ResourceTypesServiceListResourceTypesResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResourceTypesServiceListResourceTypesResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourceTypesServiceListResourceTypesResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourceTypesServiceListResourceTypesResponseValidationError{}

// Validate checks the field values on ResourceId with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ResourceId) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ResourceId with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ResourceIdMultiError, or
// nil if none found.
func (m *ResourceId) ValidateAll() error {
	return m.validate(true)
}

func (m *ResourceId) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := len(m.GetResourceType()); l < 1 || l > 1024 {
		err := ResourceIdValidationError{
			field:  "ResourceType",
			reason: "value length must be between 1 and 1024 bytes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := len(m.GetResource()); l < 1 || l > 1024 {
		err := ResourceIdValidationError{
			field:  "Resource",
			reason: "value length must be between 1 and 1024 bytes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return ResourceIdMultiError(errors)
	}

	return nil
}

// ResourceIdMultiError is an error wrapping multiple validation errors
// returned by ResourceId.ValidateAll() if the designated constraints aren't met.
type ResourceIdMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourceIdMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourceIdMultiError) AllErrors() []error { return m }

// ResourceIdValidationError is the validation error returned by
// ResourceId.Validate if the designated constraints aren't met.
type ResourceIdValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourceIdValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourceIdValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ResourceIdValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourceIdValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourceIdValidationError) ErrorName() string { return "ResourceIdValidationError" }

// Error satisfies the builtin error interface
func (e ResourceIdValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResourceId.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourceIdValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourceIdValidationError{}

// Validate checks the field values on Resource with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Resource) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Resource with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ResourceMultiError, or nil
// if none found.
func (m *Resource) ValidateAll() error {
	return m.validate(true)
}

func (m *Resource) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetId()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ResourceValidationError{
					field:  "Id",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ResourceValidationError{
					field:  "Id",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetId()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ResourceValidationError{
				field:  "Id",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if all {
		switch v := interface{}(m.GetParentResourceId()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ResourceValidationError{
					field:  "ParentResourceId",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ResourceValidationError{
					field:  "ParentResourceId",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetParentResourceId()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ResourceValidationError{
				field:  "ParentResourceId",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetDisplayName() != "" {

		if l := len(m.GetDisplayName()); l < 1 || l > 1024 {
			err := ResourceValidationError{
				field:  "DisplayName",
				reason: "value length must be between 1 and 1024 bytes, inclusive",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	for idx, item := range m.GetAnnotations() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourceValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourceValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourceValidationError{
					field:  fmt.Sprintf("Annotations[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ResourceMultiError(errors)
	}

	return nil
}

// ResourceMultiError is an error wrapping multiple validation errors returned
// by Resource.ValidateAll() if the designated constraints aren't met.
type ResourceMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourceMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourceMultiError) AllErrors() []error { return m }

// ResourceValidationError is the validation error returned by
// Resource.Validate if the designated constraints aren't met.
type ResourceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ResourceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourceValidationError) ErrorName() string { return "ResourceValidationError" }

// Error satisfies the builtin error interface
func (e ResourceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResource.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourceValidationError{}

// Validate checks the field values on ResourcesServiceListResourcesRequest
// with the rules defined in the proto definition for this message. If any
// rules are violated, the first error encountered is returned, or nil if
// there are no violations.
func (m *ResourcesServiceListResourcesRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ResourcesServiceListResourcesRequest
// with the rules defined in the proto definition for this message. If any
// rules are violated, the result is a list of violation errors wrapped in
// ResourcesServiceListResourcesRequestMultiError, or nil if none found.
func (m *ResourcesServiceListResourcesRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ResourcesServiceListResourcesRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := len(m.GetResourceTypeId()); l < 1 || l > 1024 {
		err := ResourcesServiceListResourcesRequestValidationError{
			field:  "ResourceTypeId",
			reason: "value length must be between 1 and 1024 bytes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if all {
		switch v := interface{}(m.GetParentResourceId()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ResourcesServiceListResourcesRequestValidationError{
					field:  "ParentResourceId",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ResourcesServiceListResourcesRequestValidationError{
					field:  "ParentResourceId",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetParentResourceId()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ResourcesServiceListResourcesRequestValidationError{
				field:  "ParentResourceId",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetPageSize() != 0 {

		if m.GetPageSize() > 250 {
			err := ResourcesServiceListResourcesRequestValidationError{
				field:  "PageSize",
				reason: "value must be less than or equal to 250",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	if m.GetPageToken() != "" {

		if l := len(m.GetPageToken()); l < 1 || l > 2048 {
			err := ResourcesServiceListResourcesRequestValidationError{
				field:  "PageToken",
				reason: "value length must be between 1 and 2048 bytes, inclusive",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	for idx, item := range m.GetAnnotations() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourcesServiceListResourcesRequestValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourcesServiceListResourcesRequestValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourcesServiceListResourcesRequestValidationError{
					field:  fmt.Sprintf("Annotations[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ResourcesServiceListResourcesRequestMultiError(errors)
	}

	return nil
}

// ResourcesServiceListResourcesRequestMultiError is an error wrapping multiple
// validation errors returned by
// ResourcesServiceListResourcesRequest.ValidateAll() if the designated
// constraints aren't met.
type ResourcesServiceListResourcesRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourcesServiceListResourcesRequestMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourcesServiceListResourcesRequestMultiError) AllErrors() []error { return m }

// ResourcesServiceListResourcesRequestValidationError is the validation error
// returned by ResourcesServiceListResourcesRequest.Validate if the designated
// constraints aren't met.
type ResourcesServiceListResourcesRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourcesServiceListResourcesRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourcesServiceListResourcesRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ResourcesServiceListResourcesRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourcesServiceListResourcesRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourcesServiceListResourcesRequestValidationError) ErrorName() string {
	return "ResourcesServiceListResourcesRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ResourcesServiceListResourcesRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResourcesServiceListResourcesRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourcesServiceListResourcesRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourcesServiceListResourcesRequestValidationError{}

// Validate checks the field values on ResourcesServiceListResourcesResponse
// with the rules defined in the proto definition for this message. If any
// rules are violated, the first error encountered is returned, or nil if
// there are no violations.
func (m *ResourcesServiceListResourcesResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ResourcesServiceListResourcesResponse
// with the rules defined in the proto definition for this message. If any
// rules are violated, the result is a list of violation errors wrapped in
// ResourcesServiceListResourcesResponseMultiError, or nil if none found.
func (m *ResourcesServiceListResourcesResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ResourcesServiceListResourcesResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetList() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourcesServiceListResourcesResponseValidationError{
						field:  fmt.Sprintf("List[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourcesServiceListResourcesResponseValidationError{
						field:  fmt.Sprintf("List[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourcesServiceListResourcesResponseValidationError{
					field:  fmt.Sprintf("List[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if m.GetNextPageToken() != "" {

		if l := len(m.GetNextPageToken()); l < 1 || l > 2048 {
			err := ResourcesServiceListResourcesResponseValidationError{
				field:  "NextPageToken",
				reason: "value length must be between 1 and 2048 bytes, inclusive",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	for idx, item := range m.GetAnnotations() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ResourcesServiceListResourcesResponseValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ResourcesServiceListResourcesResponseValidationError{
						field:  fmt.Sprintf("Annotations[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ResourcesServiceListResourcesResponseValidationError{
					field:  fmt.Sprintf("Annotations[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ResourcesServiceListResourcesResponseMultiError(errors)
	}

	return nil
}

// ResourcesServiceListResourcesResponseMultiError is an error wrapping
// multiple validation errors returned by
// ResourcesServiceListResourcesResponse.ValidateAll() if the designated
// constraints aren't met.
type ResourcesServiceListResourcesResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ResourcesServiceListResourcesResponseMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ResourcesServiceListResourcesResponseMultiError) AllErrors() []error { return m }

// ResourcesServiceListResourcesResponseValidationError is the validation error
// returned by ResourcesServiceListResourcesResponse.Validate if the
// designated constraints aren't met.
type ResourcesServiceListResourcesResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ResourcesServiceListResourcesResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ResourcesServiceListResourcesResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ResourcesServiceListResourcesResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ResourcesServiceListResourcesResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ResourcesServiceListResourcesResponseValidationError) ErrorName() string {
	return "ResourcesServiceListResourcesResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ResourcesServiceListResourcesResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResourcesServiceListResourcesResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ResourcesServiceListResourcesResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ResourcesServiceListResourcesResponseValidationError{}
