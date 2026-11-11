; Differential test: add, adc, sub, sbb, cmp, inc, dec, neg, not on
; 8/16/32/64-bit registers (incl. r8-r15), immediates and memory.
default rel
%include "dump64.inc"

; T op, a, b, pre: run op on every operand size with a in rax/r8 and b in
; rbx/r9, preceded by pre (clc/stc/nop).
%macro T 4
    mov rax, %2
    mov rbx, %3
    %4
    %1 al, bl
    DUMP
    mov rax, %2
    mov rbx, %3
    %4
    %1 ax, bx
    DUMP
    mov rax, %2
    mov rbx, %3
    %4
    %1 eax, ebx            ; 32-bit result zero-extends into rax
    DUMP
    mov rax, %2
    mov rbx, %3
    %4
    %1 rax, rbx
    DUMP
    mov r8, %2
    mov r9, %3
    %4
    %1 r8b, r9b
    DUMP
    mov r8, %2
    mov r9, %3
    %4
    %1 r8d, r9d
    DUMP
%endmacro

%macro VALUES 2
    T %1, 0, 0, %2
    T %1, 1, 1, %2
    T %1, 1, 2, %2
    T %1, 0x7f, 1, %2
    T %1, 0x80, 0x80, %2
    T %1, 0xff, 1, %2
    T %1, 0x0f, 0x01, %2
    T %1, 0x7fff, 1, %2
    T %1, 0x8000, 0xffff, %2
    T %1, 0x7fffffff, 1, %2
    T %1, 0x80000000, 0x80000000, %2
    T %1, 0xffffffff, 1, %2
    T %1, 0x7fffffffffffffff, 1, %2
    T %1, 0x8000000000000000, 0x8000000000000000, %2
    T %1, -1, 1, %2
    T %1, -1, -1, %2
    T %1, 0x123456789abcdef0, 0x0fedcba987654321, %2
    T %1, 0xfedcba9876543210, 0x1122334455667788, %2
    T %1, 5, 7, %2
%endmacro

%macro U 3
    mov rax, %2
    %3
    %1 al
    DUMP
    mov rax, %2
    %3
    %1 ax
    DUMP
    mov rax, %2
    %3
    %1 eax
    DUMP
    mov rax, %2
    %3
    %1 rax
    DUMP
    mov r15, %2
    %3
    %1 r15w
    DUMP
%endmacro

%macro UVALUES 2
    U %1, 0, %2
    U %1, 1, %2
    U %1, 0x0f, %2
    U %1, 0x7f, %2
    U %1, 0x80, %2
    U %1, 0xff, %2
    U %1, 0x8000, %2
    U %1, 0x7fffffff, %2
    U %1, 0x80000000, %2
    U %1, 0xffffffff, %2
    U %1, 0x7fffffffffffffff, %2
    U %1, 0x8000000000000000, %2
    U %1, -1, %2
%endmacro

section .data
m64: dq 0
m32: dd 0
m8:  db 0

section .text
global _start
_start:
    INIT
    LABEL "add"
    VALUES add, nop
    LABEL "adc cf=0"
    VALUES adc, clc
    LABEL "adc cf=1"
    VALUES adc, stc
    LABEL "sub"
    VALUES sub, nop
    LABEL "sbb cf=0"
    VALUES sbb, clc
    LABEL "sbb cf=1"
    VALUES sbb, stc
    LABEL "cmp"
    VALUES cmp, nop
    LABEL "inc"
    UVALUES inc, stc
    UVALUES inc, clc
    LABEL "dec"
    UVALUES dec, stc
    UVALUES dec, clc
    LABEL "neg"
    UVALUES neg, nop
    LABEL "not"
    UVALUES not, stc

    LABEL "immediates"
    mov rax, 0x7ffffffffffffff0
    add rax, 0x10
    DUMP
    add rax, -1                 ; imm8 sign-extended to 64 bits
    DUMP
    sub rax, 0x7fffffff         ; imm32 sign-extended
    DUMP
    add rax, -0x80000000
    DUMP
    mov eax, 0xfffffff0
    add eax, 0x20
    DUMP
    cmp rax, -1
    DUMP
    adc r10, 0x100
    DUMP
    sbb r11w, 0x7fff
    DUMP

    LABEL "memory operands"
    mov qword [m64], -16
    add qword [m64], 0x20
    mov rax, [m64]
    DUMP
    sub [m64], rax
    mov rbx, [m64]
    DUMP
    mov dword [m32], 0x7fffffff
    inc dword [m32]
    mov ecx, [m32]
    DUMP
    mov byte [m8], 0x80
    neg byte [m8]
    movzx edx, byte [m8]
    DUMP
    mov rsi, 3
    add rsi, [m64]
    DUMP
    cmp qword [m64], 0
    DUMP

    LABEL "byte registers spl bpl sil dil and ah"
    mov rsi, 0x1234
    mov rdi, 0x80
    add sil, dil
    DUMP
    mov rax, 0x1234
    add ah, al
    DUMP
    mov rbp, 0xff
    inc bpl
    xor ebp, ebp
    DUMP
    EXIT 0
