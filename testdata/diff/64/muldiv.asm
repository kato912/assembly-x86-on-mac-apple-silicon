; Differential test: mul, imul (1/2/3 operands), div, idiv on all sizes.
; After mul/imul only CF and OF are defined; after div/idiv no flags are.
default rel
%include "dump64.inc"

; M1 op, a, b: one-operand multiply on every size (rdx preloaded so its
; partial overwrite is visible)
%macro M1 3
    mov rax, %2
    mov rbx, %3
    mov rdx, 0x1111111111111111
    %1 bl
    DUMP FL_MUL
    mov rax, %2
    mov rbx, %3
    mov rdx, 0x1111111111111111
    %1 bx
    DUMP FL_MUL
    mov rax, %2
    mov rbx, %3
    mov rdx, 0x1111111111111111
    %1 ebx
    DUMP FL_MUL
    mov rax, %2
    mov rbx, %3
    mov rdx, 0x1111111111111111
    %1 rbx
    DUMP FL_MUL
%endmacro

%macro MVALUES 1
    M1 %1, 0, 0
    M1 %1, 3, 4
    M1 %1, -1, -1
    M1 %1, -1, 2
    M1 %1, 0x7f, 0x7f
    M1 %1, 0x80, 0xff
    M1 %1, 0x10, 0x10
    M1 %1, 0x7fff, 2
    M1 %1, 0x8000, 0xffff
    M1 %1, 0x10000, 0x10000
    M1 %1, 0x7fffffff, 0x7fffffff
    M1 %1, 0x80000000, 2
    M1 %1, 0x100000000, 0x100000000
    M1 %1, 0x7fffffffffffffff, 2
    M1 %1, 0x8000000000000000, -1
    M1 %1, 0x123456789abcdef, 0xfedcba987654321
%endmacro

; I2 a, b: two- and three-operand imul
%macro I2 2
    mov rax, %1
    mov rbx, %2
    imul ax, bx
    DUMP FL_MUL
    mov rax, %1
    mov rbx, %2
    imul eax, ebx
    DUMP FL_MUL
    mov rax, %1
    mov rbx, %2
    imul rax, rbx
    DUMP FL_MUL
    mov rbx, %1
    imul rcx, rbx, -3
    DUMP FL_MUL
    imul ecx, ebx, 1000
    DUMP FL_MUL
    imul r9w, bx, 0x100
    DUMP FL_MUL
%endmacro

; D op, size-register-suffix values: divide rdx:rax by rbx at each size.
; Values are chosen so no division overflows.
%macro D8 3                     ; op, ax, divisor
    mov rax, %2
    mov rbx, %3
    %1 bl
    DUMP FL_NONE
%endmacro
%macro D16 4                    ; op, dx, ax, divisor
    mov rdx, %2
    mov rax, %3
    mov rbx, %4
    %1 bx
    DUMP FL_NONE
%endmacro
%macro D32 4
    mov rdx, %2
    mov rax, %3
    mov rbx, %4
    %1 ebx
    DUMP FL_NONE
%endmacro
%macro D64 4
    mov rdx, %2
    mov rax, %3
    mov rbx, %4
    %1 rbx
    DUMP FL_NONE
%endmacro

section .data
mq: dq 7
md: dd -3

section .text
global _start
_start:
    INIT
    LABEL "mul"
    MVALUES mul
    LABEL "imul one operand"
    MVALUES imul
    LABEL "imul two/three operands"
    I2 3, 4
    I2 -5, 7
    I2 0x7fff, 2
    I2 0x4000, 4
    I2 0x7fffffff, 2
    I2 0x40000000, -2
    I2 0x123456789, 0x10
    I2 0x4000000000000000, 2
    I2 -1, -1
    LABEL "imul memory"
    mov rax, 6
    imul rax, [mq]
    DUMP FL_MUL
    imul eax, [md], 100
    DUMP FL_MUL
    mov rax, 5
    mul qword [mq]
    DUMP FL_MUL

    LABEL "div"
    D8 div, 100, 7
    D8 div, 0xff, 1
    D8 div, 0x1ff, 2
    D8 div, 0xfeff, 0xff
    D16 div, 0, 1000, 7
    D16 div, 1, 0, 2
    D16 div, 0xfffe, 0xffff, 0xffff
    D32 div, 0, 1000000, 7
    D32 div, 5, 10, 6
    D32 div, 0xfffffffe, 0xffffffff, 0xffffffff
    D64 div, 0, 1000000000000, 7
    D64 div, 1, 0, 2
    D64 div, 0x123456789, 0xabcdef0123456789, 0xfedcba9876543210
    LABEL "idiv"
    D8 idiv, -100, 7
    D8 idiv, 100, -7
    D8 idiv, -128, 1
    D8 idiv, 0x7f, -1
    D16 idiv, -1, -1000, 7
    D16 idiv, 0, 1000, -7
    D16 idiv, -1, 0x8000, 1
    D32 idiv, -1, -1000000, 7
    D32 idiv, 0, 1000000, -7
    D32 idiv, -1, 0x80000000, 1
    D32 idiv, -1, 0x80000001, -1
    D64 idiv, -1, -1000000000000, 7
    D64 idiv, 0, 1000000000000, -7
    D64 idiv, -1, 0x8000000000000000, 1
    D64 idiv, -1, -9, 2
    LABEL "div memory"
    xor edx, edx
    mov rax, 100
    div qword [mq]
    DUMP FL_NONE
    mov eax, 100
    cdq
    idiv dword [md]
    DUMP FL_NONE
    EXIT 0
