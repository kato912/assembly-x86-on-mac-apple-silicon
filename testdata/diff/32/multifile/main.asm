; Differential test: a program split over two files, with global/extern in
; both directions, shared data, and calls into the other file.
%include "dump32.inc"
extern square, sum_array, bump, counter, table, table_len
global base_value

section .data
base_value dd 1000

section .text
global _start
_start:
    INIT
    LABEL "calls into mathlib.asm"
    mov eax, 12
    call square                 ; eax = 144
    DUMP FL_MUL                 ; (imul leaves SF/ZF/PF/AF undefined)
    mov esi, table
    mov ecx, [table_len]
    call sum_array              ; eax = 31
    xor esi, esi
    DUMP FL_NONE                ; (flags come from an address computation)
    LABEL "data shared both ways"
    call bump                   ; counter += base_value (defined here)
    call bump
    mov ebx, [counter]          ; 2005
    mov edx, [table + 4*7]      ; extern array, indexed
    DUMP
    EXIT 0
