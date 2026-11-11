; input32.asm — read a line from the keyboard and greet the user (32-bit)
;   run:  nasmx86 input32.asm
;   or:   echo Somchai | nasmx86 input32.asm

section .data
    ask     db  "What is your name? "
    asklen  equ $ - ask
    hello   db  "Hello, "
    hellen  equ $ - hello

section .bss
    name    resb 64             ; buffer for the input

section .text
    global _start

_start:
    ; write(1, ask, asklen)
    mov     eax, 4
    mov     ebx, 1
    mov     ecx, ask
    mov     edx, asklen
    int     0x80

    ; n = read(0, name, 64)    -> eax = number of bytes read (incl. newline)
    mov     eax, 3
    mov     ebx, 0
    mov     ecx, name
    mov     edx, 64
    int     0x80
    mov     esi, eax            ; keep the length

    ; write(1, hello, hellen)
    mov     eax, 4
    mov     ebx, 1
    mov     ecx, hello
    mov     edx, hellen
    int     0x80

    ; write(1, name, n)
    mov     eax, 4
    mov     ebx, 1
    mov     ecx, name
    mov     edx, esi
    int     0x80

    ; exit(0)
    mov     eax, 1
    mov     ebx, 0
    int     0x80
