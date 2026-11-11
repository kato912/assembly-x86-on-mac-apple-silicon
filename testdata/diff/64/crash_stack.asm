; linux-only
; exit: 139
; Differential test: unbounded recursion overflows the stack (SIGSEGV).
default rel
%include "dump64.inc"
section .text
global _start
_start:
    INIT
    LABEL "recursing forever"
recurse:
    push rax
    call recurse
