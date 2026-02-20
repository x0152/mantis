package protocols

import "time"

type ArtifactSessionRecord struct {
	Store      any
	LastAccess time.Time
}

type ArtifactSessionStorage interface {
	Get(sessionID string) (ArtifactSessionRecord, bool)
	Set(sessionID string, session ArtifactSessionRecord)
	Delete(sessionID string)
	List() map[string]ArtifactSessionRecord
}
