; Differential test: bt/bts/btr/btc, bsf/bsr, bswap, setcc, cmovcc.
%include "dump32.inc"

%macro BT 3
    FLAGS 0x202
    mov eax, %2
    mov ecx, %3
    %1 eax, ecx
    DUMP FL_BT
    FLAGS 0x8d7
    mov eax, %2
    %1 eax, %3
    DUMP FL_BT
    FLAGS 0x202
    mov eax, %2
    %1 ax, %3
    DUMP FL_BT
%endmacro

%macro BTALL 1
    BT %1, 0x5a5a5a5a, 0
    BT %1, 0x5a5a5a5a, 1
    BT %1, 0x5a5a5a5a, 15
    BT %1, 0x5a5a5a5a, 31
    BT %1, 0x5a5a5a5a, 33
    BT %1, -1, 7
%endmacro

%macro SCAN 1
    FLAGS 0x202
    mov ebx, %1
    bsf eax, ebx
    DUMP FL_BSF
    bsr ecx, ebx
    DUMP FL_BSF
    bsf dx, bx
    DUMP FL_BSF
%endmacro

%macro CC 2
    mov eax, %1
    mov ebx, %2
    cmp eax, ebx
    seto  [conds + 0]
    setno [conds + 1]
    setb  [conds + 2]
    setae [conds + 3]
    sete  [conds + 4]
    setne [conds + 5]
    setbe [conds + 6]
    seta  [conds + 7]
    sets  [conds + 8]
    setns [conds + 9]
    setp  [conds + 10]
    setnp [conds + 11]
    setl  [conds + 12]
    setge [conds + 13]
    setle [conds + 14]
    setg  [conds + 15]
    PRINTMEM conds, 17
    mov ecx, 1
    mov edx, 2
    mov esi, 3
    mov edi, 4
    cmovl ecx, edx
    cmovg esi, edi
    cmovbe di, dx
    DUMP
%endmacro

section .data
conds:  times 16 db 0
        db 10
bitmem: dd 0, 0, 0

section .text
global _start
_start:
    INIT
    LABEL "bt"
    BTALL bt
    LABEL "bts"
    BTALL bts
    LABEL "btr"
    BTALL btr
    LABEL "btc"
    BTALL btc
    LABEL "bt memory"
    mov ecx, 40
    bts [bitmem + 4], ecx
    mov ecx, -2
    bts [bitmem + 4], ecx
    btc dword [bitmem], 3
    mov eax, [bitmem]
    mov ebx, [bitmem + 4]
    mov edx, [bitmem + 8]
    DUMP FL_BT
    LABEL "bsf/bsr"
    SCAN 1
    SCAN 0x80
    SCAN 0x8000
    SCAN 0x80000001
    SCAN 0x00f0f000
    LABEL "bswap"
    FLAGS 0x202
    mov eax, 0x11223344
    bswap eax
    mov ebp, 0x80000001
    bswap ebp
    DUMP
    LABEL "setcc/cmovcc"
    CC 0, 0
    CC 1, 2
    CC 2, 1
    CC -1, 1
    CC 1, -1
    CC 0x80000000, 1
    CC 0x7fffffff, -1
    EXIT 0
