package main

// nasmx86 setup and nasmx86 shellenv put nasmx86's ld, gcc, ddd and gdb
// first on PATH in every new terminal, so the commands of a Linux course
// work on a Mac as typed. brew install runs nasmx86 setup and brew
// uninstall runs nasmx86 setup --remove.

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/pathfind"
)

const (
	rcBegin = "# >>> nasmx86 >>>"
	rcEnd   = "# <<< nasmx86 <<<"
)

// rcBlock is what nasmx86 setup adds to the shell's startup file.
const rcBlock = rcBegin + `
# Linux-style ld, gcc, ddd and gdb for x86 programs (nasmx86 runs them).
# Added by nasmx86 setup; nasmx86 setup --remove (or brew uninstall nasmx86) takes it out.
if command -v nasmx86 >/dev/null 2>&1; then eval "$(nasmx86 shellenv)"; fi
` + rcEnd + "\n"

// toolsDirName is the directory of links named ld, gcc, ddd and gdb that
// point to nasmx86.
const toolsDirName = "nasmx86-tools"

// setupSupported reports whether nasmx86 should stand in for ld and gcc
// here: on a Mac. Linux has the real tools, and runs the programs itself.
func setupSupported() bool { return runtime.GOOS == "darwin" }

// homeDir is the user's home directory. It asks the system rather than
// trusting $HOME: Homebrew runs install steps with HOME set to a temporary
// directory. NASMX86_HOME overrides it (for tests).
func homeDir() (string, error) {
	if h := os.Getenv("NASMX86_HOME"); h != "" {
		return h, nil
	}
	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		return u.HomeDir, nil
	}
	return os.UserHomeDir()
}

// rcFiles are the startup files nasmx86 setup changes: ~/.zshrc (zsh is
// the Mac's shell), and ~/.bash_profile for bash users.
func rcFiles(all bool) []string {
	home, err := homeDir()
	if err != nil {
		return nil
	}
	files := []string{filepath.Join(home, ".zshrc")}
	if all || filepath.Base(os.Getenv("SHELL")) == "bash" {
		files = append(files, filepath.Join(home, ".bash_profile"))
	}
	return files
}

// withoutBlock removes nasmx86's block from a startup file's text.
func withoutBlock(text string) string {
	for {
		start := strings.Index(text, rcBegin)
		if start < 0 {
			return text
		}
		end := strings.Index(text[start:], rcEnd)
		if end < 0 {
			return text
		}
		end += start + len(rcEnd)
		if end < len(text) && text[end] == '\n' {
			end++
		}
		// Also drop the blank line setup put before the block.
		if start > 0 && strings.HasSuffix(text[:start], "\n\n") {
			start--
		}
		text = text[:start] + text[end:]
	}
}

// withBlock adds (or refreshes) nasmx86's block at the end of a startup file.
func withBlock(text string) string {
	text = withoutBlock(text)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if text != "" {
		text += "\n"
	}
	return text + rcBlock
}

// updateFile rewrites path in place (Homebrew's sandbox allows writing the
// file but not creating others next to it); a missing file starts empty.
func updateFile(path string, change func(string) string) (bool, error) {
	old, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	text := change(string(old))
	if text == string(old) {
		return false, nil
	}
	if err != nil && text == "" {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(text), 0o644)
}

func setupMain(args []string) int {
	remove := false
	for _, a := range args {
		switch a {
		case "--remove", "--uninstall", "-r":
			remove = true
		case "--th", "--thai":
			i18n.Lang = "th"
		case "-h", "--help":
			fmt.Print(i18n.T("setup.usage"))
			return 0
		default:
			fmt.Fprintln(os.Stderr, "nasmx86 setup: "+i18n.T("err.badOption", a))
			return 2
		}
	}
	if remove {
		code := 0
		for _, f := range rcFiles(true) {
			changed, err := updateFile(f, withoutBlock)
			if err != nil {
				fmt.Fprintln(os.Stderr, "nasmx86: "+i18n.T("setup.cannotWrite", f, err))
				code = 1
			} else if changed {
				fmt.Println("nasmx86: " + i18n.T("setup.removed", tilde(f)))
			}
		}
		for _, d := range toolsDirs() {
			removeTools(d)
		}
		return code
	}
	if !setupSupported() {
		fmt.Println("nasmx86: " + i18n.T("setup.notNeeded"))
		return 0
	}
	code := 0
	for _, f := range rcFiles(false) {
		if _, err := updateFile(f, withBlock); err != nil {
			fmt.Fprintln(os.Stderr, "nasmx86: "+i18n.T("setup.cannotWrite", f, err))
			code = 1
			continue
		}
		fmt.Println("nasmx86: " + i18n.T("setup.added", tilde(f)))
	}
	ensureTools() // brew's sandbox may not allow it; shellenv does it then
	return code
}

func tilde(path string) string {
	if home, err := homeDir(); err == nil && strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}
	return path
}

// shellenvMain prints the shell command that puts nasmx86's tools first
// on PATH. It runs in every new terminal, so it stays quiet on errors.
func shellenvMain(args []string) int {
	if !setupSupported() {
		return 0
	}
	dir, err := ensureTools()
	if err != nil {
		return 0
	}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(d) == dir {
			return 0
		}
	}
	fmt.Printf("export PATH='%s':\"$PATH\"\n", strings.ReplaceAll(dir, "'", `'\''`))
	return 0
}

// toolsDirs are where the links may go: next to nasmx86 (Homebrew
// removes them with it), else in the home directory.
func toolsDirs() []string {
	var dirs []string
	if self := pathfind.Self(); self != "" {
		dirs = append(dirs, filepath.Join(filepath.Dir(self), toolsDirName))
	}
	if home, err := homeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".nasmx86", "bin"))
	}
	return dirs
}

// ensureTools makes the directory of links to nasmx86 and returns it.
func ensureTools() (string, error) {
	self := pathfind.Self()
	if self == "" {
		return "", errors.New("cannot find nasmx86 itself")
	}
	var lastErr error
	for _, dir := range toolsDirs() {
		if err := makeTools(dir, self); err != nil {
			lastErr = err
			continue
		}
		return dir, nil
	}
	return "", lastErr
}

func makeTools(dir, self string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, name := range toolNames {
		link := filepath.Join(dir, name)
		if target, err := os.Readlink(link); err == nil && target == self {
			continue
		}
		os.Remove(link)
		if err := os.Symlink(self, link); err != nil {
			return err
		}
	}
	return nil
}

func removeTools(dir string) {
	for _, name := range toolNames {
		link := filepath.Join(dir, name)
		if pathfind.IsSelf(link) || isDangling(link) {
			os.Remove(link)
		}
	}
	os.Remove(dir) // only if now empty
	if filepath.Base(dir) == "bin" && filepath.Base(filepath.Dir(dir)) == ".nasmx86" {
		os.Remove(filepath.Dir(dir))
	}
}

func isDangling(link string) bool {
	if _, err := os.Lstat(link); err != nil {
		return false
	}
	_, err := os.Stat(link)
	return err != nil
}
