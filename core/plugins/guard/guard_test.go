package guard

import (
	"context"
	"testing"

	"mantis/core/types"
)

type noopGuardRuleStore struct{}

func (noopGuardRuleStore) Create(_ context.Context, _ []types.GuardRule) ([]types.GuardRule, error) {
	return nil, nil
}

func (noopGuardRuleStore) Get(_ context.Context, _ []string) (map[string]types.GuardRule, error) {
	return map[string]types.GuardRule{}, nil
}

func (noopGuardRuleStore) List(_ context.Context, _ types.ListQuery) ([]types.GuardRule, error) {
	return nil, nil
}

func (noopGuardRuleStore) Update(_ context.Context, _ []types.GuardRule) ([]types.GuardRule, error) {
	return nil, nil
}

func (noopGuardRuleStore) Delete(_ context.Context, _ []string) error {
	return nil
}

func TestGuard_BlocksRecursiveDeleteOnRoot(t *testing.T) {
	g := New(noopGuardRuleStore{})
	ctx := context.Background()

	cases := []string{
		"rm -rf /",
		"sudo rm -fr /",
		"rm --force --recursive /",
		`rm -rf "/"`,
		"rm -rf / --no-preserve-root",
	}

	for _, command := range cases {
		v := g.Execute(ctx, "", command)
		if v == nil {
			t.Fatalf("expected violation for %q, got nil", command)
		}
		if v.Rule != "recursive delete root" {
			t.Fatalf("expected recursive delete root rule for %q, got %q", command, v.Rule)
		}
	}
}

func TestGuard_AllowsNonRootDeletes(t *testing.T) {
	g := New(noopGuardRuleStore{})
	ctx := context.Background()

	cases := []string{
		"rm -rf /tmp",
		"rm -fr /var/log/app",
		"rm -f ./file.txt",
	}

	for _, command := range cases {
		if v := g.Execute(ctx, "", command); v != nil {
			t.Fatalf("did not expect violation for %q, got %+v", command, *v)
		}
	}
}
