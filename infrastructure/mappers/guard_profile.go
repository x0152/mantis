package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func GuardProfileToRow(p types.GuardProfile) models.GuardProfileRow {
	caps, _ := json.Marshal(p.Capabilities)
	cmds, _ := json.Marshal(p.Commands)
	return models.GuardProfileRow{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Builtin:      p.Builtin,
		Capabilities: caps,
		Commands:     cmds,
	}
}

func GuardProfileFromRow(r models.GuardProfileRow) types.GuardProfile {
	var caps types.GuardCapabilities
	_ = json.Unmarshal(r.Capabilities, &caps)
	var cmds []types.CommandRule
	_ = json.Unmarshal(r.Commands, &cmds)
	if cmds == nil {
		cmds = []types.CommandRule{}
	}
	return types.GuardProfile{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		Builtin:      r.Builtin,
		Capabilities: caps,
		Commands:     cmds,
	}
}
