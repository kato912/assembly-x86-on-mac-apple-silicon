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
	case "display":
		d.infoDisplay()
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
		enb := "y"
		if w.disabled {
			enb = "n"
		}
		items = append(items, item{w.n, fmt.Sprintf("%-7d %-14s %-4s %-3s %-10s %s", w.n, "watchpoint", "keep", enb, "", w.desc)})
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

// xState remembers gdb's last examine format, unit and next address, so a
// bare "x" continues where the previous one stopped.
type xState struct {
	format byte
	unit   int
	next   uint64
}

// parseXFormat splits gdb's "/NFU" into count, format letter and unit size.
func (d *stepper) parseXFormat(f string) (count int, format byte, unit int) {
	count, format, unit = 1, d.xs.format, d.xs.unit
	if format == 0 {
		format, unit = 'x', 4
	}
	digits := ""
	for _, c := range f {
		switch {
		case c >= '0' && c <= '9':
			digits += string(c)
		case strings.ContainsRune("xduotacsiz", c):
			format = byte(c)
		case c == 'b':
			unit = 1
		case c == 'h':
			unit = 2
		case c == 'w':
			unit = 4
		case c == 'g':
			unit = 8
		}
	}
	if n, err := strconv.Atoi(digits); err == nil && n > 0 {
		count = n
	}
	if count > 1024 {
		count = 1024
	}
	return
}

// insnText renders one instruction gdb-style: "mov    eax, 0x4".
func (d *stepper) insnText(addr uint64) (string, int, bool) {
	in, err := d.c.Decode(addr)
	if err != nil {
		return "", 0, false
	}
	text := d.p.Disasm(in, addr)
	if i := strings.IndexByte(text, ' '); i > 0 {
		text = fmt.Sprintf("%-6s %s", text[:i], text[i+1:])
	}
	return text, in.Len, true
}

// examine implements gdb's x/NFU addr.
func (d *stepper) examine(format, expr string) {
	count, f, unit := d.parseXFormat(format)
	addr := d.xs.next
	if strings.TrimSpace(expr) != "" {
		v, err := d.eval(expr)
		if err != nil {
			fmt.Fprintln(d.out, err)
			return
		}
		addr = v
	}
	d.xs.format = f
	if f != 's' && f != 'i' {
		d.xs.unit = unit
	}
	label := func(a uint64) string { return d.symbolic(a) + ":" }
	switch f {
	case 's':
		for i := 0; i < count; i++ {
			b := d.c.Mem.Peek(addr, 4096)
			if len(b) == 0 {
				fmt.Fprintln(d.out, i18n.T("dbg.cannotAccess", fmt.Sprintf("%#x", addr)))
				break
			}
			n := 0
			for n < len(b) && b[n] != 0 {
				n++
			}
			fmt.Fprintf(d.out, "%s\t%s\n", label(addr), strconv.Quote(string(b[:n])))
			addr += uint64(n) + 1
		}
	case 'i':
		for i := 0; i < count; i++ {
			text, n, ok := d.insnText(addr)
			if !ok {
				fmt.Fprintln(d.out, i18n.T("dbg.cannotAccess", fmt.Sprintf("%#x", addr)))
				break
			}
			mark := "   "
			if addr == d.c.RIP {
				mark = "=> "
			}
			fmt.Fprintf(d.out, "%s%s\t%s\n", mark, label(addr), text)
			addr += uint64(n)
		}
	default:
		perLine := map[int]int{1: 8, 2: 8, 4: 4, 8: 2}[unit]
		for i := 0; i < count; i++ {
			if i%perLine == 0 {
				if i > 0 {
					fmt.Fprintln(d.out)
				}
				fmt.Fprint(d.out, label(addr))
			}
			v, err := d.readMem(addr, unit)
			if err != nil {
				fmt.Fprintln(d.out)
				fmt.Fprintln(d.out, i18n.T("dbg.cannotAccess", fmt.Sprintf("%#x", addr)))
				d.xs.next = addr
				return
			}
			var s string
			switch f {
			case 'x':
				s = fmt.Sprintf("0x%0*x", 2*unit, v)
			case 't':
				s = fmt.Sprintf("%0*b", 8*unit, v)
			default:
				s = d.formatAs(v, unit, f)
			}
			fmt.Fprint(d.out, "\t"+s)
			addr += uint64(unit)
		}
		fmt.Fprintln(d.out)
	}
	d.xs.next = addr
}

// function finds the function (non-local label) containing addr and the
// end of its code.
func (d *stepper) function(addr uint64) (name string, start, end uint64, ok bool) {
	sec := d.p.Image.SectionAt(addr)
	if sec == nil || !sec.Exec {
		return "", 0, 0, false
	}
	start, end = sec.Addr, sec.End()
	name = "??"
	for _, s := range d.p.Image.Symbols { // sorted by address
		if s.Section != sec.Name || strings.Contains(s.Name, ".") || d.p.Image.SectionAt(s.Addr) != sec {
			continue
		}
		if s.Addr <= addr {
			name, start = s.Name, s.Addr
		} else if s.Addr < end {
			end = s.Addr
			break
		}
	}
	return name, start, end, true
}

