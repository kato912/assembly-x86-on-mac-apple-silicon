; 32-bit C calls (cdecl: arguments pushed right to left, caller pops),
; compared with glibc (gcc -m32 -no-pie).
extern printf, scanf, puts, strtol, atoi, atoll, strlen, exit, malloc, free, sprintf

section .data
f1      db "%d %u %x %s %c|%5d|%-5d|", 10, 0
f2      db "%lld %llu %llx %ld", 10, 0
f3      db "%f %.2f %e %g %g %g", 10, 0
f4      db "%10.3f|%-10.1e|%+.0f|%#.0f|%G|%E", 10, 0
f5      db "%g %g %g %g %.3g %#g", 10, 0
f6      db "%f %F %f %5.1f %Lf %a", 10, 0
f7      db "read %d: %d %lld %s %lf -> %.3f", 10, 0
f8      db "strtol=%ld atoi=%d atoll=%lld len=%d", 10, 0
f9      db "printf returned %d, sprintf: [%s] %d", 10, 0
inf     db "%d %lld %15s %lf", 0
s       db "str", 0
big     db "99999999999", 0
sfmt    db "<%d>", 0
d1      dq 3.14159265358979
d2      dq 0.1
d3      dq 123456789.0
d4      dq 0.0001
d5      dq 0.00001
d6      dq 1234567.0
d7      dq 100.0
d8      dq -2.25
d9      dq 2.5
d10     dq 1.0e100
dinf    dq __?Infinity?__
dnan    dq __?QNaN?__
dneg0   dq -0.0
ld1     dt 1.5
section .bss
n1      resd 1
n2      resq 1
s1      resb 16
dv      resq 1
buf     resb 32
section .text
global main
main:
    push ebp
    mov ebp, esp

    push -3
    push 5
    push 'Z'
    push s
    push 255
    push -1
    push -7
    push f1
    call printf
    add esp, 32

    push -2                     ; long = 32 bits
    push 0xffffffff             ; llx: high half
    push 0x12345678             ;      low half
    push 0x80000000
    push 0
    push 0x7fffffff             ; lld = 0x7fffffff_ffffffff... high
    push 0xffffffff             ; low
    push f2
    call printf
    add esp, 32

    push dword [d6+4]
    push dword [d6]
    push dword [d5+4]
    push dword [d5]
    push dword [d4+4]
    push dword [d4]
    push dword [d3+4]
    push dword [d3]
    push dword [d2+4]
    push dword [d2]
    push dword [d1+4]
    push dword [d1]
    push f3
    call printf
    add esp, 52

    push dword [d10+4]
    push dword [d10]
    push dword [d2+4]
    push dword [d2]
    push dword [d9+4]
    push dword [d9]
    push dword [d9+4]
    push dword [d9]
    push dword [d3+4]
    push dword [d3]
    push dword [d1+4]
    push dword [d1]
    push f4
    call printf
    add esp, 52

    push dword [d7+4]
    push dword [d7]
    push dword [d1+4]
    push dword [d1]
    push dword [d7+4]
    push dword [d7]
    push dword [d6+4]
    push dword [d6]
    push dword [d4+4]
    push dword [d4]
    push dword [dneg0+4]
    push dword [dneg0]
    push f5
    call printf
    add esp, 52

    push dword [d2+4]
    push dword [d2]
    push dword [ld1+8]          ; long double: 12 bytes on the stack
    push dword [ld1+4]
    push dword [ld1]
    push dword [d8+4]
    push dword [d8]
    push dword [dnan+4]
    push dword [dnan]
    push dword [dinf+4]
    push dword [dinf]
    push dword [dneg0+4]
    push dword [dneg0]
    push f6
    call printf
    add esp, 56

    ; scanf with int, long long, string and double
    push dv
    push s1
    push n2
    push n1
    push inf
    call scanf
    add esp, 20
    push dword [dv+4]
    push dword [dv]
    push dword [dv+4]
    push dword [dv]
    push s1
    push dword [n2+4]
    push dword [n2]
    push dword [n1]
    push eax
    push f7
    call printf
    add esp, 40

    ; long is 32-bit here: strtol saturates at 2147483647
    push big
    call strlen
    add esp, 4
    push eax
    push big
    call atoll
    add esp, 4
    push edx
    push eax
    push big
    call atoi
    add esp, 4
    push eax
    push 10
    push 0
    push big
    call strtol
    add esp, 12
    push eax
    push f8
    call printf
    add esp, 24
    mov esi, eax                ; printf's return value (esi survives calls)

    push 42
    push sfmt
    push buf
    call sprintf
    add esp, 12
    push eax
    push buf
    push esi
    push f9
    call printf
    add esp, 16

    push 16
    call malloc
    add esp, 4
    test eax, 15                ; 16-byte aligned on i386 glibc too
    setnz bl
    push eax
    call free
    add esp, 4
    movzx eax, bl
    push eax
    call exit
