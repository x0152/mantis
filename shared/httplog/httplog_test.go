package httplog

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func captureLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	}()
	fn()
	return buf.String()
}

func runWithStatus(t *testing.T, status int, body string) string {
	t.Helper()
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if status != 0 {
			w.WriteHeader(status)
		}
		_, _ = w.Write([]byte(body))
	}))
	return captureLogs(t, func() {
		req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	})
}

func TestMiddleware_LogsServerError(t *testing.T) {
	out := runWithStatus(t, http.StatusInternalServerError, `{"detail":"boom"}`)
	if !strings.Contains(out, "ERROR") || !strings.Contains(out, "/api/test") || !strings.Contains(out, `"detail":"boom"`) {
		t.Fatalf("expected error log with detail, got %q", out)
	}
}

func TestMiddleware_LogsClientError(t *testing.T) {
	out := runWithStatus(t, http.StatusUnprocessableEntity, `{"detail":"bad"}`)
	if !strings.Contains(out, "422") || !strings.Contains(out, `"detail":"bad"`) {
		t.Fatalf("expected client error log, got %q", out)
	}
	if strings.Contains(out, "ERROR") {
		t.Fatalf("4xx should not be tagged ERROR, got %q", out)
	}
}

func TestMiddleware_StaysQuietOnSuccess(t *testing.T) {
	out := runWithStatus(t, http.StatusOK, "ok")
	if out != "" {
		t.Fatalf("expected no log for 2xx, got %q", out)
	}
}

func TestMiddleware_TruncatesLargeBodies(t *testing.T) {
	huge := strings.Repeat("x", 10000)
	out := runWithStatus(t, http.StatusInternalServerError, huge)
	if len(out) > 4096 {
		t.Fatalf("expected truncated body, got %d bytes", len(out))
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("expected ellipsis marker, got %q", out[:200])
	}
}
