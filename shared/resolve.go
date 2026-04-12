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

func ResolvePreset(ctx context.Context, presetStore protocols.Store[string, types.Preset], presetID string) (types.Preset, error) {
	if presetID == "" {
		return types.Preset{}, fmt.Errorf("preset is not set")
	}
	presets, err := presetStore.Get(ctx, []string{presetID})
	if err != nil {
		return types.Preset{}, err
	}
	p, ok := presets[presetID]
	if !ok {
		return types.Preset{}, fmt.Errorf("preset %s not found", presetID)
	}
	return p, nil
}

func ResolveModelViaPreset(ctx context.Context, presetStore protocols.Store[string, types.Preset], modelStore protocols.Store[string, types.Model], presetID string) (types.Model, error) {
	preset, err := ResolvePreset(ctx, presetStore, presetID)
	if err != nil {
		return types.Model{}, err
	}
	modelID := preset.ChatModelID
	if modelID == "" {
		modelID = preset.FallbackModelID
	}
	return ResolveModel(ctx, modelStore, modelID)
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
