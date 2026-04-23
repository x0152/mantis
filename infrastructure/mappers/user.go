package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func UserToRow(u types.User) models.UserRow {
	return models.UserRow{
		ID:         u.ID,
		Name:       u.Name,
		APIKeyHash: u.APIKeyHash,
		CreatedAt:  u.CreatedAt,
	}
}

func UserFromRow(r models.UserRow) types.User {
	return types.User{
		ID:         r.ID,
		Name:       r.Name,
		APIKeyHash: r.APIKeyHash,
		CreatedAt:  r.CreatedAt,
	}
}
