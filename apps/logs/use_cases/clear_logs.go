package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ClearLogs struct {
	store protocols.Store[string, types.SessionLog]
}

func NewClearLogs(store protocols.Store[string, types.SessionLog]) *ClearLogs {
	return &ClearLogs{store: store}
}

func (uc *ClearLogs) Execute(ctx context.Context) error {
	logs, err := uc.store.List(ctx, types.ListQuery{})
	if err != nil {
		return err
	}
	if len(logs) == 0 {
		return nil
	}
	ids := make([]string, len(logs))
	for i, l := range logs {
		ids[i] = l.ID
	}
	return uc.store.Delete(ctx, ids)
}
