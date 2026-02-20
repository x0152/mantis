package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func ConnectionToRow(c types.Connection) models.ConnectionRow {
	memories, _ := json.Marshal(c.Memories)
	return models.ConnectionRow{
		ID:          c.ID,
		Type:        c.Type,
		Name:        c.Name,
		Description: c.Description,
		ModelID:     c.ModelID,
		Config:      c.Config,
		Memories:    memories,
	}
}

func ConnectionFromRow(r models.ConnectionRow) types.Connection {
	var memories []types.Memory
	_ = json.Unmarshal(r.Memories, &memories)
	if memories == nil {
		memories = []types.Memory{}
	}
	return types.Connection{
		ID:          r.ID,
		Type:        r.Type,
		Name:        r.Name,
		Description: r.Description,
		ModelID:     r.ModelID,
		Config:      r.Config,
		Memories:    memories,
	}
}
