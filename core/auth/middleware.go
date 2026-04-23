package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"mantis/core/protocols"
	"mantis/core/types"
)

const (
	SessionCookie = "mantis_session"
	HeaderName    = "Authorization"
	HeaderPrefix  = "Bearer "
)

func Middleware(store protocols.Store[string, types.User], skip func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skip != nil && skip(r) {
				next.ServeHTTP(w, r)
				return
			}
			token := extractToken(r)
			if token == "" {
				respondUnauthorized(w)
				return
			}
			users, err := store.List(r.Context(), types.ListQuery{
				Filter: map[string]string{"api_key_hash": HashToken(token)},
				Page:   types.Page{Limit: 1},
			})
			if err != nil || len(users) == 0 {
				respondUnauthorized(w)
				return
			}
			ctx := WithIdentity(r.Context(), Identity{UserID: users[0].ID, Name: users[0].Name})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get(HeaderName); strings.HasPrefix(h, HeaderPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(h, HeaderPrefix))
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value
	}
	return ""
}

func respondUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"title":  "Unauthorized",
		"detail": "missing or invalid credentials",
	})
}
