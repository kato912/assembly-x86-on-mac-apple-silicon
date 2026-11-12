package libc

import (
	"errors"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
)

// Buffering modes, as glibc picks them: stdout is line buffered on a
// terminal and fully buffered otherwise (a pipe or a file), stderr is
// unbuffered.
const (
	unbuffered = iota
	lineBuffered
	fullyBuffered
)

const (
	bufSize     = 4096 // glibc's buffer for pipes and files
	ttyBufSize  = 1024 // and for terminals
	openMaxPath = 4096
)

// stream is a FILE.
type stream struct {
	addr     uint64 // the FILE* value the program sees
	fd       int
	name     string
	mode     int
	canRead  bool
	canWrite bool
	tty      bool

	out []byte // output not yet written
	in  []byte // input read but not yet consumed
	eof bool
	err bool
}

func (lib *Lib) initStreams() {
	base := Base(lib.c.Mode)
	lib.streams = map[uint64]*stream{}
	for i, n := range dataSymbols {
		s := &stream{addr: base + fileOff + uint64(i)*fileSize, fd: i, name: n, tty: lib.os.IsTerminal(i)}
		switch i {
		case 0:
			s.canRead = true
			s.mode = fullyBuffered
			if s.tty {
				s.mode = lineBuffered
			}
		case 1:
			s.canWrite = true
			s.mode = fullyBuffered
			if s.tty {
				s.mode = lineBuffered
			}
		case 2:
			s.canWrite = true
			s.mode = unbuffered
		}
		lib.streams[s.addr] = s
		// The stdin/stdout/stderr variables hold the FILE pointers.
		w := lib.c.Width()
		b := make([]byte, w)
		for j := range b {
			b[j] = byte(s.addr >> (8 * j))
		}
		lib.c.Mem.Poke(base+dataOff+uint64(i*w), b)
	}
	lib.nextFile = len(dataSymbols)
}

func (lib *Lib) stdStream(fd int) *stream {
	return lib.streams[Base(lib.c.Mode)+fileOff+uint64(fd)*fileSize]
}

// file returns the stream for a FILE* argument.
func (k *call) file(addr uint64) *stream {
	s := k.lib.streams[addr]
	if s == nil {
		if v := k.lib.streamPtrVar(addr); v != "" {
			k.fail(&emu.Fault{Detail: "err.libcFileVar", Args: []any{k.fn.name, v}, Hint: "hint.libcFileVar", HintArgs: []any{v}})
		}
		k.fail(&emu.Fault{Detail: "err.libcFile", Args: []any{k.fn.name, k.hex(addr)}, Hint: "hint.libcFile"})
	}
	return s
}

// streamPtrVar recognizes the address of the stdout (...) variable itself,
// passed where its value was meant (mov rdi, stdout instead of [stdout]).
func (lib *Lib) streamPtrVar(addr uint64) string {
	base := Base(lib.c.Mode)
	for i, n := range dataSymbols {
		if addr == base+dataOff+uint64(i*lib.c.Width()) {
			return n
		}
	}
	return ""
}

// put writes b to the stream, buffering like glibc.
func (lib *Lib) put(s *stream, b []byte) bool {
	if !s.canWrite {
		s.err = true
		return false
	}
	switch s.mode {
	case unbuffered:
		return lib.writeOut(s, b)
	case lineBuffered:
		s.out = append(s.out, b...)
		if indexByte(b, '\n') >= 0 || len(s.out) >= ttyBufSize {
			return lib.flush(s)
		}
	default:
		// Like glibc, a full buffer is written when more output arrives.
		s.out = append(s.out, b...)
		if len(s.out) > bufSize {
			n := len(s.out) / bufSize * bufSize
			ok := lib.writeOut(s, s.out[:n])
			s.out = append(s.out[:0], s.out[n:]...)
			return ok
		}
	}
	return true
}

func (lib *Lib) writeOut(s *stream, b []byte) bool {
	if len(b) == 0 {
		return true
	}
	if _, err := lib.os.WriteFD(s.fd, b); err != nil {
		s.err = true
		return false
	}
	return true
}

func (lib *Lib) flush(s *stream) bool {
	ok := lib.writeOut(s, s.out)
	s.out = s.out[:0]
	return ok
}

// FlushAll writes all buffered output (exit and return from main do).
func (lib *Lib) FlushAll() {
	for _, s := range lib.sortedStreams() {
		lib.flush(s)
	}
}

func (lib *Lib) sortedStreams() []*stream {
	var out []*stream
	for a := uint64(0); a < maxFiles; a++ {
		if s := lib.streams[Base(lib.c.Mode)+fileOff+a*fileSize]; s != nil {
			out = append(out, s)
		}
	}
	return out
}

// Pending returns output still in the buffers (it is lost if the
// program ends without exit or return from main).
func (lib *Lib) Pending() []byte {
	var b []byte
	for _, s := range lib.sortedStreams() {
		if s.fd == 1 || s.fd == 2 {
			b = append(b, s.out...)
		}
	}
	return b
}

// Discard drops buffered output, as the exit system call does, and warns
// about it.
func (lib *Lib) Discard() { lib.discard("warn.libcLost") }

