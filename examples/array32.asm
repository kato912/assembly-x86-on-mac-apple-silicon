; array32.asm — sum, minimum and maximum of an array (32-bit)
;   run:  nasmx86 array32.asm
;   try:  nasmx86 --step array32.asm   then  m numbers 40   to see the array

section .data
    numbers dd  12, -7, 45, 3, 28, -15, 9, 33, 0, 21
    count   equ ($ - numbers) / 4   ; 4 bytes per dword
    s_sum   db  "sum: "
    s_min   db  "min: "
    s_max   db  "max: "

section .bss
    digits  resb 12

section .text
    global _start

_start:
    mov     esi, numbers        ; esi -> first element
    mov     ecx, count          ; elements left
    xor     eax, eax            ; sum
    mov     ebx, [esi]          ; min = numbers[0]
    mov     edx, [esi]          ; max = numbers[0]
.next:
    mov     edi, [esi]          ; edi = current element
    add     eax, edi
    cmp     edi, ebx
    jge     .not_min            ; signed compare
    mov     ebx, edi
.not_min:
    cmp     edi, edx
    jle     .not_max
    mov     edx, edi
.not_max:
    add     esi, 4              ; next dword
    loop    .next

    push    edx                 ; save max and min
    push    ebx
    mov     ecx, s_sum
    call    print_labeled       ; sum: 129
    pop     eax
    mov     ecx, s_min
    call    print_labeled       ; min: -15
    pop     eax
    mov     ecx, s_max
    call    print_labeled       ; max: 45

    mov     eax, 1
    xor     ebx, ebx
    int     0x80

; print_labeled: print the 5-byte label at ecx, then eax as a signed number.
print_labeled:
    pushad
    push    eax
    mov     eax, 4              ; write(1, ecx, 5)
    mov     ebx, 1
    mov     edx, 5
    int     0x80
    pop     eax
    mov     esi, digits + 11
    mov     byte [esi], 10
    mov     edi, eax            ; remember the sign
    test    eax, eax
    jns     .positive
    neg     eax                 ; work with the absolute value
.positive:
    mov     ebx, 10
.digit:
    xor     edx, edx
    div     ebx
    add     dl, '0'
    dec     esi
    mov     [esi], dl
    test    eax, eax
    jnz     .digit
    test    edi, edi
    jns     .print
    dec     esi
    mov     byte [esi], '-'
.print:
    mov     ecx, esi
    mov     edx, digits + 12
    sub     edx, esi
    mov     eax, 4
    mov     ebx, 1
    int     0x80
    popad
    ret
