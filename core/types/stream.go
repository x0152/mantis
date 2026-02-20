package types

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type StreamEvent struct {
	Type      string
	Sequence  int
	Iteration int
	Delta     string
	ToolID    string
	LogID     string
	ModelName string
	ToolCalls []ToolCall
	IsFinal   bool
}
