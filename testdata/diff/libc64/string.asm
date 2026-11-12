; <string.h>, <stdlib.h> and <ctype.h> functions, compared with glibc.
default rel
extern printf, strlen, strcmp, strncmp, strcpy, strncpy, strcat, strchr, strrchr, strstr
extern memcmp, memset, memcpy, atoi, atol, strtol, strtoul, $abs, labs
extern toupper, tolower, isdigit, isalpha, isspace, isupper, islower, isalnum, isxdigit, ispunct
extern rand, srand

%macro P 1-5
section .data
%%f: db %1, 10, 0
section .text
    lea rdi, [%%f]
%if %0 > 1
    mov rsi, %2
%endif
%if %0 > 2
    mov rdx, %3
%endif
%if %0 > 3
    mov rcx, %4
%endif
%if %0 > 4
    mov r8, %5
%endif
    xor eax, eax
    call printf
%endmacro

; R func, a1, a2, a3: call func with up to 3 arguments, result in [r]
; SIGN: [r] = -1, 0 or 1 by the sign of eax (glibc's strcmp/memcmp
; return any negative/positive number, depending on the CPU).
%macro SIGN 0
    xor ecx, ecx
    test eax, eax
    setg cl
    shr eax, 31
    sub ecx, eax
    movsxd rcx, ecx
    mov [r], rcx
%endmacro

%macro R 1-4
%if %0 > 1
    mov rdi, %2
%endif
%if %0 > 2
    mov rsi, %3
%endif
%if %0 > 3
    mov rdx, %4
%endif
    call %1
    mov [r], rax
%endmacro

section .data
hello   db "hello", 0
help    db "help", 0
world   db "world", 0
banana  db "banana", 0
nana    db "nana", 0
num1    db "  -123abc", 0
num2    db "+77", 0
num3    db "99999999999999999999", 0
hex     db "0x1A zz", 0
neg1    db "-1", 0
chars   db "a5 Z_", 0
section .bss
r       resq 1
r2      resq 1
end     resq 1
buf     resb 64
section .text
global main
main:
    push rbp
    mov rbp, rsp

    lea rbx, [hello]
    R strlen, rbx
    P "strlen=%ld", [r]
    lea rax, [help]
    R strcmp, rbx, rax
    SIGN
    mov rax, [r]
    mov [r2], rax
    lea rax, [help]
    R strcmp, rax, rbx
    SIGN
    P "strcmp=%d %d", [r2], [r]
    lea rax, [help]
    R strncmp, rbx, rax, 3
    SIGN
    P "strncmp=%d", [r]
    R strcmp, rbx, rbx
    SIGN
    P "strcmp same=%d", [r]

    lea rax, [buf]
    R strcpy, rax, rbx
    lea rax, [buf]
    lea rcx, [world]
    R strcat, rax, rcx
    lea rax, [buf]
    P "%s", rax
    lea rax, [buf]
    lea rcx, [banana]
    R strncpy, rax, rcx, 3
    lea rax, [buf]
    P "%s", rax

    lea rbx, [banana]
    R strchr, rbx, 'n'
    sub [r], rbx
    mov rax, [r]
    mov [r2], rax
    R strrchr, rbx, 'n'
    sub [r], rbx
    P "strchr=%ld strrchr=%ld", [r2], [r]
    lea rbx, [banana]
    lea rax, [nana]
    R strstr, rbx, rax
    sub [r], rbx
    P "strstr=%ld", [r]
    lea rbx, [banana]
    R strchr, rbx, 'z'
    P "strchr none=%ld", [r]

    lea rbx, [hello]
    lea rax, [help]
    R memcmp, rbx, rax, 4
    SIGN
    P "memcmp=%d", [r]
    lea rax, [buf]
    R memset, rax, '*', 5
    mov byte [buf+5], 0
    lea rax, [buf]
    P "%s", rax
    lea rax, [buf]
    lea rcx, [world]
    R memcpy, rax, rcx, 3
    lea rax, [buf]
    P "%s", rax

    lea rax, [num1]
    R atoi, rax
    mov [r2], rax
    lea rax, [num2]
    R atoi, rax
    P "atoi=%d %d", [r2], [r]
    lea rax, [num3]
    R atol, rax
    P "atol=%ld", [r]
    lea rax, [hex]
    lea rcx, [end]
    R strtol, rax, rcx, 16
    lea rax, [hex]
    sub [end], rax
    P "strtol16=%ld end=%ld", [r], [end]
    lea rax, [hex]
    lea rcx, [end]
    R strtol, rax, rcx, 0
    lea rax, [hex]
    sub [end], rax
    P "strtol0=%ld end=%ld", [r], [end]
    lea rax, [neg1]
    R strtoul, rax, 0, 10
    P "strtoul=%lu", [r]
    R $abs, -5
    mov [r2], rax
    R labs, -7
    P "abs=%d labs=%ld", [r2], [r]

    ; ctype: glibc returns its table bits, not just 1
    lea rbx, [chars]
.ct:
    movzx r12d, byte [rbx]
    test r12d, r12d
    jz .ctdone
    R isdigit, r12
    mov r13, rax
    R isalpha, r12
    mov r14, rax
    R isspace, r12
    mov r15, rax
    P "%c: digit=%d alpha=%d space=%d", r12, r13, r14, r15
    R isupper, r12
    mov r13, rax
    R islower, r12
    mov r14, rax
    R isalnum, r12
    mov r15, rax
    P "   upper=%d lower=%d alnum=%d", r13, r14, r15
    R isxdigit, r12
    mov r13, rax
    R ispunct, r12
    mov r14, rax
    R toupper, r12
    mov r15, rax
    P "   xdigit=%d punct=%d toupper=%c", r13, r14, r15
    inc rbx
    jmp .ct
.ctdone:
    R tolower, 'Q'
    mov r13, rax
    R toupper, -1
    P "tolower=%c toupper(EOF)=%d", r13, [r]

    ; rand: the unseeded sequence, then srand(42) and srand(0)
    R rand
    mov r13, rax
    R rand
    mov r14, rax
    P "rand: %d %d", r13, r14
    R srand, 42
    R rand
    mov r13, rax
    R rand
    mov r14, rax
    R rand
    P "srand(42): %d %d %d", r13, r14, [r]
    R srand, 0
    R rand
    P "srand(0): %d", [r]
    mov edi, 0x87654321
    call srand
    R rand
    P "srand(big): %d", [r]

    xor eax, eax
    leave
    ret
