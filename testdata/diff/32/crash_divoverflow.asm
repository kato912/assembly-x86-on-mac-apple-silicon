; exit: 136
; Differential test: div with a quotient too big for eax (edx not cleared)
; raises the divide error (SIGFPE).
%include "dump32.inc"
section .text
global _start
_start:
    INIT
    LABEL "before"
    mov eax, 10
    mov edx, 5
    mov ebx, 2
    div ebx
    EXIT 0
