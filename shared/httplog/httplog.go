package httplog

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"time"
)

const maxLoggedBody = 4096

type captureWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
	wrote  bool
}

func (c *captureWriter) WriteHeader(code int) {
	c.status = code
	c.wrote = true
	c.ResponseWriter.WriteHeader(code)
}

func (c *captureWriter) Write(b []byte) (int, error) {
	if !c.wrote {
		c.status = http.StatusOK
		c.wrote = true
	}
	if c.status >= 400 && c.body.Len() < maxLoggedBody {
		remaining := maxLoggedBody - c.body.Len()
		if remaining > len(b) {
			remaining = len(b)
		}
		c.body.Write(b[:remaining])
	}
	return c.ResponseWriter.Write(b)
}

func (c *captureWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		cw := &captureWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(cw, r)
		dur := time.Since(start)

		if cw.status >= 500 {
			log.Printf("http %s %s -> %d (%s) ERROR: %s",
				r.Method, r.URL.Path, cw.status, dur, summarizeBody(cw.body.Bytes()))
			return
		}
		if cw.status >= 400 {
			log.Printf("http %s %s -> %d (%s) %s",
				r.Method, r.URL.Path, cw.status, dur, summarizeBody(cw.body.Bytes()))
			return
		}
	})
}

func summarizeBody(b []byte) string {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return ""
	}
	const maxLen = 1500
	if len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
