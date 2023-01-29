package notify

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriterNotifier_Send(t *testing.T) {
	buf := &strings.Builder{}
	err := (&WriterNotifier{Writer: buf}).Send(context.Background(), "tag", "text")
	require.NoError(t, err)
	assert.Equal(t, "text", buf.String())
}
