package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const gonkaTestPrivateKey = "1a30d0695812c21d6c6bfc59630c1753888c23fdbe63f897686c95f2924879d2"

func TestGonkaGetInferenceLimit_Balance(t *testing.T) {
	var hit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = r.URL.Path
		if !strings.HasPrefix(r.URL.Path, "/chain-api/cosmos/bank/v1beta1/balances/") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"balances":[{"denom":"ngonka","amount":"145230000000"},{"denom":"other","amount":"1"}]}`)
	}))
	defer server.Close()

	g := NewGonka()
	limit, err := g.GetInferenceLimit(context.Background(), server.URL, gonkaTestPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if limit.Type != "balance" {
		t.Fatalf("expected balance type, got %+v", limit)
	}
	if limit.Label != "Balance: 145.23 GNK" {
		t.Fatalf("unexpected label: %q", limit.Label)
	}
	if hit == "" {
		t.Fatalf("server not hit")
	}
}

func TestGonkaGetInferenceLimit_ZeroBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"balances":[]}`)
	}))
	defer server.Close()

	g := NewGonka()
	limit, err := g.GetInferenceLimit(context.Background(), server.URL, gonkaTestPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if limit.Type != "balance" || limit.Label != "Balance: 0 GNK" {
		t.Fatalf("unexpected limit: %+v", limit)
	}
}

func TestGonkaGetInferenceLimit_MissingPrivateKey(t *testing.T) {
	g := NewGonka()
	_, err := g.GetInferenceLimit(context.Background(), "http://node", "")
	if err == nil || !strings.Contains(err.Error(), "private key") {
		t.Fatalf("expected private key error, got %v", err)
	}
}

func TestGonkaGetInferenceLimit_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	g := NewGonka()
	_, err := g.GetInferenceLimit(context.Background(), server.URL, gonkaTestPrivateKey)
	if err == nil || !strings.Contains(err.Error(), "gonka chain API error 500") {
		t.Fatalf("expected chain API error, got %v", err)
	}
}
