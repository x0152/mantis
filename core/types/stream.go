package types

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type LLMUsage struct {
	PromptTokens     int `json:"promptTokens,omitempty"`
	CompletionTokens int `json:"completionTokens,omitempty"`
	TotalTokens      int `json:"totalTokens,omitempty"`
	CachedTokens     int `json:"cachedTokens,omitempty"`
}

type StreamEvent struct {
	Type       string
	Sequence   int
	Iteration  int
	Delta      string
	ToolID     string
	LogID      string
	ModelID    string
	ModelName  string
	PresetID   string
	PresetName string
	ModelRole  string
	ToolCalls  []ToolCall
	Usage      *LLMUsage
	IsFinal    bool
}
