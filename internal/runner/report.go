package runner

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/arch/x86/x86asm"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
)

// Locate maps an address to the input file (index into Files) and the
// source line that produced it.
func (p *Program) Locate(addr uint64) (file int, loc nasm.Loc, ok bool) {
	if p.Image == nil {
		return 0, nasm.Loc{}, false
	}
	sec := p.Image.SectionAt(addr)
	if sec == nil || sec.Obj >= len(p.Files) {
		return 0, nasm.Loc{}, false
	}
	loc, ok = p.Files[sec.Obj].Lines.Lookup(sec.Name, addr-sec.Addr)
	return sec.Obj, loc, ok
}

// LocAt maps an address to the source line that produced it.
func (p *Program) LocAt(addr uint64) (nasm.Loc, bool) {
	_, loc, ok := p.Locate(addr)
	return loc, ok
}

// Where describes the source location of the instruction at addr, e.g.
// "line 12: mov eax, [ebx]" (prefixed with the file name when the program
// has several files or the line comes from an %include).
func (p *Program) Where(addr uint64) string {
	if addr == p.ExitStub && p.ExitStub != 0 {
		return i18n.T("loc.stub")
	}
	if file, loc, ok := p.Locate(addr); ok {
		s := i18n.T("loc.line", loc.Line, CodeText(loc.Text))
		switch {
		case loc.File != "":
			s = loc.File + " " + s
		case len(p.Files) > 1:
			s = filepath.Base(p.Files[file].Path) + " " + s
		}
		if loc.Macro != "" {
			s += " " + i18n.T("loc.macro", CodeText(loc.Macro))
		}
		return s
	}
	s := p.Addr(addr)
	if in, err := p.CPU.Decode(addr); err == nil {
		s += ": " + p.Disasm(in, addr)
	}
	return s
}

// FileIndex finds an input file by its name or path ("util.asm").
func (p *Program) FileIndex(name string) (int, bool) {
	for i, f := range p.Files {
		if f.Path == name || filepath.Base(f.Path) == name || filepath.Base(f.Path) == filepath.Base(name) {
			return i, true
		}
	}
	return 0, false
}

// LineAddr returns the address of the first instruction generated for
// line (or, if that line has no code, the next line that does) in input
// file file, plus that line's number and text.
func (p *Program) LineAddr(file, line int) (addr uint64, at int, text string, ok bool) {
	if file < 0 || file >= len(p.Files) || p.Files[file].Lines == nil {
		return 0, 0, "", false
	}
	lm := p.Files[file].Lines
	for _, l := range lm.Lines() {
		if l < line {
			continue
		}
		for _, so := range lm.Places(l) {
			if sec := p.Image.SectionOf(file, so.Section); sec != nil && sec.Exec {
				return sec.Addr + so.Off, l, CodeText(lm.Text(l)), true
			}
		}
	}
	return 0, 0, "", false
}

// CodeText strips a trailing ; comment and collapses runs of spaces/tabs
// (outside string literals) so source lines display compactly.
func CodeText(s string) string {
	var b strings.Builder
	var q byte
	space := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if q == 0 {
			if c == ';' {
				break
			}
			if c == ' ' || c == '\t' {
				space = true
				continue
			}
			if c == '"' || c == '\'' || c == '`' {
				q = c
			}
		} else if c == q {
			q = 0
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteByte(c)
	}
	return b.String()
}

// Addr formats an address at the program's natural width.
func (p *Program) Addr(a uint64) string {
	if p.Bits == 32 || a <= 0xffffffff {
		return fmt.Sprintf("0x%08x", a)
	}
	return fmt.Sprintf("0x%x", a)
}

// Hex formats a value of size bytes with all its digits.
func Hex(v uint64, size int) string {
	return fmt.Sprintf("0x%0*x", 2*size, v)
}

// Disasm renders an instruction, naming targets after labels.
func (p *Program) Disasm(in *x86asm.Inst, pc uint64) string {
	return emu.Disasm(in, pc, func(a uint64) (string, uint64) {
		if s, ok := p.Image.SymbolFor(a); ok && s.Addr == a {
			return s.Name, s.Addr
		}
		return "", 0
	})
}

