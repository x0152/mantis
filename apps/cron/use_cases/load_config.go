package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"mantis/core/protocols"
	"mantis/core/types"
)

type AppConfig struct {
	Cron struct {
		Channel string `json:"channel"`
		Sender  string `json:"sender"`
		ModelID string `json:"model_id"`
	} `json:"cron"`
}

type LoadConfig struct {
	configStore protocols.Store[string, types.Config]
}

func NewLoadConfig(configStore protocols.Store[string, types.Config]) *LoadConfig {
	return &LoadConfig{configStore: configStore}
}

func (uc *LoadConfig) Execute(ctx context.Context) (AppConfig, error) {
	var out AppConfig
	if uc.configStore == nil {
		return out, fmt.Errorf("config store is nil")
	}
	cfg, err := uc.configStore.Get(ctx, []string{"default"})
	if err != nil {
		return out, err
	}
	def, ok := cfg["default"]
	if !ok {
		return out, fmt.Errorf("config %q not found", "default")
	}
	if err := json.Unmarshal(def.Data, &out); err != nil {
		return out, err
	}
	return out, nil
}
