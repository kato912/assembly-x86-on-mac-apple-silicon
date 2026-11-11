; Differential test: addressing modes — RIP-relative, [base], [base+disp],
; [base+index*scale+disp], negative displacements, lea arithmetic,
; movzx/movsx/movsxd from memory and 32-bit writes zero-extending.
default rel
%include "dump64.inc"

section .data
arr:    dq 10, 20, 30, 40, 50, 60, 70, 80
words:  dw 0x8001, 0x7fff, 0xffff, 0x0001
bytes:  db 0x80, 0x7f, 0xff, 0x01
dwords: dd -2, 0x7fffffff, 0x80000000, 5

section .text
global _start
_start:
    INIT
    LABEL "rip-relative"
    mov rax, [arr]
    mov ebx, [dwords + 4]
    movzx ecx, word [words + 2]
    movsx rdx, byte [bytes]
    DUMP

    LABEL "base / base+disp / base+index*scale+disp"
    lea rbx, [arr]
    mov rax, [rbx]
    mov rcx, [rbx + 8]
    mov rdx, [rbx + 0x38]
    mov esi, 3
    mov rdi, [rbx + rsi*8]
    mov r8, [rbx + rsi*8 + 16]
    mov r9, [rbx + rsi*8 - 8]
    lea r10, [rbx + 64]
    mov r11, [r10 - 8]          ; negative disp8
    mov r12, [r10 - 0x40]       ; negative disp8 to the first element
    mov r13, 2
    mov r14, [rbx + r13*4 + 8]  ; scale 4: arr+16
    mov r15, [rbx + r13*2 + 4]  ; scale 2: arr+8
    xor ebx, ebx
    xor r10d, r10d
    DUMP

    LABEL "lea arithmetic"
    mov rax, 7
    mov rcx, 5
    lea rdx, [rax + rcx]
    lea rsi, [rax + rax*2]      ; *3
    lea rdi, [rax*8 + 3]
    lea r8, [rcx + rax*4 - 100]
    lea r9d, [rax + rcx*8 + 1]  ; 32-bit lea zero-extends
    mov r10, -1
    lea r10w, [rax + 1]         ; 16-bit lea keeps the upper bits
    lea r11, [rax - 0x12345]    ; negative disp32
    DUMP

    LABEL "index registers r8-r15"
    lea r15, [arr]
    mov r14, 5
    mov rax, [r15 + r14*8]
    mov ecx, 1
    mov rbx, [r15 + rcx*8]
    lea r13, [r15 + r14*8]
    sub r13, r15
    xor r15d, r15d
    DUMP

    LABEL "stores through all forms"
    lea rbx, [arr]
    mov rsi, 1
    mov qword [rbx], -1
    mov dword [rbx + 8], -1     ; only the low half of arr[1]
    mov word [rbx + rsi*8 + 8], 0x1234
    mov byte [rbx + rsi*8 + 15], 0x56
    mov rax, [rbx]
    mov rcx, [rbx + 8]
    mov rdx, [rbx + 16]
    xor ebx, ebx
    DUMP

    LABEL "movzx/movsx/movsxd"
    lea rsi, [words]
    movzx eax, word [rsi]
    movsx ebx, word [rsi]
    movsx rcx, word [rsi + 6]
    movzx edx, byte [rsi + 8 + 2]
    lea rdi, [dwords]
    movsxd r8, dword [rdi]
    movsxd r9, dword [rdi + 8]
    mov r10d, [rdi + 8]
    movsx r11w, byte [rsi + 8]
    xor esi, esi
    xor edi, edi
    DUMP

    LABEL "32-bit writes clear the upper half"
    mov rax, -1
    mov eax, eax
    mov rbx, -1
    add ebx, 0
    mov rcx, -1
    xor ecx, 0
    mov rdx, -1
    mov dx, 1                   ; 16-bit write keeps the upper bits
    mov rsi, -1
    mov sil, 0
    mov rdi, -1
    xchg edi, edi               ; even xchg of a 32-bit register with itself
    DUMP
    EXIT 0
