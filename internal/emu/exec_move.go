package emu

import "golang.org/x/arch/x86/x86asm"

// Data movement beyond plain mov: zero/sign extension, exchanges,
// accumulator sign extension, table lookup and flag byte transfers.

func init() {
	register(execMovx, x86asm.MOVZX, x86asm.MOVSX, x86asm.MOVSXD)
	register(execXchg, x86asm.XCHG)
	register(execXadd, x86asm.XADD)
	register(execCmpxchg, x86asm.CMPXCHG)
	register(execSignExtendAcc, x86asm.CBW, x86asm.CWDE, x86asm.CDQE, x86asm.CWD, x86asm.CDQ, x86asm.CQO)
	register(execXlat, x86asm.XLATB)
	register(execLahf, x86asm.LAHF)
	register(execSahf, x86asm.SAHF)
}

func execMovx(c *CPU, in *x86asm.Inst) error {
	dsize := argSize(in, in.Args[0])
	ssize := argSize(in, in.Args[1])
	if in.Op == x86asm.MOVSXD {
		ssize = 4
	}
	v, err := c.read(in, in.Args[1], ssize)
	if err != nil {
		return err
	}
	if in.Op != x86asm.MOVZX {
		v = uint64(signExtend(v, ssize))
	}
	return c.write(in, in.Args[0], dsize, v&mask(dsize))
}

func execXchg(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	b, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	// Write the memory operand (if any) first so a fault leaves registers intact.
	first, second := 0, 1
	va, vb := b, a
	if _, ok := in.Args[1].(x86asm.Mem); ok {
		first, second = 1, 0
		va, vb = a, b
	}
	if err := c.write(in, in.Args[first], size, va); err != nil {
		return err
	}
	return c.write(in, in.Args[second], size, vb)
}

func execXadd(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	a, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	b, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	sum, f := Add(a, b, 0, size)
	if err := c.write(in, in.Args[0], size, sum); err != nil {
		return err
	}
	if err := c.write(in, in.Args[1], size, a); err != nil {
		return err
	}
	c.setFlags(f, Arith)
	return nil
}

var accum = map[int]x86asm.Reg{1: x86asm.AL, 2: x86asm.AX, 4: x86asm.EAX, 8: x86asm.RAX}

func execCmpxchg(c *CPU, in *x86asm.Inst) error {
	size := argSize(in, in.Args[0])
	dst, err := c.read(in, in.Args[0], size)
	if err != nil {
		return err
	}
	src, err := c.read(in, in.Args[1], size)
	if err != nil {
		return err
	}
	acc := c.getReg(accum[size])
	_, f := Sub(acc, dst, 0, size)
	if f&ZF != 0 {
		if err := c.write(in, in.Args[0], size, src); err != nil {
			return err
		}
	} else {
		if _, isMem := in.Args[0].(x86asm.Mem); isMem {
			if err := c.write(in, in.Args[0], size, dst); err != nil { // always a write cycle
				return err
			}
		}
		c.setReg(accum[size], dst)
	}
	c.setFlags(f, Arith)
	return nil
}

func execSignExtendAcc(c *CPU, in *x86asm.Inst) error {
	switch in.Op {
	case x86asm.CBW:
		c.setReg(x86asm.AX, uint64(int64(int8(c.getReg(x86asm.AL)))))
	case x86asm.CWDE:
		c.setReg(x86asm.EAX, uint64(int64(int16(c.getReg(x86asm.AX)))))
	case x86asm.CDQE:
		c.setReg(x86asm.RAX, uint64(int64(int32(c.getReg(x86asm.EAX)))))
	case x86asm.CWD:
		c.setReg(x86asm.DX, uint64(int64(int16(c.getReg(x86asm.AX)))>>15))
	case x86asm.CDQ:
		c.setReg(x86asm.EDX, uint64(int64(int32(c.getReg(x86asm.EAX)))>>31))
	case x86asm.CQO:
		c.setReg(x86asm.RDX, uint64(int64(c.getReg(x86asm.RAX))>>63))
	}
	return nil
}

func execXlat(c *CPU, in *x86asm.Inst) error {
	m, _ := in.Args[0].(x86asm.Mem)
	addr := c.ea(m, in) + c.getReg(x86asm.AL)
	switch in.AddrSize {
	case 32:
		addr &= 0xffffffff
	case 16:
		addr &= 0xffff
	}
	c.faultMem = &m
	v, err := c.Mem.Read(addr, 1)
	if err != nil {
		return err
	}
	c.setReg(x86asm.AL, v)
	return nil
}

const lahfMask = SF | ZF | AF | PF | CF

func execLahf(c *CPU, in *x86asm.Inst) error {
	c.setReg(x86asm.AH, c.RFLAGS&lahfMask|reservedOne)
	return nil
}

func execSahf(c *CPU, in *x86asm.Inst) error {
	c.setFlags(c.getReg(x86asm.AH), lahfMask)
	return nil
}
