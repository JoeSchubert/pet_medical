package middleware

import (
	"net/http"
)

// AdminRequired requires AuthRequired to have run first (user in context). Returns 403 if not admin.
func AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := GetUser(r.Context())
		if u == nil || u.Role != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
