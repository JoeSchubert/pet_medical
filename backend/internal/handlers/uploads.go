package handlers

import (
	"net/http"
	"path/filepath"
	"strings"
)

// ServeUploads serves files from baseDir. URL path after prefix (e.g. /api/uploads/) is
// interpreted as relative path under baseDir. Path traversal is rejected.
func ServeUploads(baseDir, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, prefix)
		path = strings.TrimPrefix(path, "/")
		if path == "" || strings.Contains(path, "..") {
			http.NotFound(w, r)
			return
		}
		abs := filepath.Join(baseDir, filepath.FromSlash(path))
		clean := filepath.Clean(abs)
		baseAbs := filepath.Clean(baseDir)
		rel, err := filepath.Rel(baseAbs, clean)
		if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, clean)
	})
}
