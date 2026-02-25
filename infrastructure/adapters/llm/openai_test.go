package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mantis/core/types"
)

func collectStreamEvents(ch <-chan types.StreamEvent) []types.StreamEvent {
	var out []types.StreamEvent
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestChatStream_ToolCallsWithCorrectFinishReason(t *testing.T) {
	// Standard OpenAI behavior: finish_reason = "tool_calls"
	sseData := `data: {"choices":[{"delta":{"content":"<think>reasoning</think>"},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"execute_command","arguments":""}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"command\":\"uname -a\"}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	o := NewOpenAI()
	ch, err := o.ChatStream(context.Background(), server.URL, "test-key", nil, "test-model", nil, "")
	if err != nil {
		t.Fatal(err)
	}

	events := collectStreamEvents(ch)

	hasToolCalls := false
	for _, e := range events {
		if e.Type == "tool_calls" {
			hasToolCalls = true
			if len(e.ToolCalls) != 1 || e.ToolCalls[0].Name != "execute_command" {
				t.Errorf("unexpected tool_calls: %+v", e.ToolCalls)
			}
		}
	}
	if !hasToolCalls {
		t.Error("expected tool_calls event, got none")
	}
}

func TestChatStream_ToolCallsWithStopFinishReason(t *testing.T) {
	// LM Studio bug: model sends tool_calls in delta but finish_reason = "stop"
	sseData := `data: {"choices":[{"delta":{"content":"<think>reasoning</think>"},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"execute_command","arguments":""}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"command\":\"uname -a\"}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"stop"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	o := NewOpenAI()
	ch, err := o.ChatStream(context.Background(), server.URL, "test-key", nil, "test-model", nil, "")
	if err != nil {
		t.Fatal(err)
	}

	events := collectStreamEvents(ch)

	hasToolCalls := false
	for _, e := range events {
		if e.Type == "tool_calls" {
			hasToolCalls = true
		}
	}
	if hasToolCalls {
		t.Log("tool_calls emitted even with finish_reason=stop (good)")
	} else {
		t.Error("BUG: tool_calls LOST because finish_reason='stop' instead of 'tool_calls' — LM Studio compatibility issue")
	}
}

func TestChatStream_ToolCallsNoFinishReason(t *testing.T) {
	// Some providers don't send finish_reason at all, stream just ends
	sseData := `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"execute_command","arguments":"{\"command\":\"ls\"}"}}]},"finish_reason":null}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	o := NewOpenAI()
	ch, err := o.ChatStream(context.Background(), server.URL, "test-key", nil, "test-model", nil, "")
	if err != nil {
		t.Fatal(err)
	}

	events := collectStreamEvents(ch)

	hasToolCalls := false
	for _, e := range events {
		if e.Type == "tool_calls" {
			hasToolCalls = true
		}
	}
	if hasToolCalls {
		t.Log("tool_calls emitted on stream end (good)")
	} else {
		t.Error("BUG: tool_calls LOST — accumulated tool calls not emitted when stream ends without finish_reason=tool_calls")
	}
}

func TestChatStream_ThinkingModeSkipStripsThinkTags(t *testing.T) {
	sseData := `data: {"choices":[{"delta":{"content":"<think>reasoning</think>Plan: do stuff"},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"execute_command","arguments":"{\"command\":\"ls\"}"}}]},"finish_reason":"tool_calls"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	o := NewOpenAI()
	ch, err := o.ChatStream(context.Background(), server.URL, "test-key", nil, "test-model", nil, "skip")
	if err != nil {
		t.Fatal(err)
	}

	events := collectStreamEvents(ch)

	for _, e := range events {
		if e.Type == "text" && e.Delta == "Plan: do stuff" {
			return // success
		}
		if e.Type == "text" && strings.Contains(e.Delta, "<think>") {
			t.Errorf("think tags leaked through: %q", e.Delta)
		}
	}
	// "Plan: do stuff" text event expected
	hasText := false
	for _, e := range events {
		if e.Type == "text" {
			hasText = true
		}
	}
	if !hasText {
		t.Error("no text event found after thinking skip")
	}
}

func TestChatStream_ToolCallsWithSparseIndexes(t *testing.T) {
	// Some providers may emit non-contiguous tool_call indexes.
	sseData := `data: {"choices":[{"delta":{"tool_calls":[{"index":3,"id":"call_4","type":"function","function":{"name":"execute_command","arguments":""}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":3,"function":{"arguments":"{\"command\":\"uptime\"}"}}]},"finish_reason":"tool_calls"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	o := NewOpenAI()
	ch, err := o.ChatStream(context.Background(), server.URL, "test-key", nil, "test-model", nil, "")
	if err != nil {
		t.Fatal(err)
	}

	events := collectStreamEvents(ch)

	for _, e := range events {
		if e.Type != "tool_calls" {
			continue
		}
		if len(e.ToolCalls) != 1 {
			t.Fatalf("expected 1 tool call, got %d: %+v", len(e.ToolCalls), e.ToolCalls)
		}
		got := e.ToolCalls[0]
		if got.ID != "call_4" || got.Name != "execute_command" || got.Arguments != "{\"command\":\"uptime\"}" {
			t.Fatalf("unexpected tool call content: %+v", got)
		}
		return
	}
	t.Fatal("expected tool_calls event, got none")
}
