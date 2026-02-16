package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/pet-medical/api/internal/debuglog"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		debuglog.Debugf("request: %s %s", r.Method, r.URL.Path)
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		dur := time.Since(start)
		log.Printf("[HTTP] %s %s %d %d %s", r.Method, r.URL.Path, wrapped.status, wrapped.bytes, dur)
		debuglog.Debugf("response: %s %s -> %d %d bytes in %s", r.Method, r.URL.Path, wrapped.status, wrapped.bytes, dur)
	})
}
