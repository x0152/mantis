package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mantis/core/types"
)

const (
	containerNamePrefix = "mantis-sb-"
	imageTagPrefix      = "mantis-sb/"
	labelMarker         = "mantis.sandbox"
	labelName           = "mantis.sandbox.name"
	defaultNetwork      = "mantis-sandbox-net"
	apiVersion          = "v1.43"
)

type Runtime struct {
	socketPath string
	network    string
	client     *http.Client
}

func New(socketPath, network string) *Runtime {
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}
	if network == "" {
		network = defaultNetwork
	}
	return &Runtime{
		socketPath: socketPath,
		network:    network,
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", socketPath)
				},
			},
		},
	}
}

func (r *Runtime) Network() string { return r.network }

func (r *Runtime) url(path string, query url.Values) string {
	u := "http://unix/" + apiVersion + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

func (r *Runtime) do(req *http.Request) (*http.Response, error) {
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("docker api %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func containerName(name string) string { return containerNamePrefix + name }
func imageTag(name string) string      { return imageTagPrefix + name + ":latest" }

func (r *Runtime) Build(ctx context.Context, name string, dockerfile []byte) (io.ReadCloser, error) {
	if len(dockerfile) == 0 {
		return nil, errors.New("empty Dockerfile")
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name:    "Dockerfile",
		Mode:    0644,
		Size:    int64(len(dockerfile)),
		ModTime: time.Now(),
	}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(dockerfile); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("t", imageTag(name))
	q.Set("dockerfile", "Dockerfile")
	q.Set("rm", "1")
	q.Set("labels", fmt.Sprintf(`{%q:"1",%q:%q}`, labelMarker, labelName, name))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/build", q), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-tar")
	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}
	return newBuildStream(resp.Body), nil
}

type buildStream struct {
	src     io.ReadCloser
	dec     *json.Decoder
	pending bytes.Buffer
	done    bool
	err     error
}

func newBuildStream(src io.ReadCloser) *buildStream {
	return &buildStream{src: src, dec: json.NewDecoder(src)}
}

func (s *buildStream) Read(p []byte) (int, error) {
	for s.pending.Len() == 0 && !s.done {
		var msg struct {
			Stream      string `json:"stream"`
			Status      string `json:"status"`
			ErrorDetail struct {
				Message string `json:"message"`
			} `json:"errorDetail"`
			Error string `json:"error"`
		}
		if err := s.dec.Decode(&msg); err != nil {
			s.done = true
			if err != io.EOF {
				s.err = err
			}
			break
		}
		switch {
		case msg.Stream != "":
			s.pending.WriteString(msg.Stream)
		case msg.Error != "":
			s.pending.WriteString("error: " + msg.Error + "\n")
		case msg.Status != "":
			s.pending.WriteString(msg.Status + "\n")
		}
	}
	if s.pending.Len() > 0 {
		return s.pending.Read(p)
	}
	if s.err != nil {
		return 0, s.err
	}
	return 0, io.EOF
}

func (s *buildStream) Close() error { return s.src.Close() }

func (r *Runtime) Run(ctx context.Context, spec types.RuntimeRunSpec) (types.RuntimeContainer, error) {
	if spec.Name == "" {
		return types.RuntimeContainer{}, errors.New("name required")
	}
	image := spec.Image
	if image == "" {
		image = imageTag(spec.Name)
	}
	network := spec.Network
	if network == "" {
		network = r.network
	}

	labels := map[string]string{labelMarker: "1", labelName: spec.Name}
	for k, v := range spec.Labels {
		labels[k] = v
	}
	env := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		env = append(env, k+"="+v)
	}

	body := map[string]any{
		"Image":        image,
		"Labels":       labels,
		"Env":          env,
		"AttachStdin":  false,
		"AttachStdout": false,
		"AttachStderr": false,
		"Tty":          false,
		"HostConfig": map[string]any{
			"NetworkMode":   network,
			"RestartPolicy": map[string]any{"Name": "unless-stopped"},
		},
	}
	if len(spec.Cmd) > 0 {
		body["Cmd"] = spec.Cmd
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return types.RuntimeContainer{}, err
	}

	_ = r.removeIfExists(ctx, spec.Name)
	if err := r.ensureNetwork(ctx, network); err != nil {
		return types.RuntimeContainer{}, err
	}

	q := url.Values{}
	q.Set("name", containerName(spec.Name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/containers/create", q), bytes.NewReader(raw))
	if err != nil {
		return types.RuntimeContainer{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.do(req)
	if err != nil {
		return types.RuntimeContainer{}, err
	}
	resp.Body.Close()

	startReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/containers/"+containerName(spec.Name)+"/start", nil), nil)
	if err != nil {
		return types.RuntimeContainer{}, err
	}
	startResp, err := r.do(startReq)
	if err != nil {
		return types.RuntimeContainer{}, err
	}
	startResp.Body.Close()

	return r.Inspect(ctx, spec.Name)
}

func (r *Runtime) removeIfExists(ctx context.Context, name string) error {
	q := url.Values{}
	q.Set("force", "1")
	q.Set("v", "1")
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.url("/containers/"+containerName(name), q), nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (r *Runtime) ensureNetwork(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url("/networks/"+name, nil), nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := json.Marshal(map[string]any{"Name": name, "Driver": "bridge"})
	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/networks/create", nil), bytes.NewReader(body))
	if err != nil {
		return err
	}
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := r.do(createReq)
	if err != nil {
		return err
	}
	createResp.Body.Close()
	return nil
}

func (r *Runtime) Stop(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/containers/"+containerName(name)+"/stop", nil), nil)
	if err != nil {
		return err
	}
	resp, err := r.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (r *Runtime) Remove(ctx context.Context, name string) error {
	return r.removeIfExists(ctx, name)
}

func (r *Runtime) List(ctx context.Context) ([]types.RuntimeContainer, error) {
	q := url.Values{}
	q.Set("all", "1")
	q.Set("filters", fmt.Sprintf(`{"label":[%q]}`, labelMarker+"=1"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url("/containers/json", q), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw []struct {
		Names   []string          `json:"Names"`
		Image   string            `json:"Image"`
		State   string            `json:"State"`
		Labels  map[string]string `json:"Labels"`
		Created int64             `json:"Created"`
		NetworkSettings struct {
			Networks map[string]struct {
				IPAddress string `json:"IPAddress"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]types.RuntimeContainer, 0, len(raw))
	for _, c := range raw {
		out = append(out, types.RuntimeContainer{
			Name:      c.Labels[labelName],
			Image:     c.Image,
			Status:    c.State,
			Host:      containerName(c.Labels[labelName]),
			Labels:    c.Labels,
			CreatedAt: time.Unix(c.Created, 0),
		})
	}
	return out, nil
}

func (r *Runtime) Inspect(ctx context.Context, name string) (types.RuntimeContainer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url("/containers/"+containerName(name)+"/json", nil), nil)
	if err != nil {
		return types.RuntimeContainer{}, err
	}
	resp, err := r.do(req)
	if err != nil {
		return types.RuntimeContainer{}, err
	}
	defer resp.Body.Close()

	var raw struct {
		Config struct {
			Image  string            `json:"Image"`
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
		State struct {
			Status string `json:"Status"`
		} `json:"State"`
		Created string `json:"Created"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return types.RuntimeContainer{}, err
	}
	created, _ := time.Parse(time.RFC3339Nano, raw.Created)
	return types.RuntimeContainer{
		Name:      raw.Config.Labels[labelName],
		Image:     raw.Config.Image,
		Status:    raw.State.Status,
		Host:      containerName(raw.Config.Labels[labelName]),
		Labels:    raw.Config.Labels,
		CreatedAt: created,
	}, nil
}

func (r *Runtime) Logs(ctx context.Context, name string, tail int, follow bool) (io.ReadCloser, error) {
	q := url.Values{}
	q.Set("stdout", "1")
	q.Set("stderr", "1")
	if tail > 0 {
		q.Set("tail", strconv.Itoa(tail))
	} else {
		q.Set("tail", "all")
	}
	if follow {
		q.Set("follow", "1")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url("/containers/"+containerName(name)+"/logs", q), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}
	return newLogStream(resp.Body), nil
}

type logStream struct {
	src     io.ReadCloser
	pending bytes.Buffer
}

func newLogStream(src io.ReadCloser) *logStream { return &logStream{src: src} }

func (s *logStream) Read(p []byte) (int, error) {
	for s.pending.Len() == 0 {
		var header [8]byte
		if _, err := io.ReadFull(s.src, header[:]); err != nil {
			return 0, err
		}
		size := binary.BigEndian.Uint32(header[4:8])
		if size == 0 {
			continue
		}
		if _, err := io.CopyN(&s.pending, s.src, int64(size)); err != nil {
			return 0, err
		}
	}
	return s.pending.Read(p)
}

func (s *logStream) Close() error { return s.src.Close() }
