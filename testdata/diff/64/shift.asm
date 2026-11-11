; Differential test: shl/sal, shr, sar, rol, ror, rcl, rcr, shld, shrd with
; immediate and cl counts on every operand size. Flags the CPU leaves
; undefined for a given count are masked out.
default rel
%include "dump64.inc"

; SH op, v8, v16, v32, v64, count, mask, pre
%macro SH 8
    FLAGS 0x8d7
    mov rax, %2
    %8
    %1 al, %6
    DUMP %7
    FLAGS 0x8d7
    mov rax, %3
    %8
    %1 ax, %6
    DUMP %7
    FLAGS 0x8d7
    mov rax, %4
    %8
    %1 eax, %6
    DUMP %7
    FLAGS 0x8d7
    mov rax, %5
    %8
    %1 rax, %6
    DUMP %7
    FLAGS 0x8d7
    mov cl, %6                  ; same through cl
    mov rbx, %5
    %8
    %1 rbx, cl
    DUMP %7
%endmacro

; SHIFTS op, pre: shl/shr/sar-style counts (CF undefined when count >= size,
; so 8-bit counts stay below 8)
%macro SHIFTS 2
    SH %1, 0x81, 0x8001, 0x80000001, 0x8000000000000001, 0, FL_ALL, %2
    SH %1, 0x81, 0x8001, 0x80000001, 0x8000000000000001, 1, FL_LOGIC, %2
    SH %1, 0x7f, 0x7fff, 0x7fffffff, 0x7fffffffffffffff, 1, FL_LOGIC, %2
    SH %1, 0x40, 0x4000, 0x40000000, 0x4000000000000000, 1, FL_LOGIC, %2
    SH %1, 0xc3, 0xc003, 0xc0000003, 0xc000000000000003, 1, FL_LOGIC, %2
    SH %1, 0x81, 0x8001, 0x80000001, 0x8000000000000001, 2, FL_SHIFTN, %2
    SH %1, 0x55, 0x5555, 0x55555555, 0x5555555555555555, 3, FL_SHIFTN, %2
    SH %1, 0xff, 0xffff, 0xffffffff, 0xffffffffffffffff, 4, FL_SHIFTN, %2
    SH %1, 0x80, 0x8000, 0x80000000, 0x8000000000000000, 7, FL_SHIFTN, %2
    SH %1, 0x01, 0x0001, 0x00000001, 0x0000000000000001, 7, FL_SHIFTN, %2
%endmacro

; BIG op, v32, v64, count, mask: counts only valid for 32/64-bit operands
%macro BIG 5
    FLAGS 0x8d7
    mov eax, %2
    %1 eax, %4
    DUMP %5
    FLAGS 0x8d7
    mov rax, %3
    %1 rax, %4
    DUMP %5
    FLAGS 0x8d7
    mov rdx, %3
    mov cl, %4
    %1 rdx, cl
    DUMP %5
%endmacro

%macro BIGS 1
    BIG %1, 0x80000001, 0x8000000000000001, 8, FL_SHIFTN
    BIG %1, 0x12345678, 0x123456789abcdef0, 15, FL_SHIFTN
    BIG %1, 0x12345678, 0x123456789abcdef0, 16, FL_SHIFTN
    BIG %1, 0x80000000, 0x8000000000000000, 31, FL_SHIFTN
    BIG %1, 0xdeadbeef, 0xdeadbeefcafebabe, 32, FL_SHIFTN
    BIG %1, 0xdeadbeef, 0xdeadbeefcafebabe, 33, FL_SHIFTN
    BIG %1, 0x80000000, 0x8000000000000000, 63, FL_SHIFTN
%endmacro

