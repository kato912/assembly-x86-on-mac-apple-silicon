; linux-only
; exit: 136
; Differential test: division by zero kills the program with SIGFPE.
default rel
%include "dump64.inc"
section .text
global _start
_start:
    INIT
    LABEL "before the crash"
    mov eax, 1
    xor ecx, ecx
    div rcx
    LABEL "not reached"
    EXIT 0
