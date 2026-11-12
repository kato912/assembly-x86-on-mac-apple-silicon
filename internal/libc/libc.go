// Package libc is nasmx86's built-in C library: the functions assembly
// courses call from asm (printf, scanf, puts, malloc, strlen, ...). They
// are written in Go and called with the real x86 calling conventions —
// System V AMD64 in 64-bit mode (arguments in rdi, rsi, rdx, rcx, r8, r9,
// then the stack) and cdecl in 32-bit mode (arguments on the stack) — so
// a program that works here works when linked with gcc against glibc.
package libc

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

// Where the library lives in the program's address space. In 64-bit mode
// it must be within ±2 GB of the program so "call printf" (a 32-bit
// relative call) reaches it.
const (
	Base64 = 0x7f000000
	Base32 = 0xf7e00000

	slotSize   = 16      // bytes per function entry point
	dataOff    = 0x10000 // stdin/stdout/stderr variables and FILE objects
	fileOff    = dataOff + 0x100
	fileSize   = 0x100
	maxFiles   = 64
	regionSize = 0x20000
)

// Signals reported for library crashes.
const (
	sigSEGV = 11
	sigABRT = 6
)

type retKind int

const (
	retInt      retKind = iota // int: in eax (zero-extended into rax)
	retLong                    // long, size_t, pointers: all of rax
	retLongLong                // long long: rax, or edx:eax in 32-bit mode
	retVoid
)

// Kinds of variadic function, for the stack alignment check.
const (
	fixedArgs = iota
	printfLike
	scanfLike
)

type function struct {
	name string
	ret  retKind
	vari int
	impl func(k *call) uint64
}

// functions is the library, in entry-point order. Names are sorted so
// the addresses do not depend on map order.
var functions []function

func def(name string, ret retKind, vari int, impl func(k *call) uint64) {
	functions = append(functions, function{name, ret, vari, impl})
}

func init() {
	defStdio()
	defString()
	defStdlib()
	sort.Slice(functions, func(i, j int) bool { return functions[i].name < functions[j].name })
}

// dataSymbols are the library's variables.
var dataSymbols = []string{"stdin", "stdout", "stderr"}

// Names returns every symbol the library defines.
func Names() []string {
	var n []string
	for _, f := range functions {
		n = append(n, f.name)
	}
	return append(n, dataSymbols...)
}

// Region returns where the library's functions are and their names (for a
// debugger's symbol table).
func Region(bits int) (base, size uint64, funcs []string) {
	for _, f := range functions {
		funcs = append(funcs, f.name)
	}
	return Base(bits), uint64(len(functions) * slotSize), funcs
}

// Base returns the library's load address for a 32- or 64-bit program.
func Base(bits int) uint64 {
	if bits == 32 {
		return Base32
	}
	return Base64
}

// Symbols returns the addresses the linker resolves C library names to.
func Symbols(bits int) map[string]uint64 {
	base := Base(bits)
	m := map[string]uint64{}
	for i, f := range functions {
		m[f.name] = base + uint64(i*slotSize)
	}
	w := uint64(bits / 8)
	for i, n := range dataSymbols {
		m[n] = base + dataOff + uint64(i)*w
	}
	return m
}

// Lib is the library's state for one run of a program.
type Lib struct {
	c   *emu.CPU
	os  *emu.Linux
	sys map[string]uint64

	// Warn reports a likely bug that does not stop the program.
	Warn func(key string, args ...any)
	// ExitStub is the return address of main; exiting from it flushes
	// output like returning from main in C does.
	ExitStub uint64

	streams  map[uint64]*stream
	nextFile int
	heap     heap
	seed     *randState
	warned   map[string]bool
}

