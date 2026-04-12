package types

import "encoding/json"

type Skill struct {
	ID           string          `json:"id"`
	ConnectionID string          `json:"connectionId"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Parameters   json.RawMessage `json:"parameters"`
	Script       string          `json:"script"`
}
