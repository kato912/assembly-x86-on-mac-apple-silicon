package i18n

import (
	"regexp"
	"testing"
)

var verbRE = regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z%]`)

// Every Thai message must use the same format verbs, in the same order,
// as its English original.
func TestThaiVerbsMatch(t *testing.T) {
	for key, m := range catalog {
		if m.th == "" {
			continue
		}
		en := verbRE.FindAllString(m.en, -1)
		th := verbRE.FindAllString(m.th, -1)
		if len(en) != len(th) {
			t.Errorf("%s: en verbs %v, th verbs %v", key, en, th)
			continue
		}
		for i := range en {
			if en[i] != th[i] {
				t.Errorf("%s: verb %d differs: %s vs %s", key, i, en[i], th[i])
			}
		}
	}
}

func TestAllTranslated(t *testing.T) {
	for key, m := range catalog {
		if m.th == "" && key != "err.prefix" {
			t.Errorf("%s has no Thai translation", key)
		}
	}
}

func TestLang(t *testing.T) {
	defer func() { Lang = "en" }()
	Lang = "th"
	if got := T("loc.line", 3, "nop"); got != "บรรทัด 3: nop" {
		t.Errorf("got %q", got)
	}
	Lang = "en"
	if got := T("loc.line", 3, "nop"); got != "line 3: nop" {
		t.Errorf("got %q", got)
	}
}
