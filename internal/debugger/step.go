package debugger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86asm"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// OpenTerminal opens the controlling terminal for debugger commands, so
// the program itself can still read its own (possibly redirected) stdin.
func OpenTerminal() (io.ReadCloser, error) {
	if runtime.GOOS == "windows" {
		return os.Open("CONIN$")
	}
	return os.Open("/dev/tty")
}

type stepper struct {
	s   *Session
	p   *runner.Program
	c   *emu.CPU
	in  *bufio.Reader
	out io.Writer

	bps     map[uint64]int // address -> breakpoint number
	nextBP  int
	stopped bool // the program has exited or crashed
	code    int  // nasmx86 exit code once stopped
}

// Interactive runs the --step debugger, reading commands from in and
// writing to out. It returns nasmx86's exit code.
func Interactive(s *Session, in io.Reader, out io.Writer) int {
	d := &stepper{s: s, p: s.P, c: s.C, in: bufio.NewReader(in), out: out, bps: map[uint64]int{}}
	fmt.Fprintln(out, i18n.T("dbg.banner"))
	for {
		if !d.stopped {
			fmt.Fprintln(out, "→ "+d.p.Where(d.c.RIP))
		}
		fmt.Fprint(out, "(nasmx86) ")
		line, err := d.in.ReadString('\n')
		if err != nil && line == "" {
			fmt.Fprintln(out)
			return d.code
		}
		f := strings.Fields(line)
		cmd, args := "", f
		if len(f) > 0 {
			cmd, args = strings.ToLower(f[0]), f[1:]
		}
		switch cmd {
		case "", "s", "step", "si":
			n := 1
			if len(args) > 0 {
				if v, err := strconv.Atoi(args[0]); err == nil && v > 0 {
					n = v
				}
			}
			for i := 0; i < n && d.alive(); i++ {
				r := d.s.Step()
				if r.Ran {
					fmt.Fprintln(out, r.Format(d.c.Steps))
				}
				if r.Err != nil {
					d.finish(r.Err)
				}
			}
		case "n", "next", "ni":
			if d.alive() {
				d.next()
			}
		case "c", "cont", "continue", "r!", "run":
			if d.alive() {
				d.run(nil)
			}
		case "b", "break":
			d.breakCmd(args)
		case "d", "delete":
			d.deleteCmd(args)
		case "r", "reg", "regs", "registers", "info":
			fmt.Fprint(out, d.p.FormatRegs())
		case "m", "mem", "memory":
			d.memCmd(args)
		case "x", "stack", "bt":
			d.stackCmd(args)
		case "l", "list":
			d.listCmd()
		case "q", "quit", "exit":
			return d.code
		case "h", "help", "?":
			fmt.Fprint(out, i18n.T("dbg.help"))
		default:
			fmt.Fprintln(out, i18n.T("dbg.unknown", cmd))
		}
	}
}

func (d *stepper) alive() bool {
	if d.stopped {
		fmt.Fprintln(d.out, i18n.T("dbg.finished", d.code))
		return false
	}
	return true
}

func (d *stepper) finish(err error) {
	d.stopped = true
	d.code = d.p.Finish(err)
	var ex *emu.Exit
	if errors.As(err, &ex) {
		fmt.Fprintln(d.out, i18n.T("dbg.exited", d.code))
	} else {
		fmt.Fprintln(d.out, i18n.T("dbg.crashed"))
	}
}

// run executes until a breakpoint, until stop() is true, or until the
// program ends. Ctrl-C interrupts it.
func (d *stepper) run(stop func() bool) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	for first := true; ; first = false {
		select {
		case <-sig:
			fmt.Fprintln(d.out, i18n.T("dbg.interrupted"))
			return
		default:
		}
		if !first {
			if n, ok := d.bps[d.c.RIP]; ok {
				fmt.Fprintln(d.out, i18n.T("dbg.hitBreak", n))
				return
			}
			if stop != nil && stop() {
				return
			}
		}
		if r := d.s.Step(); r.Err != nil {
			d.finish(r.Err)
			return
		}
	}
}

// next steps over a call: it runs until the call returns.
func (d *stepper) next() {
	in, err := d.c.Decode(d.c.RIP)
	if err != nil || in.Op != x86asm.CALL {
		r := d.s.Step()
		if r.Ran {
			fmt.Fprintln(d.out, r.Format(d.c.Steps))
		}
		if r.Err != nil {
			d.finish(r.Err)
		}
		return
	}
	ret, sp := d.c.RIP+uint64(in.Len), d.c.SP()
	d.run(func() bool { return d.c.RIP == ret && d.c.SP() >= sp })
}

