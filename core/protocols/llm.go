package protocols

import (
	"context"

	"mantis/core/types"
)

type LLMMessage struct {
	Role       string
	Content    string
	ToolCalls  []types.ToolCall
	ToolCallID string
}

type LLM interface {
	ChatStream(ctx context.Context, baseURL, apiKey string, messages []LLMMessage, model string, tools []types.Tool, thinkingMode string) (<-chan types.StreamEvent, error)
}
