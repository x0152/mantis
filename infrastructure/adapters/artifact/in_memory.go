package artifact

import (
	"sync"

	"mantis/core/protocols"
)

type InMemorySessionStorage struct {
	mu       sync.Mutex
	sessions map[string]protocols.ArtifactSessionRecord
}

func NewInMemorySessionStorage() *InMemorySessionStorage {
	return &InMemorySessionStorage{
		sessions: make(map[string]protocols.ArtifactSessionRecord),
	}
}

func (s *InMemorySessionStorage) Get(sessionID string) (protocols.ArtifactSessionRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.sessions[sessionID]
	return v, ok
}

func (s *InMemorySessionStorage) Set(sessionID string, session protocols.ArtifactSessionRecord) {
	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()
}

func (s *InMemorySessionStorage) Delete(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

func (s *InMemorySessionStorage) List() map[string]protocols.ArtifactSessionRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]protocols.ArtifactSessionRecord, len(s.sessions))
	for k, v := range s.sessions {
		out[k] = v
	}
	return out
}
