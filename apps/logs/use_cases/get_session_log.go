package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetSessionLog struct {
	store protocols.Store[string, types.SessionLog]
}

func NewGetSessionLog(store protocols.Store[string, types.SessionLog]) *GetSessionLog {
	return &GetSessionLog{store: store}
}

func (uc *GetSessionLog) Execute(ctx context.Context, id string) (types.SessionLog, error) {
	result, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.SessionLog{}, err
	}
	s, ok := result[id]
	if !ok {
		return types.SessionLog{}, base.ErrNotFound
	}
	return s, nil
}
