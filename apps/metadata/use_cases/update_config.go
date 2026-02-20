package usecases

import (
	"context"
	"encoding/json"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateConfig struct {
	store protocols.Store[string, types.Config]
}

func NewUpdateConfig(store protocols.Store[string, types.Config]) *UpdateConfig {
	return &UpdateConfig{store: store}
}

func (uc *UpdateConfig) Execute(ctx context.Context, data json.RawMessage) (types.Config, error) {
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return types.Config{}, base.ErrValidation
	}

	cfg := types.Config{ID: configID, Data: data}

	existing, err := uc.store.Get(ctx, []string{configID})
	if err != nil {
		return types.Config{}, err
	}
	if _, ok := existing[configID]; ok {
		result, err := uc.store.Update(ctx, []types.Config{cfg})
		if err != nil {
			return types.Config{}, err
		}
		return result[0], nil
	}

	result, err := uc.store.Create(ctx, []types.Config{cfg})
	if err != nil {
		return types.Config{}, err
	}
	return result[0], nil
}
