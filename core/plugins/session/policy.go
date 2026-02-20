package session

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Mode string

const (
	ModeLatestOrCreate Mode = "latest_or_create"
	ModeCreateNew      Mode = "create_new"
	ModeEnsure         Mode = "ensure"
)

type Input struct {
	Mode Mode

	// For ModeEnsure only.
	SessionID string

	// For ModeLatestOrCreate only.
	ExcludePrefixes []string
}

type Output struct {
	Session types.ChatSession
	Created bool
}

type Policy struct {
	store protocols.Store[string, types.ChatSession]

	nowFn   func() time.Time
	newIDFn func() string
}

func NewPolicy(store protocols.Store[string, types.ChatSession]) *Policy {
	return &Policy{
		store:   store,
		nowFn:   time.Now,
		newIDFn: func() string { return uuid.New().String() },
	}
}

func (p *Policy) Execute(ctx context.Context, in Input) (Output, error) {
	switch in.Mode {
	case ModeLatestOrCreate:
		return p.latestOrCreate(ctx, in.ExcludePrefixes)
	case ModeCreateNew:
		s, err := p.create(ctx, "")
		return Output{Session: s, Created: err == nil}, err
	case ModeEnsure:
		if strings.TrimSpace(in.SessionID) == "" {
			return Output{}, fmt.Errorf("session_id is required for mode %q", ModeEnsure)
		}
		return p.ensure(ctx, in.SessionID)
	default:
		return Output{}, fmt.Errorf("unknown session mode: %q", in.Mode)
	}
}

func (p *Policy) latestOrCreate(ctx context.Context, excludePrefixes []string) (Output, error) {
	items, err := p.store.List(ctx, types.ListQuery{})
	if err != nil {
		return Output{}, err
	}

	var latest types.ChatSession
	for _, s := range items {
		if hasAnyPrefix(s.ID, excludePrefixes) {
			continue
		}
		if latest.ID == "" || s.CreatedAt.After(latest.CreatedAt) {
			latest = s
		}
	}
	if latest.ID != "" {
		return Output{Session: latest}, nil
	}

	s, err := p.create(ctx, "")
	return Output{Session: s, Created: err == nil}, err
}

func (p *Policy) ensure(ctx context.Context, sessionID string) (Output, error) {
	existing, err := p.store.Get(ctx, []string{sessionID})
	if err != nil {
		return Output{}, err
	}
	if s, ok := existing[sessionID]; ok {
		return Output{Session: s}, nil
	}

	s, err := p.create(ctx, sessionID)
	return Output{Session: s, Created: err == nil}, err
}

func (p *Policy) create(ctx context.Context, id string) (types.ChatSession, error) {
	if strings.TrimSpace(id) == "" {
		id = p.newIDFn()
	}
	now := p.nowFn()
	created, err := p.store.Create(ctx, []types.ChatSession{{ID: id, CreatedAt: now}})
	if err != nil {
		return types.ChatSession{}, err
	}
	if len(created) == 0 {
		return types.ChatSession{}, fmt.Errorf("session was not created")
	}
	return created[0], nil
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
