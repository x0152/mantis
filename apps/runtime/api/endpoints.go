package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

const registeredConnectionPrefix = "sb-"

type Endpoints struct {
	rt              protocols.Runtime
	connectionStore protocols.Store[string, types.Connection]
	token           string
}

func NewEndpoints(rt protocols.Runtime, connectionStore protocols.Store[string, types.Connection], token string) *Endpoints {
	return &Endpoints{rt: rt, connectionStore: connectionStore, token: token}
}

func (e *Endpoints) Mount(r chi.Router) {
	r.Route("/api/runtime", func(r chi.Router) {
		r.Use(e.authMiddleware)
		r.Post("/build", e.build)
		r.Post("/run", e.run)
		r.Get("/containers", e.list)
		r.Get("/containers/{name}", e.inspect)
		r.Get("/containers/{name}/logs", e.logs)
		r.Post("/containers/{name}/stop", e.stop)
		r.Delete("/containers/{name}", e.remove)
		r.Post("/register", e.register)
		r.Delete("/register/{name}", e.unregister)
	})
}

func (e *Endpoints) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if e.token == "" || r.Header.Get("X-Runtime-Token") == e.token {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}

func (e *Endpoints) build(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	dockerfile, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stream, err := e.rt.Build(r.Context(), name, dockerfile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()
	pipeStream(w, stream)
}

func (e *Endpoints) run(w http.ResponseWriter, r *http.Request) {
	var input RunInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	container, err := e.rt.Run(r.Context(), types.RuntimeRunSpec{
		Name:    input.Name,
		Image:   input.Image,
		Network: input.Network,
		Env:     input.Env,
		Labels:  input.Labels,
		Cmd:     input.Cmd,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ContainerOutput{Container: container})
}

func (e *Endpoints) list(w http.ResponseWriter, r *http.Request) {
	items, err := e.rt.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ContainerListOutput{Containers: items})
}

func (e *Endpoints) inspect(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	c, err := e.rt.Inspect(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, ContainerOutput{Container: c})
}

func (e *Endpoints) logs(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
	follow := r.URL.Query().Get("follow") == "1" || r.URL.Query().Get("follow") == "true"

	ctx := r.Context()
	if follow {
		ctx = withCancelOnClose(ctx, r)
	}
	stream, err := e.rt.Logs(ctx, name, tail, follow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer stream.Close()
	pipeStream(w, stream)
}

func (e *Endpoints) stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := e.rt.Stop(r.Context(), name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (e *Endpoints) remove(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	_ = e.deregisterByName(r.Context(), name)
	if err := e.rt.Remove(r.Context(), name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (e *Endpoints) register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if e.connectionStore == nil {
		http.Error(w, "connection store not configured", http.StatusInternalServerError)
		return
	}

	container, err := e.rt.Inspect(r.Context(), input.Name)
	if err != nil {
		http.Error(w, "container not found: "+err.Error(), http.StatusNotFound)
		return
	}

	username := input.Username
	if username == "" {
		username = "mantis"
	}
	password := input.Password
	if password == "" {
		password = "mantis"
	}
	port := input.Port
	if port == 0 {
		port = 22
	}
	profileIDs := input.ProfileIDs
	if profileIDs == nil {
		profileIDs = []string{}
	}

	config, err := json.Marshal(map[string]any{
		"host":     container.Host,
		"port":     port,
		"username": username,
		"password": password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	connectionName := registeredConnectionPrefix + input.Name
	existing, err := e.findConnectionByName(r.Context(), connectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	conn := types.Connection{
		Type:          "ssh",
		Name:          connectionName,
		Description:   input.Description,
		Config:        config,
		ProfileIDs:    profileIDs,
		Memories:      []types.Memory{},
		MemoryEnabled: true,
	}

	var saved types.Connection
	if existing != nil {
		existing.Description = input.Description
		existing.Config = config
		existing.ProfileIDs = profileIDs
		updated, uerr := e.connectionStore.Update(r.Context(), []types.Connection{*existing})
		if uerr != nil {
			http.Error(w, uerr.Error(), http.StatusInternalServerError)
			return
		}
		saved = updated[0]
	} else {
		conn.ID = uuid.New().String()
		created, cerr := e.connectionStore.Create(r.Context(), []types.Connection{conn})
		if cerr != nil {
			http.Error(w, cerr.Error(), http.StatusInternalServerError)
			return
		}
		saved = created[0]
	}

	writeJSON(w, http.StatusOK, RegisterOutput{Connection: saved})
}

func (e *Endpoints) unregister(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := e.deregisterByName(r.Context(), name); err != nil {
		if errors.Is(err, errConnectionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var errConnectionNotFound = fmt.Errorf("registered connection not found")

func (e *Endpoints) deregisterByName(ctx context.Context, sandboxName string) error {
	if e.connectionStore == nil {
		return nil
	}
	conn, err := e.findConnectionByName(ctx, registeredConnectionPrefix+sandboxName)
	if err != nil {
		return err
	}
	if conn == nil {
		return errConnectionNotFound
	}
	return e.connectionStore.Delete(ctx, []string{conn.ID})
}

func (e *Endpoints) findConnectionByName(ctx context.Context, name string) (*types.Connection, error) {
	items, err := e.connectionStore.List(ctx, types.ListQuery{Filter: map[string]string{"name": name}, Page: types.Page{Limit: 1}})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func pipeStream(w http.ResponseWriter, src io.Reader) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, _ := w.(http.Flusher)
	buf := make([]byte, 4096)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func withCancelOnClose(ctx context.Context, r *http.Request) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-r.Context().Done()
		cancel()
	}()
	return ctx
}
