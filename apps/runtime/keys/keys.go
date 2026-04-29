package keys

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

	"mantis/core/protocols"
	"mantis/core/types"
)

const defaultID = "default"

type Issuer struct {
	store protocols.Store[string, types.SandboxKey]
}

func NewIssuer(store protocols.Store[string, types.SandboxKey]) *Issuer {
	return &Issuer{store: store}
}

func (i *Issuer) Ensure(ctx context.Context) (types.SandboxKey, error) {
	existing, err := i.store.Get(ctx, []string{defaultID})
	if err != nil {
		return types.SandboxKey{}, fmt.Errorf("load sandbox key: %w", err)
	}
	if k, ok := existing[defaultID]; ok {
		return k, nil
	}

	priv, pub, err := generateEd25519()
	if err != nil {
		return types.SandboxKey{}, err
	}
	key := types.SandboxKey{
		ID:         defaultID,
		PrivateKey: priv,
		PublicKey:  pub,
		CreatedAt:  time.Now().UTC(),
	}
	created, err := i.store.Create(ctx, []types.SandboxKey{key})
	if err != nil {
		return types.SandboxKey{}, fmt.Errorf("persist sandbox key: %w", err)
	}
	if len(created) == 0 {
		return types.SandboxKey{}, errors.New("persist sandbox key: empty result")
	}
	return created[0], nil
}

func generateEd25519() (privatePEM, publicAuthorized string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("generate ed25519: %w", err)
	}
	pemBlock, err := ssh.MarshalPrivateKey(priv, "mantis-sandbox")
	if err != nil {
		return "", "", fmt.Errorf("marshal private key: %w", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", "", fmt.Errorf("marshal public key: %w", err)
	}
	return string(pem.EncodeToMemory(pemBlock)), string(ssh.MarshalAuthorizedKey(sshPub)), nil
}
