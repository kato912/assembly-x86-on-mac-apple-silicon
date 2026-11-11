; Differential test: a 64-bit program split over three files (code in
; strlib.asm, data in data.asm), all symbols RIP-relative.
default rel
%include "dump64.inc"
extern strlen_, upcase, message, msg_copy

section .text
global _start
_start:
    INIT
    LABEL "strlib.asm + data.asm"
    lea rdi, [message]
    call strlen_                ; rax = 12
    mov rbx, rax
    lea rdi, [message]
    lea rsi, [msg_copy]
    call upcase
    PRINTMEM msg_copy, 12
    LABEL ""
    xor esi, esi
    xor edi, edi
    DUMP
    EXIT 0
