package types

type Model struct {
	ID           string `json:"id"`
	ConnectionID string `json:"connectionId"`
	Name         string `json:"name"`
	ThinkingMode string `json:"thinkingMode"`
}
