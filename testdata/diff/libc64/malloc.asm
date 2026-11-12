; malloc, calloc, realloc, free: glibc's chunk layout and reuse rules.
default rel
extern malloc, calloc, realloc, free, printf, strcpy, strdup, puts
section .data
fmt1    db "aligned=%ld gap=%ld gap2=%ld", 10, 0
fmt2    db "reuse=%d calloc=%d", 10, 0
txt     db "keep me", 0
section .bss
p1 resq 1
p2 resq 1
p3 resq 1
p4 resq 1
section .text
global main
main:
    push rbp
    mov rbp, rsp
    mov edi, 10
    call malloc
    mov [p1], rax
    mov edi, 10
    call malloc
    mov [p2], rax
    mov edi, 100
    call malloc
    mov [p3], rax
    ; glibc: 16-byte aligned, 0x20-byte chunks for small requests
    mov rsi, [p1]
    and rsi, 15
    mov rdx, [p2]
    sub rdx, [p1]
    mov rcx, [p3]
    sub rcx, [p2]
    lea rdi, [fmt1]
    xor eax, eax
    call printf

    ; a freed chunk is handed out again for the same size (tcache)
    mov rdi, [p1]
    call free
    mov edi, 10
    call malloc
    xor ebx, ebx
    cmp rax, [p1]
    sete bl
    ; calloc returns zeroed memory, even when reusing a chunk
    mov rdi, [p3]
    mov byte [rdi+8], 99
    call free
    mov edi, 1
    mov esi, 100
    call calloc
    movzx edx, byte [rax+8]
    mov esi, ebx
    lea rdi, [fmt2]
    xor eax, eax
    call printf

    ; realloc keeps the contents
    mov rdi, [p2]
    lea rsi, [txt]
    call strcpy
    mov rdi, [p2]
    mov esi, 1000
    call realloc
    mov rdi, rax
    call puts
    lea rdi, [txt]
    call strdup
    mov rdi, rax
    call puts
    mov edi, 0
    call free                   ; free(NULL) does nothing
    xor eax, eax
    leave
    ret
