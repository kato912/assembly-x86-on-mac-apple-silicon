; hello32.asm — Hello, World! for 32-bit Linux (int 0x80)
;   run:  nasmx86 hello32.asm

section .data
    msg     db  "Hello, World!", 10     ; 10 = newline
    len     equ $ - msg                 ; length of msg in bytes

section .text
    global _start

_start:
    mov     eax, 4          ; sys_write
    mov     ebx, 1          ; fd 1 = stdout
    mov     ecx, msg        ; buffer
    mov     edx, len        ; number of bytes
    int     0x80

    mov     eax, 1          ; sys_exit
    mov     ebx, 0          ; exit code 0
    int     0x80
