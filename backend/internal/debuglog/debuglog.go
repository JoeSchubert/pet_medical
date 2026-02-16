package debuglog

import (
	"log"
	"sync"
)

var (
	enabled bool
	mu      sync.RWMutex
)

// SetEnabled sets whether debug logging is on. Call once at startup.
func SetEnabled(on bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = on
}

// Enabled returns whether debug logging is enabled.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// Debugf logs to the standard logger only when debug is enabled. Format and args are like log.Printf.
func Debugf(format string, args ...interface{}) {
	if !Enabled() {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}
