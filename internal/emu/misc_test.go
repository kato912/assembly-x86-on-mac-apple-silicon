package emu

import "testing"

func TestEnterLeave(t *testing.T) {
	c := run(t, 32, `
	mov ebp, 0x1234
	call f
	jmp over
f:
	enter 16, 0
	mov eax, esp
	mov ebx, ebp
	leave
	ret
over:
`)
	// call pushed 4 bytes, enter pushed ebp (4) and reserved 16.
	expect(t, c, RBX, testStack-8)
	expect(t, c, RAX, testStack-8-16)
	expect(t, c, RBP, 0x1234)
	expect(t, c, RSP, testStack)
}

func TestPushadPopadPushfPopf(t *testing.T) {
	c := run(t, 32, `
	mov eax, 1
	mov ecx, 2
	mov edx, 3
	mov ebx, 4
	mov ebp, 5
	mov esi, 6
	mov edi, 7
	pushad
	mov eax, [esp + 12]   ; saved esp
	mov [saved], eax
	xor eax, eax
	xor ecx, ecx
	xor edx, edx
	xor ebx, ebx
	xor ebp, ebp
	xor esi, esi
	xor edi, edi
	popad
	stc
	pushfd
	clc
	popfd                 ; CF back to 1
	jmp over
saved: dd 0
over:
	mov eax, [saved]
`)
	expect(t, c, RAX, testStack)
	expect(t, c, RCX, 2)
	expect(t, c, RDX, 3)
	expect(t, c, RBX, 4)
	expect(t, c, RBP, 5)
	expect(t, c, RSI, 6)
	expect(t, c, RDI, 7)
	expect(t, c, RSP, testStack)
	expectFlags(t, c, CF, CF)

	c = run(t, 64, `
	pushfq
	pop rax
	or rax, 0x400         ; set DF via popf
	push rax
	popfq
`)
	expectFlags(t, c, DF, DF)
}

func TestLoops(t *testing.T) {
	c := run(t, 32, `
	xor eax, eax
	mov ecx, 5
l1:
	add eax, ecx
	loop l1               ; eax = 15
	mov ebx, 10
	mov ecx, 0
	jecxz skip
	mov ebx, 0
skip:
	; loopne: find the first zero byte (stops at index 2)
	mov esi, data - 1
	mov ecx, 4
l2:
	inc esi
	cmp byte [esi], 0
	loopne l2
	sub esi, data
	jmp over
data: db 1, 2, 0, 3
over:
`)
	expect(t, c, RAX, 15)
	expect(t, c, RBX, 10)
	expect(t, c, RSI, 2)
	expect(t, c, RCX, 1) // loopne decrements before testing ZF

	c = run(t, 64, `
	mov rcx, 3
	xor eax, eax
l:	inc eax
	loop l
	jrcxz done
	mov eax, 99
done:
`)
	expect(t, c, RAX, 3)
}

func TestFlagOps(t *testing.T) {
	c := run(t, 32, `
	stc
	cmc
	setc al
	stc
	setc bl
	clc
	setc cl
	std
	cld
`)
	if c.R[RAX]&0xff != 0 || c.R[RBX]&0xff != 1 || c.R[RCX]&0xff != 0 {
		t.Errorf("al=%d bl=%d cl=%d", c.R[RAX]&0xff, c.R[RBX]&0xff, c.R[RCX]&0xff)
	}
	expectFlags(t, c, DF, 0)
}

func TestStringBackwards(t *testing.T) {
	c := run(t, 32, `
	std
	mov esi, src + 3
	mov edi, dst + 3
	mov ecx, 4
	rep movsb
	cld
	mov eax, [dst]
	jmp over
src: db "abcd"
dst: dd 0
over:
`)
	expect(t, c, RAX, 0x64636261)
}
