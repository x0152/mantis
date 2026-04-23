package llm

import (
	"context"
	"fmt"
	"strings"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Router struct {
	defaultProvider string
	providers       map[string]protocols.LLM
}

func NewRouter(defaultProvider string, providers map[string]protocols.LLM) *Router {
	normalized := make(map[string]protocols.LLM, len(providers))
	for provider, adapter := range providers {
		key := strings.ToLower(strings.TrimSpace(provider))
		if key == "" || adapter == nil {
			continue
		}
		normalized[key] = adapter
	}
	fallback := strings.ToLower(strings.TrimSpace(defaultProvider))
	if fallback == "" {
		fallback = "openai"
	}
	return &Router{defaultProvider: fallback, providers: normalized}
}

func (r *Router) ChatStream(ctx context.Context, provider, baseURL, apiKey string, messages []protocols.LLMMessage, model string, tools []types.Tool, thinkingMode string) (<-chan types.StreamEvent, error) {
	key := strings.ToLower(strings.TrimSpace(provider))
	if key == "" {
		key = r.defaultProvider
	}

	adapter, ok := r.providers[key]
	if !ok {
		adapter, ok = r.providers[r.defaultProvider]
		if !ok {
			return nil, fmt.Errorf("LLM provider %q is not configured", key)
		}
		key = r.defaultProvider
	}

	return adapter.ChatStream(ctx, key, baseURL, apiKey, messages, model, tools, thinkingMode)
}
