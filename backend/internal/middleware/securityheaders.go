package middleware

import (
	"net/http"
	"strconv"
)

// SecurityHeaders adds common security headers to responses.
// Set hstsMaxAge > 0 to enable HSTS (only use when serving over HTTPS).
func SecurityHeaders(hstsMaxAge int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			if hstsMaxAge > 0 {
				w.Header().Set("Strict-Transport-Security", "max-age="+strconv.Itoa(hstsMaxAge))
			}
			next.ServeHTTP(w, r)
		})
	}
}
