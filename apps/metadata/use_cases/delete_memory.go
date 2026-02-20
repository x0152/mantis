package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteMemory struct {
	store protocols.Store[string, types.Connection]
}

func NewDeleteMemory(store protocols.Store[string, types.Connection]) *DeleteMemory {
	return &DeleteMemory{store: store}
}

func (uc *DeleteMemory) Execute(ctx context.Context, connectionID, memoryID string) (types.Connection, error) {
	existing, err := uc.store.Get(ctx, []string{connectionID})
	if err != nil {
		return types.Connection{}, err
	}
	c, ok := existing[connectionID]
	if !ok {
		return types.Connection{}, base.ErrNotFound
	}
	filtered := make([]types.Memory, 0, len(c.Memories))
	for _, m := range c.Memories {
		if m.ID != memoryID {
			filtered = append(filtered, m)
		}
	}
	c.Memories = filtered
	result, err := uc.store.Update(ctx, []types.Connection{c})
	if err != nil {
		return types.Connection{}, err
	}
	return result[0], nil
}
