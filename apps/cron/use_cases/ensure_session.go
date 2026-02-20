package usecases

import (
	"context"

	sessionplugin "mantis/core/plugins/session"
)

type EnsureSession struct {
	policy *sessionplugin.Policy
}

func NewEnsureSession(policy *sessionplugin.Policy) *EnsureSession {
	return &EnsureSession{policy: policy}
}

func (uc *EnsureSession) Execute(ctx context.Context, sessionID string) error {
	_, err := uc.policy.Execute(ctx, sessionplugin.Input{
		Mode:      sessionplugin.ModeEnsure,
		SessionID: sessionID,
	})
	return err
}
