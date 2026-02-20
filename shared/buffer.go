package shared

import (
	"encoding/json"
	"sync"

	"mantis/core/types"
)

type BufferEntry struct {
	Content string
	Steps   []types.Step
}

type Buffer struct {
	mu   sync.RWMutex
	data map[string]*BufferEntry
}

func NewBuffer() *Buffer {
	return &Buffer{data: make(map[string]*BufferEntry)}
}

func (b *Buffer) entry(id string) *BufferEntry {
	e, ok := b.data[id]
	if !ok {
		e = &BufferEntry{}
		b.data[id] = e
	}
	return e
}

func (b *Buffer) SetContent(id, content string) {
	b.mu.Lock()
	b.entry(id).Content = content
	b.mu.Unlock()
}

func (b *Buffer) SetStep(id string, step types.Step) {
	b.mu.Lock()
	e := b.entry(id)
	for i, s := range e.Steps {
		if s.ID == step.ID {
			e.Steps[i] = step
			b.mu.Unlock()
			return
		}
	}
	e.Steps = append(e.Steps, step)
	b.mu.Unlock()
}

func (b *Buffer) Get(id string) (BufferEntry, bool) {
	b.mu.RLock()
	e, ok := b.data[id]
	if !ok {
		b.mu.RUnlock()
		return BufferEntry{}, false
	}
	copy := *e
	copy.Steps = make([]types.Step, len(e.Steps))
	for i := range e.Steps {
		copy.Steps[i] = e.Steps[i]
	}
	b.mu.RUnlock()
	return copy, true
}

func (b *Buffer) Delete(id string) {
	b.mu.Lock()
	delete(b.data, id)
	b.mu.Unlock()
}

func StepsToJSON(steps []types.Step) json.RawMessage {
	if len(steps) == 0 {
		return nil
	}
	data, _ := json.Marshal(steps)
	return data
}
