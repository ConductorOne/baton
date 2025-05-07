package pagination

import (
	"encoding/json"
	"fmt"
)

var ErrNoToken = fmt.Errorf("pagination token cannot be nil")

type genericSerializedPaginationBag[T any] struct {
	States       []T `json:"states"`
	CurrentState *T  `json:"current_state"`
}

// GenBag holds pagination states that can be serialized for use as page tokens. It acts as a stack that you can push and pop
// pagination operations from.
// This is a generic version of the Bag struct.
type GenBag[T any] struct {
	states       []T
	currentState *T
}

func GenBagFromToken[T any](pToken *Token) (*GenBag[T], error) {
	if pToken == nil {
		return nil, ErrNoToken
	}

	bag := &GenBag[T]{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, err
	}
	return bag, nil
}

func (pb *GenBag[T]) push(s T) {
	if pb.currentState == nil {
		pb.currentState = &s
		return
	}

	pb.states = append(pb.states, *pb.currentState)
	pb.currentState = &s
}

func (pb *GenBag[T]) pop() *T {
	if pb.currentState == nil {
		return nil
	}

	ret := *pb.currentState

	if len(pb.states) > 0 {
		pb.currentState = &pb.states[len(pb.states)-1]
		pb.states = pb.states[:len(pb.states)-1]
	} else {
		pb.currentState = nil
	}

	return &ret
}

// Push pushes a new page state onto the stack.
func (pb *GenBag[T]) Push(state T) {
	pb.push(state)
}

// Pop returns the current page action, and makes the top of the stack the current.
func (pb *GenBag[T]) Pop() *T {
	return pb.pop()
}

// Current returns the current page state for the bag.
func (pb *GenBag[T]) Current() *T {
	if pb.currentState == nil {
		return nil
	}

	current := *pb.currentState
	return &current
}

// Unmarshal takes an input string and unmarshals it onto the state object.
func (pb *GenBag[T]) Unmarshal(input string) error {
	var target genericSerializedPaginationBag[T]

	if input != "" {
		err := json.Unmarshal([]byte(input), &target)
		if err != nil {
			return fmt.Errorf("page token corrupt: %w", err)
		}

		pb.states = target.States
		pb.currentState = target.CurrentState
	} else {
		pb.states = nil
		pb.currentState = nil
	}

	return nil
}

// Marshal returns a string encoding of the state object.
func (pb *GenBag[T]) Marshal() (string, error) {
	if pb.currentState == nil {
		return "", nil
	}

	data, err := json.Marshal(genericSerializedPaginationBag[T]{
		States:       pb.states,
		CurrentState: pb.currentState,
	})
	if err != nil {
		return "", err
	}

	return string(data), nil
}
