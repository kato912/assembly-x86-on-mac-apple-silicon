; factorial64.asm — read n, print n! using recursion (64-bit, System V ABI)
;   run:  echo 20 | nasmx86 factorial64.asm

section .data
    prompt  db  "n = "
    plen    equ $ - prompt
    msg     db  "n! = "
    mlen    equ $ - msg

section .bss
    input   resb 16
    digits  resb 24

section .text
    global _start

_start:
    mov     rax, 1              ; write(1, prompt, plen)
    mov     rdi, 1
    mov     rsi, prompt
    mov     rdx, plen
    syscall

    mov     rax, 0              ; read(0, input, 16)
    mov     rdi, 0
    mov     rsi, input
    mov     rdx, 16
    syscall

    xor     edi, edi            ; parse decimal -> rdi
    mov     rsi, input
.parse:
    movzx   eax, byte [rsi]
    sub     eax, '0'
    cmp     eax, 9
    ja      .parsed
    imul    rdi, rdi, 10
    add     rdi, rax
    inc     rsi
    jmp     .parse
.parsed:
    call    factorial           ; first argument in rdi, result in rax
    mov     rbx, rax

    mov     rax, 1
    mov     rdi, 1
    mov     rsi, msg
    mov     rdx, mlen
    syscall
    mov     rax, rbx
    call    print_number

    mov     rax, 60             ; exit(0)
    xor     edi, edi
    syscall

; uint64 factorial(uint64 n)
factorial:
    cmp     rdi, 1
    jbe     .base
    push    rdi                 ; save n across the recursive call
    dec     rdi
    call    factorial           ; rax = (n-1)!
    pop     rdi
    imul    rax, rdi            ; rax = n * (n-1)!
    ret
.base:
    mov     eax, 1
    ret

print_number:                   ; print rax (unsigned) and a newline
    mov     rsi, digits + 23
    mov     byte [rsi], 10
    mov     rcx, 10
.digit:
    xor     edx, edx
    div     rcx
    add     dl, '0'
    dec     rsi
    mov     [rsi], dl
    test    rax, rax
    jnz     .digit
    mov     rdx, digits + 24
    sub     rdx, rsi            ; length
    mov     rax, 1
    mov     rdi, 1
    syscall
    ret
