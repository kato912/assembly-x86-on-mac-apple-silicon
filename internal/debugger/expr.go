package debugger

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86asm"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
)

// Debugger expressions use NASM-like syntax:
//
//	expr    = term { ('+' | '-') term }
//	term    = unary { '*' unary }
//	unary   = '-' unary | primary
//	primary = number | 'c' | '(' expr ')' | [size] '[' expr ']' | register | label
//	size    = byte | word | dword | qword
//
// Numbers: 42, 0x2a, 2ah, 0b101010. A [..] reads memory (default size: the
// CPU width). Labels evaluate to their address, as in NASM.

var sizeWords = map[string]int{"byte": 1, "word": 2, "dword": 4, "qword": 8}

type exprParser struct {
	d   *stepper
	s   string
	pos int
}

func (d *stepper) parser(s string) *exprParser { return &exprParser{d: d, s: s} }

// eval evaluates a whole expression.
func (d *stepper) eval(s string) (uint64, error) {
	p := d.parser(s)
	v, err := p.expr()
	if err != nil {
		return 0, err
	}
	if p.skip(); p.pos != len(p.s) {
		return 0, p.fail()
	}
	return v, nil
}

func (p *exprParser) fail() error {
	return errors.New(i18n.T("dbg.badExpr", strings.TrimSpace(p.s)))
}

func (p *exprParser) skip() {
	for p.pos < len(p.s) && (p.s[p.pos] == ' ' || p.s[p.pos] == '\t') {
		p.pos++
	}
}

func (p *exprParser) peek() byte {
	p.skip()
	if p.pos < len(p.s) {
		return p.s[p.pos]
	}
	return 0
}

func (p *exprParser) expr() (uint64, error) {
	v, err := p.term()
	if err != nil {
		return 0, err
	}
	for {
		switch p.peek() {
		case '+':
			p.pos++
			w, err := p.term()
			if err != nil {
				return 0, err
			}
			v += w
		case '-':
			p.pos++
			w, err := p.term()
			if err != nil {
				return 0, err
			}
			v -= w
		default:
			return v, nil
		}
	}
}

func (p *exprParser) term() (uint64, error) {
	v, err := p.unary()
	if err != nil {
		return 0, err
	}
	for p.peek() == '*' {
		p.pos++
		w, err := p.unary()
		if err != nil {
			return 0, err
		}
		v *= w
	}
	return v, nil
}

func (p *exprParser) unary() (uint64, error) {
	if p.peek() == '-' {
		p.pos++
		v, err := p.unary()
		return -v, err
	}
	return p.primary()
}

// word reads an identifier or number token.
func (p *exprParser) word() string {
	p.skip()
	start := p.pos
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		if c == '_' || c == '.' || c == '$' || c == '?' || c == '@' || c == '#' || c == '~' ||
			c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			p.pos++
			continue
		}
		break
	}
	return p.s[start:p.pos]
}

func (p *exprParser) primary() (uint64, error) {
	switch c := p.peek(); {
	case c == '(':
		p.pos++
		v, err := p.expr()
		if err != nil {
			return 0, err
		}
		if p.peek() != ')' {
			return 0, p.fail()
		}
		p.pos++
		return v, nil
	case c == '\'' || c == '"':
		if p.pos+2 < len(p.s) && p.s[p.pos+2] == c {
			v := uint64(p.s[p.pos+1])
			p.pos += 3
			return v, nil
		}
		return 0, p.fail()
	case c == '[':
		addr, err := p.address()
		if err != nil {
			return 0, err
		}
		return p.d.readMem(addr, p.d.c.Width())
	}
	w := p.word()
	if w == "" {
		return 0, p.fail()
	}
	if size, ok := sizeWords[strings.ToLower(w)]; ok {
		addr, err := p.address()
		if err != nil {
			return 0, err
		}
		return p.d.readMem(addr, size)
	}
	if v, ok := parseNumber(w); ok {
		return v, nil
	}
	if r, ok := regByName(w); ok {
		return p.d.c.Value(r), nil
	}
	if strings.EqualFold(w, "eip") || strings.EqualFold(w, "rip") {
		return p.d.c.RIP, nil
	}
	if a, ok := p.d.label(w); ok {
		return a, nil
	}
	return 0, errors.New(i18n.T("dbg.badExpr", w))
}

// address parses "[expr]" and returns expr (the address, not the memory).
func (p *exprParser) address() (uint64, error) {
	if p.peek() != '[' {
		return 0, p.fail()
	}
	p.pos++
	a, err := p.expr()
	if err != nil {
		return 0, err
	}
	if p.peek() != ']' {
		return 0, p.fail()
	}
	p.pos++
	return a, nil
}

func parseNumber(w string) (uint64, bool) {
	lw := strings.ToLower(strings.ReplaceAll(w, "_", ""))
	if lw == "" || lw[0] < '0' || lw[0] > '9' {
		return 0, false
	}
	var v uint64
	var err error
	switch {
	case strings.HasPrefix(lw, "0x"):
		v, err = strconv.ParseUint(lw[2:], 16, 64)
	case strings.HasPrefix(lw, "0b"):
		v, err = strconv.ParseUint(lw[2:], 2, 64)
	case strings.HasSuffix(lw, "h"):
		v, err = strconv.ParseUint(lw[:len(lw)-1], 16, 64)
	default:
		v, err = strconv.ParseUint(lw, 10, 64)
	}
	return v, err == nil
}

func regByName(w string) (x86asm.Reg, bool) { return emu.RegByName(w) }

// label finds a label's address; ".loop" or "loop" also match a local
// label such as "_start.loop".
func (d *stepper) label(name string) (uint64, bool) {
	if s, ok := d.p.Image.Lookup(name); ok {
		return s.Addr, true
	}
	for _, s := range d.p.Image.Symbols {
		if strings.HasSuffix(s.Name, "."+strings.TrimPrefix(name, ".")) {
			return s.Addr, true
		}
	}
	return 0, false
}

// readMem reads size bytes, reporting unreadable memory as an error.
func (d *stepper) readMem(addr uint64, size int) (uint64, error) {
	b := d.c.Mem.Peek(addr, size)
	if len(b) < size {
		return 0, errors.New(i18n.T("dbg.unreadable", d.p.Addr(addr)))
	}
	var v uint64
	for i := size - 1; i >= 0; i-- {
		v = v<<8 | uint64(b[i])
	}
	return v, nil
}

func regSizeOf(r x86asm.Reg) int {
	switch {
	case r >= x86asm.AL && r <= x86asm.R15B:
		return 1
	case r >= x86asm.AX && r <= x86asm.R15W:
		return 2
	case r >= x86asm.EAX && r <= x86asm.R15L:
		return 4
	}
	return 8
}
