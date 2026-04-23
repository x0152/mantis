package models

import (
	"time"

	"github.com/uptrace/bun"
)

type UserRow struct {
	bun.BaseModel `bun:"table:users"`
	ID            string    `bun:"id,pk"`
	Name          string    `bun:"name"`
	APIKeyHash    string    `bun:"api_key_hash"`
	CreatedAt     time.Time `bun:"created_at"`
}
