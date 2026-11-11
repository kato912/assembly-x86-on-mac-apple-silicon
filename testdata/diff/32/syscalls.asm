; Differential test: read, write, brk, creat/open/close/lseek/unlink with
; int 0x80, plus the exit status.
; exit: 7
%include "dump32.inc"

section .data
fname:  db "nasmx86_difftest_32.tmp", 0
text:   db "some file data", 10
tlen    equ $ - text
section .bss
buf:    resb 128
fd:     resd 1
heap0:  resd 1

section .text
global _start
_start:
    INIT
    LABEL "read (echo reversed words count)"
    mov eax, 3                  ; read(0, buf, 128)
    xor ebx, ebx
    mov ecx, buf
    mov edx, 128
    int 0x80
    mov esi, eax                ; bytes read
    mov eax, 4                  ; write(1, buf, n)
    mov ebx, 1
    mov ecx, buf
    mov edx, esi
    int 0x80
    xor ecx, ecx
    DUMP
    LABEL "brk"
    mov eax, 45
    xor ebx, ebx
    int 0x80
    mov [heap0], eax
    lea ebx, [eax + 0x2000]
    mov eax, 45
    int 0x80
    sub eax, [heap0]
    mov ebx, [heap0]
    mov dword [ebx + 0x1ffc], 0x1234
    mov ecx, [ebx + 0x1ffc]
    xor ebx, ebx
    DUMP
    LABEL "files"
    mov eax, 8                  ; creat(fname, 0644)
    mov ebx, fname
    mov ecx, 0o644
    int 0x80
    mov [fd], eax
    mov ebx, eax                ; write(fd, text, tlen)
    mov eax, 4
    mov ecx, text
    mov edx, tlen
    int 0x80
    mov esi, eax
    mov eax, 6                  ; close
    mov ebx, [fd]
    int 0x80
    mov eax, 5                  ; open(fname, O_RDONLY)
    mov ebx, fname
    xor ecx, ecx
    int 0x80
    mov [fd], eax
    mov ebx, eax                ; lseek(fd, -5, SEEK_END)
    mov eax, 19
    mov ecx, -5
    mov edx, 2
    int 0x80
    mov edi, eax
    mov eax, 3                  ; read(fd, buf, 128)
    mov ebx, [fd]
    mov ecx, buf
    mov edx, 128
    int 0x80
    mov ebp, eax
    PRINTMEM buf, 5
    mov eax, 6
    mov ebx, [fd]
    int 0x80
    mov eax, 10                 ; unlink
    mov ebx, fname
    int 0x80
    mov edx, eax
    mov eax, 5                  ; open again: -ENOENT
    mov ebx, fname
    xor ecx, ecx
    int 0x80
    xor ebx, ebx
    xor ecx, ecx
    DUMP
    EXIT 7
