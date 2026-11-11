; io.asm — output helpers used by main.asm (32-bit, int 0x80).

global print_str, print_uint

section .bss
    digits  resb 12

section .text

; print_str: write edx bytes from ecx to stdout. Preserves all registers.
print_str:
    pushad
    mov     eax, 4
    mov     ebx, 1
    int     0x80
    popad
    ret

; print_uint: print eax as an unsigned decimal number and a newline.
print_uint:
    pushad
    mov     edi, digits + 11
    mov     byte [edi], 10
    mov     ebx, 10
.digit:
    xor     edx, edx
    div     ebx
    add     dl, '0'
    dec     edi
    mov     [edi], dl
    test    eax, eax
    jnz     .digit
    mov     ecx, edi
    mov     edx, digits + 12
    sub     edx, edi
    call    print_str
    popad
    ret
