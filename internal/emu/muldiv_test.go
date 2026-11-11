package emu

import "testing"

func TestImulDivideUnit(t *testing.T) {
	imul := []struct {
		a, b     uint64
		size     int
		res      uint64
		overflow bool
	}{
		{3, 4, 4, 12, false},
		{0xffffffff, 2, 4, 0xfffffffe, false}, // -1 * 2 = -2
		{0x40000000, 2, 4, 0x80000000, true},
		{0x80, 0xff, 1, 0x80, true},                  // -128 * -1 = 128 overflows int8
		{0x7f, 0x02, 1, 0xfe, true},                  // 254 doesn't fit
		{0xffff, 0xffff, 2, 1, false},                // -1 * -1
		{1 << 62, 2, 8, 1 << 63, true},               // overflow into sign bit
		{^uint64(0), ^uint64(0), 8, 1, false},        // -1 * -1
		{^uint64(0) - 1, 1 << 62, 8, 1 << 63, false}, // -2 * 2^62 = -2^63 fits
	}
	for _, c := range imul {
		res, ov := Imul(c.a, c.b, c.size)
		if res != c.res || ov != c.overflow {
			t.Errorf("imul %#x*%#x/%d = %#x,%v want %#x,%v", c.a, c.b, c.size, res, ov, c.res, c.overflow)
		}
	}

	div := []struct {
		hi, lo, d uint64
		size      int
		signed    bool
		q, r      uint64
		ok        bool
	}{
		{0, 100, 7, 4, false, 14, 2, true},
		{1, 0, 2, 4, false, 0x80000000, 0, true}, // 2^32 / 2
		{2, 0, 2, 4, false, 0, 0, false},         // 2^33 / 2 overflows
		{0, 0xff, 2, 1, false, 0x7f, 1, true},
		{1, 0, 1, 1, false, 0, 0, false},                                   // 256 / 1 overflows al
		{0xffffffff, 0xfffffff9, 2, 4, true, 0xfffffffd, 0xffffffff, true}, // -7 / 2 = -3 rem -1
		{0, 7, 0xfffffffe, 4, true, 0xfffffffd, 1, true},                   // 7 / -2 = -3 rem 1
		{0xffffffff, 0x80000000, 0xffffffff, 4, true, 0, 0, false},         // INT_MIN / -1
		{0xffff, 0xff80, 0xffff, 2, true, 0x80, 0, true},                   // -128 / -1 = 128 fits in 16 bits
		{0xffff, 0x8000, 0xffff, 2, true, 0, 0, false},                     // -32768 / -1 overflows
		{0, 10, 3, 8, false, 3, 1, true},
		{^uint64(0), ^uint64(0) - 9, 3, 8, true, ^uint64(0) - 2, ^uint64(0), true}, // -10 / 3 = -3 rem -1
		{^uint64(0), 1 << 63, ^uint64(0), 8, true, 0, 0, false},                    // INT64_MIN / -1
	}
	for _, c := range div {
		q, r, ok := Divide(c.hi, c.lo, c.d, c.size, c.signed)
		if ok != c.ok || (ok && (q != c.q || r != c.r)) {
			t.Errorf("div %#x:%#x / %#x (size %d signed %v) = %#x r %#x ok %v; want %#x r %#x ok %v",
				c.hi, c.lo, c.d, c.size, c.signed, q, r, ok, c.q, c.r, c.ok)
		}
	}
}

func TestMulDivInstructions(t *testing.T) {
	c := run(t, 32, `
	mov eax, 0x10000
	mov ebx, 0x10000
	mul ebx            ; edx:eax = 0x1_00000000
	mov esi, edx
	mov edi, eax
	mov al, 200
	mov bl, 2
	mul bl             ; ax = 400
	movzx ebp, ax
`)
	expect(t, c, RSI, 1)
	expect(t, c, RDI, 0)
	expect(t, c, RBP, 400)
	expectFlags(t, c, CF|OF, CF|OF)

	c = run(t, 32, `
	mov eax, -6
	mov ecx, 7
	imul ecx           ; edx:eax = -42
	mov esi, edx
	mov edi, eax
	mov ebx, 5
	imul ebx, ebx, -3  ; -15
	mov ecx, 4
	imul ecx, ebx      ; -60
`)
	expect(t, c, RSI, 0xffffffff)
	expect(t, c, RDI, uint64(uint32(0xffffffd6)))
	expect(t, c, RBX, uint64(uint32(0xfffffff1)))
	expect(t, c, RCX, uint64(uint32(0xffffffc4)))
	expectFlags(t, c, CF|OF, 0)

	c = run(t, 64, `
	mov rax, 100
	xor edx, edx
	mov rcx, 7
	div rcx            ; q 14 r 2
	mov rbx, rax
	mov rsi, rdx
	mov rax, -100
	cqo
	idiv rcx           ; q -14 r -2
	mov ax, 1000
	mov dl, 7
	div dl             ; al = 142, ah = 6
	movzx edi, ax
`)
	expect(t, c, RBX, 14)
	expect(t, c, RSI, 2)
	expect(t, c, RDX, 0xffffffffffffff07) // -2 from idiv, then mov dl, 7
	expect(t, c, RDI, 6<<8|142)
}

func TestDivFaults(t *testing.T) {
	f := runFault(t, 32, `
	mov eax, 10
	xor ebx, ebx
	div ebx
`)
	if f.Kind != FaultDivZero || f.Div.Size != 4 {
		t.Errorf("got %+v", f)
	}
	f = runFault(t, 32, `
	mov eax, 10
	mov edx, 5         ; forgot xor edx, edx
	mov ebx, 2
	div ebx
`)
	if f.Kind != FaultDivOverflow || f.Div.Hi != 5 || f.Div.Lo != 10 || f.Div.Divisor != 2 {
		t.Errorf("got %+v", f)
	}
}

func TestNegNot(t *testing.T) {
	c := run(t, 32, `
	mov eax, 5
	neg eax
	mov ebx, eax
	mov ecx, 0
	neg ecx            ; CF = 0
	mov edx, 0x0f0f0f0f
	not edx
`)
	expect(t, c, RBX, 0xfffffffb)
	expect(t, c, RDX, 0xf0f0f0f0)
	expectFlags(t, c, CF|ZF, ZF)
}
