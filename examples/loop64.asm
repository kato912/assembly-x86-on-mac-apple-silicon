; loop64.asm — multiplication table of 7 (64-bit)
;   run:  nasmx86 loop64.asm

section .data
    times_  db  " x 7 = "
    tlen    equ $ - times_

section .bss
    digits  resb 24

section .text
    global _start

_start:
    mov     r12, 1              ; r12..r15 survive our calls (callee-saved)
.loop:
    mov     rax, r12
    call    print_number_nonl   ; "i"
    mov     rax, 1
    mov     rdi, 1
    mov     rsi, times_
    mov     rdx, tlen
    syscall                     ; " x 7 = "
    imul    rax, r12, 7
    call    print_number        ; "i*7\n"
    inc     r12
    cmp     r12, 12
    jle     .loop

    mov     rax, 60
    xor     edi, edi
    syscall

print_number:                   ; print rax and a newline
    mov     r8b, 10
    jmp     print_common
print_number_nonl:              ; print rax without a newline
    xor     r8d, r8d
print_common:
    lea     rsi, [digits + 23]
    mov     [rsi], r8b
    mov     rcx, 10
.digit:
    xor     edx, edx
    div     rcx
    add     dl, '0'
    dec     rsi
    mov     [rsi], dl
    test    rax, rax
    jnz     .digit
    lea     rdx, [digits + 23]
    sub     rdx, rsi
    test    r8b, r8b
    jz      .write
    inc     rdx                 ; include the newline
.write:
    mov     rax, 1
    mov     rdi, 1
    syscall
    ret
