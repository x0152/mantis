package agent

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const defaultMaxIterations = 10

type LoopInput struct {
	ActionInput
	MaxIterations int
	MessageID     string
}

type AgentLoop struct {
	action *AgentAction
}

func NewAgentLoop(action *AgentAction) *AgentLoop {
	return &AgentLoop{action: action}
}

func (l *AgentLoop) Execute(ctx context.Context, in LoopInput) (<-chan types.StreamEvent, error) {
	maxIter := in.MaxIterations
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
	}

	toolMap := map[string]types.Tool{}
	for _, t := range in.Tools {
		toolMap[t.Name] = t
	}

	ch := make(chan types.StreamEvent, 32)
	go func() {
		defer close(ch)

		messages := make([]protocols.LLMMessage, len(in.Messages))
		copy(messages, in.Messages)

		for iter := 0; iter < maxIter; iter++ {
			actionCh, err := l.action.Execute(ctx, ActionInput{
				BaseURL: in.BaseURL, APIKey: in.APIKey,
				Model: in.Model, Messages: messages, Tools: in.Tools,
				ThinkingMode: in.ThinkingMode,
			})
			if err != nil {
				ch <- types.StreamEvent{Type: "error", Delta: err.Error(), Iteration: iter, IsFinal: true}
				return
			}

			var reply strings.Builder
			var toolCalls []types.ToolCall

			for event := range actionCh {
				event.Iteration = iter
				switch event.Type {
				case "text":
					reply.WriteString(event.Delta)
					ch <- event
				case "thinking":
					ch <- event
				case "tool_calls":
					toolCalls = event.ToolCalls
				case "error":
					ch <- event
					return
				}
			}

		if len(toolCalls) == 0 {
			return
		}

		messages = append(messages, protocols.LLMMessage{
			Role:      "assistant",
			Content:   reply.String(),
			ToolCalls: toolCalls,
		})

			for _, tc := range toolCalls {
				tool, ok := toolMap[tc.Name]
				if !ok {
					messages = append(messages, protocols.LLMMessage{
						Role: "tool", ToolCallID: tc.ID,
						Content: "error: unknown tool " + tc.Name,
					})
					continue
				}

				stepID := uuid.New().String()
				label := tc.Name
				if tool.Label != nil {
					label = tool.Label(tc.Arguments)
				}

				step := types.Step{
					ID: stepID, Tool: tc.Name, Label: label, Icon: tool.Icon,
					Args: tc.Arguments, Status: "running",
					StartedAt: time.Now().UTC().Format(time.RFC3339),
				}
				stepJSON, _ := json.Marshal(step)
				ch <- types.StreamEvent{Type: "tool_start", Delta: string(stepJSON), ToolID: stepID, Iteration: iter}

			toolCtx := shared.ContextWithStep(ctx, stepID, in.MessageID)

			type toolResult struct {
				result string
				err    error
			}
			resCh := make(chan toolResult, 1)
			toolDone := make(chan struct{})
			go func() {
				r, e := tool.Execute(toolCtx, tc.Arguments)
				close(toolDone)
				resCh <- toolResult{r, e}
			}()

			go func() {
				ticker := time.NewTicker(50 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-toolDone:
						return
					case <-ticker.C:
						if meta := shared.ToolMetaFromContext(toolCtx); meta != nil && meta.LogID != "" {
							ch <- types.StreamEvent{Type: "tool_meta", ToolID: stepID, LogID: meta.LogID, ModelName: meta.ModelName, Iteration: iter}
							return
						}
					}
				}
			}()

			res := <-resCh
			result := res.result
			if res.err != nil {
				result = "error: " + res.err.Error()
			}

			ev := types.StreamEvent{Type: "tool_end", Delta: result, ToolID: stepID, Iteration: iter}
			if meta := shared.ToolMetaFromContext(toolCtx); meta != nil {
				ev.LogID = meta.LogID
				ev.ModelName = meta.ModelName
			}
			ch <- ev

				messages = append(messages, protocols.LLMMessage{
					Role: "tool", ToolCallID: tc.ID, Content: result,
				})
			}
		}

		ch <- types.StreamEvent{Type: "error", Delta: "max iterations reached", IsFinal: true}
	}()

	return ch, nil
}
