package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type Vision struct {
	client *http.Client
}

func NewVision() *Vision {
	transport := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}
	return &Vision{client: &http.Client{Transport: transport}}
}

func (v *Vision) Describe(ctx context.Context, baseURL, apiKey, model string, image []byte, format, prompt string) (string, error) {
	mime := "image/" + format
	if format == "jpg" {
		mime = "image/jpeg"
	}
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(image))

	payload := map[string]any{
		"model":      model,
		"max_tokens": 4096,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": prompt},
					{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
				},
			},
		},
		"enable_thinking": false,
		"chat_template_kwargs": map[string]any{
			"enable_thinking": false,
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Vision API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content          json.RawMessage `json:"content"`
				ReasoningContent string          `json:"reasoning_content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("vision parse error: %w; body: %s", err, preview(respBody))
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from vision model; body: %s", preview(respBody))
	}

	msg := result.Choices[0].Message
	content := parseVisionContent(msg.Content)
	if content == "" {
		content = strings.TrimSpace(msg.ReasoningContent)
	}

	if content == "" {
		log.Printf("vision: empty content, model=%s finish_reason=%q raw=%s",
			model, result.Choices[0].FinishReason, preview(respBody))
		if result.Choices[0].FinishReason == "length" {
			return "", fmt.Errorf("vision model hit max_tokens before producing output; try a smaller prompt/image or a different model")
		}
	}
	return content, nil
}

// parseVisionContent handles OpenAI-compatible `content` that can be:
// - a plain string: "hello"
// - an array of parts: [{"type":"text","text":"hello"}, {"type":"output_text","text":"..."}]
// - null/empty
func parseVisionContent(raw json.RawMessage) string {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	if len(raw) > 0 && raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return strings.TrimSpace(s)
		}
		return ""
	}
	if len(raw) > 0 && raw[0] == '[' {
		var parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &parts); err != nil {
			return ""
		}
		var b strings.Builder
		for _, p := range parts {
			if p.Text == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(p.Text)
		}
		return strings.TrimSpace(b.String())
	}
	return ""
}

func preview(b []byte) string {
	const max = 600
	s := string(b)
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
