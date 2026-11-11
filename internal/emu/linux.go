package emu

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Linux errno values (returned negated in rax/eax).
const (
	EPERM   = 1
	ENOENT  = 2
	EIO     = 5
	EBADF   = 9
	ENOMEM  = 12
	EACCES  = 13
	EFAULT  = 14
	EEXIST  = 17
	ENOTDIR = 20
	EISDIR  = 21
	EINVAL  = 22
	EMFILE  = 24
	ENOSPC  = 28
	ESPIPE  = 29
	ENOSYS  = 38
)

// Linux is a minimal Linux kernel personality: process setup and the
// system calls an introductory course needs.
type Linux struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Warn reports likely bugs (bad buffer pointer, unknown syscall...).
	Warn func(key string, args ...any)
	// Trace, when set, receives a one-line description of every system call.
	Trace func(desc string)

	files  map[int]*fdEntry
	warned map[uint64]bool

	brkStart, brkCur uint64
	heap             *Region
}

type fdEntry struct {
	r io.Reader
	w io.Writer
	f *os.File // nil for stdio wrappers
}

// NewLinux returns a kernel wired to the given stdio streams.
func NewLinux(stdin io.Reader, stdout, stderr io.Writer) *Linux {
	l := &Linux{Stdin: stdin, Stdout: stdout, Stderr: stderr}
	l.files = map[int]*fdEntry{
		0: {r: stdin, w: stdout}, // a terminal's fd 0 is writable too
		1: {w: stdout},
		2: {w: stderr},
	}
	return l
}

func (l *Linux) warn(key string, args ...any) {
	if l.Warn != nil {
		l.Warn(key, args...)
	}
}

var sys32 = map[uint64]string{1: "exit", 3: "read", 4: "write", 5: "open", 6: "close",
	8: "creat", 10: "unlink", 13: "time", 19: "lseek", 20: "getpid", 45: "brk", 252: "exit_group", 295: "openat"}
var sys64 = map[uint64]string{0: "read", 1: "write", 2: "open", 3: "close", 8: "lseek",
	12: "brk", 39: "getpid", 60: "exit", 85: "creat", 87: "unlink", 201: "time", 231: "exit_group", 257: "openat"}

// SyscallName returns the Linux name of system call nr for the ABI.
func SyscallName(nr uint64, int80 bool) string {
	if int80 {
		return sys32[nr]
	}
	return sys64[nr]
}

// Syscall implements Syscaller.
func (l *Linux) Syscall(c *CPU, int80 bool) error {
	var nr uint64
	var a [6]uint64
	if int80 {
		nr = c.R[RAX] & 0xffffffff
		for i, r := range []int{RBX, RCX, RDX, RSI, RDI, RBP} {
			a[i] = c.R[r] & 0xffffffff
		}
	} else {
		nr = c.R[RAX]
		for i, r := range []int{RDI, RSI, RDX, R10, R8, R9} {
			a[i] = c.R[r]
		}
	}
	name := SyscallName(nr, int80)
	ret, desc, ok, err := l.dispatch(c, name, int80, a)
	if err != nil {
		if l.Trace != nil && desc != "" {
			l.Trace(desc)
		}
		return err
	}
	if !ok {
		if l.warned == nil {
			l.warned = map[uint64]bool{}
		}
		if !l.warned[nr] || l.Trace != nil {
			l.warned[nr] = true
			l.warn("warn.enosys", nr)
		}
		desc = fmt.Sprintf("syscall %d(%#x, %#x, %#x)", nr, a[0], a[1], a[2])
		if name != "" {
			desc = fmt.Sprintf("%s(%#x, %#x, %#x)", name, a[0], a[1], a[2])
		}
		ret = -ENOSYS
	}
	if l.Trace != nil {
		l.Trace(fmt.Sprintf("%s = %d", desc, ret))
	}
	if int80 && c.Mode == 32 {
		c.R[RAX] = uint64(uint32(ret))
	} else {
		c.R[RAX] = uint64(ret)
	}
	return nil
}

func (l *Linux) dispatch(c *CPU, name string, int80 bool, a [6]uint64) (ret int64, desc string, ok bool, err error) {
	switch name {
	case "exit", "exit_group":
		code := int(a[0] & 0xff)
		return 0, fmt.Sprintf("%s(%d)", name, int32(a[0])), true, &Exit{Code: code}
	case "write":
		ret, desc = l.sysWrite(c, a[0], a[1], a[2])
	case "read":
		ret, desc = l.sysRead(c, a[0], a[1], a[2])
	case "open":
		ret, desc = l.sysOpen(c, a[0], a[1], a[2])
	case "openat":
		if int32(a[0]) != atFDCWD {
			if p, err := cString(c, a[1]); err != nil || len(p) == 0 || p[0] != '/' {
				return -ENOSYS, fmt.Sprintf("openat(%d, %#x, ...)", int32(a[0]), a[1]), false, nil
			}
		}
		ret, desc = l.sysOpen(c, a[1], a[2], a[3])
		desc = "openat(AT_FDCWD, " + strings.TrimPrefix(desc, "open(")
	case "creat":
		ret, desc = l.sysOpen(c, a[0], oWRONLY|oCREAT|oTRUNC, a[1])
		desc = "creat(" + strings.TrimPrefix(desc, "open(")
	case "unlink":
		ret, desc = l.sysUnlink(c, a[0])
	case "close":
		ret, desc = l.sysClose(int(int32(a[0])))
	case "lseek":
		off := int64(a[1])
		if int80 {
			off = int64(int32(a[1]))
		}
		ret, desc = l.sysLseek(int(int32(a[0])), off, int(a[2]))
	case "brk":
		ret = int64(l.sysBrk(c, c.wrap(a[0])))
		desc = fmt.Sprintf("brk(%#x)", a[0])
		if c.Mode == 32 {
			ret = int64(uint32(ret))
		}
	case "getpid":
		ret, desc = int64(os.Getpid()), "getpid()"
	case "time":
		ret, desc = l.sysTime(c, a[0], int80)
	default:
		return 0, "", false, nil
	}
	return ret, desc, true, nil
}

