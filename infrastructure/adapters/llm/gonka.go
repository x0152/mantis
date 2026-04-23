package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	gonkaopenai "github.com/gonka-ai/gonka-openai/go"
	"github.com/openai/openai-go"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const (
	gonkaDenom    = "ngonka"
	gonkaDecimals = 9
)

type Gonka struct {
	*OpenAI
}

func NewGonka() *Gonka {
	return &Gonka{OpenAI: NewOpenAI()}
}

func (g *Gonka) ListModels(ctx context.Context, baseURL, apiKey string) ([]types.ProviderModel, error) {
	client, err := g.newClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}

	pager := client.Models.ListAutoPaging(ctx)
	items := make([]types.ProviderModel, 0, 64)
	for pager.Next() {
		id := strings.TrimSpace(pager.Current().ID)
		if id == "" {
			continue
		}
		items = append(items, types.ProviderModel{ID: id})
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (g *Gonka) GetInferenceLimit(ctx context.Context, baseURL, apiKey string) (types.InferenceLimit, error) {
	sourceURL := strings.TrimSpace(baseURL)
	if sourceURL == "" {
		return types.InferenceLimit{}, fmt.Errorf("gonka source URL is required")
	}
	privateKey := strings.TrimSpace(apiKey)
	if privateKey == "" {
		return types.InferenceLimit{}, fmt.Errorf("Gonka private key is required")
	}

	address, err := gonkaopenai.GonkaAddress(privateKey)
	if err != nil {
		return types.InferenceLimit{}, fmt.Errorf("derive Gonka address: %w", err)
	}

	amount, err := queryGonkaBalance(ctx, sourceURL, address, gonkaDenom)
	if err != nil {
		return types.InferenceLimit{}, err
	}

	gnk := tokenFloat(amount, gonkaDecimals)
	return types.InferenceLimit{
		Type:  "balance",
		Label: fmt.Sprintf("Balance: %s GNK", formatBalanceFloat(gnk)),
	}, nil
}

func (g *Gonka) ChatStream(ctx context.Context, _ string, baseURL, apiKey string, messages []protocols.LLMMessage, model string, tools []types.Tool, thinkingMode string) (<-chan types.StreamEvent, error) {
	client, err := g.newClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}

	params := openai.ChatCompletionNewParams{
		Model:    model,
		Messages: buildGonkaMessages(messages),
	}
	if reqTools := buildGonkaTools(tools); len(reqTools) > 0 {
		params.Tools = reqTools
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	ch := make(chan types.StreamEvent, 32)

	go func() {
		defer close(ch)
		defer stream.Close()

		seq := 0
		toolCalls := map[int]*types.ToolCall{}

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}

			if reasoning := gonkaReasoningContent(chunk); reasoning != "" {
				ch <- types.StreamEvent{Type: "thinking", Delta: reasoning, Sequence: seq}
				seq++
			}

			choice := chunk.Choices[0]
			if choice.Delta.Content != "" {
				ch <- types.StreamEvent{Type: "text", Delta: choice.Delta.Content, Sequence: seq}
				seq++
			}

			for _, tc := range choice.Delta.ToolCalls {
				idx := int(tc.Index)
				existing, ok := toolCalls[idx]
				if !ok {
					existing = &types.ToolCall{ID: tc.ID, Name: tc.Function.Name}
					toolCalls[idx] = existing
				}
				if tc.ID != "" {
					existing.ID = tc.ID
				}
				if tc.Function.Name != "" {
					existing.Name = tc.Function.Name
				}
				existing.Arguments += tc.Function.Arguments
			}

			if choice.FinishReason == "tool_calls" {
				calls := orderedToolCalls(toolCalls)
				ch <- types.StreamEvent{Type: "tool_calls", ToolCalls: calls, Sequence: seq, IsFinal: true}
				return
			}
		}

		if len(toolCalls) > 0 {
			calls := orderedToolCalls(toolCalls)
			ch <- types.StreamEvent{Type: "tool_calls", ToolCalls: calls, Sequence: seq, IsFinal: true}
			return
		}

		if err := stream.Err(); err != nil {
			ch <- types.StreamEvent{Type: "error", Delta: err.Error(), IsFinal: true}
		}
	}()

	if thinkingMode != "" {
		return shared.ApplyThinkingStream(ch, thinkingMode), nil
	}

	return ch, nil
}

