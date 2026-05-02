package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"mantis/apps/gonka/inferenced"
	usecases "mantis/apps/gonka/use_cases"
	"mantis/core/base"
)

type UseCases struct {
	CreateWallet  *usecases.CreateWallet
	DeriveAddress *usecases.DeriveAddress
	GetBalance    *usecases.GetBalance
	GetAccount    *usecases.GetAccount
	GetConfig     *usecases.GetConfig
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "gonka-config", Method: http.MethodGet, Path: "/api/gonka/config"}, e.config)
	huma.Register(api, huma.Operation{OperationID: "gonka-create-wallet", Method: http.MethodPost, Path: "/api/gonka/wallet", DefaultStatus: 201}, e.createWallet)
	huma.Register(api, huma.Operation{OperationID: "gonka-derive-address", Method: http.MethodPost, Path: "/api/gonka/wallet/derive"}, e.deriveAddress)
	huma.Register(api, huma.Operation{OperationID: "gonka-wallet-balance", Method: http.MethodGet, Path: "/api/gonka/wallet/balance"}, e.balance)
	huma.Register(api, huma.Operation{OperationID: "gonka-wallet-account", Method: http.MethodGet, Path: "/api/gonka/wallet/account"}, e.account)
}

func (e *Endpoints) config(ctx context.Context, _ *struct{}) (*ConfigOutput, error) {
	cfg := e.uc.GetConfig.Execute(ctx)
	out := &ConfigOutput{}
	out.Body = cfg
	return out, nil
}

func (e *Endpoints) createWallet(ctx context.Context, _ *struct{}) (*WalletOutput, error) {
	wallet, err := e.uc.CreateWallet.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &WalletOutput{}
	out.Body.Address = wallet.Address
	out.Body.PrivateKeyHex = wallet.PrivateKeyHex
	out.Body.Mnemonic = wallet.Mnemonic
	out.Body.Words = wallet.Words
	return out, nil
}

func (e *Endpoints) deriveAddress(ctx context.Context, input *DeriveAddressInput) (*DeriveAddressOutput, error) {
	address, err := e.uc.DeriveAddress.Execute(ctx, input.Body.PrivateKeyHex)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &DeriveAddressOutput{}
	out.Body.Address = address
	return out, nil
}

func (e *Endpoints) balance(ctx context.Context, input *BalanceInput) (*BalanceOutput, error) {
	balance, err := e.uc.GetBalance.Execute(ctx, input.Address, input.NodeURL)
	if err != nil {
		return nil, mapErr(err)
	}
	return &BalanceOutput{Body: balance}, nil
}

func (e *Endpoints) account(ctx context.Context, input *AccountInput) (*AccountOutput, error) {
	acc, err := e.uc.GetAccount.Execute(ctx, input.Address, input.NodeURL)
	if err != nil {
		return nil, mapErr(err)
	}
	return &AccountOutput{Body: acc}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, inferenced.ErrNotInstalled):
		return huma.NewError(http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, base.ErrNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, base.ErrValidation):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, err.Error())
	}
}
