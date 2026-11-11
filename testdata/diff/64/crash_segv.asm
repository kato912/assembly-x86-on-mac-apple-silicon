; linux-only
; exit: 139
; Differential test: reading address 0 kills the program with SIGSEGV.
default rel
%include "dump64.inc"
section .text
global _start
_start:
    INIT
    LABEL "before the crash"
    xor ebx, ebx
    mov rax, [rbx]
    EXIT 0
