package shared

import (
	"context"
	"fmt"
	"testing"

	"mantis/core/types"
)

type memStore[ID comparable, Entity any] struct {
	data map[ID]Entity
}

func (s *memStore[ID, Entity]) Create(_ context.Context, items []Entity) ([]Entity, error) {
	return items, nil
}
func (s *memStore[ID, Entity]) Get(_ context.Context, ids []ID) (map[ID]Entity, error) {
	out := make(map[ID]Entity)
	for _, id := range ids {
		if v, ok := s.data[id]; ok {
			out[id] = v
		}
	}
	return out, nil
}
func (s *memStore[ID, Entity]) List(_ context.Context, _ types.ListQuery) ([]Entity, error) {
	return nil, nil
}
func (s *memStore[ID, Entity]) Update(_ context.Context, items []Entity) ([]Entity, error) {
	return items, nil
}
func (s *memStore[ID, Entity]) Delete(_ context.Context, _ []ID) error {
	return nil
}

// --- ResolveModel ---

func TestResolveModel_Found(t *testing.T) {
	store := &memStore[string, types.Model]{
		data: map[string]types.Model{
			"m1": {ID: "m1", Name: "GPT-4"},
		},
	}
	m, err := ResolveModel(context.Background(), store, "m1")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "m1" || m.Name != "GPT-4" {
		t.Fatalf("unexpected model: %+v", m)
	}
}

func TestResolveModel_NotFound(t *testing.T) {
	store := &memStore[string, types.Model]{data: map[string]types.Model{}}
	_, err := ResolveModel(context.Background(), store, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveModel_EmptyID(t *testing.T) {
	_, err := ResolveModel(context.Background(), nil, "")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

// --- ResolvePreset ---

func TestResolvePreset_Found(t *testing.T) {
	store := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", Name: "Default", ChatModelID: "m1"},
		},
	}
	p, err := ResolvePreset(context.Background(), store, "pr1")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "pr1" || p.ChatModelID != "m1" {
		t.Fatalf("unexpected preset: %+v", p)
	}
}

func TestResolvePreset_NotFound(t *testing.T) {
	store := &memStore[string, types.Preset]{data: map[string]types.Preset{}}
	_, err := ResolvePreset(context.Background(), store, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolvePreset_EmptyID(t *testing.T) {
	_, err := ResolvePreset(context.Background(), nil, "")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

// --- ResolveModelViaPreset ---

func TestResolveModelViaPreset_ChatModel(t *testing.T) {
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", ChatModelID: "gpt-4", FallbackModelID: "claude"},
		},
	}
	models := &memStore[string, types.Model]{
		data: map[string]types.Model{
			"gpt-4":  {ID: "gpt-4", Name: "GPT-4"},
			"claude": {ID: "claude", Name: "Claude"},
		},
	}
	m, err := ResolveModelViaPreset(context.Background(), presets, models, "pr1")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "gpt-4" {
		t.Fatalf("expected gpt-4 (primary), got %s", m.ID)
	}
}

func TestResolveModelViaPreset_FallbackModel(t *testing.T) {
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", FallbackModelID: "claude"},
		},
	}
	models := &memStore[string, types.Model]{
		data: map[string]types.Model{
			"claude": {ID: "claude", Name: "Claude"},
		},
	}
	m, err := ResolveModelViaPreset(context.Background(), presets, models, "pr1")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "claude" {
		t.Fatalf("expected claude (fallback), got %s", m.ID)
	}
}

func TestResolveModelViaPreset_PresetNotFound(t *testing.T) {
	presets := &memStore[string, types.Preset]{data: map[string]types.Preset{}}
	models := &memStore[string, types.Model]{data: map[string]types.Model{}}
	_, err := ResolveModelViaPreset(context.Background(), presets, models, "missing")
	if err == nil {
		t.Fatal("expected error for missing preset")
	}
}

func TestResolveModelViaPreset_ModelNotFound(t *testing.T) {
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", ChatModelID: "unknown-model"},
		},
	}
	models := &memStore[string, types.Model]{data: map[string]types.Model{}}
	_, err := ResolveModelViaPreset(context.Background(), presets, models, "pr1")
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}

// --- ResolveConnection ---

func TestResolveConnection_Found(t *testing.T) {
	store := &memStore[string, types.LlmConnection]{
		data: map[string]types.LlmConnection{
			"c1": {ID: "c1", Provider: "openai"},
		},
	}
	c, err := ResolveConnection(context.Background(), store, "c1")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "c1" || c.Provider != "openai" {
		t.Fatalf("unexpected connection: %+v", c)
	}
}

func TestResolveConnection_NotFound(t *testing.T) {
	store := &memStore[string, types.LlmConnection]{data: map[string]types.LlmConnection{}}
	_, err := ResolveConnection(context.Background(), store, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveConnection_EmptyID(t *testing.T) {
	_, err := ResolveConnection(context.Background(), nil, "")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

type errorStore struct{}

func (s *errorStore) Create(_ context.Context, _ []types.Model) ([]types.Model, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore) Get(_ context.Context, _ []string) (map[string]types.Model, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore) List(_ context.Context, _ types.ListQuery) ([]types.Model, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore) Update(_ context.Context, _ []types.Model) ([]types.Model, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore) Delete(_ context.Context, _ []string) error {
	return fmt.Errorf("db error")
}

func TestResolveModel_StoreError(t *testing.T) {
	_, err := ResolveModel(context.Background(), &errorStore{}, "m1")
	if err == nil {
		t.Fatal("expected error from store")
	}
}
