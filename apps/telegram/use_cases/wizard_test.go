package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"mantis/core/base"
)

type tgFakeUpdate struct {
	UpdateID int            `json:"update_id"`
	Message  *tgFakeMessage `json:"message,omitempty"`
}

type tgFakeMessage struct {
	Chat tgFakeChat `json:"chat"`
	From tgFakeUser `json:"from"`
	Text string     `json:"text"`
}

type tgFakeChat struct {
	ID int64 `json:"id"`
}

type tgFakeUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type fakeTelegram struct {
	mu        sync.Mutex
	updates   []tgFakeUpdate
	bot       tgFakeUser
	authErr   bool
	notFound  bool
	sentMsgs  []string
}

func newFakeTelegram(bot tgFakeUser) *fakeTelegram {
	return &fakeTelegram{bot: bot}
}

func (f *fakeTelegram) push(u tgFakeUpdate) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates = append(f.updates, u)
}

func (f *fakeTelegram) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		switch {
		case strings.HasSuffix(r.URL.Path, "/getMe"):
			if f.authErr {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if f.notFound {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"ok":false,"error_code":404,"description":"Not Found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]any{
					"id":         f.bot.ID,
					"username":   f.bot.Username,
					"first_name": f.bot.FirstName,
				},
			})
		case strings.Contains(r.URL.Path, "/getUpdates"):
			f.mu.Lock()
			out := make([]tgFakeUpdate, len(f.updates))
			copy(out, f.updates)
			f.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": out})
		case strings.HasSuffix(r.URL.Path, "/sendMessage"):
			body, _ := io.ReadAll(r.Body)
			f.mu.Lock()
			f.sentMsgs = append(f.sentMsgs, string(body))
			f.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		default:
			http.NotFound(w, r)
		}
	})
}

func newTestWizard(server *httptest.Server) *Wizard {
	w := NewWizard()
	w.baseURL = server.URL
	w.client = server.Client()
	return w
}

func TestWizardVerifyValidatesInput(t *testing.T) {
	w := NewWizard()
	if _, err := w.Verify(context.Background(), "  "); err == nil || !errors.Is(err, base.ErrValidation) {
		t.Fatalf("expected validation error for empty token, got %v", err)
	}
}

func TestWizardVerifyReturnsBotAndCode(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{ID: 1001, Username: "mantis_test_bot", FirstName: "Mantis"})
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	bot, err := w.Verify(context.Background(), "TOKEN")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if bot.ID != 1001 || bot.Username != "mantis_test_bot" || bot.Name != "Mantis" {
		t.Fatalf("unexpected bot info: %+v", bot)
	}
	if len(bot.Code) != wizardCodeLen {
		t.Fatalf("expected %d-digit code, got %q", wizardCodeLen, bot.Code)
	}
	for _, c := range bot.Code {
		if c < '0' || c > '9' {
			t.Fatalf("code must be digits only, got %q", bot.Code)
		}
	}
	wantLink := "https://t.me/mantis_test_bot"
	if bot.Link != wantLink {
		t.Fatalf("expected link %q, got %q", wantLink, bot.Link)
	}
	if !strings.HasPrefix(bot.DeepLink, wantLink+"?start=") {
		t.Fatalf("unexpected deep link %q", bot.DeepLink)
	}
}

func TestWizardVerifyIsIdempotentForSameToken(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{ID: 1, Username: "bot", FirstName: "Bot"})
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	first, err := w.Verify(context.Background(), "TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	second, err := w.Verify(context.Background(), "TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	if first.Code != second.Code {
		t.Fatalf("expected same code on repeated verify, got %q vs %q", first.Code, second.Code)
	}
}

func TestWizardVerifySurfacesUnauthorized(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{})
	fake.authErr = true
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	if _, err := w.Verify(context.Background(), "BAD"); err == nil || !errors.Is(err, base.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestWizardVerifySurfacesNotFoundAsValidation(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{})
	fake.notFound = true
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	_, err := w.Verify(context.Background(), "fake-token")
	if err == nil || !errors.Is(err, base.ErrValidation) {
		t.Fatalf("expected validation error for 404, got %v", err)
	}
	if !strings.Contains(err.Error(), "Not Found") && !strings.Contains(err.Error(), "bot not found") {
		t.Fatalf("expected human-readable description, got %q", err)
	}
}

