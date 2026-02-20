package usecases

import (
	"context"

	sessionplugin "mantis/core/plugins/session"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetCurrentSession struct {
	policy *sessionplugin.Policy
}

func NewGetCurrentSession(store protocols.Store[string, types.ChatSession]) *GetCurrentSession {
	return &GetCurrentSession{policy: sessionplugin.NewPolicy(store)}
}

func (uc *GetCurrentSession) Execute(ctx context.Context) (types.ChatSession, error) {
	out, err := uc.policy.Execute(ctx, sessionplugin.Input{
		Mode:            sessionplugin.ModeLatestOrCreate,
		ExcludePrefixes: []string{"cron:"},
	})
	if err != nil {
		return types.ChatSession{}, err
	}
	return out.Session, nil
}
