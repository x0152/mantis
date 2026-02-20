package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Input struct {
	ExplicitModelID string
	ChannelID       string

	// Optional config fallback.
	// If ConfigID is empty, "default" is used.
	ConfigID   string
	ConfigPath []string
}

type Output struct {
	ModelID string
	Source  string // explicit | channel | config | none
}

type Resolver struct {
	channelStore protocols.Store[string, types.Channel]
	configStore  protocols.Store[string, types.Config]
}

func NewResolver(
	channelStore protocols.Store[string, types.Channel],
	configStore protocols.Store[string, types.Config],
) *Resolver {
	return &Resolver{channelStore: channelStore, configStore: configStore}
}

func (r *Resolver) Execute(ctx context.Context, in Input) (Output, error) {
	if id := strings.TrimSpace(in.ExplicitModelID); id != "" {
		return Output{ModelID: id, Source: "explicit"}, nil
	}

	if strings.TrimSpace(in.ChannelID) != "" {
		id, err := r.resolveFromChannel(ctx, in.ChannelID)
		if err != nil {
			return Output{}, err
		}
		if id != "" {
			return Output{ModelID: id, Source: "channel"}, nil
		}
	}

	if len(in.ConfigPath) > 0 {
		id, err := r.resolveFromConfig(ctx, in.ConfigID, in.ConfigPath)
		if err != nil {
			return Output{}, err
		}
		if id != "" {
			return Output{ModelID: id, Source: "config"}, nil
		}
	}

	return Output{Source: "none"}, nil
}

func (r *Resolver) resolveFromChannel(ctx context.Context, channelID string) (string, error) {
	if r.channelStore == nil {
		return "", fmt.Errorf("channel store is nil")
	}
	existing, err := r.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return "", err
	}
	ch, ok := existing[channelID]
	if !ok {
		return "", fmt.Errorf("channel %q not found", channelID)
	}
	return strings.TrimSpace(ch.ModelID), nil
}

func (r *Resolver) resolveFromConfig(ctx context.Context, configID string, path []string) (string, error) {
	if r.configStore == nil {
		return "", fmt.Errorf("config store is nil")
	}
	if strings.TrimSpace(configID) == "" {
		configID = "default"
	}
	cfg, err := r.configStore.Get(ctx, []string{configID})
	if err != nil {
		return "", err
	}
	item, ok := cfg[configID]
	if !ok {
		return "", fmt.Errorf("config %q not found", configID)
	}

	var raw any
	if err := json.Unmarshal(item.Data, &raw); err != nil {
		return "", err
	}

	cur := raw
	for _, key := range path {
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", nil
		}
		next, ok := obj[key]
		if !ok {
			return "", nil
		}
		cur = next
	}
	val, ok := cur.(string)
	if !ok {
		return "", nil
	}
	return strings.TrimSpace(val), nil
}
