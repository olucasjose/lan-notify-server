package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

var translations map[string]string

func init() {
	translations = make(map[string]string)

	// Determine Language
	lang := os.Getenv("LC_ALL")
	if lang == "" {
		lang = os.Getenv("LANG")
	}

	locale := "en" // Default
	if strings.HasPrefix(strings.ToLower(lang), "pt") {
		locale = "pt-BR"
	}

	// Load file
	data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", locale))
	if err != nil {
		// Fallback to en if specific locale not found
		data, _ = localesFS.ReadFile("locales/en.json")
	}

	if err := json.Unmarshal(data, &translations); err != nil {
		panic(fmt.Sprintf("failed to parse i18n file: %v", err))
	}
}

// T returns the translated string for the given key.
// It formats the string using fmt.Sprintf if args are provided.
func T(key string, args ...interface{}) string {
	val, ok := translations[key]
	if !ok {
		return key // fallback to key itself if not found
	}
	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}
