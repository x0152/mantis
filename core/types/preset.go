package types

type Preset struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	ChatModelID     string   `json:"chatModelId"`
	SummaryModelID  string   `json:"summaryModelId"`
	ImageModelID    string   `json:"imageModelId"`
	FallbackModelID string   `json:"fallbackModelId"`
	Temperature     *float64 `json:"temperature"`
	SystemPrompt    string   `json:"systemPrompt"`
}
