package libc

import (
	"math"
	"strconv"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

// spec is one parsed % conversion.
type spec struct {
	minus, plus, space, hash, zero bool
	width, prec                    int // -1 = not given
	length                         string
	conv                           byte
}

// format implements printf's format string with the variadic arguments
// that follow the fixed ones. It returns the formatted bytes.
func (k *call) format(f []byte) []byte {
	var out []byte
	for i := 0; i < len(f); i++ {
		if f[i] != '%' {
			out = append(out, f[i])
			continue
		}
		start := i
		i++
		var sp spec
		sp.width, sp.prec = -1, -1
	flags:
		for ; i < len(f); i++ {
			switch f[i] {
			case '-':
				sp.minus = true
			case '+':
				sp.plus = true
			case ' ':
				sp.space = true
			case '#':
				sp.hash = true
			case '0':
				sp.zero = true
			case '\'', 'I':
			default:
				break flags
			}
		}
		if i < len(f) && f[i] == '*' {
			w := int32(k.vaWord())
			k.describe(strconv.Itoa(int(w)))
			if w < 0 {
				sp.minus, w = true, -w
			}
			sp.width = int(w)
			i++
		} else {
			for ; i < len(f) && f[i] >= '0' && f[i] <= '9'; i++ {
				sp.width = max(sp.width, 0)*10 + int(f[i]-'0')
			}
		}
		if i < len(f) && f[i] == '.' {
			i++
			sp.prec = 0
			if i < len(f) && f[i] == '*' {
				p := int32(k.vaWord())
				k.describe(strconv.Itoa(int(p)))
				sp.prec = int(p)
				if p < 0 {
					sp.prec = -1
				}
				i++
			} else {
				for ; i < len(f) && f[i] >= '0' && f[i] <= '9'; i++ {
					sp.prec = sp.prec*10 + int(f[i]-'0')
				}
			}
		}
		for i < len(f) && strings.IndexByte("hlLqjzt", f[i]) >= 0 {
			sp.length += string(f[i])
			i++
		}
		if i >= len(f) {
			out = append(out, f[start:]...)
			break
		}
		sp.conv = f[i]
		switch sp.conv {
		case '%':
			out = append(out, '%')
		case 'd', 'i':
			v := k.intArg(sp.length, true)
			k.describe(strconv.FormatInt(int64(v), 10))
			out = append(out, fmtInt(sp, v, true)...)
		case 'u', 'o', 'x', 'X':
			v := k.intArg(sp.length, false)
			if sp.conv == 'u' {
				k.describe(strconv.FormatUint(v, 10))
			} else {
				k.describe("0x" + strconv.FormatUint(v, 16))
			}
			out = append(out, fmtInt(sp, v, false)...)
		case 'c':
			v := byte(k.vaWord())
			k.describe(strconv.QuoteRune(rune(v)))
			out = append(out, pad(sp, []byte{v}, false)...)
		case 's':
			addr := k.vaWord()
			if k.c.Mode == 32 {
				addr &= 0xffffffff
			}
			var s []byte
			if addr == 0 {
				s = []byte("(null)")
				if sp.prec >= 0 && sp.prec < 6 {
					s = nil
				}
				k.describe("NULL")
			} else {
				s = k.cstr(addr, sp.prec)
				k.describe(quote(s))
			}
			out = append(out, pad(sp, s, false)...)
		case 'p':
			v := k.vaWord()
			if k.c.Mode == 32 {
				v &= 0xffffffff
			}
			k.describePtr(v)
			if v == 0 {
				sp.zero = false
				out = append(out, pad(sp, []byte("(nil)"), false)...)
				break
			}
			sp.hash, sp.length = true, "l"
			sp.conv = 'x'
			out = append(out, fmtInt(sp, v, false)...)
		case 'n':
			addr := k.vaWord()
			k.describePtr(addr)
			k.writeN(addr, k.intSize(sp.length), uint64(len(out)))
		case 'f', 'F', 'e', 'E', 'g', 'G', 'a', 'A':
			v := k.floatArg(sp)
			k.describe(strconv.FormatFloat(v, 'g', -1, 64))
			out = append(out, fmtFloat(sp, v)...)
		default:
			// glibc prints an unknown conversion as it is.
			out = append(out, f[start:i+1]...)
		}
	}
	return out
}

// intSize is the size of an integer with the given length modifier.
func (k *call) intSize(length string) int {
	switch length {
	case "hh":
		return 1
	case "h":
		return 2
	case "ll", "q", "j", "L":
		return 8
	case "l", "z", "t":
		return k.c.Width()
	}
	return 4
}

// intArg fetches an integer argument and narrows it to its C type.
func (k *call) intArg(length string, signed bool) uint64 {
	size := k.intSize(length)
	var v uint64
	if size == 8 {
		v = k.vaInt64()
	} else {
		v = k.vaWord()
	}
	switch size {
	case 1:
		if signed {
			return uint64(int64(int8(v)))
		}
		return v & 0xff
	case 2:
		if signed {
			return uint64(int64(int16(v)))
		}
		return v & 0xffff
	case 4:
		if signed {
			return uint64(int64(int32(v)))
		}
		return v & 0xffffffff
	}
	return v
}

// floatArg fetches a double. In 32-bit mode doubles are on the stack; in
// 64-bit mode they are passed in xmm registers, which nasmx86 does not
// emulate.
func (k *call) floatArg(sp spec) float64 {
	if k.c.Mode == 64 {
		k.fail(&emu.Fault{Detail: "err.libcFloat", Args: []any{k.fn.name, "%" + sp.length + string(sp.conv)}, Hint: "hint.libcFloat", Sig: sigSEGV})
	}
	if sp.length == "L" { // long double: 12 bytes on the stack
		lo := k.vaWord() | k.vaWord()<<32
		hi := k.vaWord() & 0xffff
		return x87ToFloat(hi, lo)
	}
	return math.Float64frombits(k.vaInt64())
}

func x87ToFloat(se, mant uint64) float64 {
	sign := se >> 15
	exp := int(se & 0x7fff)
	var v float64
	switch {
	case exp == 0 && mant == 0:
		v = 0
	case exp == 0x7fff:
		if mant<<1 == 0 {
			v = math.Inf(1)
		} else {
			v = math.NaN()
		}
	default:
		v = math.Ldexp(float64(mant), exp-16383-63)
	}
	if sign != 0 {
		v = -v
	}
	return v
}

// fmtInt formats an integer conversion (d i u o x X) like C.
func fmtInt(sp spec, v uint64, signed bool) []byte {
	neg := false
	if signed && int64(v) < 0 {
		neg, v = true, -v
	}
	base := 10
	switch sp.conv {
	case 'o':
		base = 8
	case 'x', 'X':
		base = 16
	}
	digits := strconv.FormatUint(v, base)
	if sp.conv == 'X' {
		digits = strings.ToUpper(digits)
	}
	if sp.prec == 0 && v == 0 {
		digits = ""
	}
	if sp.prec > len(digits) {
		digits = strings.Repeat("0", sp.prec-len(digits)) + digits
	}
	var prefix string
	switch {
	case neg:
		prefix = "-"
	case signed && sp.plus:
		prefix = "+"
	case signed && sp.space:
		prefix = " "
	}
	if sp.hash {
		switch {
		case sp.conv == 'o' && !strings.HasPrefix(digits, "0"):
			digits = "0" + digits
		case sp.conv == 'x' && v != 0:
			prefix += "0x"
		case sp.conv == 'X' && v != 0:
			prefix += "0X"
		}
	}
	zero := sp.zero && !sp.minus && sp.prec < 0
	if zero && sp.width > len(prefix)+len(digits) {
		digits = strings.Repeat("0", sp.width-len(prefix)-len(digits)) + digits
	}
	return pad(sp, []byte(prefix+digits), false)
}

// fmtFloat formats f F e E g G a A like glibc.
func fmtFloat(sp spec, v float64) []byte {
	upper := sp.conv >= 'A' && sp.conv <= 'Z'
	neg := math.Signbit(v)
	if neg {
		v = -v
	}
	var prefix string
	switch {
	case neg:
		prefix = "-"
	case sp.plus:
		prefix = "+"
	case sp.space:
		prefix = " "
	}
	var body string
	special := false
	switch {
	case math.IsInf(v, 0):
		body, special = "inf", true
	case math.IsNaN(v):
		body, special = "nan", true
	default:
		prec := sp.prec
		if prec < 0 {
			prec = 6
		}
		switch sp.conv | 0x20 {
		case 'f':
			body = strconv.FormatFloat(v, 'f', prec, 64)
			if sp.hash && prec == 0 {
				body += "."
			}
		case 'e':
			body = fmtExp(v, prec, sp.hash)
		case 'g':
			if prec == 0 {
				prec = 1
			}
			e := strconv.FormatFloat(v, 'e', prec-1, 64)
			x, _ := strconv.Atoi(e[strings.IndexByte(e, 'e')+1:])
			if prec > x && x >= -4 {
				body = strconv.FormatFloat(v, 'f', prec-1-x, 64)
			} else {
				body = fmtExp(v, prec-1, sp.hash)
			}
			if !sp.hash {
				body = trimZeros(body)
			} else if !strings.Contains(body, ".") {
				if i := strings.IndexByte(body, 'e'); i >= 0 {
					body = body[:i] + "." + body[i:]
				} else {
					body += "."
				}
			}
		case 'a':
			p := -1
			if sp.prec >= 0 {
				p = sp.prec
			}
			body = strconv.FormatFloat(v, 'x', p, 64)
			// Go writes p+01; C writes p+1.
			if i := strings.IndexByte(body, 'p'); i >= 0 {
				exp := body[i+2:]
				for len(exp) > 1 && exp[0] == '0' {
					exp = exp[1:]
				}
				body = body[:i+2] + exp
			}
		}
	}
	if upper {
		body = strings.ToUpper(body)
	}
	zero := sp.zero && !sp.minus && !special
	if zero && sp.width > len(prefix)+len(body) {
		ins := strings.Repeat("0", sp.width-len(prefix)-len(body))
		if strings.HasPrefix(body, "0x") || strings.HasPrefix(body, "0X") {
			body = body[:2] + ins + body[2:]
		} else {
			body = ins + body
		}
	}
	return pad(sp, []byte(prefix+body), false)
}

// fmtExp is %e: d.ddde±dd.
func fmtExp(v float64, prec int, hash bool) string {
	s := strconv.FormatFloat(v, 'e', prec, 64)
	if hash && prec == 0 {
		i := strings.IndexByte(s, 'e')
		s = s[:i] + "." + s[i:]
	}
	return s
}

// trimZeros removes trailing zeros after the decimal point (%g).
func trimZeros(s string) string {
	exp := ""
	if i := strings.IndexByte(s, 'e'); i >= 0 {
		s, exp = s[:i], s[i:]
	}
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s + exp
}

// pad applies the field width.
func pad(sp spec, b []byte, _ bool) []byte {
	if sp.width <= len(b) {
		return b
	}
	fill := strings.Repeat(" ", sp.width-len(b))
	if sp.minus {
		return append(b, fill...)
	}
	return append([]byte(fill), b...)
}
