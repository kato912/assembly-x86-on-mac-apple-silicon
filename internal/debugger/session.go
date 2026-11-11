// Package debugger implements --trace and the --step interactive debugger.
package debugger

import (
	"fmt"
	"strings"

	"golang.org/x/arch/x86/x86asm"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// maxShownWrites limits how many memory writes one trace line lists.
const maxShownWrites = 3

type memWrite struct {
	addr uint64
	size int
	val  uint64
}

// Frame is one active function call, tracked from executed call/ret
// instructions (a "shadow" call stack).
type Frame struct {
	Call   uint64 // address of the call instruction
	Return uint64 // where it will return to
	Target uint64 // the function that was called
	SP     uint64 // stack pointer right after the call (return address slot)
}

// Session wraps a program and records what each executed instruction did.
type Session struct {
	P *runner.Program
	C *emu.CPU

	frames []Frame

	sys    []string
	writes []memWrite
	nwrite int
}

// NewSession hooks the program's CPU so changes can be reported.
func NewSession(p *runner.Program) *Session {
	s := &Session{P: p, C: p.CPU}
	p.Linux.Trace = func(desc string) { s.sys = append(s.sys, desc) }
	p.CPU.Mem.OnWrite = func(addr uint64, size int, val uint64) {
		s.nwrite++
		if len(s.writes) < maxShownWrites {
			s.writes = append(s.writes, memWrite{addr, size, val})
		}
	}
	return s
}

type snapshot struct {
	r     [16]uint64
	flags uint64
}

func (s *Session) snap() snapshot { return snapshot{s.C.R, s.C.RFLAGS} }

// StepResult describes one executed instruction.
type StepResult struct {
	PC      uint64
	Where   string // "line 12: mov eax, [ebx]"
	Changes string // "eax=0x00000005 (5)  ZF=1"
	Err     error  // non-nil if the program exited or faulted
	Ran     bool   // the instruction executed (false when it faulted before running)
}

// Step executes one instruction and describes its effects.
func (s *Session) Step() StepResult {
	before := s.snap()
	pc := s.C.RIP
	var op x86asm.Op
	var length int
	if in, derr := s.C.Decode(pc); derr == nil {
		op, length = in.Op, in.Len
	}
	s.sys, s.writes, s.nwrite = s.sys[:0], s.writes[:0], 0
	err := s.C.Step()
	_, exited := err.(*emu.Exit)
	// A faulting instruction did not complete; the fault is reported instead.
	res := StepResult{PC: pc, Err: err, Ran: err == nil || exited}
	if err == nil {
		switch op {
		case x86asm.CALL:
			s.frames = append(s.frames, Frame{Call: pc, Return: pc + uint64(length), Target: s.C.RIP, SP: s.C.SP()})
		case x86asm.RET:
			for len(s.frames) > 0 && s.frames[len(s.frames)-1].SP < s.C.SP() {
				s.frames = s.frames[:len(s.frames)-1]
			}
		}
	}
	if res.Ran {
		res.Where = s.P.Where(pc)
		res.Changes = s.changes(before)
	}
	return res
}

// Frames returns the active calls, outermost first.
func (s *Session) Frames() []Frame { return s.frames }

// changes lists what the last instruction changed.
func (s *Session) changes(b snapshot) string {
	var parts []string
	for _, d := range s.sys {
		parts = append(parts, "→ "+d)
	}
	n := 8
	if s.P.Bits == 64 {
		n = 16
	}
	for _, i := range []int{emu.RAX, emu.RBX, emu.RCX, emu.RDX, emu.RSI, emu.RDI, emu.RBP, emu.RSP,
		emu.R8, emu.R9, emu.R10, emu.R11, emu.R12, emu.R13, emu.R14, emu.R15}[:n] {
		if s.C.R[i] != b.r[i] {
			parts = append(parts, emu.RegName(i, s.P.Bits)+"="+s.P.ValueText(i, s.C.R[i]))
		}
	}
	var fl []string
	for _, f := range emu.FlagNames {
		if (s.C.RFLAGS^b.flags)&f.Bit != 0 {
			v := 0
			if s.C.RFLAGS&f.Bit != 0 {
				v = 1
			}
			fl = append(fl, fmt.Sprintf("%s=%d", f.Name, v))
		}
	}
	if len(fl) > 0 {
		parts = append(parts, strings.Join(fl, " "))
	}
	for _, w := range s.writes {
		if w.size > 8 { // a block written by a system call such as read
			parts = append(parts, fmt.Sprintf("[%s..]=%d bytes", s.P.Addr(w.addr), w.size))
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s]=%s", s.P.Addr(w.addr), runner.Hex(w.val, w.size)))
	}
	if s.nwrite > len(s.writes) {
		parts = append(parts, fmt.Sprintf("(+%d more writes)", s.nwrite-len(s.writes)))
	}
	return strings.Join(parts, "  ")
}

// Format renders a step as one trace line.
func (r StepResult) Format(n uint64) string {
	num := fmt.Sprintf("[%d]", n)
	if r.Changes == "" {
		return fmt.Sprintf("%-6s %s", num, r.Where)
	}
	return fmt.Sprintf("%-6s %-38s %s", num, r.Where, r.Changes)
}
