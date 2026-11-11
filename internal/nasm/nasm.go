// Package nasm locates and runs the NASM assembler and inspects its output.
package nasm

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
)

// ErrNotFound is returned by Find when nasm is not installed.
var ErrNotFound = errors.New("nasm not found")

// Find returns the path of the nasm executable. Besides $PATH it checks the
// usual install locations, because Homebrew's bin dir or the Windows NASM
// installer's folder are often missing from PATH.
func Find() (string, error) {
	if p := os.Getenv("NASMX86_NASM"); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", ErrNotFound
		}
		return p, nil
	}
	if p, err := exec.LookPath("nasm"); err == nil {
		return p, nil
	}
	var cands []string
	switch runtime.GOOS {
	case "darwin":
		cands = []string{"/opt/homebrew/bin/nasm", "/usr/local/bin/nasm"}
	case "linux":
		cands = []string{"/usr/bin/nasm", "/usr/local/bin/nasm", "/home/linuxbrew/.linuxbrew/bin/nasm"}
	case "windows":
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LOCALAPPDATA"} {
			if d := os.Getenv(env); d != "" {
				cands = append(cands, filepath.Join(d, "NASM", "nasm.exe"), filepath.Join(d, "bin", "NASM", "nasm.exe"))
			}
		}
		if exe, err := os.Executable(); err == nil {
			cands = append(cands, filepath.Join(filepath.Dir(exe), "nasm.exe"))
		}
	}
	for _, c := range cands {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", ErrNotFound
}

var versionRE = regexp.MustCompile(`version\s+(\d+)\.(\d+)`)

// Version returns nasm's major and minor version, or 0, 0 if unknown.
func Version(nasmPath string) (major, minor int) {
	out, err := exec.Command(nasmPath, "-v").Output()
	if err != nil {
		return 0, 0
	}
	m := versionRE.FindSubmatch(out)
	if m == nil {
		return 0, 0
	}
	major, _ = strconv.Atoi(string(m[1]))
	minor, _ = strconv.Atoi(string(m[2]))
	return major, minor
}

// Result holds the files produced by Assemble.
type Result struct {
	Object  string
	Listing string
}

// ExitError reports that nasm ran but failed (its messages were already
// shown to the user).
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("nasm failed with exit code %d", e.Code) }

// Assemble runs nasm on src producing an ELF object (elf32 or elf64) and a
// listing in dir. nasm's own diagnostics are copied to diag unchanged.
func Assemble(nasmPath, src string, bits int, extra []string, dir string, diag io.Writer) (*Result, error) {
	format := "elf64"
	if bits == 32 {
		format = "elf32"
	}
	res := &Result{
		Object:  filepath.Join(dir, "prog.o"),
		Listing: filepath.Join(dir, "prog.lst"),
	}
	args := []string{"-f", format, "-g", "-l", res.Listing, "-o", res.Object}
	// Let %include find files next to the source file.
	if d := filepath.Dir(src); d != "" {
		args = append(args, "-I"+d+string(filepath.Separator))
	}
	if major, _ := Version(nasmPath); major >= 3 && !hasWarningFlag(extra) {
		// NASM 3 warns about every absolute address in 64-bit code; the
		// NASM 2.x used on course machines does not, so keep output identical.
		args = append(args, "-w-implicit-abs-deprecated")
	}
	args = append(args, extra...)
	args = append(args, src)

	cmd := exec.Command(nasmPath, args...)
	cmd.Stdout = diag
	cmd.Stderr = diag
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, &ExitError{Code: ee.ExitCode()}
		}
		return nil, err
	}
	return res, nil
}

func hasWarningFlag(args []string) bool {
	for _, a := range args {
		if len(a) >= 2 && (a[:2] == "-w" || a[:2] == "-W") {
			return true
		}
	}
	return false
}
