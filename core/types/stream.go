package types

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
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
	IsFinal    bool
}
