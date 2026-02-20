package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ClearHistory struct {
	sessionStore protocols.Store[string, types.ChatSession]
	messageStore protocols.Store[string, types.ChatMessage]
}

func NewClearHistory(
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
) *ClearHistory {
	return &ClearHistory{sessionStore: sessionStore, messageStore: messageStore}
}

func (uc *ClearHistory) Execute(ctx context.Context) error {
	messages, err := uc.messageStore.List(ctx, types.ListQuery{})
	if err != nil {
		return err
	}
	if len(messages) > 0 {
		ids := make([]string, len(messages))
		for i, m := range messages {
			ids[i] = m.ID
		}
		if err := uc.messageStore.Delete(ctx, ids); err != nil {
			return err
		}
	}

	sessions, err := uc.sessionStore.List(ctx, types.ListQuery{})
	if err != nil {
		return err
	}
	if len(sessions) > 0 {
		ids := make([]string, len(sessions))
		for i, s := range sessions {
			ids[i] = s.ID
		}
		if err := uc.sessionStore.Delete(ctx, ids); err != nil {
			return err
		}
	}

	return nil
}
