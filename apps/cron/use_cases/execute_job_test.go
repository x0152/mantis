package usecases

import (
	"context"
	"testing"

	"mantis/core/types"
)

type channelStoreMock struct {
	items []types.Channel
}

func (m *channelStoreMock) Create(_ context.Context, items []types.Channel) ([]types.Channel, error) {
	return items, nil
}

func (m *channelStoreMock) Get(_ context.Context, _ []string) (map[string]types.Channel, error) {
	return map[string]types.Channel{}, nil
}

func (m *channelStoreMock) List(_ context.Context, _ types.ListQuery) ([]types.Channel, error) {
	return m.items, nil
}

func (m *channelStoreMock) Update(_ context.Context, items []types.Channel) ([]types.Channel, error) {
	return items, nil
}

func (m *channelStoreMock) Delete(_ context.Context, _ []string) error {
	return nil
}

func TestResolveSender_UsesAllowedUserFromTelegramChannel(t *testing.T) {
	uc := &ExecuteJob{
		channelStore: &channelStoreMock{
			items: []types.Channel{
				{Type: "telegram", Token: "token-1", AllowedUserIDs: []int64{123456789}},
			},
		},
	}

	sender, err := uc.resolveSender(context.Background(), "telegram")
	if err != nil {
		t.Fatalf("resolveSender returned error: %v", err)
	}
	if sender == nil {
		t.Fatal("expected sender, got nil")
	}
	if got := sender.Recipient(); got != "123456789" {
		t.Fatalf("unexpected recipient: %s", got)
	}
	if got := sender.Channel(); got != "telegram" {
		t.Fatalf("unexpected channel: %s", got)
	}
}

func TestResolveSender_ReturnsNilWhenChannelNotConfigured(t *testing.T) {
	uc := &ExecuteJob{
		channelStore: &channelStoreMock{
			items: []types.Channel{
				{Type: "telegram", Token: "token-1", AllowedUserIDs: []int64{42}},
			},
		},
	}

	sender, err := uc.resolveSender(context.Background(), "")
	if err != nil {
		t.Fatalf("resolveSender returned error: %v", err)
	}
	if sender != nil {
		t.Fatal("expected nil sender when channel is not configured")
	}
}

func TestResolveSender_ErrorsWhenTelegramChannelHasNoAllowedUsers(t *testing.T) {
	uc := &ExecuteJob{
		channelStore: &channelStoreMock{
			items: []types.Channel{
				{Type: "telegram", Token: "token-1", AllowedUserIDs: []int64{}},
			},
		},
	}

	_, err := uc.resolveSender(context.Background(), "telegram")
	if err == nil {
		t.Fatal("expected error for empty telegram recipient")
	}
}
