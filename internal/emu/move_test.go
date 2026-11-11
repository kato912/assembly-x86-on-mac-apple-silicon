package emu

import "testing"

func TestExtendMoves(t *testing.T) {
	c := run(t, 64, `
	mov rax, -1
	mov al, 0x80
	movzx ebx, al          ; 0x80
	movsx rcx, al          ; 0xffffffffffffff80
	movsx edx, al          ; 0xffffff80 (zero-extended into rdx)
	mov r8, 0x8000
	movzx r9, r8w          ; 0x8000
	movsx r10, r8w         ; 0xffffffffffff8000
	mov r11d, 0x80000000
	movsxd r12, r11d       ; 0xffffffff80000000
	movzx esi, byte [rel bytes + 1]   ; 0xfe
	movsx edi, word [rel bytes]       ; 0xfffffeff
	jmp over
bytes: db 0xff, 0xfe
over:
`)
	expect(t, c, RBX, 0x80)
	expect(t, c, RCX, 0xffffffffffffff80)
	expect(t, c, RDX, 0xffffff80)
	expect(t, c, R9, 0x8000)
	expect(t, c, R10, 0xffffffffffff8000)
	expect(t, c, R12, 0xffffffff80000000)
	expect(t, c, RSI, 0xfe)
	expect(t, c, RDI, 0xfffffeff)
}

func TestAccumulatorExtend(t *testing.T) {
	c := run(t, 32, `
	mov eax, 0x12345680
	cbw                ; ax = 0xff80
	mov ebx, eax       ; 0x1234ff80
	cwde               ; eax = 0xffffff80
	mov ecx, eax
	mov eax, -5
	cdq                ; edx = 0xffffffff
	mov esi, edx
	mov eax, 5
	cdq                ; edx = 0
	mov edi, edx
	mov ax, 0x8000
	cwd                ; dx = 0xffff
`)
	expect(t, c, RBX, 0x1234ff80)
	expect(t, c, RCX, 0xffffff80)
	expect(t, c, RSI, 0xffffffff)
	expect(t, c, RDI, 0)
	expect(t, c, RDX, 0xffff)

	c = run(t, 64, `
	mov eax, 0x80000000
	cdqe               ; rax = 0xffffffff80000000
	mov rbx, rax
	cqo                ; rdx = -1
`)
	expect(t, c, RBX, 0xffffffff80000000)
	expect(t, c, RDX, ^uint64(0))
}

func TestXchgXaddCmpxchg(t *testing.T) {
	c := run(t, 32, `
	mov eax, 1
	mov ebx, 2
	xchg eax, ebx
	mov ecx, 10
	mov [buf], ecx
	mov edx, 5
	xchg edx, [buf]    ; edx = 10, buf = 5
	mov esi, [buf]
	mov edi, 3
	xadd [buf], edi    ; buf = 8, edi = 5
	mov ebp, [buf]
	jmp over
buf: dd 0
over:
`)
	expect(t, c, RAX, 2)
	expect(t, c, RBX, 1)
	expect(t, c, RDX, 10)
	expect(t, c, RSI, 5)
	expect(t, c, RDI, 5)
	expect(t, c, RBP, 8)

	c = run(t, 64, `
	mov rax, 7
	mov rbx, 7
	mov rcx, 9
	cmpxchg rbx, rcx   ; equal: rbx = 9, ZF = 1
	mov rdx, rbx
	mov rax, 1
	cmpxchg rbx, rcx   ; not equal: rax = 9, ZF = 0
`)
	expect(t, c, RDX, 9)
	expect(t, c, RAX, 9)
	expectFlags(t, c, ZF, 0)
}

func TestXlatLahfSahf(t *testing.T) {
	c := run(t, 32, `
	mov ebx, table
	mov al, 2
	xlatb              ; al = 'c'
	movzx ecx, al
	mov dh, 0xff
	add dh, 1          ; CF = 1, ZF = 1
	mov ah, 0
	lahf               ; ah = CF | bit1 | (ZF/PF from earlier ops)
	movzx edx, ah
	mov ah, 0x41       ; ZF | CF
	sahf
	jmp over
table: db "abcd"
over:
`)
	expect(t, c, RCX, 'c')
	if c.R[RDX]&0x03 != 0x03 {
		t.Errorf("lahf ah = %#x", c.R[RDX])
	}
	expectFlags(t, c, ZF|CF|SF, ZF|CF)
}
