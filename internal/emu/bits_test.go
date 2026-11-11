package emu

import "testing"

func TestBitTest(t *testing.T) {
	c := run(t, 32, `
	mov eax, 0b1010
	bt eax, 1          ; CF = 1
	setc bl
	bt eax, 2          ; CF = 0
	setc bh
	bts eax, 0         ; 0b1011
	btr eax, 3         ; 0b0011
	btc eax, 4         ; 0b10011
	mov ecx, 35        ; bit 35 of an array = bit 3 of the second dword
	bts [arr], ecx
	mov edx, [arr + 4]
	mov esi, -1        ; bit -1 = top bit of the byte before arr
	bt [arr], esi      ; CF = 1 (0x80 in pad)
	setc ch
	jmp over
pad: db 0x80
arr: dd 0, 0
over:
`)
	expect(t, c, RAX, 0b10011)
	if c.R[RBX]&0xffff != 0x0001 {
		t.Errorf("bx = %#x", c.R[RBX]&0xffff)
	}
	expect(t, c, RDX, 8)
	if c.R[RCX]>>8&0xff != 1 {
		t.Errorf("ch = %#x", c.R[RCX]>>8&0xff)
	}
}

func TestBitScanBswap(t *testing.T) {
	c := run(t, 64, `
	mov eax, 0x00f0
	bsf ebx, eax       ; 4
	bsr ecx, eax       ; 7
	mov rdx, 77
	xor esi, esi
	bsf rdx, rsi       ; ZF = 1, rdx unchanged
	mov edi, 0x11223344
	bswap edi
	mov r8, 0x0102030405060708
	bswap r8
`)
	expect(t, c, RBX, 4)
	expect(t, c, RCX, 7)
	expect(t, c, RDX, 77)
	expect(t, c, RDI, 0x44332211)
	expect(t, c, R8, 0x0807060504030201)
	expectFlags(t, c, ZF, ZF)
}

func TestSetccCmovcc(t *testing.T) {
	c := run(t, 32, `
	mov eax, 5
	cmp eax, 7
	setl bl            ; 5 < 7 signed
	setb bh            ; 5 < 7 unsigned
	setg cl
	sete ch
	mov edx, -1
	cmp edx, 1
	setl dl            ; -1 < 1 signed  -> 1
	seta dh            ; 0xffffffff > 1 unsigned -> 1
	mov esi, 10
	mov edi, 20
	cmp esi, edi
	cmovl esi, edi     ; esi = 20
	mov ebp, 1
	cmovg ebp, edi     ; not taken
`)
	if c.R[RBX]&0xffff != 0x0101 || c.R[RCX]&0xffff != 0 || c.R[RDX]&0xffff != 0x0101 {
		t.Errorf("bx=%#x cx=%#x dx=%#x", c.R[RBX]&0xffff, c.R[RCX]&0xffff, c.R[RDX]&0xffff)
	}
	expect(t, c, RSI, 20)
	expect(t, c, RBP, 1)

	c = run(t, 64, `
	mov rax, -1
	mov ebx, 3
	test ebx, ebx
	cmovz eax, ebx     ; condition false, but eax is still zero-extended
`)
	expect(t, c, RAX, 0xffffffff)
}
