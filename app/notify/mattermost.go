package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
)

// Mattermost sends messages to Mattermost via webhook.
type Mattermost struct {
	log *log.Logger
	cl  *http.Client
	url string
}

// NewMattermost makes a new Mattermost notifier.
func NewMattermost(lg *log.Logger, cl http.Client, url string) *Mattermost {
	return &Mattermost{cl: &cl, url: url, log: lg}
}

// String returns the name of the notifier.
func (m *Mattermost) String() string {
	return fmt.Sprintf("mattermost hook at: %s", extractBaseURL(m.url))
}

// Send sends a message to Mattermost.
func (m *Mattermost) Send(ctx context.Context, text string) error {
	b, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := m.cl.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			m.log.Printf("[WARN] can't close request body, %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// MattermostBot sends messages to Mattermost via webhook.
type MattermostBot struct {
	cl        *http.Client
	log       *log.Logger
	baseURL   string
	userID    string
	channelID string
}

// NewMattermostBot makes a new Mattermost notifier.
func NewMattermostBot(lg *log.Logger, cl http.Client, baseURL, token, channelID string) (bot *MattermostBot, err error) {
	bot = &MattermostBot{
		cl: requester.New(cl,
			middleware.Header("Authorization", "Bearer "+token),
		).Client(),

		baseURL:   strings.TrimSuffix(baseURL, "/"), // remove trailing slash
		channelID: channelID,
		log:       lg,
	}

	if bot.userID, err = bot.me(context.Background()); err != nil {
		return nil, fmt.Errorf("get bot's userID: %w", err)
	}

	return bot, nil
}

func (b *MattermostBot) me(ctx context.Context) (userID string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL+"/api/v4/users/me", nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := b.cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			b.log.Printf("[WARN] can't close request body, %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var u struct {
		ID string `json:"id"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return u.ID, nil
}

// String returns the name of the notifier.
func (b *MattermostBot) String() string {
	return fmt.Sprintf("mattermost bot %.4s... channel: %.4s... at: %s",
		b.userID, b.channelID, extractBaseURL(b.baseURL))
}

// Send sends a message to Mattermost.
func (b *MattermostBot) Send(ctx context.Context, text string) error {
	bts, err := json.Marshal(map[string]string{
		"channel_id": b.channelID,
		"message":    text,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/api/v4/posts", bytes.NewReader(bts))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := b.cl.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			b.log.Printf("[WARN] can't close request body, %s", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var m struct {
		ID string `json:"id"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	b.log.Printf("[INFO] sent message %s to channel %s", m.ID, b.channelID)

	return nil
}
