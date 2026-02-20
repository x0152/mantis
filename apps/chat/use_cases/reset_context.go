package usecases

import (
	"context"

	sessionplugin "mantis/core/plugins/session"
	"mantis/core/protocols"
	"mantis/core/types"
)

type ResetContext struct {
	policy *sessionplugin.Policy
}

func NewResetContext(store protocols.Store[string, types.ChatSession]) *ResetContext {
	return &ResetContext{policy: sessionplugin.NewPolicy(store)}
}

func (uc *ResetContext) Execute(ctx context.Context) (types.ChatSession, error) {
	out, err := uc.policy.Execute(ctx, sessionplugin.Input{Mode: sessionplugin.ModeCreateNew})
	if err != nil {
		return types.ChatSession{}, err
	}
	return out.Session, nil
}
