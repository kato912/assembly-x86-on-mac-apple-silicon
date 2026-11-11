; stat/fstat/mkdir/rmdir/rename/nanosleep/clock_gettime/gettimeofday with
; int 0x80 (32-bit struct layouts). Only deterministic results are compared.
%include "dump32.inc"

section .data
fname:  db "nasmx86_sc2_32.tmp", 0
dname:  db "nasmx86_sc2_32.dir", 0
dname2: db "nasmx86_sc2_32.dir2", 0
text:   db "seven b"
nap:    dd 0, 1000000
section .bss
st:     resb 64
ts:     resd 2
fd:     resd 1

section .text
global _start
_start:
    INIT
    LABEL "fstat / stat"
    mov eax, 8                  ; creat(fname, 0644)
    mov ebx, fname
    mov ecx, 0o644
    int 0x80
    mov [fd], eax
    mov ebx, eax                ; write(fd, text, 7)
    mov eax, 4
    mov ecx, text
    mov edx, 7
    int 0x80
    mov eax, 108                ; fstat(fd, st)
    mov ebx, [fd]
    mov ecx, st
    int 0x80
    mov esi, eax
    mov edi, [st + 20]          ; st_size = 7
    movzx ebp, word [st + 8]
    and ebp, 0o170000           ; S_IFREG
    mov eax, 6
    mov ebx, [fd]
    int 0x80
    mov eax, 106                ; stat(fname, st)
    mov ebx, fname
    mov ecx, st
    int 0x80
    mov ebx, eax
    mov ecx, [st + 20]
    xor edx, edx
    DUMP FL_NONE

    LABEL "mkdir / rename / rmdir / time"
    mov eax, 39                 ; mkdir
    mov ebx, dname
    mov ecx, 0o755
    int 0x80
    mov esi, eax
    mov eax, 38                 ; rename
    mov ebx, dname
    mov ecx, dname2
    int 0x80
    mov edi, eax
    mov eax, 40                 ; rmdir
    mov ebx, dname2
    int 0x80
    mov ebp, eax
    mov eax, 40                 ; rmdir again -> -ENOENT
    mov ebx, dname2
    int 0x80
    mov edx, eax
    mov eax, 10                 ; unlink(fname)
    mov ebx, fname
    int 0x80
    mov eax, 162                ; nanosleep(&nap, NULL)
    mov ebx, nap
    xor ecx, ecx
    int 0x80
    mov ebx, eax
    mov eax, 265                ; clock_gettime(CLOCK_MONOTONIC, ts)
    mov ecx, ts
    push ebx
    mov ebx, 1
    int 0x80
    pop ebx
    mov ecx, eax
    mov eax, 78                 ; gettimeofday(ts, NULL)
    push ebx
    mov ebx, ts
    push ecx
    xor ecx, ecx
    int 0x80
    pop ecx
    pop ebx
    DUMP FL_NONE
    EXIT 0
