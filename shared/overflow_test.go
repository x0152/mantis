package shared

import (
	"errors"
	"testing"
)

func TestIsContextOverflow(t *testing.T) {
	cases := []struct {
		msg  string
		want bool
	}{
		{"This model's maximum context length is 128000 tokens", true},
		{"Error: context length exceeded", true},
		{"prompt is too long for this model", true},
		{"Request too large for model gpt-4", true},
		{"tokens_limit reached", true},
		{"Please reduce the length of the messages or completion", true},
		{"Please reduce your prompt", true},
		{"context window is 32k, but got 40k", true},
		{"context canceled", false},
		{"connection refused", false},
		{"429 Too Many Requests", false},
		{"", false},
	}
	for _, c := range cases {
		got := IsContextOverflow(errors.New(c.msg))
		if got != c.want {
			t.Errorf("IsContextOverflow(%q) = %v, want %v", c.msg, got, c.want)
		}
	}
	if IsContextOverflow(nil) {
		t.Errorf("IsContextOverflow(nil) should be false")
	}
}
