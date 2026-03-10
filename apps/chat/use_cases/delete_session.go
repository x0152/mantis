package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteSession struct {
	sessionStore protocols.Store[string, types.ChatSession]
	messageStore protocols.Store[string, types.ChatMessage]
}

func NewDeleteSession(
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
) *DeleteSession {
	return &DeleteSession{sessionStore: sessionStore, messageStore: messageStore}
}

func (uc *DeleteSession) Execute(ctx context.Context, id string) error {
	messages, err := uc.messageStore.List(ctx, types.ListQuery{
		Filter: map[string]string{"session_id": id},
	})
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
	return uc.sessionStore.Delete(ctx, []string{id})
}
