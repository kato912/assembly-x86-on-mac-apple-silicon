package runner

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
)

// runFiles writes the named sources into a temp dir and runs them together
// (the first one is the main file).
func runFiles(t *testing.T, files ...[2]string) (string, string, int) {
	t.Helper()
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	var paths []string
	for _, f := range files {
		p := filepath.Join(dir, f[0])
		if err := os.WriteFile(p, []byte(f[1]), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	var out, errb bytes.Buffer
	code := Run(Config{Source: paths[0], Extra: paths[1:], MaxSteps: DefaultMaxSteps,
		Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb})
	return out.String(), errb.String(), code
}

const mainSrc = `
extern print, greeting, glen
global _start
section .text
_start:
    mov ecx, greeting
    mov edx, [glen]
    call print
    mov eax, 1
    mov ebx, 0
    int 0x80
`

const ioSrc = `
global print, greeting, glen
section .data
greeting db "hello from io.asm", 10
glen     dd $ - greeting
section .text
print:                  ; write(1, ecx, edx)
    mov eax, 4
    mov ebx, 1
    int 0x80
    ret
`

func TestMultiFile(t *testing.T) {
	out, errs, code := runFiles(t, [2]string{"main.asm", mainSrc}, [2]string{"io.asm", ioSrc})
	if out != "hello from io.asm\n" || code != 0 || errs != "" {
		t.Errorf("out=%q err=%q code=%d", out, errs, code)
	}
}

func TestMultiFileErrors(t *testing.T) {
	// A crash inside the second file names that file.
	_, errs, code := runFiles(t, [2]string{"main.asm", `
extern boom
global _start
section .text
_start:
    call boom
`}, [2]string{"util.asm", `
global boom
section .text
boom:
    xor esi, esi
    mov eax, [esi]
    ret
`})
	if code != 139 || !strings.Contains(errs, "util.asm line 6: mov eax, [esi]") {
		t.Errorf("code=%d\n%s", code, errs)
	}

	// Forgetting "global" in the other file gets a specific hint.
	_, errs, code = runFiles(t, [2]string{"main.asm", "extern helper\nglobal _start\nsection .text\n_start: call helper\n"},
		[2]string{"util.asm", "section .text\nhelper: ret\n"})
	if code != ExitBuild || !strings.Contains(errs, "global helper") {
		t.Errorf("missing global: code=%d\n%s", code, errs)
	}

	// The same global label in two files.
	_, errs, code = runFiles(t, [2]string{"a.asm", "global _start, f\nsection .text\n_start:\nf: ret\n"},
		[2]string{"b.asm", "global f\nsection .text\nf: ret\n"})
	if code != ExitBuild || !strings.Contains(errs, "f") || !strings.Contains(errs, "a.asm, b.asm") {
		t.Errorf("duplicate: code=%d\n%s", code, errs)
	}
}

func TestObjectFileInput(t *testing.T) {
	nasmPath, err := nasm.Find()
	if err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	ioPath := filepath.Join(dir, "io.asm")
	os.WriteFile(ioPath, []byte(ioSrc), 0o644)
	res, err := nasm.Assemble(nasmPath, ioPath, 32, nil, dir, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(dir, "main.asm")
	os.WriteFile(mainPath, []byte(mainSrc), 0o644)
	var out, errb bytes.Buffer
	code := Run(Config{Source: mainPath, Extra: []string{res.Object}, MaxSteps: DefaultMaxSteps,
		Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb})
	if out.String() != "hello from io.asm\n" || code != 0 {
		t.Errorf("out=%q code=%d err=%s", out.String(), code, errb.String())
	}
}
