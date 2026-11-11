; Differential test: 32-bit addressing modes and extensions from memory.
%include "dump32.inc"

section .data
arr:    dd 10, 20, 30, 40, 50, 60, 70, 80
words:  dw 0x8001, 0x7fff
bytes:  db 0x80, 0x7f, 0xff, 0x01

section .text
global _start
_start:
    INIT
    LABEL "addressing modes"
    mov eax, [arr]                  ; [disp]
    mov ebx, arr
    mov ecx, [ebx]                  ; [base]
    mov edx, [ebx + 12]             ; [base + disp]
    mov esi, 2
    mov edi, [ebx + esi*4]          ; [base + index*scale]
    mov ebp, [ebx + esi*8 + 4]      ; [base + index*scale + disp]
    sub ebx, arr
    DUMP
    mov eax, [arr + esi*4]          ; [index*scale + disp]
    mov ecx, [arr + esi*2 + 2]      ; scale 2
    lea edx, [arr + 32]
    mov edx, [edx - 4]              ; negative displacement
    lea ebx, [arr + 32]
    mov edi, [ebx - 0x20]
    xor ebx, ebx
    xor ebp, ebp
    DUMP
    LABEL "lea"
    mov eax, 7
    mov ecx, 5
    lea edx, [eax + ecx]
    lea esi, [eax + eax*4]
    lea edi, [eax*8 - 1]
    lea ebx, [ecx + eax*2 + 100]
    lea bp, [eax + 1]
    DUMP
    LABEL "stores and extensions"
    mov dword [arr], -1
    mov word [arr + 4], 0x1234
    mov byte [arr + 8], 0x56
    mov eax, [arr]
    mov ebx, [arr + 4]
    mov ecx, [arr + 8]
    movzx edx, word [words]
    movsx esi, word [words]
    movzx edi, byte [bytes]
    movsx ebp, byte [bytes]
    DUMP
    movsx ax, byte [bytes + 2]
    movzx bx, byte [bytes + 2]
    xor ecx, ecx
    xor edx, edx
    xor esi, esi
    xor edi, edi
    xor ebp, ebp
    DUMP
    EXIT 0
