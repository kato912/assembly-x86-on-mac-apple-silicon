package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
)

// Programs calling the built-in C library. Expected outputs are what the
// same programs print when linked with gcc against glibc.
func TestLibc(t *testing.T) {
	cases := []struct {
		name, src, stdin string
		out, errs        string // errs: substring of stderr ("" = must be empty)
		code             int
	}{
		{name: "printf formats 64", src: `
default rel
extern printf
section .data
f1 db "%d|%5d|%-5d|%05d|%+d|% d|", 0
f2 db "%x|%X|%#x|%o|%#o|%u|", 0
f3 db "%ld|%lu|%hhd|%hd|%.0d|%.3d|%%|", 0
f4 db "%s|%.2s|%5s|%-4s|%c%c|%p|%p", 10, 0
s1 db "hello", 0
s2 db "ab", 0
section .text
global main
main:
    push rbp
    lea rdi, [f1]
    mov esi, 42
    mov edx, 42
    mov ecx, 42
    mov r8d, 42
    mov r9d, 42
    push 0          ; keep the stack aligned: 2 pushes
    push 42
    xor eax, eax
    call printf
    add rsp, 16
    lea rdi, [f2]
    mov esi, 255
    mov edx, 255
    mov ecx, 255
    mov r8d, 8
    mov r9d, 8
    push 0
    push -1
    xor eax, eax
    call printf
    add rsp, 16
    lea rdi, [f3]
    mov rsi, -1
    mov rdx, -1
    mov ecx, 255
    mov r8d, 65535
    xor r9d, r9d
    push 0
    push 7
    xor eax, eax
    call printf
    add rsp, 16
    lea rdi, [f4]
    lea rsi, [s1]
    lea rdx, [s1]
    lea rcx, [s2]
    lea r8, [s2]
    mov r9d, 'o'
    push 0          ; alignment pad (never read)
    push 0x1000
    push 0
    push 'k'
    xor eax, eax
    call printf
    add rsp, 32
    xor eax, eax
    pop rbp
    ret
`, out: "42|   42|42   |00042|+42| 42|ff|FF|0xff|10|010|4294967295|-1|18446744073709551615|-1|-1||007|%|hello|he|   ab|ab  |ok|(nil)|0x1000\n"},

		{name: "scanf and return values", stdin: "12 30\nword x\n", src: `
default rel
extern printf, scanf
section .data
in1 db "%d %d", 0
in2 db "%s %c", 0
ofmt db "%d+%d=%d n=%d w=%s c=%c", 10, 0
section .bss
a resd 1
b resd 1
w resb 16
c resb 1
section .text
global main
main:
    push rbx
    lea rdi, [in1]
    lea rsi, [a]
    lea rdx, [b]
    xor eax, eax
    call scanf
    mov ebx, eax            ; 2 items
    lea rdi, [in2]
    lea rsi, [w]
    lea rdx, [c]
    xor eax, eax
    call scanf
    add ebx, eax            ; +2
    lea rdi, [ofmt]
    mov esi, [a]
    mov edx, [b]
    lea ecx, [rsi+rdx]
    mov r8d, ebx
    lea r9, [w]
    movzx eax, byte [c]
    push rax
    push rax
    xor eax, eax
    call printf
    add rsp, 16
    lea rdi, [in1]          ; at end of input scanf returns EOF (-1) in eax
    lea rsi, [a]
    lea rdx, [b]
    xor eax, eax
    call scanf
    cmp eax, -1
    jne bad
    mov ecx, eax            ; int return: rax upper half is 0, as in real code
    shr rax, 32
    jnz bad
    xor eax, eax
    pop rbx
    ret
bad:
    mov eax, 1
    pop rbx
    ret
`, out: "12+30=42 n=4 w=word c=x\n"},

		{name: "cdecl 32-bit", src: `
extern printf, puts, exit
section .data
fmt db "%d %s %lld %.2f", 10, 0
s   db "two", 0
msg db "bye", 0
d   dq 2.5
section .text
global main
main:
    push ebp
    mov ebp, esp
    push dword [d+4]
    push dword [d]
    push 0
    push 3
    push s
    push 1
    push fmt
    call printf
    add esp, 28
    push msg
    call puts
    add esp, 4
    push 7
    call exit
`, out: "1 two 3 2.50\nbye\n", code: 7},

		{name: "stdout is buffered like glibc on a pipe", src: `
default rel
extern printf
section .data
a db "A", 10, 0
b db "B", 10
section .text
global main
main:
    push rbp
    lea rdi, [a]
    xor eax, eax
    call printf             ; buffered (stdout is not a terminal)
    mov eax, 1              ; write(1, "B\n", 2) goes out at once
    mov edi, 1
    lea rsi, [b]
    mov edx, 2
    syscall
    xor eax, eax
    pop rbp
    ret                     ; returning from main flushes "A\n"
`, out: "B\nA\n"},

		{name: "raw exit loses buffered output", src: `
extern printf
section .data
msg db "lost", 0
section .text
global _start
_start:
    push msg
    call printf
    add esp, 4
    mov eax, 1
    mov ebx, 0
    int 0x80
`, out: "", errs: `4 bytes of printf/puts output`},

		{name: "strings and memory", src: `
default rel
extern malloc, calloc, realloc, free, strcpy, strcat, strlen, strcmp, strncmp, memset, memcpy, puts, printf, strchr, atoi, sprintf, sscanf, toupper
section .data
hello db "hello", 0
world db " world", 0
num   db "  -123xyz", 0
pfmt  db "%s|%d|%d|%d|%d|%c|%s|%d", 10, 0
sfmt  db "n=%d", 0
input db "7 8", 0
iscan db "%d %d", 0
section .bss
buf    resb 64
x      resd 1
y      resd 1
r_len  resd 1
r_cmp  resd 1
r_atoi resd 1
r_up   resd 1
r_spr  resd 1
section .text
global main
main:
    push rbx
    mov edi, 32
    call malloc
    mov rbx, rax
    mov rdi, rbx
    lea rsi, [hello]
    call strcpy
    mov rdi, rbx
    lea rsi, [world]
    call strcat
    mov rdi, rbx
    call strlen
    mov [r_len], eax        ; 11
    lea rdi, [hello]
    lea rsi, [world]
    call strcmp
    mov [r_cmp], eax        ; 'h' - ' ' = 72
    lea rdi, [num]
    call atoi
    mov [r_atoi], eax       ; -123
    mov edi, 'q'
    call toupper
    mov [r_up], eax
    lea rdi, [buf]
    lea rsi, [sfmt]
    mov edx, 5
    xor eax, eax
    call sprintf            ; "n=5", returns 3
    mov [r_spr], eax
    lea rdi, [input]
    lea rsi, [iscan]
    lea rdx, [x]
    lea rcx, [y]
    xor eax, eax
    call sscanf
    mov eax, [x]
    imul eax, [y]           ; 56
    lea rdi, [pfmt]
    mov rsi, rbx
    mov edx, [r_len]
    mov ecx, [r_cmp]
    mov r8d, eax
    mov r9d, [r_spr]
    push 0                  ; alignment pad
    movsxd rax, dword [r_atoi]
    push rax
    lea rax, [buf]
    push rax
    mov eax, [r_up]
    push rax
    xor eax, eax
    call printf
    add rsp, 32
    mov rdi, rbx
    mov esi, 64
    call realloc            ; keeps the contents
    mov rbx, rax
    mov rdi, rbx
    call puts
    mov rdi, rbx
    call free
    mov edi, 4
    mov esi, 4
    call calloc
    mov eax, [rax+12]       ; calloc'd memory is zero
    pop rbx
    ret
`, out: "hello world|11|72|56|3|Q|n=5|-123\nhello world\n"},

		{name: "rand matches glibc", src: `
default rel
extern srand, rand, printf
section .data
f db "%d %d %d", 10, 0
section .text
global main
main:
    push rbx
    push r12
    push r13
    call rand               ; unseeded = srand(1)
    mov ebx, eax
    mov edi, 42
    call srand
    call rand
    mov r12d, eax
    call rand
    mov r13d, eax
    lea rdi, [f]
    mov esi, ebx
    mov edx, r12d
    mov ecx, r13d
    xor eax, eax
    call printf
    pop r13
    pop r12
    pop rbx
    xor eax, eax
    ret
`, out: "1804289383 71876166 708592740\n"},

		{name: "files", src: `
default rel
extern fopen, fprintf, fclose, fgets, puts, stderr, fputs
section .data
path  db "PATH", 0
wmode db "w", 0
rmode db "r", 0
fmt   db "line %d", 10, 0
err   db "to stderr", 10, 0
section .bss
buf resb 32
section .text
global main
main:
    push rbx
    lea rdi, [path]
    lea rsi, [wmode]
    call fopen
    mov rbx, rax
    mov rdi, rbx
    lea rsi, [fmt]
    mov edx, 42
    xor eax, eax
    call fprintf
    mov rdi, rbx
    call fclose
    lea rdi, [path]
    lea rsi, [rmode]
    call fopen
    mov rbx, rax
    lea rdi, [buf]
    mov esi, 32
    mov rdx, rbx
    call fgets
    lea rdi, [buf]
    call puts
    mov rdi, rbx
    call fclose
    lea rdi, [err]
    mov rsi, [stderr]
    call fputs
    xor eax, eax
    pop rbx
    ret
`, out: "line 42\n\n", errs: "to stderr"},

		{name: "caller-saved registers are clobbered", src: `
default rel
extern puts
section .data
s db "x", 0
section .text
global main
main:
    push rbx
    mov ecx, 5              ; a loop counter in rcx does not survive a call
    mov ebx, 5              ; rbx does (callee-saved)
    lea rdi, [s]
    call puts
    cmp ebx, 5
    jne bad
    cmp ecx, 5
    je bad
    xor eax, eax
    pop rbx
    ret
bad:
    mov eax, 1
    pop rbx
    ret
`, out: "x\n"},

		{name: "PIE-style calls through the GOT and PLT", src: `
default rel
extern printf, fflush, stdout
section .data
f db "got %d", 10, 0
section .text
global main
main:
    push rbp
    lea rdi, [f]
    mov esi, 7
    xor eax, eax
    call [rel printf wrt ..got]
    mov rax, [rel stdout wrt ..got]
    mov rdi, [rax]
    call fflush wrt ..plt
    xor eax, eax
    pop rbp
    ret
`, out: "got 7\n"},

		{name: "scanf value instead of address", stdin: "5\n", src: `
default rel
extern scanf
section .data
f db "%d", 0
n dd 3
section .text
global main
main:
    push rbp
    lea rdi, [f]
    mov esi, [n]
    xor eax, eax
    call scanf
    pop rbp
    ret
`, errs: "not its value", code: 139},

		{name: "misaligned stack", src: `
default rel
extern scanf
section .data
f db "%d", 0
n dd 0
section .text
global main
main:
    lea rdi, [f]
    lea rsi, [n]
    xor eax, eax
    call scanf
    ret
`, errs: "multiple of 16", code: 139},

		{name: "double free", src: `
default rel
extern malloc, free
section .text
global main
main:
    push rbx
    mov edi, 8
    call malloc
    mov rbx, rax
    mov rdi, rax
    call free
    mov rdi, rbx
    call free
    pop rbx
    ret
`, errs: "double free", code: 134},

		{name: "FILE variable instead of its value", src: `
default rel
extern fputs, stdout
section .data
s db "x", 0
section .text
global main
main:
    push rbp
    lea rdi, [s]
    mov rsi, stdout
    call fputs
    pop rbp
    ret
`, errs: "use [stdout]", code: 139},

		{name: "%f in 64-bit", src: `
default rel
extern printf
section .data
f db "%f", 0
section .text
global main
main:
    push rbp
    lea rdi, [f]
    mov eax, 1
    call printf
    pop rbp
    ret
`, errs: "xmm", code: 139},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := c.src
			if strings.Contains(src, `"PATH"`) {
				src = strings.Replace(src, "PATH", filepath.ToSlash(filepath.Join(t.TempDir(), "f.txt")), 1)
			}
			out, errs, code := runSource(t, src, c.stdin)
			if out != c.out || code != c.code {
				t.Errorf("out=%q code=%d, want out=%q code=%d\nstderr: %s", out, code, c.out, c.code, errs)
			}
			if c.errs == "" && errs != "" || !strings.Contains(errs, c.errs) {
				t.Errorf("stderr=%q, want it to contain %q", errs, c.errs)
			}
		})
	}
}

// buildSource assembles and loads src without running it.
func buildSource(t *testing.T, src string) (*Program, int) {
	t.Helper()
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	path := filepath.Join(t.TempDir(), "prog.asm")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	return Build(Config{Source: path, MaxSteps: DefaultMaxSteps, Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb})
}

// A program that never calls the C library does not get it mapped.
func TestNoLibcWithoutExterns(t *testing.T) {
	p, code := buildSource(t, "global _start\n_start: mov eax, 1\nint 0x80\n")
	if p == nil {
		t.Fatalf("build failed: %d", code)
	}
	defer p.Close()
	if p.Lib != nil || p.CPU.NativeCall != nil {
		t.Error("C library installed for a program without externs")
	}
}
