package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Mattermost sends messages to Mattermost via webhook.
type Mattermost struct {
	cl  *http.Client
	url string
}

// NewMattermost makes a new Mattermost notifier.
func NewMattermost(cl http.Client, url string) *Mattermost {
	return &Mattermost{cl: &cl, url: url}
}

// String returns the name of the notifier.
func (m *Mattermost) String() string {
	return fmt.Sprintf("mattermost hook at: %s", extractBaseURL(m.url))
}

// Send sends a message to Mattermost.
func (m *Mattermost) Send(ctx context.Context, _, text string) error {
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
