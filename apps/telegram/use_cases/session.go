package usecases

import (
	"context"
	"fmt"
	"sync"

	sessionplugin "mantis/core/plugins/session"
)

type SessionMode string

const (
	SessionModeGetOrCreate SessionMode = "get_or_create"
	SessionModeReset       SessionMode = "reset"
)

type Session struct {
	policy *sessionplugin.Policy
	mu     sync.Mutex
}

func NewSession(policy *sessionplugin.Policy) *Session {
	return &Session{policy: policy}
}

func (uc *Session) Execute(ctx context.Context, mode SessionMode) (string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	switch mode {
	case SessionModeGetOrCreate:
		out, err := uc.policy.Execute(ctx, sessionplugin.Input{
			Mode:            sessionplugin.ModeLatestOrCreate,
			ExcludePrefixes: []string{"cron:"},
		})
		if err != nil {
			fallback, ferr := uc.policy.Execute(ctx, sessionplugin.Input{Mode: sessionplugin.ModeCreateNew})
			if ferr != nil {
				return "", ferr
			}
			return fallback.Session.ID, nil
		}
		return out.Session.ID, nil
	case SessionModeReset:
		out, err := uc.policy.Execute(ctx, sessionplugin.Input{Mode: sessionplugin.ModeCreateNew})
		if err != nil {
			return "", err
		}
		return out.Session.ID, nil
	default:
		return "", fmt.Errorf("unknown session mode: %q", mode)
	}
}
