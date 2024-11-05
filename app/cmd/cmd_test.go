package cmd

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMattermostHookGroup_build(t *testing.T) {
	called := make([]int, 3)
	mu := sync.Mutex{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		idx, err := strconv.Atoi(r.URL.Query().Get("idx"))
		require.NoError(t, err)
		called[idx]++
	}))
	defer ts.Close()

	group := MattermostHookGroup{
		URL: []string{
			ts.URL + "?idx=0",
			ts.URL + "?idx=1",
			ts.URL + "?idx=2",
		},
		Timeout: 5 * time.Second,
	}

	dest, err := group.build()
	require.NoError(t, err)

	err = dest.Send(context.Background(), "test")
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, []int{1, 1, 1}, called)
}
