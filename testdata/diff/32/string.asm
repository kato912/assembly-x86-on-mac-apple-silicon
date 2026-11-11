; Differential test: string instructions with rep/repe/repne, both directions.
%include "dump32.inc"

%macro OFFSI 1
    pushfd
    sub esi, %1
    popfd
%endmacro
%macro OFFDI 1
    pushfd
    sub edi, %1
    popfd
%endmacro

section .data
src:    db "Hello, assembly world!", 0
srclen  equ $ - src
other:  db "Hello, assembly WORLD!", 0
nl:     db 10
section .bss
dst:    resb 64

section .text
global _start
_start:
    INIT
    LABEL "rep movsb / movsw / movsd"
    mov esi, src
    mov edi, dst
    mov ecx, srclen
    rep movsb
    PRINTMEM dst, srclen - 1
    PRINTMEM nl, 1
    OFFSI src
    OFFDI dst
    DUMP
    mov esi, src
    mov edi, dst + 32
    mov ecx, 3
    rep movsd
    movsw
    PRINTMEM dst + 32, 14
    PRINTMEM nl, 1
    OFFSI src
    OFFDI dst
    DUMP
    LABEL "rep stosb / stosw / stosd"
    mov edi, dst
    mov al, '='
    mov ecx, 6
    rep stosb
    mov ax, '<>'
    mov ecx, 2
    rep stosw
    mov eax, '1234'
    stosd
    PRINTMEM dst, 14
    PRINTMEM nl, 1
    OFFDI dst
    DUMP
    LABEL "lods"
    mov esi, src
    lodsb
    mov ebx, eax
    lodsw
    mov ecx, eax
    lodsd
    mov edx, eax
    OFFSI src
    DUMP
    LABEL "cmps / scas"
    mov esi, src
    mov edi, other
    mov ecx, srclen
    repe cmpsb
    OFFSI src
    OFFDI other
    DUMP
    mov edi, src
    xor eax, eax
    mov ecx, -1
    repne scasb
    not ecx
    OFFDI src
    DUMP
    mov edi, src
    mov al, 'w'
    mov ecx, srclen
    repne scasb
    OFFDI src
    DUMP
    mov esi, src
    mov edi, other
    cmpsd
    OFFSI src
    OFFDI other
    DUMP
    LABEL "backwards"
    std
    mov esi, src + srclen - 2
    mov edi, dst + srclen - 2
    mov ecx, srclen - 1
    rep movsb
    cld
    PRINTMEM dst, srclen - 1
    PRINTMEM nl, 1
    OFFSI src
    OFFDI dst
    DUMP
    EXIT 0
