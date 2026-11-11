; Differential test: movs/stos/lods/cmps/scas in every size, with rep,
; repe and repne, forwards and backwards (std). Pointers are rebased to
; offsets before each dump.
default rel
%include "dump64.inc"

; OFFSI a / OFFDI b: turn rsi / rdi into an offset from label a / b so
; dumps are comparable (flags and rax are preserved)
%macro OFFSI 1
    pushfq
    push rax
    lea rax, [%1]
    sub rsi, rax
    pop rax
    popfq
%endmacro
%macro OFFDI 1
    pushfq
    push rax
    lea rax, [%1]
    sub rdi, rax
    pop rax
    popfq
%endmacro
%macro OFFSETS 2
    OFFSI %1
    OFFDI %2
    xor eax, eax
%endmacro
%define REBASE OFFSETS src, dst

section .data
src:    db "The quick brown fox jumps over the lazy dog", 0
srclen  equ $ - src
other:  db "The quick brown fox jumps over the lazy cat", 0
nl:     db 10
section .bss
dst:    resb 128

section .text
global _start
_start:
    INIT
    LABEL "rep movsb"
    lea rsi, [src]
    lea rdi, [dst]
    mov ecx, srclen
    rep movsb
    PRINTMEM dst, srclen - 1
    PRINTMEM nl, 1
    REBASE
    DUMP

    LABEL "movsw/movsd/movsq"
    lea rsi, [src]
    lea rdi, [dst]
    movsw
    movsd
    movsq
    REBASE
    DUMP
    lea rsi, [src]
    lea rdi, [dst + 64]
    mov ecx, 3
    rep movsq
    PRINTMEM dst, 88
    PRINTMEM nl, 1
    lea rax, [dst + 64]
    sub rdi, rax
    lea rax, [src]
    sub rsi, rax
    xor eax, eax
    DUMP

    LABEL "rep stos"
    lea rdi, [dst]
    mov al, '*'
    mov ecx, 10
    rep stosb
    mov ax, '-+'
    mov ecx, 4
    rep stosw
    mov eax, 'abcd'
    stosd
    mov rax, 'ABCDEFGH'
    stosq
    PRINTMEM dst, 30
    PRINTMEM nl, 1
    lea rax, [dst]
    sub rdi, rax
    xor eax, eax
    DUMP

    LABEL "lods"
    lea rsi, [src]
    lodsb
    mov ebx, eax
    lodsw
    mov ecx, eax
    lodsd
    mov edx, eax
    lodsq
    mov r8, rax
    xor eax, eax
    OFFSI src
    DUMP

    LABEL "repe cmpsb / repne scasb"
    lea rsi, [src]
    lea rdi, [other]
    mov ecx, srclen
    repe cmpsb                  ; stops after the first difference
    OFFSETS src, other
    DUMP
    lea rdi, [src]              ; strlen
    xor al, al
    mov rcx, -1
    repne scasb
    not rcx
    dec rcx
    OFFDI src
    DUMP
    lea rdi, [src]              ; find 'x'
    mov al, 'x'
    mov ecx, srclen
    repne scasb
    OFFDI src
    DUMP
    lea rdi, [src]              ; search for a missing byte: rcx runs out
    mov al, '#'
    mov ecx, 5
    repne scasb
    OFFDI src
    DUMP
    lea rsi, [src]              ; cmpsw/cmpsd/cmpsq single steps
    lea rdi, [other]
    cmpsq
    OFFSETS src, other
    DUMP
    lea rsi, [src + 36]
    lea rdi, [other + 36]
    cmpsd
    OFFSETS src, other
    DUMP
    lea rsi, [src + 40]
    lea rdi, [other + 40]
    cmpsw
    OFFSETS src, other
    DUMP

    LABEL "rep with rcx = 0"
    lea rsi, [src]
    lea rdi, [dst]
    xor ecx, ecx
    FLAGS 0x8d7
    rep movsb
    repe cmpsb                  ; flags untouched when rcx = 0
    REBASE
    DUMP

    LABEL "backwards (std)"
    lea rsi, [src + srclen - 2]
    lea rdi, [dst + srclen - 2]
    mov ecx, srclen - 1
    std
    rep movsb
    cld
    PRINTMEM dst, srclen - 1
    PRINTMEM nl, 1
    REBASE
    DUMP
    std
    lea rsi, [src + 8]
    lodsd
    lodsd
    mov ebx, eax
    lea rdi, [dst + 16]
    mov eax, 0x41424344
    stosd
    stosd
    cld
    PRINTMEM dst + 8, 16
    PRINTMEM nl, 1
    REBASE
    DUMP
    EXIT 0
