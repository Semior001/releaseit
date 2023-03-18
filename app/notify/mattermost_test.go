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
	err := svc.Send(context.Background(), "release notes")
	assert.NoError(t, err)
}

func TestMattermost_String(t *testing.T) {
	svc := NewMattermost(*http.DefaultClient, "https://example.com/hooks/123")
	assert.Equal(t, "mattermost hook at: https://example.com", svc.String())
}

func TestNewMattermostBot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))

		assert.Equal(t, "/api/v4/users/me", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "123",
		})
		require.NoError(t, err)
	}))
	defer ts.Close()

	svc, err := NewMattermostBot(*http.DefaultClient, ts.URL, "token", "channelID")
	require.NoError(t, err)
	assert.Equal(t, ts.URL, svc.url)
	assert.Equal(t, "123", svc.userID)
	assert.Equal(t, "channelID", svc.channelID)
}

func TestMattermostBot_String(t *testing.T) {
	svc := &MattermostBot{
		url:       "https://example.com",
		userID:    "123",
		channelID: "channelID",
	}
	assert.Equal(t, "mattermost bot 123 at: https://example.com, channel: channelID", svc.String())
}

func TestMattermostBot_Send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/posts", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body struct {
			ChannelID string `json:"channel_id"`
			Message   string `json:"message"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "channelID", body.ChannelID)
		assert.Equal(t, "release notes", body.Message)

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]interface{}{"id": "123"})
		require.NoError(t, err)
	}))
	defer ts.Close()

	svc := &MattermostBot{
		cl:        http.DefaultClient,
		url:       ts.URL,
		userID:    "",
		channelID: "channelID",
	}

	err := svc.Send(context.Background(), "release notes")
	assert.NoError(t, err)
}
