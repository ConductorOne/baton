package annotations

import (
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Annotations []*anypb.Any

// Append appends the proto message to the annotations slice.
func (a *Annotations) Append(msgs ...proto.Message) {
	for _, msg := range msgs {
		madeAny, err := anypb.New(msg)
		if err != nil {
			panic(fmt.Errorf("failed to anypb.New: %w", err))
		}
		*a = append(*a, madeAny)
	}
}

// Update updates the annotations slice.
func (a *Annotations) Update(msg proto.Message) {
	var newAnnotations []*anypb.Any
	for _, v := range *a {
		if v.MessageIs(msg) {
			updatedAny, err := anypb.New(msg)
			if err != nil {
				panic(fmt.Errorf("failed to anypb.New: %w", err))
			}
			newAnnotations = append(newAnnotations, updatedAny)
		} else {
			newAnnotations = append(newAnnotations, v)
		}
	}
	*a = newAnnotations
}

// Contains checks if the message is in the annotations slice.
func (a *Annotations) Contains(msg proto.Message) bool {
	for _, v := range *a {
		if v.MessageIs(msg) {
			return true
		}
	}

	return false
}

// Pick checks if the message is in the annotations slice.
func (a *Annotations) Pick(needle proto.Message) (bool, error) {
	for _, v := range *a {
		if v.MessageIs(needle) {
			if err := v.UnmarshalTo(needle); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

// WithRateLimiting takes a pointer to a RateLimitDescription and appends it to the Annotations slice.
func (a *Annotations) WithRateLimiting(rateLimit *v2.RateLimitDescription) *Annotations {
	a.Append(rateLimit)
	return a
}
