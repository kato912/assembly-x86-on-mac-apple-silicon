; factorial32.asm — read n, print n! using recursion (32-bit)
;   run:  echo 5 | nasmx86 factorial32.asm
;   try:  echo 4 | nasmx86 --trace factorial32.asm

section .data
    prompt  db  "n = "
    plen    equ $ - prompt
    msg     db  "n! = "
    mlen    equ $ - msg

section .bss
    input   resb 16
    digits  resb 12

section .text
    global _start

_start:
    mov     eax, 4              ; write(1, prompt, plen)
    mov     ebx, 1
    mov     ecx, prompt
    mov     edx, plen
    int     0x80

    mov     eax, 3              ; read(0, input, 16)
    mov     ebx, 0
    mov     ecx, input
    mov     edx, 16
    int     0x80

    ; convert the decimal text to a number in eax
    xor     eax, eax
    mov     esi, input
.parse:
    movzx   ecx, byte [esi]
    sub     ecx, '0'
    cmp     ecx, 9
    ja      .parsed             ; stop at the first non-digit (unsigned trick)
    imul    eax, eax, 10
    add     eax, ecx
    inc     esi
    jmp     .parse
.parsed:
    push    eax
    call    factorial           ; eax = n!
    add     esp, 4
    mov     edi, eax

    mov     eax, 4
    mov     ebx, 1
    mov     ecx, msg
    mov     edx, mlen
    int     0x80
    mov     eax, edi
    call    print_number

    mov     eax, 1
    xor     ebx, ebx
    int     0x80

; unsigned factorial(unsigned n): n <= 1 ? 1 : n * factorial(n - 1)
factorial:
    push    ebp
    mov     ebp, esp
    mov     eax, [ebp + 8]      ; n
    cmp     eax, 1
    jbe     .base
    dec     eax
    push    eax
    call    factorial           ; eax = (n-1)!
    add     esp, 4
    mul     dword [ebp + 8]     ; edx:eax = (n-1)! * n
    jmp     .done
.base:
    mov     eax, 1
.done:
    pop     ebp
    ret

print_number:                   ; print eax (unsigned) and a newline
    pushad
    mov     esi, digits + 11
    mov     byte [esi], 10
    mov     ebx, 10
.digit:
    xor     edx, edx
    div     ebx
    add     dl, '0'
    dec     esi
    mov     [esi], dl
    test    eax, eax
    jnz     .digit
    mov     ecx, esi
    mov     edx, digits + 12
    sub     edx, esi
    mov     eax, 4
    mov     ebx, 1
    int     0x80
    popad
    ret
