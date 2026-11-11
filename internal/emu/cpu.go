// Package emu implements an x86 (IA-32 and x86-64) user-mode CPU emulator
// for the instruction subset taught in introductory assembly courses, plus
// a small Linux system-call layer.
package emu

import (
	"errors"
	"strings"

	"golang.org/x/arch/x86/x86asm"
)

// Syscaller handles `int 0x80` and `syscall`.
type Syscaller interface {
	// Syscall performs the system call described by the CPU registers.
	// int80 is true for the 32-bit `int 0x80` ABI.
	Syscall(c *CPU, int80 bool) error
}

// CodeRange is a range of addresses holding instructions.
type CodeRange struct{ Start, End uint64 }

type handler func(c *CPU, in *x86asm.Inst) error

var handlers = map[x86asm.Op]handler{}

func register(h handler, ops ...x86asm.Op) {
	for _, op := range ops {
		handlers[op] = h
	}
}

type decoded struct {
	inst x86asm.Inst
	h    handler
}

// CPU is the emulated processor state.
type CPU struct {
	Mode   int // 32 or 64
	R      [16]uint64
	RIP    uint64
	RFLAGS uint64

	Mem      *Memory
	Sys      Syscaller
	Code     []CodeRange // where instructions may be fetched from
	Steps    uint64      // instructions executed so far
	MaxSteps uint64      // 0 = unlimited

	// StackLimit is the lowest valid stack address; accesses just below it
	// are reported as a stack overflow.
	StackLimit uint64

	pc, next uint64 // current instruction address and fall-through address
	branched bool   // current instruction transferred control

	havePrev     bool
	prevPC       uint64
	prevLen      int
	prevOp       x86asm.Op
	prevBranched bool

	faultMem   *x86asm.Mem // memory operand being accessed (for error messages)
	faultStack bool        // implicit stack access in progress

	cache map[uint64]*decoded
}

// New returns a CPU in the given mode (32 or 64) using mem.
func New(mode int, mem *Memory, sys Syscaller) *CPU {
	return &CPU{Mode: mode, Mem: mem, Sys: sys, RFLAGS: InitialFlags, cache: make(map[uint64]*decoded)}
}

// Width returns the natural register width in bytes (4 or 8).
func (c *CPU) Width() int { return c.Mode / 8 }

func (c *CPU) wrap(a uint64) uint64 {
	if c.Mode == 32 {
		return a & 0xffffffff
	}
	return a
}

// SP returns the stack pointer.
func (c *CPU) SP() uint64 { return c.wrap(c.R[RSP]) }

func (c *CPU) setSP(v uint64) { c.R[RSP] = c.wrap(v) }

// InCode reports whether addr lies in a code range.
func (c *CPU) InCode(addr uint64) bool {
	for _, r := range c.Code {
		if addr >= r.Start && addr < r.End {
			return true
		}
	}
	return false
}

// Decode returns the instruction at addr without executing it.
func (c *CPU) Decode(addr uint64) (*x86asm.Inst, error) {
	d, err := c.fetch(addr)
	if err != nil {
		return nil, err
	}
	return &d.inst, nil
}

func (c *CPU) fetch(pc uint64) (*decoded, error) {
	if d, ok := c.cache[pc]; ok {
		return d, nil
	}
	if !c.InCode(pc) {
		return nil, c.badPC(pc)
	}
	var buf [15]byte
	n, err := c.Mem.Fetch(pc, buf[:])
	if err != nil {
		return nil, c.badPC(pc)
	}
	inst, err := x86asm.Decode(buf[:n], c.Mode)
	if err != nil {
		if c.Mode == 32 && n >= 2 && buf[0] == 0x0f && buf[1] == 0x05 {
			// SYSCALL does not exist in 32-bit mode; say so instead of "invalid".
			return nil, &Fault{Kind: FaultSyscallIn32, PC: pc}
		}
		return nil, &Fault{Kind: FaultInvalid, PC: pc, Bytes: append([]byte(nil), buf[:n]...)}
	}
	h := handlers[inst.Op]
	if h == nil {
		h = unsupported
	}
	d := &decoded{inst: inst, h: h}
	c.cache[pc] = d
	return d, nil
}

func (c *CPU) badPC(pc uint64) *Fault {
	f := &Fault{Kind: FaultExec, PC: pc, From: c.prevPC}
	if c.havePrev {
		switch {
		case !c.prevBranched && c.prevPC+uint64(c.prevLen) == pc:
			f.Kind = FaultRunOff
		case c.prevOp == x86asm.RET:
			f.Kind = FaultBadReturn
		}
	}
	return f
}

