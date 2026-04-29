package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func SandboxKeyToRow(k types.SandboxKey) models.SandboxKeyRow {
	return models.SandboxKeyRow{
		ID:         k.ID,
		PrivateKey: k.PrivateKey,
		PublicKey:  k.PublicKey,
		CreatedAt:  k.CreatedAt,
	}
}

func SandboxKeyFromRow(r models.SandboxKeyRow) types.SandboxKey {
	return types.SandboxKey{
		ID:         r.ID,
		PrivateKey: r.PrivateKey,
		PublicKey:  r.PublicKey,
		CreatedAt:  r.CreatedAt,
	}
}
