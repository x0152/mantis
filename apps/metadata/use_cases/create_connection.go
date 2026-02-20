package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateConnection struct {
	store protocols.Store[string, types.Connection]
}

func NewCreateConnection(store protocols.Store[string, types.Connection]) *CreateConnection {
	return &CreateConnection{store: store}
}

func (uc *CreateConnection) Execute(ctx context.Context, connType, name, description, modelID string, config json.RawMessage) (types.Connection, error) {
	c := types.Connection{
		ID:          uuid.New().String(),
		Type:        connType,
		Name:        name,
		Description: description,
		ModelID:     modelID,
		Config:      config,
		Memories:    []types.Memory{},
	}
	result, err := uc.store.Create(ctx, []types.Connection{c})
	if err != nil {
		return types.Connection{}, err
	}
	return result[0], nil
}