// Install maps the library into the program and hooks it to the CPU.
func Install(c *emu.CPU, l *emu.Linux) *Lib {
	base := Base(c.Mode)
	lib := &Lib{c: c, os: l, sys: Symbols(c.Mode), warned: map[string]bool{}}
	code := make([]byte, len(functions)*slotSize)
	for i := range code {
		code[i] = 0xc3 // ret: what a disassembler shows; never executed
	}
	c.Mem.Map(base, dataOff, emu.PermR|emu.PermX, "C library", emu.KindStub)
	c.Mem.Poke(base, code)
	c.Mem.Map(base+dataOff, regionSize-dataOff, emu.PermR|emu.PermW, "C library data", emu.KindStub)
	c.Native = emu.CodeRange{Start: base, End: base + uint64(len(code))}
	c.NativeCall = lib.dispatch
	lib.initStreams()
	l.BeforeExit = lib.beforeExit
	return lib
}

// FuncAt names the library function at addr ("" if none).
func (lib *Lib) FuncAt(addr uint64) string { return FuncAt(lib.c.Mode, addr) }

// FuncAt names the library function whose entry point is addr.
func FuncAt(bits int, addr uint64) string {
	base := Base(bits)
	if addr < base || (addr-base)%slotSize != 0 {
		return ""
	}
	if i := int((addr - base) / slotSize); i < len(functions) {
		return functions[i].name
	}
	return ""
}

func (lib *Lib) warn(key string, args ...any) {
	if lib.Warn != nil {
		lib.Warn(key, args...)
	}
}

// call is one call of a library function.
type call struct {
	lib  *Lib
	c    *emu.CPU
	fn   *function
	next int      // next variadic argument slot
	args []string // argument descriptions for --trace
}

// libFault is panicked by helpers and turned into the returned error.
type libFault struct{ err error }

func (k *call) fail(f *emu.Fault) {
	if f.Kind == 0 {
		f.Kind = emu.FaultLibc
	}
	if f.Sig == 0 {
		f.Sig = sigSEGV
	}
	panic(libFault{f})
}

// bad reports a pointer argument that is not valid memory.
func (k *call) bad(addr uint64) {
	hint, args := "hint.libcPtr", []any{}
	switch {
	case addr == 0:
		hint = "hint.libcNull"
	case addr < 0x10000 || k.c.Mem.FindRegion(addr) == nil && addr < 0x100000000:
		hint = "hint.libcValue"
	}
	k.fail(&emu.Fault{Detail: "err.libcPtr", Args: []any{k.fn.name, k.hex(addr)}, Hint: hint, HintArgs: args})
}

func (k *call) hex(a uint64) string {
	if k.c.Mode == 32 || a <= 0xffffffff {
		return fmt.Sprintf("0x%08x", a)
	}
	return fmt.Sprintf("0x%x", a)
}

func (lib *Lib) dispatch(c *emu.CPU, pc uint64) (err error) {
	i := int((pc - c.Native.Start) / slotSize)
	if (pc-c.Native.Start)%slotSize != 0 || i >= len(functions) {
		return &emu.Fault{Kind: emu.FaultExec, PC: pc}
	}
	fn := &functions[i]
	k := &call{lib: lib, c: c, fn: fn}
	c.NativeRan = fn.name
	defer func() {
		if r := recover(); r != nil {
			lf, ok := r.(libFault)
			if !ok {
				panic(r)
			}
			err = lf.err
			_, exited := err.(*emu.Exit)
			lib.trace(k, "", !exited)
		}
	}()
	lib.checkStack(k)
	ret := fn.impl(k)
	lib.setReturn(fn, ret)
	lib.trace(k, lib.retText(fn, ret), false)
	if err := c.Return(); err != nil {
		var mf *emu.MemFault
		if errors.As(err, &mf) {
			return &emu.Fault{Kind: emu.FaultStackOverflow, Mem: mf}
		}
		return err
	}
	return nil
}

