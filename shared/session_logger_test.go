package shared

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"mantis/core/types"
)

type mockSessionLogStore struct {
	mu   sync.Mutex
	logs map[string]types.SessionLog
}

func newMockStore() *mockSessionLogStore {
	return &mockSessionLogStore{logs: make(map[string]types.SessionLog)}
}

func (m *mockSessionLogStore) Create(_ context.Context, items []types.SessionLog) ([]types.SessionLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range items {
		m.logs[item.ID] = item
	}
	return items, nil
}

func (m *mockSessionLogStore) Update(_ context.Context, items []types.SessionLog) ([]types.SessionLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range items {
		m.logs[item.ID] = item
	}
	return items, nil
}

func (m *mockSessionLogStore) Get(_ context.Context, ids []string) (map[string]types.SessionLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]types.SessionLog{}
	for _, id := range ids {
		if l, ok := m.logs[id]; ok {
			out[id] = l
		}
	}
	return out, nil
}

func (m *mockSessionLogStore) List(_ context.Context, _ types.ListQuery) ([]types.SessionLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []types.SessionLog
	for _, l := range m.logs {
		out = append(out, l)
	}
	return out, nil
}

func (m *mockSessionLogStore) Delete(_ context.Context, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		delete(m.logs, id)
	}
	return nil
}

func (m *mockSessionLogStore) getFirst() types.SessionLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.logs {
		return l
	}
	return types.SessionLog{}
}

func makeToolStartDelta(name, args string) string {
	step := map[string]any{
		"id":   "step-1",
		"tool": name,
		"args": args,
	}
	b, _ := json.Marshal(step)
	return string(b)
}

func TestSessionLogger_TextBeforeToolCall(t *testing.T) {
	store := newMockStore()
	logger := NewSessionLogger(store)

	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "Plan: I will check the system"},
		{Type: "tool_start", Delta: makeToolStartDelta("execute_command", `{"command":"uname -a"}`)},
		{Type: "tool_end", Delta: "Linux server 5.15.0"},
		{Type: "text", Delta: "Done. The system is Linux."},
	})

	out := logger.Wrap(context.Background(), "conn-1", "ssh", "uname -a", src)
	// Drain output
	for range out {
	}

	// Wait briefly for async save
	time.Sleep(50 * time.Millisecond)

	session := store.getFirst()
	if session.Status != "finished" {
		t.Fatalf("status = %q, want finished", session.Status)
	}
	if session.Prompt != "uname -a" {
		t.Errorf("prompt = %q, want %q", session.Prompt, "uname -a")
	}

	// Expected order: thought, command, output, thought
	if len(session.Entries) != 4 {
		t.Fatalf("expected 4 entries, got %d: %+v", len(session.Entries), session.Entries)
	}

	expected := []struct {
		typ     string
		content string
	}{
		{"thought", "Plan: I will check the system"},
		{"command", makeToolStartDelta("execute_command", `{"command":"uname -a"}`)},
		{"output", "Linux server 5.15.0"},
		{"thought", "Done. The system is Linux."},
	}

	for i, want := range expected {
		got := session.Entries[i]
		if got.Type != want.typ {
			t.Errorf("entry[%d].Type = %q, want %q", i, got.Type, want.typ)
		}
		if got.Content != want.content {
			t.Errorf("entry[%d].Content = %q, want %q", i, got.Content, want.content)
		}
	}
}

func TestSessionLogger_ThinkingEventsIgnored(t *testing.T) {
	store := newMockStore()
	logger := NewSessionLogger(store)

	// "thinking" events (reasoning_content) are NOT handled — they're silently lost
	src := feedEvents([]types.StreamEvent{
		{Type: "thinking", Delta: "deep reasoning here"},
		{Type: "text", Delta: "Plan: check the server"},
		{Type: "tool_start", Delta: makeToolStartDelta("execute_command", `{"command":"ls"}`)},
		{Type: "tool_end", Delta: "file1 file2"},
	})

	out := logger.Wrap(context.Background(), "conn-1", "ssh", "ls", src)
	for range out {
	}
	time.Sleep(50 * time.Millisecond)

	session := store.getFirst()
	entries := session.Entries

	// BUG: "thinking" events are dropped — no entry is created for them.
	// We should have: thought(deep reasoning), thought(Plan), command, output = 4 entries
	// But currently we get: thought(Plan), command, output = 3 entries

	hasThinkingEntry := false
	for _, e := range entries {
		if e.Content == "deep reasoning here" {
			hasThinkingEntry = true
		}
	}
	if hasThinkingEntry {
		t.Log("thinking events ARE captured (good)")
	} else {
		t.Error("BUG: thinking events are silently dropped — no log entry created for reasoning_content")
	}
}

