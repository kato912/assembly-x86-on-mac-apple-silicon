// Package gdbstub lets gdb (and front-ends built on it, such as DDD) debug
// a program running in nasmx86. It implements the target side of gdb's
// Remote Serial Protocol, as gdbserver or QEMU's gdb stub do: gdb loads
// the ELF executable nasmx86 wrote (symbols and DWARF line tables) and
// controls the emulated CPU over a TCP connection
// (target extended-remote localhost:PORT).
package gdbstub

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// Signal numbers in stop replies (gdb's numbering, the same as Linux's
// for these).
const (
	sigINT  = 2
	sigTRAP = 5
)

// Server debugs one program for one gdb connection.
type Server struct {
	p *runner.Program

	// Log receives a line for each packet when non-nil (debugging the stub).
	Log io.Writer

	w      *bufio.Writer
	wmu    sync.Mutex // the reader goroutine also writes (acks)
	noAck  atomic.Bool
	intr   atomic.Bool
	pkts   chan string
	bps    map[uint64]bool
	watch  map[uint64]int // write watchpoints: address -> length
	hit    uint64         // address of the watchpoint the last step triggered
	hitSet bool

	exited   bool
	exitCode int
	fault    *emu.Fault // the program crashed; continuing ends it
	killed   bool
}

// New prepares a server for p (already built and loaded).
func New(p *runner.Program) *Server {
	s := &Server{p: p, bps: map[uint64]bool{}, watch: map[uint64]int{}}
	s.hook()
	return s
}

// hook attaches breakpoint and watchpoint checks to the (re)loaded CPU.
func (s *Server) hook() {
	c := s.p.CPU
	c.MaxSteps = 0 // gdb can interrupt a runaway program
	c.BreakNative = func(a uint64) bool { return s.bps[a] }
	c.Mem.OnWrite = func(addr uint64, size int, _ uint64) {
		for a, n := range s.watch {
			if addr < a+uint64(n) && a < addr+uint64(size) {
				s.hit, s.hitSet = a, true
			}
		}
	}
}

// Serve handles one gdb connection until gdb detaches, kills the program
// for good, or disconnects. It returns the program's exit code.
func (s *Server) Serve(conn net.Conn) int {
	defer conn.Close()
	s.w = bufio.NewWriter(conn)
	s.pkts = make(chan string, 16)
	go s.read(bufio.NewReader(conn))
	for pkt := range s.pkts {
		if s.Log != nil {
			fmt.Fprintf(s.Log, "<- %s\n", trim(pkt))
		}
		reply, done := s.handle(pkt)
		if reply != nil {
			s.send(*reply)
		}
		if done {
			break
		}
	}
	return s.exitCode
}

// read parses packets ($data#cs) and interrupts (Ctrl-C, byte 0x03).
func (s *Server) read(r *bufio.Reader) {
	defer close(s.pkts)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return
		}
		switch b {
		case 0x03:
			s.intr.Store(true)
		case '$':
			var data []byte
			for {
				c, err := r.ReadByte()
				if err != nil {
					return
				}
				if c == '#' {
					break
				}
				if c == '}' { // escaped byte
					e, err := r.ReadByte()
					if err != nil {
						return
					}
					c = e ^ 0x20
				}
				data = append(data, c)
			}
			r.ReadByte() // checksum (TCP is reliable; not checked)
			r.ReadByte()
			if !s.noAck.Load() {
				s.wmu.Lock()
				s.w.WriteByte('+')
				s.w.Flush()
				s.wmu.Unlock()
			}
			s.pkts <- string(data)
		}
		// '+' and '-' acknowledgements are ignored.
	}
}

func (s *Server) send(data string) {
	if s.Log != nil {
		fmt.Fprintf(s.Log, "-> %s\n", trim(data))
	}
	s.wmu.Lock()
	defer s.wmu.Unlock()
	var sum byte
	s.w.WriteByte('$')
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c == '$' || c == '#' || c == '}' || c == '*' {
			s.w.WriteByte('}')
			sum += '}'
			c ^= 0x20
		}
		s.w.WriteByte(c)
		sum += c
	}
	fmt.Fprintf(s.w, "#%02x", sum)
	s.w.Flush()
}

// output sends text to gdb's console (an "O" packet).
func (s *Server) output(text string) {
	s.send("O" + hex.EncodeToString([]byte(text)))
}

