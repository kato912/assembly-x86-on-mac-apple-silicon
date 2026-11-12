package libc

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

func defStdlib() {
	def("exit", retVoid, fixedArgs, func(k *call) uint64 {
		code := int(k.argInt(0))
		k.describe(strconv.Itoa(code))
		k.lib.FlushAll()
		k.lib.closeFiles()
		panic(libFault{&emu.Exit{Code: code & 0xff}})
	})
	def("_exit", retVoid, fixedArgs, func(k *call) uint64 {
		code := int(k.argInt(0))
		k.describe(strconv.Itoa(code))
		k.lib.Discard()
		panic(libFault{&emu.Exit{Code: code & 0xff}})
	})
	def("abort", retVoid, fixedArgs, func(k *call) uint64 {
		k.fail(&emu.Fault{Detail: "err.libcAbort", Sig: sigABRT})
		return 0
	})

	def("atoi", retInt, fixedArgs, func(k *call) uint64 {
		v, _ := k.strtol(k.strArg(0), 10, true)
		return v
	})
	atol := func(k *call) uint64 {
		v, _ := k.strtol(k.strArg(0), 10, true)
		return v
	}
	def("atol", retLong, fixedArgs, atol)
	def("atoll", retLongLong, fixedArgs, func(k *call) uint64 {
		v, _ := strtol(k.strArg(0), 10, true, 64)
		return v
	})
	strto := func(signed bool) func(k *call) uint64 {
		return func(k *call) uint64 {
			p := k.argPtr(0)
			if k.c.Mode == 32 {
				p &= 0xffffffff
			}
			s := k.cstr(p, -1)
			k.describe(quote(s))
			end := k.argPtr(1)
			k.describePtr(end)
			base := int(k.argInt(2))
			k.describe(strconv.Itoa(base))
			v, n := k.strtol(s, base, signed)
			if end != 0 {
				k.writeN(end, k.c.Width(), p+uint64(n))
			}
			return v
		}
	}
	def("strtol", retLong, fixedArgs, strto(true))
	def("strtoul", retLong, fixedArgs, strto(false))
	def("abs", retInt, fixedArgs, func(k *call) uint64 {
		v := int32(k.arg(0))
		k.describe(strconv.Itoa(int(v)))
		if v < 0 {
			v = -v
		}
		return uint64(int64(v))
	})
	def("labs", retLong, fixedArgs, func(k *call) uint64 {
		v := k.argLong(0)
		k.describe(strconv.FormatInt(v, 10))
		if v < 0 {
			v = -v
		}
		return uint64(v)
	})

	def("malloc", retLong, fixedArgs, func(k *call) uint64 {
		n := k.size(0)
		k.describe(strconv.FormatUint(n, 10))
		return k.lib.malloc(k, n)
	})
	def("calloc", retLong, fixedArgs, func(k *call) uint64 {
		n, size := k.size(0), k.size(1)
		k.describe(strconv.FormatUint(n, 10))
		k.describe(strconv.FormatUint(size, 10))
		hi, total := mul64(n, size)
		if hi != 0 || (k.c.Mode == 32 && total > math.MaxUint32) {
			return 0
		}
		p := k.lib.malloc(k, total)
		if p != 0 && total > 0 {
			k.c.Mem.Poke(p, make([]byte, total))
		}
		return p
	})
	def("realloc", retLong, fixedArgs, func(k *call) uint64 {
		p := k.ptrArg(0)
		n := k.size(1)
		k.describe(strconv.FormatUint(n, 10))
		return k.lib.realloc(k, p, n)
	})
	def("free", retVoid, fixedArgs, func(k *call) uint64 {
		k.lib.free(k, k.ptrArg(0))
		return 0
	})

	def("rand", retInt, fixedArgs, func(k *call) uint64 {
		if k.lib.seed == nil {
			k.lib.seed = newRand(1)
		}
		return uint64(k.lib.seed.next())
	})
	def("srand", retVoid, fixedArgs, func(k *call) uint64 {
		s := uint32(k.arg(0))
		k.describe(strconv.FormatUint(uint64(s), 10))
		k.lib.seed = newRand(s)
		return 0
	})
	def("time", retLong, fixedArgs, func(k *call) uint64 {
		p := k.ptrArg(0)
		t := uint64(time.Now().Unix())
		if k.c.Mode == 32 {
			t &= 0xffffffff
		}
		if p != 0 {
			k.writeN(p, k.c.Width(), t)
		}
		return t
	})
	def("sleep", retInt, fixedArgs, func(k *call) uint64 {
		s := uint32(k.arg(0))
		k.describe(strconv.FormatUint(uint64(s), 10))
		time.Sleep(time.Duration(s) * time.Second)
		return 0
	})
	def("usleep", retInt, fixedArgs, func(k *call) uint64 {
		us := uint32(k.arg(0))
		k.describe(strconv.FormatUint(uint64(us), 10))
		time.Sleep(time.Duration(us) * time.Microsecond)
		return 0
	})
}

