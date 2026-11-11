; dump64.inc — helpers for nasmx86 differential tests (64-bit Linux, syscall).
; Same interface as dump32.inc; see there for details.
;
; All data references are RIP-relative so the tests can also be built as
; position-independent code (define MACOS to build for macOS, which lets
; the 64-bit tests be checked on Apple Silicon under Rosetta 2 too).

%define FL_ALL    0xCD5
%define FL_LOGIC  0xCC5
%define FL_MUL    0xC01
%define FL_NONE   0x400
%define FL_SHIFTN 0x4C5
%define FL_ROTN   0x4D5
%define FL_BT     0x441
%define FL_BSF    0x440

%ifdef MACOS
  %define SYS_WRITE 0x2000004
  %define SYS_EXIT  0x2000001
%else
  %define SYS_WRITE 1
  %define SYS_EXIT  60
%endif

section .data
__dmask:  dq FL_ALL
__sp0:    dq 0
__hexdig: db "0123456789abcdef"
__names:  db "rax=rbx=rcx=rdx=rsi=rdi=rbp=r8 =r9 =r10=r11=r12=r13=r14=r15="
section .bss
__buf:    resb 512
section .text

; INIT records the starting rsp and zeroes every register and flag, so the
; program starts from the same state on every platform.
%macro INIT 0
    mov [rel __sp0], rsp
    xor eax, eax
    xor ebx, ebx
    xor ecx, ecx
    xor edx, edx
    xor esi, esi
    xor edi, edi
    xor ebp, ebp
    xor r8d, r8d
    xor r9d, r9d
    xor r10d, r10d
    xor r11d, r11d
    xor r12d, r12d
    xor r13d, r13d
    xor r14d, r14d
    xor r15d, r15d
    push qword 0x202
    popfq
%endmacro

; FLAGS value: set the flags to a known pattern (so undefined flags left by
; an earlier instruction cannot leak into a later, fully compared dump).
%macro FLAGS 1
    push qword %1
    popfq
%endmacro

%macro DUMP 0-1 FL_ALL
    mov qword [rel __dmask], %1
    call __dump
%endmacro

%macro EXIT 1
    mov eax, SYS_EXIT
    mov edi, %1
    syscall
%endmacro

%macro LABEL 1
    jmp %%skip
%%str: db %1, 10
%%len equ $ - %%str
%%skip:
    push qword %%len
    push rax                    ; placeholder, replaced by the address
    lea rax, [rel %%str]
    xchg rax, [rsp]
    call __write
%endmacro

%macro PRINTMEM 2               ; addr (label), len
    push qword %2
    push rax
    lea rax, [rel %1]
    xchg rax, [rsp]
    call __write
%endmacro

; __write(ptr, len) — callee pops its arguments, preserves everything.
__write:
    pushfq
    push rax
    push rcx
    push rdx
    push rsi
    push rdi
    push r11
    mov rsi, [rsp + 64]
    mov rdx, [rsp + 72]
    mov eax, SYS_WRITE
    mov edi, 1
    syscall
    pop r11
    pop rdi
    pop rsi
    pop rdx
    pop rcx
    pop rax
    popfq
    ret 16

__dump:
    pushfq
    push r15
    push r14
    push r13
    push r12
    push r11
    push r10
    push r9
    push r8
    push rbp
    push rdi
    push rsi
    push rdx
    push rcx
    push rbx
    push rax
    mov rbp, rsp
    lea rdi, [rel __buf]
    xor ecx, ecx
.reg:
    lea rsi, [rel __names]
    mov eax, [rsi + rcx*4]
    mov [rdi], eax
    add rdi, 4
    mov rax, [rbp + rcx*8]
    mov r9d, 16
    call __hex
    mov byte [rdi], ' '
    inc rdi
    inc ecx
    cmp ecx, 15
    jb .reg
    mov dword [rdi], 'fl= '
    add rdi, 3
    mov rax, [rbp + 15*8]
    and rax, [rel __dmask]
    mov r9d, 3
    call __hex
    mov dword [rdi], ' sp='
    add rdi, 4
    lea rax, [rbp + 17*8]
    sub rax, [rel __sp0]
    mov r9d, 16
    call __hex
    mov byte [rdi], 10
    inc rdi
    lea rsi, [rel __buf]
    mov rdx, rdi
    sub rdx, rsi
    mov eax, SYS_WRITE
    mov edi, 1
    syscall
    pop rax
    pop rbx
    pop rcx
    pop rdx
    pop rsi
    pop rdi
    pop rbp
    pop r8
    pop r9
    pop r10
    pop r11
    pop r12
    pop r13
    pop r14
    pop r15
    popfq
    ret

; __hex: write rax as r9 hex digits at rdi, advancing rdi.
__hex:
    lea r10, [rel __hexdig]
    mov r11, r9
.d: dec r11
    mov rdx, rax
    and edx, 15
    mov dl, [r10 + rdx]
    mov [rdi + r11], dl
    shr rax, 4
    test r11, r11
    jnz .d
    add rdi, r9
    ret
