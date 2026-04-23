package types

import "time"

type User struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	APIKeyHash string    `json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
}
