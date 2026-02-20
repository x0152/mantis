package usecases

import (
	"context"
	"sort"
	"time"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const pendingTimeout = 10 * time.Minute

type ListMessages struct {
	store  protocols.Store[string, types.ChatMessage]
	buffer *shared.Buffer
}

func NewListMessages(store protocols.Store[string, types.ChatMessage], buffer *shared.Buffer) *ListMessages {
	return &ListMessages{store: store, buffer: buffer}
}

func (uc *ListMessages) Execute(ctx context.Context, limit, offset int, sessionID, source string) ([]types.ChatMessage, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	items, err := uc.list(ctx, limit, offset, sessionID, source)
	if err != nil {
		return nil, err
	}

	var stale []types.ChatMessage
	var result []types.ChatMessage
	for _, m := range items {
		if m.Status == "pending" && time.Since(m.CreatedAt) > pendingTimeout {
			m.Status = "error"
			m.Content = "[Error] generation interrupted"
			stale = append(stale, m)
		}
		result = append(result, m)
	}
	if len(stale) > 0 {
		_, _ = uc.store.Update(ctx, stale)
	}

	for i, m := range result {
		if m.Status == "pending" {
			if entry, ok := uc.buffer.Get(m.ID); ok {
				if entry.Content != "" {
					result[i].Content = entry.Content
				}
				result[i].Steps = shared.StepsToJSON(entry.Steps)
			}
		}
	}

	if result == nil {
		result = []types.ChatMessage{}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (uc *ListMessages) list(ctx context.Context, limit, offset int, sessionID, source string) ([]types.ChatMessage, error) {
	query := types.ListQuery{
		Page: types.Page{Limit: limit, Offset: offset},
		Sort: []types.Sort{{Field: "created_at", Dir: types.SortDirDesc}},
	}
	if source != "" {
		query.Filter = map[string]string{"source": source}
	} else {
		query.FilterNot = map[string]string{"source": "cron"}
	}
	if sessionID != "" {
		if query.Filter == nil {
			query.Filter = map[string]string{}
		}
		query.Filter["session_id"] = sessionID
	}
	items, err := uc.store.List(ctx, query)
	if items == nil {
		items = []types.ChatMessage{}
	}
	return items, err
}
