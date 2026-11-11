; Differential test: xchg, xadd, cmpxchg, cbw/cwde/cwd/cdq, xlat,
; lahf/sahf, clc/stc/cmc, nop.
%include "dump32.inc"

section .data
mem:    dd 0x11111111
table:  db "0123456789abcdef"

section .text
global _start
_start:
    INIT
    LABEL "xchg / xadd"
    mov eax, 1
    mov ebx, 2
    xchg eax, ebx
    xchg cl, ch
    mov edx, 0x55
    xchg edx, [mem]
    mov esi, [mem]
    mov edi, 3
    xadd [mem], edi
    mov ebp, [mem]
    DUMP
    LABEL "cmpxchg"
    FLAGS 0x202
    mov eax, 0x58
    mov ecx, 9
    cmpxchg [mem], ecx          ; equal
    mov edx, [mem]
    DUMP
    FLAGS 0x202
    mov eax, 1
    cmpxchg [mem], ecx          ; not equal
    DUMP
    FLAGS 0x202
    mov al, 0x80
    mov bl, 0x7f
    cmpxchg bl, cl
    DUMP
    LABEL "sign extension"
    mov eax, 0x12345680
    cbw
    DUMP
    cwde
    DUMP
    mov eax, 0x8000
    cwd
    DUMP
    mov eax, -3
    cdq
    DUMP
    mov eax, 3
    cdq
    DUMP
    LABEL "xlat / lahf / sahf / flags"
    mov ebx, table
    mov al, 13
    xlatb
    movzx ecx, al
    xor ebx, ebx
    FLAGS 0x8d7
    lahf
    movzx edx, ah
    FLAGS 0x202
    mov ah, 0xd5
    sahf
    DUMP
    clc
    DUMP
    stc
    DUMP
    cmc
    DUMP
    nop
    DUMP
    EXIT 0
