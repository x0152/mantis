package usecases

import (
	"context"
	"fmt"
	"strings"

	"mantis/core/base"
	"mantis/infrastructure/gonkachain"
)

type AccountStatus struct {
	Address         string `json:"address"`
	Exists          bool   `json:"exists"`
	PubKeyPublished bool   `json:"pubKeyPublished"`
	AccountNumber   uint64 `json:"accountNumber,omitempty"`
}

type GetAccount struct{}

func NewGetAccount() *GetAccount {
	return &GetAccount{}
}

func (uc *GetAccount) Execute(ctx context.Context, address, nodeURL string) (AccountStatus, error) {
	addr := strings.TrimSpace(address)
	src := strings.TrimSpace(nodeURL)
	if addr == "" {
		return AccountStatus{}, fmt.Errorf("%w: address is required", base.ErrValidation)
	}
	if src == "" {
		return AccountStatus{}, fmt.Errorf("%w: nodeUrl is required", base.ErrValidation)
	}
	acc, err := gonkachain.QueryAccount(ctx, src, addr)
	if err != nil {
		return AccountStatus{}, err
	}
	return AccountStatus{
		Address:         addr,
		Exists:          acc.Address != "",
		PubKeyPublished: acc.HasPubKey,
		AccountNumber:   acc.AccountNum,
	}, nil
}
