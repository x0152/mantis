package model

import (
	"context"
	"fmt"
	"strings"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Input struct {
	ExplicitModelID string
	ChannelID       string
	DefaultPreset   string
}

type Output struct {
	ModelID    string
	Source     string // explicit | channel | settings | none
	PresetID   string
	PresetName string
	ModelRole  string // primary | fallback | explicit | legacy
}

type Resolver struct {
	channelStore  protocols.Store[string, types.Channel]
	settingsStore protocols.Store[string, types.Settings]
	presetStore   protocols.Store[string, types.Preset]
}

func NewResolver(
	channelStore protocols.Store[string, types.Channel],
	settingsStore protocols.Store[string, types.Settings],
	presetStore protocols.Store[string, types.Preset],
) *Resolver {
	return &Resolver{channelStore: channelStore, settingsStore: settingsStore, presetStore: presetStore}
}

func (r *Resolver) Execute(ctx context.Context, in Input) (Output, error) {
	if id := strings.TrimSpace(in.ExplicitModelID); id != "" {
		return Output{ModelID: id, Source: "explicit", ModelRole: "explicit"}, nil
	}

	var channelLegacy Output
	if strings.TrimSpace(in.ChannelID) != "" {
		out, err := r.resolveFromChannel(ctx, in.ChannelID)
		if err != nil {
			return Output{}, err
		}
		if out.ModelID != "" {
			if in.DefaultPreset != "" && out.ModelRole == "legacy" {
				channelLegacy = out
			} else {
				out.Source = "channel"
				return out, nil
			}
		}
	}

	if in.DefaultPreset != "" {
		presetID, err := r.resolveDefaultPresetID(ctx, in.DefaultPreset)
		if err == nil && presetID != "" {
			if presetOut := r.resolvePresetToModel(ctx, presetID); presetOut.ModelID != "" {
				presetOut.Source = "settings"
				return presetOut, nil
			}
		}
	}

	if channelLegacy.ModelID != "" {
		channelLegacy.Source = "channel"
		return channelLegacy, nil
	}

	return Output{Source: "none"}, nil
}

func (r *Resolver) resolvePresetToModel(ctx context.Context, presetID string) Output {
	if r.presetStore == nil || presetID == "" {
		return Output{}
	}
	presets, err := r.presetStore.Get(ctx, []string{presetID})
	if err != nil {
		return Output{}
	}
	p, ok := presets[presetID]
	if !ok {
		return Output{}
	}
	if strings.TrimSpace(p.ChatModelID) != "" {
		return Output{
			ModelID:  strings.TrimSpace(p.ChatModelID),
			PresetID: presetID, PresetName: p.Name,
			ModelRole: "primary",
		}
	}
	if strings.TrimSpace(p.FallbackModelID) != "" {
		return Output{
			ModelID:  strings.TrimSpace(p.FallbackModelID),
			PresetID: presetID, PresetName: p.Name,
			ModelRole: "fallback",
		}
	}
	return Output{
		PresetID: presetID, PresetName: p.Name,
	}
}

func (r *Resolver) resolveFromChannel(ctx context.Context, channelID string) (Output, error) {
	if r.channelStore == nil {
		return Output{}, fmt.Errorf("channel store is nil")
	}
	existing, err := r.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return Output{}, err
	}
	ch, ok := existing[channelID]
	if !ok {
		return Output{}, fmt.Errorf("channel %q not found", channelID)
	}
	if pid := strings.TrimSpace(ch.PresetID); pid != "" {
		if out := r.resolvePresetToModel(ctx, pid); out.ModelID != "" {
			return out, nil
		}
	}
	if mid := strings.TrimSpace(ch.ModelID); mid != "" {
		return Output{ModelID: mid, ModelRole: "legacy"}, nil
	}
	return Output{}, nil
}

func (r *Resolver) resolveDefaultPresetID(ctx context.Context, slot string) (string, error) {
	if r.settingsStore == nil {
		return "", fmt.Errorf("settings store is nil")
	}
	cfg, err := r.settingsStore.Get(ctx, []string{"default"})
	if err != nil {
		return "", err
	}
	item, ok := cfg["default"]
	if !ok {
		return "", fmt.Errorf("settings %q not found", "default")
	}
	switch strings.ToLower(strings.TrimSpace(slot)) {
	case "chat":
		return strings.TrimSpace(item.ChatPresetID), nil
	case "server":
		return strings.TrimSpace(item.ServerPresetID), nil
	default:
		return "", fmt.Errorf("unknown default preset %q", slot)
	}
}
