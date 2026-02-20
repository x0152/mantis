package shared

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	// DefaultMaxArtifactBytes is the per-file limit for temporary artifacts.
	DefaultMaxArtifactBytes = 10 * 1024 * 1024

	// DefaultMaxTotalArtifactBytes is a soft cap for all artifacts in one session.
	// Keeps memory usage bounded even if the model accumulates multiple files.
	DefaultMaxTotalArtifactBytes = 50 * 1024 * 1024

	// DefaultArtifactTTL controls how long artifacts are kept in memory.
	// Artifacts are never persisted; they expire automatically.
	DefaultArtifactTTL = 30 * time.Minute
)

type Artifact struct {
	ID        string
	Name      string
	Format    string
	MIME      string
	Bytes     []byte
	SizeBytes int64
	SHA256    string
	CreatedAt time.Time
}

type ArtifactMeta struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Format    string `json:"format,omitempty"`
	MIME      string `json:"mime,omitempty"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type OutgoingArtifact struct {
	ArtifactID string `json:"artifact_id"`
	FileName   string `json:"file_name"`
	Caption    string `json:"caption,omitempty"`
}

type ArtifactStore struct {
	MaxFileBytes  int64
	MaxTotalBytes int64
	TTL           time.Duration

	mu         sync.Mutex
	totalBytes int64
	items      map[string]Artifact
	outgoing   map[string][]OutgoingArtifact // request_id -> files queued for delivery
}

func NewArtifactStore() *ArtifactStore {
	return &ArtifactStore{
		MaxFileBytes:  DefaultMaxArtifactBytes,
		MaxTotalBytes: DefaultMaxTotalArtifactBytes,
		TTL:           DefaultArtifactTTL,
		items:         make(map[string]Artifact),
		outgoing:      make(map[string][]OutgoingArtifact),
	}
}

func (s *ArtifactStore) Put(name string, data []byte, mime string) (ArtifactMeta, error) {
	if s == nil {
		return ArtifactMeta{}, errors.New("artifact store is nil")
	}
	if name == "" {
		name = "artifact"
	}
	if int64(len(data)) > s.MaxFileBytes {
		return ArtifactMeta{}, fmt.Errorf("artifact %q too large: %d bytes (max %d)", name, len(data), s.MaxFileBytes)
	}

	sum := sha256.Sum256(data)
	format := formatFromName(name)
	if format == "" {
		format = formatFromMIME(mime)
	}
	now := time.Now().UTC()
	meta := Artifact{
		ID:        uuid.New().String(),
		Name:      name,
		Format:    format,
		MIME:      mime,
		Bytes:     data,
		SizeBytes: int64(len(data)),
		SHA256:    hex.EncodeToString(sum[:]),
		CreatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneExpiredLocked(now)

	if s.MaxTotalBytes > 0 && s.totalBytes+meta.SizeBytes > s.MaxTotalBytes {
		return ArtifactMeta{}, fmt.Errorf("artifact store total size exceeded: %d + %d > %d", s.totalBytes, meta.SizeBytes, s.MaxTotalBytes)
	}

	s.items[meta.ID] = meta
	s.totalBytes += meta.SizeBytes
	return metaToPublic(meta), nil
}

func (s *ArtifactStore) Get(id string) (Artifact, bool) {
	if s == nil {
		return Artifact{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(time.Now().UTC())
	a, ok := s.items[id]
	return a, ok
}

func (s *ArtifactStore) List() []ArtifactMeta {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(time.Now().UTC())
	out := make([]ArtifactMeta, 0, len(s.items))
	for _, a := range s.items {
		out = append(out, metaToPublic(a))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *ArtifactStore) MarkToSend(requestID, artifactID, fileName, caption string) error {
	if s == nil {
		return errors.New("artifact store is nil")
	}
	if requestID == "" {
		return errors.New("request_id is required")
	}
	if artifactID == "" {
		return errors.New("artifact_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(time.Now().UTC())
	a, ok := s.items[artifactID]
	if !ok {
		return fmt.Errorf("unknown artifact_id: %s", artifactID)
	}
	if fileName == "" {
		fileName = a.Name
	}
	s.outgoing[requestID] = append(s.outgoing[requestID], OutgoingArtifact{
		ArtifactID: artifactID,
		FileName:   fileName,
		Caption:    caption,
	})
	return nil
}

func (s *ArtifactStore) Outgoing(requestID string) []OutgoingArtifact {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(time.Now().UTC())
	queued := s.outgoing[requestID]
	out := make([]OutgoingArtifact, len(queued))
	copy(out, queued)
	return out
}

func (s *ArtifactStore) TakeOutgoing(requestID string) []OutgoingArtifact {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(time.Now().UTC())
	queued := s.outgoing[requestID]
	if len(queued) == 0 {
		return nil
	}
	out := make([]OutgoingArtifact, len(queued))
	copy(out, queued)
	delete(s.outgoing, requestID)
	return out
}

type ArtifactStats struct {
	Count      int
	TotalBytes int64
	Outgoing   int
}

func (s *ArtifactStore) Stats() ArtifactStats {
	if s == nil {
		return ArtifactStats{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(time.Now().UTC())
	outgoing := 0
	for _, list := range s.outgoing {
		outgoing += len(list)
	}
	return ArtifactStats{
		Count:      len(s.items),
		TotalBytes: s.totalBytes,
		Outgoing:   outgoing,
	}
}

func (s *ArtifactStore) PruneExpired() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.pruneExpiredLocked(time.Now().UTC())
	s.mu.Unlock()
}

func (s *ArtifactStore) pruneExpiredLocked(now time.Time) {
	if s.TTL <= 0 {
		return
	}
	if len(s.items) == 0 && len(s.outgoing) == 0 {
		return
	}

	// Drop expired artifacts first.
	for id, a := range s.items {
		if a.CreatedAt.Add(s.TTL).Before(now) {
			delete(s.items, id)
			s.totalBytes -= a.SizeBytes
		}
	}
	if s.totalBytes < 0 {
		s.totalBytes = 0
	}

	// Drop queued outgoing entries referencing missing/expired artifacts.
	if len(s.outgoing) > 0 {
		for reqID, list := range s.outgoing {
			dst := list[:0]
			for _, o := range list {
				if _, ok := s.items[o.ArtifactID]; ok {
					dst = append(dst, o)
				}
			}
			if len(dst) == 0 {
				delete(s.outgoing, reqID)
			} else {
				s.outgoing[reqID] = dst
			}
		}
	}
}

func ArtifactInlinePreview(a Artifact, maxBytes int) string {
	if maxBytes <= 0 {
		maxBytes = 8 * 1024
	}
	if len(a.Bytes) == 0 {
		return ""
	}

	b := a.Bytes
	if len(b) > maxBytes {
		b = b[:maxBytes]
	}

	looksBinary := bytes.IndexByte(b, 0) != -1 || !utf8.Valid(b)
	if looksBinary {
		// Keep it short even for binary: show a small base64 prefix.
		const maxB64 = 2048
		enc := base64.StdEncoding.EncodeToString(b)
		if len(enc) > maxB64 {
			enc = enc[:maxB64] + "...(truncated)"
		}
		return fmt.Sprintf("[binary preview: base64]\n%s", enc)
	}

	text := string(b)
	if len(a.Bytes) > maxBytes {
		text += "\n...(truncated)"
	}
	return text
}

func metaToPublic(a Artifact) ArtifactMeta {
	return ArtifactMeta{
		ID:        a.ID,
		Name:      a.Name,
		Format:    a.Format,
		MIME:      a.MIME,
		SizeBytes: a.SizeBytes,
		SHA256:    a.SHA256,
	}
}

func formatFromName(name string) string {
	if name == "" {
		return ""
	}
	base := path.Base(name)
	ext := path.Ext(base)
	if ext == "" {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

func formatFromMIME(mime string) string {
	if mime == "" {
		return ""
	}
	m := strings.ToLower(strings.TrimSpace(strings.SplitN(mime, ";", 2)[0]))
	switch m {
	case "text/plain":
		return "txt"
	case "application/json":
		return "json"
	case "text/markdown":
		return "md"
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "audio/mpeg":
		return "mp3"
	case "audio/mp4":
		return "m4a"
	case "audio/ogg":
		return "ogg"
	case "audio/wav", "audio/x-wav":
		return "wav"
	default:
		return ""
	}
}
