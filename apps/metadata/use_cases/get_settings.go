package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type GetSettings struct {
	store protocols.Store[string, types.Settings]
}

func NewGetSettings(store protocols.Store[string, types.Settings]) *GetSettings {
	return &GetSettings{store: store}
}

func (uc *GetSettings) Execute(ctx context.Context) (types.Settings, error) {
	result, err := uc.store.Get(ctx, []string{settingsID})
	if err != nil {
		return types.Settings{}, err
	}
	item, ok := result[settingsID]
	if ok {
		if item.UserMemories == nil {
			item.UserMemories = []string{}
		}
		return item, nil
	}
	return types.Settings{
		ID:            settingsID,
		MemoryEnabled: true,
		UserMemories:  []string{},
	}, nil
}
