package debugger

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86asm"

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
		if len(b) == 0 {
			fmt.Fprintln(d.out, i18n.T("dbg.unreadable", d.p.Addr(v)))
			return
		}
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
	if d.gdbSet(raw) {
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

// funcName returns the function (non-local label) containing addr.
func (d *stepper) funcName(addr uint64) string {
	sym, ok := d.p.Image.SymbolFor(addr)
	if !ok {
		return "?"
	}
	if i := strings.IndexByte(sym.Name, '.'); i > 0 {
		return sym.Name[:i] // "fact.base" is the local label .base inside fact
	}
	return sym.Name
}

// btCmd prints the call stack: where we are, then each active call.
func (d *stepper) btCmd() {
	frames := d.s.Frames()
	type row struct{ fn, where string }
	rows := []row{{d.funcName(d.c.RIP), d.p.Where(d.c.RIP)}}
	for i := len(frames) - 1; i >= 0; i-- {
		rows = append(rows, row{d.funcName(frames[i].Call), d.p.Where(frames[i].Call)})
	}
	w := 0
	for _, r := range rows {
		w = max(w, len(r.fn))
	}
	for i, r := range rows {
		if len(rows) > 24 && i == 12 {
			fmt.Fprintf(d.out, "   ... %d more ...\n", len(rows)-24)
		}
		if len(rows) > 24 && i >= 12 && i < len(rows)-12 {
			continue
		}
		fmt.Fprintf(d.out, "#%-3d %-*s  %s\n", i, w, r.fn, r.where)
	}
}

// finishCmd runs until the current function returns to its caller.
func (d *stepper) finishCmd() {
	frames := d.s.Frames()
	if len(frames) == 0 {
		fmt.Fprintln(d.out, i18n.T("dbg.finishTop"))
		return
	}
	depth := len(frames)
	fn := d.funcName(frames[depth-1].Target)
	d.run(func() bool { return len(d.s.Frames()) < depth })
	if !d.stopped && len(d.s.Frames()) < depth {
		reg := emu.RegName(emu.RAX, d.p.Bits)
		fmt.Fprintln(d.out, i18n.T("dbg.returned", fn, reg, d.p.ValueText(emu.RAX, d.c.R[emu.RAX])))
	}
}

// watchpoint stops execution when a register or memory location changes.
type watchpoint struct {
	n        int
	disabled bool
	desc     string // "dword [counter]" or "ecx"
	isReg    bool
	reg      x86asm.Reg
	addr     uint64
	size     int
	old      uint64
}

func (d *stepper) watchValue(w *watchpoint) (uint64, bool) {
	if w.isReg {
		return d.c.Value(w.reg), true
	}
	v, err := d.readMem(w.addr, w.size)
	return v, err == nil
}

// watchCmd implements "watch <register>" and "watch [size] <address>",
// e.g. watch ecx, watch counter, watch byte [buf+3]. Alone it lists them.
func (d *stepper) watchCmd(args []string) {
	raw := strings.TrimSpace(strings.Join(args, " "))
	if raw == "" {
		if len(d.watches) == 0 {
			fmt.Fprintln(d.out, i18n.T("dbg.noWatches"))
		}
		for _, w := range d.watches {
			v, _ := d.watchValue(w)
			fmt.Fprintf(d.out, "  #%d  %s = %s\n", w.n, w.desc, formatValue(v, w.size))
		}
		return
	}
	w := &watchpoint{size: d.c.Width()}
	if r, ok := regByName(raw); ok {
		w.isReg, w.reg, w.size, w.desc = true, r, regSizeOf(r), strings.ToLower(raw)
	} else {
		p := d.parser(raw)
		sizeName := map[int]string{1: "byte", 2: "word", 4: "dword", 8: "qword"}[w.size]
		if n, ok := sizeWords[strings.ToLower(p.word())]; ok {
			w.size, sizeName = n, strings.ToLower(raw[:p.pos])
		} else {
			p.pos = 0
		}
		rest := strings.TrimSpace(raw[p.pos:])
		inner := strings.TrimSuffix(strings.TrimPrefix(rest, "["), "]")
		addr, err := d.eval(inner)
		if err != nil {
			fmt.Fprintln(d.out, err)
			fmt.Fprintln(d.out, i18n.T("dbg.watchUsage"))
			return
		}
		w.addr = addr
		w.desc = sizeName + " [" + strings.TrimSpace(inner) + "]"
	}
	v, ok := d.watchValue(w)
	if !ok {
		fmt.Fprintln(d.out, i18n.T("dbg.unreadable", d.p.Addr(w.addr)))
		return
	}
	w.old = v
	d.nextBP++
	w.n = d.nextBP
	d.watches = append(d.watches, w)
	fmt.Fprintln(d.out, i18n.T("dbg.watchSet", w.n, w.desc, formatValue(v, w.size)))
}

// checkWatches reports every watchpoint whose value changed during the
// instruction at pc. It returns true if any did.
func (d *stepper) checkWatches(pc uint64) bool {
	hit := false
	for _, w := range d.watches {
		v, ok := d.watchValue(w)
		if !ok || v == w.old {
			continue
		}
		if w.disabled {
			w.old = v
			continue
		}
		short := func(x uint64) string {
			if w.size < 8 {
				x &= 1<<(8*uint(w.size)) - 1
			}
			return fmt.Sprintf("%s (%d)", runner.Hex(x, w.size), x)
		}
		fmt.Fprintln(d.out, i18n.T("dbg.watchHit", w.n, w.desc, short(w.old), short(v), d.p.Where(pc)))
		w.old = v
		hit = true
	}
	return hit
}
