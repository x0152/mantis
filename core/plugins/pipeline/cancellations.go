package pipeline

import (
	"context"
	"sync"
	"time"
)

const cancelWaitTimeout = 5 * time.Second

type cancelEntry struct {
	token  uint64
	cancel context.CancelFunc
	done   chan struct{}
}

type Cancellations struct {
	mu      sync.Mutex
	entries map[string]*cancelEntry
	nextTok uint64
}

func NewCancellations() *Cancellations {
	return &Cancellations{entries: make(map[string]*cancelEntry)}
}

func (c *Cancellations) Begin(parent context.Context, key string) (context.Context, func()) {
	if c == nil || key == "" {
		ctx, cancel := context.WithCancel(parent)
		return ctx, cancel
	}
	ctx, cancel := context.WithCancel(parent)
	doneCh := make(chan struct{})

	c.mu.Lock()
	c.nextTok++
	entry := &cancelEntry{token: c.nextTok, cancel: cancel, done: doneCh}
	if prev, ok := c.entries[key]; ok {
		prev.cancel()
	}
	c.entries[key] = entry
	c.mu.Unlock()

	release := func() {
		c.mu.Lock()
		if cur, ok := c.entries[key]; ok && cur.token == entry.token {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		cancel()
		select {
		case <-doneCh:
		default:
			close(doneCh)
		}
	}
	return ctx, release
}

func (c *Cancellations) Cancel(key string) bool {
	if c == nil || key == "" {
		return false
	}
	c.mu.Lock()
	entry, ok := c.entries[key]
	c.mu.Unlock()
	if !ok {
		return false
	}
	entry.cancel()
	select {
	case <-entry.done:
	case <-time.After(cancelWaitTimeout):
	}
	return true
}

func (c *Cancellations) Has(key string) bool {
	if c == nil || key == "" {
		return false
	}
	c.mu.Lock()
	_, ok := c.entries[key]
	c.mu.Unlock()
	return ok
}