func (l *Linux) sysWrite(c *CPU, fd, buf, count uint64) (int64, string) {
	desc := fmt.Sprintf("write(%d, %#x, %d)", int32(fd), buf, int64(count))
	e := l.files[int(int32(fd))]
	if e == nil || e.w == nil {
		l.warn("warn.badfd", "write", int32(fd))
		return -EBADF, desc
	}
	if int64(count) < 0 {
		return -EINVAL, desc
	}
	if count == 0 {
		return 0, desc
	}
	data, err := c.Mem.ReadBytes(c.wrap(buf), int(count))
	if err != nil {
		l.warn("warn.efault", "write", buf)
		return -EFAULT, desc
	}
	desc = fmt.Sprintf("write(%d, %#x %s, %d)", int32(fd), buf, quote(data, 32), count)
	n, werr := e.w.Write(data)
	if werr != nil && n == 0 {
		return -EIO, desc
	}
	return int64(n), desc
}

// quote renders bytes like strace does: "Hello\n"... truncated to max.
func quote(b []byte, max int) string {
	s := fmt.Sprintf("%q", string(b))
	if len(b) > max {
		s = fmt.Sprintf("%q...", string(b[:max]))
	}
	return s
}

// Process memory layout constants.
const (
	StackTop64   = 0x7ffffffff000
	StackTop32   = 0xfffdd000
	StackSize    = 8 << 20
	stackGuard   = 16 << 20
	StubAddr32   = 0x08047000
	StubAddr64   = 0x3ff000
	auxvPageSize = 6
	auxvEntry    = 9
	auxvRandom   = 25
	auxvExecFn   = 31
)

// SetupStack maps the stack and builds the initial Linux process stack
// (argc, argv, envp, auxv) like the kernel does. It returns the initial
// stack pointer, plus the addresses of argv and envp arrays.
func SetupStack(c *CPU, argv, envp []string, entry uint64) (sp, argvPtr, envpPtr uint64) {
	top := uint64(StackTop64)
	if c.Mode == 32 {
		top = StackTop32
	}
	c.Mem.Map(top-StackSize, StackSize, PermR|PermW, "stack", KindStack)
	c.Mem.Map(top-StackSize-stackGuard, stackGuard, 0, "stack guard", KindStackGuard)
	c.StackLimit = top - StackSize
	w := uint64(c.Width())

	p := top - 16 // keep a little zero padding at the very top
	putStr := func(s string) uint64 {
		p -= uint64(len(s) + 1)
		c.Mem.Poke(p, append([]byte(s), 0))
		return p
	}
	execfn := putStr(argv[0])
	envAddrs := make([]uint64, len(envp))
	for i := len(envp) - 1; i >= 0; i-- {
		envAddrs[i] = putStr(envp[i])
	}
	argAddrs := make([]uint64, len(argv))
	for i := len(argv) - 1; i >= 0; i-- {
		argAddrs[i] = putStr(argv[i])
	}
	p -= 16
	random := p
	c.Mem.Poke(random, []byte("nasmx86-random!!"))
	p &^= 15

	auxv := [][2]uint64{{auxvPageSize, PageSize}, {auxvEntry, entry}, {auxvRandom, random}, {auxvExecFn, execfn}, {0, 0}}
	words := 1 + len(argv) + 1 + len(envp) + 1 + 2*len(auxv)
	sp = (p - uint64(words)*w) &^ 15
	q := sp
	put := func(v uint64) {
		c.Mem.Poke(q, le(v, int(w)))
		q += w
	}
	put(uint64(len(argv)))
	argvPtr = q
	for _, a := range argAddrs {
		put(a)
	}
	put(0)
	envpPtr = q
	for _, e := range envAddrs {
		put(e)
	}
	put(0)
	for _, kv := range auxv {
		put(kv[0])
		put(kv[1])
	}
	c.setSP(sp)
	return sp, argvPtr, envpPtr
}

func le(v uint64, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(v >> (8 * i))
	}
	return b
}

// InstallExitStub maps a tiny routine that calls exit(eax). It is used as
// the return address when the program's entry point is `main`, so that
// `ret` from main exits like a C program does.
func InstallExitStub(c *CPU) uint64 {
	var addr uint64
	var code []byte
	if c.Mode == 32 {
		addr = StubAddr32
		code = []byte{0x89, 0xc3, 0xb8, 0x01, 0, 0, 0, 0xcd, 0x80} // mov ebx,eax; mov eax,1; int 0x80
	} else {
		addr = StubAddr64
		code = []byte{0x89, 0xc7, 0xb8, 0x3c, 0, 0, 0, 0x0f, 0x05} // mov edi,eax; mov eax,60; syscall
	}
	c.Mem.Map(addr, PageSize, PermR|PermX, "exit stub", KindStub)
	c.Mem.Poke(addr, code)
	c.Code = append(c.Code, CodeRange{addr, addr + uint64(len(code))})
	return addr
}