; ROTS op, pre: rotate counts (8-bit counts may exceed 8 for rotates)
%macro ROTS 2
    SH %1, 0x81, 0x8001, 0x80000001, 0x8000000000000001, 0, FL_ALL, %2
    SH %1, 0x81, 0x8001, 0x80000001, 0x8000000000000001, 1, FL_ALL, %2
    SH %1, 0x40, 0x4000, 0x40000000, 0x4000000000000000, 1, FL_ALL, %2
    SH %1, 0x01, 0x0001, 0x00000001, 0x0000000000000001, 1, FL_ALL, %2
    SH %1, 0x81, 0x8001, 0x80000001, 0x8000000000000001, 2, FL_ROTN, %2
    SH %1, 0x5a, 0x5a5a, 0x5a5a5a5a, 0x5a5a5a5a5a5a5a5a, 3, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 0x123456789abcdef0, 7, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 0x123456789abcdef0, 8, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 0x123456789abcdef0, 9, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 0x123456789abcdef0, 17, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 0x123456789abcdef0, 31, FL_ROTN, %2
%endmacro

; DS op, dst, src, count, mask: shld/shrd on 16/32/64-bit
%macro DS 5
    FLAGS 0x8d7
    mov rax, %2
    mov rbx, %3
    %1 eax, ebx, %4
    DUMP %5
    FLAGS 0x8d7
    mov rax, %2
    mov rbx, %3
    %1 rax, rbx, %4
    DUMP %5
    FLAGS 0x8d7
    mov rax, %2
    mov rbx, %3
    mov cl, %4
    %1 rax, rbx, cl
    DUMP %5
%endmacro

%macro DS16 5
    FLAGS 0x8d7
    mov rax, %2
    mov rbx, %3
    %1 ax, bx, %4
    DUMP %5
%endmacro

section .data
mw: dw 0x8001
md: dd 0x80000001

section .text
global _start
_start:
    INIT
    LABEL "shl"
    SHIFTS shl, nop
    BIGS shl
    LABEL "shr"
    SHIFTS shr, nop
    BIGS shr
    LABEL "sar"
    SHIFTS sar, nop
    BIGS sar
    LABEL "rol"
    ROTS rol, nop
    LABEL "ror"
    ROTS ror, nop
    LABEL "rcl cf=0"
    ROTS rcl, clc
    LABEL "rcl cf=1"
    ROTS rcl, stc
    LABEL "rcr cf=0"
    ROTS rcr, clc
    LABEL "rcr cf=1"
    ROTS rcr, stc

    LABEL "shld"
    DS shld, 0x12345678, 0x9abcdef0, 1, FL_LOGIC
    DS shld, 0x12345678, 0x9abcdef0, 4, FL_SHIFTN
    DS shld, 0x87654321, 0x0fedcba9, 16, FL_SHIFTN
    DS shld, 0x87654321, 0x0fedcba9, 31, FL_SHIFTN
    DS shld, 0x87654321, 0x0fedcba9, 0, FL_ALL
    DS16 shld, 0x1234, 0xabcd, 1, FL_LOGIC
    DS16 shld, 0x1234, 0xabcd, 8, FL_SHIFTN
    DS16 shld, 0x8234, 0xabcd, 15, FL_SHIFTN
    LABEL "shrd"
    DS shrd, 0x12345678, 0x9abcdef0, 1, FL_LOGIC
    DS shrd, 0x12345678, 0x9abcdef0, 4, FL_SHIFTN
    DS shrd, 0x87654321, 0x0fedcba9, 16, FL_SHIFTN
    DS shrd, 0x87654321, 0x0fedcba9, 31, FL_SHIFTN
    DS16 shrd, 0x1234, 0xabcd, 1, FL_LOGIC
    DS16 shrd, 0x1234, 0xabcd, 8, FL_SHIFTN
    mov rax, 0x0123456789abcdef
    mov rbx, 0xfedcba9876543210
    shld rax, rbx, 40
    DUMP FL_SHIFTN
    shrd rax, rbx, 63
    DUMP FL_SHIFTN

    LABEL "memory"
    shl word [mw], 1
    movzx eax, word [mw]
    DUMP FL_LOGIC
    sar dword [md], 4
    mov ebx, [md]
    DUMP FL_SHIFTN
    mov cl, 3
    rol dword [md], cl
    mov ebx, [md]
    DUMP FL_ROTN
    shrd [md], ebx, 5
    mov edx, [md]
    DUMP FL_SHIFTN
    EXIT 0
