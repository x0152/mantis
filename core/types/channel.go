package types

type Channel struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	Token          string  `json:"token"`
	ModelID        string  `json:"modelId"`
	PresetID       string  `json:"presetId"`
	AllowedUserIDs []int64 `json:"allowedUserIds"`
}
