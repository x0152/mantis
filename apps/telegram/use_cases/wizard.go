package usecases

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"mantis/core/base"
)

type WizardBot struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Link     string `json:"link"`
	Code     string `json:"code"`
	DeepLink string `json:"deepLink"`
}

type WizardUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name"`
}

type Wizard struct {
	client  *http.Client
	baseURL string
	mu      sync.Mutex
	state   map[string]*wizardState
}

const defaultTelegramBaseURL = "https://api.telegram.org"

type wizardState struct {
	bot       wizardBot
	code      string
	offset    int
	linked    *WizardUser
	updatedAt time.Time
}

type wizardBot struct {
	ID       int64
	Username string
	Name     string
}

const (
	wizardStateTTL = 30 * time.Minute
	wizardCodeLen  = 6
)

func NewWizard() *Wizard {
	return &Wizard{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultTelegramBaseURL,
		state:   make(map[string]*wizardState),
	}
}

func (w *Wizard) Verify(ctx context.Context, token string) (*WizardBot, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return nil, fmt.Errorf("%w: token is required", base.ErrValidation)
	}
	bot, err := w.fetchMe(ctx, t)
	if err != nil {
		return nil, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.gcLocked()
	st, ok := w.state[t]
	if !ok || st.code == "" {
		st = &wizardState{bot: bot, code: randomCode(wizardCodeLen), updatedAt: time.Now()}
		w.state[t] = st
	} else {
		st.bot = bot
		st.updatedAt = time.Now()
	}

	return buildWizardBot(bot, st.code), nil
}

func (w *Wizard) Status(ctx context.Context, token string) (*WizardUser, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return nil, fmt.Errorf("%w: token is required", base.ErrValidation)
	}

	w.mu.Lock()
	w.gcLocked()
	st, ok := w.state[t]
	if !ok {
		w.mu.Unlock()
		return nil, fmt.Errorf("%w: verify the token first", base.ErrValidation)
	}
	if st.linked != nil {
		linked := *st.linked
		w.mu.Unlock()
		return &linked, nil
	}
	offset := st.offset
	code := st.code
	w.mu.Unlock()

	updates, err := w.getUpdates(ctx, t, offset)
	if err != nil {
		return nil, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	st = w.state[t]
	if st == nil {
		return nil, nil
	}

	for _, u := range updates {
		if u.UpdateID >= st.offset {
			st.offset = u.UpdateID + 1
		}
		if u.Message == nil {
			continue
		}
		if !messageMatchesCode(u.Message.Text, code) {
			continue
		}
		from := u.Message.From
		usr := WizardUser{
			ID:       from.ID,
			Username: from.Username,
			Name:     strings.TrimSpace(from.FirstName + " " + from.LastName),
		}
		st.linked = &usr
		st.updatedAt = time.Now()
		_ = w.sendMessage(ctx, t, u.Message.Chat.ID, fmt.Sprintf(
			"Linked %s — Mantis can now reach you here.",
			fallback(usr.Name, "you"),
		))
		linked := usr
		return &linked, nil
	}
	st.updatedAt = time.Now()
	return nil, nil
}

func (w *Wizard) gcLocked() {
	for k, s := range w.state {
		if time.Since(s.updatedAt) > wizardStateTTL {
			delete(w.state, k)
		}
	}
}

func (w *Wizard) fetchMe(ctx context.Context, token string) (wizardBot, error) {
	body, err := w.tgGet(ctx, fmt.Sprintf("%s/bot%s/getMe", w.baseURL, token))
	if err != nil {
		return wizardBot{}, err
	}
	var resp struct {
		OK     bool `json:"ok"`
		Result struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return wizardBot{}, err
	}
	if !resp.OK {
		return wizardBot{}, fmt.Errorf("%w: %s", base.ErrValidation, resp.Description)
	}
	return wizardBot{
		ID:       resp.Result.ID,
		Username: resp.Result.Username,
		Name:     strings.TrimSpace(resp.Result.FirstName),
	}, nil
}

type wizUpdate struct {
	UpdateID int         `json:"update_id"`
	Message  *wizMessage `json:"message"`
}

type wizMessage struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"from"`
	Text string `json:"text"`
}

func (w *Wizard) getUpdates(ctx context.Context, token string, offset int) ([]wizUpdate, error) {
	u := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d&timeout=0&allowed_updates=%%5B%%22message%%22%%5D", w.baseURL, token, offset)
	body, err := w.tgGet(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		OK          bool        `json:"ok"`
		Result      []wizUpdate `json:"result"`
		Description string      `json:"description"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%w: %s", base.ErrValidation, resp.Description)
	}
	return resp.Result, nil
}

func (w *Wizard) sendMessage(ctx context.Context, token string, chatID int64, text string) error {
	payload, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text})
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/bot%s/sendMessage", w.baseURL, token), strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (w *Wizard) tgGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: could not reach Telegram (%s)", base.ErrValidation, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return body, nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("%w: invalid bot token", base.ErrValidation)
	}
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return nil, fmt.Errorf("%w: %s", base.ErrValidation, telegramDescription(body, resp.StatusCode))
	}
	return nil, fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
}

func telegramDescription(body []byte, status int) string {
	if status == http.StatusNotFound {
		return "bot not found — check the token"
	}
	if status == http.StatusConflict {
		return "this bot is already running elsewhere — stop the existing channel before re-linking, or remove it from the Hosts → Channels page"
	}
	var parsed struct {
		Description string `json:"description"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		desc := strings.TrimSpace(parsed.Description)
		if desc != "" {
			return desc
		}
	}
	return fmt.Sprintf("telegram rejected request (%d)", status)
}

func buildWizardBot(b wizardBot, code string) *WizardBot {
	out := &WizardBot{
		ID:       b.ID,
		Username: b.Username,
		Name:     b.Name,
		Code:     code,
	}
	if b.Username != "" {
		out.Link = "https://t.me/" + b.Username
		out.DeepLink = "https://t.me/" + b.Username + "?start=" + url.QueryEscape(code)
	}
	return out
}

func messageMatchesCode(text, code string) bool {
	clean := strings.TrimSpace(text)
	if clean == "" || code == "" {
		return false
	}
	if clean == code || strings.Contains(clean, code) {
		return true
	}
	if strings.HasPrefix(clean, "/start") {
		rest := strings.TrimSpace(strings.TrimPrefix(clean, "/start"))
		if at := strings.IndexByte(rest, ' '); at >= 0 {
			rest = strings.TrimSpace(rest[at+1:])
		} else if strings.HasPrefix(rest, "@") {
			if sp := strings.IndexByte(rest, ' '); sp >= 0 {
				rest = strings.TrimSpace(rest[sp+1:])
			} else {
				rest = ""
			}
		}
		return rest == code
	}
	return false
}

func randomCode(n int) string {
	if n <= 0 {
		n = 6
	}
	max := big.NewInt(10)
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			b.WriteByte('0')
			continue
		}
		b.WriteByte(byte('0' + v.Int64()))
	}
	return b.String()
}

func fallback(s, def string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}
