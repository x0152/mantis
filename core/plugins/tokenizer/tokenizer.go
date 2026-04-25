package tokenizer

import (
	"regexp"
	"strings"
	"sync"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Tokenizer interface {
	Count(text string) int
	CountMessages(msgs []protocols.LLMMessage) int
	CountChatMessage(msg types.ChatMessage) int
	CountChatMessages(msgs []types.ChatMessage) int
	Family() string
}

type CharEstimator struct {
	Name            string
	CharsPerToken   float64
	MessageOverhead int
	ToolCallOvhd    int
	ToolResultOvhd  int
}

func (t *CharEstimator) Family() string { return t.Name }

func (t *CharEstimator) Count(s string) int {
	if s == "" {
		return 0
	}
	cpt := t.CharsPerToken
	if cpt <= 0 {
		cpt = 4
	}
	n := float64(len([]rune(s))) / cpt
	if n < 1 {
		return 1
	}
	return int(n + 0.5)
}

func (t *CharEstimator) CountMessages(msgs []protocols.LLMMessage) int {
	total := 0
	for _, m := range msgs {
		total += t.Count(m.Role)
		total += t.Count(m.Content)
		total += t.MessageOverhead
		for _, tc := range m.ToolCalls {
			total += t.Count(tc.Name) + t.Count(tc.Arguments) + t.ToolCallOvhd
		}
		if m.ToolCallID != "" {
			total += t.Count(m.ToolCallID) + t.ToolResultOvhd
		}
	}
	return total
}

func (t *CharEstimator) CountChatMessage(m types.ChatMessage) int {
	total := t.Count(m.Content) + t.MessageOverhead
	if len(m.Steps) > 0 {
		total += t.Count(string(m.Steps))
	}
	return total
}

func (t *CharEstimator) CountChatMessages(msgs []types.ChatMessage) int {
	total := 0
	for _, m := range msgs {
		total += t.CountChatMessage(m)
	}
	return total
}

func NewCharEstimator() *CharEstimator {
	return &CharEstimator{
		Name:            "default",
		CharsPerToken:   4,
		MessageOverhead: 4,
		ToolCallOvhd:    4,
		ToolResultOvhd:  2,
	}
}

type entry struct {
	pattern *regexp.Regexp
	factory func() Tokenizer
}

type Registry struct {
	mu       sync.RWMutex
	entries  []entry
	fallback Tokenizer
}

func NewRegistry() *Registry {
	return &Registry{fallback: NewCharEstimator()}
}

func (r *Registry) Register(pattern *regexp.Regexp, factory func() Tokenizer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry{pattern: pattern, factory: factory})
}

func (r *Registry) SetFallback(t Tokenizer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = t
}

func (r *Registry) For(modelName string) Tokenizer {
	name := strings.ToLower(strings.TrimSpace(modelName))
	r.mu.RLock()
	defer r.mu.RUnlock()
	if name != "" {
		for _, e := range r.entries {
			if e.pattern.MatchString(name) {
				return e.factory()
			}
		}
	}
	return r.fallback
}

func (r *Registry) Default() Tokenizer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.fallback
}

var (
	globalOnce sync.Once
	global     *Registry
)

func ensureGlobal() *Registry {
	globalOnce.Do(func() {
		global = NewRegistry()
		registerBuiltins(global)
	})
	return global
}

func Global() *Registry         { return ensureGlobal() }
func For(name string) Tokenizer { return ensureGlobal().For(name) }
func Default() Tokenizer        { return ensureGlobal().Default() }
func Register(pattern *regexp.Regexp, factory func() Tokenizer) {
	ensureGlobal().Register(pattern, factory)
}

func registerBuiltins(r *Registry) {
	r.Register(regexp.MustCompile(`(?i)qwen|qwq`), func() Tokenizer {
		return &CharEstimator{Name: "qwen", CharsPerToken: 3.5, MessageOverhead: 4, ToolCallOvhd: 4, ToolResultOvhd: 2}
	})
	r.Register(regexp.MustCompile(`(?i)gpt|o1|o3|o4|openai`), func() Tokenizer {
		return &CharEstimator{Name: "gpt", CharsPerToken: 4, MessageOverhead: 4, ToolCallOvhd: 4, ToolResultOvhd: 2}
	})
	r.Register(regexp.MustCompile(`(?i)claude|anthropic`), func() Tokenizer {
		return &CharEstimator{Name: "claude", CharsPerToken: 3.8, MessageOverhead: 4, ToolCallOvhd: 4, ToolResultOvhd: 2}
	})
	r.Register(regexp.MustCompile(`(?i)llama|mistral|mixtral`), func() Tokenizer {
		return &CharEstimator{Name: "llama", CharsPerToken: 3.7, MessageOverhead: 4, ToolCallOvhd: 4, ToolResultOvhd: 2}
	})
	r.Register(regexp.MustCompile(`(?i)gemini|gemma|palm`), func() Tokenizer {
		return &CharEstimator{Name: "gemini", CharsPerToken: 4, MessageOverhead: 4, ToolCallOvhd: 4, ToolResultOvhd: 2}
	})
	r.Register(regexp.MustCompile(`(?i)deepseek`), func() Tokenizer {
		return &CharEstimator{Name: "deepseek", CharsPerToken: 3.6, MessageOverhead: 4, ToolCallOvhd: 4, ToolResultOvhd: 2}
	})
}
