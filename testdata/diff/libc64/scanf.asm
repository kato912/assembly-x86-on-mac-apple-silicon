; scanf conversions, return values and leftover input, compared with glibc.
default rel
extern scanf, printf, sscanf, getchar, fgets, stdin

; S "format", ptr1..ptr4: scanf, then print its return value.
%macro S 1-5
section .data
%%f: db %1, 0
section .text
    lea rdi, [%%f]
%if %0 > 1
    lea rsi, [%2]
%endif
%if %0 > 2
    lea rdx, [%3]
%endif
%if %0 > 3
    lea rcx, [%4]
%endif
%if %0 > 4
    lea r8, [%5]
%endif
    xor eax, eax
    call scanf
    lea rdi, [rfmt]
    mov esi, eax
    xor eax, eax
    call printf
%endmacro

; P "format", a1..a4: printf with integer values (rsi..r8)
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

section .data
rfmt    db "ret=%d ", 0
section .bss
i1      resq 1
i2      resq 1
i3      resq 1
q1      resq 1
s1      resb 32
s2      resb 32
c1      resb 4
d1      resq 1
f1      resd 1
line    resb 64
section .text
global main
main:
    push rbp
    mov rbp, rsp

    S "%d%d", i1, i2
    movsxd rsi, dword [i1]
    movsxd rdx, dword [i2]
    P "%d %d", rsi, rdx

    S "%i %i %i", i1, i2, i3
    P "%d %d %d", [i1], [i2], [i3]

    S "%x %o %u", i1, i2, i3
    P "%x %o %u", [i1], [i2], [i3]

    S "%3d%d", i1, i2
    P "%d %d", [i1], [i2]

    S "%s %5s", s1, s2
    P "[%s] [%s]", s1, s2
    S "%s", s1
    P "[%s]", s1

    S " %c%c", c1, c1+1
    movzx esi, byte [c1]
    movzx edx, byte [c1+1]
    P "%c%c", rsi, rdx

    S ` %[a-z]%*[ ]%[^\n]`, s1, s2
    P "[%s] [%s]", s1, s2

    S "%d,%d", i1, i2           ; input "5;6": only one item
    P "%d", [i1]
    S `%*[^\n]`                 ; skip the rest of the line
    call getchar                ; and the newline
    P "getchar=%d", rax

    S "%d", i1                  ; "abc": no number, returns 0, abc stays
    lea rdi, [line]
    mov esi, 64
    mov rdx, [stdin]
    call fgets
    P "left: %s", line

    mov qword [q1], 0
    mov qword [i2], 0
    mov qword [i3], 0
    S "%ld %hhd %hd", q1, i2, i3
    movsx rdx, byte [i2]
    movsx rcx, word [i3]
    P "%ld %d %d", [q1], rdx, rcx

    S "%d %n%d", i1, i2, i3     ; %n is not counted
    P "%d n=%d %d", [i1], [i2], [i3]

    S "%lf %f", d1, f1
    mov eax, [f1]
    P "%lx %x", [d1], rax

    S "%d", i1                  ; 99999999999 does not fit an int
    P "%d", [i1]

    S "%d", i1                  ; end of input: EOF
    P ""

    ; sscanf on a string
    lea rdi, [sstr]
    lea rsi, [sfmt]
    lea rdx, [s1]
    lea rcx, [i1]
    xor eax, eax
    call sscanf
    mov [i2], eax
    P "sscanf=%d [%s] %d", [i2], s1, [i1]

    xor eax, eax
    leave
    ret
section .data
sstr    db "  item=42 rest", 0
sfmt    db " %4[a-z]=%d", 0
