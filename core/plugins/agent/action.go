package agent

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ActionInput struct {
	BaseURL      string
	APIKey       string
	Model        string
	Messages     []protocols.LLMMessage
	Tools        []types.Tool
	ThinkingMode string
}

type AgentAction struct {
	llm protocols.LLM
}

func NewAgentAction(llm protocols.LLM) *AgentAction {
	return &AgentAction{llm: llm}
}

func (a *AgentAction) Execute(ctx context.Context, in ActionInput) (<-chan types.StreamEvent, error) {
	return a.llm.ChatStream(ctx, in.BaseURL, in.APIKey, in.Messages, in.Model, in.Tools, in.ThinkingMode)
}
