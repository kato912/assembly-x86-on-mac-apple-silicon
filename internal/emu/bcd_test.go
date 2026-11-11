package emu

import "testing"

func TestDaaDas(t *testing.T) {
	cases := []struct {
		name      string
		al, flags uint64
		sub       bool
		want      uint64
		cf, af    bool
	}{
		{"79+35=114", 0xae, 0, false, 0x14, true, true},
		{"35+48=83", 0x7d, 0, false, 0x83, false, true},
		{"29+09=38 (AF from add)", 0x32, AF, false, 0x38, false, true},
		{"99+01=100", 0x9a, 0, false, 0x00, true, true},
		{"50+50=100 (CF from add)", 0xa0, 0, false, 0x00, true, false},
		{"12+34=46 no adjust", 0x46, 0, false, 0x46, false, false},
		{"35-47=-12 -> 88", 0xee, CF | AF, true, 0x88, true, true},
		{"47-35=12", 0x12, 0, true, 0x12, false, false},
		{"40-01=39", 0x3f, AF, true, 0x39, false, true},
	}
	for _, c := range cases {
		al, f := Daa(c.al, c.flags, c.sub)
		if al != c.want || (f&CF != 0) != c.cf || (f&AF != 0) != c.af {
			t.Errorf("%s: al=%#x CF=%v AF=%v, want %#x CF=%v AF=%v", c.name, al, f&CF != 0, f&AF != 0, c.want, c.cf, c.af)
		}
	}
}

func TestAaaAas(t *testing.T) {
	cases := []struct {
		name      string
		ax, flags uint64
		sub       bool
		want      uint64
		carry     bool
	}{
		{"8+6=14", 0x000e, 0, false, 0x0104, true},
		{"9+9=18 (AF)", 0x0012, AF, false, 0x0108, true},
		{"3+4=7", 0x0007, 0, false, 0x0007, false},
		{"0x0105+0x05 -> 0x0200", 0x010a, 0, false, 0x0200, true},
		{"13-5=8", 0x01fe, AF, true, 0x0008, true},
		{"9-4=5", 0x0005, 0, true, 0x0005, false},
	}
	for _, c := range cases {
		ax, f := Aaa(c.ax, c.flags, c.sub)
		if ax != c.want || (f&CF != 0) != c.carry || (f&AF != 0) != c.carry {
			t.Errorf("%s: ax=%#x CF=%v, want %#x CF=%v", c.name, ax, f&CF != 0, c.want, c.carry)
		}
	}
}

func TestBCDInstructions(t *testing.T) {
	c := run(t, 32, `
	mov al, 0x79
	add al, 0x35
	daa                ; al = 0x14, CF = 1 (79 + 35 = 114)
	movzx ebx, al
	setc bh
	mov al, 63
	aam                ; ah = 6, al = 3
	movzx ecx, ax
	aad                ; al = 63, ah = 0
	movzx edx, ax
	mov ax, 0x0208     ; unpacked 28
	add al, 5          ; + 5
	aaa                ; ax = 0x0303 (33)
	movzx esi, ax
	mov al, 0x2b
	aam 16             ; any base: ah = 2, al = 0xb
	movzx edi, ax
`)
	expect(t, c, RBX, 0x0114)
	expect(t, c, RCX, 0x0603)
	expect(t, c, RDX, 63)
	expect(t, c, RSI, 0x0303)
	expect(t, c, RDI, 0x020b)

	f := runFault(t, 32, `
	mov al, 5
	aam 0
`)
	if f.Kind != FaultDivZero {
		t.Errorf("aam 0: %v", f)
	}
}
