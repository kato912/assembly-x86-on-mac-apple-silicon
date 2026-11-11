; Differential test: conditional jumps, loop/loope/loopne, jecxz/jcxz,
; indirect jumps through registers, memory and a jump table.
%include "dump32.inc"

%macro J 1
    xor ebx, ebx
    %assign k 0
    %rep 7
        mov eax, [pairs + k*8]
        cmp eax, [pairs + k*8 + 4]
        j%1 %%taken %+ k
        jmp %%next %+ k
    %%taken %+ k:
        bts ebx, k
    %%next %+ k:
        %assign k k+1
    %endrep
    xor eax, eax
    DUMP FL_LOGIC
%endmacro

section .data
pairs:  dd 0, 0
        dd 1, 2
        dd 2, 1
        dd -1, 1
        dd 1, -1
        dd 0x80000000, 1
        dd 0x7fffffff, -1
table:  dd case0, case1, case2

section .text
global _start
_start:
    INIT
    LABEL "jcc"
    J o
    J no
    J b
    J ae
    J e
    J ne
    J be
    J a
    J s
    J ns
    J p
    J np
    J l
    J ge
    J le
    J g
    LABEL "loops"
    xor eax, eax
    mov ecx, 100
l1: add eax, ecx
    loop l1
    DUMP
    mov esi, pairs
    mov ecx, 14
    xor edx, edx
l2: inc edx
    cmp dword [esi + edx*4], 1
    loopne l2
    xor esi, esi
    DUMP
    mov ecx, 5
    xor edx, edx
l3: inc edx
    cmp edx, 3
    loope l3
    DUMP
    LABEL "jecxz / jcxz"
    xor ecx, ecx
    mov eax, 1
    jecxz z1
    mov eax, 2
z1: mov ecx, 0x10000            ; cx = 0, ecx != 0
    mov ebx, 1
    jcxz z2
    mov ebx, 2
z2: mov edx, 1
    jecxz z3
    mov edx, 2
z3: DUMP
    LABEL "indirect"
    xor edi, edi
    xor ecx, ecx
again:
    jmp [table + ecx*4]
case0:
    add edi, 1
    jmp next
case1:
    add edi, 10
    jmp next
case2:
    add edi, 100
next:
    inc ecx
    cmp ecx, 3
    jb again
    mov eax, via
    jmp eax
    mov edi, -1
via:
    xor eax, eax
    DUMP
    EXIT 0
