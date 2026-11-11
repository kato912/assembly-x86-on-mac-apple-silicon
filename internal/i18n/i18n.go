// Package i18n holds nasmx86's user-facing messages in English and Thai.
package i18n

import (
	"fmt"
	"os"
	"strings"
)

// Lang is the active language: "en" (default) or "th".
var Lang = "en"

// Init picks the language from the NASMX86_LANG environment variable.
func Init() {
	if l := strings.ToLower(os.Getenv("NASMX86_LANG")); strings.HasPrefix(l, "th") {
		Lang = "th"
	}
}

type msg struct{ en, th string }

// T formats the message with the given key in the active language.
// Unknown keys are returned verbatim so a missing entry is obvious.
func T(key string, args ...any) string {
	m, ok := catalog[key]
	if !ok {
		if len(args) > 0 {
			return key + fmt.Sprint(args...)
		}
		return key
	}
	s := m.en
	if Lang == "th" && m.th != "" {
		s = m.th
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}
