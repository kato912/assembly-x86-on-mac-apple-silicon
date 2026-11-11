package debugger

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
)

// gdb compatibility: commands and output formats as lab sheets written for
// gdb expect them (info registers, p/x, x/4xw, disassemble, display ...).

// Segment register values a Linux process sees (what gdb shows natively).
var segRegs32 = []struct {
	name string
	v    uint64
}{{"cs", 0x23}, {"ss", 0x2b}, {"ds", 0x2b}, {"es", 0x2b}, {"fs", 0}, {"gs", 0}}
var segRegs64 = []struct {
	name string
	v    uint64
}{{"cs", 0x33}, {"ss", 0x2b}, {"ds", 0}, {"es", 0}, {"fs", 0}, {"gs", 0}}

// gdbFlags renders eflags like gdb: "[ PF ZF IF ]".
func gdbFlags(f uint64) string {
	names := []struct {
		n string
		b uint64
	}{{"CF", emu.CF}, {"PF", emu.PF}, {"AF", emu.AF}, {"ZF", emu.ZF}, {"SF", emu.SF},
		{"TF", emu.TF}, {"IF", emu.IF}, {"DF", emu.DF}, {"OF", emu.OF}}
	var on []string
	for _, x := range names {
		if f&x.b != 0 {
			on = append(on, x.n)
		}
	}
	return "[ " + strings.Join(on, " ") + " ]"
}

// symbolic renders an address like gdb: "0x8049005 <_start+5>".
func (d *stepper) symbolic(a uint64) string {
	s := fmt.Sprintf("%#x", a)
	if sym, ok := d.p.Image.SymbolFor(a); ok && sym.Section != "" {
		if a == sym.Addr {
			s += " <" + sym.Name + ">"
		} else {
			s += fmt.Sprintf(" <%s+%d>", sym.Name, a-sym.Addr)
		}
	}
	return s
}

// infoRegisters prints registers in gdb's "info registers" layout. With
// names it prints only those (info registers eax eflags).
func (d *stepper) infoRegisters(names []string) {
	type row struct{ name, hex, natural string }
	var rows []row
	w := d.c.Width()
	n := 8
	if w == 8 {
		n = 16
	}
	order := []int{emu.RAX, emu.RCX, emu.RDX, emu.RBX, emu.RSP, emu.RBP, emu.RSI, emu.RDI,
		emu.R8, emu.R9, emu.R10, emu.R11, emu.R12, emu.R13, emu.R14, emu.R15}[:n]
	for _, i := range order {
		v := d.c.R[i] & widthMask(w)
		nat := strconv.FormatInt(emuSignExtend(v, w), 10)
		if i == emu.RSP || i == emu.RBP {
			nat = fmt.Sprintf("%#x", v)
		}
		rows = append(rows, row{emu.RegName(i, d.p.Bits), fmt.Sprintf("%#x", v), nat})
	}
	ip := "eip"
	segs := segRegs32
	if w == 8 {
		ip, segs = "rip", segRegs64
	}
	rows = append(rows, row{ip, fmt.Sprintf("%#x", d.c.RIP), d.symbolic(d.c.RIP)})
	rows = append(rows, row{"eflags", fmt.Sprintf("%#x", d.c.RFLAGS), gdbFlags(d.c.RFLAGS)})
	for _, s := range segs {
		rows = append(rows, row{s.name, fmt.Sprintf("%#x", s.v), strconv.FormatUint(s.v, 10)})
	}
	if w == 8 {
		rows = append(rows, row{"fs_base", "0x0", "0"}, row{"gs_base", "0x0", "0"})
	}
	print := func(r row) { fmt.Fprintf(d.out, "%-15s%-19s%s\n", r.name, r.hex, r.natural) }
	if len(names) == 0 {
		for _, r := range rows {
			print(r)
		}
		return
	}
	for _, name := range names {
		name = strings.ToLower(strings.TrimPrefix(name, "$"))
		switch name {
		case "pc":
			name = ip
		case "sp":
			name = emu.RegName(emu.RSP, d.p.Bits)
		case "fp":
			name = emu.RegName(emu.RBP, d.p.Bits)
		case "rflags", "flags":
			name = "eflags"
		}
		found := false
		for _, r := range rows {
			if r.name == name {
				print(r)
				found = true
			}
		}
		if !found {
			if r, ok := regByName(name); ok { // sub-registers: al, ax, r8d ...
				v := d.c.Value(r)
				print(row{name, fmt.Sprintf("%#x", v), strconv.FormatInt(emuSignExtend(v, regSizeOf(r)), 10)})
				continue
			}
			fmt.Fprintln(d.out, i18n.T("dbg.badRegister", name))
		}
	}
}

