package emu

import (
	"fmt"

	"golang.org/x/arch/x86/x86asm"
)

// FaultKind classifies why a program stopped abnormally.
type FaultKind int

const (
	FaultMemory        FaultKind = iota // invalid read or write
	FaultExec                           // control jumped to an address that is not code
	FaultRunOff                         // execution fell off the end of the code
	FaultBadReturn                      // ret popped something that is not a return address
	FaultStackOverflow                  // stack grew past its limit
	FaultDivZero                        // div/idiv by zero
	FaultDivOverflow                    // quotient too large for the destination
	FaultUnsupported                    // valid x86 instruction nasmx86 does not emulate
	FaultInvalid                        // bytes that are not a valid instruction
	FaultSyscallIn32                    // `syscall` used in a 32-bit program
	FaultBadInterrupt                   // int N other than 0x80
	FaultBreakpoint                     // int3
	FaultMaxSteps                       // instruction limit reached (infinite loop?)
	FaultPrivileged                     // kernel-only instruction
	FaultLibc                           // misuse of a C library function (Detail explains)
)

// Fault is a CPU exception or other abnormal stop. It carries raw facts;
// package report turns it into a friendly, translated message.
type Fault struct {
	Kind FaultKind
	PC   uint64       // address of the faulting instruction (or bad target)
	From uint64       // for Exec/RunOff/BadReturn: the instruction that got us there
	Inst *x86asm.Inst // the faulting instruction when decoded
	Mem  *MemFault    // for memory faults
	Arg  *x86asm.Mem  // memory operand that faulted, when known
	Regs []RegValue   // registers used to form the faulting address
	Int  int          // interrupt number for FaultBadInterrupt
	Div  struct {     // for division faults
		Hi, Lo, Divisor uint64
		Size            int // operand size in bytes
		Signed          bool
	}
	Steps    uint64
	Bytes    []byte // raw bytes for FaultInvalid
	Detail   string // message key for FaultLibc
	Args     []any  // its arguments
	Hint     string // optional hint key for FaultLibc
	HintArgs []any
	Sig      int // signal to report instead of the default
}

// RegValue is a register name and its value at the time of a fault.
type RegValue struct {
	Name  string
	Value uint64
	Size  int
}

func (f *Fault) Error() string {
	switch f.Kind {
	case FaultMemory:
		return f.Mem.Error()
	case FaultExec:
		return fmt.Sprintf("jumped to %#x, which is not code", f.PC)
	case FaultRunOff:
		return "ran past the end of the code"
	case FaultBadReturn:
		return fmt.Sprintf("ret jumped to %#x, which is not code", f.PC)
	case FaultStackOverflow:
		return "stack overflow"
	case FaultDivZero:
		return "division by zero"
	case FaultDivOverflow:
		return "division overflow"
	case FaultUnsupported:
		return "unsupported instruction"
	case FaultInvalid:
		return "invalid instruction"
	case FaultSyscallIn32:
		return "syscall in 32-bit mode"
	case FaultBadInterrupt:
		return fmt.Sprintf("int %#x", f.Int)
	case FaultBreakpoint:
		return "int3 breakpoint"
	case FaultMaxSteps:
		return fmt.Sprintf("instruction limit reached (%d)", f.Steps)
	case FaultPrivileged:
		return "privileged instruction"
	case FaultLibc:
		return f.Detail
	}
	return "fault"
}

// Signal returns the Linux signal number a native run would die with.
func (f *Fault) Signal() int {
	if f.Sig != 0 {
		return f.Sig
	}
	switch f.Kind {
	case FaultDivZero, FaultDivOverflow:
		return 8 // SIGFPE
	case FaultUnsupported, FaultInvalid, FaultSyscallIn32:
		return 4 // SIGILL
	case FaultBreakpoint:
		return 5 // SIGTRAP
	case FaultMaxSteps:
		return 0
	}
	return 11 // SIGSEGV
}

// Exit is returned by Step/Run when the program exits normally.
type Exit struct {
	Code   int
	ByHalt bool // stopped by hlt rather than the exit syscall
}

func (e *Exit) Error() string { return fmt.Sprintf("exit %d", e.Code) }
