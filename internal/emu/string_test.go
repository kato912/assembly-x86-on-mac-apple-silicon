package emu

import "testing"

func TestStringOps(t *testing.T) {
	c := run(t, 32, `
	mov esi, src
	mov edi, dst
	mov ecx, 5
	rep movsb
	mov eax, [dst]
	mov ebx, ecx
	mov edi, src
	xor al, al
	mov ecx, -1
	repne scasb
	not ecx
	dec ecx
	mov edx, ecx
	mov esi, src
	mov edi, other
	mov ecx, 5
	repe cmpsb
	mov ebp, ecx
	jmp over
src:   db "hello", 0
other: db "help!", 0
dst:   times 8 db 0
over:
`)
	expect(t, c, RAX, 0x6c6c6500) // "hell", then xor al, al
	expect(t, c, RBX, 0)
	expect(t, c, RDX, 5)
	expect(t, c, RBP, 1)
	expectFlags(t, c, ZF, 0)

	c = run(t, 64, `
	default rel
	lea rdi, [buf]
	mov rax, 0x1122334455667788
	mov ecx, 3
	rep stosq
	lea rsi, [buf]
	lodsq
	mov rbx, rax
	lodsb
	movzx edx, al
	jmp over
buf: times 4 dq 0
over:
`)
	expect(t, c, RBX, 0x1122334455667788)
	expect(t, c, RDX, 0x88)
	expect(t, c, RCX, 0)
}
