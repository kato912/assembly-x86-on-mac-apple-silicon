; exit: 139
; Differential test: a DOS interrupt (int 21h) is not allowed (SIGSEGV).
%include "dump32.inc"
section .text
global _start
_start:
    INIT
    LABEL "before"
    mov ah, 9
    int 21h
    EXIT 0
