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

	"mantis/apps/runtime/templates"
	"mantis/core/protocols"
	"mantis/core/types"
)

const dockerfileHashLabel = "mantis.sandbox.dockerfile_hash"

type Bootstrapper struct {
	rt              protocols.Runtime
	connectionStore protocols.Store[string, types.Connection]
}

func NewBootstrapper(rt protocols.Runtime, connectionStore protocols.Store[string, types.Connection]) *Bootstrapper {
	return &Bootstrapper{rt: rt, connectionStore: connectionStore}
}

func (b *Bootstrapper) Run(ctx context.Context) error {
	if err := b.seedBuiltins(ctx); err != nil {
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
		if err := b.ensureSandbox(ctx, conn, sandboxName); err != nil {
			log.Printf("runtime bootstrap: sandbox %s: %v", sandboxName, err)
		}
	}
	return nil
}

func (b *Bootstrapper) seedBuiltins(ctx context.Context) error {
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
		config, _ := json.Marshal(map[string]any{
			"host":     "mantis-sb-" + t.Name,
			"port":     22,
			"username": "mantis",
			"password": "mantis",
		})
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

func (b *Bootstrapper) ensureSandbox(ctx context.Context, conn types.Connection, sandboxName string) error {
	wantHash := dockerfileHash(conn.Dockerfile)
	container, err := b.rt.Inspect(ctx, sandboxName)
	if err == nil && container.Status == "running" && container.Labels[dockerfileHashLabel] == wantHash {
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
	spec := types.RuntimeRunSpec{
		Name:   sandboxName,
		Env:    envForSandbox(sandboxName),
		Labels: map[string]string{dockerfileHashLabel: wantHash},
	}
	if _, err := b.rt.Run(ctx, spec); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func dockerfileHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func envForSandbox(name string) map[string]string {
	if name != "runtimectl" {
		return nil
	}
	return map[string]string{
		"MANTIS_URL":           "http://app:8080",
		"MANTIS_RUNTIME_TOKEN": os.Getenv("RUNTIME_API_TOKEN"),
	}
}