// formatAs renders v (size bytes) in a gdb print format letter.
func (d *stepper) formatAs(v uint64, size int, f byte) string {
	v &= widthMask(size)
	switch f {
	case 'x':
		return fmt.Sprintf("%#x", v)
	case 'z':
		return fmt.Sprintf("0x%0*x", 2*size, v)
	case 'o':
		if v == 0 {
			return "0"
		}
		return fmt.Sprintf("%#o", v)
	case 't':
		return strconv.FormatUint(v, 2)
	case 'u':
		return strconv.FormatUint(v, 10)
	case 'c':
		c := byte(v)
		if c >= 32 && c < 127 {
			return fmt.Sprintf("%d '%c'", int8(c), c)
		}
		return fmt.Sprintf("%d %q", int8(c), string(rune(c)))
	case 'a':
		return d.symbolic(v)
	default: // 'd' and anything else
		return strconv.FormatInt(emuSignExtend(v, size), 10)
	}
}

// gdbPrint implements p/FMT expr with gdb's "$N = value" output.
func (d *stepper) gdbPrint(format, expr string) {
	if expr == "" {
		fmt.Fprintln(d.out, i18n.T("dbg.pUsage"))
		return
	}
	f := byte('d')
	if format != "" {
		f = format[len(format)-1]
	}
	if f == 's' {
		d.printCmd("p/s", strings.Fields(expr))
		return
	}
	v, err := d.eval(expr)
	if err != nil {
		fmt.Fprintln(d.out, err)
		return
	}
	d.histN++
	val := d.formatAs(v, d.gdbSize(expr), f)
	if e := strings.ToLower(strings.TrimSpace(expr)); e == "$eflags" || e == "$rflags" {
		val = gdbFlags(v)
		if f != 'd' {
			val = d.formatAs(v, 4, f)
		}
	}
	fmt.Fprintf(d.out, "$%d = %s\n", d.histN, val)
}

// gdbSize is the display size of an expression: a register's own size, a
// cast's size, otherwise the CPU width (int-sized in 32-bit code).
func (d *stepper) gdbSize(expr string) int {
	e := strings.TrimSpace(expr)
	if strings.HasPrefix(e, "$") {
		if r, ok := regByName(e[1:]); ok {
			return regSizeOf(r)
		}
	}
	p := d.parser(e)
	if p.peek() == '(' || p.peek() == '*' || p.peek() == '{' {
		p.pos++
		if p.s[p.pos-1] == '*' && p.peek() == '(' {
			p.pos++
		}
		if size, _, _, ok := p.cType(); ok {
			return size
		}
	}
	return d.valueSize(expr)
}

var gdbExprRE = regexp.MustCompile(`\$|&|\{|\*\s*\(|\(\s*(unsigned\s+|signed\s+)?(char|short|int|long|u?int(8|16|32|64)_t)\b`)

// looksGdb reports whether an expression uses gdb/C syntax ($eax, &x,
// (int)x, *(int*)p, {int}p), so its value is printed the way gdb does.
func looksGdb(expr string) bool { return gdbExprRE.MatchString(expr) }

// infoCmd implements gdb's info subcommands.
func (d *stepper) infoCmd(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(d.out, i18n.T("dbg.infoUsage"))
		return
	}
	switch strings.ToLower(args[0]) {
	case "registers", "reg", "r", "all-registers", "registers-all":
		d.infoRegisters(args[1:])
	case "breakpoints", "break", "b", "watchpoints", "watch":
		d.infoBreakpoints()
	case "frame", "f", "stack", "s":
		d.btCmd()
	case "line", "source", "program":
		fmt.Fprintln(d.out, "→ "+d.p.Where(d.c.RIP))
	default:
		fmt.Fprintln(d.out, i18n.T("dbg.infoUsage"))
	}
}

// infoBreakpoints lists breakpoints and watchpoints like gdb's table.
func (d *stepper) infoBreakpoints() {
	type item struct {
		n    int
		line string
	}
	var items []item
	for a, b := range d.bps {
		kind, disp := "breakpoint", "keep"
		if b.temp {
			disp = "del"
		}
		enb := "y"
		if !b.enabled {
			enb = "n"
		}
		items = append(items, item{b.n, fmt.Sprintf("%-7d %-14s %-4s %-3s %-10s %s", b.n, kind, disp, enb, fmt.Sprintf("%#08x", a), d.p.Where(a))})
	}
	for _, w := range d.watches {
		items = append(items, item{w.n, fmt.Sprintf("%-7d %-14s %-4s %-3s %-10s %s", w.n, "watchpoint", "keep", "y", "", w.desc)})
	}
	if len(items) == 0 {
		fmt.Fprintln(d.out, i18n.T("dbg.noBreaks"))
		return
	}
	for i := 0; i < len(items); i++ { // sort by number
		for j := i + 1; j < len(items); j++ {
			if items[j].n < items[i].n {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	fmt.Fprintf(d.out, "%-7s %-14s %-4s %-3s %-10s %s\n", "Num", "Type", "Disp", "Enb", "Address", "What")
	for _, it := range items {
		fmt.Fprintln(d.out, it.line)
	}
}
