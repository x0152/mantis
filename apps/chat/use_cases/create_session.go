package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateSession struct {
	store protocols.Store[string, types.ChatSession]
}

func NewCreateSession(store protocols.Store[string, types.ChatSession]) *CreateSession {
	return &CreateSession{store: store}
}

func (uc *CreateSession) Execute(ctx context.Context, title string) (types.ChatSession, error) {
	session := types.ChatSession{
		ID:        uuid.New().String(),
		Title:     title,
		CreatedAt: time.Now(),
	}
	created, err := uc.store.Create(ctx, []types.ChatSession{session})
	if err != nil {
		return types.ChatSession{}, err
	}
	if len(created) == 0 {
		return types.ChatSession{}, nil
	}
	return created[0], nil
}
