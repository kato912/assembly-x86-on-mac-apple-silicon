; Differential test: shifts and rotates on 8/16/32-bit operands.
%include "dump32.inc"

; SH op, v8, v16, v32, count, mask, pre
%macro SH 7
    FLAGS 0x8d7
    mov eax, %2
    %7
    %1 al, %5
    DUMP %6
    FLAGS 0x8d7
    mov eax, %3
    %7
    %1 ax, %5
    DUMP %6
    FLAGS 0x8d7
    mov eax, %4
    %7
    %1 eax, %5
    DUMP %6
    FLAGS 0x8d7
    mov cl, %5
    mov ebx, %4
    %7
    %1 ebx, cl
    DUMP %6
%endmacro

%macro SHIFTS 2
    SH %1, 0x81, 0x8001, 0x80000001, 0, FL_ALL, %2
    SH %1, 0x81, 0x8001, 0x80000001, 1, FL_LOGIC, %2
    SH %1, 0x7f, 0x7fff, 0x7fffffff, 1, FL_LOGIC, %2
    SH %1, 0x40, 0x4000, 0x40000000, 1, FL_LOGIC, %2
    SH %1, 0x81, 0x8001, 0x80000001, 2, FL_SHIFTN, %2
    SH %1, 0x55, 0x5555, 0x55555555, 3, FL_SHIFTN, %2
    SH %1, 0x80, 0x8000, 0x80000000, 7, FL_SHIFTN, %2
%endmacro

%macro ROTS 2
    SH %1, 0x81, 0x8001, 0x80000001, 0, FL_ALL, %2
    SH %1, 0x81, 0x8001, 0x80000001, 1, FL_ALL, %2
    SH %1, 0x40, 0x4000, 0x40000000, 1, FL_ALL, %2
    SH %1, 0x81, 0x8001, 0x80000001, 2, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 7, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 9, FL_ROTN, %2
    SH %1, 0x12, 0x1234, 0x12345678, 31, FL_ROTN, %2
%endmacro

%macro BIG 4                    ; op, value, count, mask (32-bit only)
    FLAGS 0x8d7
    mov eax, %2
    %1 eax, %3
    DUMP %4
%endmacro

section .text
global _start
_start:
    INIT
    LABEL "shl"
    SHIFTS shl, nop
    BIG shl, 0x80000001, 16, FL_SHIFTN
    BIG shl, 0x80000001, 31, FL_SHIFTN
    BIG shl, 0x80000001, 32, FL_ALL
    BIG shl, 0x80000001, 33, FL_LOGIC
    LABEL "shr"
    SHIFTS shr, nop
    BIG shr, 0x80000001, 16, FL_SHIFTN
    BIG shr, 0x80000001, 31, FL_SHIFTN
    LABEL "sar"
    SHIFTS sar, nop
    BIG sar, 0x80000001, 16, FL_SHIFTN
    BIG sar, 0x80000001, 31, FL_SHIFTN
    LABEL "rol"
    ROTS rol, nop
    LABEL "ror"
    ROTS ror, nop
    LABEL "rcl"
    ROTS rcl, clc
    ROTS rcl, stc
    LABEL "rcr"
    ROTS rcr, clc
    ROTS rcr, stc
    LABEL "shld / shrd"
    FLAGS 0x8d7
    mov eax, 0x12345678
    mov ebx, 0x9abcdef0
    shld eax, ebx, 1
    DUMP FL_LOGIC
    shld eax, ebx, 12
    DUMP FL_SHIFTN
    mov cl, 20
    shrd eax, ebx, cl
    DUMP FL_SHIFTN
    shrd eax, ebx, 1
    DUMP FL_LOGIC
    mov ax, 0x1234
    mov bx, 0xabcd
    shld ax, bx, 4
    DUMP FL_SHIFTN
    shrd ax, bx, 8
    DUMP FL_SHIFTN
    EXIT 0