func (lib *Lib) closeFiles() {
	for _, s := range lib.sortedStreams() {
		if s.fd > 2 {
			lib.closeFile(s)
		}
	}
}

func mul64(a, b uint64) (hi, lo uint64) {
	if a == 0 || b == 0 {
		return 0, 0
	}
	lo = a * b
	if lo/b != a {
		return 1, lo
	}
	return 0, lo
}

// strtol is C's strtol/strtoul for long (the CPU's word size): it skips
// leading space, reads a sign and digits, and saturates on overflow. It
// returns the value and how many bytes it consumed (0 if no number).
func (k *call) strtol(s []byte, base int, signed bool) (uint64, int) {
	bits := 64
	if k.c.Mode == 32 {
		bits = 32
	}
	return strtol(s, base, signed, bits)
}

func strtol(s []byte, base int, signed bool, bits int) (uint64, int) {
	i := 0
	for i < len(s) && isSpace(int(s[i])) {
		i++
	}
	neg := false
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		neg = s[i] == '-'
		i++
	}
	if (base == 0 || base == 16) && i+1 < len(s) && s[i] == '0' && (s[i+1]|0x20) == 'x' && i+2 < len(s) && isDigit(int(s[i+2]), 16) {
		i += 2
		base = 16
	} else if base == 0 {
		base = 10
		if i < len(s) && s[i] == '0' {
			base = 8
		}
	}
	if base < 2 || base > 36 {
		return 0, 0
	}
	start := i
	for i < len(s) && isDigit(int(s[i]), base) {
		i++
	}
	if i == start {
		return 0, 0
	}
	u, err := strconv.ParseUint(strings.ToLower(string(s[start:i])), base, 64)
	over := err != nil
	var v uint64
	if signed {
		lim := uint64(1)<<(bits-1) - 1
		switch {
		case !neg && (over || u > lim):
			v = lim
		case neg && (over || u > lim+1):
			v = -(lim + 1)
		case neg:
			v = -u
		default:
			v = u
		}
	} else {
		lim := ^uint64(0) >> (64 - bits)
		switch {
		case over || u > lim:
			v = lim
		case neg:
			v = -u
		default:
			v = u
		}
	}
	if bits == 32 {
		v &= 0xffffffff
	}
	return v, i
}

// randState is glibc's random(): an additive feedback generator, so
// srand(1) then rand() gives the same numbers as on Linux (1804289383...).
type randState struct {
	r [34]int32
	i int
}

func newRand(seed uint32) *randState {
	if seed == 0 {
		seed = 1
	}
	all := make([]int32, 344)
	all[0] = int32(seed)
	for i := 1; i < 31; i++ {
		hi := int64(all[i-1]) / 127773
		lo := int64(all[i-1]) % 127773
		word := 16807*lo - 2836*hi
		if word < 0 {
			word += 2147483647
		}
		all[i] = int32(word)
	}
	for i := 31; i < 34; i++ {
		all[i] = all[i-31]
	}
	for i := 34; i < 344; i++ {
		all[i] = all[i-31] + all[i-3]
	}
	s := &randState{}
	copy(s.r[:], all[344-34:])
	return s
}

