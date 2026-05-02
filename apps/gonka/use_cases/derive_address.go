package usecases

import (
	"context"
	"fmt"
	"strings"

	gonkaopenai "github.com/gonka-ai/gonka-openai/go"

	"mantis/core/base"
)

type DeriveAddress struct{}

func NewDeriveAddress() *DeriveAddress {
	return &DeriveAddress{}
}

func (uc *DeriveAddress) Execute(_ context.Context, privateKeyHex string) (string, error) {
	pk := strings.TrimSpace(privateKeyHex)
	if pk == "" {
		return "", fmt.Errorf("%w: private key is required", base.ErrValidation)
	}
	address, err := gonkaopenai.GonkaAddress(pk)
	if err != nil {
		return "", fmt.Errorf("%w: invalid private key: %s", base.ErrValidation, err)
	}
	return address, nil
}
