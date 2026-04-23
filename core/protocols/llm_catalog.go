package protocols

import (
	"context"

	"mantis/core/types"
)

type LLMCatalog interface {
	ListModels(ctx context.Context, baseURL, apiKey string) ([]types.ProviderModel, error)
	GetInferenceLimit(ctx context.Context, baseURL, apiKey string) (types.InferenceLimit, error)
}
