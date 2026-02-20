package types

type GuardRule struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Pattern      string `json:"pattern"`
	ConnectionID string `json:"connectionId"`
	Enabled      bool   `json:"enabled"`
}
