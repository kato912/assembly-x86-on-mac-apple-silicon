package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
)

// runSource assembles and runs src, returning stdout, stderr and exit code.
func runSource(t *testing.T, src string, stdin string, args ...string) (string, string, int) {
	t.Helper()
	return runSourceCfg(t, src, func(c *Config) {
		c.Stdin = strings.NewReader(stdin)
		c.Args = args
	})
}

// runSourceCfg is runSource with a hook to adjust the configuration.
func runSourceCfg(t *testing.T, src string, adjust func(*Config)) (string, string, int) {
	t.Helper()
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "prog.asm")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	cfg := Config{Source: path, MaxSteps: DefaultMaxSteps, Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb}
	adjust(&cfg)
	code := Run(cfg)
	return out.String(), errb.String(), code
}

func TestHello32(t *testing.T) {
	out, errs, code := runSource(t, `
section .data
msg db "Hello, World!", 10
len equ $ - msg
section .text
global _start
_start:
    mov eax, 4
    mov ebx, 1
    mov ecx, msg
    mov edx, len
    int 0x80
    mov eax, 1
    mov ebx, 3
    int 0x80
`, "")
	if out != "Hello, World!\n" || code != 3 || errs != "" {
		t.Errorf("out=%q err=%q code=%d", out, errs, code)
	}
}

func TestHello64(t *testing.T) {
	out, errs, code := runSource(t, `
section .data
msg db "Hello, World!", 10
len equ $ - msg
section .text
global _start
_start:
    mov rax, 1
    mov rdi, 1
    mov rsi, msg
    mov rdx, len
    syscall
    mov rax, 60
    xor rdi, rdi
    syscall
`, "")
	if out != "Hello, World!\n" || code != 0 || errs != "" {
		t.Errorf("out=%q err=%q code=%d", out, errs, code)
	}
}

func TestArgvAndMainReturn(t *testing.T) {
	// main returns argc in eax; the C-like exit stub turns it into the exit code.
	_, errs, code := runSource(t, `
section .text
global main
main:
    mov eax, edi
    ret
`, "", "a", "b")
	if code != 3 || errs != "" {
		t.Errorf("code=%d err=%q", code, errs)
	}
	// _start: argc is at [esp], argv[1] at [esp+8]
	out, _, code := runSource(t, `
section .text
global _start
_start:
    mov ecx, [esp+8]      ; argv[1]
    mov eax, 4
    mov ebx, 1
    mov edx, 3
    int 0x80
    mov eax, 1
    mov ebx, [esp]        ; argc
    int 0x80
`, "", "xyz")
	if out != "xyz" || code != 2 {
		t.Errorf("out=%q code=%d", out, code)
	}
}

func TestUndefinedSymbolMessage(t *testing.T) {
	_, errs, code := runSource(t, "extern prnitf\nglobal main\nsection .text\nmain: call prnitf\n", "")
	if code != ExitBuild || !strings.Contains(errs, "prnitf") || !strings.Contains(errs, "printf, puts") {
		t.Errorf("code=%d err=%q", code, errs)
	}
}

func TestYasmStyleSource(t *testing.T) {
	// Program in the style of YASM course material: ddq/resdq, octal "q".
	_, errs, code := runSource(t, `
section .data
S_IRUSR equ 00400q
wide    ddq 0x1234, -2
section .bss
big     resdq 2
section .text
global _start
_start:
    mov rax, [wide]         ; low half of the first ddq
    cmp rax, 0x1234
    jne fail
    mov rax, [wide + 24]    ; high half of -2 is all ones
    cmp rax, -1
    jne fail
    mov rax, big + 32       ; resdq 2 reserved 32 bytes
    sub rax, big
    mov rdi, rax
    add rdi, S_IRUSR - 256  ; 32 + 0
    mov rax, 60
    syscall
fail:
    mov rax, 60
    mov rdi, 1
    syscall
`, "")
	if code != 32 || errs != "" {
		t.Errorf("code=%d err=%q", code, errs)
	}
}
