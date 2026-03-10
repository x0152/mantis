package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListSessions struct {
	store protocols.Store[string, types.ChatSession]
}

func NewListSessions(store protocols.Store[string, types.ChatSession]) *ListSessions {
	return &ListSessions{store: store}
}

func (uc *ListSessions) Execute(ctx context.Context, limit, offset int) ([]types.ChatSession, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	items, err := uc.store.List(ctx, types.ListQuery{
		Page:      types.Page{Limit: limit, Offset: offset},
		Sort:      []types.Sort{{Field: "created_at", Dir: types.SortDirDesc}},
		FilterNot: map[string]string{"id": "cron:%"},
	})
	if err != nil {
		return nil, err
	}

	var result []types.ChatSession
	for _, s := range items {
		if len(s.ID) > 5 && s.ID[:5] == "cron:" {
			continue
		}
		result = append(result, s)
	}
	if result == nil {
		result = []types.ChatSession{}
	}
	return result, nil
}
