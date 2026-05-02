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

	defaultMemoryBytes = int64(512 * 1024 * 1024)
	defaultNanoCPUs    = int64(1_000_000_000)
	defaultPidsLimit   = int64(256)
)

// infrastructureCaps are the capabilities the sandbox init script + sshd
// require to function. They are always granted (unless the runtime is in
// privileged mode, which implies all caps anyway), regardless of what the
// user configures through RUNTIME_SANDBOX_CAPS or per-template CapAdd.
//
//   - CHOWN, DAC_OVERRIDE, FOWNER     — init creates /home/mantis/.ssh
//     inside a volume owned by the mantis user.
//   - SETUID, SETGID                  — sshd drops privileges to the user.
//   - SYS_CHROOT                      — sshd privilege-separation sandbox.
//   - KILL                            — sshd manages child processes.
//   - AUDIT_WRITE                     — sshd writes login audit records.
//   - NET_BIND_SERVICE                — sshd binds port 22.
var infrastructureCaps = []string{
	"CHOWN", "DAC_OVERRIDE", "FOWNER",
	"SETUID", "SETGID",
	"SYS_CHROOT", "KILL",
	"AUDIT_WRITE", "NET_BIND_SERVICE",
}

type Options struct {
	SocketPath string
	Network    string
	// DefaultCaps is added to every sandbox on top of per-template CapAdd.
	// Capability names accept either "NET_ADMIN" or "CAP_NET_ADMIN" form.
	DefaultCaps []string
	// Privileged grants every capability and disables seccomp/AppArmor —
	// equivalent to `docker run --privileged`. Implies all caps.
	Privileged bool
}

type Runtime struct {
	socketPath  string
	network     string
	defaultCaps []string
	privileged  bool
	client      *http.Client
}

