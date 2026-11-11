; Differential test: bt/bts/btr/btc, bsf/bsr, bswap, setcc and cmovcc.
default rel
%include "dump64.inc"

%macro BT 3                     ; op, value, bit
    FLAGS 0x202
    mov rax, %2
    mov rcx, %3
    %1 eax, ecx
    DUMP FL_BT
    FLAGS 0x8d7
    mov rax, %2
    %1 rax, %3
    DUMP FL_BT
    FLAGS 0x202
    mov rax, %2
    mov rcx, %3
    %1 ax, cx
    DUMP FL_BT
%endmacro

%macro BTALL 1
    BT %1, 0x5a5a5a5a5a5a5a5a, 0
    BT %1, 0x5a5a5a5a5a5a5a5a, 1
    BT %1, 0x5a5a5a5a5a5a5a5a, 15
    BT %1, 0x5a5a5a5a5a5a5a5a, 31
    BT %1, 0x5a5a5a5a5a5a5a5a, 33
    BT %1, 0x8000000000000000, 63
    BT %1, -1, 7
    BT %1, 0, 40
%endmacro

%macro SCAN 1                   ; value
    FLAGS 0x202
    mov rbx, %1
    bsf rax, rbx
    DUMP FL_BSF
    bsr rcx, rbx
    DUMP FL_BSF
    bsf edx, ebx
    DUMP FL_BSF
    bsr esi, ebx
    DUMP FL_BSF
    bsr di, bx
    DUMP FL_BSF
%endmacro

; CC a, b: compare and record all 16 conditions with setcc into memory,
; and cmovcc into registers
%macro CC 2
    mov rax, %1
    mov rbx, %2
    cmp rax, rbx
    lea rdi, [conds]
    seto  [rdi + 0]
    setno [rdi + 1]
    setb  [rdi + 2]
    setae [rdi + 3]
    sete  [rdi + 4]
    setne [rdi + 5]
    setbe [rdi + 6]
    seta  [rdi + 7]
    sets  [rdi + 8]
    setns [rdi + 9]
    setp  [rdi + 10]
    setnp [rdi + 11]
    setl  [rdi + 12]
    setge [rdi + 13]
    setle [rdi + 14]
    setg  [rdi + 15]
    xor edi, edi                ; (xor changes flags: compare again)
    cmp rax, rbx
    PRINTMEM conds, 16
    mov ecx, 1
    mov edx, 2
    mov r8d, 3
    mov r9d, 4
    mov r10, -1
    mov r11, -1
    mov r12, -1
    mov r13, -1
    cmovl ecx, edx
    cmovg r8d, r9d
    cmova r10, rdx
    cmovbe r11, rdx
    cmovs r12d, edx             ; 32-bit cmov zero-extends even when not taken
    cmovns r13w, dx
    DUMP
    xor r10d, r10d
    xor r11d, r11d
    xor r12d, r12d
    xor r13d, r13d
%endmacro

section .data
conds: times 16 db 0
         db 10
bitmem:  dq 0, 0, 0

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
    LABEL "bt memory with register offsets"
    lea rbx, [bitmem + 8]
    mov ecx, 70
    bts [rbx], ecx              ; bit 70 of bitmem+8 -> bitmem+16, bit 6
    mov rcx, -3
    bts [rbx], rcx              ; bit -3 -> bitmem+7, bit 5
    mov ecx, 5
    btc dword [rbx], 5
    bt [rbx], ecx
    xor ebx, ebx
    mov rax, [bitmem]
    mov rdx, [bitmem + 8]
    mov rsi, [bitmem + 16]
    xor ecx, ecx
    DUMP FL_BT

    LABEL "bsf/bsr"
    SCAN 1
    SCAN 0x80
    SCAN 0x00f0
    SCAN 0x8000
    SCAN 0x80000001
    SCAN 0x0000010000000000
    SCAN -1
    FLAGS 0x202
    xor ebx, ebx
    bsf rax, rbx                ; source 0: ZF=1
    mov eax, 0                  ; (destination is undefined; overwrite it)
    DUMP FL_BSF

    LABEL "bswap"
    FLAGS 0x202
    mov eax, 0x11223344
    bswap eax
    mov rbx, 0x0102030405060708
    bswap rbx
    mov r14, -2
    bswap r14
    DUMP

    LABEL "setcc/cmovcc"
    CC 0, 0
    CC 1, 2
    CC 2, 1
    CC -1, 1
    CC 1, -1
    CC 0x8000000000000000, 1
    CC 0x7fffffffffffffff, -1
    CC 3, 3
    CC 0xff, 0
    EXIT 0
