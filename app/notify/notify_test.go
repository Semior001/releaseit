package notify

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDestinations_String(t *testing.T) {
	assert.Equal(t, "[mock, mock1]", Destinations{
		&DestinationMock{StringFunc: func() string { return "mock" }},
		&DestinationMock{StringFunc: func() string { return "mock1" }},
	}.String())
}

func TestDestinations_Send(t *testing.T) {
	t.Run("without errors", func(t *testing.T) {
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
	})

	t.Run("with errors", func(t *testing.T) {
		err, err1 := errors.New("err0"), errors.New("err1")
		mq := &DestinationMock{
			SendFunc:   func(ctx context.Context, tagName string, text string) error { return err },
			StringFunc: func() string { return "mock0" },
		}
		mq1 := &DestinationMock{
			SendFunc:   func(ctx context.Context, tagName string, text string) error { return err1 },
			StringFunc: func() string { return "mock1" },
		}

		dests := Destinations{mq, mq1}

		errs := dests.Send(context.Background(), "v1.0.0", "release notes")
		var merr *multierror.Error
		require.ErrorAsf(t, errs, &merr, "expected multierror")
		assert.Equal(t, 2, len(merr.Errors))
		sort.Slice(merr.Errors, func(i, j int) bool {
			return merr.Errors[i].Error() < merr.Errors[j].Error()
		})
		assert.ErrorIs(t, merr.Errors[0], err)
		assert.ErrorIs(t, merr.Errors[1], err1)
	})
}
