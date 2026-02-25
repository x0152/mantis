package guard

import (
	"context"
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"mantis/core/protocols"
	"mantis/core/types"
)

type Violation struct {
	Rule    string
	Message string
}

type Guard struct {
	store protocols.Store[string, types.GuardProfile]
}

func New(store protocols.Store[string, types.GuardProfile]) *Guard {
	return &Guard{store: store}
}

func (g *Guard) Execute(ctx context.Context, profileIDs []string, command string) *Violation {
	if len(profileIDs) == 0 {
		return nil
	}

	profiles, err := g.store.Get(ctx, profileIDs)
	if err != nil || len(profiles) == 0 {
		return nil
	}

	merged := mergeProfiles(profiles)

	if merged.Capabilities.Unrestricted {
		return nil
	}

	return checkCommand(merged, command, 0)
}

func (g *Guard) Profiles(ctx context.Context, profileIDs []string) []types.GuardProfile {
	if len(profileIDs) == 0 {
		return nil
	}
	m, err := g.store.Get(ctx, profileIDs)
	if err != nil {
		return nil
	}
	out := make([]types.GuardProfile, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	return out
}

func (g *Guard) Describe(ctx context.Context, profileIDs []string) string {
	profiles := g.Profiles(ctx, profileIDs)
	if len(profiles) == 0 {
		return ""
	}
	merged := mergeProfiles(profileIDsToMap(profiles))
	if merged.Capabilities.Unrestricted {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Guard: only allowed commands will execute. If blocked — inform the user.\n")
	sb.WriteString("Capabilities: ")
	var caps []string
	if merged.Capabilities.Pipes {
		caps = append(caps, "pipes")
	}
	if merged.Capabilities.Redirects {
		caps = append(caps, "redirects")
	}
	if merged.Capabilities.CmdSubst {
		caps = append(caps, "command substitution")
	}
	if merged.Capabilities.Background {
		caps = append(caps, "background processes")
	}
	if merged.Capabilities.Sudo {
		caps = append(caps, "sudo")
	}
	if merged.Capabilities.CodeExec {
		caps = append(caps, "code execution")
	}
	if merged.Capabilities.Download {
		caps = append(caps, "downloads")
	}
	if merged.Capabilities.Install {
		caps = append(caps, "package install")
	}
	if merged.Capabilities.WriteFS {
		caps = append(caps, "filesystem writes")
	}
	if merged.Capabilities.NetworkOut {
		caps = append(caps, "outbound network")
	}
	if merged.Capabilities.Cron {
		caps = append(caps, "cron/at")
	}
	if len(caps) == 0 {
		caps = append(caps, "none")
	}
	sb.WriteString(strings.Join(caps, ", "))

	if len(merged.commands) > 0 {
		sb.WriteString("\nAllowed commands: ")
		names := make([]string, 0, len(merged.commands))
		for name := range merged.commands {
			names = append(names, name)
		}
		sb.WriteString(strings.Join(names, ", "))
	}
	return sb.String()
}

func profileIDsToMap(profiles []types.GuardProfile) map[string]types.GuardProfile {
	m := make(map[string]types.GuardProfile, len(profiles))
	for _, p := range profiles {
		m[p.ID] = p
	}
	return m
}

type mergedProfile struct {
	Capabilities types.GuardCapabilities
	commands     map[string]mergedCommand
}

type mergedCommand struct {
	allowAll   bool
	allowedArgs map[string]bool
	allowedSQL  map[string]bool
}

func mergeProfiles(profiles map[string]types.GuardProfile) mergedProfile {
	m := mergedProfile{commands: make(map[string]mergedCommand)}
	for _, p := range profiles {
		m.Capabilities.Pipes = m.Capabilities.Pipes || p.Capabilities.Pipes
		m.Capabilities.Redirects = m.Capabilities.Redirects || p.Capabilities.Redirects
		m.Capabilities.CmdSubst = m.Capabilities.CmdSubst || p.Capabilities.CmdSubst
		m.Capabilities.Background = m.Capabilities.Background || p.Capabilities.Background
		m.Capabilities.Sudo = m.Capabilities.Sudo || p.Capabilities.Sudo
		m.Capabilities.CodeExec = m.Capabilities.CodeExec || p.Capabilities.CodeExec
		m.Capabilities.Download = m.Capabilities.Download || p.Capabilities.Download
		m.Capabilities.Install = m.Capabilities.Install || p.Capabilities.Install
		m.Capabilities.WriteFS = m.Capabilities.WriteFS || p.Capabilities.WriteFS
		m.Capabilities.NetworkOut = m.Capabilities.NetworkOut || p.Capabilities.NetworkOut
		m.Capabilities.Cron = m.Capabilities.Cron || p.Capabilities.Cron
		m.Capabilities.Unrestricted = m.Capabilities.Unrestricted || p.Capabilities.Unrestricted

		for _, cmd := range p.Commands {
			existing, ok := m.commands[cmd.Command]
			if !ok {
				existing = mergedCommand{allowedArgs: make(map[string]bool), allowedSQL: make(map[string]bool)}
			}
			if len(cmd.AllowedArgs) == 0 && len(cmd.AllowedSQL) == 0 {
				existing.allowAll = true
			}
			for _, a := range cmd.AllowedArgs {
				existing.allowedArgs[a] = true
			}
			for _, s := range cmd.AllowedSQL {
				existing.allowedSQL[strings.ToUpper(s)] = true
			}
			m.commands[cmd.Command] = existing
		}
	}
	return m
}

const maxRecursionDepth = 3

func checkCommand(mp mergedProfile, command string, depth int) *Violation {
	if depth > maxRecursionDepth {
		return &Violation{Rule: "recursion-limit", Message: "too many nesting levels — run each step separately"}
	}

	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		return &Violation{Rule: "parse-error", Message: fmt.Sprintf("syntax error: %s", err)}
	}

	var violation *Violation
	syntax.Walk(prog, func(node syntax.Node) bool {
		if violation != nil {
			return false
		}
		switch n := node.(type) {
		case *syntax.BinaryCmd:
			if n.Op == syntax.Pipe || n.Op == syntax.PipeAll {
				if !mp.Capabilities.Pipes {
					violation = &Violation{Rule: "pipes-disabled", Message: "pipe (|) is not allowed — run commands separately"}
					return false
				}
				if hasPipeToShell(n) && !mp.Capabilities.CodeExec {
					violation = &Violation{Rule: "pipe-to-shell", Message: "piping to shell (| sh/bash) is not allowed — run commands directly"}
					return false
				}
			}
		case *syntax.Redirect:
			if !mp.Capabilities.Redirects {
				violation = &Violation{Rule: "redirects-disabled", Message: "redirect (>, <) is not allowed — use stdout only"}
				return false
			}
		case *syntax.CmdSubst:
			if !mp.Capabilities.CmdSubst {
				violation = &Violation{Rule: "cmd-subst-disabled", Message: "$() substitution is not allowed — run the inner command separately"}
				return false
			}
		case *syntax.Stmt:
			if n.Background {
				if !mp.Capabilities.Background {
					violation = &Violation{Rule: "background-disabled", Message: "background (&) is not allowed — run in foreground"}
					return false
				}
			}
		case *syntax.CallExpr:
			v := checkCallExpr(mp, n, command, depth)
			if v != nil {
				violation = v
				return false
			}
		}
		return true
	})
	return violation
}

var shellInterpreters = map[string]string{
	"bash": "-c", "sh": "-c", "zsh": "-c", "dash": "-c",
	"python": "-c", "python3": "-c", "perl": "-e", "ruby": "-e", "node": "-e",
}

var sqlInterpreters = map[string]string{
	"psql": "-c", "mysql": "-e", "sqlite3": "",
}

var downloadCommands = map[string]bool{
	"curl": true, "wget": true, "scp": true, "rsync": true,
}

var installCommands = map[string]bool{
	"apt": true, "apt-get": true, "yum": true, "dnf": true,
	"pip": true, "pip3": true, "npm": true, "yarn": true,
	"brew": true, "snap": true, "gem": true, "cargo": true,
	"pacman": true, "zypper": true, "apk": true,
}

var writeCommands = map[string]bool{
	"cp": true, "mv": true, "mkdir": true, "touch": true,
	"tee": true, "chmod": true, "chown": true, "chgrp": true,
	"ln": true, "install": true, "rsync": true,
}

var cronCommands = map[string]bool{
	"crontab": true, "at": true,
}

func allowedCommandNames(mp mergedProfile) string {
	names := make([]string, 0, len(mp.commands))
	for name := range mp.commands {
		names = append(names, name)
	}
	if len(names) > 20 {
		return strings.Join(names[:20], ", ") + fmt.Sprintf(" (and %d more)", len(names)-20)
	}
	return strings.Join(names, ", ")
}

func allowedArgsList(mc mergedCommand) string {
	args := make([]string, 0, len(mc.allowedArgs))
	for a := range mc.allowedArgs {
		args = append(args, a)
	}
	return strings.Join(args, ", ")
}

func checkCallExpr(mp mergedProfile, call *syntax.CallExpr, originalCmd string, depth int) *Violation {
	args := resolveArgs(call)
	if len(args) == 0 {
		return nil
	}

	cmdName := args[0]

	if cmdName == "sudo" {
		if !mp.Capabilities.Sudo {
			return &Violation{Rule: "sudo-disabled", Message: "sudo is not allowed — run without sudo"}
		}
		args = args[1:]
		if len(args) == 0 {
			return nil
		}
		cmdName = args[0]
	}

	if cmdName == "nohup" {
		if !mp.Capabilities.Background {
			return &Violation{Rule: "background-disabled", Message: "nohup is not allowed — run in foreground"}
		}
		args = args[1:]
		if len(args) == 0 {
			return nil
		}
		cmdName = args[0]
	}

	if flag, ok := shellInterpreters[cmdName]; ok {
		code := extractFlag(args[1:], flag)
		if code != "" {
			if !mp.Capabilities.CodeExec {
				return &Violation{Rule: "code-exec-disabled", Message: fmt.Sprintf("%s -c is not allowed — run commands directly", cmdName)}
			}
			if cmdName == "bash" || cmdName == "sh" || cmdName == "zsh" || cmdName == "dash" {
				return checkCommand(mp, code, depth+1)
			}
		}
	}

	if flag, ok := sqlInterpreters[cmdName]; ok {
		query := extractFlag(args[1:], flag)
		if query != "" {
			return checkSQL(mp, cmdName, query)
		}
	}

	if downloadCommands[cmdName] && !mp.Capabilities.Download {
		return &Violation{Rule: "download-disabled", Message: fmt.Sprintf("\"%s\" blocked — downloads not allowed", cmdName)}
	}
	if installCommands[cmdName] && !mp.Capabilities.Install {
		return &Violation{Rule: "install-disabled", Message: fmt.Sprintf("\"%s\" blocked — package install not allowed", cmdName)}
	}
	if writeCommands[cmdName] && !mp.Capabilities.WriteFS {
		return &Violation{Rule: "write-fs-disabled", Message: fmt.Sprintf("\"%s\" blocked — filesystem writes not allowed (read-only)", cmdName)}
	}
	if cronCommands[cmdName] && !mp.Capabilities.Cron {
		return &Violation{Rule: "cron-disabled", Message: fmt.Sprintf("\"%s\" blocked — scheduling not allowed", cmdName)}
	}

	mc, ok := mp.commands[cmdName]
	if !ok {
		return &Violation{Rule: "command-not-allowed", Message: fmt.Sprintf("\"%s\" is not allowed. Allowed: %s", cmdName, allowedCommandNames(mp))}
	}

	if mc.allowAll {
		return nil
	}

	if len(mc.allowedArgs) > 0 && len(args) > 1 {
		firstArg := args[1]
		if !mc.allowedArgs[firstArg] {
			return &Violation{
				Rule:    "arg-not-allowed",
				Message: fmt.Sprintf("\"%s %s\" not allowed. Allowed for %s: %s", cmdName, firstArg, cmdName, allowedArgsList(mc)),
			}
		}
	}

	return nil
}

func allowedSQLList(mc mergedCommand) string {
	ops := make([]string, 0, len(mc.allowedSQL))
	for s := range mc.allowedSQL {
		ops = append(ops, s)
	}
	return strings.Join(ops, ", ")
}

func checkSQL(mp mergedProfile, cmdName, query string) *Violation {
	mc, ok := mp.commands[cmdName]
	if !ok {
		return &Violation{Rule: "command-not-allowed", Message: fmt.Sprintf("\"%s\" is not allowed. Allowed: %s", cmdName, allowedCommandNames(mp))}
	}
	if mc.allowAll || len(mc.allowedSQL) == 0 {
		return nil
	}

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil
	}

	if trimmed[0] == '\\' {
		token := strings.Fields(trimmed)[0]
		if mc.allowedSQL[strings.ToUpper(token)] {
			return nil
		}
		return &Violation{
			Rule:    "sql-not-allowed",
			Message: fmt.Sprintf("SQL \"%s\" not allowed via %s. Allowed: %s", token, cmdName, allowedSQLList(mc)),
		}
	}

	keyword := strings.ToUpper(strings.Fields(trimmed)[0])
	if mc.allowedSQL[keyword] {
		return nil
	}
	return &Violation{
		Rule:    "sql-not-allowed",
		Message: fmt.Sprintf("SQL %s not allowed via %s. Allowed: %s", keyword, cmdName, allowedSQLList(mc)),
	}
}

