package debugger

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// valueSize guesses how wide the value of an expression is, for display:
// a lone register has its own size, "dword [x]" is 4, otherwise the CPU width.
func (d *stepper) valueSize(expr string) int {
	e := strings.TrimSpace(expr)
	if r, ok := regByName(e); ok {
		return regSizeOf(r)
	}
	if i := strings.IndexByte(e, ' '); i > 0 {
		if n, ok := sizeWords[strings.ToLower(e[:i])]; ok && strings.HasPrefix(strings.TrimSpace(e[i:]), "[") {
			return n
		}
	}
	if n, ok := sizeWords[strings.ToLower(strings.SplitN(e, "[", 2)[0])]; ok {
		return n
	}
	return d.c.Width()
}

// formatValue renders v (size bytes) as hex, unsigned and signed decimal,
// and a character when it is a printable byte.
func formatValue(v uint64, size int) string {
	if size < 8 {
		v &= 1<<(8*uint(size)) - 1
	}
	s := fmt.Sprintf("%s = %d", runner.Hex(v, size), v)
	signed := int64(v)
	switch size {
	case 1:
		signed = int64(int8(v))
	case 2:
		signed = int64(int16(v))
	case 4:
		signed = int64(int32(v))
	}
	if signed < 0 {
		s += fmt.Sprintf(" (signed %d)", signed)
	}
	if size == 1 && v >= 32 && v < 127 {
		s += fmt.Sprintf(" '%c'", rune(v))
	}
	return s
}

// printCmd implements "p expr" and "p/s expr" (string at an address).
func (d *stepper) printCmd(cmd string, args []string) {
	expr := strings.TrimSpace(strings.Join(args, " "))
	if expr == "" {
		fmt.Fprintln(d.out, i18n.T("dbg.pUsage"))
		return
	}
	v, err := d.eval(expr)
	if err != nil {
		fmt.Fprintln(d.out, err)
		return
	}
	if cmd == "p/s" {
		b := d.c.Mem.Peek(v, 256)
		n := 0
		for n < len(b) && b[n] != 0 {
			n++
		}
		fmt.Fprintf(d.out, "%s = %s (%d bytes)\n", d.p.Addr(v), strconv.Quote(string(b[:n])), n)
		return
	}
	size := d.valueSize(expr)
	fmt.Fprintf(d.out, "%s = %s\n", expr, formatValue(v, size))
	if _, isReg := regByName(expr); !isReg {
		if _, isLabel := d.label(expr); isLabel {
			fmt.Fprintln(d.out, "  "+i18n.T("dbg.pLabel", expr))
		}
	}
}

var flagBits = map[string]uint64{}

func init() {
	for _, f := range emu.FlagNames {
		flagBits[strings.ToLower(f.Name)] = f.Bit
	}
}

// setCmd implements:
//
//	set <register> <expr>        set eax 5, set al 'A', set esp esp+4
//	set <flag> 0|1               set zf 1
//	set eip <label|address>      continue somewhere else
//	set [size] [<addr>] <expr>   set dword [count] 0, set byte [msg] 'h'
func (d *stepper) setCmd(args []string) {
	raw := strings.TrimSpace(strings.Join(args, " "))
	if raw == "" {
		fmt.Fprintln(d.out, i18n.T("dbg.setUsage"))
		return
	}
	p := d.parser(raw)
	first := p.word()
	lfirst := strings.ToLower(first)
	value := func(rest string) (uint64, bool) {
		if strings.TrimSpace(rest) == "" {
			fmt.Fprintln(d.out, i18n.T("dbg.setUsage"))
			return 0, false
		}
		v, err := d.eval(rest)
		if err != nil {
			fmt.Fprintln(d.out, err)
			return 0, false
		}
		return v, true
	}
	if bit, ok := flagBits[lfirst]; ok {
		v, ok := value(raw[p.pos:])
		if !ok {
			return
		}
		if v != 0 {
			d.c.RFLAGS |= bit
		} else {
			d.c.RFLAGS &^= bit
		}
		fmt.Fprintf(d.out, "%s = %d   (%s)\n", strings.ToUpper(lfirst), b2i(v != 0), emu.FormatFlags(d.c.RFLAGS))
		return
	}
	if r, ok := regByName(first); ok {
		v, ok := value(raw[p.pos:])
		if !ok {
			return
		}
		d.c.SetValue(r, v)
		fmt.Fprintf(d.out, "%s = %s\n", lfirst, formatValue(d.c.Value(r), regSizeOf(r)))
		return
	}
	if lfirst == "eip" || lfirst == "rip" {
		v, ok := value(raw[p.pos:])
		if !ok {
			return
		}
		if !d.c.InCode(v) {
			fmt.Fprintln(d.out, i18n.T("dbg.notCode", d.p.Addr(v)))
			return
		}
		d.c.RIP = v
		fmt.Fprintln(d.out, "→ "+d.p.Where(v))
		return
	}
	// Memory: [size] [addr] value
	size := d.c.Width()
	p.pos = 0
	if n, ok := sizeWords[lfirst]; ok {
		size = n
		p.word()
	}
	addr, err := p.address()
	if err != nil {
		fmt.Fprintln(d.out, i18n.T("dbg.setUsage"))
		return
	}
	v, ok := value(raw[p.pos:])
	if !ok {
		return
	}
	if err := d.c.Mem.Write(addr, size, v); err != nil {
		fmt.Fprintln(d.out, i18n.T("dbg.setMemFail", d.p.Addr(addr)))
		return
	}
	name := map[int]string{1: "byte", 2: "word", 4: "dword", 8: "qword"}[size]
	fmt.Fprintf(d.out, "%s [%s] = %s\n", name, d.p.Addr(addr), formatValue(v, size))
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
