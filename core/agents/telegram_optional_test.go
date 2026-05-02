package agents

import (
	"context"
	"strings"
	"testing"

	"mantis/core/types"
)

type channelStoreMock struct {
	channels map[string]types.Channel
}

func (s *channelStoreMock) Create(_ context.Context, items []types.Channel) ([]types.Channel, error) {
	for _, c := range items {
		s.channels[c.ID] = c
	}
	return items, nil
}

func (s *channelStoreMock) Get(_ context.Context, ids []string) (map[string]types.Channel, error) {
	out := make(map[string]types.Channel)
	for _, id := range ids {
		if c, ok := s.channels[id]; ok {
			out[id] = c
		}
	}
	return out, nil
}

func (s *channelStoreMock) List(_ context.Context, _ types.ListQuery) ([]types.Channel, error) {
	out := make([]types.Channel, 0, len(s.channels))
	for _, c := range s.channels {
		out = append(out, c)
	}
	return out, nil
}

func (s *channelStoreMock) Update(_ context.Context, items []types.Channel) ([]types.Channel, error) {
	for _, c := range items {
		s.channels[c.ID] = c
	}
	return items, nil
}

func (s *channelStoreMock) Delete(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(s.channels, id)
	}
	return nil
}

func TestResolveTelegramRecipient_NotConfigured(t *testing.T) {
	_, _, err := resolveTelegramRecipient(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when channel store is nil")
	}
	if !strings.Contains(err.Error(), "Telegram is not connected") {
		t.Fatalf("expected friendly hint, got %q", err)
	}
	if !strings.Contains(err.Error(), "TG_BOT_TOKEN") {
		t.Fatalf("expected env hint, got %q", err)
	}
}

func TestResolveTelegramRecipient_EmptyStore(t *testing.T) {
	store := &channelStoreMock{channels: map[string]types.Channel{}}
	_, _, err := resolveTelegramRecipient(context.Background(), store)
	if err == nil {
		t.Fatal("expected error for empty store")
	}
	if !strings.Contains(err.Error(), "Telegram is not connected") {
		t.Fatalf("expected friendly hint, got %q", err)
	}
}

func TestResolveTelegramRecipient_TokenButNoRecipient(t *testing.T) {
	store := &channelStoreMock{channels: map[string]types.Channel{
		"tg": {ID: "tg", Type: "telegram", Token: "TOKEN"},
	}}
	_, _, err := resolveTelegramRecipient(context.Background(), store)
	if err == nil {
		t.Fatal("expected error when recipient is missing")
	}
	if !strings.Contains(err.Error(), "no recipient") {
		t.Fatalf("expected recipient hint, got %q", err)
	}
}

func TestResolveTelegramRecipient_FullyConfigured(t *testing.T) {
	store := &channelStoreMock{channels: map[string]types.Channel{
		"tg": {ID: "tg", Type: "telegram", Token: "TOKEN", AllowedUserIDs: []int64{42}},
	}}
	token, chatID, err := resolveTelegramRecipient(context.Background(), store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "TOKEN" || chatID != 42 {
		t.Fatalf("expected TOKEN/42, got %q/%d", token, chatID)
	}
}

func TestSendNotificationTool_TelegramOptionalErrorMessage(t *testing.T) {
	agent := &MantisAgent{}
	tool := agent.sendNotificationTool()
	_, err := tool.Execute(context.Background(), `{"text":"hello"}`)
	if err == nil {
		t.Fatal("expected error when Telegram is not connected")
	}
	if !strings.Contains(err.Error(), "Telegram is not connected") {
		t.Fatalf("expected friendly hint in tool error, got %q", err)
	}
}
