package api

import (
	usecases "mantis/apps/gonka/use_cases"
)

type WalletOutput struct {
	Body struct {
		Address       string   `json:"address"`
		PrivateKeyHex string   `json:"privateKeyHex"`
		Mnemonic      string   `json:"mnemonic"`
		Words         []string `json:"words"`
	}
}

type DeriveAddressInput struct {
	Body struct {
		PrivateKeyHex string `json:"privateKeyHex" required:"true" minLength:"1"`
	}
}

type DeriveAddressOutput struct {
	Body struct {
		Address string `json:"address"`
	}
}

type BalanceInput struct {
	Address string `query:"address" required:"true"`
	NodeURL string `query:"nodeUrl" required:"true"`
}

type BalanceOutput struct {
	Body usecases.Balance
}

type AccountInput struct {
	Address string `query:"address" required:"true"`
	NodeURL string `query:"nodeUrl" required:"true"`
}

type AccountOutput struct {
	Body usecases.AccountStatus
}

type ConfigOutput struct {
	Body usecases.Config
}
