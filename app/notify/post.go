package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Post sends a POST request to the given URL with release notes.
type Post struct {
	Log    *log.Logger
	URL    string
	Client *http.Client

	once sync.Once
}

// String returns the string representation of the notifier.
func (p *Post) String() string {
	return fmt.Sprintf("post to %s", extractBaseURL(p.URL))
}

// Send sends a POST request to the given URL with release notes.
func (p *Post) Send(ctx context.Context, text string) error {
	body, err := json.Marshal(map[string]interface{}{"text": text})
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
	defer func() {
		if err = resp.Body.Close(); err != nil {
			p.Log.Printf("[WARN] can't close request body, %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
