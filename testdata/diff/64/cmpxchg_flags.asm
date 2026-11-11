; linux-only: Rosetta 2 computes cmpxchg's flags as dest - accumulator, real
; x86 CPUs as accumulator - dest (like cmp), so only native Linux can judge.
default rel
%include "dump64.inc"

%macro CX 2                     ; accumulator, destination (not equal)
    FLAGS 0x202
    mov rax, %1
    mov rbx, %2
    mov rcx, 7
    cmpxchg rbx, rcx
    DUMP
    FLAGS 0x202
    mov rax, %1
    mov [mem], rbx
    cmpxchg [mem], ecx
    DUMP
    FLAGS 0x202
    mov rax, %1
    mov rbx, %2
    cmpxchg bl, cl
    DUMP
%endmacro

section .data
mem: dq 0

section .text
global _start
_start:
    INIT
    LABEL "cmpxchg flags (not equal)"
    CX 1, 99
    CX 99, 1
    CX 0x80, 0x7f
    CX 0x8000000000000000, 1
    CX -1, 0
    EXIT 0
