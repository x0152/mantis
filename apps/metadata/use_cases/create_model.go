package usecases

import (
	"context"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateModel struct {
	store protocols.Store[string, types.Model]
}

func NewCreateModel(store protocols.Store[string, types.Model]) *CreateModel {
	return &CreateModel{store: store}
}

const (
	defaultContextWindow = 128000
	defaultReserveTokens = 20000
)

func (uc *CreateModel) Execute(ctx context.Context, m types.Model) (types.Model, error) {
	m.ID = uuid.New().String()
	applyModelDefaults(&m)
	result, err := uc.store.Create(ctx, []types.Model{m})
	if err != nil {
		return types.Model{}, err
	}
	return result[0], nil
}

func applyModelDefaults(m *types.Model) {
	if m.ContextWindow <= 0 {
		m.ContextWindow = defaultContextWindow
	}
	if m.ReserveTokens < 0 {
		m.ReserveTokens = 0
	}
	if m.ReserveTokens == 0 {
		m.ReserveTokens = defaultReserveTokens
	}
	if m.CompactTokens <= 0 {
		budget := m.ContextWindow - m.ReserveTokens
		if budget <= 0 {
			budget = m.ContextWindow
		}
		m.CompactTokens = budget
	}
}
