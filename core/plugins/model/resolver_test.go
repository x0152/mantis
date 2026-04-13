package model

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

func TestResolve_ExplicitModel(t *testing.T) {
	r := NewResolver(nil, nil, nil)
	out, err := r.Execute(context.Background(), Input{ExplicitModelID: "gpt-4"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "gpt-4" {
		t.Fatalf("expected gpt-4, got %s", out.ModelID)
	}
	if out.Source != "explicit" {
		t.Fatalf("expected explicit source, got %s", out.Source)
	}
	if out.ModelRole != "explicit" {
		t.Fatalf("expected explicit role, got %s", out.ModelRole)
	}
}

func TestResolve_ChannelWithPreset(t *testing.T) {
	channels := &memStore[string, types.Channel]{
		data: map[string]types.Channel{
			"ch1": {ID: "ch1", PresetID: "pr1"},
		},
	}
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", Name: "Fast", ChatModelID: "claude-3"},
		},
	}
	r := NewResolver(channels, nil, presets)
	out, err := r.Execute(context.Background(), Input{ChannelID: "ch1"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "claude-3" {
		t.Fatalf("expected claude-3, got %s", out.ModelID)
	}
	if out.Source != "channel" {
		t.Fatalf("expected channel source, got %s", out.Source)
	}
	if out.PresetID != "pr1" {
		t.Fatalf("expected preset pr1, got %s", out.PresetID)
	}
	if out.ModelRole != "primary" {
		t.Fatalf("expected primary role, got %s", out.ModelRole)
	}
}

func TestResolve_ChannelLegacy_OverriddenBySettingsPreset(t *testing.T) {
	channels := &memStore[string, types.Channel]{
		data: map[string]types.Channel{
			"ch1": {ID: "ch1", ModelID: "legacy-model"},
		},
	}
	settings := &memStore[string, types.Settings]{
		data: map[string]types.Settings{
			"default": {ID: "default", ChatPresetID: "pr1"},
		},
	}
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", Name: "Default", ChatModelID: "preferred-model"},
		},
	}
	r := NewResolver(channels, settings, presets)
	out, err := r.Execute(context.Background(), Input{ChannelID: "ch1", DefaultPreset: "chat"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "preferred-model" {
		t.Fatalf("settings preset should override legacy channel; got %s", out.ModelID)
	}
	if out.Source != "settings" {
		t.Fatalf("expected settings source, got %s", out.Source)
	}
}

func TestResolve_ChannelLegacy_Fallback(t *testing.T) {
	channels := &memStore[string, types.Channel]{
		data: map[string]types.Channel{
			"ch1": {ID: "ch1", ModelID: "legacy-model"},
		},
	}
	r := NewResolver(channels, nil, nil)
	out, err := r.Execute(context.Background(), Input{ChannelID: "ch1"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "legacy-model" {
		t.Fatalf("expected legacy-model, got %s", out.ModelID)
	}
	if out.Source != "channel" {
		t.Fatalf("expected channel source, got %s", out.Source)
	}
	if out.ModelRole != "legacy" {
		t.Fatalf("expected legacy role, got %s", out.ModelRole)
	}
}

func TestResolve_SettingsPreset_Chat(t *testing.T) {
	settings := &memStore[string, types.Settings]{
		data: map[string]types.Settings{
			"default": {ID: "default", ChatPresetID: "pr1"},
		},
	}
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", Name: "GPT", ChatModelID: "gpt-4o"},
		},
	}
	r := NewResolver(nil, settings, presets)
	out, err := r.Execute(context.Background(), Input{DefaultPreset: "chat"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", out.ModelID)
	}
	if out.Source != "settings" {
		t.Fatalf("expected settings, got %s", out.Source)
	}
}

func TestResolve_SettingsPreset_FallbackModel(t *testing.T) {
	settings := &memStore[string, types.Settings]{
		data: map[string]types.Settings{
			"default": {ID: "default", ChatPresetID: "pr1"},
		},
	}
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"pr1": {ID: "pr1", Name: "Backup", FallbackModelID: "fallback-model"},
		},
	}
	r := NewResolver(nil, settings, presets)
	out, err := r.Execute(context.Background(), Input{DefaultPreset: "chat"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "fallback-model" {
		t.Fatalf("expected fallback-model, got %s", out.ModelID)
	}
	if out.ModelRole != "fallback" {
		t.Fatalf("expected fallback role, got %s", out.ModelRole)
	}
}

func TestResolve_NothingConfigured(t *testing.T) {
	r := NewResolver(nil, nil, nil)
	out, err := r.Execute(context.Background(), Input{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Source != "none" {
		t.Fatalf("expected none, got %s", out.Source)
	}
	if out.ModelID != "" {
		t.Fatalf("expected empty model, got %s", out.ModelID)
	}
}

func TestResolve_ChannelNotFound(t *testing.T) {
	channels := &memStore[string, types.Channel]{
		data: map[string]types.Channel{},
	}
	r := NewResolver(channels, nil, nil)
	_, err := r.Execute(context.Background(), Input{ChannelID: "missing"})
	if err == nil {
		t.Fatal("expected error for missing channel")
	}
}

func TestResolve_ExplicitOverridesChannel(t *testing.T) {
	channels := &memStore[string, types.Channel]{
		data: map[string]types.Channel{
			"ch1": {ID: "ch1", ModelID: "channel-model"},
		},
	}
	r := NewResolver(channels, nil, nil)
	out, err := r.Execute(context.Background(), Input{
		ExplicitModelID: "override",
		ChannelID:       "ch1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "override" {
		t.Fatalf("explicit should override channel, got %s", out.ModelID)
	}
}

func TestResolve_ServerPreset(t *testing.T) {
	settings := &memStore[string, types.Settings]{
		data: map[string]types.Settings{
			"default": {ID: "default", ServerPresetID: "sp1"},
		},
	}
	presets := &memStore[string, types.Preset]{
		data: map[string]types.Preset{
			"sp1": {ID: "sp1", Name: "Server", ChatModelID: "server-model"},
		},
	}
	r := NewResolver(nil, settings, presets)
	out, err := r.Execute(context.Background(), Input{DefaultPreset: "server"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ModelID != "server-model" {
		t.Fatalf("expected server-model, got %s", out.ModelID)
	}
}

type errorStore[ID comparable, Entity any] struct{}

func (s *errorStore[ID, Entity]) Create(_ context.Context, _ []Entity) ([]Entity, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore[ID, Entity]) Get(_ context.Context, _ []ID) (map[ID]Entity, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore[ID, Entity]) List(_ context.Context, _ types.ListQuery) ([]Entity, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore[ID, Entity]) Update(_ context.Context, _ []Entity) ([]Entity, error) {
	return nil, fmt.Errorf("db error")
}
func (s *errorStore[ID, Entity]) Delete(_ context.Context, _ []ID) error {
	return fmt.Errorf("db error")
}

func TestResolve_ChannelStoreError(t *testing.T) {
	r := NewResolver(&errorStore[string, types.Channel]{}, nil, nil)
	_, err := r.Execute(context.Background(), Input{ChannelID: "ch1"})
	if err == nil {
		t.Fatal("expected error from store")
	}
}
