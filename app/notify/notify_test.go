package notify

import (
	"context"
	"testing"

	"github.com/Semior001/releaseit/app/store"
	"github.com/stretchr/testify/assert"
)

type testDestination struct {
	name   string
	called bool
}

func (t testDestination) String() string {
	return t.name
}

func (t testDestination) Send(context.Context, store.Changelog) error {
	t.called = true
	return nil
}

func TestService_String(t *testing.T) {
	s := (&Service{
		dest: []Destination{
			testDestination{name: "test1"},
			testDestination{name: "test2"},
			testDestination{name: "test3"},
		},
	}).String()
	assert.Equal(t, "aggregated notifier with next notifiers: [test1 test2 test3]", s)
}