func (g *Gonka) newClient(baseURL, apiKey string) (*gonkaopenai.GonkaOpenAI, error) {
	sourceURL := strings.TrimSpace(baseURL)
	if sourceURL == "" {
		return nil, fmt.Errorf("gonka source URL is required")
	}
	privateKey := strings.TrimSpace(apiKey)
	if privateKey == "" {
		return nil, fmt.Errorf("Gonka private key is required")
	}

	return gonkaopenai.NewGonkaOpenAI(gonkaopenai.Options{
		GonkaPrivateKey: privateKey,
		SourceUrl:       sourceURL,
	})
}

func buildGonkaMessages(messages []protocols.LLMMessage) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case "system":
			out = append(out, openai.SystemMessage(m.Content))
		case "assistant":
			if len(m.ToolCalls) == 0 {
				out = append(out, openai.AssistantMessage(m.Content))
				continue
			}
			assistant := openai.ChatCompletionAssistantMessageParam{
				ToolCalls: buildGonkaAssistantToolCalls(m.ToolCalls),
			}
			if strings.TrimSpace(m.Content) != "" {
				assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(m.Content),
				}
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})
		case "tool":
			if strings.TrimSpace(m.ToolCallID) == "" {
				out = append(out, openai.UserMessage(m.Content))
				continue
			}
			out = append(out, openai.ToolMessage(m.Content, m.ToolCallID))
		default:
			out = append(out, openai.UserMessage(m.Content))
		}
	}
	return out
}

func buildGonkaAssistantToolCalls(toolCalls []types.ToolCall) []openai.ChatCompletionMessageToolCallParam {
	out := make([]openai.ChatCompletionMessageToolCallParam, 0, len(toolCalls))
	for i, tc := range toolCalls {
		name := strings.TrimSpace(tc.Name)
		if name == "" {
			continue
		}
		id := strings.TrimSpace(tc.ID)
		if id == "" {
			id = fmt.Sprintf("call_%d", i)
		}
		out = append(out, openai.ChatCompletionMessageToolCallParam{
			ID: id,
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      name,
				Arguments: tc.Arguments,
			},
		})
	}
	return out
}

func buildGonkaTools(tools []types.Tool) []openai.ChatCompletionToolParam {
	out := make([]openai.ChatCompletionToolParam, 0, len(tools))
	for _, t := range tools {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		fn := openai.FunctionDefinitionParam{Name: name}
		if desc := strings.TrimSpace(t.Description); desc != "" {
			fn.Description = openai.String(desc)
		}
		if len(t.Parameters) > 0 {
			fn.Parameters = t.Parameters
		}
		out = append(out, openai.ChatCompletionToolParam{Function: fn})
	}
	return out
}

func queryGonkaBalance(ctx context.Context, sourceURL, address, denom string) (*big.Int, error) {
	base := strings.TrimSuffix(strings.TrimRight(sourceURL, "/"), "/v1")
	url := base + "/chain-api/cosmos/bank/v1beta1/balances/" + address

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query gonka balance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gonka balance API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode gonka balances: %w", err)
	}

	for _, b := range payload.Balances {
		if b.Denom != denom {
			continue
		}
		v, ok := new(big.Int).SetString(b.Amount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid gonka balance amount: %q", b.Amount)
		}
		return v, nil
	}

	return big.NewInt(0), nil
}

func tokenFloat(amount *big.Int, decimals int) float64 {
	if amount == nil {
		return 0
	}
	denom := new(big.Float).SetFloat64(1)
	for i := 0; i < decimals; i++ {
		denom.Mul(denom, big.NewFloat(10))
	}
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(amount), denom).Float64()
	return f
}

func formatBalanceFloat(v float64) string {
	switch {
	case v >= 1000:
		return fmt.Sprintf("%.0f", v)
	case v >= 1:
		return fmt.Sprintf("%.2f", v)
	case v > 0:
		return fmt.Sprintf("%.4f", v)
	default:
		return "0"
	}
}

func gonkaReasoningContent(chunk openai.ChatCompletionChunk) string {
	raw := strings.TrimSpace(chunk.RawJSON())
	if raw == "" {
		return ""
	}
	var payload streamChunk
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ""
	}
	if len(payload.Choices) == 0 {
		return ""
	}
	return payload.Choices[0].Delta.ReasoningContent
}
