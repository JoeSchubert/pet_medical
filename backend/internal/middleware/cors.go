package middleware

import (
	"net/http"
	"strings"
)

func CORS(origins string) func(http.Handler) http.Handler {
	allowed := strings.Split(origins, ",")
	for i := range allowed {
		allowed[i] = strings.TrimSpace(allowed[i])
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowOrigin := ""
			if origin != "" {
				for _, o := range allowed {
					if o == "*" || o == origin {
						allowOrigin = o
						break
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
