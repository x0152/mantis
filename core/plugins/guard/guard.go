package guard

import (
	"context"
	"fmt"
	"regexp"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Violation struct {
	Rule    string
	Message string
}

type Rule interface {
	Execute(command string) *Violation
}

type RegexRule struct {
	name    string
	pattern *regexp.Regexp
	message string
}

func NewRegexRule(name, pattern, message string) *RegexRule {
	return &RegexRule{name: name, pattern: regexp.MustCompile(pattern), message: message}
}

func (r *RegexRule) Execute(command string) *Violation {
	if r.pattern.MatchString(command) {
		return &Violation{Rule: r.name, Message: r.message}
	}
	return nil
}

type Guard struct {
	store    protocols.Store[string, types.GuardRule]
	builtins []Rule
}

func New(store protocols.Store[string, types.GuardRule]) *Guard {
	return &Guard{store: store, builtins: defaultBuiltins()}
}

func (g *Guard) Execute(ctx context.Context, connectionID, command string) *Violation {
	for _, r := range g.builtins {
		if v := r.Execute(command); v != nil {
			return v
		}
	}

	for _, r := range g.loadRules(ctx, connectionID) {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			continue
		}
		if re.MatchString(command) {
			msg := r.Description
			if msg == "" {
				msg = fmt.Sprintf("matched rule: %s", r.Name)
			}
			return &Violation{Rule: r.Name, Message: msg}
		}
	}

	return nil
}

func (g *Guard) Rules(ctx context.Context, connectionID string) []types.GuardRule {
	return g.loadRules(ctx, connectionID)
}

func (g *Guard) loadRules(ctx context.Context, connectionID string) []types.GuardRule {
	all, err := g.store.List(ctx, types.ListQuery{})
	if err != nil {
		return nil
	}
	var out []types.GuardRule
	for _, r := range all {
		if !r.Enabled {
			continue
		}
		if r.ConnectionID != "" && r.ConnectionID != connectionID {
			continue
		}
		out = append(out, r)
	}
	return out
}

func defaultBuiltins() []Rule {
	defs := []struct{ name, pattern, msg string }{
		{"recursive delete root", `(?:^|\s)rm\s+(?:-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*|--recursive(?:\s+--force)?|--force(?:\s+--recursive)?)\s+(?:--no-preserve-root\s+)?["']?/["']?(?:\s|$)`, "recursive force delete on system root"},
		{"write to block device", `>\s*/dev/sd`, "direct write to block device"},
		{"format filesystem", `mkfs\.`, "filesystem formatting"},
		{"fork bomb", `:\(\)\s*\{`, "fork bomb detected"},
		{"disk overwrite", `dd\s+if=.+of=/dev/`, "destructive disk operation"},
	}
	var out []Rule
	for _, d := range defs {
		out = append(out, NewRegexRule(d.name, d.pattern, d.msg))
	}
	return out
}