// Crashed drops buffered output after a crash, which loses it on Linux.
func (lib *Lib) Crashed() { lib.discard("warn.libcLostCrash") }

func (lib *Lib) discard(key string) {
	lost := lib.Pending()
	for _, s := range lib.sortedStreams() {
		s.out = s.out[:0]
	}
	if len(lost) > 0 {
		lib.warn(key, len(lost), quote(lost))
	}
}

// beforeExit runs at the exit system call: returning from main flushes
// output (C's exit does that); a raw exit system call loses it.
func (lib *Lib) beforeExit(c *emu.CPU) {
	pc := c.PC()
	if lib.ExitStub != 0 && pc >= lib.ExitStub && pc < lib.ExitStub+16 {
		lib.FlushAll()
		return
	}
	lib.Discard()
}

// fill reads more input into s; it returns false at end of file.
func (lib *Lib) fill(s *stream) bool {
	if s.eof || !s.canRead {
		return false
	}
	// glibc flushes a line-buffered stdout before reading a line-buffered
	// (terminal) stream, so a prompt shows up before the program waits.
	if s.mode == lineBuffered {
		if out := lib.stdStream(1); out != nil && out.mode == lineBuffered {
			lib.flush(out)
		}
	}
	buf := make([]byte, bufSize)
	n, err := lib.os.ReadFD(s.fd, buf)
	if errors.Is(err, emu.ErrInterrupted) {
		panic(libFault{err}) // dispatch puts back what this call consumed
	}
	if n > 0 {
		s.in = append(s.in, buf[:n]...)
		if lib.readInCall == nil {
			lib.readInCall = map[*stream][]byte{}
		}
		lib.readInCall[s] = append(lib.readInCall[s], buf[:n]...)
		return true
	}
	if err != nil || n == 0 {
		s.eof = true
	}
	return false
}

// getc reads one byte, or returns -1 at end of file.
func (lib *Lib) getc(s *stream) int {
	if len(s.in) == 0 && !lib.fill(s) {
		return -1
	}
	c := s.in[0]
	s.in = s.in[1:]
	return int(c)
}

// peekc looks at the next byte without consuming it (-1 at end of file).
func (lib *Lib) peekc(s *stream) int {
	if len(s.in) == 0 && !lib.fill(s) {
		return -1
	}
	return int(s.in[0])
}

func (lib *Lib) ungetc(s *stream, c byte) {
	s.in = append([]byte{c}, s.in...)
	s.eof = false
}

// openFile implements fopen's mode strings.
func (lib *Lib) openFile(path, mode string) *stream {
	const (
		oRDONLY = 0
		oWRONLY = 1
		oRDWR   = 2
		oCREAT  = 0o100
		oTRUNC  = 0o1000
		oAPPEND = 0o2000
	)
	if mode == "" {
		return nil
	}
	var flags uint64
	plus := len(mode) > 1 && (mode[1] == '+' || len(mode) > 2 && mode[2] == '+')
	switch mode[0] {
	case 'r':
		flags = oRDONLY
	case 'w':
		flags = oWRONLY | oCREAT | oTRUNC
	case 'a':
		flags = oWRONLY | oCREAT | oAPPEND
	default:
		return nil
	}
	if plus {
		flags = flags&^3 | oRDWR
	}
	fd := lib.os.OpenFile(path, flags, 0o666)
	if fd < 0 {
		return nil
	}
	slot := -1
	for i := len(dataSymbols); i < maxFiles; i++ {
		if lib.streams[Base(lib.c.Mode)+fileOff+uint64(i)*fileSize] == nil {
			slot = i
			break
		}
	}
	if slot < 0 {
		lib.os.CloseFD(int(fd))
		return nil
	}
	s := &stream{
		addr:     Base(lib.c.Mode) + fileOff + uint64(slot)*fileSize,
		fd:       int(fd),
		name:     path,
		mode:     fullyBuffered,
		canRead:  mode[0] == 'r' || plus,
		canWrite: mode[0] != 'r' || plus,
	}
	lib.streams[s.addr] = s
	return s
}

func (lib *Lib) closeFile(s *stream) int {
	ok := lib.flush(s)
	delete(lib.streams, s.addr)
	if lib.os.CloseFD(s.fd) < 0 || !ok {
		return -1
	}
	return 0
}

// inputState is a readable stream's buffered input at the start of a call.
type inputState struct {
	s   *stream
	in  []byte
	eof bool
}

// saveInput records the buffered input of the readable streams, so that an
// interrupted call can be undone (restoreInput). Consuming input only
// reslices s.in, so keeping the slice keeps the bytes; what the call reads
// is collected in lib.readInCall.
func (lib *Lib) saveInput() []inputState {
	lib.readInCall = nil
	var saved []inputState
	for _, s := range lib.streams {
		if s.canRead {
			saved = append(saved, inputState{s, s.in, s.eof})
		}
	}
	return saved
}

// restoreInput gives the streams back the input they had when the call
// started, followed by what the call read.
func (lib *Lib) restoreInput(saved []inputState) {
	for _, st := range saved {
		in := append([]byte(nil), st.in...)
		st.s.in, st.s.eof = append(in, lib.readInCall[st.s]...), st.eof
	}
	lib.readInCall = nil
}
