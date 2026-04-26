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
}

type BuiltinMeta struct {
	Name        string
	Description string
	ProfileID   string
}

var builtinMeta = []BuiltinMeta{
	{Name: "base", ProfileID: "base", Description: "General-purpose Linux sandbox — shell, files, networking utilities."},
	{Name: "python", ProfileID: "python", Description: "Python 3 sandbox — scripts, data analysis, pip packages."},
	{Name: "browser", ProfileID: "browser", Description: "Headless Chromium + Playwright — web navigation, screenshots, PDF, parsing."},
	{Name: "ffmpeg", ProfileID: "media", Description: "FFmpeg + MediaInfo + ImageMagick — video, audio, image processing."},
	{Name: "db", ProfileID: "database", Description: "Database clients — psql, mysql, redis-cli, sqlite3."},
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
		})
	}
	return out, nil
}

func Render(name string) (string, error) {
	dockerfile, err := fs.ReadFile(sandboxes.FS, path.Join(name, "Dockerfile"))
	if err != nil {
		return "", err
	}
	return inlineCopies(sandboxes.FS, name, string(dockerfile))
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
