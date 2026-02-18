package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pet-medical/api/internal/config"
)

// ThrottleByPath returns a middleware that rate-limits by client IP with different limits per path:
// - /api/auth/login: strict (e.g. 5/min) to prevent brute force
// - /api/auth/refresh, /api/auth/logout: moderate (e.g. 20/min)
// - /api/* (rest): general (e.g. 120/min)
// Uses config for client IP (X-Forwarded-For when from trusted proxy). Returns 429 with Retry-After when exceeded.
func ThrottleByPath(cfg *config.Config, authLoginPerMin, authOtherPerMin, apiPerMin int) func(http.Handler) http.Handler {
	if authLoginPerMin <= 0 {
		authLoginPerMin = 5
	}
	if authOtherPerMin <= 0 {
		authOtherPerMin = 20
	}
	if apiPerMin <= 0 {
		apiPerMin = 120
	}
	login := newThrottler(authLoginPerMin)
	authOther := newThrottler(authOtherPerMin)
	api := newThrottler(apiPerMin)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := cfg.ClientIP(r)
			path := r.URL.Path

			var t *throttler
			switch {
			case path == "/api/auth/login":
				t = login
			case path == "/api/auth/refresh" || path == "/api/auth/logout" || path == "/api/auth/change-password":
				t = authOther
			case len(path) > 4 && path[:4] == "/api":
				t = api
			default:
				next.ServeHTTP(w, r)
				return
			}

			key := path + ":" + ip
			ok, retryAfter := t.allow(key)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", retryAfter)
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"error.too_many_requests"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// throttler limits requests per key to limitPerMin per minute (sliding window).
type throttler struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	requests map[string][]time.Time
}

func newThrottler(perMin int) *throttler {
	return &throttler{
		limit:    perMin,
		window:   time.Minute,
		requests: make(map[string][]time.Time),
	}
}

func (t *throttler) allow(key string) (ok bool, retryAfter string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-t.window)
	slice := t.requests[key]
	// Prune old entries
	var kept []time.Time
	for _, ts := range slice {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= t.limit {
		// Oldest in window - suggest retry after it expires
		oldest := kept[0]
		sec := int(oldest.Add(t.window).Sub(now).Seconds())
		if sec < 1 {
			sec = 1
		}
		t.requests[key] = kept
		return false, strconv.Itoa(sec)
	}
	kept = append(kept, now)
	t.requests[key] = kept
	return true, ""
}
