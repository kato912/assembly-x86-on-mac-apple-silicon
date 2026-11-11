package nasm

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	bitsRE   = regexp.MustCompile(`^\[?\s*(?:bits\s+(16|32|64)|use(16|32|64))\b`)
	reg64RE  = regexp.MustCompile(`\b(?:r[abcd]x|r[sd]i|r[sb]p|r(?:8|9|1[0-5])[dwb]?|[sd]il|[sb]pl)\b`)
	sysRE    = regexp.MustCompile(`\bsyscall\b`)
	int80RE  = regexp.MustCompile(`\bint\s+(?:0x80|80h|0x80h|128|0h80|\$80)\b`)
	defRelRE = regexp.MustCompile(`^default\s+rel\b`)
)

// DetectBits guesses whether a NASM source file is 32- or 64-bit code.
// It returns the mode and a short human-readable reason.
//
// Order of evidence: an explicit BITS directive, use of 64-bit registers or
// `default rel`, then `syscall` (64) vs `int 0x80` (32). With no evidence at
// all it defaults to 64-bit.
func DetectBits(src []byte) (bits int, reason string) {
	var e evidence
	e.scan(src, "", nil, 0)
	return e.decide()
}

// DetectBitsFile is DetectBits for a file on disk; it also looks inside
// files pulled in with %include (searched next to the source and in dirs).
func DetectBitsFile(path string, src []byte, dirs []string) (bits int, reason string) {
	var e evidence
	e.scan(src, filepath.Dir(path), dirs, 0)
	return e.decide()
}

type evidence struct {
	bits                  int
	bitsWhy               string
	reg64, syscall, int80 string
	seen                  map[string]bool
}

func (e *evidence) decide() (int, string) {
	switch {
	case e.bits != 0:
		return e.bits, e.bitsWhy
	case e.reg64 != "":
		return 64, e.reg64
	case e.syscall != "":
		return 64, e.syscall
	case e.int80 != "":
		return 32, e.int80
	}
	return 64, "default"
}

func (e *evidence) scan(src []byte, dir string, dirs []string, depth int) {
	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if e.bits != 0 {
			return
		}
		raw := stripComment(sc.Text())
		t := strings.TrimSpace(strings.ToLower(raw))
		if t == "" {
			continue
		}
		if m := bitsRE.FindStringSubmatch(t); m != nil {
			e.bits, e.bitsWhy = map[string]int{"16": 16, "32": 32, "64": 64}[m[1]+m[2]], "bits "+m[1]+m[2]
			return
		}
		if m := includeRE.FindStringSubmatch(strings.TrimSpace(raw)); m != nil && depth < 4 && dir != "" {
			e.include(strings.TrimSpace(m[1]), dir, dirs, depth)
			continue
		}
		code := stripStrings(t)
		if e.reg64 == "" {
			if defRelRE.MatchString(code) {
				e.reg64 = "default rel"
			} else if m := reg64RE.FindString(code); m != "" {
				e.reg64 = "uses register " + m
			}
		}
		if e.syscall == "" && sysRE.MatchString(code) {
			e.syscall = "uses syscall"
		}
		if e.int80 == "" && int80RE.MatchString(code) {
			e.int80 = "uses int 0x80"
		}
	}
}

func (e *evidence) include(name, dir string, dirs []string, depth int) {
	for _, d := range append([]string{dir}, dirs...) {
		p := filepath.Join(d, name)
		if e.seen[p] {
			return
		}
		if b, err := os.ReadFile(p); err == nil {
			if e.seen == nil {
				e.seen = map[string]bool{}
			}
			e.seen[p] = true
			e.scan(b, filepath.Dir(p), dirs, depth+1)
			return
		}
	}
}

// stripComment removes a ';' comment, respecting quoted strings.
func stripComment(s string) string {
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case q != 0:
			if c == q {
				q = 0
			}
		case c == '"' || c == '\'' || c == '`':
			q = c
		case c == ';':
			return s[:i]
		}
	}
	return s
}

// stripStrings blanks out quoted strings so their contents are not mistaken
// for instructions or registers.
func stripStrings(s string) string {
	b := []byte(s)
	var q byte
	for i, c := range b {
		switch {
		case q != 0:
			if c == q {
				q = 0
			} else {
				b[i] = ' '
			}
		case c == '"' || c == '\'' || c == '`':
			q = c
		}
	}
	return string(b)
}
