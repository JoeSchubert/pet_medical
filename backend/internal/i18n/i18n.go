package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	mu     sync.RWMutex
	locale string
	msgs   map[string]string
)

// Init loads messages for the given language code (e.g. "en"). Falls back to "en" if file missing.
func Init(lang string) {
	mu.Lock()
	defer mu.Unlock()
	locale = lang
	msgs = loadLocale(lang)
	if msgs == nil && lang != "en" {
		msgs = loadLocale("en")
	}
	if msgs == nil {
		msgs = make(map[string]string)
	}
}

func loadLocale(lang string) map[string]string {
	data, err := localesFS.ReadFile("locales/" + lang + ".json")
	if err != nil {
		return nil
	}
	var out map[string]string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

// T returns the translation for key. If not found, returns key.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if s, ok := msgs[key]; ok && s != "" {
		return s
	}
	return key
}

// Tf returns T(key) formatted with fmt.Sprintf. Use %s, %d, %v etc. in the locale string.
func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}
