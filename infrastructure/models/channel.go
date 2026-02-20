package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

type ChannelRow struct {
	bun.BaseModel  `bun:"table:channels"`
	ID             string          `bun:"id,pk"`
	Type           string          `bun:"type"`
	Name           string          `bun:"name"`
	Token          string          `bun:"token"`
	ModelID        string          `bun:"model_id"`
	AllowedUserIDs json.RawMessage `bun:"allowed_user_ids,type:jsonb"`
}
