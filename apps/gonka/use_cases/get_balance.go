package usecases

import (
	"context"
	"fmt"
	"strings"

	"mantis/core/base"
	"mantis/infrastructure/gonkachain"
)

type Balance struct {
	Address string  `json:"address"`
	NodeURL string  `json:"nodeUrl"`
	Amount  string  `json:"amount"`
	GNK     float64 `json:"gnk"`
	Label   string  `json:"label"`
}

type GetBalance struct{}

func NewGetBalance() *GetBalance {
	return &GetBalance{}
}

func (uc *GetBalance) Execute(ctx context.Context, address, nodeURL string) (Balance, error) {
	addr := strings.TrimSpace(address)
	src := strings.TrimSpace(nodeURL)
	if addr == "" {
		return Balance{}, fmt.Errorf("%w: address is required", base.ErrValidation)
	}
	if src == "" {
		return Balance{}, fmt.Errorf("%w: nodeUrl is required", base.ErrValidation)
	}
	amount, err := gonkachain.QueryBalance(ctx, src, addr)
	if err != nil {
		return Balance{}, err
	}
	gnk := gonkachain.TokenFloat(amount)
	return Balance{
		Address: addr,
		NodeURL: src,
		Amount:  amount.String(),
		GNK:     gnk,
		Label:   fmt.Sprintf("%s GNK", gonkachain.FormatBalance(gnk)),
	}, nil
}
