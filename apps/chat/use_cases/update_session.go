package usecases

import (
	"context"
	"fmt"

	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateSession struct {
	store protocols.Store[string, types.ChatSession]
}

func NewUpdateSession(store protocols.Store[string, types.ChatSession]) *UpdateSession {
	return &UpdateSession{store: store}
}

func (uc *UpdateSession) Execute(ctx context.Context, id, title string) (types.ChatSession, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.ChatSession{}, err
	}
	session, ok := existing[id]
	if !ok {
		return types.ChatSession{}, fmt.Errorf("session %q not found", id)
	}
	session.Title = title
	updated, err := uc.store.Update(ctx, []types.ChatSession{session})
	if err != nil {
		return types.ChatSession{}, err
	}
	return updated[0], nil
}