func TestWizardStatusRequiresVerifyFirst(t *testing.T) {
	w := NewWizard()
	if _, err := w.Status(context.Background(), "TOKEN"); err == nil || !errors.Is(err, base.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestWizardStatusReturnsNilUntilCodeSeen(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{ID: 1, Username: "bot", FirstName: "Bot"})
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	if _, err := w.Verify(context.Background(), "TOKEN"); err != nil {
		t.Fatal(err)
	}

	fake.push(tgFakeUpdate{UpdateID: 1, Message: &tgFakeMessage{
		Chat: tgFakeChat{ID: 42},
		From: tgFakeUser{ID: 42, Username: "egor", FirstName: "Egor"},
		Text: "hello",
	}})

	user, err := w.Status(context.Background(), "TOKEN")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestWizardStatusLinksUserOnCodeMatch(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{ID: 1, Username: "bot", FirstName: "Bot"})
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	bot, err := w.Verify(context.Background(), "TOKEN")
	if err != nil {
		t.Fatal(err)
	}

	fake.push(tgFakeUpdate{UpdateID: 1, Message: &tgFakeMessage{
		Chat: tgFakeChat{ID: 42},
		From: tgFakeUser{ID: 42, Username: "egor", FirstName: "Egor", LastName: "I"},
		Text: bot.Code,
	}})

	user, err := w.Status(context.Background(), "TOKEN")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if user == nil || user.ID != 42 || user.Username != "egor" || user.Name != "Egor I" {
		t.Fatalf("unexpected linked user: %+v", user)
	}

	again, err := w.Status(context.Background(), "TOKEN")
	if err != nil {
		t.Fatalf("status failed second time: %v", err)
	}
	if again == nil || again.ID != 42 {
		t.Fatalf("expected same linked user on repeat, got %+v", again)
	}

	if len(fake.sentMsgs) == 0 || !strings.Contains(fake.sentMsgs[0], "Linked") {
		t.Fatalf("expected confirmation message, got %v", fake.sentMsgs)
	}
}

func TestWizardStatusLinksOnStartCommand(t *testing.T) {
	fake := newFakeTelegram(tgFakeUser{ID: 1, Username: "bot", FirstName: "Bot"})
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	w := newTestWizard(srv)
	bot, err := w.Verify(context.Background(), "TOKEN")
	if err != nil {
		t.Fatal(err)
	}

	fake.push(tgFakeUpdate{UpdateID: 1, Message: &tgFakeMessage{
		Chat: tgFakeChat{ID: 7},
		From: tgFakeUser{ID: 7, FirstName: "Anon"},
		Text: "/start " + bot.Code,
	}})

	user, err := w.Status(context.Background(), "TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil || user.ID != 7 {
		t.Fatalf("expected linked user via /start, got %+v", user)
	}
}

func TestMessageMatchesCode(t *testing.T) {
	cases := []struct {
		text string
		code string
		want bool
	}{
		{"123456", "123456", true},
		{" 123456 ", "123456", true},
		{"my code is 123456 thanks", "123456", true},
		{"/start 123456", "123456", true},
		{"/start@mantis_bot 123456", "123456", true},
		{"/start", "123456", false},
		{"123455", "123456", false},
		{"", "123456", false},
		{"hello", "123456", false},
	}
	for _, c := range cases {
		if got := messageMatchesCode(c.text, c.code); got != c.want {
			t.Errorf("messageMatchesCode(%q, %q) = %v, want %v", c.text, c.code, got, c.want)
		}
	}
}

func TestRandomCodeShape(t *testing.T) {
	for i := 0; i < 20; i++ {
		code := randomCode(wizardCodeLen)
		if len(code) != wizardCodeLen {
			t.Fatalf("expected len %d, got %d (%q)", wizardCodeLen, len(code), code)
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Fatalf("non-digit in code %q", code)
			}
		}
	}
}

func TestBuildWizardBotWithoutUsernameOmitsLink(t *testing.T) {
	got := buildWizardBot(wizardBot{ID: 5, Name: "anon"}, "111111")
	if got.Link != "" || got.DeepLink != "" {
		t.Fatalf("expected empty links for bot without username, got %+v", got)
	}
}
