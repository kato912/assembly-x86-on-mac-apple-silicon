package debugger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

func build(t *testing.T, src, stdin string) (*runner.Program, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	path := filepath.Join(t.TempDir(), "p.asm")
	os.WriteFile(path, []byte(src), 0o644)
	var out, errb bytes.Buffer
	p, code := runner.Build(runner.Config{Source: path, MaxSteps: 10000,
		Stdin: strings.NewReader(stdin), Stdout: &out, Stderr: &errb})
	if p == nil {
		t.Fatalf("build failed (%d): %s", code, errb.String())
	}
	t.Cleanup(p.Close)
	return p, &out, &errb
}

const loopSrc = `
section .bss
buf resb 4
section .text
global _start
_start:
    xor eax, eax        ; sum = 0
    mov ecx, 2
again:
    add eax, ecx
    mov [buf], eax
    loop again
    mov ebx, eax
    mov eax, 1
    int 0x80
`

func TestTrace(t *testing.T) {
	p, _, _ := build(t, loopSrc, "")
	var tr bytes.Buffer
	err := Trace(NewSession(p), &tr)
	if code := p.Finish(err); code != 3 {
		t.Errorf("exit code %d", code)
	}
	lines := strings.Split(strings.TrimSpace(tr.String()), "\n")
	if len(lines) != 11 {
		t.Fatalf("%d trace lines:\n%s", len(lines), tr.String())
	}
	want := []string{
		"[1]    line 7: xor eax, eax",
		"PF=1 ZF=1",
		"[3]    line 10: add eax, ecx",
		"eax=0x00000002 (2)  PF=0 ZF=0",
		"[0x0804a000]=0x00000002",
		"ecx=0x00000001 (1)",
		"→ exit(3)",
	}
	for _, w := range want {
		if !strings.Contains(tr.String(), w) {
			t.Errorf("trace lacks %q:\n%s", w, tr.String())
		}
	}
}
