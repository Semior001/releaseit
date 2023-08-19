package notes

import (
	"bytes"
	"context"
	"github.com/Semior001/releaseit/app/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
