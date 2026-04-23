package usecases

import (
	"context"
	"strings"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetConnectionInferenceLimit struct {
	connStore protocols.Store[string, types.LlmConnection]
	catalogs  map[string]protocols.LLMCatalog
}

func NewGetConnectionInferenceLimit(connStore protocols.Store[string, types.LlmConnection], catalogs map[string]protocols.LLMCatalog) *GetConnectionInferenceLimit {
	return &GetConnectionInferenceLimit{connStore: connStore, catalogs: catalogs}
}

func (uc *GetConnectionInferenceLimit) Execute(ctx context.Context, connectionID string) (types.InferenceLimit, error) {
	items, err := uc.connStore.Get(ctx, []string{connectionID})
	if err != nil {
		return types.InferenceLimit{}, err
	}
	conn, ok := items[connectionID]
	if !ok {
		return types.InferenceLimit{}, base.ErrNotFound
	}
	catalog, ok := uc.catalogs[strings.ToLower(strings.TrimSpace(conn.Provider))]
	if !ok || catalog == nil {
		return types.InferenceLimit{
			Type:  "unlimited",
			Label: "No inference limit reported",
		}, nil
	}
	limit, err := catalog.GetInferenceLimit(ctx, conn.BaseURL, conn.APIKey)
	if err != nil {
		return types.InferenceLimit{}, err
	}
	return limit, nil
}
