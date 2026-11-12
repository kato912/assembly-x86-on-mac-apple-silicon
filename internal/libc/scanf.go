package libc

import (
	"math"
	"strconv"
	"strings"
)

// source is where scanf reads from: a stream (stdin, a file) or a string
// (sscanf).
type source interface {
	peek() int // next byte, -1 at end of input
	get() int
	unget(c byte)
}

type streamSource struct {
	lib *Lib
	s   *stream
}

func (r streamSource) peek() int     { return r.lib.peekc(r.s) }
func (r streamSource) get() int      { return r.lib.getc(r.s) }
func (r streamSource) unget(c byte)  { r.lib.ungetc(r.s, c) }
func (r *stringSource) peek() int    { return r.at(0) }
func (r *stringSource) unget(_ byte) { r.pos-- }
func (r *stringSource) at(i int) int {
	if r.pos+i < len(r.b) {
		return int(r.b[r.pos+i])
	}
	return -1
}
func (r *stringSource) get() int {
	c := r.at(0)
	if c >= 0 {
		r.pos++
	}
	return c
}

type stringSource struct {
	b   []byte
	pos int
}

func isSpace(c int) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r'
}

// counter wraps a source to count consumed bytes (for %n).
type counter struct {
	src source
	n   int
}

func (r *counter) peek() int { return r.src.peek() }
func (r *counter) get() int {
	c := r.src.get()
	if c >= 0 {
		r.n++
	}
	return c
}
func (r *counter) unget(c byte) { r.src.unget(c); r.n-- }

func (r *counter) skipSpace() {
	for isSpace(r.peek()) {
		r.get()
	}
}

// scan implements scanf's format string: it returns the number of
// assigned items, or -1 (EOF) if the input ended before the first
// conversion.
func (k *call) scan(src source, f []byte) int64 {
	r := &counter{src: src}
	assigned := 0
	converted := false
	eofFail := func() int64 {
		if !converted {
			return -1
		}
		return int64(assigned)
	}
	for i := 0; i < len(f); i++ {
		ch := f[i]
		if isSpace(int(ch)) {
			r.skipSpace()
			continue
		}
		if ch != '%' {
			c := r.peek()
			if c < 0 {
				return eofFail()
			}
			if byte(c) != ch {
				return int64(assigned)
			}
			r.get()
			continue
		}
		i++
		if i >= len(f) {
			break
		}
		suppress := false
		if f[i] == '*' {
			suppress = true
			i++
		}
		width := -1
		for ; i < len(f) && f[i] >= '0' && f[i] <= '9'; i++ {
			width = max(width, 0)*10 + int(f[i]-'0')
		}
		length := ""
		for i < len(f) && strings.IndexByte("hlLqjzt", f[i]) >= 0 {
			length += string(f[i])
			i++
		}
		if i >= len(f) {
			break
		}
		conv := f[i]
		if conv != 'c' && conv != '[' && conv != 'n' {
			r.skipSpace()
		}
		switch conv {
		case '%':
			c := r.peek()
			if c < 0 {
				return eofFail()
			}
			if c != '%' {
				return int64(assigned)
			}
			r.get()
			continue
		case 'n':
			if !suppress {
				addr := k.vaWord()
				k.describePtr(addr)
				k.writeN(k.ptr(addr), k.intSize(length), uint64(r.n))
			}
			continue
		}
		if r.peek() < 0 {
			return eofFail()
		}
		switch conv {
		case 'd', 'i', 'u', 'o', 'x', 'X', 'p':
			base := map[byte]int{'d': 10, 'u': 10, 'i': 0, 'o': 8, 'x': 16, 'X': 16, 'p': 16}[conv]
			text, ok := scanInt(r, width, base)
			if !ok {
				return int64(assigned)
			}
			converted = true
			if suppress {
				continue
			}
			size := k.intSize(length)
			if conv == 'p' {
				size = k.c.Width()
			}
			v := k.convertInt(text, base, conv == 'd' || conv == 'i', size)
			addr := k.vaWord()
			k.describePtr(addr)
			k.writeN(k.ptr(addr), size, v)
			assigned++
		case 'f', 'F', 'e', 'E', 'g', 'G', 'a', 'A':
			text := scanFloat(r, width)
			v, err := strconv.ParseFloat(text, 64)
			if text == "" || err != nil && !isRangeErr(err) {
				return int64(assigned)
			}
			converted = true
			if suppress {
				continue
			}
			addr := k.vaWord()
			k.describePtr(addr)
			switch length {
			case "l":
				k.writeN(k.ptr(addr), 8, math.Float64bits(v))
			case "L":
				k.write(k.ptr(addr), floatToX87(v))
			default:
				k.writeN(k.ptr(addr), 4, uint64(math.Float32bits(float32(v))))
			}
			assigned++
		case 's':
			var b []byte
			for width != 0 {
				c := r.peek()
				if c < 0 || isSpace(c) {
					break
				}
				b = append(b, byte(r.get()))
				width--
			}
			converted = true
			if !suppress {
				addr := k.vaWord()
				k.describePtr(addr)
				k.write(k.ptr(addr), append(b, 0))
				assigned++
			}
		case 'c':
			if width < 0 {
				width = 1
			}
			var b []byte
			for len(b) < width {
				c := r.get()
				if c < 0 {
					break
				}
				b = append(b, byte(c))
			}
			if len(b) == 0 {
				return eofFail()
			}
			converted = true
			if !suppress {
				addr := k.vaWord()
				k.describePtr(addr)
				k.write(k.ptr(addr), b)
				assigned++
			}
		case '[':
			set, end := parseScanset(f, i+1)
			i = end
			var b []byte
			for width != 0 {
				c := r.peek()
				if c < 0 || !set[c] {
					break
				}
				b = append(b, byte(r.get()))
				width--
			}
			if len(b) == 0 {
				return int64(assigned)
			}
			converted = true
			if !suppress {
				addr := k.vaWord()
				k.describePtr(addr)
				k.write(k.ptr(addr), append(b, 0))
				assigned++
			}
		default:
			return int64(assigned)
		}
	}
	return int64(assigned)
}

