package notes

import (
	"bytes"
	"context"
	"reflect"
	"sort"
	"testing"
	"text/template"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/task"
	tengine "github.com/Semior001/releaseit/app/task/engine"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestEvalAddon_listTaskUsers(t *testing.T) {
	t.Run("no opts", func(t *testing.T) {
		addon := &EvalAddon{}

		fns, err := addon.Funcs(context.Background())
		require.NoError(t, err)

		tmpl, err := template.New("").
			Funcs(fns).
			Parse(`{{ listTaskUsers . }}`)
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		require.NoError(t, tmpl.Execute(buf, task.Ticket{
			Author:   task.User{Username: "user1"},
			Assignee: task.User{Username: "user2"},
			Watchers: []task.User{
				{Username: "user3"},
				{Username: "user4"},
			},
		}))

		assert.Equal(t, `author user1, assignee user2, watchers: user3, user4`, buf.String())
	})

	t.Run("all opts", func(t *testing.T) {
		addon := &EvalAddon{}

		fns, err := addon.Funcs(context.Background())
		require.NoError(t, err)

		tmpl, err := template.New("").
			Funcs(fns).
			Parse(`{{ listTaskUsers . "@" "автор" "исполнитель" "наблюдатели" }}`)
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		require.NoError(t, tmpl.Execute(buf, task.Ticket{
			Author:   task.User{Username: "user1"},
			Assignee: task.User{Username: "user2"},
			Watchers: []task.User{
				{Username: "user3"},
				{Username: "user4"},
			},
		}))

		assert.Equal(t, `автор @user1, исполнитель @user2, наблюдатели: @user3, @user4`, buf.String())
	})
}

func TestEvalAddon_brackets(t *testing.T) {
	t.Run("round", func(t *testing.T) {
		addon := &EvalAddon{}

		fns, err := addon.Funcs(context.Background())
		require.NoError(t, err)

		tmpl, err := template.New("").
			Funcs(fns).
			Parse(`{{ brackets "abacaba" }}`)
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		require.NoError(t, tmpl.Execute(buf, nil))

		assert.Equal(t, "(abacaba)", buf.String())
	})

	t.Run("square", func(t *testing.T) {
		addon := &EvalAddon{}

		fns, err := addon.Funcs(context.Background())
		require.NoError(t, err)

		tmpl, err := template.New("").
			Funcs(fns).
			Parse(`{{ brackets "abacaba" true }}`)
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		require.NoError(t, tmpl.Execute(buf, nil))

		assert.Equal(t, "[abacaba]", buf.String())
	})

	t.Run("empty", func(t *testing.T) {
		addon := &EvalAddon{}

		fns, err := addon.Funcs(context.Background())
		require.NoError(t, err)

		tmpl, err := template.New("").
			Funcs(fns).
			Parse(`{{ brackets "" true }}`)
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		require.NoError(t, tmpl.Execute(buf, nil))

		assert.Equal(t, "", buf.String())
	})
}

func TestEvalAddon_listPRs(t *testing.T) {
	addon := &EvalAddon{}

	fns, err := addon.Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").
		Funcs(fns).
		Parse(`{{ listPRs . }}`)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, []git.PullRequest{
		{Title: "PR1", URL: "https://pr1"},
		{Title: "PR2", URL: "https://pr2"},
	}))

	assert.Equal(t, "[PR1](https://pr1), [PR2](https://pr2)", buf.String())
}

func TestEvalAddon_mdTaskLink(t *testing.T) {
	addon := &EvalAddon{}

	fns, err := addon.Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").
		Funcs(fns).
		Parse(`{{ mdTaskLink . }}`)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, task.Ticket{ID: "TASK-1", URL: "https://task1"}))

	assert.Equal(t, "[TASK-1](https://task1)", buf.String())
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
