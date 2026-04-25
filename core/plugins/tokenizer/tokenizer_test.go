package tokenizer

import "testing"

func TestForSelectsByFamily(t *testing.T) {
	cases := map[string]string{
		"qwen/qwen3.5-35b-a3b":  "qwen",
		"qwen3.6-27b":           "qwen",
		"gpt-4o-mini":           "gpt",
		"o3-mini":               "gpt",
		"claude-sonnet-4":       "claude",
		"llama-3.1-70b":         "llama",
		"mistral-large":         "llama",
		"deepseek-v3":           "deepseek",
		"gemini-2.5-pro":        "gemini",
		"some-unknown-model":    "default",
		"":                      "default",
	}
	for name, want := range cases {
		got := For(name).Family()
		if got != want {
			t.Fatalf("For(%q): got %q, want %q", name, got, want)
		}
	}
}

func TestCountMatchesExpected(t *testing.T) {
	gpt := For("gpt-4o")
	qwen := For("qwen3.5-35b-a3b")
	s := "Hello, this is a piece of text used to check that tokenizers produce different estimates."
	gc := gpt.Count(s)
	qc := qwen.Count(s)
	if gc == 0 || qc == 0 {
		t.Fatalf("expected non-zero counts, got gpt=%d qwen=%d", gc, qc)
	}
	if gc == qc {
		t.Fatalf("expected different counts across families, got gpt=%d qwen=%d", gc, qc)
	}
}
