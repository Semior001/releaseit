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

func TestTelegram_String(t *testing.T) {
	assert.Equal(t, "telegram to chatID chat_id", NewTelegram(TelegramParams{ChatID: "chat_id"}).String())
}

func TestTelegram_Send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/bottoken/sendMessage", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "@chat_id", q.Get("chat_id"))
		assert.Equal(t, "Markdown", q.Get("parse_mode"))
		assert.Equal(t, "true", q.Get("disable_web_page_preview"))

		var msg tgMsg
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&msg))
		assert.Equal(t, "text", msg.Text)
		assert.Equal(t, "MarkdownV2", msg.ParseMode)

		_, err := w.Write([]byte(`{"ok": true}`))
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	svc := NewTelegram(TelegramParams{
		ChatID: "chat_id",
		Client: http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				// hijack the request to test server
				req.URL.Host = ts.URL[7:]
				req.URL.Scheme = "http"
				return http.DefaultTransport.RoundTrip(req)
			}),
		},
		Token:                 "token",
		DisableWebPagePreview: true,
	})

	err := svc.Send(context.Background(), "tag", "text")
	require.NoError(t, err)
}
