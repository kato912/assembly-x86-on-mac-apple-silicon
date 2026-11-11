; linux-only
; exit: 42
; Differential test: the exit status reaches the shell (exit_group).
default rel
%include "dump64.inc"
section .text
global _start
_start:
    INIT
    LABEL "exit_group(42)"
    mov eax, 231
    mov edi, 42
    syscall
