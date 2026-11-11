; dump32.inc — helpers for nasmx86 differential tests (32-bit Linux, int 0x80).
;
; A test program prints a line of registers + flags after each operation
; with DUMP. The same program is run natively (nasm + ld on x86 Linux) and
; through nasmx86, and the outputs must be identical.
;
;   INIT                 save the starting esp (call once at _start)
;   DUMP [mask]          print eax..ebp, (flags & mask), esp - starting esp
;   LABEL "text"         print a line of text
;   PRINTMEM addr, len   print len bytes of memory
;   EXIT code            exit(code)
;
; Every macro except EXIT preserves all registers and flags. Registers that
; hold addresses differ between the two runs, so tests clear or rebase them
; before DUMP. Flags that the CPU leaves undefined are masked out.

%define FL_ALL    0xCD5         ; CF PF AF ZF SF DF OF
%define FL_LOGIC  0xCC5         ; no AF (undefined after and/or/xor/test, shifts)
%define FL_MUL    0xC01         ; CF OF (mul/imul)
%define FL_NONE   0x400         ; DF only (div/idiv leave all arithmetic flags undefined)
%define FL_SHIFTN 0x4C5         ; shifts by >1: no AF, no OF
%define FL_ROTN   0x4D5         ; rotates by >1: no OF
%define FL_BT     0x441         ; bt*: CF (and ZF, which is unaffected)
%define FL_BSF    0x440         ; bsf/bsr: ZF

section .data
__dmask:  dd FL_ALL
__sp0:    dd 0
__hexdig: db "0123456789abcdef"
__names:  db "eax=ebx=ecx=edx=esi=edi=ebp="
__offs:   db 28, 16, 24, 20, 4, 0, 8      ; pushad slot of each register
section .bss
__buf:    resb 256
section .text

; INIT records the starting esp and zeroes every register and flag.
%macro INIT 0
    mov [__sp0], esp
    xor eax, eax
    xor ebx, ebx
    xor ecx, ecx
    xor edx, edx
    xor esi, esi
    xor edi, edi
    xor ebp, ebp
    push dword 0x202
    popfd
%endmacro

; FLAGS value: set the flags to a known pattern (so undefined flags left by
; an earlier instruction cannot leak into a later, fully compared dump).
%macro FLAGS 1
    push dword %1
    popfd
%endmacro

%macro DUMP 0-1 FL_ALL
    mov dword [__dmask], %1
    call __dump
%endmacro

%macro EXIT 1
    mov eax, 1
    mov ebx, %1
    int 0x80
%endmacro

%macro LABEL 1
    jmp %%skip
%%str: db %1, 10
%%len equ $ - %%str
%%skip:
    push dword %%len
    push dword %%str
    call __write
%endmacro

%macro PRINTMEM 2
    push dword %2
    push dword %1
    call __write
%endmacro

; __write(ptr, len) — callee pops its arguments, preserves everything.
__write:
    pushfd
    pushad
    mov ecx, [esp + 40]
    mov edx, [esp + 44]
    mov eax, 4
    mov ebx, 1
    int 0x80
    popad
    popfd
    ret 8

__dump:
    pushfd
    pushad
    mov ebp, esp
    mov edi, __buf
    xor ecx, ecx
.reg:
    mov eax, [__names + ecx*4]
    mov [edi], eax
    add edi, 4
    movzx ebx, byte [__offs + ecx]
    mov eax, [ebp + ebx]
    mov esi, 8
    call __hex
    mov byte [edi], ' '
    inc edi
    inc ecx
    cmp ecx, 7
    jb .reg
    mov dword [edi], 'fl= '
    add edi, 3
    mov eax, [ebp + 32]
    and eax, [__dmask]
    mov esi, 3
    call __hex
    mov dword [edi], ' sp='
    add edi, 4
    lea eax, [ebp + 40]
    sub eax, [__sp0]
    mov esi, 8
    call __hex
    mov byte [edi], 10
    inc edi
    mov ecx, __buf
    mov edx, edi
    sub edx, ecx
    mov eax, 4
    mov ebx, 1
    int 0x80
    popad
    popfd
    ret

; __hex: write eax as esi hex digits at edi, advancing edi.
__hex:
    mov edx, esi
.d: dec edx
    mov ebx, eax
    and ebx, 15
    mov bl, [__hexdig + ebx]
    mov [edi + edx], bl
    shr eax, 4
    test edx, edx
    jnz .d
    add edi, esi
    ret
