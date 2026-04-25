package usecases

import (
	"context"
	"strings"

	"mantis/core/base"
	modelplugin "mantis/core/plugins/model"
	"mantis/core/plugins/summarizer"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

type ContextStatus struct {
	SessionID         string `json:"sessionId"`
	ContextWindow     int    `json:"contextWindow"`
	ReserveTokens     int    `json:"reserveTokens"`
	CompactThreshold  int    `json:"compactThreshold"`
	SummaryVersion    int    `json:"summaryVersion"`
	LastContextTokens int    `json:"lastContextTokens"`
	ModelName         string `json:"modelName,omitempty"`
}

type GetContextStatus struct {
	sessionStore protocols.Store[string, types.ChatSession]
	modelStore   protocols.Store[string, types.Model]
	resolver     *modelplugin.Resolver
}

func NewGetContextStatus(
	sessionStore protocols.Store[string, types.ChatSession],
	modelStore protocols.Store[string, types.Model],
	resolver *modelplugin.Resolver,
) *GetContextStatus {
	return &GetContextStatus{sessionStore: sessionStore, modelStore: modelStore, resolver: resolver}
}

func (uc *GetContextStatus) Execute(ctx context.Context, sessionID string) (ContextStatus, error) {
	id := strings.TrimSpace(sessionID)
	if id == "" {
		return ContextStatus{}, base.ErrNotFound
	}

	status := ContextStatus{SessionID: id}

	if !strings.HasPrefix(id, "plan:") {
		sessions, err := uc.sessionStore.Get(ctx, []string{id})
		if err != nil {
			return ContextStatus{}, err
		}
		if s, ok := sessions[id]; ok {
			status.SummaryVersion = s.SummaryVersion
			status.LastContextTokens = s.LastContextTokens
		}
	}

	out, err := uc.resolver.Execute(ctx, modelplugin.Input{DefaultPreset: "chat"})
	if err == nil && out.ModelID != "" {
		if model, err := shared.ResolveModel(ctx, uc.modelStore, out.ModelID); err == nil {
			status.ContextWindow = model.ContextWindow
			status.ReserveTokens = model.ReserveTokens
			status.CompactThreshold = summarizer.EffectiveThreshold(model)
			status.ModelName = model.Name
		}
	}
	if status.CompactThreshold <= 0 {
		status.CompactThreshold = summarizer.DefaultCompactTokens
	}
	return status, nil
}
