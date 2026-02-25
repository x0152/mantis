package usecases

import (
	"context"
	"testing"
	"time"

	"mantis/core/types"
	"mantis/shared"
)

type chatMessageStoreMock struct {
	items     []types.ChatMessage
	lastQuery types.ListQuery
	updated   []types.ChatMessage
}

func (m *chatMessageStoreMock) Create(_ context.Context, _ []types.ChatMessage) ([]types.ChatMessage, error) {
	return nil, nil
}

func (m *chatMessageStoreMock) Get(_ context.Context, _ []string) (map[string]types.ChatMessage, error) {
	return map[string]types.ChatMessage{}, nil
}

func (m *chatMessageStoreMock) List(_ context.Context, query types.ListQuery) ([]types.ChatMessage, error) {
	m.lastQuery = query
	return m.items, nil
}

func (m *chatMessageStoreMock) Update(_ context.Context, items []types.ChatMessage) ([]types.ChatMessage, error) {
	m.updated = append([]types.ChatMessage{}, items...)
	return items, nil
}

func (m *chatMessageStoreMock) Delete(_ context.Context, _ []string) error {
	return nil
}

func TestListMessages_StalePendingAndBufferOverlay(t *testing.T) {
	now := time.Now()
	store := &chatMessageStoreMock{
		items: []types.ChatMessage{
			{ID: "m3", Status: "pending", Content: "old", CreatedAt: now.Add(-pendingTimeout - time.Second)},
			{ID: "m2", Status: "pending", Content: "pending", CreatedAt: now.Add(-time.Minute)},
			{ID: "m1", Status: "", Content: "done", CreatedAt: now.Add(-2 * time.Minute)},
		},
	}
	buffer := shared.NewBuffer()
	buffer.SetContent("m2", "stream")
	buffer.SetStep("m2", types.Step{ID: "s1", Tool: "t"})

	uc := NewListMessages(store, buffer)
	out, err := uc.Execute(context.Background(), 0, -1, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if store.lastQuery.Page.Limit != 10 || store.lastQuery.Page.Offset != 0 {
		t.Fatalf("unexpected pagination: %+v", store.lastQuery.Page)
	}
	if store.lastQuery.FilterNot["source"] != "cron" {
		t.Fatalf("unexpected filter_not: %+v", store.lastQuery.FilterNot)
	}
	if len(store.lastQuery.Sort) != 1 || store.lastQuery.Sort[0].Field != "created_at" || store.lastQuery.Sort[0].Dir != types.SortDirDesc {
		t.Fatalf("unexpected sort: %+v", store.lastQuery.Sort)
	}

	if len(store.updated) != 1 || store.updated[0].ID != "m3" || store.updated[0].Status != "error" {
		t.Fatalf("unexpected stale updates: %+v", store.updated)
	}
	if len(out) != 3 {
		t.Fatalf("unexpected result length: %d", len(out))
	}
	if out[0].ID != "m3" || out[1].ID != "m1" || out[2].ID != "m2" {
		t.Fatalf("unexpected order: %s %s %s", out[0].ID, out[1].ID, out[2].ID)
	}
	if out[2].Content != "stream" {
		t.Fatalf("unexpected buffered content: %q", out[2].Content)
	}
	if len(out[2].Steps) == 0 {
		t.Fatal("expected buffered steps")
	}
	if out[0].Content != "[Error] generation interrupted" {
		t.Fatalf("unexpected stale content: %q", out[0].Content)
	}
}

func TestListMessages_SourceAndSessionFilters(t *testing.T) {
	store := &chatMessageStoreMock{items: []types.ChatMessage{}}
	uc := NewListMessages(store, shared.NewBuffer())

	_, err := uc.Execute(context.Background(), 5, 2, "s1", "telegram")
	if err != nil {
		t.Fatal(err)
	}

	if store.lastQuery.Filter["source"] != "telegram" || store.lastQuery.Filter["session_id"] != "s1" {
		t.Fatalf("unexpected filter: %+v", store.lastQuery.Filter)
	}
	if len(store.lastQuery.FilterNot) != 0 {
		t.Fatalf("unexpected filter_not: %+v", store.lastQuery.FilterNot)
	}
}
