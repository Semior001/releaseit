// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package notify

import (
	"context"
	"sync"
)

// Ensure, that DestinationMock does implement Destination.
// If this is not the case, regenerate this file with moq.
var _ Destination = &DestinationMock{}

// DestinationMock is a mock implementation of Destination.
//
// 	func TestSomethingThatUsesDestination(t *testing.T) {
//
// 		// make and configure a mocked Destination
// 		mockedDestination := &DestinationMock{
// 			SendFunc: func(ctx context.Context, text string) error {
// 				panic("mock out the Send method")
// 			},
// 			StringFunc: func() string {
// 				panic("mock out the String method")
// 			},
// 		}
//
// 		// use mockedDestination in code that requires Destination
// 		// and then make assertions.
//
// 	}
type DestinationMock struct {
	// SendFunc mocks the Send method.
	SendFunc func(ctx context.Context, text string) error

	// StringFunc mocks the String method.
	StringFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Send holds details about calls to the Send method.
		Send []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Text is the text argument value.
			Text string
		}
		// String holds details about calls to the String method.
		String []struct {
		}
	}
	lockSend   sync.RWMutex
	lockString sync.RWMutex
}

// Send calls SendFunc.
func (mock *DestinationMock) Send(ctx context.Context, text string) error {
	if mock.SendFunc == nil {
		panic("DestinationMock.SendFunc: method is nil but Destination.Send was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Text string
	}{
		Ctx:  ctx,
		Text: text,
	}
	mock.lockSend.Lock()
	mock.calls.Send = append(mock.calls.Send, callInfo)
	mock.lockSend.Unlock()
	return mock.SendFunc(ctx, text)
}

// SendCalls gets all the calls that were made to Send.
// Check the length with:
//     len(mockedDestination.SendCalls())
func (mock *DestinationMock) SendCalls() []struct {
	Ctx  context.Context
	Text string
} {
	var calls []struct {
		Ctx  context.Context
		Text string
	}
	mock.lockSend.RLock()
	calls = mock.calls.Send
	mock.lockSend.RUnlock()
	return calls
}

// String calls StringFunc.
func (mock *DestinationMock) String() string {
	if mock.StringFunc == nil {
		panic("DestinationMock.StringFunc: method is nil but Destination.String was just called")
	}
	callInfo := struct {
	}{}
	mock.lockString.Lock()
	mock.calls.String = append(mock.calls.String, callInfo)
	mock.lockString.Unlock()
	return mock.StringFunc()
}

// StringCalls gets all the calls that were made to String.
// Check the length with:
//     len(mockedDestination.StringCalls())
func (mock *DestinationMock) StringCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockString.RLock()
	calls = mock.calls.String
	mock.lockString.RUnlock()
	return calls
}
