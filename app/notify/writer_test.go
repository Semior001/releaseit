package notify

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriterNotifier_String(t *testing.T) {
	assert.Equal(t, "writer to stdout", (&WriterNotifier{Name: "stdout"}).String())
}

func TestWriterNotifier_Send(t *testing.T) {
	buf := &strings.Builder{}
	err := (&WriterNotifier{Writer: buf}).Send(context.Background(), "text")
	require.NoError(t, err)
	assert.Equal(t, "text", buf.String())
}
