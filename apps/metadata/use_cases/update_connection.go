package usecases

import (
	"context"
	"encoding/json"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateConnection struct {
	store protocols.Store[string, types.Connection]
}

func NewUpdateConnection(store protocols.Store[string, types.Connection]) *UpdateConnection {
	return &UpdateConnection{store: store}
}

func (uc *UpdateConnection) Execute(ctx context.Context, id, connType, name, description, modelID string, config json.RawMessage, profileIDs []string, memoryEnabled bool) (types.Connection, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.Connection{}, err
	}
	old, ok := existing[id]
	if !ok {
		return types.Connection{}, base.ErrNotFound
	}
	if profileIDs == nil {
		profileIDs = []string{}
	}
	c := types.Connection{
		ID: id, Type: connType, Name: name, Description: description,
		ModelID: modelID, Config: config, Memories: old.Memories, ProfileIDs: profileIDs,
		MemoryEnabled: memoryEnabled,
	}
	result, err := uc.store.Update(ctx, []types.Connection{c})
	if err != nil {
		return types.Connection{}, err
	}
	return result[0], nil
}
