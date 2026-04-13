package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

type ListSessions struct {
	store  protocols.Store[string, types.ChatSession]
	buffer *shared.Buffer
}

func NewListSessions(store protocols.Store[string, types.ChatSession], buffer *shared.Buffer) *ListSessions {
	return &ListSessions{store: store, buffer: buffer}
}

func (uc *ListSessions) Execute(ctx context.Context, limit, offset int) ([]types.ChatSession, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	items, err := uc.store.List(ctx, types.ListQuery{
		Page: types.Page{Limit: limit, Offset: offset},
		Sort: []types.Sort{{Field: "created_at", Dir: types.SortDirDesc}},
	})
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []types.ChatSession{}
	}

	if uc.buffer != nil {
		active := make(map[string]struct{})
		for _, id := range uc.buffer.ActiveSessionIDs() {
			active[id] = struct{}{}
		}
		for i := range items {
			if _, ok := active[items[i].ID]; ok {
				items[i].Active = true
			}
		}
	}

	return items, nil
}
