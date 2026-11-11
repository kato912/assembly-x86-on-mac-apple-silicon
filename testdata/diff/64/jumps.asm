; Differential test: every conditional jump taken and not taken, loop,
; loope, loopne, jrcxz/jecxz, indirect jumps and a jump table.
default rel
%include "dump64.inc"

; J cc: count in rbx how many of the comparisons below take jcc
%macro J 1
    xor ebx, ebx
    %assign k 0
    %rep 8
        mov rax, [pairs + k*16]
        cmp rax, [pairs + k*16 + 8]
        j%1 %%taken %+ k
        jmp %%next %+ k
    %%taken %+ k:
        bts rbx, k
    %%next %+ k:
        %assign k k+1
    %endrep
    xor eax, eax
    DUMP FL_LOGIC
%endmacro

section .data
pairs:  dq 0, 0
        dq 1, 2
        dq 2, 1
        dq -1, 1
        dq 1, -1
        dq 0x8000000000000000, 1
        dq 0x7fffffffffffffff, -1
        dq 5, 5
table:  dq case0, case1, case2

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
    J c
    J nc
    J z
    J nz
    J nae
    J nbe
    J nge
    J nle

    LABEL "loop"
    xor eax, eax
    mov ecx, 10
.l1:
    add rax, rcx
    loop .l1
    DUMP
    LABEL "loope / loopne"
    lea rsi, [pairs]
    mov ecx, 16
    xor edx, edx
.l2:
    inc edx
    cmp qword [rsi + rdx*8], 0
    loopne .l2                  ; until a zero quadword
    mov r8, rcx
    mov r9, rdx
    mov ecx, 16
    xor edx, edx
.l3:
    inc edx
    cmp qword [rsi + rdx*8 + 8], 1
    loope .l3
    xor esi, esi
    DUMP FL_ALL

    LABEL "jrcxz / jecxz"
    xor ecx, ecx
    mov eax, 1
    jrcxz .z1
    mov eax, 2
.z1:
    mov rcx, 0x100000000        ; ecx = 0 but rcx != 0
    mov ebx, 1
    jrcxz .z2
    mov ebx, 2
.z2:
    mov edx, 1
    jecxz .z3
    mov edx, 2
.z3:
    DUMP

    LABEL "indirect jumps / jump table"
    xor r12d, r12d
    xor ecx, ecx
again_:
    lea rax, [table]
    jmp [rax + rcx*8]
case0:
    add r12, 1
    jmp next_case
case1:
    add r12, 10
    jmp next_case
case2:
    add r12, 100
next_case:
    inc ecx
    cmp ecx, 3
    jb again_
    lea rdx, [via_reg]
    jmp rdx
    mov r12, -1                 ; skipped
via_reg:
    xor eax, eax
    xor edx, edx
    DUMP
    EXIT 0
