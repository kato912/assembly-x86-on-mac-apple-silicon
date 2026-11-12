// Package pathfind finds programs on PATH while skipping nasmx86 itself.
// nasmx86 stands in for Linux's ld, gcc, ddd and gdb under those names (a
// directory of links to it goes first on PATH), so looking up "ld" or
// "ddd" must not find nasmx86's own link again.
package pathfind

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Self returns the path of the running nasmx86 executable, with symbolic
// links resolved, or "".
func Self() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		return real
	}
	return exe
}

// IsSelf reports whether path is (a link to) the running nasmx86.
func IsSelf(path string) bool {
	self := Self()
	if self == "" {
		return false
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	if real == self {
		return true
	}
	a, err1 := os.Stat(real)
	b, err2 := os.Stat(self)
	return err1 == nil && err2 == nil && os.SameFile(a, b)
}

// Other looks for an executable called name on PATH, then in dirs,
// skipping nasmx86 itself.
func Other(name string, dirs ...string) (string, bool) {
	for _, dir := range append(filepath.SplitList(os.Getenv("PATH")), dirs...) {
		if dir == "" {
			dir = "."
		}
		p := filepath.Join(dir, name)
		if !isExecutable(p) || IsSelf(p) {
			continue
		}
		return p, true
	}
	return "", false
}

func isExecutable(p string) bool {
	if _, err := exec.LookPath(p); err != nil {
		return false
	}
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// WithoutDir returns a PATH value without dir.
func WithoutDir(path, dir string) string {
	var keep []string
	for _, d := range filepath.SplitList(path) {
		if filepath.Clean(d) != filepath.Clean(dir) {
			keep = append(keep, d)
		}
	}
	return strings.Join(keep, string(filepath.ListSeparator))
}
