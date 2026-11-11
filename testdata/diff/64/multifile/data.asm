; Second file of the 64-bit multifile test: data only.
global message, msg_copy

section .data
message: db "hello, world", 0

section .bss
msg_copy: resb 32