// Describe turns a fault into a friendly, translated message: the source
// line, what went wrong, and a hint.
func (p *Program) Describe(f *emu.Fault) string {
	at := f.PC
	switch f.Kind {
	case emu.FaultExec, emu.FaultRunOff, emu.FaultBadReturn:
		at = f.From
	}
	var what string
	var hints []string
	w := p.CPU.Width()
	switch f.Kind {
	case emu.FaultMemory:
		key := "err.read"
		if f.Mem.Access == emu.AccessWrite {
			key = "err.write"
		}
		what = i18n.T(key, p.Addr(f.Mem.Addr)) + p.regList(f.Regs)
		sec := p.Image.SectionAt(f.Mem.Addr)
		switch {
		case f.Mem.Addr < 0x1000:
			hints = append(hints, i18n.T("hint.null"))
		case f.Mem.Access == emu.AccessWrite && sec != nil && !sec.Write:
			hints = append(hints, i18n.T("hint.readonly", sec.Name))
		case f.Mem.Region == nil && f.Mem.Addr < 0x10000:
			hints = append(hints, i18n.T("hint.smallAddr"))
		default:
			hints = append(hints, i18n.T("hint.unmapped"))
		}
	case emu.FaultStackOverflow:
		what = i18n.T("err.stackOverflow") + p.regList([]emu.RegValue{{Name: emu.RegName(emu.RSP, p.Bits), Value: p.CPU.SP(), Size: w}})
		hints = append(hints, i18n.T("hint.stackOverflow"))
	case emu.FaultExec:
		what = i18n.T("err.exec", p.Addr(f.PC))
		if f.PC == 0 {
			hints = append(hints, i18n.T("hint.execZero"))
		} else {
			hints = append(hints, i18n.T("hint.exec"))
		}
	case emu.FaultRunOff:
		what = i18n.T("err.runoff")
		hints = append(hints, i18n.T("hint.exit"+fmt.Sprint(p.Bits)))
	case emu.FaultBadReturn:
		if p.Image.EntryName != "main" && p.CPU.SP() == p.InitialSP+uint64(w) {
			what = i18n.T("err.retStart")
			hints = append(hints, i18n.T("hint.exit"+fmt.Sprint(p.Bits)))
		} else {
			what = i18n.T("err.badRet", p.Addr(f.PC))
			hints = append(hints, i18n.T("hint.badRet"))
		}
	case emu.FaultDivZero:
		what = i18n.T("err.divZero", p.operandText(f))
	case emu.FaultDivOverflow:
		what, hints = p.divOverflow(f)
	case emu.FaultUnsupported:
		what = i18n.T("err.unsupported", p.instText(f))
		if isFloatOrSIMD(f.Inst) {
			hints = append(hints, i18n.T("hint.fpu"))
		}
	case emu.FaultPrivileged:
		what = i18n.T("err.privileged", p.instText(f))
	case emu.FaultInvalid:
		what = i18n.T("err.invalid", fmt.Sprintf("% x", f.Bytes[:min(len(f.Bytes), 4)]))
		hints = append(hints, i18n.T("hint.invalid"))
	case emu.FaultSyscallIn32:
		what = i18n.T("err.syscall32")
	case emu.FaultBadInterrupt:
		what = i18n.T("err.badInt", f.Int)
		switch f.Int {
		case 0x21, 0x10, 0x16, 0x20:
			hints = append(hints, i18n.T("hint.dos"))
		default:
			hints = append(hints, i18n.T("hint.int"))
		}
	case emu.FaultBreakpoint:
		what = i18n.T("err.int3")
		hints = append(hints, i18n.T("hint.int3"))
	case emu.FaultMaxSteps:
		what = i18n.T("err.maxSteps", f.Steps)
		hints = append(hints, i18n.T("hint.maxSteps"))
	default:
		what = f.Error()
	}
	var b strings.Builder
	b.WriteString(p.Where(at))
	b.WriteString("  →  ")
	b.WriteString(what)
	for _, h := range hints {
		b.WriteString("\n    ")
		b.WriteString(i18n.T("hint.prefix"))
		b.WriteString(h)
	}
	return b.String()
}

func (p *Program) regList(regs []emu.RegValue) string {
	if len(regs) == 0 {
		return ""
	}
	parts := make([]string, len(regs))
	for i, r := range regs {
		parts[i] = r.Name + " = " + Hex(r.Value, r.Size)
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

func (p *Program) instText(f *emu.Fault) string {
	if f.Inst == nil {
		return "?"
	}
	return p.Disasm(f.Inst, f.PC)
}

// operandText names the divisor operand, e.g. "ebx" or "dword [x]".
func (p *Program) operandText(f *emu.Fault) string {
	if f.Inst == nil || f.Inst.Args[0] == nil {
		return "?"
	}
	if r, ok := f.Inst.Args[0].(x86asm.Reg); ok {
		return strings.ToLower(r.String())
	}
	s := p.instText(f)
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return s[i+1:]
	}
	return s
}

func (p *Program) divOverflow(f *emu.Fault) (string, []string) {
	var dividend, dest, fix string
	d := f.Div
	switch d.Size {
	case 1:
		dividend = "ax = " + Hex(d.Hi<<8|d.Lo, 2)
		dest, fix = "al", "movzx ax, al / cbw"
	case 2:
		dividend = "dx:ax = " + Hex(d.Hi, 2) + ":" + Hex(d.Lo, 2)
		dest, fix = "ax", "xor dx, dx / cwd"
	case 4:
		dividend = "edx:eax = " + Hex(d.Hi, 4) + ":" + Hex(d.Lo, 4)
		dest, fix = "eax", "xor edx, edx / cdq"
	default:
		dividend = "rdx:rax = " + Hex(d.Hi, 8) + ":" + Hex(d.Lo, 8)
		dest, fix = "rax", "xor rdx, rdx / cqo"
	}
	what := i18n.T("err.divOverflow", dividend, d.Divisor, dest)
	parts := strings.Split(fix, " / ")
	hint := i18n.T("hint.divUnsigned", parts[0])
	if d.Signed {
		hint = i18n.T("hint.divSigned", parts[1])
	}
	return what, []string{hint}
}

func isFloatOrSIMD(in *x86asm.Inst) bool {
	if in == nil {
		return false
	}
	name := in.Op.String()
	if strings.HasPrefix(name, "F") {
		return true
	}
	for _, a := range in.Args {
		if r, ok := a.(x86asm.Reg); ok && (r >= x86asm.F0 && r <= x86asm.X15) {
			return true
		}
	}
	return strings.Contains(name, "XMM") || strings.HasSuffix(name, "PS") || strings.HasSuffix(name, "PD") ||
		strings.HasSuffix(name, "SS") || strings.HasSuffix(name, "SD") || strings.HasPrefix(name, "V")
}