func trim(s string) string {
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

func str(s string) *string { return &s }

const (
	ok    = "OK"
	empty = ""
)

// handle answers one packet; done ends the connection.
func (s *Server) handle(pkt string) (reply *string, done bool) {
	if pkt == "" {
		return str(empty), false
	}
	switch {
	case strings.HasPrefix(pkt, "qSupported"):
		return str("PacketSize=4000;QStartNoAckMode+;qXfer:features:read+;swbreak+;hwbreak+;vContSupported+"), false
	case pkt == "QStartNoAckMode":
		s.noAck.Store(true) // this packet was acknowledged already; no more after it
		return str(ok), false
	case pkt == "!":
		return str(ok), false
	case pkt == "?":
		return str(s.stopReply()), false
	case pkt == "qAttached" || strings.HasPrefix(pkt, "qAttached:"):
		return str("0"), false
	case pkt == "qC":
		return str("QC01"), false
	case pkt == "qfThreadInfo":
		if s.exited {
			return str("l"), false // no process
		}
		return str("m01"), false
	case pkt == "qsThreadInfo":
		return str("l"), false
	case strings.HasPrefix(pkt, "qSymbol"):
		return str(ok), false
	case pkt[0] == 'H':
		return str(ok), false
	case strings.HasPrefix(pkt, "T"): // thread alive?
		return str(ok), false
	case strings.HasPrefix(pkt, "qXfer:features:read:"):
		return str(s.features(strings.TrimPrefix(pkt, "qXfer:features:read:"))), false
	case pkt == "g":
		return str(s.readRegs()), false
	case pkt[0] == 'G':
		s.writeRegs(pkt[1:])
		return str(ok), false
	case pkt[0] == 'p':
		n, err := strconv.ParseUint(pkt[1:], 16, 32)
		if err != nil {
			return str("E01"), false
		}
		return str(s.readReg(int(n))), false
	case pkt[0] == 'P':
		n, v, found := strings.Cut(pkt[1:], "=")
		i, err := strconv.ParseUint(n, 16, 32)
		if !found || err != nil {
			return str("E01"), false
		}
		s.writeReg(int(i), v)
		return str(ok), false
	case pkt[0] == 'm':
		return str(s.readMem(pkt[1:])), false
	case pkt[0] == 'M':
		return str(s.writeMem(pkt[1:])), false
	case pkt[0] == 'Z' || pkt[0] == 'z':
		return str(s.breakpoint(pkt)), false
	case pkt == "vCont?":
		return str("vCont;c;C;s;S"), false
	case strings.HasPrefix(pkt, "vCont;"):
		action := strings.TrimPrefix(pkt, "vCont;")
		action, _, _ = strings.Cut(action, ";")
		action, _, _ = strings.Cut(action, ":")
		return str(s.resume(action != "" && (action[0] == 's' || action[0] == 'S'))), false
	case pkt[0] == 'c' || pkt[0] == 'C':
		return str(s.resume(false)), false
	case pkt[0] == 's' || pkt[0] == 'S':
		return str(s.resume(true)), false
	case strings.HasPrefix(pkt, "vRun"):
		s.restart()
		return str(s.stopReply()), false
	case strings.HasPrefix(pkt, "vKill"):
		s.kill()
		return str(ok), false
	case pkt == "k":
		s.kill()
		return nil, false
	case pkt[0] == 'D':
		s.send(ok)
		return nil, true
	case pkt == "R00" || pkt[0] == 'R':
		s.restart()
		return nil, false
	}
	return str(empty), false // not supported: gdb falls back or does without
}

// stopReply describes the current state, as for the "?" packet.
func (s *Server) stopReply() string {
	switch {
	case s.exited && s.killed:
		return "X09"
	case s.exited:
		return fmt.Sprintf("W%02x", s.exitCode&0xff)
	case s.fault != nil:
		return fmt.Sprintf("T%02xthread:01;", s.fault.Signal())
	}
	return fmt.Sprintf("T%02xthread:01;", sigTRAP)
}

func (s *Server) restart() {
	s.p.Reload()
	s.hook()
	s.exited, s.killed, s.fault, s.exitCode, s.hitSet = false, false, nil, 0, false
}

func (s *Server) kill() {
	s.exited, s.killed = true, true
}

// resume runs (or single-steps) the program and returns the stop reply.
func (s *Server) resume(step bool) string {
	switch {
	case s.exited:
		return s.stopReply()
	case s.fault != nil:
		// Continuing after a crash delivers the signal: the process dies.
		sig := s.fault.Signal()
		s.exited, s.exitCode = true, 128+sig
		return fmt.Sprintf("X%02x", sig)
	}
	c := s.p.CPU
	s.intr.Store(false)
	for n := 0; ; n++ {
		// Like an int3 on real hardware, a breakpoint at the resume address
		// stops at once; gdb removes it first when it means to step over it.
		if s.bps[c.RIP] && !(step && n == 0) {
			return fmt.Sprintf("T%02xswbreak:;thread:01;", sigTRAP)
		}
		if n&1023 == 1023 && s.intr.Load() {
			s.intr.Store(false)
			return fmt.Sprintf("T%02xthread:01;", sigINT)
		}
		s.hitSet = false
		err := c.Step()
		if err != nil {
			return s.stopped(err)
		}
		if s.hitSet {
			return fmt.Sprintf("T%02xwatch:%x;thread:01;", sigTRAP, s.hit)
		}
		if step {
			return fmt.Sprintf("T%02xthread:01;", sigTRAP)
		}
	}
}

// stopped handles the program exiting or crashing.
func (s *Server) stopped(err error) string {
	var ex *emu.Exit
	if errors.As(err, &ex) {
		s.exited, s.exitCode = true, s.p.Finish(err)
		return fmt.Sprintf("W%02x", s.exitCode&0xff)
	}
	var f *emu.Fault
	if errors.As(err, &f) {
		// Explain it in nasmx86's words, in the terminal and in gdb.
		s.output(s.p.Describe(f) + "\n")
		code := s.p.Finish(err)
		if f.Kind == emu.FaultBreakpoint {
			// int3 in the program: stop like a breakpoint, past the int3.
			s.p.CPU.RIP = f.PC + 1
			return fmt.Sprintf("T%02xthread:01;", sigTRAP)
		}
		s.fault, s.exitCode = f, code
		return fmt.Sprintf("T%02xthread:01;", f.Signal())
	}
	s.output("nasmx86: " + err.Error() + "\n")
	s.exited, s.exitCode = true, runner.ExitBuild
	return "X06"
}

// --- breakpoints ------------------------------------------------------

func (s *Server) breakpoint(pkt string) string {
	parts := strings.Split(pkt[1:], ",")
	if len(parts) < 3 {
		return "E01"
	}
	addr, err := strconv.ParseUint(parts[1], 16, 64)
	if err != nil {
		return "E01"
	}
	n, _ := strconv.ParseUint(strings.SplitN(parts[2], ";", 2)[0], 16, 32)
	insert := pkt[0] == 'Z'
	switch parts[0] {
	case "0", "1": // software and hardware breakpoints are the same here
		if insert {
			s.bps[addr] = true
		} else {
			delete(s.bps, addr)
		}
	case "2": // write watchpoint
		if insert {
			s.watch[addr] = max(int(n), 1)
		} else {
			delete(s.watch, addr)
		}
	default:
		return empty // read/access watchpoints: gdb uses software ones
	}
	return ok
}

// --- memory -----------------------------------------------------------

func (s *Server) readMem(args string) string {
	a, l, found := strings.Cut(args, ",")
	addr, err1 := strconv.ParseUint(a, 16, 64)
	n, err2 := strconv.ParseUint(l, 16, 32)
	if !found || err1 != nil || err2 != nil {
		return "E01"
	}
	b := s.p.CPU.Mem.Peek(addr, int(min(n, 0x1000)))
	if len(b) == 0 && n > 0 {
		return "E14" // EFAULT
	}
	return hex.EncodeToString(b)
}

func (s *Server) writeMem(args string) string {
	head, data, found := strings.Cut(args, ":")
	a, l, found2 := strings.Cut(head, ",")
	addr, err1 := strconv.ParseUint(a, 16, 64)
	n, err2 := strconv.ParseUint(l, 16, 32)
	b, err3 := hex.DecodeString(data)
	if !found || !found2 || err1 != nil || err2 != nil || err3 != nil || uint64(len(b)) != n {
		return "E01"
	}
	m := s.p.CPU.Mem
	for i := range b {
		if r := m.FindRegion(addr + uint64(i)); r == nil || r.Perm == 0 {
			return "E14"
		}
	}
	m.Poke(addr, b)
	s.p.CPU.InvalidateCode()
	return ok
}
