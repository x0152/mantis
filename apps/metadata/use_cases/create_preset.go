package usecases

import (
	"context"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreatePreset struct {
	store protocols.Store[string, types.Preset]
}

func NewCreatePreset(store protocols.Store[string, types.Preset]) *CreatePreset {
	return &CreatePreset{store: store}
}

func (uc *CreatePreset) Execute(ctx context.Context, p types.Preset) (types.Preset, error) {
	p.ID = uuid.New().String()
	result, err := uc.store.Create(ctx, []types.Preset{p})
	if err != nil {
		return types.Preset{}, err
	}
	return result[0], nil
}
