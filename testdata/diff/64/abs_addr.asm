; linux-only: absolute 32-bit addresses are not allowed in macOS binaries.
; Differential test: [disp32], label as immediate, [label + index*scale].
%include "dump64.inc"

section .data
arr:    dd 11, 22, 33, 44
q:      dq 0x0123456789abcdef

section .text
global _start
_start:
    INIT
    LABEL "absolute addressing"
    mov eax, [arr + 4]          ; [disp32]
    mov rbx, [q]
    mov esi, 2
    mov ecx, [arr + rsi*4]      ; [disp32 + index*scale]
    mov edx, [arr + rsi*4 + 4]
    mov edi, arr                ; address as a 32-bit immediate
    sub edi, arr
    mov r8, q                   ; ... and as a 64-bit immediate
    sub r8, q
    mov dword [arr], 99
    mov r9d, [arr]
    push qword arr              ; push of an address
    pop r10
    sub r10, arr
    DUMP
    EXIT 0