// Step executes one instruction.
func (c *CPU) Step() error {
	if c.MaxSteps > 0 && c.Steps >= c.MaxSteps {
		return &Fault{Kind: FaultMaxSteps, PC: c.RIP, Steps: c.Steps}
	}
	pc := c.RIP
	d, err := c.fetch(pc)
	if err != nil {
		return err
	}
	c.pc = pc
	c.next = c.wrap(pc + uint64(d.inst.Len))
	c.RIP = c.next
	c.branched = false
	c.faultMem, c.faultStack = nil, false
	err = d.h(c, &d.inst)
	c.Steps++
	if err != nil {
		var ex *Exit
		if errors.As(err, &ex) {
			return err
		}
		c.RIP = pc
		return c.decorate(err, d)
	}
	c.prevPC, c.prevLen, c.prevOp, c.prevBranched, c.havePrev = pc, d.inst.Len, d.inst.Op, c.branched, true
	return nil
}

// Run executes until the program exits or faults.
func (c *CPU) Run() error {
	for {
		if err := c.Step(); err != nil {
			return err
		}
	}
}

// LastBranched reports whether the most recently executed instruction
// transferred control (taken jump, call, ret...).
func (c *CPU) LastBranched() bool { return c.prevBranched }

func (c *CPU) decorate(err error, d *decoded) error {
	var mf *MemFault
	if errors.As(err, &mf) {
		f := &Fault{Kind: FaultMemory, PC: c.pc, Inst: &d.inst, Mem: mf}
		if (mf.Region != nil && mf.Region.Kind == KindStackGuard) ||
			(c.faultStack && c.StackLimit != 0 && mf.Addr < c.StackLimit && mf.Addr >= c.StackLimit-64<<20) {
			f.Kind = FaultStackOverflow
		}
		if c.faultMem != nil {
			f.Arg = c.faultMem
			for _, r := range []x86asm.Reg{c.faultMem.Base, c.faultMem.Index} {
				if isGPR(r) {
					f.Regs = append(f.Regs, RegValue{Name: nasmRegName(r), Value: c.getReg(r), Size: regSize(r)})
				}
			}
		} else if c.faultStack {
			f.Regs = append(f.Regs, RegValue{Name: RegName(RSP, c.Mode), Value: c.SP(), Size: c.Width()})
		}
		return f
	}
	var f *Fault
	if errors.As(err, &f) {
		if f.Inst == nil {
			f.Inst = &d.inst
		}
		if f.PC == 0 {
			f.PC = c.pc
		}
		return f
	}
	return err
}

// Disasm renders an instruction in Intel syntax with NASM-style register
// names, e.g. "mov eax, dword ptr [ebx+0x4]".
func Disasm(in *x86asm.Inst, pc uint64, sym func(uint64) (string, uint64)) string {
	s := x86asm.IntelSyntax(*in, pc, sym)
	// x86asm prints REX byte registers as spl/bpl/... already; fix r8l -> r8d.
	for i := 8; i <= 15; i++ {
		s = strings.ReplaceAll(s, "r"+itoa(i)+"l", "r"+itoa(i)+"d")
	}
	return s
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return "1" + string(rune('0'+i-10))
}

// ---- operand helpers ----

func (c *CPU) ea(m x86asm.Mem, in *x86asm.Inst) uint64 {
	var a uint64
	switch m.Base {
	case 0:
	case x86asm.RIP, x86asm.EIP, x86asm.IP:
		a = c.next
	default:
		a = c.getReg(m.Base)
	}
	if m.Index != 0 {
		a += c.getReg(m.Index) * uint64(m.Scale)
	}
	a += uint64(m.Disp)
	switch in.AddrSize {
	case 16:
		a &= 0xffff
	case 32:
		a &= 0xffffffff
	}
	return a
}

// argSize returns the operand size in bytes of an instruction argument.
func argSize(in *x86asm.Inst, a x86asm.Arg) int {
	switch a := a.(type) {
	case x86asm.Reg:
		return regSize(a)
	case x86asm.Mem:
		if in.MemBytes > 0 {
			return in.MemBytes
		}
	}
	return in.DataSize / 8
}

