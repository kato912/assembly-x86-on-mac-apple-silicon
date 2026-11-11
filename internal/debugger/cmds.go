package debugger

import (
	"fmt"
	"strconv"
	"strings"

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
