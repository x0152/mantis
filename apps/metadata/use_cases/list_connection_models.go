package usecases

import (
	"context"
	"strings"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type ListConnectionModels struct {
	connStore protocols.Store[string, types.LlmConnection]
	catalogs  map[string]protocols.LLMCatalog
}

func NewListConnectionModels(connStore protocols.Store[string, types.LlmConnection], catalogs map[string]protocols.LLMCatalog) *ListConnectionModels {
	return &ListConnectionModels{connStore: connStore, catalogs: catalogs}
}

func (uc *ListConnectionModels) Execute(ctx context.Context, connectionID string) ([]types.ProviderModel, error) {
	items, err := uc.connStore.Get(ctx, []string{connectionID})
	if err != nil {
		return nil, err
	}
	conn, ok := items[connectionID]
	if !ok {
		return nil, base.ErrNotFound
	}
	catalog, ok := uc.catalogs[strings.ToLower(strings.TrimSpace(conn.Provider))]
	if !ok || catalog == nil {
		return []types.ProviderModel{}, nil
	}
	models, err := catalog.ListModels(ctx, conn.BaseURL, conn.APIKey)
	if models == nil {
		models = []types.ProviderModel{}
	}
	return models, err
}