func (c *CPU) read(in *x86asm.Inst, a x86asm.Arg, size int) (uint64, error) {
	switch a := a.(type) {
	case x86asm.Reg:
		if !isGPR(a) {
			return 0, &Fault{Kind: FaultUnsupported}
		}
		return c.getReg(a), nil
	case x86asm.Mem:
		m := a
		c.faultMem = &m
		return c.Mem.Read(c.ea(a, in), size)
	case x86asm.Imm:
		return uint64(a) & mask(size), nil
	case x86asm.Rel:
		return uint64(int64(a)), nil
	}
	return 0, &Fault{Kind: FaultUnsupported}
}

func (c *CPU) write(in *x86asm.Inst, a x86asm.Arg, size int, v uint64) error {
	switch a := a.(type) {
	case x86asm.Reg:
		if !isGPR(a) {
			return &Fault{Kind: FaultUnsupported}
		}
		c.setReg(a, v)
		return nil
	case x86asm.Mem:
		m := a
		c.faultMem = &m
		return c.Mem.Write(c.ea(a, in), size, v&mask(size))
	}
	return &Fault{Kind: FaultUnsupported}
}

// stackWidth is the push/pop size in bytes for an instruction.
func (c *CPU) stackWidth(in *x86asm.Inst) int {
	if in.DataSize == 16 {
		return 2
	}
	return c.Width()
}

func (c *CPU) push(v uint64, size int) error {
	sp := c.wrap(c.SP() - uint64(size))
	c.faultMem, c.faultStack = nil, true
	if err := c.Mem.Write(sp, size, v&mask(size)); err != nil {
		return err
	}
	c.faultStack = false
	c.setSP(sp)
	return nil
}

func (c *CPU) pop(size int) (uint64, error) {
	sp := c.SP()
	c.faultMem, c.faultStack = nil, true
	v, err := c.Mem.Read(sp, size)
	if err != nil {
		return 0, err
	}
	c.faultStack = false
	c.setSP(sp + uint64(size))
	return v, nil
}

func (c *CPU) jump(target uint64) {
	c.RIP = c.wrap(target)
	c.branched = true
}

// branchTarget resolves the target of jmp/call/jcc.
func (c *CPU) branchTarget(in *x86asm.Inst) (uint64, error) {
	switch a := in.Args[0].(type) {
	case x86asm.Rel:
		return c.wrap(c.next + uint64(int64(a))), nil
	case x86asm.Imm:
		return uint64(a), nil
	}
	size := c.Width()
	if in.DataSize == 16 {
		size = 2
	}
	if r, ok := in.Args[0].(x86asm.Reg); ok {
		size = regSize(r)
	}
	return c.read(in, in.Args[0], size)
}

func unsupported(c *CPU, in *x86asm.Inst) error {
	switch in.Op {
	case x86asm.IN, x86asm.OUT, x86asm.INSB, x86asm.INSW, x86asm.INSD, x86asm.OUTSB, x86asm.OUTSW,
		x86asm.OUTSD, x86asm.CLI, x86asm.STI, x86asm.LGDT, x86asm.LIDT, x86asm.LLDT, x86asm.LTR,
		x86asm.WRMSR, x86asm.RDMSR, x86asm.INVLPG, x86asm.CLTS, x86asm.SYSRET, x86asm.SYSEXIT, x86asm.IRET, x86asm.IRETD, x86asm.IRETQ:
		return &Fault{Kind: FaultPrivileged}
	}
	for _, a := range in.Args {
		if r, ok := a.(x86asm.Reg); ok && r >= x86asm.CR0 {
			return &Fault{Kind: FaultPrivileged}
		}
	}
	return &Fault{Kind: FaultUnsupported}
}

// ccOf maps Jcc/SETcc/CMOVcc opcodes to their 4-bit condition code. It is
// built during package variable initialization, before any init() runs.
var ccOf = buildCondTable()

func buildCondTable() map[x86asm.Op]int {
	t := map[x86asm.Op]int{}
	codes := map[string]int{"O": 0, "NO": 1, "B": 2, "AE": 3, "E": 4, "NE": 5, "BE": 6, "A": 7,
		"S": 8, "NS": 9, "P": 10, "NP": 11, "L": 12, "GE": 13, "LE": 14, "G": 15}
	for op := x86asm.Op(1); op < 4096; op++ {
		name := op.String()
		if strings.HasPrefix(name, "Op(") {
			continue
		}
		for _, p := range []string{"J", "SET", "CMOV"} {
			if strings.HasPrefix(name, p) {
				if cc, ok := codes[name[len(p):]]; ok {
					t[op] = cc
				}
			}
		}
	}
	return t
}