// resolve turns a breakpoint argument — a line number ("12"), a line in
// another file ("util.asm:5"), a label or an address — into a code address
// and a description.
func (d *stepper) resolve(arg string) (uint64, string, error) {
	file, lineArg := 0, arg
	if i := strings.LastIndexByte(arg, ':'); i > 0 {
		f, ok := d.p.FileIndex(arg[:i])
		if !ok {
			return 0, "", errors.New(i18n.T("dbg.noFile", arg[:i]))
		}
		file, lineArg = f, arg[i+1:]
	}
	if n, err := strconv.Atoi(lineArg); err == nil {
		addr, line, text, ok := d.p.LineAddr(file, n)
		if !ok {
			return 0, "", errors.New(i18n.T("dbg.noCode", n))
		}
		desc := i18n.T("loc.line", line, text)
		if len(d.p.Files) > 1 {
			desc = filepath.Base(d.p.Files[file].Path) + " " + desc
		}
		return addr, desc, nil
	}
	a, err := d.eval(arg)
	if err != nil {
		return 0, "", err
	}
	if !d.c.InCode(a) {
		return 0, "", errors.New(i18n.T("dbg.notCode", arg))
	}
	return a, d.p.Where(a), nil
}

func (d *stepper) breakCmd(args []string) {
	if len(args) == 0 {
		if len(d.bps) == 0 {
			fmt.Fprintln(d.out, i18n.T("dbg.noBreaks"))
		}
		type bp struct {
			n    int
			addr uint64
		}
		var list []bp
		for a, n := range d.bps {
			list = append(list, bp{n, a})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].n < list[j].n })
		for _, b := range list {
			fmt.Fprintf(d.out, "  #%d  %s\n", b.n, d.p.Where(b.addr))
		}
		return
	}
	a, desc, err := d.resolve(strings.Join(args, ""))
	if err != nil {
		fmt.Fprintln(d.out, err)
		return
	}
	if n, ok := d.bps[a]; ok {
		fmt.Fprintln(d.out, i18n.T("dbg.breakSet", n, desc))
		return
	}
	d.nextBP++
	d.bps[a] = d.nextBP
	fmt.Fprintln(d.out, i18n.T("dbg.breakSet", d.nextBP, desc))
}

func (d *stepper) deleteCmd(args []string) {
	if len(args) == 0 {
		d.bps = map[uint64]int{}
		fmt.Fprintln(d.out, i18n.T("dbg.breaksCleared"))
		return
	}
	arg := strings.TrimPrefix(args[0], "#")
	if n, err := strconv.Atoi(arg); err == nil {
		for a, b := range d.bps {
			if b == n {
				delete(d.bps, a)
				fmt.Fprintln(d.out, i18n.T("dbg.breakDeleted", n))
				return
			}
		}
	}
	if a, _, err := d.resolve(args[0]); err == nil {
		if n, ok := d.bps[a]; ok {
			delete(d.bps, a)
			fmt.Fprintln(d.out, i18n.T("dbg.breakDeleted", n))
			return
		}
	}
	fmt.Fprintln(d.out, i18n.T("dbg.noSuchBreak", args[0]))
}

// eval evaluates "label", "0x1234", "1234", "esp", "msg+4", "rbp-8".
func (d *stepper) eval(expr string) (uint64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		return 0, errors.New(i18n.T("dbg.badExpr", expr))
	}
	var total uint64
	sign := uint64(1)
	start := 0
	for i := 0; i <= len(expr); i++ {
		if i < len(expr) && !(i > start && (expr[i] == '+' || expr[i] == '-')) {
			continue
		}
		v, err := d.term(expr[start:i])
		if err != nil {
			return 0, err
		}
		total += sign * v
		if i < len(expr) {
			sign = 1
			if expr[i] == '-' {
				sign = ^uint64(0)
			}
		}
		start = i + 1
	}
	return total, nil
}

func (d *stepper) term(t string) (uint64, error) {
	lt := strings.ToLower(t)
	switch {
	case strings.HasPrefix(lt, "0x"):
		if v, err := strconv.ParseUint(lt[2:], 16, 64); err == nil {
			return v, nil
		}
	case strings.HasSuffix(lt, "h") && len(lt) > 1 && lt[0] >= '0' && lt[0] <= '9':
		if v, err := strconv.ParseUint(lt[:len(lt)-1], 16, 64); err == nil {
			return v, nil
		}
	case lt != "" && lt[0] >= '0' && lt[0] <= '9':
		if v, err := strconv.ParseUint(lt, 10, 64); err == nil {
			return v, nil
		}
	}
	if r, ok := emu.RegByName(lt); ok {
		return d.c.Value(r), nil
	}
	if lt == "eip" || lt == "rip" {
		return d.c.RIP, nil
	}
	if s, ok := d.p.Image.Lookup(t); ok {
		return s.Addr, nil
	}
	for _, s := range d.p.Image.Symbols { // local labels: _start.loop -> .loop / loop
		if strings.HasSuffix(s.Name, "."+strings.TrimPrefix(t, ".")) {
			return s.Addr, nil
		}
	}
	return 0, errors.New(i18n.T("dbg.badExpr", t))
}

