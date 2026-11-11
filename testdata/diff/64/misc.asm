; Differential test: xchg, xadd, cmpxchg, cbw/cwde/cdqe, cwd/cdq/cqo, xlat,
; lahf/sahf, clc/stc/cmc, nop forms.
default rel
%include "dump64.inc"

section .data
mem:    dq 0x1111111111111111
table:  db "0123456789abcdef"

section .text
global _start
_start:
    INIT
    LABEL "xchg"
    mov rax, 1
    mov rbx, 2
    xchg rax, rbx
    mov ecx, 3
    mov rdx, -1
    xchg ecx, edx               ; 32-bit: both zero-extended
    mov si, 0x1234
    mov di, 0x5678
    xchg si, di
    xchg al, ah
    mov r8, 0x55
    xchg r8, [mem]
    mov r9, [mem]
    DUMP

    LABEL "xadd / cmpxchg"
    mov rax, 5
    mov rbx, 7
    xadd rax, rbx
    DUMP
    mov qword [mem], 10
    mov rcx, 3
    xadd [mem], rcx
    mov rdx, [mem]
    DUMP
    mov rax, 13
    mov rcx, 99
    cmpxchg [mem], rcx          ; equal: mem = 99
    mov rdx, [mem]
    DUMP
    mov rax, 1
    cmpxchg [mem], rcx          ; not equal: rax = 99
    DUMP 0x440                  ; (other flags: see cmpxchg_flags.asm)
    mov eax, 5
    mov ebx, 5
    mov ecx, 6
    cmpxchg ebx, ecx
    DUMP
    mov al, 1
    mov bl, 2
    cmpxchg bl, cl
    DUMP 0x440
    FLAGS 0x202

    LABEL "sign extension"
    mov rax, 0x1234567890abcd80
    cbw
    DUMP
    cwde
    DUMP
    cdqe
    DUMP
    mov rax, 0x8000
    cwd
    DUMP
    mov eax, 0x80000000
    cdq
    DUMP
    mov rax, -5
    cqo
    DUMP
    mov rax, 5
    cqo
    DUMP

    LABEL "xlat"
    lea rbx, [table]
    mov al, 11
    xlatb
    movzx ecx, al
    xor ebx, ebx
    DUMP

    LABEL "lahf / sahf / clc / stc / cmc"
    FLAGS 0x8d7
    lahf
    movzx ecx, ah
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
    cmc
    DUMP

    LABEL "nop forms"
    nop
    db 0x0f, 0x1f, 0x44, 0x00, 0x00   ; nop dword [rax+rax]
    db 0x66, 0x90                     ; xchg ax, ax = nop
    pause
    DUMP
    EXIT 0
