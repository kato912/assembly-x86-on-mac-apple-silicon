package emu

import "golang.org/x/arch/x86/x86asm"

// Stack frames, register/flag save and restore, counted loops and flag
// control instructions.

func init() {
	register(execEnter, x86asm.ENTER)
	register(execLeave, x86asm.LEAVE)
	register(execPusha, x86asm.PUSHA, x86asm.PUSHAD)
	register(execPopa, x86asm.POPA, x86asm.POPAD)
	register(execPushf, x86asm.PUSHF, x86asm.PUSHFD, x86asm.PUSHFQ)
	register(execPopf, x86asm.POPF, x86asm.POPFD, x86asm.POPFQ)
	register(execLoop, x86asm.LOOP, x86asm.LOOPE, x86asm.LOOPNE)
	register(execJcxz, x86asm.JCXZ, x86asm.JECXZ, x86asm.JRCXZ)
	register(execFlagOp, x86asm.CLC, x86asm.STC, x86asm.CMC, x86asm.CLD, x86asm.STD)
	register(func(*CPU, *x86asm.Inst) error { return &Fault{Kind: FaultInvalid} }, x86asm.UD1, x86asm.UD2)
}

func execEnter(c *CPU, in *x86asm.Inst) error {
	size := uint64(in.Args[0].(x86asm.Imm)) & 0xffff
	level := uint64(in.Args[1].(x86asm.Imm)) & 0x1f
	w := c.stackWidth(in)
	bp := c.R[RBP] & mask(w)
	if err := c.push(bp, w); err != nil {
		return err
	}
	frame := c.SP()
	if level > 0 {
		for i := uint64(1); i < level; i++ {
			bp = c.wrap(bp - uint64(w))
			c.faultStack = true
			v, err := c.Mem.Read(bp, w)
			if err != nil {
				return err
			}
			if err := c.push(v, w); err != nil {
				return err
			}
		}
		if err := c.push(frame, w); err != nil {
			return err
		}
	}
	newSP := c.wrap(c.SP() - size)
	// Touch the new frame's lowest byte so a huge frame faults here, as
	// hardware does, instead of on some later access.
	if size > 0 {
		c.faultStack = true
		if err := c.Mem.Check(newSP, 1, AccessWrite); err != nil {
			return err
		}
		c.faultStack = false
	}
	c.setReg(stackReg(RBP, w), frame)
	c.setSP(newSP)
	return nil
}

// stackReg returns the x86asm register for GPR i at width w bytes.
func stackReg(i, w int) x86asm.Reg {
	switch w {
	case 2:
		return x86asm.AX + x86asm.Reg(i)
	case 4:
		return x86asm.EAX + x86asm.Reg(i)
	}
	return x86asm.RAX + x86asm.Reg(i)
}

func execLeave(c *CPU, in *x86asm.Inst) error {
	w := c.stackWidth(in)
	oldSP := c.R[RSP]
	c.setSP(c.R[RBP])
	v, err := c.pop(w)
	if err != nil {
		c.R[RSP] = oldSP
		return err
	}
	c.setReg(stackReg(RBP, w), v)
	return nil
}

func execPusha(c *CPU, in *x86asm.Inst) error {
	w := 4
	if in.Op == x86asm.PUSHA {
		w = 2
	}
	sp := c.R[RSP]
	for _, r := range []int{RAX, RCX, RDX, RBX, -1, RBP, RSI, RDI} {
		v := sp
		if r >= 0 {
			v = c.R[r]
		}
		if err := c.push(v, w); err != nil {
			c.R[RSP] = sp
			return err
		}
	}
	return nil
}

func execPopa(c *CPU, in *x86asm.Inst) error {
	w := 4
	if in.Op == x86asm.POPA {
		w = 2
	}
	sp := c.R[RSP]
	c.faultStack = true
	if err := c.Mem.Check(c.SP(), 8*w, AccessRead); err != nil {
		return err
	}
	for _, r := range []int{RDI, RSI, RBP, -1, RBX, RDX, RCX, RAX} {
		v, _ := c.pop(w)
		if r >= 0 {
			c.setReg(stackReg(r, w), v)
		}
	}
	_ = sp
	return nil
}

func execPushf(c *CPU, in *x86asm.Inst) error {
	// VM and RF always read as zero in the pushed image.
	return c.push(c.RFLAGS&^(1<<16|1<<17), c.stackWidth(in))
}

func execPopf(c *CPU, in *x86asm.Inst) error {
	w := c.stackWidth(in)
	v, err := c.pop(w)
	if err != nil {
		return err
	}
	m := popfMask & mask(w)
	c.RFLAGS = c.RFLAGS&^m | v&m | reservedOne
	return nil
}

// countReg returns the counter register (cx/ecx/rcx) selected by the
// instruction's address size.
func countReg(in *x86asm.Inst) x86asm.Reg {
	switch in.AddrSize {
	case 16:
		return x86asm.CX
	case 32:
		return x86asm.ECX
	}
	return x86asm.RCX
}

func execLoop(c *CPU, in *x86asm.Inst) error {
	r := countReg(in)
	n := (c.getReg(r) - 1) & mask(regSize(r))
	c.setReg(r, n)
	take := n != 0
	switch in.Op {
	case x86asm.LOOPE:
		take = take && c.flag(ZF)
	case x86asm.LOOPNE:
		take = take && !c.flag(ZF)
	}
	if take {
		return execJmp(c, in)
	}
	return nil
}

func execJcxz(c *CPU, in *x86asm.Inst) error {
	var r x86asm.Reg
	switch in.Op {
	case x86asm.JCXZ:
		r = x86asm.CX
	case x86asm.JECXZ:
		r = x86asm.ECX
	default:
		r = x86asm.RCX
	}
	if c.getReg(r) == 0 {
		return execJmp(c, in)
	}
	return nil
}

func execFlagOp(c *CPU, in *x86asm.Inst) error {
	switch in.Op {
	case x86asm.CLC:
		c.setFlags(0, CF)
	case x86asm.STC:
		c.setFlags(CF, CF)
	case x86asm.CMC:
		c.RFLAGS ^= CF
	case x86asm.CLD:
		c.setFlags(0, DF)
	case x86asm.STD:
		c.setFlags(DF, DF)
	}
	return nil
}
