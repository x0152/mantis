package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func ChannelToRow(c types.Channel) models.ChannelRow {
	allowed, _ := json.Marshal(c.AllowedUserIDs)
	return models.ChannelRow{
		ID:             c.ID,
		Type:           c.Type,
		Name:           c.Name,
		Token:          c.Token,
		ModelID:        c.ModelID,
		AllowedUserIDs: allowed,
	}
}

func ChannelFromRow(r models.ChannelRow) types.Channel {
	var allowed []int64
	_ = json.Unmarshal(r.AllowedUserIDs, &allowed)
	if allowed == nil {
		allowed = []int64{}
	}
	return types.Channel{
		ID:             r.ID,
		Type:           r.Type,
		Name:           r.Name,
		Token:          r.Token,
		ModelID:        r.ModelID,
		AllowedUserIDs: allowed,
	}
}
