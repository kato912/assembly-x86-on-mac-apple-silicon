package debugger

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86asm"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/libc"
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

// tv is a value with the little type information gdb-style expressions
// need: casts to pointers ("(int*)x") and bare labels ("(int)counter"
// reads the variable, as gdb does for symbols without debug info).
type tv struct {
	v   uint64
	ptr int  // size of the pointed-to type after a (type*) cast
	lbl bool // the value is a bare label's address
}

func (p *exprParser) expr() (uint64, error) {
	t, err := p.exprTV()
	return t.v, err
}

func (p *exprParser) exprTV() (tv, error) {
	t, err := p.term()
	if err != nil {
		return tv{}, err
	}
	for {
		op := p.peek()
		if op != '+' && op != '-' {
			return t, nil
		}
		p.pos++
		w, err := p.term()
		if err != nil {
			return tv{}, err
		}
		if op == '+' {
			t = tv{v: t.v + w.v}
		} else {
			t = tv{v: t.v - w.v}
		}
	}
}

func (p *exprParser) term() (tv, error) {
	t, err := p.unary()
	if err != nil {
		return tv{}, err
	}
	for p.peek() == '*' {
		p.pos++
		w, err := p.unary()
		if err != nil {
			return tv{}, err
		}
		t = tv{v: t.v * w.v}
	}
	return t, nil
}

func (p *exprParser) unary() (tv, error) {
	switch p.peek() {
	case '-':
		p.pos++
		t, err := p.unary()
		return tv{v: -t.v}, err
	case '*': // gdb dereference: *(int*)&x, *$esp
		p.pos++
		t, err := p.unary()
		if err != nil {
			return tv{}, err
		}
		size := t.ptr
		if size == 0 {
			size = p.d.c.Width()
		}
		v, err := p.d.readMem(t.v, size)
		return tv{v: v}, err
	case '&': // gdb address-of: &counter is just the label's address
		p.pos++
		t, err := p.unary()
		return tv{v: t.v}, err
	}
	return p.primary()
}

// cType parses a C type name such as "int", "unsigned char" or "int*".
// It returns false (and consumes nothing) if the text is not a type.
func (p *exprParser) cType() (size int, signed, pointer bool, ok bool) {
	start := p.pos
	signed = true
	var words []string
	for {
		save := p.pos
		w := strings.ToLower(p.word())
		switch w {
		case "unsigned":
			signed = false
		case "signed", "char", "short", "int", "long":
		case "int8_t", "uint8_t", "int16_t", "uint16_t", "int32_t", "uint32_t", "int64_t", "uint64_t":
			signed = !strings.HasPrefix(w, "u")
		default:
			p.pos = save
			goto done
		}
		words = append(words, w)
	}
done:
	if len(words) == 0 {
		p.pos = start
		return 0, false, false, false
	}
	joined := strings.Join(words, " ")
	switch {
	case strings.Contains(joined, "8_t") || strings.Contains(joined, "char"):
		size = 1
	case strings.Contains(joined, "16_t") || strings.Contains(joined, "short"):
		size = 2
	case strings.Contains(joined, "64_t") || strings.Count(joined, "long") == 2:
		size = 8
	case strings.Contains(joined, "long"):
		size = p.d.c.Width() // long is 4 bytes on i386, 8 on x86-64
	default:
		size = 4
	}
	if p.peek() == '*' {
		p.pos++
		pointer = true
	}
	return size, signed, pointer, true
}

func (p *exprParser) primary() (tv, error) {
	switch c := p.peek(); {
	case c == '(':
		save := p.pos
		p.pos++
		if size, signed, ptr, ok := p.cType(); ok && p.peek() == ')' {
			p.pos++
			t, err := p.unary()
			if err != nil {
				return tv{}, err
			}
			switch {
			case ptr:
				return tv{v: t.v, ptr: size}, nil
			case t.lbl: // (int)counter reads the variable
				v, err := p.d.readMem(t.v, size)
				return tv{v: v}, err
			}
			v := t.v
			if size < 8 {
				v &= 1<<(8*uint(size)) - 1
				if signed {
					v = uint64(emuSignExtend(v, size))
				}
			}
			return tv{v: v}, nil
		}
		p.pos = save + 1
		t, err := p.exprTV()
		if err != nil {
			return tv{}, err
		}
		if p.peek() != ')' {
			return tv{}, p.fail()
		}
		p.pos++
		return tv{v: t.v}, nil
	case c == '{': // gdb {int} addr
		p.pos++
		size, _, _, ok := p.cType()
		if !ok || p.peek() != '}' {
			return tv{}, p.fail()
		}
		p.pos++
		t, err := p.unary()
		if err != nil {
			return tv{}, err
		}
		v, err := p.d.readMem(t.v, size)
		return tv{v: v}, err
	case c == '$': // gdb register: $eax, $pc, $sp, $fp, $eflags
		p.pos++
		name := strings.ToLower(p.word())
		switch name {
		case "pc", "eip", "rip":
			return tv{v: p.d.c.RIP}, nil
		case "sp":
			return tv{v: p.d.c.SP()}, nil
		case "fp":
			return tv{v: p.d.c.R[emu.RBP] & widthMask(p.d.c.Width())}, nil
		case "eflags", "rflags", "flags":
			return tv{v: p.d.c.RFLAGS}, nil
		}
		if r, ok := regByName(name); ok {
			return tv{v: p.d.c.Value(r)}, nil
		}
		return tv{}, errors.New(i18n.T("dbg.badExpr", "$"+name))
	case c == '\'' || c == '"':
		if p.pos+2 < len(p.s) && p.s[p.pos+2] == c {
			v := uint64(p.s[p.pos+1])
			p.pos += 3
			return tv{v: v}, nil
		}
		return tv{}, p.fail()
	case c == '[':
		addr, err := p.address()
		if err != nil {
			return tv{}, err
		}
		v, err := p.d.readMem(addr, p.d.c.Width())
		return tv{v: v}, err
	}
	w := p.word()
	if w == "" {
		return tv{}, p.fail()
	}
	if size, ok := sizeWords[strings.ToLower(w)]; ok {
		addr, err := p.address()
		if err != nil {
			return tv{}, err
		}
		v, err := p.d.readMem(addr, size)
		return tv{v: v}, err
	}
	if v, ok := parseNumber(w); ok {
		return tv{v: v}, nil
	}
	if r, ok := regByName(w); ok {
		return tv{v: p.d.c.Value(r)}, nil
	}
	if strings.EqualFold(w, "eip") || strings.EqualFold(w, "rip") {
		return tv{v: p.d.c.RIP}, nil
	}
	if a, ok := p.d.label(w); ok {
		return tv{v: a, lbl: true}, nil
	}
	return tv{}, errors.New(i18n.T("dbg.badExpr", w))
}

func widthMask(w int) uint64 {
	if w >= 8 {
		return ^uint64(0)
	}
	return 1<<(8*uint(w)) - 1
}

func emuSignExtend(v uint64, size int) int64 {
	switch size {
	case 1:
		return int64(int8(v))
	case 2:
		return int64(int16(v))
	case 4:
		return int64(int32(v))
	}
	return int64(v)
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
	if d.p.Lib != nil {
		if a, ok := libc.Symbols(d.p.Bits)[name]; ok { // p stdout, x/gx &stdout
			return a, true
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
