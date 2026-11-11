package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTimeSyscalls64(t *testing.T) {
	_, errs, code := runSource(t, `
section .bss
ts1: resq 2
ts2: resq 2
tv:  resq 2
section .data
nap: dq 0, 2000000          ; 2 ms
section .text
global _start
_start:
    mov eax, 228            ; clock_gettime(CLOCK_MONOTONIC, ts1)
    mov edi, 1
    mov rsi, ts1
    syscall
    test rax, rax
    jnz fail
    mov eax, 35             ; nanosleep(&nap, NULL)
    mov rdi, nap
    xor esi, esi
    syscall
    test rax, rax
    jnz fail
    mov eax, 228            ; clock_gettime(CLOCK_MONOTONIC, ts2)
    mov edi, 1
    mov rsi, ts2
    syscall
    mov rax, [ts2]          ; elapsed ns = (s2-s1)*1e9 + (n2-n1)
    sub rax, [ts1]
    imul rax, rax, 1000000000
    add rax, [ts2+8]
    sub rax, [ts1+8]
    cmp rax, 2000000
    jb fail                 ; slept at least 2 ms
    mov eax, 96             ; gettimeofday(tv, NULL)
    mov rdi, tv
    xor esi, esi
    syscall
    cmp qword [tv], 1700000000
    jb fail
    cmp qword [tv+8], 1000000
    jae fail                ; microseconds < 1e6
    mov eax, 228            ; bad clock id -> -EINVAL
    mov edi, 999
    mov rsi, ts1
    syscall
    cmp rax, -22
    jne fail
    mov eax, 60
    xor edi, edi
    syscall
fail:
    mov eax, 60
    mov edi, 1
    syscall
`, "")
	if code != 0 || errs != "" {
		t.Errorf("code=%d err=%q", code, errs)
	}
}

func TestFileInfoSyscalls(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.txt")
	os.WriteFile(file, []byte("0123456789abc"), 0o644) // 13 bytes
	sub := filepath.Join(dir, "sub")
	moved := filepath.Join(dir, "moved")
	// 64-bit: stat -> st_size at offset 48, st_mode at offset 24
	_, errs, code := runSource(t, `
section .data
path:  db "`+file+`", 0
sub:   db "`+sub+`", 0
moved: db "`+moved+`", 0
section .bss
st: resb 144
section .text
global _start
_start:
    mov eax, 4              ; stat(path, st)
    mov rdi, path
    mov rsi, st
    syscall
    test rax, rax
    jnz fail
    cmp qword [st + 48], 13
    jne fail
    mov eax, [st + 24]
    and eax, 0o170000
    cmp eax, 0o100000       ; S_IFREG
    jne fail
    mov eax, 2              ; open(path, O_RDONLY)
    mov rdi, path
    xor esi, esi
    syscall
    mov rdi, rax            ; fstat(fd, st)
    mov eax, 5
    mov rsi, st
    syscall
    cmp qword [st + 48], 13
    jne fail
    mov eax, 83             ; mkdir(sub, 0755)
    mov rdi, sub
    mov esi, 0o755
    syscall
    test rax, rax
    jnz fail
    mov eax, 83             ; again -> -EEXIST
    mov rdi, sub
    mov esi, 0o755
    syscall
    cmp rax, -17
    jne fail
    mov eax, 82             ; rename(sub, moved)
    mov rdi, sub
    mov rsi, moved
    syscall
    test rax, rax
    jnz fail
    mov eax, 84             ; rmdir(moved)
    mov rdi, moved
    syscall
    test rax, rax
    jnz fail
    mov eax, 84             ; rmdir(path) -> -ENOTDIR
    mov rdi, path
    syscall
    cmp rax, -20
    jne fail
    mov eax, 60
    xor edi, edi
    syscall
fail:
    mov eax, 60
    mov edi, 1
    syscall
`, "")
	if code != 0 || errs != "" {
		t.Errorf("64-bit: code=%d err=%q", code, errs)
	}
	if _, err := os.Stat(moved); !os.IsNotExist(err) {
		t.Errorf("directory was not removed")
	}

	// 32-bit (int 0x80): struct stat has st_size at offset 20, st_mode (16 bits) at 8
	_, errs, code = runSource(t, `
section .data
path: db "`+file+`", 0
section .bss
st: resb 64
section .text
global _start
_start:
    mov eax, 5              ; open
    mov ebx, path
    xor ecx, ecx
    int 0x80
    mov ebx, eax            ; fstat(fd, st)
    mov eax, 108
    mov ecx, st
    int 0x80
    test eax, eax
    jnz fail
    cmp dword [st + 20], 13
    jne fail
    movzx eax, word [st + 8]
    and eax, 0o170000
    cmp eax, 0o100000
    jne fail
    mov eax, 1
    xor ebx, ebx
    int 0x80
fail:
    mov eax, 1
    mov ebx, 1
    int 0x80
`, "")
	if code != 0 || errs != "" {
		t.Errorf("32-bit: code=%d err=%q", code, errs)
	}
}
