package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Semior001/releaseit/app/store"
	"github.com/Semior001/releaseit/app/store/service"
)

// Telegram implements Destination to send changelogs to specified
// telegram chats.
type Telegram struct {
	TelegramParams
}

// TelegramParams defines parameters needed to initialize Telegram notifier.
type TelegramParams struct {
	ReleaseNotesBuilder service.ReleaseNotesBuilder
	Log                 *log.Logger

	ChatID                string
	Pin                   bool
	Client                *http.Client
	Timeout               time.Duration
	Token                 string
	DisableWebPagePreview bool

	apiBaseURL string // changed only in tests
}

const telegramAPIBaseURL = "https://api.telegram.org/bot"

// NewTelegram makes telegram bot for notifications
func NewTelegram(params TelegramParams) *Telegram {
	res := Telegram{TelegramParams: params}

	if res.Log == nil {
		res.Log = log.Default()
	}
	if res.apiBaseURL == "" {
		res.apiBaseURL = telegramAPIBaseURL
	}
	if res.Client == nil {
		res.Client = &http.Client{Timeout: 5 * time.Second}
	}

	return &res
}

func (t *Telegram) sendMessage(ctx context.Context, msg string, chatID string) error {
	if _, err := strconv.ParseInt(chatID, 10, 64); err != nil {
		chatID = "@" + chatID // if chatID not a number enforce @ prefix
	}

	u := fmt.Sprintf("%s%s/sendMessage?chat_id=%s&parse_mode=Markdown&disable_web_page_preview=%t",
		t.apiBaseURL, t.Token, chatID, t.DisableWebPagePreview)
	r, err := http.NewRequest("POST", u, strings.NewReader(msg))
	if err != nil {
		return fmt.Errorf("make telegram request: %w", err)
	}
	r.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := t.Client.Do(r.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("get telegram response: %w", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			t.Log.Printf("[WARN] can't close request body, %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		tgErr := tgError{}
		if err = json.NewDecoder(resp.Body).Decode(&tgErr); err == nil {
			return fmt.Errorf("unexpected telegram API status code %d, error: %q",
				resp.StatusCode, tgErr.Description)
		}
		return fmt.Errorf("unexpected telegram API status code %d", resp.StatusCode)
	}

	tgResp := struct {
		OK bool `json:"ok"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("can't decode telegram response: %w", err)
	}

	if !tgResp.OK {
		return fmt.Errorf("unexpected telegram API response: %t", tgResp.OK)
	}

	return nil
}

// String returns the string representation to identify this notifier.
func (t *Telegram) String() string {
	return fmt.Sprintf("telegram to chatID %s", t.ChatID)
}

// Send changelog via Telegram.
func (t *Telegram) Send(ctx context.Context, changelog store.Changelog) error {
	text, err := t.ReleaseNotesBuilder.Build(changelog)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	if err := t.sendMessage(ctx, text, t.ChatID); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

type tgError struct {
	Description string `json:"description"`
}
