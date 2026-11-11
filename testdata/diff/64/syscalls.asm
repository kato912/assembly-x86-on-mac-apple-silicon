; linux-only: uses read, brk and file system calls.
; Differential test: read stdin line by line and echo it upper-cased,
; grow the heap with brk, write/read/seek/close/unlink a file.
default rel
%include "dump64.inc"

section .data
fname:  db "nasmx86_difftest_64.tmp", 0
text:   db "file contents 123", 10
tlen    equ $ - text
section .bss
buf:    resb 256
fd:     resq 1
heap0:  resq 1

section .text
global _start
_start:
    INIT
    LABEL "read + upper-case"
.read:
    xor eax, eax                ; read(0, buf, 256)
    xor edi, edi
    lea rsi, [buf]
    mov edx, 256
    syscall
    test rax, rax
    jle .eof
    mov rcx, rax
    lea rsi, [buf]
.up:
    mov dl, [rsi]
    cmp dl, 'a'
    jb .skip
    cmp dl, 'z'
    ja .skip
    sub byte [rsi], 32
.skip:
    inc rsi
    loop .up
    mov rdx, rax                ; write(1, buf, n)
    mov eax, 1
    mov edi, 1
    lea rsi, [buf]
    syscall
    jmp .read
.eof:
    xor esi, esi
    xor edx, edx
    xor ecx, ecx
    xor r11d, r11d
    DUMP

    LABEL "brk"
    mov eax, 12                 ; brk(0)
    xor edi, edi
    syscall
    mov [heap0], rax
    lea rdi, [rax + 0x3000]
    mov eax, 12
    syscall
    sub rax, [heap0]            ; grew by 0x3000
    mov rbx, rax
    mov rdi, [heap0]
    mov qword [rdi + 0x2ff8], 0x77  ; the new memory is usable
    mov rcx, [rdi + 0x2ff8]
    xor edi, edi
    xor r11d, r11d
    xor eax, eax
    DUMP

    LABEL "files"
    mov eax, 2                  ; open(fname, O_WRONLY|O_CREAT|O_TRUNC, 0644)
    lea rdi, [fname]
    mov esi, 0x241
    mov edx, 0o644
    syscall
    mov [fd], rax
    mov rdi, rax                ; write(fd, text, tlen)
    mov eax, 1
    lea rsi, [text]
    mov edx, tlen
    syscall
    mov rbx, rax
    mov eax, 3                  ; close(fd)
    mov rdi, [fd]
    syscall
    mov r12, rax
    mov eax, 2                  ; open(fname, O_RDONLY)
    lea rdi, [fname]
    xor esi, esi
    syscall
    mov [fd], rax
    mov rdi, rax                ; lseek(fd, 5, SEEK_SET)
    mov eax, 8
    mov esi, 5
    xor edx, edx
    syscall
    mov r13, rax
    xor eax, eax                ; read(fd, buf, 256)
    mov rdi, [fd]
    lea rsi, [buf]
    mov edx, 256
    syscall
    mov r14, rax
    PRINTMEM buf, tlen - 5
    mov eax, 3
    mov rdi, [fd]
    syscall
    mov eax, 87                 ; unlink(fname)
    lea rdi, [fname]
    syscall
    mov r15, rax
    mov eax, 2                  ; opening it again fails with -ENOENT
    lea rdi, [fname]
    xor esi, esi
    syscall
    xor ecx, ecx
    xor edx, edx
    xor esi, esi
    xor edi, edi
    xor r11d, r11d
    DUMP
    EXIT 0
