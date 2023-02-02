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

func TestMattermost_Send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/hooks/123", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body struct {
			Text string `json:"text"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "release notes", body.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	svc := NewMattermost(*http.DefaultClient, ts.URL+"/hooks/123")
	err := svc.Send(context.Background(), "tag", "release notes")
	assert.NoError(t, err)
}

func TestMattermost_String(t *testing.T) {
	svc := NewMattermost(*http.DefaultClient, "https://example.com/hooks/123")
	assert.Equal(t, "mattermost hook at: https://example.com", svc.String())
}