func (s *randState) next() int32 {
	// r[n] = r[n-31] + r[n-3], kept in a ring of the last 34 values.
	n := len(s.r)
	v := s.r[(s.i+n-31)%n] + s.r[(s.i+n-3)%n]
	s.r[s.i] = v
	s.i = (s.i + 1) % n
	return int32(uint32(v) >> 1)
}

// heap is malloc's state. Chunks are laid out like glibc's (a 16-byte
// header in 64-bit mode, sizes rounded to 16) in memory obtained with brk,
// so addresses look like the ones a real program prints.
type heap struct {
	top, end uint64
	used     map[uint64]uint64 // user pointer -> chunk size
	freed    map[uint64]uint64
	bins     map[uint64][]uint64 // chunk size -> freed pointers, newest last
}

func (lib *Lib) chunkSize(n uint64) uint64 {
	w := uint64(lib.c.Width())
	s := (n + w + 15) &^ 15
	return max(s, 4*w)
}

func (lib *Lib) malloc(k *call, n uint64) uint64 {
	h := &lib.heap
	if h.used == nil {
		h.used, h.freed, h.bins = map[uint64]uint64{}, map[uint64]uint64{}, map[uint64][]uint64{}
	}
	limit := uint64(1) << 40
	if lib.c.Mode == 32 {
		limit = 1 << 31
	}
	if n >= limit {
		return 0
	}
	size := lib.chunkSize(n)
	if b := h.bins[size]; len(b) > 0 {
		p := b[len(b)-1]
		h.bins[size] = b[:len(b)-1]
		delete(h.freed, p)
		h.used[p] = size
		return p
	}
	hdr := 2 * uint64(lib.c.Width())
	if h.top == 0 || h.top+size+hdr > h.end {
		cur := lib.os.Brk(lib.c, 0)
		// Chunks start hdr bytes before a 16-byte aligned user pointer.
		switch {
		case h.top == 0:
			h.top = (cur+0x290+hdr+15)&^15 - hdr // where glibc's first chunk starts
		case cur != h.end:
			// The program moved the break itself: continue after it.
			h.top = (cur+hdr+15)&^15 - hdr
		}
		want := (h.top + size + hdr + 0x21000 + 0xfff) &^ 0xfff
		if got := lib.os.Brk(lib.c, want); got < want {
			return 0
		}
		h.end = want
	}
	chunk := h.top
	h.top += size
	p := chunk + hdr
	// The size field glibc keeps just before the pointer (with the
	// "previous chunk in use" bit).
	w := lib.c.Width()
	sz := make([]byte, w)
	for i := range sz {
		sz[i] = byte((size | 1) >> (8 * i))
	}
	lib.c.Mem.Poke(p-uint64(w), sz)
	h.used[p] = size
	return p
}

func (lib *Lib) free(k *call, p uint64) {
	h := &lib.heap
	if p == 0 {
		return
	}
	size, ok := h.used[p]
	if !ok {
		if _, dbl := h.freed[p]; dbl {
			k.fail(&emu.Fault{Detail: "err.libcDoubleFree", Args: []any{k.hex(p)}, Hint: "hint.libcDoubleFree", Sig: sigABRT})
		}
		k.fail(&emu.Fault{Detail: "err.libcBadFree", Args: []any{k.fn.name, k.hex(p)}, Hint: "hint.libcBadFree", Sig: sigABRT})
	}
	delete(h.used, p)
	h.freed[p] = size
	h.bins[size] = append(h.bins[size], p)
}

func (lib *Lib) realloc(k *call, p, n uint64) uint64 {
	if p == 0 {
		return lib.malloc(k, n)
	}
	h := &lib.heap
	size, ok := h.used[p]
	if !ok {
		k.fail(&emu.Fault{Detail: "err.libcBadFree", Args: []any{k.fn.name, k.hex(p)}, Hint: "hint.libcBadFree", Sig: sigABRT})
	}
	if n == 0 {
		lib.free(k, p)
		return 0
	}
	if lib.chunkSize(n) <= size {
		return p
	}
	q := lib.malloc(k, n)
	if q == 0 {
		return 0
	}
	usable := size - uint64(lib.c.Width())
	lib.c.Mem.Poke(q, lib.c.Mem.Peek(p, int(min(usable, n))))
	lib.free(k, p)
	return q
}
