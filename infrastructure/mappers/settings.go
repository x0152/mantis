package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func SettingsToRow(s types.Settings) models.SettingsRow {
	memories, _ := json.Marshal(s.UserMemories)
	return models.SettingsRow{
		ID:             s.ID,
		ChatPresetID:   s.ChatPresetID,
		ServerPresetID: s.ServerPresetID,
		MemoryEnabled:  s.MemoryEnabled,
		UserMemories:   memories,
	}
}

func SettingsFromRow(r models.SettingsRow) types.Settings {
	var memories []string
	_ = json.Unmarshal(r.UserMemories, &memories)
	if memories == nil {
		memories = []string{}
	}
	return types.Settings{
		ID:             r.ID,
		ChatPresetID:   r.ChatPresetID,
		ServerPresetID: r.ServerPresetID,
		MemoryEnabled:  r.MemoryEnabled,
		UserMemories:   memories,
	}
}
