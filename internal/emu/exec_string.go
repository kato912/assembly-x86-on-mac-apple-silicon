package emu

import "golang.org/x/arch/x86/x86asm"

// String instructions (movs, stos, lods, cmps, scas) with rep/repe/repne.

type strKind uint8

const (
	strMovs strKind = iota
	strStos
	strLods
	strCmps
	strScas
)

type strOp struct {
	kind strKind
	size int
}

var strOps = map[x86asm.Op]strOp{
	x86asm.MOVSB: {strMovs, 1}, x86asm.MOVSW: {strMovs, 2}, x86asm.MOVSD: {strMovs, 4}, x86asm.MOVSQ: {strMovs, 8},
	x86asm.STOSB: {strStos, 1}, x86asm.STOSW: {strStos, 2}, x86asm.STOSD: {strStos, 4}, x86asm.STOSQ: {strStos, 8},
	x86asm.LODSB: {strLods, 1}, x86asm.LODSW: {strLods, 2}, x86asm.LODSD: {strLods, 4}, x86asm.LODSQ: {strLods, 8},
	x86asm.CMPSB: {strCmps, 1}, x86asm.CMPSW: {strCmps, 2}, x86asm.CMPSD: {strCmps, 4}, x86asm.CMPSQ: {strCmps, 8},
	x86asm.SCASB: {strScas, 1}, x86asm.SCASW: {strScas, 2}, x86asm.SCASD: {strScas, 4}, x86asm.SCASQ: {strScas, 8},
}

func init() {
	for op := range strOps {
		register(execString, op)
	}
}

// repPrefix returns 0xF3 (rep/repe), 0xF2 (repne) or 0.
func repPrefix(in *x86asm.Inst) uint8 {
	var r uint8
	for _, p := range in.Prefix {
		if p == 0 {
			break
		}
		switch uint8(p) {
		case 0xF3, 0xF2:
			r = uint8(p)
		}
	}
	return r
}

func execString(c *CPU, in *x86asm.Inst) error {
	op := strOps[in.Op]
	rep := repPrefix(in)
	asz := in.AddrSize / 8
	am := mask(asz)
	siReg, diReg := x86asm.RSI, x86asm.RDI
	switch asz {
	case 4:
		siReg, diReg = x86asm.ESI, x86asm.EDI
	case 2:
		siReg, diReg = x86asm.SI, x86asm.DI
	}
	setIdx := func(r int, v uint64) {
		switch asz {
		case 8:
			c.R[r] = v
		case 4:
			c.R[r] = v & 0xffffffff
		default:
			c.R[r] = c.R[r]&^0xffff | v&0xffff
		}
	}
	delta := uint64(op.size)
	if c.flag(DF) {
		delta = -delta
	}
	acc := accum[op.size]
	readAt := func(r x86asm.Reg) (uint64, error) {
		c.faultMem = &x86asm.Mem{Base: r}
		return c.Mem.Read(c.R[regTable[r].idx]&am, op.size)
	}
	first := true
	for {
		if rep != 0 && c.R[RCX]&am == 0 {
			break
		}
		if !first {
			// Each extra iteration counts toward the instruction limit; when
			// it is hit, stop here and resume the rep on the next Step.
			if c.MaxSteps > 0 && c.Steps >= c.MaxSteps {
				c.RIP = c.pc
				return nil
			}
			c.Steps++
		}
		first = false
		si, di := c.R[RSI]&am, c.R[RDI]&am
		switch op.kind {
		case strMovs:
			v, err := readAt(siReg)
			if err != nil {
				return err
			}
			c.faultMem = &x86asm.Mem{Base: diReg}
			if err := c.Mem.Write(di, op.size, v); err != nil {
				return err
			}
			setIdx(RSI, si+delta)
			setIdx(RDI, di+delta)
		case strStos:
			c.faultMem = &x86asm.Mem{Base: diReg}
			if err := c.Mem.Write(di, op.size, c.getReg(acc)); err != nil {
				return err
			}
			setIdx(RDI, di+delta)
		case strLods:
			v, err := readAt(siReg)
			if err != nil {
				return err
			}
			c.setReg(acc, v)
			setIdx(RSI, si+delta)
		case strCmps:
			a, err := readAt(siReg)
			if err != nil {
				return err
			}
			b, err := readAt(diReg)
			if err != nil {
				return err
			}
			_, f := Sub(a, b, 0, op.size)
			c.setFlags(f, Arith)
			setIdx(RSI, si+delta)
			setIdx(RDI, di+delta)
		case strScas:
			b, err := readAt(diReg)
			if err != nil {
				return err
			}
			_, f := Sub(c.getReg(acc), b, 0, op.size)
			c.setFlags(f, Arith)
			setIdx(RDI, di+delta)
		}
		if rep == 0 {
			break
		}
		setIdx(RCX, (c.R[RCX]&am)-1)
		if op.kind == strCmps || op.kind == strScas {
			if (rep == 0xF3 && !c.flag(ZF)) || (rep == 0xF2 && c.flag(ZF)) {
				break
			}
		}
	}
	return nil
}