func resolveArgs(call *syntax.CallExpr) []string {
	var args []string
	for _, word := range call.Args {
		var sb strings.Builder
		for _, part := range word.Parts {
			switch p := part.(type) {
			case *syntax.Lit:
				sb.WriteString(p.Value)
			case *syntax.SglQuoted:
				sb.WriteString(p.Value)
			case *syntax.DblQuoted:
				for _, qp := range p.Parts {
					if lit, ok := qp.(*syntax.Lit); ok {
						sb.WriteString(lit.Value)
					}
				}
			}
		}
		if s := sb.String(); s != "" {
			args = append(args, s)
		}
	}
	return args
}

func extractFlag(args []string, flag string) string {
	if flag == "" && len(args) > 0 {
		return args[0]
	}
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, flag+"=") {
			return a[len(flag)+1:]
		}
		if strings.HasPrefix(a, flag) && len(a) > len(flag) {
			return a[len(flag):]
		}
	}
	return ""
}

func hasPipeToShell(binary *syntax.BinaryCmd) bool {
	if binary.Op != syntax.Pipe && binary.Op != syntax.PipeAll {
		return false
	}
	stmt := binary.Y
	if stmt == nil || stmt.Cmd == nil {
		return false
	}
	if call, ok := stmt.Cmd.(*syntax.CallExpr); ok {
		args := resolveArgs(call)
		if len(args) > 0 {
			_, isShell := shellInterpreters[args[0]]
			return isShell
		}
	}
	return false
}
