; Differential test: add, adc, sub, sbb, cmp, inc, dec, neg, not
; on 8/16/32-bit registers, immediates and memory operands.
%include "dump32.inc"

; T op, a, b, pre, mask: load a/b into eax/ebx, run pre (clc/stc/nop),
; then op on al,bl / ax,bx / eax,ebx and dump after each.
%macro T 5
    mov eax, %2
    mov ebx, %3
    %4
    %1 al, bl
    DUMP %5
    mov eax, %2
    mov ebx, %3
    %4
    %1 ax, bx
    DUMP %5
    mov eax, %2
    mov ebx, %3
    %4
    %1 eax, ebx
    DUMP %5
%endmacro

%macro VALUES 3                 ; op, pre, mask
    T %1, 0, 0, %2, %3
    T %1, 1, 1, %2, %3
    T %1, 1, 2, %2, %3
    T %1, 0x7f, 1, %2, %3
    T %1, 0x80, 0x80, %2, %3
    T %1, 0xff, 1, %2, %3
    T %1, 0x0f, 0x01, %2, %3
    T %1, 0x10, 0x01, %2, %3
    T %1, 0x7fff, 1, %2, %3
    T %1, 0x8000, 0xffff, %2, %3
    T %1, 0x7fffffff, 1, %2, %3
    T %1, 0x80000000, 0x80000000, %2, %3
    T %1, 0xffffffff, 1, %2, %3
    T %1, 0xffffffff, 0xffffffff, %2, %3
    T %1, 0x12345678, 0x9abcdef0, %2, %3
    T %1, 0x9abcdef0, 0x12345678, %2, %3
    T %1, 5, 7, %2, %3
%endmacro

; U op, a, pre: unary op on al / ax / eax
%macro U 3
    mov eax, %2
    %3
    %1 al
    DUMP
    mov eax, %2
    %3
    %1 ax
    DUMP
    mov eax, %2
    %3
    %1 eax
    DUMP
%endmacro

%macro UVALUES 2
    U %1, 0, %2
    U %1, 1, %2
    U %1, 0x0f, %2
    U %1, 0x10, %2
    U %1, 0x7f, %2
    U %1, 0x80, %2
    U %1, 0xff, %2
    U %1, 0x7fff, %2
    U %1, 0x8000, %2
    U %1, 0xffff, %2
    U %1, 0x7fffffff, %2
    U %1, 0x80000000, %2
    U %1, 0xffffffff, %2
%endmacro

section .data
m8:  db 0
m16: dw 0
m32: dd 0

section .text
global _start
_start:
    INIT
    LABEL "add"
    VALUES add, nop, FL_ALL
    LABEL "adc cf=0"
    VALUES adc, clc, FL_ALL
    LABEL "adc cf=1"
    VALUES adc, stc, FL_ALL
    LABEL "sub"
    VALUES sub, nop, FL_ALL
    LABEL "sbb cf=0"
    VALUES sbb, clc, FL_ALL
    LABEL "sbb cf=1"
    VALUES sbb, stc, FL_ALL
    LABEL "cmp"
    VALUES cmp, nop, FL_ALL
    LABEL "inc (cf preserved)"
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
    mov eax, 0x7ffffff0
    add eax, 0x10               ; imm8 form
    DUMP
    add eax, 0x12345            ; imm32 form
    DUMP
    sub eax, -1                 ; sign-extended imm8
    DUMP
    mov ax, 0xfff0
    add ax, 0x20
    DUMP
    mov al, 0x7f
    add al, 1
    DUMP
    cmp eax, -1
    DUMP
    cmp al, 0x80
    DUMP
    adc eax, 0x100
    DUMP
    sbb ax, 0x7fff
    DUMP

    LABEL "memory operands"
    mov dword [m32], 0xfffffff0
    add dword [m32], 0x20
    DUMP
    mov eax, [m32]
    DUMP
    sub [m32], eax
    DUMP
    mov ebx, 3
    add ebx, [m32]
    DUMP
    mov byte [m8], 0x80
    sub byte [m8], 1
    DUMP
    movzx ecx, byte [m8]
    DUMP
    mov word [m16], 0x7fff
    inc word [m16]
    DUMP
    movzx edx, word [m16]
    DUMP
    neg dword [m32]
    DUMP
    not byte [m8]
    movzx esi, byte [m8]
    DUMP
    cmp dword [m32], 0
    DUMP
    dec dword [m32]
    mov edi, [m32]
    DUMP

    LABEL "8-bit high registers"
    mov eax, 0x1234
    add ah, al
    DUMP
    mov ebx, 0xff80
    sub bh, bl
    DUMP
    mov ecx, 0x0102
    add ch, cl
    DUMP
    mov edx, 0xffff
    inc dh
    DUMP
    EXIT 0
