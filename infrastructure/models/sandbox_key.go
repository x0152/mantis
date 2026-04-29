package models

import (
	"time"

	"github.com/uptrace/bun"
)

type SandboxKeyRow struct {
	bun.BaseModel `bun:"table:sandbox_keys"`
	ID            string    `bun:"id,pk"`
	PrivateKey    string    `bun:"private_key"`
	PublicKey     string    `bun:"public_key"`
	CreatedAt     time.Time `bun:"created_at,nullzero,default:now()"`
}
