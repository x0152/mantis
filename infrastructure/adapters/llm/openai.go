package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

type OpenAI struct {
	client *http.Client
}

func NewOpenAI() *OpenAI {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	return &OpenAI{client: &http.Client{Transport: transport}}
}

type reqMessage struct {
	Role       string        `json:"role"`
	Content    *string       `json:"content"`
	ToolCalls  []reqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type reqToolCall struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Function reqFunction `json:"function"`
}

type reqFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type reqTool struct {
	Type     string     `json:"type"`
	Function reqToolDef `json:"function"`
}

type reqToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type chatReq struct {
	Model         string        `json:"model"`
	Messages      []reqMessage  `json:"messages"`
	Tools         []reqTool     `json:"tools,omitempty"`
	Stream        bool          `json:"stream"`
	StreamOptions streamOptions `json:"stream_options"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type streamDelta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
	ToolCalls        []struct {
		Index    int    `json:"index"`
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls"`
}

type streamChunk struct {
	Choices []struct {
		Delta        streamDelta `json:"delta"`
		FinishReason *string     `json:"finish_reason"`
	} `json:"choices"`
	Usage *streamUsage `json:"usage"`
}

type streamUsage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	PromptTokensDetails     *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
}

type modelsResp struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (o *OpenAI) ListModels(ctx context.Context, baseURL, apiKey string) ([]types.ProviderModel, error) {
	base := strings.TrimRight(baseURL, "/")
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM models API error %d: %s", resp.StatusCode, string(b))
	}
	var payload modelsResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	items := make([]types.ProviderModel, 0, len(payload.Data))
	for _, m := range payload.Data {
		if m.ID == "" {
			continue
		}
		items = append(items, types.ProviderModel{ID: m.ID})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (o *OpenAI) GetInferenceLimit(_ context.Context, _ string, _ string) (types.InferenceLimit, error) {
	return types.InferenceLimit{
		Type:  "unlimited",
		Label: "No inference limit reported",
	}, nil
}

func (o *OpenAI) ChatStream(ctx context.Context, _ string, baseURL, apiKey string, messages []protocols.LLMMessage, model string, tools []types.Tool, thinkingMode string) (<-chan types.StreamEvent, error) {
	msgs := buildMessages(messages)
	reqTools := buildTools(tools)

	payload := chatReq{Model: model, Messages: msgs, Stream: true, StreamOptions: streamOptions{IncludeUsage: true}}
	if len(reqTools) > 0 {
		payload.Tools = reqTools
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")

	log.Printf("LLM stream request: POST %s model=%s messages=%d tools=%d", baseURL+"/chat/completions", model, len(messages), len(tools))
	resp, err := o.client.Do(req)
	if err != nil {
		log.Printf("LLM stream connect error: %v", err)
		return nil, err
	}
	log.Printf("LLM stream response: status=%d", resp.StatusCode)
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan types.StreamEvent, 32)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		seq := 0

		toolCalls := map[int]*types.ToolCall{}
		var lastUsage *types.LLMUsage

		emitUsage := func() {
			if lastUsage == nil {
				return
			}
			ch <- types.StreamEvent{Type: "usage", Usage: lastUsage, Sequence: seq}
			seq++
			lastUsage = nil
		}

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				log.Printf("LLM stream done after %d events", seq)
				break
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if chunk.Usage != nil {
				lastUsage = convertUsage(chunk.Usage)
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			choice := chunk.Choices[0]
			delta := choice.Delta

			if delta.ReasoningContent != "" {
				ch <- types.StreamEvent{Type: "thinking", Delta: delta.ReasoningContent, Sequence: seq}
				seq++
			}

			if delta.Content != "" {
				ch <- types.StreamEvent{Type: "text", Delta: delta.Content, Sequence: seq}
				seq++
			}

			for _, tc := range delta.ToolCalls {
				existing, ok := toolCalls[tc.Index]
				if !ok {
					existing = &types.ToolCall{ID: tc.ID, Name: tc.Function.Name}
					toolCalls[tc.Index] = existing
				}
				if tc.ID != "" {
					existing.ID = tc.ID
				}
				if tc.Function.Name != "" {
					existing.Name = tc.Function.Name
				}
				existing.Arguments += tc.Function.Arguments
			}

			if choice.FinishReason != nil && *choice.FinishReason == "tool_calls" {
				calls := orderedToolCalls(toolCalls)
				emitUsage()
				ch <- types.StreamEvent{Type: "tool_calls", ToolCalls: calls, Sequence: seq, IsFinal: true}
				return
			}
		}

		if len(toolCalls) > 0 {
			calls := orderedToolCalls(toolCalls)
			emitUsage()
			ch <- types.StreamEvent{Type: "tool_calls", ToolCalls: calls, Sequence: seq, IsFinal: true}
			return
		}

		emitUsage()

		if err := scanner.Err(); err != nil {
			log.Printf("LLM stream read error: %v", err)
			ch <- types.StreamEvent{Type: "error", Delta: err.Error(), IsFinal: true}
		}
	}()

	if thinkingMode != "" {
		return shared.ApplyThinkingStream(ch, thinkingMode), nil
	}

	return ch, nil
}

func buildMessages(messages []protocols.LLMMessage) []reqMessage {
	out := make([]reqMessage, len(messages))
	for i, m := range messages {
		content := m.Content
		msg := reqMessage{Role: m.Role, Content: &content}
		if len(m.ToolCalls) > 0 {
			if content == "" {
				msg.Content = nil
			}
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, reqToolCall{
					ID: tc.ID, Type: "function",
					Function: reqFunction{Name: tc.Name, Arguments: tc.Arguments},
				})
			}
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		out[i] = msg
	}
	return out
}

func buildTools(tools []types.Tool) []reqTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]reqTool, len(tools))
	for i, t := range tools {
		out[i] = reqTool{
			Type: "function",
			Function: reqToolDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return out
}

func convertUsage(u *streamUsage) *types.LLMUsage {
	if u == nil {
		return nil
	}
	out := &types.LLMUsage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
	if u.PromptTokensDetails != nil {
		out.CachedTokens = u.PromptTokensDetails.CachedTokens
	}
	return out
}

func orderedToolCalls(toolCalls map[int]*types.ToolCall) []types.ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}
	indexes := make([]int, 0, len(toolCalls))
	for i := range toolCalls {
		indexes = append(indexes, i)
	}
	sort.Ints(indexes)

	calls := make([]types.ToolCall, 0, len(indexes))
	for _, i := range indexes {
		if tc, ok := toolCalls[i]; ok && tc != nil {
			calls = append(calls, *tc)
		}
	}
	return calls
}
