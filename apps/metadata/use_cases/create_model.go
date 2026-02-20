package usecases

import (
	"context"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateModel struct {
	store protocols.Store[string, types.Model]
}

func NewCreateModel(store protocols.Store[string, types.Model]) *CreateModel {
	return &CreateModel{store: store}
}

func (uc *CreateModel) Execute(ctx context.Context, connectionID, name, thinkingMode string) (types.Model, error) {
	m := types.Model{
		ID:           uuid.New().String(),
		ConnectionID: connectionID,
		Name:         name,
		ThinkingMode: thinkingMode,
	}
	result, err := uc.store.Create(ctx, []types.Model{m})
	if err != nil {
		return types.Model{}, err
	}
	return result[0], nil
}