// disassemble implements gdb's disassemble [function|address].
func (d *stepper) disassemble(expr string) {
	at := d.c.RIP
	if strings.TrimSpace(expr) != "" {
		v, err := d.eval(expr)
		if err != nil {
			fmt.Fprintln(d.out, err)
			return
		}
		at = v
	}
	name, start, end, ok := d.function(at)
	if !ok {
		fmt.Fprintln(d.out, i18n.T("dbg.notCode", fmt.Sprintf("%#x", at)))
		return
	}
	fmt.Fprintln(d.out, i18n.T("dbg.disasHeader", name))
	for a, n := start, 0; a < end && n < 200; n++ {
		text, length, ok := d.insnText(a)
		if !ok {
			break
		}
		mark := "   "
		if a == d.c.RIP {
			mark = "=> "
		}
		line := ""
		if loc, ok := d.p.LocAt(a); ok {
			line = fmt.Sprintf("  ; line %d", loc.Line)
		}
		fmt.Fprintf(d.out, "%s%#08x <+%d>:\t%-28s%s\n", mark, a, a-start, text, line)
		a += uint64(length)
	}
	fmt.Fprintln(d.out, i18n.T("dbg.disasEnd"))
}

// autoDisplay is an expression shown every time the program stops.
type autoDisplay struct {
	n       int
	format  string
	expr    string
	enabled bool
}

// examineStyle reports whether a display format is an x/ format (i, s, a
// count or a unit) rather than a print format.
func examineStyle(format string) bool {
	return strings.ContainsAny(format, "isbhwg0123456789")
}

func (d *stepper) showDisplay(a *autoDisplay) {
	if examineStyle(a.format) {
		fmt.Fprintf(d.out, "%d: x/%s %s\n", a.n, a.format, a.expr)
		save := d.xs
		d.examine(a.format, a.expr)
		d.xs = save
		return
	}
	v, err := d.eval(a.expr)
	if err != nil {
		fmt.Fprintf(d.out, "%d: %s = <%v>\n", a.n, a.expr, err)
		return
	}
	f := byte('d')
	label := a.expr
	if a.format != "" {
		f = a.format[len(a.format)-1]
		label = "/" + a.format + " " + a.expr
	}
	val := d.formatAs(v, d.gdbSize(a.expr), f)
	if e := strings.ToLower(a.expr); (e == "$eflags" || e == "$rflags") && a.format == "" {
		val = gdbFlags(v)
	}
	fmt.Fprintf(d.out, "%d: %s = %s\n", a.n, label, val)
}

// showDisplays prints every enabled auto-display (after the program stops).
func (d *stepper) showDisplays() {
	for _, a := range d.displays {
		if a.enabled {
			d.showDisplay(a)
		}
	}
}

// displayCmd implements display[/FMT] expr (and plain display).
func (d *stepper) displayCmd(format, expr string) {
	if strings.TrimSpace(expr) == "" {
		d.showDisplays()
		return
	}
	d.nextDisplay++
	a := &autoDisplay{n: d.nextDisplay, format: format, expr: strings.TrimSpace(expr), enabled: true}
	d.displays = append(d.displays, a)
	if !d.stopped {
		d.showDisplay(a)
	}
}

// undisplayCmd removes auto-displays (all of them without arguments).
func (d *stepper) undisplayCmd(args []string) {
	if len(args) == 0 {
		d.displays = nil
		return
	}
	for _, arg := range args {
		n, _ := strconv.Atoi(arg)
		for i, a := range d.displays {
			if a.n == n {
				d.displays = append(d.displays[:i], d.displays[i+1:]...)
				break
			}
		}
	}
}

func (d *stepper) infoDisplay() {
	if len(d.displays) == 0 {
		fmt.Fprintln(d.out, i18n.T("dbg.noDisplays"))
		return
	}
	fmt.Fprintln(d.out, "Auto-display expressions now in effect:\nNum Enb Expression")
	for _, a := range d.displays {
		enb := "y"
		if !a.enabled {
			enb = "n"
		}
		f := ""
		if a.format != "" {
			f = "/" + a.format + " "
		}
		fmt.Fprintf(d.out, "%d:   %s  %s%s\n", a.n, enb, f, a.expr)
	}
}

// enableCmd implements enable/disable [n...] for breakpoints, watchpoints
// and (with "display") auto-displays.
func (d *stepper) enableCmd(on bool, args []string) {
	displays := len(args) > 0 && strings.HasPrefix(strings.ToLower(args[0]), "disp")
	if displays {
		args = args[1:]
	}
	match := func(n int) bool {
		if len(args) == 0 {
			return true
		}
		for _, a := range args {
			if v, err := strconv.Atoi(strings.TrimPrefix(a, "#")); err == nil && v == n {
				return true
			}
		}
		return false
	}
	if displays {
		for _, a := range d.displays {
			if match(a.n) {
				a.enabled = on
			}
		}
		return
	}
	for _, b := range d.bps {
		if match(b.n) {
			b.enabled = on
		}
	}
	for _, w := range d.watches {
		if match(w.n) {
			w.disabled = !on
		}
	}
}

// untilCmd implements until/advance <location>: run to it once.
func (d *stepper) untilCmd(args []string) {
	if len(args) == 0 {
		d.next()
		return
	}
	a, _, err := d.resolve(strings.Join(args, ""))
	if err != nil {
		fmt.Fprintln(d.out, err)
		return
	}
	d.run(func() bool { return d.c.RIP == a })
}
