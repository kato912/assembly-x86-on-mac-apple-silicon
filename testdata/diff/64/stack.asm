; Differential test: push/pop forms, pushfq/popfq, call/ret (direct,
; register, memory, ret imm), enter/leave and recursion. Stack addresses
; appear only as offsets from the starting rsp (sp= in each dump).
default rel
%include "dump64.inc"

section .data
fptr:   dq add3
val:    dq 0x1122334455667788
w16:    dw 0xbeef

section .text
global _start
_start:
    INIT
    LABEL "push/pop"
    mov rax, 0x1234
    push rax
    push -1                     ; imm8, sign-extended to 64 bits
    push 0x12345678             ; imm32
    push qword [val]
    DUMP
    pop rbx
    pop rcx
    pop rdx
    pop rsi
    DUMP
    push word 0x4142            ; 16-bit push
    push word [w16]
    DUMP
    pop di
    pop si
    DUMP
    push qword [val]
    pop qword [val + 0]         ; pop to memory
    mov r8, [val]
    mov r9, rsp                 ; push rsp pushes the old value
    push rsp
    pop r10
    sub r10, r9
    xor r9d, r9d
    DUMP

    LABEL "pushfq/popfq"
    FLAGS 0x8d7
    pushfq
    FLAGS 0x202
    popfq
    DUMP
    pushfq
    pop rax
    and eax, 0xcd5
    DUMP
    push qword 0x401            ; DF + CF
    popfq
    DUMP
    cld
    clc

    LABEL "call/ret"
    mov rdi, 10
    call add3
    DUMP
    lea rbx, [add3]
    call rbx                    ; through a register
    xor ebx, ebx
    DUMP
    call [fptr]                 ; through memory
    DUMP
    push 7
    push 8
    call sum2                   ; callee pops arguments with ret 16
    DUMP

    LABEL "enter/leave"
    mov rbp, 0x99
    call framed
    DUMP
    mov rdi, 5
    call fact
    DUMP FL_MUL                 ; flags come from imul
    EXIT 0

add3:
    add rdi, 3
    ret

sum2:                           ; returns [rsp+8] + [rsp+16], pops both
    mov rax, [rsp + 8]
    add rax, [rsp + 16]
    DUMP                        ; inside a call: sp is 8 lower
    ret 16

framed:
    enter 32, 0
    mov qword [rbp - 8], 0x55
    mov rax, [rbp - 8]
    mov rcx, rbp                ; rebase rbp to an offset for the dump
    sub rcx, rsp
    push rbp
    xor ebp, ebp
    DUMP
    pop rbp
    xor ecx, ecx
    leave
    ret

fact:                           ; rax = rdi!
    push rbp
    mov rbp, rsp
    cmp rdi, 1
    jbe .one
    push rdi
    dec rdi
    call fact
    pop rdi
    imul rax, rdi
    jmp .done
.one:
    mov eax, 1
    push rbp
    xor ebp, ebp
    DUMP                        ; deepest point of the recursion
    pop rbp
.done:
    mov rsp, rbp
    pop rbp
    ret
