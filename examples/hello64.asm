; hello64.asm — Hello, World! for 64-bit Linux (syscall)
;   run:  nasmx86 hello64.asm

section .data
    msg     db  "Hello, World!", 10
    len     equ $ - msg

section .text
    global _start

_start:
    mov     rax, 1          ; sys_write
    mov     rdi, 1          ; fd 1 = stdout
    mov     rsi, msg        ; buffer
    mov     rdx, len        ; number of bytes
    syscall

    mov     rax, 60         ; sys_exit
    xor     rdi, rdi        ; exit code 0
    syscall
