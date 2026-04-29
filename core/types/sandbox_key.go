package types

import "time"

type SandboxKey struct {
	ID         string    `json:"id"`
	PrivateKey string    `json:"privateKey"`
	PublicKey  string    `json:"publicKey"`
	CreatedAt  time.Time `json:"createdAt"`
}
