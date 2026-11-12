package gdbstub

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

// reg is one register in gdb's numbering for i386 / amd64 (the order of
// gdb's 32bit-core.xml + 32bit-sse.xml and 64bit-core.xml +
// 64bit-sse.xml).
type reg struct {
	name string
	size int // bytes
	gpr  int // emu register index, -1 for others
	kind string
}

const (
	kGPR = "gpr"
	kIP  = "ip"
	kFL  = "flags"
	kSeg = "seg"
	kFPU = "fpu"
	kXMM = "xmm"
)

func regList(bits int) []reg {
	var rs []reg
	if bits == 64 {
		for _, r := range []struct {
			n string
			i int
		}{{"rax", emu.RAX}, {"rbx", emu.RBX}, {"rcx", emu.RCX}, {"rdx", emu.RDX}, {"rsi", emu.RSI}, {"rdi", emu.RDI},
			{"rbp", emu.RBP}, {"rsp", emu.RSP}, {"r8", emu.R8}, {"r9", emu.R9}, {"r10", emu.R10}, {"r11", emu.R11},
			{"r12", emu.R12}, {"r13", emu.R13}, {"r14", emu.R14}, {"r15", emu.R15}} {
			rs = append(rs, reg{r.n, 8, r.i, kGPR})
		}
		rs = append(rs, reg{"rip", 8, -1, kIP})
	} else {
		for i, n := range []string{"eax", "ecx", "edx", "ebx", "esp", "ebp", "esi", "edi"} {
			rs = append(rs, reg{n, 4, i, kGPR}) // hardware order = emu order
		}
		rs = append(rs, reg{"eip", 4, -1, kIP})
	}
	rs = append(rs, reg{"eflags", 4, -1, kFL})
	for _, n := range []string{"cs", "ss", "ds", "es", "fs", "gs"} {
		rs = append(rs, reg{n, 4, -1, kSeg})
	}
	for i := 0; i < 8; i++ {
		rs = append(rs, reg{fmt.Sprintf("st%d", i), 10, -1, kFPU})
	}
	for _, n := range []string{"fctrl", "fstat", "ftag", "fiseg", "fioff", "foseg", "fooff", "fop"} {
		rs = append(rs, reg{n, 4, -1, kFPU})
	}
	nx := 8
	if bits == 64 {
		nx = 16
	}
	for i := 0; i < nx; i++ {
		rs = append(rs, reg{fmt.Sprintf("xmm%d", i), 16, -1, kXMM})
	}
	return append(rs, reg{"mxcsr", 4, -1, kXMM})
}

// Values Linux gives a process for registers nasmx86 does not emulate.
func (s *Server) fixed(r reg) uint64 {
	bits := s.p.Bits
	switch r.name {
	case "cs":
		if bits == 64 {
			return 0x33
		}
		return 0x23
	case "ss":
		return 0x2b
	case "ds", "es":
		if bits == 32 {
			return 0x2b
		}
	case "fctrl":
		return 0x37f
	case "mxcsr":
		return 0x1f80
	}
	return 0
}

func (s *Server) regValue(r reg) []byte {
	c := s.p.CPU
	b := make([]byte, r.size)
	var v uint64
	switch r.kind {
	case kGPR:
		v = c.R[r.gpr]
	case kIP:
		v = c.RIP
	case kFL:
		v = c.RFLAGS
	default:
		v = s.fixed(r)
	}
	for i := 0; i < r.size && i < 8; i++ {
		b[i] = byte(v >> (8 * i))
	}
	return b
}

func (s *Server) setReg(r reg, b []byte) {
	c := s.p.CPU
	var v uint64
	for i := min(len(b), 8) - 1; i >= 0; i-- {
		v = v<<8 | uint64(b[i])
	}
	switch r.kind {
	case kGPR:
		if r.size == 4 {
			v &= 0xffffffff
		}
		c.R[r.gpr] = v
	case kIP:
		c.RIP = v
	case kFL:
		c.RFLAGS = v&0xed5 | 0x2 // CF PF AF ZF SF IF DF OF; bit 1 is always set
	}
}

func (s *Server) readRegs() string {
	var b strings.Builder
	for _, r := range regList(s.p.Bits) {
		b.WriteString(hex.EncodeToString(s.regValue(r)))
	}
	return b.String()
}

func (s *Server) writeRegs(h string) {
	data, err := hex.DecodeString(h)
	if err != nil {
		return
	}
	for _, r := range regList(s.p.Bits) {
		if len(data) < r.size {
			return
		}
		s.setReg(r, data[:r.size])
		data = data[r.size:]
	}
}

