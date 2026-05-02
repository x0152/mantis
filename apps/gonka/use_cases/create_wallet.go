package usecases

import (
	"context"

	"mantis/apps/gonka/inferenced"
)

type CreateWallet struct {
	runner *inferenced.Runner
}

func NewCreateWallet(runner *inferenced.Runner) *CreateWallet {
	return &CreateWallet{runner: runner}
}

func (uc *CreateWallet) Execute(ctx context.Context) (inferenced.Wallet, error) {
	return uc.runner.CreateWallet(ctx)
}
