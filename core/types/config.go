package types

import "encoding/json"

type LlmConnection struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	BaseURL  string `json:"baseUrl"`
	APIKey   string `json:"apiKey"`
}

type Config struct {
	ID   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}
