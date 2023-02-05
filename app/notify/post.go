package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Post sends a POST request to the given URL with release notes.
type Post struct {
	URL    string
	Extras map[string]string
	Client *http.Client
}

// String returns the string representation of the notifier.
func (p *Post) String() string {
	return fmt.Sprintf("post to %s", extractBaseURL(p.URL))
}

// Send sends a POST request to the given URL with release notes.
func (p *Post) Send(ctx context.Context, text string) error {
	body, err := json.Marshal(map[string]interface{}{
		"extras": p.Extras,
		"text":   text,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
