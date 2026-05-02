package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	runtimetemplates "mantis/apps/runtime/templates"
	dockerruntime "mantis/infrastructure/adapters/runtime/docker"
)

const dockerfileHashLabel = "mantis.sandbox.dockerfile_hash"

func main() {
	log.SetFlags(0)
	verbose := flag.Bool("verbose", false, "stream Docker build output to stderr")
	timeout := flag.Duration("timeout", 30*time.Minute, "overall timeout for prebuild")
	socketWait := flag.Duration("socket-wait", 30*time.Second, "how long to wait for the docker socket to appear")
	flag.Parse()

	socket := envOr("DOCKER_SOCKET", "/var/run/docker.sock")
	if err := waitForSocket(socket, *socketWait); err != nil {
		log.Fatalf("sandbox-prebuild: %v", err)
	}

	rt := dockerruntime.New(dockerruntime.Options{
		SocketPath: socket,
		Network:    envOr("RUNTIME_NETWORK", ""),
	})

	tpls, err := runtimetemplates.Builtin()
	if err != nil {
		log.Fatalf("sandbox-prebuild: render templates: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	failures := 0
	for _, t := range tpls {
		hash := dockerfileHash(t.Dockerfile)
		if labels, err := rt.ImageLabels(ctx, t.Name); err == nil && labels != nil && labels[dockerfileHashLabel] == hash {
			log.Printf("sandbox-prebuild: %-12s up-to-date (sha=%s)", t.Name, hash)
			continue
		}
		log.Printf("sandbox-prebuild: %-12s building (sha=%s)", t.Name, hash)
		started := time.Now()
		if err := build(ctx, rt, t.Name, t.Dockerfile, hash, *verbose); err != nil {
			failures++
			log.Printf("sandbox-prebuild: %-12s FAILED after %s: %v", t.Name, time.Since(started).Round(time.Second), err)
			continue
		}
		log.Printf("sandbox-prebuild: %-12s ready (%s)", t.Name, time.Since(started).Round(time.Second))
	}

	if failures > 0 {
		log.Fatalf("sandbox-prebuild: %d/%d sandbox images failed to build", failures, len(tpls))
	}
	log.Printf("sandbox-prebuild: all %d sandbox images ready", len(tpls))
}

func build(ctx context.Context, rt *dockerruntime.Runtime, name, dockerfile, hash string, verbose bool) error {
	stream, err := rt.BuildWithLabels(ctx, name, []byte(dockerfile), map[string]string{dockerfileHashLabel: hash})
	if err != nil {
		return err
	}
	defer stream.Close()
	sink := io.Discard
	if verbose {
		sink = os.Stderr
	}
	if _, err := io.Copy(sink, stream); err != nil {
		return err
	}
	return nil
}

func dockerfileHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func waitForSocket(path string, deadline time.Duration) error {
	until := time.Now().Add(deadline)
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		if time.Now().After(until) {
			return fmt.Errorf("docker socket %s not available after %s", path, deadline)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
