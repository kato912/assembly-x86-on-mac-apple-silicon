; examples/multifile — a program split over two files (32-bit).
;   run:  nasmx86 main.asm io.asm
;
; io.asm provides print_str and print_uint; they are marked "global" there
; and "extern" here, exactly like when linking with ld.

extern print_str, print_uint

section .data
    title   db  "sum of 1..100 = "
    tlen    equ $ - title

section .text
    global _start

_start:
    mov     ecx, title          ; print_str(ecx = text, edx = length)
    mov     edx, tlen
    call    print_str

    xor     eax, eax            ; eax = 1 + 2 + ... + 100
    mov     ecx, 100
.add:
    add     eax, ecx
    loop    .add
    call    print_uint          ; prints eax and a newline

    mov     eax, 1              ; exit(0)
    xor     ebx, ebx
    int     0x80
