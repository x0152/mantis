package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListSessionLogs struct {
	store protocols.Store[string, types.SessionLog]
}

func NewListSessionLogs(store protocols.Store[string, types.SessionLog]) *ListSessionLogs {
	return &ListSessionLogs{store: store}
}

func (uc *ListSessionLogs) Execute(ctx context.Context, connectionID string, limit, offset int) ([]types.SessionLog, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	query := types.ListQuery{
		Page: types.Page{Limit: limit, Offset: offset},
		Sort: []types.Sort{{Field: "started_at", Dir: types.SortDirDesc}},
	}
	if connectionID != "" {
		query.Filter = map[string]string{"connection_id": connectionID}
	}
	items, err := uc.store.List(ctx, query)
	if items == nil {
		items = []types.SessionLog{}
	}
	return items, err
}
