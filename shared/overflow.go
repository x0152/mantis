package shared

import (
	"errors"
	"regexp"
	"strings"
)

var overflowPattern = regexp.MustCompile(`(?i)(context[\s_-]?length|context[\s_-]?window|maximum context|too many tokens|token[s]?[\s_-]?limit|prompt is too long|input.*(?:too long|exceed)|reduce.*(prompt|input|length)|request.*too large|max[_\s-]?tokens[_\s-]?exceeded)`)

func IsContextOverflow(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "context canceled") || strings.Contains(msg, "context deadline exceeded") {
		return false
	}
	return overflowPattern.MatchString(msg)
}

var errContextOverflow = errors.New("context overflow")

func NewContextOverflowError(detail string) error {
	if strings.TrimSpace(detail) == "" {
		return errContextOverflow
	}
	return errors.New("context overflow: " + detail)
}
