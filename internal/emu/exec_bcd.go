package emu

import "golang.org/x/arch/x86/x86asm"

// Binary-coded decimal adjustments (32-bit mode only; they are invalid in
// 64-bit mode on real CPUs, and the decoder rejects them there too).
// The semantics follow the pseudocode in the Intel SDM. Flags the manual
// calls undefined are computed from the result like SF/ZF/PF.

func init() {
	register(execDaaDas, x86asm.DAA, x86asm.DAS)
	register(execAaaAas, x86asm.AAA, x86asm.AAS)
	register(execAam, x86asm.AAM)
	register(execAad, x86asm.AAD)
}

// Daa returns AL and the flags after DAA (sub=false) or DAS (sub=true).
func Daa(al uint64, flags uint64, sub bool) (uint64, uint64) {
	al &= 0xff
	oldAL, oldCF := al, flags&CF != 0
	af := flags&AF != 0
	cf := false
	if al&0x0f > 9 || af {
		if sub {
			cf = oldCF || al < 6
			al = (al - 6) & 0xff
		} else {
			cf = oldCF || al+6 > 0xff
			al = (al + 6) & 0xff
		}
		af = true
	} else {
		af = false
	}
	if oldAL > 0x99 || oldCF {
		if sub {
			al = (al - 0x60) & 0xff
		} else {
			al = (al + 0x60) & 0xff
		}
		cf = true
	} else if !sub {
		cf = false // DAA clears CF here; DAS keeps the value from the low digit
	}
	f := szp(al, 1)
	if cf {
		f |= CF
	}
	if af {
		f |= AF
	}
	return al, f
}

func execDaaDas(c *CPU, in *x86asm.Inst) error {
	al, f := Daa(c.getReg(x86asm.AL), c.RFLAGS, in.Op == x86asm.DAS)
	c.setReg(x86asm.AL, al)
	c.setFlags(f, Arith)
	return nil
}

// Aaa returns AX and the flags after AAA (sub=false) or AAS (sub=true).
func Aaa(ax uint64, flags uint64, sub bool) (uint64, uint64) {
	ax &= 0xffff
	var f uint64
	if ax&0x0f > 9 || flags&AF != 0 {
		if sub {
			ax = (ax - 6) & 0xffff
			ax = (ax - 0x100) & 0xffff
		} else {
			ax = (ax + 0x106) & 0xffff
		}
		f = AF | CF
	}
	ax &^= 0xf0 // AL <- AL AND 0FH
	return ax, f | szp(ax&0xff, 1)
}

func execAaaAas(c *CPU, in *x86asm.Inst) error {
	ax, f := Aaa(c.getReg(x86asm.AX), c.RFLAGS, in.Op == x86asm.AAS)
	c.setReg(x86asm.AX, ax)
	c.setFlags(f, Arith)
	return nil
}

func bcdBase(in *x86asm.Inst) uint64 {
	if imm, ok := in.Args[0].(x86asm.Imm); ok {
		return uint64(imm) & 0xff
	}
	return 10
}

func execAam(c *CPU, in *x86asm.Inst) error {
	base := bcdBase(in)
	al := c.getReg(x86asm.AL)
	if base == 0 {
		f := &Fault{Kind: FaultDivZero}
		f.Div.Lo, f.Div.Size = al, 1
		return f
	}
	ah, al := al/base, al%base
	c.setReg(x86asm.AX, ah<<8|al)
	c.setFlags(szp(al, 1), Arith)
	return nil
}

func execAad(c *CPU, in *x86asm.Inst) error {
	base := bcdBase(in)
	ax := c.getReg(x86asm.AX)
	al := (ax&0xff + (ax>>8)*base) & 0xff
	c.setReg(x86asm.AX, al)
	c.setFlags(szp(al, 1), Arith)
	return nil
}
