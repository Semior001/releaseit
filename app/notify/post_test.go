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

func TestPost_Send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			TagName string `json:"tag_name"`
			Text    string `json:"text"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "tag", body.TagName)
		assert.Equal(t, "text", body.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := (&Post{
		URL:    ts.URL,
		Client: http.DefaultClient,
	}).Send(context.Background(), "tag", "text")
	require.NoError(t, err)
}
