package shared

import (
	"mantis/core/plugins/tokenizer"
	"mantis/core/protocols"
)

func EstimateTokens(s string) int {
	return tokenizer.Default().Count(s)
}

func EstimateMessagesTokens(messages []protocols.LLMMessage) int {
	return tokenizer.Default().CountMessages(messages)
}

func EstimateTokensFor(modelName, s string) int {
	return tokenizer.For(modelName).Count(s)
}

func EstimateMessagesTokensFor(modelName string, messages []protocols.LLMMessage) int {
	return tokenizer.For(modelName).CountMessages(messages)
}
