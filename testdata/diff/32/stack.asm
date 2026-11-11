; Differential test: push/pop forms, pushad/popad, pushfd/popfd, call/ret,
; ret imm, cdecl calls with stack arguments, enter/leave, recursion.
%include "dump32.inc"

section .data
fptr:   dd twice
val:    dd 0x11223344

section .text
global _start
_start:
    INIT
    LABEL "push/pop"
    push 5
    push -1
    push 0x12345678
    push dword [val]
    push word 0x4142
    DUMP
    pop ax
    pop ebx
    pop ecx
    pop edx
    pop esi
    DUMP
    mov eax, esp
    push esp
    pop edi
    sub edi, eax
    xor eax, eax
    DUMP

    LABEL "pushad/popad"
    mov eax, 1
    mov ebx, 2
    mov ecx, 3
    mov edx, 4
    mov esi, 5
    mov edi, 6
    mov ebp, 7
    pushad
    mov eax, [esp + 12]         ; the esp value pushad saved
    sub eax, esp
    mov [val], eax
    xor eax, eax
    xor ebx, ebx
    xor ecx, ecx
    xor edx, edx
    xor esi, esi
    xor edi, edi
    xor ebp, ebp
    popad
    mov eax, [val]
    DUMP

    LABEL "pushfd/popfd"
    FLAGS 0x8d7
    pushfd
    FLAGS 0x202
    popfd
    DUMP
    pushfd
    pop eax
    and eax, 0xcd5
    DUMP

    LABEL "calls"
    mov eax, 21
    call twice
    DUMP
    mov ebx, twice
    call ebx
    xor ebx, ebx
    DUMP
    call [fptr]
    DUMP
    push 30                     ; cdecl: caller cleans up
    push 12
    call add2
    add esp, 8
    DUMP
    push 9                      ; stdcall-style: ret 8
    push 4
    call sub2
    DUMP

    LABEL "enter/leave and recursion"
    mov ebp, 0x99
    call framed
    DUMP
    push 6
    call fact
    add esp, 4
    DUMP FL_MUL
    EXIT 0

twice:
    add eax, eax
    ret

add2:
    push ebp
    mov ebp, esp
    mov eax, [ebp + 8]
    add eax, [ebp + 12]
    pop ebp
    ret

sub2:
    mov eax, [esp + 4]
    sub eax, [esp + 8]
    DUMP
    ret 8

framed:
    enter 16, 0
    mov dword [ebp - 4], 0x55
    mov eax, [ebp - 4]
    mov ecx, ebp
    sub ecx, esp
    push ebp
    xor ebp, ebp
    DUMP
    pop ebp
    xor ecx, ecx
    leave
    ret

fact:                           ; eax = n! for n = [esp+4]
    push ebp
    mov ebp, esp
    mov eax, [ebp + 8]
    cmp eax, 1
    jbe .done
    dec eax
    push eax
    call fact
    add esp, 4
    imul eax, [ebp + 8]
.done:
    pop ebp
    ret
