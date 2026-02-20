package shared

import (
	"context"
	"fmt"

	"mantis/core/protocols"
	"mantis/core/types"
)

func ResolveModel(ctx context.Context, store protocols.Store[string, types.Model], modelID string) (types.Model, error) {
	if modelID == "" {
		return types.Model{}, fmt.Errorf("LLM is not connected")
	}
	models, err := store.Get(ctx, []string{modelID})
	if err != nil {
		return types.Model{}, err
	}
	model, ok := models[modelID]
	if !ok {
		return types.Model{}, fmt.Errorf("model %s not found", modelID)
	}
	return model, nil
}

func ResolveConnection(ctx context.Context, store protocols.Store[string, types.LlmConnection], connectionID string) (types.LlmConnection, error) {
	if connectionID == "" {
		return types.LlmConnection{}, fmt.Errorf("llm connection is not set")
	}
	conns, err := store.Get(ctx, []string{connectionID})
	if err != nil {
		return types.LlmConnection{}, err
	}
	conn, ok := conns[connectionID]
	if !ok {
		return types.LlmConnection{}, fmt.Errorf("llm connection %q not found", connectionID)
	}
	return conn, nil
}
