package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrUnauthorized = errors.New("unauthorized")

type Identity struct {
	UserID string
	Name   string
}

type ctxKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
