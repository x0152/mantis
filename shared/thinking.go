package shared

import (
	"regexp"
	"strings"

	"mantis/core/types"
)

// Some models emit "thinking" blocks as <think>...</think> or <thinking>...</thinking>.
// We also support a couple of common variants to make ThinkingMode more reliable.
var (
	thinkingBlockRe = regexp.MustCompile(`(?is)<\s*(think|thinking|analysis|reasoning)\b[^>]*>.*?<\s*/\s*(think|thinking|analysis|reasoning)\s*>`)
	thinkingOpenRe  = regexp.MustCompile(`(?is)<\s*(think|thinking|analysis|reasoning)\b[^>]*>.*$`)
	thinkingTagsRe  = regexp.MustCompile(`(?is)</?\s*(think|thinking|analysis|reasoning)\b[^>]*>`)
)

func ApplyThinkingMode(content, mode string) string {
	switch mode {
	case "skip":
		// Remove complete blocks first, then be lenient about a missing closing tag.
		out := thinkingBlockRe.ReplaceAllString(content, "")
		out = thinkingOpenRe.ReplaceAllString(out, "")
		return strings.TrimSpace(out)
	case "inline":
		// Keep content but strip the tags.
		out := thinkingTagsRe.ReplaceAllString(content, "")
		return strings.TrimSpace(out)
	default:
		return content
	}
}

func ApplyThinkingStream(src <-chan types.StreamEvent, mode string) <-chan types.StreamEvent {
	out := make(chan types.StreamEvent, 32)
	go func() {
		defer close(out)
		var textParts []string

		flushText := func() {
			if len(textParts) == 0 {
				return
			}
			combined := ApplyThinkingMode(strings.Join(textParts, ""), mode)
			if combined != "" {
				out <- types.StreamEvent{Type: "text", Delta: combined}
			}
			textParts = nil
		}

		for event := range src {
			if event.Type == "text" {
				textParts = append(textParts, event.Delta)
				continue
			}
			if event.Type == "thinking" {
				switch mode {
				case "skip":
					continue
				case "inline":
					flushText()
					out <- types.StreamEvent{Type: "text", Delta: event.Delta, Sequence: event.Sequence}
				default:
					flushText()
					out <- event
				}
				continue
			}
			flushText()
			out <- event
		}

		if len(textParts) > 0 {
			combined := ApplyThinkingMode(strings.Join(textParts, ""), mode)
			if combined != "" {
				out <- types.StreamEvent{Type: "text", Delta: combined, IsFinal: true}
			}
		}
	}()
	return out
}
