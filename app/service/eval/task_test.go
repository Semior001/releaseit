package eval

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/task"
	"github.com/Semior001/releaseit/app/task/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestTask_getTicket(t *testing.T) {
	eng := &engine.Tracker{
		Interface: &engine.InterfaceMock{
			GetFunc: func(ctx context.Context, id string) (task.Ticket, error) {
				return task.Ticket{ID: "1", Name: "task"}, nil
			},
		},
	}

	res := execTaskTmpl(t, eng, `{{ getTicket "1" }}`, nil)
	assert.Equal(t, fmt.Sprintf("%v", task.Ticket{ID: "1", Name: "task"}), res)
}

func TestTask_listTickets(t *testing.T) {
	eng := &engine.Tracker{
		Interface: &engine.InterfaceMock{
			ListFunc: func(ctx context.Context, ids []string) ([]task.Ticket, error) {
				return []task.Ticket{
					{ID: "1", Name: "task"},
					{ID: "2", Name: "task-2"},
				}, nil
			},
		},
	}

	res := execTaskTmpl(t, eng, `{{ listTickets .List false }}`, struct{ List []string }{List: []string{"1", "2"}})
	assert.Equal(t, fmt.Sprintf("%v", []task.Ticket{
		{ID: "1", Name: "task"},
		{ID: "2", Name: "task-2"},
	}), res)
}

func execTaskTmpl(t *testing.T, eng *engine.Tracker, expr string, data any) string {
	fns, err := (&Task{Tracker: eng}).Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").Funcs(fns).Parse(expr)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, data))

	return buf.String()
}

func TestTask_String(t *testing.T) {
	assert.Equal(t, "task", (&Task{}).String())
}
