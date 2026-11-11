package emu

import "testing"

func TestShiftFlags(t *testing.T) {
	cases := []struct {
		name  string
		op    ShiftOp
		a     uint64
		cnt   uint64
		size  int
		cf    bool
		res   uint64
		flags uint64 // expected values of the flags in check
		check uint64
	}{
		{"shl al,1", SHL, 0x81, 1, 1, false, 0x02, CF | OF, Arith &^ AF},
		{"shl eax,31", SHL, 1, 31, 4, false, 0x80000000, SF | PF, CF | SF | ZF | PF},
		{"shl eax,4", SHL, 0xffffffff, 4, 4, false, 0xfffffff0, CF | SF | PF, CF | SF | ZF | PF},
		{"shl al,8", SHL, 0x01, 8, 1, false, 0, CF | ZF | PF, CF | ZF | PF},
		{"shl rax,63", SHL, 1, 63, 8, false, 1 << 63, SF | PF, CF | SF | ZF | PF},
		{"shr al,1", SHR, 0x81, 1, 1, false, 0x40, CF | OF, Arith &^ AF},
		{"shr eax,4", SHR, 0x80000018, 4, 4, false, 0x08000001, CF, CF | SF | ZF},
		{"sar al,1", SAR, 0x81, 1, 1, false, 0xc0, CF | SF | PF, Arith &^ AF},
		{"sar eax,31", SAR, 0x80000000, 31, 4, false, 0xffffffff, SF | PF, CF | SF | ZF | PF},
		{"sar al,20", SAR, 0x80, 20, 1, false, 0xff, CF | SF | PF, CF | SF | ZF | PF},
		{"rol al,1", ROL, 0x81, 1, 1, false, 0x03, CF | OF, CF | OF},
		{"ror al,1", ROR, 0x81, 1, 1, false, 0xc0, CF, CF | OF},
		{"rol eax,8", ROL, 0x12345678, 8, 4, false, 0x34567812, 0, CF},
		{"ror eax,4", ROR, 0x12345678, 4, 4, false, 0x81234567, CF, CF},
		{"rcl al,1", RCL, 0x80, 1, 1, false, 0x00, CF | OF, CF | OF},
		{"rcl al,1 cf", RCL, 0x00, 1, 1, true, 0x01, 0, CF | OF},
		{"rcr al,1", RCR, 0x01, 1, 1, true, 0x80, CF | OF, CF | OF},
		{"rcr eax,4", RCR, 0x0000000f, 4, 4, false, 0xe0000000, CF, CF},
		{"rcl al,9 (no-op)", RCL, 0x5a, 9, 1, true, 0x5a, CF, CF},
	}
	for _, c := range cases {
		res, f, which := Shift(c.op, c.a, c.cnt, c.size, c.cf)
		if res != c.res {
			t.Errorf("%s: result %#x want %#x", c.name, res, c.res)
		}
		if which&c.check != c.check&which || f&c.check != c.flags {
			t.Errorf("%s: flags %s want %s", c.name, FormatFlags(f&c.check), FormatFlags(c.flags))
		}
	}
	// A masked count of zero leaves every flag alone.
	if _, _, which := Shift(SHL, 1, 32, 4, false); which != 0 {
		t.Errorf("count 32 on 32-bit should not touch flags")
	}
	if _, _, which := Shift(ROL, 1, 64, 8, false); which != 0 {
		t.Errorf("count 64 on 64-bit should not touch flags")
	}
}

func TestDoubleShift(t *testing.T) {
	res, f, _ := DoubleShift(true, 0x12345678, 0x9abcdef0, 8, 4)
	if res != 0x3456789a || f&(CF|OF) != 0 {
		t.Errorf("shld: %#x %s", res, FormatFlags(f))
	}
	res, f, _ = DoubleShift(false, 0x12345678, 0x9abcdef0, 8, 4)
	if res != 0xf0123456 || f&(CF|SF) != SF {
		t.Errorf("shrd: %#x %s", res, FormatFlags(f))
	}
	res, _, _ = DoubleShift(true, 0x0123456789abcdef, 0xfedcba9876543210, 4, 8)
	if res != 0x123456789abcdeff {
		t.Errorf("shld 64: %#x", res)
	}
}

func TestShiftInstructions(t *testing.T) {
	c := run(t, 32, `
	mov eax, 1
	mov cl, 5
	shl eax, cl        ; 32
	mov ebx, eax
	sal ebx, 1         ; 64
	mov edx, -64
	sar edx, 2         ; -16
	mov esi, 0x80000000
	shr esi, 31        ; 1
	mov edi, 0x12345678
	rol edi, 4         ; 0x23456781
	mov ebp, 0x12345678
	shrd ebp, edi, 8   ; 0x81123456
	mov word [buf], 0x8001
	shl word [buf], 1  ; 0x0002 CF=1
	movzx ecx, word [buf]
	jmp over
buf: dw 0
over:
`)
	expect(t, c, RAX, 32)
	expect(t, c, RBX, 64)
	expect(t, c, RDX, 0xfffffff0)
	expect(t, c, RSI, 1)
	expect(t, c, RDI, 0x23456781)
	expect(t, c, RBP, 0x81123456)
	expect(t, c, RCX, 2)
	expectFlags(t, c, CF, CF)

	// A 32-bit shift by zero still zero-extends the destination in 64-bit mode.
	c = run(t, 64, `
	mov rax, -1
	xor ecx, ecx
	shl eax, cl
`)
	expect(t, c, RAX, 0xffffffff)
}
