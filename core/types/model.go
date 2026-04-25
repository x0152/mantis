package types

type Model struct {
	ID            string `json:"id"`
	ConnectionID  string `json:"connectionId"`
	Name          string `json:"name"`
	ThinkingMode  string `json:"thinkingMode"`
	ContextWindow int    `json:"contextWindow"`
	ReserveTokens int    `json:"reserveTokens"`
	CompactTokens int    `json:"compactTokens"`
}
