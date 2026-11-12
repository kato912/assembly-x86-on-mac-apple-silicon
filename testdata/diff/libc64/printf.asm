; printf and friends, compared with glibc (gcc -no-pie).
default rel
extern printf, puts, putchar, fputs, fflush, sprintf, snprintf, stdout

; P "format", a1..a5: printf with up to 5 integer arguments (rsi..r9).
%macro P 1-6
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
%if %0 > 5
    mov r9, %6
%endif
    xor eax, eax
    call printf
    mov [rv], eax
%endmacro

section .data
abc     db "abcdef", 0
empty   db 0
hello   db "hello", 0
retfmt  db "returned %d", 10, 0
bfmt    db "[%s] %d", 10, 0
A       db "A", 10, 0
B       db "B", 10
C       db "C", 10, 0
section .bss
rv      resd 1
cnt     resd 1
buf     resb 64
section .text
global main
main:
    push rbp
    mov rbp, rsp

    P "%d|%i|%5d|%-5d|%05d", 42, -42, 42, 42, -42
    P "%+d|% d|%+5d|%-+5d|%+05d", 42, 42, 42, -42, 42
    P "%x|%X|%#x|%#X|%#o", 3054, 3054, 3054, 3054, 8
    P "%u|%lu|%ld|%li|%lx", -1, -1, -1, -2, -1
    P "%hhd|%hhu|%hd|%hu|%hx", 0x1ff, 0x1ff, 0x18000, 0x18000, 0x18000
    P "[%.0d|%.0x|%#.0o|%.5d|%8.3d]", 0, 0, 0, 42, -42
    P "%lld|%llu|%zd|%zu|%jd", 0x7fffffffffffffff, 0x8000000000000000, -5, 5, -9000000000
    P "%#x|%-#8x|%#08x|%08X|%-8o|", 0, 42, 42, 255, 8
    P "%c|%5c|%-3c|%c%c", 'a', 'b', 'c', 'o', 'k'
    P "%s|%10s|%-10s|%.3s|%10.3s|", abc, abc, abc, abc, abc
    P "[%s][%5s][%-5s]", empty, empty, hello
    P "%s|%.3s|%.6s|%10s", 0, 0, 0, 0
    P "%p|%p|%10p|%-10p|", 0, 0x1234, 0, 0x1000
    P "%%|100%%|%5.2s%%", abc
    P "no args"
    P "%d%%", 50
    mov eax, [rv]
    lea rdi, [retfmt]
    mov esi, eax
    xor eax, eax
    call printf                 ; "50%\n" -> returned 4

    ; width and precision from arguments
    push 0
    push 7                      ; 6th value goes on the stack
    P "[%*d][%-*d][%.*d]", 5, 1, 5, 2, 3
    add rsp, 16

    ; more than 5 values: the rest on the stack
    push 0
    push 9
    push 8
    push 7
    P "%d %d %d %d %d %d %d %d", 1, 2, 3, 4, 5
    add rsp, 32
    ; %n stores the count so far
    sub rsp, 16
    P "abc%n", cnt
    add rsp, 16
    P "count=%d", [cnt]

    ; sprintf / snprintf
    lea rdi, [buf]
    lea rsi, [bfmt]
    lea rdx, [hello]
    mov ecx, 7
    xor eax, eax
    call sprintf
    mov [rv], eax
    lea rdi, [buf]
    xor eax, eax
    call printf
    P "sprintf returned %d", [rv]
    lea rdi, [buf]
    mov esi, 4
    lea rdx, [bfmt]
    lea rcx, [hello]
    mov r8d, 7
    xor eax, eax
    call snprintf
    mov [rv], eax
    P "snprintf: %s %d", buf, [rv]

    ; puts / putchar / fputs return values
    lea rdi, [hello]
    call puts
    mov [rv], eax
    mov edi, '!'
    call putchar
    mov edi, 10
    call putchar
    lea rdi, [hello]
    mov rsi, [stdout]
    call fputs
    P " puts=%d", [rv]

    ; buffered printf vs the write system call: on a pipe, printf output
    ; waits in the buffer until fflush (or exit)
    lea rdi, [A]
    xor eax, eax
    call printf
    mov eax, 1
    mov edi, 1
    lea rsi, [B]
    mov edx, 2
    syscall
    mov rdi, [stdout]
    call fflush
    lea rdi, [C]
    xor eax, eax
    call printf

    mov eax, 3
    leave
    ret
; exit: 3
