package usecases

import (
	"context"

	"mantis/apps/gonka/inferenced"
)

type Config struct {
	DefaultNodeURL      string `json:"defaultNodeUrl"`
	InferencedAvailable bool   `json:"inferencedAvailable"`
	HasPresetPrivateKey bool   `json:"hasPresetPrivateKey"`
	HasPresetNodeURL    bool   `json:"hasPresetNodeUrl"`
	MinBalanceGNK       string `json:"minBalanceGnk"`
}

type GetConfig struct {
	runner          *inferenced.Runner
	defaultNodeURL  string
	hasPresetPK     bool
	hasPresetNodeURL bool
	minBalance       string
}

type GetConfigOptions struct {
	DefaultNodeURL   string
	HasPresetPK      bool
	HasPresetNodeURL bool
	MinBalanceGNK    string
}

func NewGetConfig(runner *inferenced.Runner, opts GetConfigOptions) *GetConfig {
	return &GetConfig{
		runner:           runner,
		defaultNodeURL:   opts.DefaultNodeURL,
		hasPresetPK:      opts.HasPresetPK,
		hasPresetNodeURL: opts.HasPresetNodeURL,
		minBalance:       opts.MinBalanceGNK,
	}
}

func (uc *GetConfig) Execute(_ context.Context) Config {
	return Config{
		DefaultNodeURL:      uc.defaultNodeURL,
		InferencedAvailable: uc.runner != nil && uc.runner.Available(),
		HasPresetPrivateKey: uc.hasPresetPK,
		HasPresetNodeURL:    uc.hasPresetNodeURL,
		MinBalanceGNK:       uc.minBalance,
	}
}
