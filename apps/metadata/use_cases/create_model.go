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

const defaultCompactTokens = 100000

func (uc *CreateModel) Execute(ctx context.Context, connectionID, name, thinkingMode string, compactTokens int) (types.Model, error) {
	if compactTokens <= 0 {
		compactTokens = defaultCompactTokens
	}
	m := types.Model{
		ID:            uuid.New().String(),
		ConnectionID:  connectionID,
		Name:          name,
		ThinkingMode:  thinkingMode,
		CompactTokens: compactTokens,
	}
	result, err := uc.store.Create(ctx, []types.Model{m})
	if err != nil {
		return types.Model{}, err
	}
	return result[0], nil
}
