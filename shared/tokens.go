package shared

import (
	"mantis/core/protocols"
)

const tokenCharRatio = 4

func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + tokenCharRatio - 1) / tokenCharRatio
}

func EstimateMessagesTokens(messages []protocols.LLMMessage) int {
	total := 0
	for _, m := range messages {
		total += EstimateTokens(m.Role)
		total += EstimateTokens(m.Content)
		total += 4
		for _, tc := range m.ToolCalls {
			total += EstimateTokens(tc.Name) + EstimateTokens(tc.Arguments) + 4
		}
		if m.ToolCallID != "" {
			total += EstimateTokens(m.ToolCallID) + 2
		}
	}
	return total
}
