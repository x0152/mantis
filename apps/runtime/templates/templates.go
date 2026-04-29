package templates

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"mantis/sandboxes"
)

type Template struct {
	Name        string
	Description string
	ProfileID   string
	Dockerfile  string
	CapAdd      []string
}

type BuiltinMeta struct {
	Name        string
	Description string
	ProfileID   string
	CapAdd      []string
}

var builtinMeta = []BuiltinMeta{
	{Name: "base", ProfileID: "base", Description: "General-purpose workhorse sandbox — Python 3.12 + scientific stack, DB clients (psql, mysql, redis, sqlite), shell and networking utilities."},
	{Name: "browser", ProfileID: "browser", Description: "Headless Chromium + Playwright — web navigation, screenshots, PDF, parsing.", CapAdd: []string{"SYS_ADMIN"}},
	{Name: "ffmpeg", ProfileID: "media", Description: "FFmpeg + MediaInfo + ImageMagick — video, audio, image processing."},
	{Name: "netsec", ProfileID: "netsec", Description: "Network / pentest toolkit — nmap, dig, nikto, ffuf, hashcat + net-* wrappers with hard timeouts."},
	{Name: "runtimectl", ProfileID: "runtimectl", Description: "Runtime controller. Ask it in plain language to provision a new sandbox (e.g. \"need rust + cargo + curl\"); it builds, runs and registers the container."},
}

func Builtin() ([]Template, error) {
	out := make([]Template, 0, len(builtinMeta))
	for _, m := range builtinMeta {
		df, err := Render(m.Name)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", m.Name, err)
		}
		out = append(out, Template{
			Name:        m.Name,
			Description: m.Description,
			ProfileID:   m.ProfileID,
			Dockerfile:  df,
			CapAdd:      m.CapAdd,
		})
	}
	return out, nil
}

func Lookup(name string) (BuiltinMeta, bool) {
	for _, m := range builtinMeta {
		if m.Name == name {
			return m, true
		}
	}
	return BuiltinMeta{}, false
}

func Render(name string) (string, error) {
	dockerfile, err := fs.ReadFile(sandboxes.FS, path.Join(name, "Dockerfile"))
	if err != nil {
		return "", err
	}
	expanded, err := inlineCopies(sandboxes.FS, name, string(dockerfile))
	if err != nil {
		return "", err
	}
	return harden(expanded), nil
}

const initScript = `#!/bin/sh
set -eu
mkdir -p /run/sshd
chown mantis:mantis /home/mantis
chmod 755 /home/mantis
mkdir -p /home/mantis/.ssh
if [ -n "${MANTIS_SSH_PUBLIC_KEY:-}" ]; then
    printf '%s\n' "$MANTIS_SSH_PUBLIC_KEY" > /home/mantis/.ssh/authorized_keys
    chmod 600 /home/mantis/.ssh/authorized_keys
fi
: > /home/mantis/.ssh/environment
env | sed -n 's/^\(MANTIS_[A-Z_]*\)=\(.*\)$/\1=\2/p' | grep -v '^MANTIS_SSH_PUBLIC_KEY=' >> /home/mantis/.ssh/environment || true
chmod 600 /home/mantis/.ssh/environment
chmod 700 /home/mantis/.ssh
chown -R mantis:mantis /home/mantis/.ssh
exec /usr/sbin/sshd -D -e \
    -o PasswordAuthentication=no \
    -o PermitRootLogin=no \
    -o KbdInteractiveAuthentication=no \
    -o UsePAM=no \
    -o PermitUserEnvironment=yes
`

func harden(dockerfile string) string {
	var out strings.Builder
	for _, line := range strings.Split(dockerfile, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "CMD ") || strings.HasPrefix(trimmed, "ENTRYPOINT ") {
			continue
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	// `adduser -D` (Alpine) and `useradd -m` (Debian/Ubuntu) both create the
	// account with a `!` in /etc/shadow, which makes sshd reject the user as
	// "account locked". Replace it with `*` so password login stays
	// impossible but the account is otherwise valid for pubkey auth.
	out.WriteString("RUN echo 'mantis:*' | chpasswd -e\n")
	out.WriteString("RUN mkdir -p /usr/local/sbin && ssh-keygen -A\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(initScript))
	fmt.Fprintf(&out, "RUN printf %%s %q | base64 -d > /usr/local/sbin/mantis-init && chmod 755 /usr/local/sbin/mantis-init\n", encoded)
	out.WriteString(`CMD ["/usr/local/sbin/mantis-init"]` + "\n")
	return out.String()
}

func inlineCopies(root fs.FS, sandbox, dockerfile string) (string, error) {
	var out strings.Builder
	for _, line := range strings.Split(dockerfile, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "COPY ") {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) != 3 {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		src, dst := fields[1], fields[2]
		if strings.HasPrefix(src, "/") || strings.Contains(src, "..") {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}

		src = strings.TrimPrefix(src, "./")
		srcPath := path.Join(sandbox, src)

		entry, err := fs.Stat(root, srcPath)
		if err != nil {
			return "", fmt.Errorf("COPY %s: %w", src, err)
		}
		if entry.IsDir() {
			if err := emitDirectoryCopy(&out, root, srcPath, dst); err != nil {
				return "", err
			}
		} else {
			content, err := fs.ReadFile(root, srcPath)
			if err != nil {
				return "", err
			}
			emitFileCopy(&out, dst, content)
		}
	}
	return out.String(), nil
}

func emitDirectoryCopy(out *strings.Builder, root fs.FS, srcDir, dstDir string) error {
	entries, err := fs.ReadDir(root, srcDir)
	if err != nil {
		return err
	}
	dstDir = strings.TrimSuffix(dstDir, "/")
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		content, err := fs.ReadFile(root, path.Join(srcDir, e.Name()))
		if err != nil {
			return err
		}
		emitFileCopy(out, path.Join(dstDir, e.Name()), content)
	}
	return nil
}

func emitFileCopy(out *strings.Builder, dst string, content []byte) {
	encoded := base64.StdEncoding.EncodeToString(content)
	fmt.Fprintf(out, "RUN mkdir -p %q && printf %%s %q | base64 -d > %q && chmod a+rx %q\n",
		path.Dir(dst), encoded, dst, dst)
}