// ptr checks a scanf destination: passing a value ([x]) instead of an
// address (x) is the classic mistake, reported by k.write's fault.
func (k *call) ptr(addr uint64) uint64 {
	if k.c.Mode == 32 {
		return addr & 0xffffffff
	}
	return addr
}

func isRangeErr(err error) bool {
	ne, ok := err.(*strconv.NumError)
	return ok && ne.Err == strconv.ErrRange
}

// scanInt reads the longest prefix that can be an integer in base (0 =
// C's auto-detect). Like glibc, it cannot push back more than one byte, so
// "-x" consumes the '-' and fails.
func scanInt(r *counter, width, base int) (string, bool) {
	var b []byte
	more := func() bool { return width < 0 || len(b) < width }
	if c := r.peek(); (c == '+' || c == '-') && more() {
		b = append(b, byte(r.get()))
	}
	digits := 0
	if (base == 0 || base == 16) && r.peek() == '0' && more() {
		b = append(b, byte(r.get()))
		digits++
		if c := r.peek(); (c == 'x' || c == 'X') && more() {
			r.get()
			b = append(b, 'x')
			digits = 0
			base = 16
			if !isDigit(r.peek(), 16) {
				// "0x" followed by a non-digit: the 0 is the number.
				return strings.TrimSuffix(string(b), "x"), true
			}
		} else if base == 0 {
			base = 8
		}
	}
	if base == 0 {
		base = 10
	}
	for more() && isDigit(r.peek(), base) {
		b = append(b, byte(r.get()))
		digits++
	}
	return string(b), digits > 0
}

func isDigit(c, base int) bool {
	switch {
	case c >= '0' && c <= '9':
		return c-'0' < base
	case c >= 'a' && c <= 'z':
		return c-'a'+10 < base
	case c >= 'A' && c <= 'Z':
		return c-'A'+10 < base
	}
	return false
}

