package middleware

import (
	"net/http"
	"strconv"

	"github.com/pet-medical/api/internal/config"
)

// SecurityHeaders adds common security headers to responses.
// When hstsMaxAge > 0 and cfg is non-nil, HSTS is set only when the request is HTTPS (direct TLS or X-Forwarded-Proto from a trusted proxy), per RFC 6797.
func SecurityHeaders(hstsMaxAge int, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			if hstsMaxAge > 0 && cfg != nil && cfg.IsRequestHTTPS(r) {
				w.Header().Set("Strict-Transport-Security", "max-age="+strconv.Itoa(hstsMaxAge))
			}
			next.ServeHTTP(w, r)
		})
	}
}