func (lib *Lib) retText(fn *function, ret uint64) string {
	switch fn.ret {
	case retVoid:
		return ""
	case retInt:
		return fmt.Sprint(int32(ret))
	case retLongLong:
		return fmt.Sprint(int64(ret))
	}
	if lib.c.Mode == 32 {
		ret &= 0xffffffff
	}
	if ret < 0x10000 {
		return fmt.Sprint(ret)
	}
	return fmt.Sprintf("%#x", ret)
}

// trace shows the call in --trace / the debugger like ltrace does:
// printf("x = %d\n", 5) = 6.
func (lib *Lib) trace(k *call, ret string, crashed bool) {
	if lib.os.Trace == nil {
		return
	}
	s := k.fn.name + "(" + strings.Join(k.args, ", ") + ")"
	switch {
	case crashed:
		s += " → crash"
	case ret != "":
		s += " = " + ret
	}
	lib.os.Trace(s)
}

// checkStack enforces the x86-64 ABI rule that rsp is a multiple of 16
// at a call (so it ends in 8 at the function's first instruction). glibc's
// printf and scanf crash with a segmentation fault when it is broken
// (printf only when al, the number of vector registers used, is not 0).
func (lib *Lib) checkStack(k *call) {
	c := lib.c
	if c.Mode != 64 || c.SP()%16 == 8 {
		return
	}
	sp := c.SP() + 8 // rsp at the call instruction
	crash := k.fn.vari == scanfLike || (k.fn.vari == printfLike && c.R[emu.RAX]&0xff != 0)
	if crash {
		k.fail(&emu.Fault{Detail: "err.libcAlign", Args: []any{k.fn.name, k.hex(sp)}, Hint: "hint.libcAlign"})
	}
	if !lib.warned["align"] {
		lib.warned["align"] = true
		lib.warn("warn.libcAlign", k.fn.name, k.hex(sp))
	}
}

// setReturn puts the result in rax/eax and clobbers the other registers
// the calling convention lets a C function change, as real glibc does, so
// a program that keeps a value in rcx across "call printf" fails here too.
func (lib *Lib) setReturn(fn *function, ret uint64) {
	c := lib.c
	if c.Mode == 32 {
		if fn.ret != retVoid {
			c.R[emu.RAX] = ret & 0xffffffff
		}
		c.R[emu.RCX] = 0
		c.R[emu.RDX] = 0xffffffff & (Base32 + dataOff)
		if fn.ret == retLongLong {
			c.R[emu.RDX] = ret >> 32
		}
		return
	}
	switch fn.ret {
	case retInt:
		c.R[emu.RAX] = ret & 0xffffffff
	case retLong, retLongLong:
		c.R[emu.RAX] = ret
	}
	c.R[emu.RCX] = c.Native.Start + 0x1234 // like the address of glibc's syscall
	c.R[emu.RDX] = 0
	c.R[emu.RSI] = Base64 + dataOff + 0x800
	c.R[emu.RDI] = 0
	c.R[emu.R8] = 0
	c.R[emu.R9] = 0x7c
	c.R[emu.R10] = 0
	c.R[emu.R11] = 0x246 // like the flags saved by syscall
}

// --- argument access ---------------------------------------------------

// arg returns fixed argument n (0-based).
func (k *call) arg(n int) uint64 {
	v, err := k.c.Arg(n)
	if err != nil {
		k.fail(&emu.Fault{Kind: emu.FaultStackOverflow})
	}
	return v
}

// argInt is an int argument.
func (k *call) argInt(n int) int64 { return int64(int32(k.arg(n))) }

// argLong is a long (or size_t when cast to unsigned) argument.
func (k *call) argLong(n int) int64 {
	if k.c.Mode == 32 {
		return int64(int32(k.arg(n)))
	}
	return int64(k.arg(n))
}

// argPtr is a pointer argument.
func (k *call) argPtr(n int) uint64 { return k.arg(n) }

// startVarargs makes the variadic arguments begin after fixed fixed ones.
func (k *call) startVarargs(fixed int) { k.next = fixed }

