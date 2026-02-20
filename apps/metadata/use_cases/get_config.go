package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetConfig struct {
	store protocols.Store[string, types.Config]
}

func NewGetConfig(store protocols.Store[string, types.Config]) *GetConfig {
	return &GetConfig{store: store}
}

func (uc *GetConfig) Execute(ctx context.Context) (types.Config, error) {
	result, err := uc.store.Get(ctx, []string{configID})
	if err != nil {
		return types.Config{}, err
	}
	cfg, ok := result[configID]
	if !ok {
		return types.Config{}, base.ErrNotFound
	}
	return cfg, nil
}
