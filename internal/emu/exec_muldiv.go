package emu

import (
	"math/big"
	"math/bits"

	"golang.org/x/arch/x86/x86asm"
)

// neg, not, and the multiply/divide family.

func init() {
	register(execNeg, x86asm.NEG)
	register(execNot, x86asm.NOT)
	register(execMul, x86asm.MUL)
	register(execImul, x86asm.IMUL)
	register(execDiv, x86asm.DIV, x86asm.IDIV)
}

func execNeg(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	res, f := Sub(0, a, 0, size)
	if err := c.write(in, in.Args[0], size, res); err != nil {
		return err
	}
	c.setFlags(f, Arith)
	return nil
}

func execNot(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	return c.write(in, in.Args[0], size, ^a)
}

// halves returns the (low, high) register pair used by mul/div.
func halves(size int) (lo, hi x86asm.Reg) {
	switch size {
	case 1:
		return x86asm.AL, x86asm.AH
	case 2:
		return x86asm.AX, x86asm.DX
	case 4:
		return x86asm.EAX, x86asm.EDX
	}
	return x86asm.RAX, x86asm.RDX
}

// MulFlags returns the flags after mul/imul: CF=OF=overflow. SF/ZF/PF are
// architecturally undefined; nasmx86 sets them from the low half.
func MulFlags(lo uint64, size int, overflow bool) uint64 {
	f := szp(lo, size)
	if overflow {
		f |= CF | OF
	}
	return f
}

func execMul(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	b, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	loR, hiR := halves(size)
	a := c.getReg(loR)
	var lo, hi uint64
	if size == 8 {
		hi, lo = bits.Mul64(a, b)
	} else {
		p := a * b
		lo = p & mask(size)
		hi = p >> (8 * uint(size))
	}
	if size == 1 {
		c.setReg(x86asm.AX, hi<<8|lo)
	} else {
		c.setReg(loR, lo)
		c.setReg(hiR, hi)
	}
	c.setFlags(MulFlags(lo, size, hi != 0), Arith)
	return nil
}

// mul128 returns the full signed 128-bit product of two int64 values.
func mul128(a, b int64) (hi, lo uint64) {
	hi, lo = bits.Mul64(uint64(a), uint64(b))
	if a < 0 {
		hi -= uint64(b)
	}
	if b < 0 {
		hi -= uint64(a)
	}
	return hi, lo
}

// Imul computes the truncated signed product and whether it overflowed.
func Imul(a, b uint64, size int) (res uint64, overflow bool) {
	sa, sb := signExtend(a, size), signExtend(b, size)
	if size == 8 {
		hi, lo := mul128(sa, sb)
		return lo, hi != uint64(int64(lo)>>63)
	}
	p := sa * sb
	res = uint64(p) & mask(size)
	return res, p != signExtend(res, size)
}

func execImul(c *CPU, in *x86asm.Inst) error {
	if in.Args[1] == nil { // one operand: rDX:rAX = rAX * src
		size := argSize(in, in.Args[0])
		b, err := c.read(in, in.Args[0], size)
		if err != nil {
			return err
		}
		loR, hiR := halves(size)
		a := c.getReg(loR)
		var lo, hi uint64
		var overflow bool
		if size == 8 {
			hi, lo = mul128(int64(a), int64(b))
			overflow = hi != uint64(int64(lo)>>63)
		} else {
			p := signExtend(a, size) * signExtend(b, size)
			lo = uint64(p) & mask(size)
			hi = uint64(p>>(8*uint(size))) & mask(size)
			overflow = p != signExtend(lo, size)
		}
		if size == 1 {
			c.setReg(x86asm.AX, hi<<8|lo)
		} else {
			c.setReg(loR, lo)
			c.setReg(hiR, hi)
		}
		c.setFlags(MulFlags(lo, size, overflow), Arith)
		return nil
	}
	size := argSize(in, in.Args[0])
	src1, src2 := in.Args[0], in.Args[1]
	if in.Args[2] != nil {
		src1, src2 = in.Args[1], in.Args[2]
	}
	a, err := c.read(in, src1, size)
	if err != nil {
		return err
	}
	b, err := c.read(in, src2, size)
	if err != nil {
		return err
	}
	res, overflow := Imul(a, b, size)
	if err := c.write(in, in.Args[0], size, res); err != nil {
		return err
	}
	c.setFlags(MulFlags(res, size, overflow), Arith)
	return nil
}

func execDiv(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	d, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	loR, hiR := halves(size)
	var lo, hi uint64
	if size == 1 {
		ax := c.getReg(x86asm.AX)
		lo, hi = ax&0xff, ax>>8
	} else {
		lo, hi = c.getReg(loR), c.getReg(hiR)
	}
	signed := in.Op == x86asm.IDIV
	fault := func(k FaultKind) error {
		f := &Fault{Kind: k}
		f.Div.Hi, f.Div.Lo, f.Div.Divisor, f.Div.Size, f.Div.Signed = hi, lo, d, size, signed
		return f
	}
	if d == 0 {
		return fault(FaultDivZero)
	}
	q, r, ok := Divide(hi, lo, d, size, signed)
	if !ok {
		return fault(FaultDivOverflow)
	}
	if size == 1 {
		c.setReg(x86asm.AX, r<<8|q)
	} else {
		c.setReg(loR, q)
		c.setReg(hiR, r)
	}
	return nil
}

// Divide divides the double-width value hi:lo by d (all size bytes wide).
// ok is false when the quotient does not fit in size bytes (#DE).
func Divide(hi, lo, d uint64, size int, signed bool) (q, r uint64, ok bool) {
	m := mask(size)
	if !signed {
		if size == 8 {
			if hi >= d {
				return 0, 0, false
			}
			q, r = bits.Div64(hi, lo, d)
			return q, r, true
		}
		n := hi<<(8*uint(size)) | lo
		q, r = n/d, n%d
		return q, r, q <= m
	}
	if size == 8 {
		n := new(big.Int).SetUint64(hi)
		n.Lsh(n, 64).Or(n, new(big.Int).SetUint64(lo))
		if hi&(1<<63) != 0 { // negative dividend: subtract 2^128
			n.Sub(n, new(big.Int).Lsh(big.NewInt(1), 128))
		}
		dv := big.NewInt(int64(d))
		bq, br := new(big.Int).QuoRem(n, dv, new(big.Int))
		if !bq.IsInt64() {
			return 0, 0, false
		}
		return uint64(bq.Int64()), uint64(br.Int64()), true
	}
	n := signExtend(hi<<(8*uint(size))|lo, 2*size)
	if size == 4 {
		n = int64(hi<<32 | lo)
	}
	dv := signExtend(d, size)
	sq, sr := n/dv, n%dv
	if sq != signExtend(uint64(sq)&m, size) {
		return 0, 0, false
	}
	return uint64(sq) & m, uint64(sr) & m, true
}
