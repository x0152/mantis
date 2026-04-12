package types

type Settings struct {
	ID             string   `json:"id"`
	ChatPresetID   string   `json:"chatPresetId"`
	ServerPresetID string   `json:"serverPresetId"`
	MemoryEnabled  bool     `json:"memoryEnabled"`
	UserMemories   []string `json:"userMemories"`
}
