package emu

import (
	"math/bits"
	"strings"
)

// RFLAGS bits.
const (
	CF uint64 = 1 << 0
	PF uint64 = 1 << 2
	AF uint64 = 1 << 4
	ZF uint64 = 1 << 6
	SF uint64 = 1 << 7
	TF uint64 = 1 << 8
	IF uint64 = 1 << 9
	DF uint64 = 1 << 10
	OF uint64 = 1 << 11
	AC uint64 = 1 << 18
	ID uint64 = 1 << 21

	// Arith is the set of flags written by add/sub-style instructions.
	Arith = CF | PF | AF | ZF | SF | OF

	reservedOne uint64 = 1 << 1
	// InitialFlags is RFLAGS as Linux hands it to a new process.
	InitialFlags = reservedOne | IF
	// popfMask is what user code may change with popf.
	popfMask = Arith | DF | AC | ID
)

// FlagNames lists the flags shown to users, in display order.
var FlagNames = []struct {
	Name string
	Bit  uint64
}{{"CF", CF}, {"PF", PF}, {"AF", AF}, {"ZF", ZF}, {"SF", SF}, {"DF", DF}, {"OF", OF}}

// FormatFlags renders the user-visible flags, e.g. "CF=0 PF=1 ...".
func FormatFlags(f uint64) string {
	var b strings.Builder
	for i, n := range FlagNames {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(n.Name)
		if f&n.Bit != 0 {
			b.WriteString("=1")
		} else {
			b.WriteString("=0")
		}
	}
	return b.String()
}

func mask(size int) uint64 {
	if size >= 8 {
		return ^uint64(0)
	}
	return 1<<(8*uint(size)) - 1
}

func signBit(size int) uint64 { return 1 << (8*uint(size) - 1) }

func signExtend(v uint64, size int) int64 {
	switch size {
	case 1:
		return int64(int8(v))
	case 2:
		return int64(int16(v))
	case 4:
		return int64(int32(v))
	}
	return int64(v)
}

// szp computes SF, ZF and PF for a result of the given size (bytes).
func szp(res uint64, size int) uint64 {
	var f uint64
	res &= mask(size)
	if res == 0 {
		f |= ZF
	}
	if res&signBit(size) != 0 {
		f |= SF
	}
	if bits.OnesCount8(uint8(res))%2 == 0 {
		f |= PF
	}
	return f
}

func b2u(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

// Add computes a + b + carry at the given size and the resulting arithmetic flags.
func Add(a, b, carry uint64, size int) (res, flags uint64) {
	m := mask(size)
	a, b = a&m, b&m
	var cf bool
	if size == 8 {
		var c uint64
		res, c = bits.Add64(a, b, carry)
		cf = c != 0
	} else {
		sum := a + b + carry
		cf = sum > m
		res = sum & m
	}
	if cf {
		flags |= CF
	}
	if (a^res)&(b^res)&signBit(size) != 0 {
		flags |= OF
	}
	if (a^b^res)&0x10 != 0 {
		flags |= AF
	}
	return res, flags | szp(res, size)
}

// Sub computes a - b - borrow at the given size and the resulting arithmetic flags.
func Sub(a, b, borrow uint64, size int) (res, flags uint64) {
	m := mask(size)
	a, b = a&m, b&m
	d, br := bits.Sub64(a, b, borrow)
	res = d & m
	if br != 0 {
		flags |= CF
	}
	if (a^b)&(a^res)&signBit(size) != 0 {
		flags |= OF
	}
	if (a^b^res)&0x10 != 0 {
		flags |= AF
	}
	return res, flags | szp(res, size)
}

// Logic returns the flags for and/or/xor/test: CF=OF=0, SF/ZF/PF from result.
func Logic(res uint64, size int) uint64 { return szp(res, size) }

// setFlags replaces the flags selected by which with those in f.
func (c *CPU) setFlags(f, which uint64) {
	c.RFLAGS = c.RFLAGS&^which | f&which
}

func (c *CPU) flag(f uint64) bool { return c.RFLAGS&f != 0 }

// cond evaluates condition code cc (the 4-bit x86 encoding).
func (c *CPU) cond(cc int) bool {
	f := c.RFLAGS
	var r bool
	switch cc >> 1 {
	case 0: // O
		r = f&OF != 0
	case 1: // B
		r = f&CF != 0
	case 2: // E
		r = f&ZF != 0
	case 3: // BE
		r = f&(CF|ZF) != 0
	case 4: // S
		r = f&SF != 0
	case 5: // P
		r = f&PF != 0
	case 6: // L
		r = (f&SF != 0) != (f&OF != 0)
	case 7: // LE
		r = f&ZF != 0 || (f&SF != 0) != (f&OF != 0)
	}
	if cc&1 != 0 {
		return !r
	}
	return r
}
