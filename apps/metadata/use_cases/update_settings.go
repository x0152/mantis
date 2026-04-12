package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateSettings struct {
	store protocols.Store[string, types.Settings]
}

func NewUpdateSettings(store protocols.Store[string, types.Settings]) *UpdateSettings {
	return &UpdateSettings{store: store}
}

func (uc *UpdateSettings) Execute(ctx context.Context, item types.Settings) (types.Settings, error) {
	item.ID = settingsID
	if item.UserMemories == nil {
		item.UserMemories = []string{}
	}

	existing, err := uc.store.Get(ctx, []string{settingsID})
	if err != nil {
		return types.Settings{}, err
	}
	if _, ok := existing[settingsID]; ok {
		result, err := uc.store.Update(ctx, []types.Settings{item})
		if err != nil {
			return types.Settings{}, err
		}
		return result[0], nil
	}

	result, err := uc.store.Create(ctx, []types.Settings{item})
	if err != nil {
		return types.Settings{}, err
	}
	return result[0], nil
}
