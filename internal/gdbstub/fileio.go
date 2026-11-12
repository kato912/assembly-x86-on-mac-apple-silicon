package gdbstub

// When gdb or DDD is started by nasmx86 (gdb ./hello, ddd ./hello &), the
// program's standard input and output go through gdb, with the File-I/O
// extension of gdb's remote protocol: on write the stub asks gdb to
// "Fwrite,1,buffer,count", gdb reads the bytes from the target's memory and
// prints them on its console, and replies "Fcount". So in DDD the
// program's output and input are in DDD's console, as when DDD runs a
// program on Linux, and nasmx86 never reads the terminal it was started
// from (which would stop it there when started in the background with &).

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"syscall"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

// fioSize is the size of the buffer gdb reads output from and writes input
// to. It lives at fioAddr, an address the program never uses: the stub
// answers gdb's memory reads and writes there itself.
const fioSize = 0x4000

func fioAddr(bits int) uint64 {
	if bits == 32 {
		return 0xffff0000 // above the 32-bit stack
	}
	return 0x7fff00000000 // far below the 64-bit stack
}

// errDisconnected means gdb went away during a File-I/O request.
var errDisconnected = errors.New("gdb disconnected")

// UseConsole routes the program's standard input and output through gdb's
// console. Call it before serving a connection.
func (s *Server) UseConsole() {
	cfg := &s.p.Cfg
	s.fio = make([]byte, fioSize)
	s.fioAt = fioAddr(s.p.Bits)
	cfg.Stdin = &console{s: s, fd: 0, r: cfg.Stdin}
	cfg.Stdout = &console{s: s, fd: 1, w: cfg.Stdout}
	cfg.Stderr = &console{s: s, fd: 2, w: cfg.Stderr}
	s.console = true
	s.restart() // load the program again with these streams
}

// console is one of the program's standard streams, going through gdb
// while the program runs. Outside a run (there is no gdb request to answer
// then) it uses the terminal's stream.
type console struct {
	s  *Server
	fd int
	r  io.Reader
	w  io.Writer
}

// IsTerminal tells the C library that this is a terminal (line-buffered
// output, flushed before reading input), as DDD's console is on Linux.
func (c *console) IsTerminal() bool { return true }

func (c *console) Write(b []byte) (int, error) {
	s := c.s
	if !s.running {
		return c.w.Write(b)
	}
	total := 0
	for len(b) > 0 {
		n := copy(s.fio, b)
		ret, _, ctrlC, err := s.fileRequest(fmt.Sprintf("Fwrite,%x,%x,%x", c.fd, s.fioAt, n))
		if ctrlC {
			s.stopNow = true
		}
		if err != nil {
			return total, err
		}
		if ret <= 0 {
			return total, io.ErrShortWrite
		}
		total += int(ret)
		b = b[ret:]
	}
	return total, nil
}

func (c *console) Read(b []byte) (int, error) {
	s := c.s
	if !s.running {
		return c.r.Read(b)
	}
	n := min(len(b), fioSize)
	ret, errno, ctrlC, err := s.fileRequest(fmt.Sprintf("Fread,%x,%x,%x", c.fd, s.fioAt, n))
	switch {
	case err != nil:
		return 0, err
	case ctrlC || (ret < 0 && errno == int(syscall.EINTR)):
		// Ctrl-C (DDD's Interrupt) while the program waits for input:
		// stop before the read, and read again on continue.
		return 0, emu.ErrInterrupted
	case ret < 0:
		return 0, syscall.Errno(errno)
	case ret == 0:
		return 0, io.EOF // Ctrl-D
	}
	return copy(b, s.fio[:ret]), nil
}

// fileRequest sends a File-I/O request and answers gdb's memory accesses
// until gdb's reply "Fretcode[,errno[,C]]".
func (s *Server) fileRequest(req string) (ret int64, errno int, ctrlC bool, err error) {
	s.send(req)
	for pkt := range s.pkts {
		if s.Log != nil {
			fmt.Fprintf(s.Log, "<- %s\n", trim(pkt))
		}
		if strings.HasPrefix(pkt, "F") {
			return parseFileReply(pkt[1:])
		}
		switch pkt[0] {
		case 'm', 'M', 'g', 'p', 'q':
			if reply, _ := s.handle(pkt); reply != nil {
				s.send(*reply)
			}
		default:
			s.send(empty) // nothing else belongs in a File-I/O exchange
		}
	}
	return 0, 0, false, errDisconnected
}

// parseFileReply parses the body of gdb's F reply: retcode, errno, C.
func parseFileReply(body string) (ret int64, errno int, ctrlC bool, err error) {
	body, _, _ = strings.Cut(body, ";")
	parts := strings.Split(body, ",")
	if ret, err = strconv.ParseInt(parts[0], 16, 64); err != nil {
		return 0, 0, false, fmt.Errorf("bad File-I/O reply F%s", body)
	}
	if len(parts) > 1 && parts[1] != "" {
		e, err := strconv.ParseInt(parts[1], 16, 32)
		if err != nil {
			return 0, 0, false, fmt.Errorf("bad File-I/O reply F%s", body)
		}
		errno = int(e)
	}
	return ret, errno, len(parts) > 2 && parts[2] == "C", nil
}

// fioRead answers a memory read in the File-I/O buffer.
func (s *Server) fioRead(addr, n uint64) ([]byte, bool) {
	if s.fio == nil || addr < s.fioAt || addr >= s.fioAt+fioSize {
		return nil, false
	}
	off := addr - s.fioAt
	return s.fio[off:min(off+n, fioSize)], true
}

// fioWrite answers a memory write in the File-I/O buffer.
func (s *Server) fioWrite(addr uint64, b []byte) bool {
	if s.fio == nil || addr < s.fioAt || addr >= s.fioAt+fioSize {
		return false
	}
	copy(s.fio[addr-s.fioAt:], b)
	return true
}
