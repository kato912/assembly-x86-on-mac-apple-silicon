package emu

import "golang.org/x/arch/x86/x86asm"

// Core instructions: data movement, basic arithmetic/logic, stack, control
// flow and the system-call gates.

func init() {
	register(execMov, x86asm.MOV)
	register(execLea, x86asm.LEA)
	register(execALU, x86asm.ADD, x86asm.ADC, x86asm.SUB, x86asm.SBB, x86asm.CMP,
		x86asm.AND, x86asm.OR, x86asm.XOR, x86asm.TEST)
	register(execIncDec, x86asm.INC, x86asm.DEC)
	register(execPush, x86asm.PUSH)
	register(execPop, x86asm.POP)
	register(execJmp, x86asm.JMP)
	register(execJcc, x86asm.JA, x86asm.JAE, x86asm.JB, x86asm.JBE, x86asm.JE, x86asm.JG, x86asm.JGE,
		x86asm.JL, x86asm.JLE, x86asm.JNE, x86asm.JNO, x86asm.JNP, x86asm.JNS, x86asm.JO, x86asm.JP, x86asm.JS)
	register(execCall, x86asm.CALL)
	register(execRet, x86asm.RET)
	register(func(*CPU, *x86asm.Inst) error { return nil }, x86asm.NOP, x86asm.PAUSE)
	register(func(*CPU, *x86asm.Inst) error { return &Exit{Code: 0, ByHalt: true} }, x86asm.HLT)
	register(execInt, x86asm.INT)
	register(execSyscall, x86asm.SYSCALL)
}

func execMov(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	v, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	return c.write(in, in.Args[0], size, v)
}

func execLea(c *CPU, in *x86asm.Inst) error {
	m, ok := in.Args[1].(x86asm.Mem)
	if !ok {
		return &Fault{Kind: FaultInvalid}
	}
	return c.write(in, in.Args[0], argSize(in, in.Args[0]), c.ea(m, in))
}

func execALU(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	b, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	var res, f uint64
	store := true
	switch in.Op {
	case x86asm.ADD:
		res, f = Add(a, b, 0, size)
	case x86asm.ADC:
		res, f = Add(a, b, b2u(c.flag(CF)), size)
	case x86asm.SUB:
		res, f = Sub(a, b, 0, size)
	case x86asm.SBB:
		res, f = Sub(a, b, b2u(c.flag(CF)), size)
	case x86asm.CMP:
		res, f = Sub(a, b, 0, size)
		store = false
	case x86asm.AND:
		res = a & b
		f = Logic(res, size)
	case x86asm.OR:
		res = a | b
		f = Logic(res, size)
	case x86asm.XOR:
		res = a ^ b
		f = Logic(res, size)
	case x86asm.TEST:
		res = a & b
		f = Logic(res, size)
		store = false
	}
	if store {
		if err := c.write(in, in.Args[0], size, res); err != nil {
			return err
		}
	}
	c.setFlags(f, Arith)
	return nil
}

func execIncDec(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	var res, f uint64
	if in.Op == x86asm.INC {
		res, f = Add(a, 1, 0, size)
	} else {
		res, f = Sub(a, 1, 0, size)
	}
	if err := c.write(in, in.Args[0], size, res); err != nil {
		return err
	}
	c.setFlags(f, Arith&^CF) // inc/dec leave CF alone
	return nil
}

func execPush(c *CPU, in *x86asm.Inst) error {
	size := c.stackWidth(in)
	v, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	return c.push(v, size)
}

func execPop(c *CPU, in *x86asm.Inst) error {
	size := c.stackWidth(in)
	v, err := c.pop(size)
	if err != nil {
		return err
	}
	if err := c.write(in, in.Args[0], size, v); err != nil {
		c.setSP(c.SP() - uint64(size))
		return err
	}
	return nil
}

func execJmp(c *CPU, in *x86asm.Inst) error {
	t, err := c.branchTarget(in)
	if err != nil {
		return err
	}
	c.jump(t)
	return nil
}

func execJcc(c *CPU, in *x86asm.Inst) error {
	if c.cond(ccOf[in.Op]) {
		return execJmp(c, in)
	}
	return nil
}

func execCall(c *CPU, in *x86asm.Inst) error {
	t, err := c.branchTarget(in)
	if err != nil {
		return err
	}
	if err := c.push(c.next, c.stackWidth(in)); err != nil {
		return err
	}
	c.jump(t)
	return nil
}

func execRet(c *CPU, in *x86asm.Inst) error {
	t, err := c.pop(c.stackWidth(in))
	if err != nil {
		return err
	}
	if imm, ok := in.Args[0].(x86asm.Imm); ok {
		c.setSP(c.SP() + uint64(imm)&0xffff)
	}
	c.jump(t)
	return nil
}

func execInt(c *CPU, in *x86asm.Inst) error {
	n := int(uint8(in.Args[0].(x86asm.Imm)))
	switch n {
	case 0x80:
		return c.Sys.Syscall(c, true)
	case 3:
		return &Fault{Kind: FaultBreakpoint}
	}
	return &Fault{Kind: FaultBadInterrupt, Int: n}
}

func execSyscall(c *CPU, in *x86asm.Inst) error {
	if c.Mode != 64 {
		return &Fault{Kind: FaultSyscallIn32}
	}
	c.R[RCX] = c.next
	c.R[R11] = c.RFLAGS
	return c.Sys.Syscall(c, false)
}
