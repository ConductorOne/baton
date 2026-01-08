package annotations

import (
	"fmt"

	c1zpb "github.com/conductorone/baton-sdk/pb/c1/c1z/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Annotations []*anypb.Any

// New - Convenience function to create annotations.
func New(msgs ...proto.Message) Annotations {
	annos := Annotations{}
	for _, msg := range msgs {
		annos.Update(msg)
	}

	return annos
}

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
	if msg == nil {
		return
	}

	var newAnnotations []*anypb.Any

	found := false
	for _, v := range *a {
		if v.MessageIs(msg) {
			updatedAny, err := anypb.New(msg)
			if err != nil {
				panic(fmt.Errorf("failed to anypb.New: %w", err))
			}
			newAnnotations = append(newAnnotations, updatedAny)
			found = true
		} else {
			newAnnotations = append(newAnnotations, v)
		}
	}

	// If we are trying to update a new message, just append it.
	if !found {
		v, err := anypb.New(msg)
		if err != nil {
			panic(fmt.Errorf("failed to anypb.New: %w", err))
		}
		newAnnotations = append(newAnnotations, v)
	}

	*a = newAnnotations
}

func (a *Annotations) Merge(newAnnotations ...*anypb.Any) {
	*a = append(*a, newAnnotations...)
}

// Contains checks if the message is in the annotations slice.
func (a *Annotations) Contains(msg proto.Message) bool {
	if msg == nil {
		return false
	}

	for _, v := range *a {
		if v.MessageIs(msg) {
			return true
		}
	}

	return false
}

func (a *Annotations) ContainsAny(msgs ...proto.Message) bool {
	if len(msgs) == 0 {
		return false
	}

	for _, v := range *a {
		for _, msg := range msgs {
			if v.MessageIs(msg) {
				return true
			}
		}
	}

	return false
}

// Pick checks if the message is in the annotations slice.
func (a *Annotations) Pick(needle proto.Message) (bool, error) {
	if needle == nil {
		return false, nil
	}

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
	a.Update(rateLimit)
	return a
}

// NOTE: the store is the only usage of this.
func GetSyncIdFromAnnotations(annos Annotations) (string, error) {
	syncDetails := &c1zpb.SyncDetails{}
	ok, err := annos.Pick(syncDetails)
	if err != nil {
		return "", err
	}
	if ok {
		return syncDetails.GetId(), nil
	}

	return "", nil
}
