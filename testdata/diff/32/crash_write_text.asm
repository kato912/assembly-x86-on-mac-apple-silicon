; exit: 139
; Differential test: writing to the read-only .text section (SIGSEGV).
%include "dump32.inc"
section .text
global _start
_start:
    INIT
    LABEL "before"
    mov dword [counter], 1
    EXIT 0
counter: dd 0
