package runner

import (
	"fmt"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

// regOrder is the display order of general-purpose registers.
var regOrder = []int{emu.RAX, emu.RBX, emu.RCX, emu.RDX, emu.RSI, emu.RDI, emu.RBP, emu.RSP,
	emu.R8, emu.R9, emu.R10, emu.R11, emu.R12, emu.R13, emu.R14, emu.R15}

// ValueText renders a register value: hex at full width plus either the
// label it points to or its signed decimal value.
func (p *Program) ValueText(i int, v uint64) string {
	w := p.CPU.Width()
	s := Hex(v, w)
	if i == emu.RSP || i == emu.RBP {
		return s
	}
	if v != 0 {
		if sym, ok := p.Image.SymbolFor(v); ok && sym.Addr == v {
			return s + " (" + sym.Name + ")"
		}
	}
	if w == 4 {
		return fmt.Sprintf("%s (%d)", s, int32(v))
	}
	return fmt.Sprintf("%s (%d)", s, int64(v))
}

// FormatRegs renders all registers and flags, two registers per line.
func (p *Program) FormatRegs() string {
	c := p.CPU
	n := 8
	if p.Bits == 64 {
		n = 16
	}
	var b strings.Builder
	col := 30
	if p.Bits == 64 {
		col = 44
	}
	for k := 0; k < n; k += 2 {
		for j := 0; j < 2; j++ {
			i := regOrder[k+j]
			cell := fmt.Sprintf("%-4s= %s", emu.RegName(i, p.Bits), p.ValueText(i, c.R[i]))
			if j == 0 {
				fmt.Fprintf(&b, "%-*s", col, cell)
			} else {
				b.WriteString(cell)
			}
		}
		b.WriteByte('\n')
	}
	ip := "eip"
	if p.Bits == 64 {
		ip = "rip"
	}
	fmt.Fprintf(&b, "%-4s= %s\n", ip, Hex(c.RIP, c.Width()))
	fmt.Fprintf(&b, "flags: %s\n", emu.FormatFlags(c.RFLAGS))
	return b.String()
}
