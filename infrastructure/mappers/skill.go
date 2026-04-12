package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func SkillToRow(s types.Skill) models.SkillRow {
	return models.SkillRow{
		ID:           s.ID,
		ConnectionID: s.ConnectionID,
		Name:         s.Name,
		Description:  s.Description,
		Parameters:   s.Parameters,
		Script:       s.Script,
	}
}

func SkillFromRow(r models.SkillRow) types.Skill {
	return types.Skill{
		ID:           r.ID,
		ConnectionID: r.ConnectionID,
		Name:         r.Name,
		Description:  r.Description,
		Parameters:   r.Parameters,
		Script:       r.Script,
	}
}
