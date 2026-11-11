; exit: 133
; Differential test: int3 stops the program with SIGTRAP.
%include "dump32.inc"
section .text
global _start
_start:
    INIT
    LABEL "before"
    int3
    EXIT 0