func New(opts Options) *Runtime {
	socketPath := opts.SocketPath
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}
	network := opts.Network
	if network == "" {
		network = defaultNetwork
	}
	return &Runtime{
		socketPath:  socketPath,
		network:     network,
		defaultCaps: normalizeCaps(opts.DefaultCaps),
		privileged:  opts.Privileged,
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

// normalizeCaps strips the optional CAP_ prefix and uppercases entries so the
// caller can mix "net_admin", "NET_ADMIN" and "CAP_NET_ADMIN" freely.
func normalizeCaps(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, c := range in {
		c = strings.TrimSpace(strings.ToUpper(c))
		if c == "" {
			continue
		}
		c = strings.TrimPrefix(c, "CAP_")
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

func mergeCaps(lists ...[]string) []string {
	total := 0
	for _, l := range lists {
		total += len(l)
	}
	merged := make([]string, 0, total)
	for _, l := range lists {
		merged = append(merged, l...)
	}
	return normalizeCaps(merged)
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
func sandboxNetwork(name string) string { return containerNamePrefix + name + "-net" }
func homeVolume(name string) string     { return containerNamePrefix + name + "-home" }

func (r *Runtime) Build(ctx context.Context, name string, dockerfile []byte) (io.ReadCloser, error) {
	return r.BuildWithLabels(ctx, name, dockerfile, nil)
}

// BuildWithLabels behaves like Build but stamps additional labels onto the
// resulting image. Useful for callers (sandbox-prebuild, bootstrap) that need
// to attach a dockerfile-content hash so subsequent runs can detect drift
// without re-building.
func (r *Runtime) BuildWithLabels(ctx context.Context, name string, dockerfile []byte, extraLabels map[string]string) (io.ReadCloser, error) {
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

	labels := map[string]string{
		labelMarker: "1",
		labelName:   name,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("t", imageTag(name))
	q.Set("dockerfile", "Dockerfile")
	q.Set("rm", "1")
	q.Set("labels", string(labelsJSON))

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

// Read drains the Docker build progress stream. Build failures arrive as
// {"error": "..."} JSON frames; we surface them as a Go error so callers that
// io.Copy the stream into a buffer don't silently treat a failed build as
// success.
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
		case msg.Error != "":
			s.pending.WriteString("error: " + msg.Error + "\n")
			s.done = true
			detail := msg.Error
			if msg.ErrorDetail.Message != "" {
				detail = msg.ErrorDetail.Message
			}
			s.err = fmt.Errorf("build failed: %s", strings.TrimSpace(detail))
		case msg.Stream != "":
			s.pending.WriteString(msg.Stream)
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
		network = sandboxNetwork(spec.Name)
	}
	volume := homeVolume(spec.Name)

	labels := map[string]string{labelMarker: "1", labelName: spec.Name}
	for k, v := range spec.Labels {
		labels[k] = v
	}
	env := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		env = append(env, k+"="+v)
	}

	hostConfig := map[string]any{
		"NetworkMode":    network,
		"RestartPolicy":  map[string]any{"Name": "unless-stopped"},
		"ReadonlyRootfs": true,
		"CapDrop":        []string{"ALL"},
		"SecurityOpt":    []string{"no-new-privileges"},
		"Memory":         defaultMemoryBytes,
		"MemorySwap":     defaultMemoryBytes,
		"NanoCpus":       defaultNanoCPUs,
		"PidsLimit":      defaultPidsLimit,
		"Sysctls": map[string]string{
			"net.ipv4.ping_group_range": "0 2147483647",
		},
		"Tmpfs": map[string]string{
			"/tmp":     "rw,size=256m,nosuid,nodev",
			"/run":     "rw,size=64m,nosuid,nodev",
			"/var/log": "rw,size=64m,nosuid,nodev",
			"/var/tmp": "rw,size=64m,nosuid,nodev",
		},
		"Mounts": []map[string]any{
			{"Type": "volume", "Source": volume, "Target": "/home/mantis"},
		},
	}
	if r.privileged {
		hostConfig["Privileged"] = true
		delete(hostConfig, "CapDrop")
		delete(hostConfig, "SecurityOpt")
	} else {
		hostConfig["CapAdd"] = mergeCaps(infrastructureCaps, r.defaultCaps, spec.CapAdd)
	}

	body := map[string]any{
		"Image":        image,
		"Labels":       labels,
		"Env":          env,
		"AttachStdin":  false,
		"AttachStdout": false,
		"AttachStderr": false,
		"Tty":          false,
		"HostConfig":   hostConfig,
	}
	if len(spec.Cmd) > 0 {
		body["Cmd"] = spec.Cmd
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return types.RuntimeContainer{}, err
	}

	_ = r.removeIfExists(ctx, spec.Name)
	if err := r.ensureNetwork(ctx, network, spec.Internal); err != nil {
		return types.RuntimeContainer{}, err
	}
	if err := r.ensureVolume(ctx, volume, spec.Name); err != nil {
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

func (r *Runtime) ensureNetwork(ctx context.Context, name string, internal bool) error {
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
	body, _ := json.Marshal(map[string]any{
		"Name":     name,
		"Driver":   "bridge",
		"Internal": internal,
		"Labels":   map[string]string{labelMarker: "1"},
	})
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

func (r *Runtime) ensureVolume(ctx context.Context, name, sandbox string) error {
	body, _ := json.Marshal(map[string]any{
		"Name": name,
		"Labels": map[string]string{
			labelMarker: "1",
			labelName:   sandbox,
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/volumes/create", nil), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (r *Runtime) removeVolume(ctx context.Context, name string) {
	q := url.Values{}
	q.Set("force", "1")
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.url("/volumes/"+name, q), nil)
	if err != nil {
		return
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (r *Runtime) removeNetwork(ctx context.Context, name string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.url("/networks/"+name, nil), nil)
	if err != nil {
		return
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
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
	if err := r.removeIfExists(ctx, name); err != nil {
		return err
	}
	r.removeVolume(ctx, homeVolume(name))
	r.removeNetwork(ctx, sandboxNetwork(name))
	return nil
}

type networkEndpoint struct {
	IPAddress string `json:"IPAddress"`
}

// pickIP returns the container IP on the preferred network, falling back to
// any non-empty address. We need this when callers (e.g. Mantis app running
// in a Kubernetes pod next to a DinD sidecar) cannot rely on Docker's
// embedded DNS to resolve container names.
func pickIP(networks map[string]networkEndpoint, preferred string) string {
	if n, ok := networks[preferred]; ok && n.IPAddress != "" {
		return n.IPAddress
	}
	for _, n := range networks {
		if n.IPAddress != "" {
			return n.IPAddress
		}
	}
	return ""
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
		Names           []string          `json:"Names"`
		Image           string            `json:"Image"`
		State           string            `json:"State"`
		Labels          map[string]string `json:"Labels"`
		Created         int64             `json:"Created"`
		NetworkSettings struct {
			Networks map[string]networkEndpoint `json:"Networks"`
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
			IP:        pickIP(c.NetworkSettings.Networks, r.network),
			Labels:    c.Labels,
			CreatedAt: time.Unix(c.Created, 0),
		})
	}
	return out, nil
}

// ImageLabels returns the labels attached to the sandbox image for the given
// name. Returns (nil, nil) when the image does not exist locally — the caller
// can treat that as "needs build". Any other error (network, malformed JSON)
// is surfaced as-is.
func (r *Runtime) ImageLabels(ctx context.Context, name string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url("/images/"+imageTag(name)+"/json", nil), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("docker api %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var raw struct {
		Config struct {
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Config.Labels == nil {
		return map[string]string{}, nil
	}
	return raw.Config.Labels, nil
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
		Created         string `json:"Created"`
		NetworkSettings struct {
			Networks map[string]networkEndpoint `json:"Networks"`
		} `json:"NetworkSettings"`
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
		IP:        pickIP(raw.NetworkSettings.Networks, r.network),
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
