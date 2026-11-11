; Differential test: mul, imul (1/2/3 operands), div, idiv on 8/16/32 bits.
%include "dump32.inc"

%macro M1 3
    mov eax, %2
    mov ebx, %3
    mov edx, 0x11111111
    %1 bl
    DUMP FL_MUL
    mov eax, %2
    mov ebx, %3
    mov edx, 0x11111111
    %1 bx
    DUMP FL_MUL
    mov eax, %2
    mov ebx, %3
    mov edx, 0x11111111
    %1 ebx
    DUMP FL_MUL
%endmacro

%macro MVALUES 1
    M1 %1, 0, 0
    M1 %1, 3, 4
    M1 %1, -1, -1
    M1 %1, -1, 2
    M1 %1, 0x7f, 0x7f
    M1 %1, 0x80, 0xff
    M1 %1, 0x7fff, 2
    M1 %1, 0x8000, 0xffff
    M1 %1, 0x10000, 0x10000
    M1 %1, 0x7fffffff, 0x7fffffff
    M1 %1, 0x80000000, 2
    M1 %1, 0x12345678, 0x9abcdef0
%endmacro

%macro I2 2
    mov eax, %1
    mov ebx, %2
    imul ax, bx
    DUMP FL_MUL
    mov eax, %1
    mov ebx, %2
    imul eax, ebx
    DUMP FL_MUL
    imul ecx, ebx, -3
    DUMP FL_MUL
    imul edx, ebx, 100000
    DUMP FL_MUL
    imul si, bx, 0x100
    DUMP FL_MUL
%endmacro

%macro D 5                      ; op, operand, edx/dx/ah, eax/ax/al, divisor
    mov edx, %3
    mov eax, %4
    mov ebx, %5
    %1 %2
    DUMP FL_NONE
%endmacro

section .data
md: dd -7

section .text
global _start
_start:
    INIT
    LABEL "mul"
    MVALUES mul
    LABEL "imul"
    MVALUES imul
    LABEL "imul 2/3 operands"
    I2 3, 4
    I2 -5, 7
    I2 0x7fff, 2
    I2 0x7fffffff, 2
    I2 0x40000000, -2
    I2 -1, -1
    mov eax, 6
    imul eax, [md]
    DUMP FL_MUL
    imul ebx, [md], 3
    DUMP FL_MUL
    LABEL "div"
    D div, bl, 0, 100, 7
    D div, bl, 0, 0x1ff, 2
    D div, bl, 0, 0xfeff, 0xff
    D div, bx, 0, 1000, 7
    D div, bx, 1, 0, 2
    D div, bx, 0xfffe, 0xffff, 0xffff
    D div, ebx, 0, 1000000, 7
    D div, ebx, 5, 10, 6
    D div, ebx, 0xfffffffe, 0xffffffff, 0xffffffff
    LABEL "idiv"
    D idiv, bl, 0, -100, 7
    D idiv, bl, 0, 100, -7
    D idiv, bl, 0, -128, 1
    D idiv, bx, -1, -1000, 7
    D idiv, bx, 0, 1000, -7
    D idiv, ebx, -1, -1000000, 7
    D idiv, ebx, 0, 1000000, -7
    D idiv, ebx, -1, 0x80000000, 1
    D idiv, ebx, -1, -9, 2
    mov eax, 100
    cdq
    idiv dword [md]
    DUMP FL_NONE
    EXIT 0