func (d *stepper) memCmd(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(d.out, i18n.T("dbg.memUsage"))
		return
	}
	n := 64
	exprArgs := args
	if len(args) > 1 {
		if v, err := strconv.ParseUint(args[len(args)-1], 0, 32); err == nil {
			n = int(v)
			exprArgs = args[:len(args)-1]
		}
	}
	if n > 4096 {
		n = 4096
	}
	addr, err := d.eval(strings.Join(exprArgs, ""))
	if err != nil {
		fmt.Fprintln(d.out, err)
		return
	}
	data := d.c.Mem.Peek(addr, n)
	if len(data) == 0 {
		fmt.Fprintln(d.out, i18n.T("dbg.unreadable", d.p.Addr(addr)))
		return
	}
	for off := 0; off < n; off += 16 {
		var hex, asc strings.Builder
		for j := 0; j < 16 && off+j < n; j++ {
			if j == 8 {
				hex.WriteByte(' ')
			}
			if off+j < len(data) {
				b := data[off+j]
				fmt.Fprintf(&hex, "%02x ", b)
				if b >= 32 && b < 127 {
					asc.WriteByte(b)
				} else {
					asc.WriteByte('.')
				}
			} else {
				hex.WriteString("?? ")
				asc.WriteByte(' ')
			}
		}
		label := ""
		if s, ok := d.p.Image.SymbolFor(addr + uint64(off)); ok && s.Addr == addr+uint64(off) {
			label = "  <" + s.Name + ">"
		}
		fmt.Fprintf(d.out, "%s  %-49s |%s|%s\n", d.p.Addr(addr+uint64(off)), hex.String(), asc.String(), label)
	}
}

func (d *stepper) stackCmd(args []string) {
	n := 8
	if len(args) > 0 {
		if v, err := strconv.Atoi(args[0]); err == nil && v > 0 {
			n = v
		}
	}
	w := d.c.Width()
	sp := d.c.SP()
	bp := d.c.R[emu.RBP] & (^uint64(0) >> (64 - 8*uint(w)))
	spName, bpName := emu.RegName(emu.RSP, d.p.Bits), emu.RegName(emu.RBP, d.p.Bits)
	for i := 0; i < n; i++ {
		a := sp + uint64(i*w)
		b := d.c.Mem.Peek(a, w)
		if len(b) < w {
			break
		}
		var v uint64
		for k := w - 1; k >= 0; k-- {
			v = v<<8 | uint64(b[k])
		}
		var marks []string
		if a == sp {
			marks = append(marks, "← "+spName)
		}
		if a == bp {
			marks = append(marks, "← "+bpName)
		}
		note := ""
		switch {
		case a == d.p.InitialSP && d.p.Image.EntryName != "main":
			note = "argc"
		case d.c.InCode(v):
			if loc, ok := d.p.LocAt(v); ok {
				note = i18n.T("dbg.retAddr", loc.Line)
			} else if v == d.p.ExitStub {
				note = i18n.T("dbg.retMain")
			}
		default:
			if s, ok := d.p.Image.SymbolFor(v); ok && s.Addr == v && v != 0 {
				note = "→ " + s.Name
			}
		}
		fmt.Fprintf(d.out, "  %s: %s  %-8s %s\n", d.p.Addr(a), runner.Hex(v, w), strings.Join(marks, " "), note)
	}
}

func (d *stepper) listCmd() {
	file, cur := 0, 0
	if f, loc, ok := d.p.Locate(d.c.RIP); ok && loc.File == "" {
		file, cur = f, loc.Line
	}
	src := d.p.Files[file].Source
	bpLines := map[int]bool{}
	for a := range d.bps {
		if f, loc, ok := d.p.Locate(a); ok && f == file && loc.File == "" {
			bpLines[loc.Line] = true
		}
	}
	if len(d.p.Files) > 1 {
		fmt.Fprintf(d.out, "  %s\n", filepath.Base(d.p.Files[file].Path))
	}
	from, to := cur-5, cur+5
	if from < 1 {
		from = 1
	}
	if to > len(src) {
		to = len(src)
	}
	for l := from; l <= to; l++ {
		mark := "  "
		if bpLines[l] {
			mark = "● "
		}
		if l == cur {
			mark = "→ "
		}
		fmt.Fprintf(d.out, "%s%4d  %s\n", mark, l, strings.ReplaceAll(src[l-1], "\t", "    "))
	}
}
