package agent

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type AgentInput struct {
	LoopInput
}

type Agent struct {
	loop *AgentLoop
}

func New(llm protocols.LLM) *Agent {
	action := NewAgentAction(llm)
	return &Agent{loop: NewAgentLoop(action)}
}

func (a *Agent) Execute(ctx context.Context, in AgentInput) (<-chan types.StreamEvent, error) {
	return a.loop.Execute(ctx, in.LoopInput)
}
