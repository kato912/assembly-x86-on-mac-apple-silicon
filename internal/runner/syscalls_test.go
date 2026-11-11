package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadEcho32(t *testing.T) {
	out, errs, code := runSource(t, `
section .bss
buf resb 64
section .text
global _start
_start:
    mov eax, 3          ; read(0, buf, 64)
    mov ebx, 0
    mov ecx, buf
    mov edx, 64
    int 0x80
    mov edx, eax        ; echo what was read
    mov eax, 4
    mov ebx, 1
    mov ecx, buf
    int 0x80
    mov eax, 1
    xor ebx, ebx
    int 0x80
`, "hello input\n")
	if out != "hello input\n" || code != 0 || errs != "" {
		t.Errorf("out=%q err=%q code=%d", out, errs, code)
	}
}

func TestFileRoundTrip64(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	src := `
section .data
path db "` + path + `", 0
text db "file data", 10
textlen equ $ - text
section .bss
buf resb 32
fd resq 1
section .text
global _start
_start:
    mov rax, 2              ; open(path, O_WRONLY|O_CREAT|O_TRUNC, 0644)
    mov rdi, path
    mov rsi, 0x241
    mov rdx, 0o644
    syscall
    mov [fd], rax
    mov rdi, rax            ; write(fd, text, textlen)
    mov rax, 1
    mov rsi, text
    mov rdx, textlen
    syscall
    mov rax, 3              ; close(fd)
    mov rdi, [fd]
    syscall
    mov rax, 2              ; open(path, O_RDONLY)
    mov rdi, path
    xor esi, esi
    syscall
    mov [fd], rax
    mov rdi, rax            ; lseek(fd, 5, SEEK_SET)
    mov rax, 8
    mov rsi, 5
    xor edx, edx
    syscall
    mov rax, 0              ; read(fd, buf, 32)
    mov rdi, [fd]
    mov rsi, buf
    mov rdx, 32
    syscall
    mov rdx, rax            ; write(1, buf, n)
    mov rax, 1
    mov rdi, 1
    mov rsi, buf
    syscall
    mov rax, 2              ; open a missing file -> -ENOENT
    mov rdi, missing
    xor esi, esi
    syscall
    neg rax
    mov rdi, rax            ; exit(ENOENT = 2)
    mov rax, 60
    syscall
section .data
missing db "/definitely/not/here", 0
`
	out, errs, code := runSource(t, src, "")
	if out != "data\n" || code != 2 || errs != "" {
		t.Errorf("out=%q err=%q code=%d", out, errs, code)
	}
	if b, _ := os.ReadFile(path); string(b) != "file data\n" {
		t.Errorf("file contains %q", b)
	}
}

func TestBrkAndTime(t *testing.T) {
	_, errs, code := runSource(t, `
section .text
global _start
_start:
    mov eax, 45         ; brk(0) -> current break
    xor ebx, ebx
    int 0x80
    mov esi, eax
    lea ebx, [eax + 8192]
    mov eax, 45         ; grow by 8 KiB
    int 0x80
    sub eax, esi
    cmp eax, 8192
    jne fail
    mov dword [esi + 8000], 42   ; new memory is usable
    mov eax, 13         ; time(NULL)
    xor ebx, ebx
    int 0x80
    cmp eax, 1700000000
    jb fail
    mov eax, 20         ; getpid
    int 0x80
    test eax, eax
    jz fail
    mov eax, 1
    mov ebx, [esi + 8000]
    int 0x80
fail:
    mov eax, 1
    mov ebx, 1
    int 0x80
`, "")
	if code != 42 || errs != "" {
		t.Errorf("code=%d err=%q", code, errs)
	}
}

func TestUnknownSyscallWarning(t *testing.T) {
	_, errs, code := runSource(t, `
section .text
global _start
_start:
    mov eax, 162        ; nanosleep: not supported
    int 0x80
    neg eax
    mov ebx, eax
    mov eax, 1
    int 0x80
`, "")
	if code != 38 || !strings.Contains(errs, "system call 162 is not supported") {
		t.Errorf("code=%d err=%q", code, errs)
	}
}
