package emu

import (
	"math/bits"

	"golang.org/x/arch/x86/x86asm"
)

// Bit test/scan instructions, byte swap and the conditional set/move family.

func init() {
	register(execBt, x86asm.BT, x86asm.BTS, x86asm.BTR, x86asm.BTC)
	register(execBitScan, x86asm.BSF, x86asm.BSR)
	register(execBswap, x86asm.BSWAP)
	var setcc, cmovcc []x86asm.Op
	for op := range ccOf {
		switch op.String()[0] {
		case 'S':
			setcc = append(setcc, op)
		case 'C':
			cmovcc = append(cmovcc, op)
		}
	}
	register(execSetcc, setcc...)
	register(execCmovcc, cmovcc...)
}

func applyBit(op x86asm.Op, v, bit uint64) uint64 {
	switch op {
	case x86asm.BTS:
		return v | 1<<bit
	case x86asm.BTR:
		return v &^ (1 << bit)
	case x86asm.BTC:
		return v ^ 1<<bit
	}
	return v
}

func execBt(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	off, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	if m, ok := in.Args[0].(x86asm.Mem); ok {
		if _, isReg := in.Args[1].(x86asm.Reg); isReg {
			// A register bit offset is signed and may reach outside the
			// operand: address the containing byte directly.
			soff := signExtend(off, size)
			addr := c.wrap(c.ea(m, in) + uint64(soff>>3))
			bit := uint64(soff & 7)
			c.faultMem = &m
			v, err := c.Mem.Read(addr, 1)
			if err != nil {
				return err
			}
			if in.Op != x86asm.BT {
				if err := c.Mem.Write(addr, 1, applyBit(in.Op, v, bit)); err != nil {
					return err
				}
			}
			c.setFlags((v>>bit&1)*CF, CF)
			return nil
		}
	}
	v, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	bit := off & uint64(8*size-1)
	if in.Op != x86asm.BT {
		if err := c.write(in, in.Args[0], size, applyBit(in.Op, v, bit)); err != nil {
			return err
		}
	}
	c.setFlags((v>>bit&1)*CF, CF)
	return nil
}

func execBitScan(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	v, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	if v == 0 {
		c.setFlags(ZF, ZF) // destination is left unchanged
		return nil
	}
	var idx int
	if in.Op == x86asm.BSF {
		idx = bits.TrailingZeros64(v)
	} else {
		idx = 63 - bits.LeadingZeros64(v)
	}
	c.setFlags(0, ZF)
	return c.write(in, in.Args[0], size, uint64(idx))
}

func execBswap(c *CPU, in *x86asm.Inst) error {
	r := in.Args[0].(x86asm.Reg)
	v := c.getReg(r)
	switch regSize(r) {
	case 8:
		v = bits.ReverseBytes64(v)
	case 4:
		v = uint64(bits.ReverseBytes32(uint32(v)))
	default:
		v = 0 // bswap of a 16-bit register is undefined; hardware clears it
	}
	c.setReg(r, v)
	return nil
}

func execSetcc(c *CPU, in *x86asm.Inst) error {
	return c.write(in, in.Args[0], 1, b2u(c.cond(ccOf[in.Op])))
}

func execCmovcc(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	v, err := c.read(in, in.Args[1], size) // the source is read even if the condition is false
	if err != nil {
		return err
	}
	if !c.cond(ccOf[in.Op]) {
		v = c.getReg(in.Args[0].(x86asm.Reg)) // a 32-bit destination is still zero-extended
	}
	return c.write(in, in.Args[0], size, v)
}
