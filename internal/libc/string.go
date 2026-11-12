package libc

import (
	"bytes"
	"strconv"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

func defString() {
	def("strlen", retLong, fixedArgs, func(k *call) uint64 {
		return uint64(len(k.strArg(0)))
	})
	def("strcpy", retLong, fixedArgs, func(k *call) uint64 {
		dst := k.ptrArg(0)
		k.write(dst, append(k.strArg(1), 0))
		return dst
	})
	def("strncpy", retLong, fixedArgs, func(k *call) uint64 {
		dst := k.ptrArg(0)
		n := k.size(2)
		src := k.cstr(k.argPtr(1), int(min(n, 1<<30)))
		k.describe(quote(src))
		k.describe(strconv.FormatUint(n, 10))
		k.check(dst, n, emu.AccessWrite)
		out := make([]byte, n)
		copy(out, src)
		k.write(dst, out)
		return dst
	})
	def("strcat", retLong, fixedArgs, func(k *call) uint64 {
		dst := k.ptrArg(0)
		end := dst + uint64(len(k.cstr(dst, -1)))
		k.write(end, append(k.strArg(1), 0))
		return dst
	})
	def("strncat", retLong, fixedArgs, func(k *call) uint64 {
		dst := k.ptrArg(0)
		n := k.size(2)
		src := k.cstr(k.argPtr(1), int(min(n, 1<<30)))
		k.describe(quote(src))
		k.describe(strconv.FormatUint(n, 10))
		end := dst + uint64(len(k.cstr(dst, -1)))
		k.write(end, append(src, 0))
		return dst
	})
	def("strcmp", retInt, fixedArgs, func(k *call) uint64 {
		return uint64(int64(compare(k.strArg(0), k.strArg(1), -1)))
	})
	def("strncmp", retInt, fixedArgs, func(k *call) uint64 {
		n := k.size(2)
		a := k.cstr(k.argPtr(0), int(min(n, 1<<30)))
		b := k.cstr(k.argPtr(1), int(min(n, 1<<30)))
		k.describe(quote(a))
		k.describe(quote(b))
		k.describe(strconv.FormatUint(n, 10))
		return uint64(int64(compare(a, b, int(min(n, 1<<30)))))
	})
	def("strchr", retLong, fixedArgs, func(k *call) uint64 {
		p := k.ptrArg(0)
		s := k.cstr(p, -1)
		c := byte(k.arg(1))
		k.describe(strconv.QuoteRune(rune(c)))
		if c == 0 {
			return p + uint64(len(s))
		}
		if i := bytes.IndexByte(s, c); i >= 0 {
			return p + uint64(i)
		}
		return 0
	})
	def("strrchr", retLong, fixedArgs, func(k *call) uint64 {
		p := k.ptrArg(0)
		s := k.cstr(p, -1)
		c := byte(k.arg(1))
		k.describe(strconv.QuoteRune(rune(c)))
		if c == 0 {
			return p + uint64(len(s))
		}
		if i := bytes.LastIndexByte(s, c); i >= 0 {
			return p + uint64(i)
		}
		return 0
	})
	def("strstr", retLong, fixedArgs, func(k *call) uint64 {
		p := k.ptrArg(0)
		s := k.cstr(p, -1)
		sub := k.strArg(1)
		if i := bytes.Index(s, sub); i >= 0 {
			return p + uint64(i)
		}
		return 0
	})
	def("strdup", retLong, fixedArgs, func(k *call) uint64 {
		s := k.strArg(0)
		p := k.lib.malloc(k, uint64(len(s)+1))
		if p != 0 {
			k.write(p, append(s, 0))
		}
		return p
	})
	def("memset", retLong, fixedArgs, func(k *call) uint64 {
		p := k.ptrArg(0)
		c := byte(k.arg(1))
		n := k.size(2)
		k.describe(strconv.Itoa(int(c)))
		k.describe(strconv.FormatUint(n, 10))
		k.check(p, n, emu.AccessWrite)
		k.write(p, bytes.Repeat([]byte{c}, int(n)))
		return p
	})
	cpy := func(k *call) uint64 {
		dst, src := k.ptrArg(0), k.ptrArg(1)
		n := k.size(2)
		k.describe(strconv.FormatUint(n, 10))
		k.check(src, n, emu.AccessRead)
		k.check(dst, n, emu.AccessWrite)
		k.write(dst, k.read(src, int(n)))
		return dst
	}
	def("memcpy", retLong, fixedArgs, cpy)
	def("memmove", retLong, fixedArgs, cpy)
	def("memcmp", retInt, fixedArgs, func(k *call) uint64 {
		a, b := k.ptrArg(0), k.ptrArg(1)
		n := k.size(2)
		k.describe(strconv.FormatUint(n, 10))
		k.check(a, n, emu.AccessRead)
		k.check(b, n, emu.AccessRead)
		return uint64(int64(compare(k.read(a, int(n)), k.read(b, int(n)), int(n))))
	})

	// <ctype.h>. Like glibc, a true result is the character class's bit
	// in glibc's table (isdigit('5') is 2048), not 1: test it with
	// test eax, eax / jnz, not cmp eax, 1.
	ctype := func(name string, bit uint64, f func(c int) bool) {
		def(name, retInt, fixedArgs, func(k *call) uint64 {
			c := int(int32(k.arg(0)))
			k.describe(charText(c))
			if c >= 0 && c < 128 && f(c) {
				return bit
			}
			return 0
		})
	}
	alpha := func(c int) bool { return c|0x20 >= 'a' && c|0x20 <= 'z' }
	digit := func(c int) bool { return c >= '0' && c <= '9' }
	ctype("isupper", 0x100, func(c int) bool { return c >= 'A' && c <= 'Z' })
	ctype("islower", 0x200, func(c int) bool { return c >= 'a' && c <= 'z' })
	ctype("isalpha", 0x400, alpha)
	ctype("isdigit", 0x800, digit)
	ctype("isxdigit", 0x1000, func(c int) bool { return isDigit(c, 16) })
	ctype("isspace", 0x2000, isSpace)
	ctype("isprint", 0x4000, func(c int) bool { return c >= ' ' && c < 127 })
	ctype("ispunct", 0x4, func(c int) bool { return c > ' ' && c < 127 && !alpha(c) && !digit(c) })
	ctype("isalnum", 0x8, func(c int) bool { return alpha(c) || digit(c) })
	def("toupper", retInt, fixedArgs, func(k *call) uint64 {
		c := int32(k.arg(0))
		k.describe(charText(int(c)))
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		return uint64(int64(c))
	})
	def("tolower", retInt, fixedArgs, func(k *call) uint64 {
		c := int32(k.arg(0))
		k.describe(charText(int(c)))
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		return uint64(int64(c))
	})
}

func charText(c int) string {
	if c >= ' ' && c < 127 {
		return strconv.QuoteRune(rune(c))
	}
	return strconv.Itoa(c)
}

// compare is strcmp/memcmp: the difference of the first differing bytes
// (as unsigned char), like glibc on x86.
func compare(a, b []byte, n int) int {
	for i := 0; n < 0 || i < n; i++ {
		var x, y int
		if i < len(a) {
			x = int(a[i])
		}
		if i < len(b) {
			y = int(b[i])
		}
		if x != y || x == 0 && n < 0 {
			return x - y
		}
		if i >= len(a) && i >= len(b) {
			return 0
		}
	}
	return 0
}

// ptrArg reads a pointer argument and shows it in --trace.
func (k *call) ptrArg(n int) uint64 {
	p := k.argPtr(n)
	if k.c.Mode == 32 {
		p &= 0xffffffff
	}
	k.describePtr(p)
	return p
}

// check faults (inside the C library, like a real crash) if n bytes at
// addr cannot be accessed. It also bounds sizes before buffers are made.
func (k *call) check(addr, n uint64, acc emu.Access) {
	if n == 0 {
		return
	}
	m := min(n, 1<<31)
	if err := k.c.Mem.Check(addr, int(m), acc); err != nil {
		if mf, ok := err.(*emu.MemFault); ok {
			k.bad(mf.Addr)
		}
		k.bad(addr)
	}
	if n > m {
		k.bad(addr + m)
	}
}
