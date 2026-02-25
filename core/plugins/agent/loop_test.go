package agent

import (
	"context"
	"testing"

	"mantis/core/protocols"
	"mantis/core/types"
)

type scriptedLLM struct {
	streams [][]types.StreamEvent
	calls   int
}

func (s *scriptedLLM) ChatStream(_ context.Context, _ string, _ string, _ []protocols.LLMMessage, _ string, _ []types.Tool, _ string) (<-chan types.StreamEvent, error) {
	ch := make(chan types.StreamEvent, 8)
	idx := s.calls
	s.calls++
	go func() {
		if idx < len(s.streams) {
			for _, ev := range s.streams[idx] {
				ch <- ev
			}
		}
		close(ch)
	}()
	return ch, nil
}

func collect(ch <-chan types.StreamEvent) []types.StreamEvent {
	var out []types.StreamEvent
	for ev := range ch {
		out = append(out, ev)
	}
	return out
}

func TestAgentLoop_ExecutesToolAndStopsWithoutError(t *testing.T) {
	llm := &scriptedLLM{
		streams: [][]types.StreamEvent{
			{
				{Type: "text", Delta: "run"},
				{Type: "tool_calls", ToolCalls: []types.ToolCall{{ID: "1", Name: "sum", Arguments: "1+1"}}},
			},
			{
				{Type: "text", Delta: "done"},
			},
		},
	}

	var gotArgs string
	loop := NewAgentLoop(NewAgentAction(llm))
	ch, err := loop.Execute(context.Background(), LoopInput{
		ActionInput: ActionInput{
			Messages: []protocols.LLMMessage{{Role: "user", Content: "x"}},
			Tools: []types.Tool{
				{
					Name: "sum",
					Execute: func(_ context.Context, args string) (string, error) {
						gotArgs = args
						return "2", nil
					},
				},
			},
		},
		MaxIterations: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	events := collect(ch)
	if gotArgs != "1+1" {
		t.Fatalf("unexpected tool args: %q", gotArgs)
	}

	hasStart := false
	hasEnd := false
	for _, ev := range events {
		if ev.Type == "tool_start" {
			hasStart = true
		}
		if ev.Type == "tool_end" && ev.Delta == "2" {
			hasEnd = true
		}
		if ev.Type == "error" {
			t.Fatalf("unexpected error event: %q", ev.Delta)
		}
	}
	if !hasStart || !hasEnd {
		t.Fatalf("missing tool events start=%v end=%v", hasStart, hasEnd)
	}
}

func TestAgentLoop_MaxIterationsReached(t *testing.T) {
	llm := &scriptedLLM{
		streams: [][]types.StreamEvent{
			{
				{Type: "tool_calls", ToolCalls: []types.ToolCall{{ID: "1", Name: "sum", Arguments: "1+1"}}},
			},
		},
	}

	loop := NewAgentLoop(NewAgentAction(llm))
	ch, err := loop.Execute(context.Background(), LoopInput{
		ActionInput: ActionInput{
			Tools: []types.Tool{
				{
					Name: "sum",
					Execute: func(_ context.Context, _ string) (string, error) {
						return "2", nil
					},
				},
			},
		},
		MaxIterations: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	events := collect(ch)
	found := false
	for _, ev := range events {
		if ev.Type == "error" && ev.Delta == "max iterations reached" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected max iterations error")
	}
}
