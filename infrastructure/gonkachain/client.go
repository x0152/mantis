package gonkachain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const (
	Denom    = "ngonka"
	Decimals = 9

	defaultTimeout = 15 * time.Second
)

type Account struct {
	Address      string
	AccountNum   uint64
	Sequence     uint64
	PubKeyType   string
	HasPubKey    bool
}

func QueryBalance(ctx context.Context, sourceURL, address string) (*big.Int, error) {
	if strings.TrimSpace(sourceURL) == "" {
		return nil, fmt.Errorf("gonka source URL is required")
	}
	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf("gonka address is required")
	}
	url := chainAPIBase(sourceURL) + "/chain-api/cosmos/bank/v1beta1/balances/" + address

	body, err := getJSON(ctx, url)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balances"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode gonka balances: %w", err)
	}

	for _, b := range payload.Balances {
		if b.Denom != Denom {
			continue
		}
		v, ok := new(big.Int).SetString(b.Amount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid gonka balance amount: %q", b.Amount)
		}
		return v, nil
	}
	return big.NewInt(0), nil
}

func QueryAccount(ctx context.Context, sourceURL, address string) (Account, error) {
	if strings.TrimSpace(sourceURL) == "" {
		return Account{}, fmt.Errorf("gonka source URL is required")
	}
	if strings.TrimSpace(address) == "" {
		return Account{}, fmt.Errorf("gonka address is required")
	}
	url := chainAPIBase(sourceURL) + "/chain-api/cosmos/auth/v1beta1/accounts/" + address

	body, err := getJSON(ctx, url)
	if err != nil {
		return Account{}, err
	}

	var payload struct {
		Account struct {
			Address       string `json:"address"`
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
			PubKey        *struct {
				Type string `json:"@type"`
				Key  string `json:"key"`
			} `json:"pub_key"`
		} `json:"account"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Account{}, fmt.Errorf("decode gonka account: %w", err)
	}

	acc := Account{Address: payload.Account.Address}
	if payload.Account.AccountNumber != "" {
		fmt.Sscanf(payload.Account.AccountNumber, "%d", &acc.AccountNum)
	}
	if payload.Account.Sequence != "" {
		fmt.Sscanf(payload.Account.Sequence, "%d", &acc.Sequence)
	}
	if payload.Account.PubKey != nil && strings.TrimSpace(payload.Account.PubKey.Key) != "" {
		acc.HasPubKey = true
		acc.PubKeyType = payload.Account.PubKey.Type
	}
	return acc, nil
}

func TokenFloat(amount *big.Int) float64 {
	if amount == nil {
		return 0
	}
	denom := new(big.Float).SetFloat64(1)
	for i := 0; i < Decimals; i++ {
		denom.Mul(denom, big.NewFloat(10))
	}
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(amount), denom).Float64()
	return f
}

func FormatBalance(v float64) string {
	switch {
	case v >= 1000:
		return fmt.Sprintf("%.0f", v)
	case v >= 1:
		return fmt.Sprintf("%.2f", v)
	case v > 0:
		return fmt.Sprintf("%.4f", v)
	default:
		return "0"
	}
}

func chainAPIBase(sourceURL string) string {
	return strings.TrimSuffix(strings.TrimRight(sourceURL, "/"), "/v1")
}

func getJSON(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: defaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gonka chain request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return body, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gonka chain API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
