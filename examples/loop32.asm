; loop32.asm — print the numbers 1 to 10 and their sum (32-bit)
;   run:  nasmx86 loop32.asm
;   try:  nasmx86 --trace loop32.asm     (watch the loop run)

section .data
    total   db  "sum = "
    totlen  equ $ - total

section .bss
    digits  resb 12             ; room for a 10-digit number + newline

section .text
    global _start

_start:
    mov     ecx, 1              ; counter i = 1
    xor     edi, edi            ; sum = 0
.loop:
    add     edi, ecx            ; sum += i
    mov     eax, ecx
    call    print_number        ; print i
    inc     ecx
    cmp     ecx, 10
    jle     .loop               ; while i <= 10

    mov     eax, 4              ; write(1, "sum = ", 6)
    mov     ebx, 1
    mov     ecx, total
    mov     edx, totlen
    int     0x80
    mov     eax, edi
    call    print_number        ; 55

    mov     eax, 1              ; exit(0)
    xor     ebx, ebx
    int     0x80

; print_number: print eax as an unsigned decimal number and a newline.
; All registers are preserved.
print_number:
    pushad
    mov     esi, digits + 11    ; fill the buffer from the end
    mov     byte [esi], 10      ; newline
    mov     ebx, 10
.next_digit:
    xor     edx, edx            ; edx:eax / 10 -> eax = quotient, edx = remainder
    div     ebx
    add     dl, '0'             ; remainder -> ASCII digit
    dec     esi
    mov     [esi], dl
    test    eax, eax
    jnz     .next_digit
    mov     ecx, esi            ; write(1, esi, digits + 12 - esi)
    mov     edx, digits + 12
    sub     edx, esi
    mov     eax, 4
    mov     ebx, 1
    int     0x80
    popad
    ret
