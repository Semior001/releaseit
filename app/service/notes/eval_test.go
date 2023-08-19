package notes

import (
	"bytes"
	"context"
	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/task"
	tengine "github.com/Semior001/releaseit/app/task/engine"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"sort"
	"testing"
	"text/template"
)

func TestEvalAddon_treeTickets(t *testing.T) {
	addon := &EvalAddon{}

	fns, err := addon.Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").
		Funcs(fns).
		Parse(testData(t, "tree.gotmpl"))
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, []task.Ticket{
		{ID: "1", ParentID: ""},
		{ID: "2", ParentID: "1"},
		{ID: "3", ParentID: "1"},
		{ID: "4", ParentID: "2"},
		{ID: "5", ParentID: "2"},
		{ID: "6"},
		{ID: "7", ParentID: "6"},
	}))

	// 1
	// ├── 2
	// │   ├── 4
	// │   └── 5
	// └── 3
	// 6
	// └── 7

	assert.Equal(t, testData(t, "tree.txt"), buf.String())
}

func TestEvalAddon_loadTicketTree(t *testing.T) {
	addon := &EvalAddon{TaskTracker: &tengine.Tracker{
		Interface: &tengine.InterfaceMock{
			ListFunc: func(ctx context.Context, ids []string) ([]task.Ticket, error) {
				sort.Strings(ids)
				switch {
				case reflect.DeepEqual(ids, []string{"TASK-1", "TASK-5"}):
					return []task.Ticket{
						{ID: "TASK-1"},
						{ID: "TASK-5", ParentID: "TASK-2"},
					}, nil
				case reflect.DeepEqual(ids, []string{"TASK-2"}):
					return []task.Ticket{{ID: "TASK-2", ParentID: "TASK-3"}}, nil
				case reflect.DeepEqual(ids, []string{"TASK-3"}):
					return []task.Ticket{{ID: "TASK-3"}}, nil
				default:
					require.FailNow(t, "unexpected ids", "ids: %v", ids)
					return nil, nil
				}
			},
		},
	}}

	fns, err := addon.Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").
		Funcs(lo.Assign(fns, template.FuncMap{
			"prTitles": func(prs []git.PullRequest) []string {
				titles := make([]string, len(prs))
				for i, pr := range prs {
					titles[i] = pr.Title
				}
				return titles
			},
			"sort": func(strs []string) []string {
				sort.Strings(strs)
				return strs
			},
		})).
		Parse(testData(t, "load-tree.gotmpl"))
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, []git.PullRequest{
		{Title: "unattached PR"},
		{Title: "[TASK-1] some real deal"},
		{Title: "[TASK-5] something else"},
		{Title: "another one unattached"},
	}))

	assert.Equal(t, testData(t, "load-tree.txt"), buf.String())
}