// convertInt converts like strtol/strtoul (saturating at the limits of
// long) and then truncates to size bytes, as glibc's scanf does.
func (k *call) convertInt(text string, base int, signed bool, size int) uint64 {
	neg := strings.HasPrefix(text, "-")
	t := strings.TrimLeft(text, "+-")
	if base == 0 {
		switch {
		case strings.HasPrefix(t, "0x"):
			base = 16
		case strings.HasPrefix(t, "0") && len(t) > 1:
			base = 8
		default:
			base = 10
		}
	}
	t = strings.TrimPrefix(t, "0x")
	long64 := k.c.Mode == 64 || size == 8
	u, err := strconv.ParseUint(t, base, 64)
	overflow := err != nil
	var v uint64
	if signed {
		lim := uint64(math.MaxInt64)
		if !long64 {
			lim = math.MaxInt32
		}
		switch {
		case !neg && (overflow || u > lim):
			v = lim
		case neg && (overflow || u > lim+1):
			v = -(lim + 1)
		case neg:
			v = -u
		default:
			v = u
		}
	} else {
		lim := uint64(math.MaxUint64)
		if !long64 {
			lim = math.MaxUint32
		}
		switch {
		case overflow || u > lim:
			v = lim
		case neg:
			v = -u
		default:
			v = u
		}
	}
	return v
}

// scanFloat reads the longest prefix that can be a floating-point number.
func scanFloat(r *counter, width int) string {
	var b []byte
	more := func() bool { return width < 0 || len(b) < width }
	if c := r.peek(); (c == '+' || c == '-') && more() {
		b = append(b, byte(r.get()))
	}
	// inf, infinity, nan
	if c := r.peek(); c == 'i' || c == 'I' || c == 'n' || c == 'N' {
		var word string
		for more() {
			c := r.peek()
			if c < 0 {
				break
			}
			w := word + strings.ToLower(string(rune(c)))
			if !strings.HasPrefix("infinity", w) && !strings.HasPrefix("nan", w) {
				break
			}
			r.get()
			word = w
		}
		switch {
		case word == "nan":
			return string(b) + "nan"
		case strings.HasPrefix(word, "inf"):
			return string(b) + "inf"
		}
		return ""
	}
	digits := 0
	for more() && isDigit(r.peek(), 10) {
		b = append(b, byte(r.get()))
		digits++
	}
	if more() && r.peek() == '.' {
		b = append(b, byte(r.get()))
		for more() && isDigit(r.peek(), 10) {
			b = append(b, byte(r.get()))
			digits++
		}
	}
	if digits == 0 {
		return ""
	}
	if c := r.peek(); (c == 'e' || c == 'E') && more() {
		b = append(b, byte(r.get()))
		if c := r.peek(); (c == '+' || c == '-') && more() {
			b = append(b, byte(r.get()))
		}
		n := 0
		for more() && isDigit(r.peek(), 10) {
			b = append(b, byte(r.get()))
			n++
		}
		if n == 0 { // "1e" — glibc keeps the consumed 'e' but the number is 1
			b = []byte(strings.TrimRight(string(b), "eE+-"))
		}
	}
	return string(b)
}

// floatToX87 converts to an 80-bit long double (10 bytes, as stored by
// fstp tword).
func floatToX87(v float64) []byte {
	out := make([]byte, 10)
	bits := math.Float64bits(v)
	sign := bits >> 63
	exp := int((bits >> 52) & 0x7ff)
	frac := bits & (1<<52 - 1)
	var e uint64
	var m uint64
	switch {
	case exp == 0 && frac == 0:
	case exp == 0x7ff:
		e, m = 0x7fff, 1<<63|frac<<11
	case exp == 0: // subnormal double: normal as long double
		fr, ex := math.Frexp(math.Abs(v))
		m = uint64(fr * (1 << 64))
		e = uint64(ex - 1 + 16383)
	default:
		e, m = uint64(exp-1023+16383), 1<<63|frac<<11
	}
	for i := 0; i < 8; i++ {
		out[i] = byte(m >> (8 * i))
	}
	se := e | sign<<15
	out[8], out[9] = byte(se), byte(se>>8)
	return out
}

// parseScanset parses %[...] starting just after '['; it returns the set
// and the index of the closing ']'.
func parseScanset(f []byte, i int) (set [256]bool, end int) {
	neg := false
	if i < len(f) && f[i] == '^' {
		neg = true
		i++
	}
	first := true
	for ; i < len(f); i++ {
		c := f[i]
		if c == ']' && !first {
			break
		}
		first = false
		if i+2 < len(f) && f[i+1] == '-' && f[i+2] != ']' && f[i+2] >= c {
			for x := int(c); x <= int(f[i+2]); x++ {
				set[x] = true
			}
			i += 2
			continue
		}
		set[c] = true
	}
	if neg {
		for x := range set {
			set[x] = !set[x]
		}
	}
	return set, i
}
