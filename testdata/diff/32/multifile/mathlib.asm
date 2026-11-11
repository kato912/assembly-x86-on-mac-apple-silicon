; Second file of the multifile test: functions and data used by main.asm.
global square, sum_array, bump, counter, table, table_len
extern base_value

section .data
counter   dd 5
table     dd 3, 1, 4, 1, 5, 9, 2, 6
table_len dd 8

section .text
square:                         ; eax = eax * eax
    imul eax, eax
    ret

sum_array:                      ; eax = sum of ecx dwords at esi
    xor eax, eax
.next:
    add eax, [esi]
    add esi, 4
    loop .next
    ret

bump:                           ; counter += base_value
    mov eax, [base_value]
    add [counter], eax
    ret
