; Differential test: and, or, xor, test, not on 8/16/32-bit operands.
%include "dump32.inc"

%macro T 3
    mov eax, %2
    mov ebx, %3
    %1 al, bl
    DUMP FL_LOGIC
    mov eax, %2
    mov ebx, %3
    %1 ax, bx
    DUMP FL_LOGIC
    mov eax, %2
    mov ebx, %3
    %1 eax, ebx
    DUMP FL_LOGIC
    mov ecx, %2
    mov edx, %3
    %1 ch, dh
    DUMP FL_LOGIC
%endmacro

%macro VALUES 1
    T %1, 0, 0
    T %1, 0, -1
    T %1, -1, -1
    T %1, 0x0f0f, 0xf0f0
    T %1, 0x8080, 0x8080
    T %1, 0x80000000, 0x7fffffff
    T %1, 0x12345678, 0x0f0f0f0f
    T %1, 0x55555555, 0xaaaaaaaa
    T %1, 0x0300, 0x0500
%endmacro

section .data
m: dd 0x00ff00ff

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
    LABEL "immediates and memory"
    mov eax, -1
    and eax, 0x7f0
    DUMP FL_LOGIC
    or eax, -8
    DUMP FL_LOGIC
    xor al, 0x81
    DUMP FL_LOGIC
    test ax, 0x8000
    DUMP FL_LOGIC
    and dword [m], 0x0f0f
    mov ebx, [m]
    DUMP FL_LOGIC
    or byte [m + 1], 0x80
    mov ebx, [m]
    DUMP FL_LOGIC
    not word [m]
    mov ebx, [m]
    DUMP FL_LOGIC
    test [m], ebx
    DUMP FL_LOGIC
    EXIT 0
