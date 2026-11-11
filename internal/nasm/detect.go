package nasm

import (
	"bufio"
	"bytes"
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
	var saw64reg, sawSyscall, saw80 string
	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.ToLower(stripComment(sc.Text()))
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if m := bitsRE.FindStringSubmatch(t); m != nil {
			v := m[1] + m[2]
			switch v {
			case "64":
				return 64, "bits 64"
			case "32":
				return 32, "bits 32"
			default:
				return 16, "bits 16"
			}
		}
		code := stripStrings(t)
		if saw64reg == "" {
			if defRelRE.MatchString(code) {
				saw64reg = "default rel"
			} else if m := reg64RE.FindString(code); m != "" {
				saw64reg = "uses register " + m
			}
		}
		if sawSyscall == "" && sysRE.MatchString(code) {
			sawSyscall = "uses syscall"
		}
		if saw80 == "" && int80RE.MatchString(code) {
			saw80 = "uses int 0x80"
		}
	}
	switch {
	case saw64reg != "":
		return 64, saw64reg
	case sawSyscall != "":
		return 64, sawSyscall
	case saw80 != "":
		return 32, saw80
	}
	return 64, "default"
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
