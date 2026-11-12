; printf64.asm — call the C library (printf/scanf) from assembly, 64-bit,
; in the style of "x86-64 Assembly Language Programming with Ubuntu".
;   run:  echo 6 7 | nasmx86 printf64.asm
; On Linux the same file builds with:
;   nasm -f elf64 printf64.asm && gcc -no-pie printf64.o -o printf64

extern printf, scanf

section .data
    prompt  db  "Enter two numbers: ", 0
    infmt   db  "%ld %ld", 0
    outfmt  db  "%ld x %ld = %ld", 10, 0

section .bss
    a       resq 1
    b       resq 1

section .text
    global main

main:
    push    rbp                 ; keep the stack 16-byte aligned for C calls
    mov     rbp, rsp

    mov     rdi, prompt         ; printf(prompt)
    xor     eax, eax            ; al = 0: no floating-point arguments
    call    printf

    mov     rdi, infmt          ; scanf("%ld %ld", &a, &b)
    mov     rsi, a
    mov     rdx, b
    xor     eax, eax
    call    scanf

    mov     rsi, [a]            ; printf(outfmt, a, b, a*b)
    mov     rdx, [b]
    mov     rcx, rsi
    imul    rcx, rdx
    mov     rdi, outfmt
    xor     eax, eax
    call    printf

    xor     eax, eax            ; return 0 from main
    pop     rbp
    ret
