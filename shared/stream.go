package shared

import (
	"errors"
	"strings"

	"mantis/core/types"
)

func CollectText(ch <-chan types.StreamEvent) (string, error) {
	var sb strings.Builder
	for event := range ch {
		switch event.Type {
		case "error":
			return "", errors.New(event.Delta)
		case "text":
			sb.WriteString(event.Delta)
		}
	}
	return sb.String(), nil
}
