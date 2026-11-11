; function32.asm — calling a function with arguments on the stack (cdecl, 32-bit)
; exit: 24   (the result is returned as the exit code)
;   run:  nasmx86 function32.asm      (exit code = result, check with: echo $?)
;   try:  nasmx86 --step function32.asm  and use  x  to watch the stack

section .text
    global _start

_start:
    push    7                   ; second argument  (pushed first)
    push    5                   ; first argument
    call    add_mul             ; eax = (5 + 7) * 2 = 24
    add     esp, 8              ; caller removes the two arguments

    mov     ebx, eax            ; exit(24)
    mov     eax, 1
    int     0x80

; int add_mul(int a, int b) — returns (a + b) * 2
add_mul:
    push    ebp                 ; standard function prologue
    mov     ebp, esp
    mov     eax, [ebp + 8]      ; a
    add     eax, [ebp + 12]     ; + b
    shl     eax, 1              ; * 2
    pop     ebp                 ; epilogue
    ret
