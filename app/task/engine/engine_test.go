package engine

import (
	"context"
	"github.com/Semior001/releaseit/app/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestTracker_List(t *testing.T) {
	t.Run("load parents", func(t *testing.T) {
		svc := &Tracker{
			Interface: &InterfaceMock{
				ListFunc: func(ctx context.Context, ids []string) ([]task.Ticket, error) {
					switch {
					case reflect.DeepEqual(ids, []string{"1", "2"}):
						return []task.Ticket{
							{ID: "1", ParentID: "3"},
							{ID: "2", ParentID: "4"},
						}, nil
					case reflect.DeepEqual(ids, []string{"3", "4"}):
						return []task.Ticket{
							{ID: "3", ParentID: "5"},
							{ID: "4"},
						}, nil
					case reflect.DeepEqual(ids, []string{"5"}):
						return []task.Ticket{{ID: "5"}}, nil
					default:
						require.FailNow(t, "unexpected call to List with ids: %v", ids)
						return nil, nil
					}
				},
			},
		}

		tickets, err := svc.List(context.Background(), []string{"1", "2"}, true)
		require.NoError(t, err)
		assert.Equal(t, []task.Ticket{
			{ID: "1", ParentID: "3"},
			{ID: "2", ParentID: "4"},
			{ID: "3", ParentID: "5"},
			{ID: "4"},
			{ID: "5"},
		}, tickets)
	})

	t.Run("do not load parents", func(t *testing.T) {
		svc := &Tracker{
			Interface: &InterfaceMock{
				ListFunc: func(ctx context.Context, ids []string) ([]task.Ticket, error) {
					assert.Equal(t, []string{"1", "2"}, ids)
					return []task.Ticket{
						{ID: "1", ParentID: "3"},
						{ID: "2", ParentID: "4"},
					}, nil
				},
			},
		}

		tickets, err := svc.List(context.Background(), []string{"1", "2"}, false)
		require.NoError(t, err)
		assert.Equal(t, []task.Ticket{
			{ID: "1", ParentID: "3"},
			{ID: "2", ParentID: "4"},
		}, tickets)
	})
}

func TestUnsupported_List(t *testing.T) {
	res, err := Unsupported{}.List(nil, nil)
	assert.EqualError(t, err, "operation not supported")
	assert.Empty(t, res)
}

func TestUnsupported_Get(t *testing.T) {
	res, err := Unsupported{}.Get(nil, "")
	assert.EqualError(t, err, "operation not supported")
	assert.Empty(t, res)
}
