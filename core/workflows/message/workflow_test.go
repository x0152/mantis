package message

import (
	"context"
	"errors"
	"strings"
	"testing"

	artifactplugin "mantis/core/plugins/artifact"
	"mantis/core/types"
)

type workflowMessageStoreMock struct {
	createErr error
	created   []types.ChatMessage
}

func (m *workflowMessageStoreMock) Create(_ context.Context, items []types.ChatMessage) ([]types.ChatMessage, error) {
	m.created = append([]types.ChatMessage{}, items...)
	if m.createErr != nil {
		return nil, m.createErr
	}
	return items, nil
}

func (m *workflowMessageStoreMock) Get(_ context.Context, _ []string) (map[string]types.ChatMessage, error) {
	return map[string]types.ChatMessage{}, nil
}

func (m *workflowMessageStoreMock) List(_ context.Context, _ types.ListQuery) ([]types.ChatMessage, error) {
	return nil, nil
}

func (m *workflowMessageStoreMock) Update(_ context.Context, items []types.ChatMessage) ([]types.ChatMessage, error) {
	return items, nil
}

func (m *workflowMessageStoreMock) Delete(_ context.Context, _ []string) error {
	return nil
}

func TestWorkflow_RequiresSessionID(t *testing.T) {
	w := &Workflow{artifactMgr: artifactplugin.NewManager(nil)}
	_, err := w.Execute(context.Background(), Input{SessionID: " "})
	if err == nil || !strings.Contains(err.Error(), "session_id is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkflow_ReturnsCreateError(t *testing.T) {
	createErr := errors.New("db down")
	store := &workflowMessageStoreMock{createErr: createErr}
	w := &Workflow{
		messageStore: store,
		artifactMgr:  artifactplugin.NewManager(nil),
	}

	_, err := w.Execute(context.Background(), Input{
		SessionID: "s1",
		Content:   "hello",
		Source:    "web",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 2 {
		t.Fatalf("expected two created messages, got %d", len(store.created))
	}
	if store.created[0].Role != "user" || store.created[1].Role != "assistant" {
		t.Fatalf("unexpected roles: %s %s", store.created[0].Role, store.created[1].Role)
	}
}
