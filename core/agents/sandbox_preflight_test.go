package agents

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"mantis/core/types"
)

type fakeRuntime struct {
	status string
	err    error
}

func (f *fakeRuntime) Build(ctx context.Context, name string, dockerfile []byte) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeRuntime) BuildWithLabels(ctx context.Context, name string, dockerfile []byte, labels map[string]string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeRuntime) Run(ctx context.Context, spec types.RuntimeRunSpec) (types.RuntimeContainer, error) {
	return types.RuntimeContainer{}, errors.New("not implemented")
}
func (f *fakeRuntime) Stop(ctx context.Context, name string) error   { return nil }
func (f *fakeRuntime) Remove(ctx context.Context, name string) error { return nil }
func (f *fakeRuntime) List(ctx context.Context) ([]types.RuntimeContainer, error) {
	return nil, nil
}
func (f *fakeRuntime) Inspect(ctx context.Context, name string) (types.RuntimeContainer, error) {
	if f.err != nil {
		return types.RuntimeContainer{}, f.err
	}
	return types.RuntimeContainer{Name: name, Status: f.status}, nil
}
func (f *fakeRuntime) ImageLabels(ctx context.Context, name string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeRuntime) Logs(ctx context.Context, name string, tail int, follow bool) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func TestCheckSandboxRunning_SkipsRemote(t *testing.T) {
	a := &MantisAgent{runtime: &fakeRuntime{status: "exited"}}
	if err := a.checkSandboxRunning(context.Background(), types.Connection{Name: "prod-vps"}); err != nil {
		t.Fatalf("remote connection should be skipped, got: %v", err)
	}
}

func TestCheckSandboxRunning_SkipsNilRuntime(t *testing.T) {
	a := &MantisAgent{}
	conn := types.Connection{Name: "python", Dockerfile: "FROM ubuntu:24.04"}
	if err := a.checkSandboxRunning(context.Background(), conn); err != nil {
		t.Fatalf("nil runtime should be skipped, got: %v", err)
	}
}

func TestCheckSandboxRunning_RunningOK(t *testing.T) {
	a := &MantisAgent{runtime: &fakeRuntime{status: "running"}}
	conn := types.Connection{Name: "python", Dockerfile: "FROM ubuntu:24.04"}
	if err := a.checkSandboxRunning(context.Background(), conn); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheckSandboxRunning_Stopped(t *testing.T) {
	a := &MantisAgent{runtime: &fakeRuntime{status: "exited"}}
	conn := types.Connection{Name: "python", Dockerfile: "FROM ubuntu:24.04"}
	err := a.checkSandboxRunning(context.Background(), conn)
	if err == nil {
		t.Fatalf("expected error for stopped sandbox")
	}
	if !strings.Contains(err.Error(), "python") || !strings.Contains(err.Error(), "exited") {
		t.Fatalf("error should mention name and state, got: %v", err)
	}
	if !strings.Contains(err.Error(), "mantisctl sandbox start python") {
		t.Fatalf("error should suggest mantisctl start, got: %v", err)
	}
}

func TestCheckSandboxRunning_InspectError(t *testing.T) {
	a := &MantisAgent{runtime: &fakeRuntime{err: errors.New("no such container")}}
	conn := types.Connection{Name: "sb-custom", Dockerfile: "FROM ubuntu:24.04"}
	err := a.checkSandboxRunning(context.Background(), conn)
	if err == nil {
		t.Fatalf("expected error when inspect fails")
	}
	if !strings.Contains(err.Error(), "sb-custom") {
		t.Fatalf("error should contain connection name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "mantisctl sandbox start custom") {
		t.Fatalf("error should strip sb- prefix in mantisctl hint, got: %v", err)
	}
}
