package middleware

import (
	"net/http"
	"strings"

	"github.com/pet-medical/api/internal/config"
)

// CORS sets Access-Control-Allow-* when the request Origin is allowed.
// When origins is empty and cfg is non-nil, only the request's effective origin is allowed (same-origin when behind a proxy; no CORS_ORIGINS env needed).
// When origins is set, it is a comma-separated list of allowed origins, or "*" for any.
func CORS(origins string, cfg *config.Config) func(http.Handler) http.Handler {
	var allowed []string
	if origins != "" {
		allowed = strings.Split(origins, ",")
		for i := range allowed {
			allowed[i] = strings.TrimSpace(allowed[i])
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowOrigin := ""
			if origin != "" {
				if len(allowed) == 0 && cfg != nil {
					if cfg.RequestOrigin(r) == origin {
						allowOrigin = origin
					}
				} else {
					for _, o := range allowed {
						if o == "*" || o == origin {
							allowOrigin = o
							break
						}
					}
				}
			}
			if allowOrigin == "" && len(allowed) > 0 && allowed[0] == "*" {
				allowOrigin = "*"
			}
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
