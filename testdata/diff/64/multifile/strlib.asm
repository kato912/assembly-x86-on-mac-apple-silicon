; Third file of the 64-bit multifile test: string functions.
default rel
global strlen_, upcase

section .text
strlen_:                        ; rax = length of the string at rdi
    xor eax, eax
.loop:
    cmp byte [rdi + rax], 0
    je .done
    inc rax
    jmp .loop
.done:
    ret

upcase:                         ; copy the string at rdi to rsi, upper-cased
    mov al, [rdi]
    cmp al, 'a'
    jb .store
    cmp al, 'z'
    ja .store
    sub al, 32
.store:
    mov [rsi], al
    inc rdi
    inc rsi
    test al, al
    jnz upcase
    ret