func TestSessionLogger_OnlyThinkingNoTools(t *testing.T) {
	store := newMockStore()
	logger := NewSessionLogger(store)

	// Simulates the full pipeline as it now works in SSH agent:
	// ApplyThinkingStream(skip) -> SessionLogger.Wrap
	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "<think>Пользователь просит выполнить uname -a</think>"},
	})

	filtered := ApplyThinkingStream(src, "skip")
	out := logger.Wrap(context.Background(), "conn-1", "ssh", "uname -a", filtered)
	for range out {
	}
	time.Sleep(50 * time.Millisecond)

	session := store.getFirst()
	// After stripping <think> block, no text remains, so no entries
	if len(session.Entries) != 0 {
		t.Errorf("expected 0 entries (all text was inside <think>), got %d: %+v", len(session.Entries), session.Entries)
	}
	if session.Status != "finished" {
		t.Errorf("status = %q, want finished", session.Status)
	}
}

func TestSessionLogger_ThinkingStrippedTextKept(t *testing.T) {
	store := newMockStore()
	logger := NewSessionLogger(store)

	// Model outputs <think>...</think> then real text, then tool call
	src := feedEvents([]types.StreamEvent{
		{Type: "text", Delta: "<think>reasoning</think>"},
		{Type: "text", Delta: "Plan: check system"},
		{Type: "tool_calls", ToolCalls: []types.ToolCall{{ID: "1", Name: "execute_command"}}},
	})

	// Full pipeline: ApplyThinkingStream strips <think>, then session_logger wraps
	filtered := ApplyThinkingStream(src, "skip")
	out := logger.Wrap(context.Background(), "conn-1", "ssh", "uname -a", filtered)
	for range out {
	}
	time.Sleep(50 * time.Millisecond)

	session := store.getFirst()
	// "Plan: check system" survives as thought, no <think> tag content
	hasCleanThought := false
	for _, e := range session.Entries {
		if e.Type == "thought" && e.Content == "Plan: check system" {
			hasCleanThought = true
		}
		if e.Content == "<think>reasoning</think>" {
			t.Error("BUG: <think> content leaked into session log")
		}
	}
	if !hasCleanThought {
		t.Error("expected clean 'Plan: check system' thought entry")
	}
}

func TestSessionLogger_IncrementalSave(t *testing.T) {
	store := newMockStore()
	logger := NewSessionLogger(store)

	events := make(chan types.StreamEvent, 10)
	out := logger.Wrap(context.Background(), "conn-1", "ssh", "test", events)

	// Drain output in background
	go func() {
		for range out {
		}
	}()

	events <- types.StreamEvent{Type: "text", Delta: "thinking..."}
	events <- types.StreamEvent{Type: "tool_start", Delta: makeToolStartDelta("execute_command", `{"command":"ls"}`)}
	time.Sleep(50 * time.Millisecond)

	// After tool_start, entries should be saved incrementally
	session := store.getFirst()
	if len(session.Entries) < 2 {
		t.Errorf("expected at least 2 entries after tool_start, got %d", len(session.Entries))
	}

	events <- types.StreamEvent{Type: "tool_end", Delta: "result"}
	close(events)
	time.Sleep(50 * time.Millisecond)

	session = store.getFirst()
	if session.Status != "finished" {
		t.Errorf("status = %q, want finished", session.Status)
	}
	if len(session.Entries) != 3 {
		t.Errorf("expected 3 entries at end, got %d", len(session.Entries))
	}
}
