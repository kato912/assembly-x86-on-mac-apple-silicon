; Differential test: BCD adjustments daa, das, aaa, aas, aam, aad (32-bit
; only). Undefined flags are masked: OF for daa/das, everything but CF/AF
; for aaa/aas, OF/AF/CF for aam/aad. (Generated case list.)
%include "dump32.inc"

%define FL_DAA 0x4D5            ; all but OF
%define FL_AAA 0x411            ; CF AF
%define FL_AAM 0x4C4            ; SF ZF PF

section .text
global _start
_start:
    INIT
    LABEL "add + daa"
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0x99
    daa
    DUMP FL_DAA
    LABEL "adc + daa (carry in)"
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    stc
    adc al, 0x99
    daa
    DUMP FL_DAA
    LABEL "sub + das"
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x19
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x29
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x35
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x48
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x50
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x79
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x90
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x99
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0x9a
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0x99
    das
    DUMP FL_DAA
    LABEL "daa/das with flags preset"
    FLAGS 0x203
    mov eax, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x0
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x0
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x1
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x1
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x5
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x5
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x9
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x9
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x10
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x10
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x19
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x19
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x29
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x29
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x35
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x35
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x48
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x48
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x50
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x50
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x79
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x79
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x90
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x90
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x99
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x99
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0xa
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0xa
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0xa
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0xa
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0xa
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0xa
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x9a
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0x9a
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x9a
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0x9a
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x9a
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0x9a
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0xf9
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0xf9
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0xf9
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0xf9
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0xf9
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0xf9
    das
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0xff
    daa
    DUMP FL_DAA
    FLAGS 0x203
    mov eax, 0xff
    das
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0xff
    daa
    DUMP FL_DAA
    FLAGS 0x212
    mov eax, 0xff
    das
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0xff
    daa
    DUMP FL_DAA
    FLAGS 0x213
    mov eax, 0xff
    das
    DUMP FL_DAA
    LABEL "add + aaa"
    FLAGS 0x202
    mov eax, 0x0
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x0
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff00
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff00
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff00
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff00
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff01
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff01
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff01
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff01
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff04
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff04
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff04
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff04
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff05
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff05
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff05
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff05
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff08
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff08
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff08
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff08
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff09
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff09
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff09
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff09
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0a
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0a
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0a
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0a
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0f
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0f
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0f
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff0f
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff10
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff10
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff10
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff10
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xfff9
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xfff9
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xfff9
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xfff9
    add al, 9
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xffff
    add al, 0
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xffff
    add al, 1
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xffff
    add al, 5
    aaa
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xffff
    add al, 9
    aaa
    DUMP FL_AAA
    LABEL "sub + aas"
    FLAGS 0x202
    mov eax, 0x0
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x0
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x4
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x5
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x8
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xa
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xf9
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0xff
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x100
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x101
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x104
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x105
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x108
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x109
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10a
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x10f
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x110
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1f9
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x1ff
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x900
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x901
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x904
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x905
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x908
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x909
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90a
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x90f
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x910
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9f9
    sub al, 9
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    sub al, 0
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    sub al, 1
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    sub al, 5
    aas
    DUMP FL_AAA
    FLAGS 0x202
    mov eax, 0x9ff
    sub al, 9
    aas
    DUMP FL_AAA
    LABEL "aam / aad"
    FLAGS 0x202
    mov eax, 0x1200
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1200
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1200
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1200
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1201
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1201
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1201
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1201
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1209
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1209
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1209
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1209
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x120a
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x120a
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x120a
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x120a
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x123f
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x123f
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x123f
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x123f
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1263
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1263
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1263
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1263
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1264
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1264
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1264
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1264
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x127f
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x127f
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x127f
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x127f
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x12ff
    aam
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x12ff
    aam 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x12ff
    aam 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x12ff
    aam 1
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x0
    aad
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x0
    aad 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x0
    aad 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x603
    aad
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x603
    aad 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x603
    aad 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x909
    aad
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x909
    aad 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x909
    aad 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0xa0a
    aad
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0xa0a
    aad 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0xa0a
    aad 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1234
    aad
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1234
    aad 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0x1234
    aad 7
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0xffff
    aad
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0xffff
    aad 16
    DUMP FL_AAM
    FLAGS 0x202
    mov eax, 0xffff
    aad 7
    DUMP FL_AAM
    EXIT 0
