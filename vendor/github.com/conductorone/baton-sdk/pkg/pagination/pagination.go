package pagination

import (
	"encoding/json"
	"fmt"
)

type StreamToken struct {
	Size   int
	Cursor string
}

type StreamState struct {
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"done"`
}

type Token struct {
	Size  int
	Token string
}

type PageState struct {
	Token          string `json:"token,omitempty"`
	ResourceTypeID string `json:"type,omitempty"`
	ResourceID     string `json:"id,omitempty"`
}

// Bag holds pagination states that can be serialized for use as page tokens. It acts as a stack that you can push and pop
// pagination operations from.
type Bag struct {
	states       []PageState
	currentState *PageState
}

type serializedPaginationBag struct {
	States       []PageState `json:"states"`
	CurrentState *PageState  `json:"current_state"`
}

func (pb *Bag) push(s PageState) {
	if pb.currentState == nil {
		pb.currentState = &s
		return
	}

	pb.states = append(pb.states, *pb.currentState)
	pb.currentState = &s
}

func (pb *Bag) pop() *PageState {
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
func (pb *Bag) Push(state PageState) {
	pb.push(state)
}

// Pop returns the current page action, and makes the top of the stack the current.
func (pb *Bag) Pop() *PageState {
	return pb.pop()
}

// Next pops the current token, and pushes a copy of it with an updated page token.
func (pb *Bag) Next(pageToken string) error {
	st := pb.pop()
	if st == nil {
		return fmt.Errorf("no active page state")
	}

	if pageToken != "" {
		newState := *st
		newState.Token = pageToken
		pb.push(newState)
	}

	return nil
}

// Next pops the current token, and pushes a copy of it with an updated page token.
func (pb *Bag) NextToken(pageToken string) (string, error) {
	// assume that `pb` was passed an empty token
	if pb.currentState == nil {
		pb.currentState = &PageState{
			Token: pageToken,
		}
		return pb.Marshal()
	}

	st := pb.pop()
	if st == nil {
		return "", fmt.Errorf("no active page state")
	}

	if pageToken != "" {
		newState := *st
		newState.Token = pageToken
		pb.push(newState)
	}

	return pb.Marshal()
}

// Current returns the current page state for the bag.
func (pb *Bag) Current() *PageState {
	if pb.currentState == nil {
		return nil
	}

	current := *pb.currentState
	return &current
}

// Unmarshal takes an input string and unmarshals it onto the state object.
func (pb *Bag) Unmarshal(input string) error {
	token := serializedPaginationBag{}

	if input != "" {
		err := json.Unmarshal([]byte(input), &token)
		if err != nil {
			return fmt.Errorf("page token corrupt: %w", err)
		}

		pb.states = token.States
		pb.currentState = token.CurrentState
	} else {
		pb.states = nil
		pb.currentState = nil
	}

	return nil
}

// Marshal returns a string encoding of the state object.
func (pb *Bag) Marshal() (string, error) {
	if pb.currentState == nil {
		return "", nil
	}

	data, err := json.Marshal(serializedPaginationBag{
		States:       pb.states,
		CurrentState: pb.currentState,
	})
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// PageToken returns the page token for the current state.
func (pb *Bag) PageToken() string {
	c := pb.Current()
	if c == nil {
		return ""
	}

	return c.Token
}

// ResourceTypeID returns the resource type id for the current state.
func (pb *Bag) ResourceTypeID() string {
	c := pb.Current()
	if c == nil {
		return ""
	}

	return c.ResourceTypeID
}

// ResourceID returns the resource ID for the current state.
func (pb *Bag) ResourceID() string {
	c := pb.Current()
	if c == nil {
		return ""
	}

	return c.ResourceID
}