// vaWord returns the next variadic integer argument (one register or
// stack slot).
func (k *call) vaWord() uint64 {
	v := k.arg(k.next)
	k.next++
	return v
}

// vaInt64 returns a 64-bit variadic argument (two slots in 32-bit mode).
func (k *call) vaInt64() uint64 {
	if k.c.Mode == 64 {
		return k.vaWord()
	}
	lo := k.vaWord()
	return lo | k.vaWord()<<32
}

// --- memory access -----------------------------------------------------

func (k *call) read(addr uint64, n int) []byte {
	if n == 0 {
		return nil
	}
	b, err := k.c.Mem.ReadBytes(addr, n)
	if err != nil {
		k.bad(badAddr(err, addr))
	}
	return b
}

func (k *call) write(addr uint64, b []byte) {
	if len(b) == 0 {
		return
	}
	if err := k.c.Mem.WriteBytes(addr, b); err != nil {
		k.bad(badAddr(err, addr))
	}
	if k.c.Mem.OnWrite != nil {
		if len(b) <= 8 {
			var v uint64
			for i := len(b) - 1; i >= 0; i-- {
				v = v<<8 | uint64(b[i])
			}
			k.c.Mem.OnWrite(addr, len(b), v)
		} else {
			k.c.Mem.OnWrite(addr, len(b), 0)
		}
	}
}

func (k *call) writeN(addr uint64, size int, v uint64) {
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(v >> (8 * i))
	}
	k.write(addr, b)
}

func (k *call) readN(addr uint64, size int) uint64 {
	b := k.read(addr, size)
	var v uint64
	for i := size - 1; i >= 0; i-- {
		v = v<<8 | uint64(b[i])
	}
	return v
}

func badAddr(err error, addr uint64) uint64 {
	var mf *emu.MemFault
	if errors.As(err, &mf) {
		if mf.Addr == addr || mf.Region == nil {
			return addr
		}
	}
	return addr
}

// cstr reads the NUL-terminated string at addr (at most max bytes when
// max >= 0, which need not be terminated, as with printf's %.5s).
func (k *call) cstr(addr uint64, max int) []byte {
	var out []byte
	for {
		n := 256
		if max >= 0 && max-len(out) < n {
			n = max - len(out)
		}
		if n == 0 {
			return out
		}
		chunk := k.c.Mem.Peek(addr+uint64(len(out)), n)
		if i := indexByte(chunk, 0); i >= 0 {
			return append(out, chunk[:i]...)
		}
		out = append(out, chunk...)
		if len(chunk) < n {
			if len(out) == 0 {
				k.bad(addr)
			}
			k.bad(addr + uint64(len(out)))
		}
	}
}

func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

// quote renders a string for --trace like ltrace: "Hello\n"...
func quote(b []byte) string {
	const max = 32
	if len(b) > max {
		return fmt.Sprintf("%q...", string(b[:max]))
	}
	return fmt.Sprintf("%q", string(b))
}

// describe adds an argument description for --trace.
func (k *call) describe(s string) {
	if k.lib.os.Trace != nil {
		k.args = append(k.args, s)
	}
}

func (k *call) describeStr(addr uint64) {
	if k.lib.os.Trace != nil {
		b := k.c.Mem.Peek(addr, 64)
		if i := indexByte(b, 0); i >= 0 {
			b = b[:i]
		}
		k.args = append(k.args, quote(b))
	}
}

func (k *call) describePtr(addr uint64) {
	if k.lib.os.Trace != nil {
		if addr == 0 {
			k.args = append(k.args, "NULL")
		} else {
			k.args = append(k.args, k.hex(addr))
		}
	}
}

// Summary lists the library's most used functions, for messages.
func Summary() string {
	return "printf, puts, putchar, scanf, getchar, fgets, malloc, free, strlen, strcpy, strcmp, atoi, rand, exit (and more: see the README)"
}
