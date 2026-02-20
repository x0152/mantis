package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type AddMemory struct {
	store protocols.Store[string, types.Connection]
}

func NewAddMemory(store protocols.Store[string, types.Connection]) *AddMemory {
	return &AddMemory{store: store}
}

func (uc *AddMemory) Execute(ctx context.Context, connectionID, content string) (types.Connection, error) {
	existing, err := uc.store.Get(ctx, []string{connectionID})
	if err != nil {
		return types.Connection{}, err
	}
	c, ok := existing[connectionID]
	if !ok {
		return types.Connection{}, base.ErrNotFound
	}
	c.Memories = append(c.Memories, types.Memory{
		ID:        uuid.New().String(),
		Content:   content,
		CreatedAt: time.Now(),
	})
	result, err := uc.store.Update(ctx, []types.Connection{c})
	if err != nil {
		return types.Connection{}, err
	}
	return result[0], nil
}
