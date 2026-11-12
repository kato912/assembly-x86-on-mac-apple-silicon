//go:build linux

package difftest

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
)

func nativeAvailable() (bool, string) {
	if os.Getenv("NASMX86_NATIVE") == "0" {
		return false, "NASMX86_NATIVE=0"
	}
	if runtime.GOARCH != "amd64" {
		return false, "not an x86-64 machine"
	}
	if _, err := exec.LookPath("ld"); err != nil {
		return false, "ld (binutils) not installed"
	}
	return true, ""
}

// runNative assembles with the same nasm options nasmx86 uses, links with
// GNU ld and runs the program with an 8 MiB stack limit.
func runNative(t *testing.T, p program) ([]byte, int, error) {
	nasmPath, err := nasm.Find()
	if err != nil {
		return nil, 0, err
	}
	bits := p.bits
	if bits == 0 {
		src, _ := os.ReadFile(p.path)
		bits, _ = nasm.DetectBitsFile(p.path, src, []string{p.incDir})
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "prog")
	ldArgs := []string{"-o", exe}
	for i, src := range append([]string{p.path}, p.extra...) {
		sub := filepath.Join(dir, strconv.Itoa(i))
		if err := os.Mkdir(sub, 0o755); err != nil {
			return nil, 0, err
		}
		res, err := nasm.Assemble(nasmPath, src, bits, []string{"-I" + p.incDir + "/"}, sub, io.Discard)
		if err != nil {
			return nil, 0, err
		}
		ldArgs = append(ldArgs, res.Object)
	}
	linker := "ld"
	switch {
	case p.libc:
		// gcc links the C runtime (which calls main) and glibc.
		linker = "gcc"
		ldArgs = append([]string{"-no-pie", "-z", "noexecstack"}, ldArgs...)
		if bits == 32 {
			ldArgs = append([]string{"-m32"}, ldArgs...)
		}
	case bits == 32:
		ldArgs = append([]string{"-m", "elf_i386"}, ldArgs...)
	}
	if out, err := exec.Command(linker, ldArgs...).CombinedOutput(); err != nil {
		return nil, 0, errors.New(linker + ": " + string(out))
	}
	cmd := exec.Command("sh", "-c", `ulimit -s 8192 2>/dev/null; exec "$0"`, exe)
	cmd.Stdin = bytes.NewReader(p.stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		ws := ee.Sys().(syscall.WaitStatus)
		if ws.Signaled() {
			code = 128 + int(ws.Signal())
		} else {
			code = ws.ExitStatus()
		}
	} else if err != nil {
		return nil, 0, err
	}
	return out.Bytes(), code, nil
}
