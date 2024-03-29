package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPost_String(t *testing.T) {
	assert.Equal(t, "post to https://example.com", (&Post{URL: "https://example.com"}).String())
}

func TestPost_Send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "text", body.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := (&Post{
		URL:    ts.URL,
		Client: http.DefaultClient,
	}).Send(context.Background(), "text")
	require.NoError(t, err)
}
