package runner

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/progfile"
)

// object assembles src into dir/name (with DWARF line information).
func object(t *testing.T, dir, name string, bits int, src string) string {
	t.Helper()
	nasmPath, err := nasm.Find()
	if err != nil {
		t.Skip("nasm not installed")
	}
	asm := filepath.Join(dir, strings.TrimSuffix(name, ".o")+".asm")
	if err := os.WriteFile(asm, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	res, err := nasm.Assemble(nasmPath, asm, bits, []string{"-F", "dwarf"}, work, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(res.Object)
	obj := filepath.Join(dir, name)
	os.WriteFile(obj, data, 0o644)
	return obj
}

func runObjects(cfg Config, objs ...string) (string, string, int) {
	var out, errb bytes.Buffer
	cfg.Source, cfg.Extra = objs[0], objs[1:]
	cfg.MaxSteps, cfg.Stdin, cfg.Stdout, cfg.Stderr = DefaultMaxSteps, strings.NewReader(""), &out, &errb
	code := Run(cfg)
	return out.String(), errb.String(), code
}

const start64 = `
section .data
msg db "from _start", 10
section .text
global _start
_start:
    mov rax, 1
    mov rdi, 1
    mov rsi, msg
    mov rdx, 12
    syscall
    mov rax, 60
    mov rdi, 4
    syscall
`

const main64 = `
extern printf
section .data
fmt db "%s", 10, 0
msg db "from main", 0
section .text
global main
main:
    push rbp
    mov rdi, fmt
    mov rsi, msg
    mov rax, 0
    call printf
    pop rbp
    mov rax, 6
    ret
`

func TestLinkLikeLD(t *testing.T) {
	dir := t.TempDir()
	start := object(t, dir, "start.o", 64, start64)
	if out, errs, code := runObjects(Config{Mode: LD, Tool: "ld"}, start); out != "from _start\n" || code != 4 || errs != "" {
		t.Errorf("ld: out=%q code=%d err=%s", out, code, errs)
	}

	// ld does not link the C library unless asked (-lc), and says to use gcc.
	mainObj := object(t, dir, "printf.o", 64, main64)
	_, errs, code := runObjects(Config{Mode: LD, Tool: "ld", Argv0: "cprintf"}, mainObj)
	if code != ExitBuild || !strings.Contains(errs, "ld: undefined reference to printf") ||
		!strings.Contains(errs, "gcc -o cprintf "+mainObj) {
		t.Errorf("ld without -lc: code=%d\n%s", code, errs)
	}

	// No _start: GNU ld warns and starts at the top of .text.
	noStart := object(t, dir, "nostart.o", 64, "section .text\nglobal go\ngo:\n    mov rax, 60\n    mov rdi, 9\n    syscall\n")
	_, errs, code = runObjects(Config{Mode: LD, Tool: "ld"}, noStart)
	if code != 9 || !strings.Contains(errs, "ld: warning: cannot find entry symbol _start; defaulting to 0x401000") {
		t.Errorf("no _start: code=%d\n%s", code, errs)
	}
	// ld -e go
	if _, errs, code := runObjects(Config{Mode: LD, Tool: "ld", Entry: "go"}, noStart); code != 9 || errs != "" {
		t.Errorf("-e go: code=%d\n%s", code, errs)
	}

	// A 32-bit object needs -m elf_i386.
	obj32 := object(t, dir, "h32.o", 32, "section .text\nglobal _start\n_start:\n    mov eax, 1\n    mov ebx, 2\n    int 0x80\n")
	_, errs, code = runObjects(Config{Mode: LD, Tool: "ld"}, obj32)
	if code != ExitBuild || !strings.Contains(errs, "h32.o is a 32-bit object") || !strings.Contains(errs, "ld -m elf_i386") {
		t.Errorf("32-bit object: code=%d\n%s", code, errs)
	}
	if _, errs, code := runObjects(Config{Mode: LD, Tool: "ld", Bits: 32}, obj32); code != 2 || errs != "" {
		t.Errorf("-m elf_i386: code=%d\n%s", code, errs)
	}
}

func TestLinkLikeGCC(t *testing.T) {
	dir := t.TempDir()
	mainObj := object(t, dir, "printf.o", 64, main64)
	if out, errs, code := runObjects(Config{Mode: GCC, Tool: "gcc"}, mainObj); out != "from main\n" || code != 6 || errs != "" {
		t.Errorf("gcc: out=%q code=%d err=%s", out, code, errs)
	}
	start := object(t, dir, "start.o", 64, start64)
	_, errs, code := runObjects(Config{Mode: GCC, Tool: "gcc", Argv0: "hello"}, start)
	if code != ExitBuild || !strings.Contains(errs, "gcc: undefined reference to main") || !strings.Contains(errs, "ld -o hello "+start) {
		t.Errorf("gcc with _start: code=%d\n%s", code, errs)
	}
	both := object(t, dir, "both.o", 64, "section .text\nglobal _start, main\n_start:\nmain:\n    ret\n")
	_, errs, code = runObjects(Config{Mode: GCC, Tool: "gcc"}, both)
	if code != ExitBuild || !strings.Contains(errs, "multiple definition of _start") {
		t.Errorf("gcc with main and _start: code=%d\n%s", code, errs)
	}
	obj32 := object(t, dir, "m32.o", 32, "section .text\nglobal main\nmain:\n    mov eax, 3\n    ret\n")
	if _, errs, code := runObjects(Config{Mode: GCC, Tool: "gcc"}, obj32); code != ExitBuild || !strings.Contains(errs, "gcc -m32") {
		t.Errorf("gcc 32-bit object: code=%d\n%s", code, errs)
	}
	if _, errs, code := runObjects(Config{Mode: GCC, Tool: "gcc", Bits: 32}, obj32); code != 3 || errs != "" {
		t.Errorf("gcc -m32: code=%d\n%s", code, errs)
	}
}

func TestProgramFile(t *testing.T) {
	dir := t.TempDir()
	obj := object(t, dir, "args.o", 64, `
section .text
global _start
_start:
    mov rsi, [rsp+8]    ; argv[0]
    mov rdx, 0
.len: cmp byte [rsi+rdx], 0
    je .print
    inc rdx
    jmp .len
.print:
    mov rax, 1
    mov rdi, 1
    syscall
    mov rax, 60
    mov rdi, [rsp]      ; argc
    syscall
`)
	crash := object(t, dir, "crash.o", 64, "section .text\nglobal _start\n_start:\n    xor rbx, rbx\n    mov rax, [rbx]\n")
	data, _ := os.ReadFile(obj)
	prog := filepath.Join(dir, "args")
	if err := progfile.Write(prog, progfile.Manifest{Linker: progfile.LD, Bits: 64, Objects: []string{"args.o"},
		Command: []string{"ld", "-o", "args", "args.o"}}, [][]byte{data}); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := Run(Config{Source: prog, Args: []string{"a", "b"}, MaxSteps: DefaultMaxSteps,
		Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb})
	if out.String() != prog || code != 3 || errb.String() != "" {
		t.Errorf("program file: argv0=%q argc=%d err=%s", out.String(), code, errb.String())
	}

	// The program's source lines come from the objects' debug information.
	data, _ = os.ReadFile(crash)
	prog = filepath.Join(dir, "crash")
	progfile.Write(prog, progfile.Manifest{Linker: progfile.LD, Bits: 64, Objects: []string{"crash.o"}}, [][]byte{data})
	out.Reset()
	errb.Reset()
	code = Run(Config{Source: prog, MaxSteps: DefaultMaxSteps, Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb})
	if code != 139 || !strings.Contains(errb.String(), "crash.asm line 5: mov rax, [rbx]") {
		t.Errorf("crash: code=%d\n%s", code, errb.String())
	}
}
