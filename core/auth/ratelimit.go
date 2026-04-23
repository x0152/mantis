package auth

import (
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LoginRateLimiter struct {
	mu      sync.Mutex
	entries map[string][]time.Time
	max     int
	window  time.Duration
}

func NewLoginRateLimiter(max int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		entries: make(map[string][]time.Time),
		max:     max,
		window:  window,
	}
}

func (l *LoginRateLimiter) Allow(key string, now time.Time) (wait time.Duration, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempts := l.prune(key, now)
	if len(attempts) >= l.max {
		return l.window - now.Sub(attempts[0]), false
	}
	return 0, true
}

func (l *LoginRateLimiter) Record(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempts := l.prune(key, now)
	l.entries[key] = append(attempts, now)
}

func (l *LoginRateLimiter) prune(key string, now time.Time) []time.Time {
	attempts := l.entries[key]
	cutoff := now.Add(-l.window)
	kept := attempts[:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.entries, key)
		return nil
	}
	l.entries[key] = kept
	return kept
}

func (l *LoginRateLimiter) Middleware(paths ...string) func(http.Handler) http.Handler {
	watched := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		watched[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := watched[r.URL.Path]; !ok {
				next.ServeHTTP(w, r)
				return
			}
			key := clientIP(r)
			now := time.Now()
			if wait, allowed := l.Allow(key, now); !allowed {
				respondTooManyRequests(w, wait)
				return
			}
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if rec.status == http.StatusUnauthorized {
				l.Record(key, now)
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func respondTooManyRequests(w http.ResponseWriter, wait time.Duration) {
	seconds := int(math.Ceil(wait.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"title":  "Too Many Requests",
		"detail": "too many failed attempts, try again later",
	})
}
