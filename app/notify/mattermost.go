package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
)

// MattermostBot is a notifier for MattermostBot.
type MattermostBot struct {
	MattermostBotParams
	token struct {
		value string
		mu    *sync.Mutex
	}
	cl *http.Client
}

// MattermostBotParams are MattermostBot notifier params.
type MattermostBotParams struct {
	Client    http.Client
	BaseURL   string
	ChannelID string

	LoginID  string
	Password string
	LDAP     bool
}

// NewMattermostBot makes a new MattermostBot notifier.
func NewMattermostBot(params MattermostBotParams) (*MattermostBot, error) {
	params.BaseURL = strings.TrimRight(params.BaseURL, "/")

	svc := &MattermostBot{MattermostBotParams: params}
	svc.token.mu = &sync.Mutex{}

	svc.cl = requester.New(svc.Client, svc.reloginMiddleware).Client()

	if err := svc.ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	if _, err := svc.login(context.Background()); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	return svc, nil
}

func (m *MattermostBot) login(ctx context.Context) (token string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u := fmt.Sprintf("%s/api/v4/users/login", m.BaseURL)

	reqBody := map[string]string{"login_id": m.LoginID, "password": m.Password}
	if m.LDAP {
		reqBody["ldap_only"] = "true"
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read body: %w", err)
		}
		return "", fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, body)
	}

	m.token.mu.Lock()
	defer m.token.mu.Unlock()
	m.token.value = resp.Header.Get("token")

	return m.token.value, nil
}

func (m *MattermostBot) ping(ctx context.Context) error {
	u := fmt.Sprintf("%s/api/v4/config/client?format=old", m.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// String returns the name of the notifier.
func (m *MattermostBot) String() string {
	return fmt.Sprintf("mattermost bot at: %s", m.BaseURL)
}

// Send sends a message to Mattermost.
func (m *MattermostBot) Send(ctx context.Context, _, text string) error {
	u := fmt.Sprintf("%s/api/v4/posts", m.BaseURL)

	b, err := json.Marshal(map[string]string{"channel_id": m.ChannelID, "message": text})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := m.cl.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (m *MattermostBot) reloginMiddleware(next http.RoundTripper) http.RoundTripper {
	return middleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			token, err := m.login(req.Context())
			if err != nil {
				return nil, fmt.Errorf("received 401, auth failed: %w", err)
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			return next.RoundTrip(req)
		}

		return resp, nil
	})
}

// MattermostHook sends messages to Mattermost via webhook.
type MattermostHook struct {
	cl      *http.Client
	baseURL string
	hookID  string
}

// NewMattermostHook makes a new MattermostHook notifier.
func NewMattermostHook(cl http.Client, baseURL, hookID string) *MattermostHook {
	return &MattermostHook{cl: &cl, baseURL: baseURL, hookID: hookID}
}

// String returns the name of the notifier.
func (m *MattermostHook) String() string {
	return fmt.Sprintf("mattermost hook at: %s", m.baseURL)
}

// Send sends a message to Mattermost.
func (m *MattermostHook) Send(ctx context.Context, _, text string) error {
	u := fmt.Sprintf("%s/hooks/%s", m.baseURL, m.hookID)

	b, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
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
