package emu

import "golang.org/x/arch/x86/x86asm"

// Shifts and rotates. The semantic core is in pure functions (Shift,
// DoubleShift) so the flag rules can be unit tested directly.

// ShiftOp selects a shift or rotate.
type ShiftOp int

const (
	SHL ShiftOp = iota // also SAL
	SHR
	SAR
	ROL
	ROR
	RCL
	RCR
)

func init() {
	register(execShift, x86asm.SHL, x86asm.SHR, x86asm.SAR, x86asm.ROL, x86asm.ROR, x86asm.RCL, x86asm.RCR)
	register(execDoubleShift, x86asm.SHLD, x86asm.SHRD)
}

var shiftOps = map[x86asm.Op]ShiftOp{x86asm.SHL: SHL, x86asm.SHR: SHR, x86asm.SAR: SAR,
	x86asm.ROL: ROL, x86asm.ROR: ROR, x86asm.RCL: RCL, x86asm.RCR: RCR}

func countMask(size int) uint64 {
	if size == 8 {
		return 0x3f
	}
	return 0x1f
}

func msb(v uint64, size int) uint64 { return (v >> (8*uint(size) - 1)) & 1 }

// Shift applies op to a (size bytes) by count, with carry-in cf for
// rcl/rcr. It returns the result, the new flag values and the set of flags
// the instruction writes. When the masked count is zero no flags change.
//
// OF is architecturally defined only for 1-bit shifts; for larger counts
// nasmx86 applies the 1-bit formula. AF is undefined and set to 0.
func Shift(op ShiftOp, a, count uint64, size int, cf bool) (res, flags, affected uint64) {
	m := mask(size)
	a &= m
	n := uint64(8 * size)
	cnt := count & countMask(size)
	if cnt == 0 {
		return a, 0, 0
	}
	var c uint64 // carry out
	var of uint64
	switch op {
	case SHL:
		if cnt < n {
			res = (a << cnt) & m
			c = (a >> (n - cnt)) & 1
		} else if cnt == n {
			c = a & 1
		}
		of = msb(res, size) ^ c
	case SHR:
		if cnt < n {
			res = a >> cnt
			c = (a >> (cnt - 1)) & 1
		} else if cnt == n {
			c = msb(a, size)
		}
		of = msb(a, size)
	case SAR:
		sa := signExtend(a, size)
		s := cnt
		if s > 63 {
			s = 63
		}
		res = uint64(sa>>s) & m
		s1 := cnt - 1
		if s1 > 63 {
			s1 = 63
		}
		c = uint64(sa>>s1) & 1
		of = 0
	case ROL, ROR:
		r := cnt % n
		res = a
		if r != 0 {
			if op == ROL {
				res = (a<<r | a>>(n-r)) & m
			} else {
				res = (a>>r | a<<(n-r)) & m
			}
		}
		if op == ROL {
			c = res & 1
			of = msb(res, size) ^ c
		} else {
			c = msb(res, size)
			of = c ^ ((res >> (n - 2)) & 1)
		}
		return res, c*CF | of*OF, CF | OF
	case RCL, RCR:
		t := cnt
		if size == 1 {
			t %= 9
		} else if size == 2 {
			t %= 17
		}
		c = b2u(cf)
		res = a
		if op == RCR {
			of = msb(a, size) ^ c
		}
		for i := uint64(0); i < t; i++ {
			if op == RCL {
				out := msb(res, size)
				res = (res<<1 | c) & m
				c = out
			} else {
				out := res & 1
				res = res>>1 | c<<(n-1)
				c = out
			}
		}
		if op == RCL {
			of = msb(res, size) ^ c
		}
		return res, c*CF | of*OF, CF | OF
	}
	return res, szp(res, size) | c*CF | of*OF, Arith
}

// DoubleShift implements shld (left=true) and shrd: dst is shifted by
// count and the vacated bits are filled from src.
func DoubleShift(left bool, dst, src, count uint64, size int) (res, flags, affected uint64) {
	m := mask(size)
	dst, src = dst&m, src&m
	n := uint64(8 * size)
	cnt := count & countMask(size)
	if cnt == 0 {
		return dst, 0, 0
	}
	var c uint64
	if cnt > n { // only possible for 16-bit operands: result is undefined
		cnt = n
	}
	if left {
		res = (dst<<cnt | src>>(n-cnt)) & m
		c = (dst >> (n - cnt)) & 1
	} else {
		res = (dst>>cnt | src<<(n-cnt)) & m
		c = (dst >> (cnt - 1)) & 1
	}
	of := msb(res, size) ^ msb(dst, size)
	return res, szp(res, size) | c*CF | of*OF, Arith
}

func execShift(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	cnt, err := c.read(in, in.Args[1], 1)
	if err != nil {
		return err
	}
	res, f, which := Shift(shiftOps[in.Op], a, cnt, size, c.flag(CF))
	// Even a zero count writes the destination (a 32-bit register is
	// zero-extended in 64-bit mode); only the flags are left untouched.
	if err := c.write(in, in.Args[0], size, res); err != nil {
		return err
	}
	c.setFlags(f, which)
	return nil
}

func execDoubleShift(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	b, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	cnt, err := c.read(in, in.Args[2], 1)
	if err != nil {
		return err
	}
	res, f, which := DoubleShift(in.Op == x86asm.SHLD, a, b, cnt, size)
	// Even a zero count writes the destination (a 32-bit register is
	// zero-extended in 64-bit mode); only the flags are left untouched.
	if err := c.write(in, in.Args[0], size, res); err != nil {
		return err
	}
	c.setFlags(f, which)
	return nil
}
