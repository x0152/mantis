package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"

	"mantis/apps/runtime/keys"
	"mantis/apps/runtime/spec"
	"mantis/apps/runtime/templates"
	"mantis/core/protocols"
	"mantis/core/types"
)

const (
	dockerfileHashLabel = "mantis.sandbox.dockerfile_hash"
	pubkeyEnvVar        = "MANTIS_SSH_PUBLIC_KEY"
)

func sandboxHost(name, ip string) string {
	if ip != "" {
		return ip
	}
	return "mantis-sb-" + name
}

func sandboxSSHConfig(name, ip, privateKey string) ([]byte, error) {
	return json.Marshal(map[string]any{
		"host":       sandboxHost(name, ip),
		"port":       22,
		"username":   "mantis",
		"privateKey": privateKey,
	})
}

type Bootstrapper struct {
	rt              protocols.Runtime
	connectionStore protocols.Store[string, types.Connection]
	keyIssuer       *keys.Issuer
	specBuilder     *spec.Builder
}

func NewBootstrapper(
	rt protocols.Runtime,
	connectionStore protocols.Store[string, types.Connection],
	keyIssuer *keys.Issuer,
	specBuilder *spec.Builder,
) *Bootstrapper {
	return &Bootstrapper{
		rt:              rt,
		connectionStore: connectionStore,
		keyIssuer:       keyIssuer,
		specBuilder:     specBuilder,
	}
}

func (b *Bootstrapper) Run(ctx context.Context) error {
	key, err := b.keyIssuer.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("issue sandbox key: %w", err)
	}

	if err := b.seedBuiltins(ctx, key); err != nil {
		log.Printf("runtime bootstrap: seed builtins: %v", err)
	}

	conns, err := b.connectionStore.List(ctx, types.ListQuery{Page: types.Page{Limit: 1000}})
	if err != nil {
		return fmt.Errorf("list connections: %w", err)
	}
	for _, conn := range conns {
		if conn.Dockerfile == "" {
			continue
		}
		sandboxName := strings.TrimPrefix(conn.Name, "sb-")
		if err := b.ensureSandbox(ctx, conn, sandboxName, key); err != nil {
			log.Printf("runtime bootstrap: sandbox %s: %v", sandboxName, err)
		}
	}
	return nil
}

func (b *Bootstrapper) seedBuiltins(ctx context.Context, key types.SandboxKey) error {
	tpls, err := templates.Builtin()
	if err != nil {
		return err
	}
	conns, err := b.connectionStore.List(ctx, types.ListQuery{Page: types.Page{Limit: 1000}})
	if err != nil {
		return err
	}
	byName := make(map[string]types.Connection, len(conns))
	for _, c := range conns {
		byName[c.Name] = c
	}
	for _, t := range tpls {
		config, _ := sandboxSSHConfig(t.Name, "", key.PrivateKey)
		existing, ok := byName[t.Name]
		if !ok {
			conn := types.Connection{
				ID:            uuid.New().String(),
				Type:          "ssh",
				Name:          t.Name,
				Description:   t.Description,
				Config:        config,
				ProfileIDs:    []string{t.ProfileID},
				Dockerfile:    t.Dockerfile,
				Memories:      []types.Memory{},
				MemoryEnabled: true,
			}
			if _, err := b.connectionStore.Create(ctx, []types.Connection{conn}); err != nil {
				log.Printf("runtime bootstrap: create builtin %s: %v", t.Name, err)
			} else {
				log.Printf("runtime bootstrap: seeded builtin %s", t.Name)
			}
			continue
		}
		if existing.Dockerfile == t.Dockerfile {
			continue
		}
		existing.Dockerfile = t.Dockerfile
		existing.Config = config
		if len(existing.ProfileIDs) == 0 {
			existing.ProfileIDs = []string{t.ProfileID}
		}
		if existing.Description == "" {
			existing.Description = t.Description
		}
		if _, err := b.connectionStore.Update(ctx, []types.Connection{existing}); err != nil {
			log.Printf("runtime bootstrap: resync builtin %s: %v", t.Name, err)
		} else {
			log.Printf("runtime bootstrap: resynced builtin %s", t.Name)
		}
	}
	return nil
}

func (b *Bootstrapper) ensureSandbox(ctx context.Context, conn types.Connection, sandboxName string, key types.SandboxKey) error {
	wantHash := dockerfileHash(conn.Dockerfile)
	container, err := b.rt.Inspect(ctx, sandboxName)
	if err == nil && container.Status == "running" && container.Labels[dockerfileHashLabel] == wantHash {
		if err := b.syncConnectionHost(ctx, conn, sandboxName, container.IP, key.PrivateKey); err != nil {
			log.Printf("runtime bootstrap: sync host %s: %v", sandboxName, err)
		}
		return nil
	}

	log.Printf("runtime bootstrap: building %s", sandboxName)
	stream, err := b.rt.Build(ctx, sandboxName, []byte(conn.Dockerfile))
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	if _, err := io.Copy(io.Discard, stream); err != nil {
		stream.Close()
		return fmt.Errorf("build stream: %w", err)
	}
	stream.Close()

	log.Printf("runtime bootstrap: starting %s", sandboxName)
	spec := b.specBuilder.Build(
		ctx,
		sandboxName,
		conn,
		envForSandbox(sandboxName, key.PublicKey),
		map[string]string{dockerfileHashLabel: wantHash},
	)
	started, err := b.rt.Run(ctx, spec)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	if err := b.syncConnectionHost(ctx, conn, sandboxName, started.IP, key.PrivateKey); err != nil {
		log.Printf("runtime bootstrap: sync host %s: %v", sandboxName, err)
	}
	return nil
}

func (b *Bootstrapper) syncConnectionHost(ctx context.Context, conn types.Connection, sandboxName, ip, privateKey string) error {
	cfg, err := sandboxSSHConfig(sandboxName, ip, privateKey)
	if err != nil {
		return err
	}
	if string(conn.Config) == string(cfg) {
		return nil
	}
	conn.Config = cfg
	_, err = b.connectionStore.Update(ctx, []types.Connection{conn})
	return err
}

func dockerfileHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func envForSandbox(name, publicKey string) map[string]string {
	env := map[string]string{pubkeyEnvVar: publicKey}
	if name == "runtimectl" {
		env["MANTIS_URL"] = "http://app:8080"
		env["MANTIS_RUNTIME_TOKEN"] = os.Getenv("RUNTIME_API_TOKEN")
	}
	return env
}
