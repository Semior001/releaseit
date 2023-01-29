package notify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDestinations_Send(t *testing.T) {
	mq := &DestinationMock{
		SendFunc: func(ctx context.Context, tagName string, text string) error { return nil },
	}
	mq1 := &DestinationMock{
		SendFunc: func(ctx context.Context, tagName string, text string) error { return nil },
	}

	dests := Destinations{mq, mq1}

	require.NoError(t, dests.Send(context.Background(), "v1.0.0", "release notes"))
	assert.Equal(t, 1, len(mq.SendCalls()))
	assert.Equal(t, 1, len(mq1.SendCalls()))

	assert.Equal(t, "v1.0.0", mq.SendCalls()[0].TagName)
	assert.Equal(t, "release notes", mq.SendCalls()[0].Text)

	assert.Equal(t, "v1.0.0", mq1.SendCalls()[0].TagName)
	assert.Equal(t, "release notes", mq1.SendCalls()[0].Text)
}
