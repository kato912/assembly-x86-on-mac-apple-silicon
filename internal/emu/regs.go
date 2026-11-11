package emu

import (
	"strings"

	"golang.org/x/arch/x86/x86asm"
)

// General-purpose register indexes (hardware encoding order).
const (
	RAX = iota
	RCX
	RDX
	RBX
	RSP
	RBP
	RSI
	RDI
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
)

var names64 = [16]string{"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15"}
var names32 = [16]string{"eax", "ecx", "edx", "ebx", "esp", "ebp", "esi", "edi", "r8d", "r9d", "r10d", "r11d", "r12d", "r13d", "r14d", "r15d"}

// RegName returns the conventional name of GPR i at the given width (32/64).
func RegName(i, bits int) string {
	if bits == 64 {
		return names64[i]
	}
	return names32[i]
}

// regLoc describes where an x86asm register lives inside the GPR file.
type regLoc struct {
	idx  int // GPR index
	size int // bytes: 1, 2, 4, 8
	high bool
	ok   bool
}

var regTable = func() (t [256]regLoc) {
	for r := x86asm.AL; r <= x86asm.BL; r++ {
		t[r] = regLoc{int(r - x86asm.AL), 1, false, true}
	}
	for r := x86asm.AH; r <= x86asm.BH; r++ {
		t[r] = regLoc{int(r - x86asm.AH), 1, true, true}
	}
	for r := x86asm.SPB; r <= x86asm.R15B; r++ {
		t[r] = regLoc{int(r-x86asm.SPB) + 4, 1, false, true}
	}
	for r := x86asm.AX; r <= x86asm.R15W; r++ {
		t[r] = regLoc{int(r - x86asm.AX), 2, false, true}
	}
	for r := x86asm.EAX; r <= x86asm.R15L; r++ {
		t[r] = regLoc{int(r - x86asm.EAX), 4, false, true}
	}
	for r := x86asm.RAX; r <= x86asm.R15; r++ {
		t[r] = regLoc{int(r - x86asm.RAX), 8, false, true}
	}
	return
}()

// regSize returns the width in bytes of a general-purpose register, or 0.
func regSize(r x86asm.Reg) int { return regTable[r].size }

// isGPR reports whether r is a general-purpose register.
func isGPR(r x86asm.Reg) bool { return regTable[r].ok }

func (c *CPU) getReg(r x86asm.Reg) uint64 {
	l := regTable[r]
	if !l.ok {
		switch r {
		case x86asm.RIP, x86asm.EIP, x86asm.IP:
			return c.next
		}
		return 0
	}
	v := c.R[l.idx]
	switch l.size {
	case 1:
		if l.high {
			return (v >> 8) & 0xff
		}
		return v & 0xff
	case 2:
		return v & 0xffff
	case 4:
		return v & 0xffffffff
	}
	return v
}

func (c *CPU) setReg(r x86asm.Reg, v uint64) {
	l := regTable[r]
	if !l.ok {
		return
	}
	old := c.R[l.idx]
	switch l.size {
	case 1:
		if l.high {
			c.R[l.idx] = old&^0xff00 | (v&0xff)<<8
		} else {
			c.R[l.idx] = old&^0xff | v&0xff
		}
	case 2:
		c.R[l.idx] = old&^0xffff | v&0xffff
	case 4:
		c.R[l.idx] = v & 0xffffffff // 32-bit writes zero-extend into the full register
	default:
		c.R[l.idx] = v
	}
}

// RegByName looks up a register by its NASM name (eax, r8d, al, ...),
// returning its x86asm form. Used by the debugger.
func RegByName(name string) (x86asm.Reg, bool) {
	name = strings.ToLower(name)
	for r := x86asm.AL; r <= x86asm.R15; r++ {
		if strings.ToLower(nasmRegName(r)) == name {
			return r, true
		}
	}
	return 0, false
}

// nasmRegName returns NASM's spelling of a register.
func nasmRegName(r x86asm.Reg) string {
	switch {
	case r >= x86asm.SPB && r <= x86asm.DIB:
		return [...]string{"spl", "bpl", "sil", "dil"}[r-x86asm.SPB]
	case r >= x86asm.R8L && r <= x86asm.R15L:
		return strings.ToLower(r.String()[:len(r.String())-1]) + "d"
	}
	return strings.ToLower(r.String())
}

// Value returns the current value of register r (for the debugger).
func (c *CPU) Value(r x86asm.Reg) uint64 { return c.getReg(r) }

// SetValue writes register r (for the debugger) with the same rules as an
// instruction would: writing al keeps the rest of rax, writing eax clears
// the upper half of rax in 64-bit mode.
func (c *CPU) SetValue(r x86asm.Reg, v uint64) { c.setReg(r, v) }