func (s *Server) readReg(n int) string {
	rs := regList(s.p.Bits)
	if n < 0 || n >= len(rs) {
		return "E01"
	}
	return hex.EncodeToString(s.regValue(rs[n]))
}

func (s *Server) writeReg(n int, h string) {
	rs := regList(s.p.Bits)
	b, err := hex.DecodeString(h)
	if n < 0 || n >= len(rs) || err != nil {
		return
	}
	s.setReg(rs[n], b)
}

// features answers qXfer:features:read:ANNEX:OFFSET,LENGTH.
func (s *Server) features(args string) string {
	annex, rng, found := strings.Cut(args, ":")
	if !found || annex != "target.xml" {
		return "E00"
	}
	o, l, _ := strings.Cut(rng, ",")
	off, err1 := strconv.ParseUint(o, 16, 32)
	n, err2 := strconv.ParseUint(l, 16, 32)
	if err1 != nil || err2 != nil {
		return "E01"
	}
	doc := targetXML(s.p.Bits)
	if off >= uint64(len(doc)) {
		return "l"
	}
	end := min(off+n, uint64(len(doc)))
	if end == uint64(len(doc)) {
		return "l" + doc[off:end]
	}
	return "m" + doc[off:end]
}

// targetXML describes the registers to gdb.
func targetXML(bits int) string {
	var b strings.Builder
	arch := "i386"
	if bits == 64 {
		arch = "i386:x86-64"
	}
	fmt.Fprintf(&b, `<?xml version="1.0"?>
<!DOCTYPE target SYSTEM "gdb-target.dtd">
<target version="1.0">
<architecture>%s</architecture>
<feature name="org.gnu.gdb.i386.core">
<flags id="i386_eflags" size="4">
<field name="CF" start="0" end="0"/><field name="" start="1" end="1"/><field name="PF" start="2" end="2"/>
<field name="AF" start="4" end="4"/><field name="ZF" start="6" end="6"/><field name="SF" start="7" end="7"/>
<field name="TF" start="8" end="8"/><field name="IF" start="9" end="9"/><field name="DF" start="10" end="10"/>
<field name="OF" start="11" end="11"/><field name="NT" start="14" end="14"/><field name="RF" start="16" end="16"/>
<field name="VM" start="17" end="17"/><field name="AC" start="18" end="18"/><field name="VIF" start="19" end="19"/>
<field name="VIP" start="20" end="20"/><field name="ID" start="21" end="21"/>
</flags>
`, arch)
	inSSE := false
	for _, r := range regList(bits) {
		if r.kind == kXMM && !inSSE {
			inSSE = true
			b.WriteString(`</feature>
<feature name="org.gnu.gdb.i386.sse">
<vector id="v4f" type="ieee_single" count="4"/><vector id="v2d" type="ieee_double" count="2"/>
<vector id="v16i8" type="int8" count="16"/><vector id="v8i16" type="int16" count="8"/>
<vector id="v4i32" type="int32" count="4"/><vector id="v2i64" type="int64" count="2"/>
<union id="vec128"><field name="v4_float" type="v4f"/><field name="v2_double" type="v2d"/>
<field name="v16_int8" type="v16i8"/><field name="v8_int16" type="v8i16"/><field name="v4_int32" type="v4i32"/>
<field name="v2_int64" type="v2i64"/><field name="uint128" type="uint128"/></union>
`)
		}
		typ, group := "int", ""
		switch {
		case r.kind == kGPR && (strings.HasSuffix(r.name, "sp") || strings.HasSuffix(r.name, "bp")):
			typ = "data_ptr"
		case r.kind == kGPR && bits == 64:
			typ = "int64"
		case r.kind == kGPR:
			typ = "int32"
		case r.kind == kIP:
			typ = "code_ptr"
		case r.kind == kFL:
			typ = "i386_eflags"
		case r.kind == kSeg:
			typ = "int32"
		case r.kind == kFPU && r.size == 10:
			typ = "i387_ext"
		case r.kind == kFPU:
			group = ` group="float"`
		case r.name == "mxcsr":
			group = ` group="vector"`
		case r.kind == kXMM:
			typ = "vec128"
		}
		fmt.Fprintf(&b, "<reg name=\"%s\" bitsize=\"%d\" type=\"%s\"%s/>\n", r.name, r.size*8, typ, group)
	}
	b.WriteString("</feature>\n</target>\n")
	return b.String()
}
