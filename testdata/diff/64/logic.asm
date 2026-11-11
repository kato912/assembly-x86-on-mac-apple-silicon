; Differential test: and, or, xor, test on all operand sizes, immediates
; and memory. AF is undefined after logic instructions and is masked out.
default rel
%include "dump64.inc"

%macro T 3
    mov rax, %2
    mov rbx, %3
    %1 al, bl
    DUMP FL_LOGIC
    mov rax, %2
    mov rbx, %3
    %1 ax, bx
    DUMP FL_LOGIC
    mov rax, %2
    mov rbx, %3
    %1 eax, ebx
    DUMP FL_LOGIC
    mov rax, %2
    mov rbx, %3
    %1 rax, rbx
    DUMP FL_LOGIC
    mov r12, %2
    mov r13, %3
    %1 r12d, r13d
    DUMP FL_LOGIC
%endmacro

%macro VALUES 1
    T %1, 0, 0
    T %1, 0, -1
    T %1, -1, -1
    T %1, 0x0f, 0xf0
    T %1, 0x80, 0x80
    T %1, 0x8000, 0xff00
    T %1, 0x80000000, 0x7fffffff
    T %1, 0x8000000000000000, 0x8000000000000001
    T %1, 0x123456789abcdef0, 0x0f0f0f0f0f0f0f0f
    T %1, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa
    T %1, 3, 5
%endmacro

section .data
m: dq 0x00ff00ff00ff00ff

section .text
global _start
_start:
    INIT
    LABEL "and"
    VALUES and
    LABEL "or"
    VALUES or
    LABEL "xor"
    VALUES xor
    LABEL "test"
    VALUES test

    LABEL "immediates"
    mov rax, -1
    and rax, 0x7f0              ; imm32
    DUMP FL_LOGIC
    or rax, -8                  ; imm8 sign-extended
    DUMP FL_LOGIC
    xor eax, 0x80000000
    DUMP FL_LOGIC
    test al, 0x80
    DUMP FL_LOGIC
    and ax, 0x00ff
    DUMP FL_LOGIC
    test rax, rax
    DUMP FL_LOGIC

    LABEL "memory"
    mov rcx, [m]
    and qword [m], 0x0f0f
    mov rdx, [m]
    DUMP FL_LOGIC
    or dword [m], 0xf0000000
    mov rdx, [m]
    DUMP FL_LOGIC
    xor byte [m], 0xff
    mov rdx, [m]
    DUMP FL_LOGIC
    test [m], rcx
    DUMP FL_LOGIC
    not qword [m]
    mov rdx, [m]
    DUMP FL_LOGIC
    xor rdx, rdx
    DUMP FL_LOGIC
    EXIT 0
