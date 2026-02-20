package shared

import (
	"testing"

	"mantis/core/types"
)

func collectEvents(ch <-chan types.StreamEvent) []types.StreamEvent {
	var out []types.StreamEvent
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func feedEvents(events []types.StreamEvent) <-chan types.StreamEvent {
	ch := make(chan types.StreamEvent, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return ch
}

func TestApplyThinkingMode_Skip(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"strips complete block", "<think>reasoning here</think>Plan: do stuff", "Plan: do stuff"},
		{"strips block with newlines", "<think>\nstep1\nstep2\n</think>\nResult", "Result"},
		{"strips only think tags", "Hello <think>inner</think> world", "Hello  world"},
		{"strips unclosed tag", "<think>partial reasoning that never closes", ""},
		{"no tags passes through", "plain text", "plain text"},
		{"empty after strip", "<think>only thinking</think>", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyThinkingMode(tt.in, "skip")
			if got != tt.want {
				t.Errorf("ApplyThinkingMode(%q, skip) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestApplyThinkingStream_SkipMode_StripsThinkTags(t *testing.T) {
	// Simulate what LM Studio does: <think>...</think> as text, then tool_calls
	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "<think>"},
		{Type: "text", Delta: "I need to run uname"},
		{Type: "text", Delta: "</think>"},
		{Type: "text", Delta: "\nPlan: check system info"},
		{Type: "tool_calls", ToolCalls: []types.ToolCall{{ID: "1", Name: "execute_command", Arguments: `{"command":"uname -a"}`}}},
	})

	events := collectEvents(ApplyThinkingStream(src, "skip"))

	// Should get: one text event "Plan: check system info", then tool_calls
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d: %+v", len(events), events)
	}
	if events[0].Type != "text" || events[0].Delta != "Plan: check system info" {
		t.Errorf("event[0] = {%s, %q}, want {text, %q}", events[0].Type, events[0].Delta, "Plan: check system info")
	}
	if events[1].Type != "tool_calls" {
		t.Errorf("event[1].Type = %s, want tool_calls", events[1].Type)
	}
}

func TestApplyThinkingStream_SkipMode_AllThinking(t *testing.T) {
	// Model outputs ONLY thinking, then tool_calls — all text should be stripped
	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "<think>I will run the command</think>"},
		{Type: "tool_calls", ToolCalls: []types.ToolCall{{ID: "1", Name: "execute_command"}}},
	})

	events := collectEvents(ApplyThinkingStream(src, "skip"))

	// Only tool_calls should survive (thinking stripped, no remaining text)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Type != "tool_calls" {
		t.Errorf("event[0].Type = %s, want tool_calls", events[0].Type)
	}
}

func TestApplyThinkingStream_SkipMode_ThinkingEvents(t *testing.T) {
	// API-level thinking events (reasoning_content) should be dropped in skip mode
	src := feedEvents([]types.StreamEvent{
		{Type: "thinking", Delta: "deep reasoning"},
		{Type: "text", Delta: "visible response"},
	})

	events := collectEvents(ApplyThinkingStream(src, "skip"))

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Type != "text" || events[0].Delta != "visible response" {
		t.Errorf("event[0] = {%s, %q}, want {text, %q}", events[0].Type, events[0].Delta, "visible response")
	}
}

func TestApplyThinkingStream_SkipMode_TokenByToken(t *testing.T) {
	// Realistic scenario: model streams thinking token-by-token then calls tool
	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "<"},
		{Type: "text", Delta: "think"},
		{Type: "text", Delta: ">"},
		{Type: "text", Delta: "Пользователь просит выполнить uname"},
		{Type: "text", Delta: "\n"},
		{Type: "text", Delta: "</"},
		{Type: "text", Delta: "think"},
		{Type: "text", Delta: ">"},
		{Type: "tool_calls", ToolCalls: []types.ToolCall{{ID: "1", Name: "execute_command"}}},
	})

	events := collectEvents(ApplyThinkingStream(src, "skip"))

	// All text was inside <think>, should be stripped. Only tool_calls remains.
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Type != "tool_calls" {
		t.Errorf("event[0].Type = %s, want tool_calls", events[0].Type)
	}
}

func TestApplyThinkingStream_SkipMode_NoToolCalls(t *testing.T) {
	// Model outputs thinking + regular text, no tool calls (stream ends)
	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "<think>reasoning</think>"},
		{Type: "text", Delta: "Final answer here"},
	})

	events := collectEvents(ApplyThinkingStream(src, "skip"))

	// The trailing text flush should contain only "Final answer here"
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Delta != "Final answer here" {
		t.Errorf("event[0].Delta = %q, want %q", events[0].Delta, "Final answer here")
	}
}
