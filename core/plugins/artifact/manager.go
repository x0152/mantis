package artifact

import (
	"strings"
	"time"

	"mantis/core/protocols"
	"mantis/shared"
)

type ManagerInput struct {
	SessionID string
}

// Manager handles artifact stores per session and delegates persistence to storage adapter.
type Manager struct {
	TTL           time.Duration
	MaxFileBytes  int64
	MaxTotalBytes int64

	storage protocols.ArtifactSessionStorage
}

func NewManager(storage protocols.ArtifactSessionStorage) *Manager {
	return &Manager{
		TTL:           shared.DefaultArtifactTTL,
		MaxFileBytes:  shared.DefaultMaxArtifactBytes,
		MaxTotalBytes: shared.DefaultMaxTotalArtifactBytes,
		storage:       storage,
	}
}

func (m *Manager) Execute(in ManagerInput) *shared.ArtifactStore {
	if m == nil {
		return shared.NewArtifactStore()
	}

	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" || m.storage == nil {
		s := shared.NewArtifactStore()
		m.configureStore(s)
		return s
	}

	now := time.Now().UTC()
	if rec, ok := m.storage.Get(sessionID); ok {
		if s, ok := rec.Store.(*shared.ArtifactStore); ok && s != nil {
			m.configureStore(s)
			rec.LastAccess = now
			rec.Store = s
			m.storage.Set(sessionID, rec)
			m.cleanup(now)
			return s
		}
	}

	s := shared.NewArtifactStore()
	m.configureStore(s)
	m.storage.Set(sessionID, protocols.ArtifactSessionRecord{
		Store:      s,
		LastAccess: now,
	})
	m.cleanup(now)
	return s
}

func (m *Manager) ForSession(sessionID string) *shared.ArtifactStore {
	return m.Execute(ManagerInput{SessionID: sessionID})
}

func (m *Manager) configureStore(s *shared.ArtifactStore) {
	if s == nil {
		return
	}
	s.TTL = m.TTL
	s.MaxFileBytes = m.MaxFileBytes
	s.MaxTotalBytes = m.MaxTotalBytes
}

func (m *Manager) cleanup(now time.Time) {
	if m == nil || m.storage == nil {
		return
	}
	ttl := m.TTL
	if ttl <= 0 {
		ttl = shared.DefaultArtifactTTL
	}

	for sessionID, rec := range m.storage.List() {
		s, ok := rec.Store.(*shared.ArtifactStore)
		if !ok || s == nil {
			m.storage.Delete(sessionID)
			continue
		}
		if now.Sub(rec.LastAccess) <= ttl {
			continue
		}
		s.PruneExpired()
		stats := s.Stats()
		if stats.Count == 0 && stats.Outgoing == 0 {
			m.storage.Delete(sessionID)
		}
	}
}
