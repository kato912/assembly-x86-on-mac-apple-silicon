package emu

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"
	"time"
)

// Linux open(2) flag bits (the host OS may use different values).
const (
	oWRONLY  = 01
	oRDWR    = 02
	oCREAT   = 0100
	oEXCL    = 0200
	oTRUNC   = 01000
	oAPPEND  = 02000
	atFDCWD  = -100
	maxPath  = 4096
	maxRead  = 1 << 20
	brkLimit = 1 << 30 // how far the heap may grow
)

// cString reads a NUL-terminated string from emulated memory.
func cString(c *CPU, addr uint64) (string, error) {
	var b []byte
	for i := uint64(0); i < maxPath; i++ {
		v, err := c.Mem.Read(c.wrap(addr+i), 1)
		if err != nil {
			return "", err
		}
		if v == 0 {
			return string(b), nil
		}
		b = append(b, byte(v))
	}
	return "", errors.New("path too long")
}

// errno maps a host error to a Linux errno value.
func errno(err error) int64 {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, fs.ErrNotExist):
		return ENOENT
	case errors.Is(err, fs.ErrExist):
		return EEXIST
	case errors.Is(err, fs.ErrPermission):
		return EACCES
	case errors.Is(err, syscall.EISDIR):
		return EISDIR
	case errors.Is(err, syscall.ENOTDIR):
		return ENOTDIR
	case errors.Is(err, syscall.EMFILE):
		return EMFILE
	case errors.Is(err, syscall.ENOSPC):
		return ENOSPC
	case errors.Is(err, fs.ErrClosed):
		return EBADF
	}
	return EIO
}

func (l *Linux) newFD(e *fdEntry) int {
	for fd := 0; ; fd++ {
		if l.files[fd] == nil {
			l.files[fd] = e
			return fd
		}
	}
}

func (l *Linux) sysRead(c *CPU, fd, buf, count uint64) (int64, string) {
	desc := fmt.Sprintf("read(%d, %#x, %d)", int32(fd), buf, int64(count))
	e := l.files[int(int32(fd))]
	if e == nil || e.r == nil {
		l.warn("warn.badfd", "read", int32(fd))
		return -EBADF, desc
	}
	if int64(count) < 0 {
		return -EINVAL, desc
	}
	if count == 0 {
		return 0, desc
	}
	// Read no more than fits in writable memory at buf, like the kernel
	// copying into the user buffer page by page.
	n := count
	if n > maxRead {
		n = maxRead
	}
	room := uint64(0)
	for room < n {
		if err := c.Mem.Check(c.wrap(buf+room), 1, AccessWrite); err != nil {
			break
		}
		room += PageSize - (buf+room)%PageSize
	}
	if room == 0 {
		l.warn("warn.efault", "read", buf)
		return -EFAULT, desc
	}
	if room < n {
		n = room
	}
	data := make([]byte, n)
	got, err := e.r.Read(data)
	if got == 0 && err != nil && !errors.Is(err, io.EOF) {
		return -errno(err), desc
	}
	c.Mem.Poke(c.wrap(buf), data[:got])
	if c.Mem.OnWrite != nil && got > 0 {
		c.Mem.OnWrite(c.wrap(buf), got, 0)
	}
	return int64(got), fmt.Sprintf("read(%d, %#x %s, %d)", int32(fd), buf, quote(data[:got], 32), int64(count))
}

func (l *Linux) sysOpen(c *CPU, pathPtr, flags, mode uint64) (int64, string) {
	path, err := cString(c, pathPtr)
	if err != nil {
		l.warn("warn.efault", "open", pathPtr)
		return -EFAULT, fmt.Sprintf("open(%#x, %#o, %#o)", pathPtr, flags, mode)
	}
	desc := fmt.Sprintf("open(%q, %#o, %#o)", path, flags&0xffffffff, mode&0o7777)
	var gf int
	switch flags & 3 {
	case 0:
		gf = os.O_RDONLY
	case oWRONLY:
		gf = os.O_WRONLY
	default:
		gf = os.O_RDWR
	}
	if flags&oCREAT != 0 {
		gf |= os.O_CREATE
	}
	if flags&oEXCL != 0 {
		gf |= os.O_EXCL
	}
	if flags&oTRUNC != 0 {
		gf |= os.O_TRUNC
	}
	if flags&oAPPEND != 0 {
		gf |= os.O_APPEND
	}
	f, err := os.OpenFile(path, gf, os.FileMode(mode&0o777))
	if err != nil {
		return -errno(err), desc
	}
	e := &fdEntry{f: f}
	if flags&3 != oWRONLY {
		e.r = f
	}
	if flags&3 != 0 {
		e.w = f
	}
	return int64(l.newFD(e)), desc
}

func (l *Linux) sysUnlink(c *CPU, pathPtr uint64) (int64, string) {
	path, err := cString(c, pathPtr)
	if err != nil {
		return -EFAULT, fmt.Sprintf("unlink(%#x)", pathPtr)
	}
	desc := fmt.Sprintf("unlink(%q)", path)
	if st, err := os.Lstat(path); err == nil && st.IsDir() {
		return -EISDIR, desc
	}
	if err := os.Remove(path); err != nil {
		return -errno(err), desc
	}
	return 0, desc
}

func (l *Linux) sysClose(fd int) (int64, string) {
	desc := fmt.Sprintf("close(%d)", fd)
	e := l.files[fd]
	if e == nil {
		return -EBADF, desc
	}
	delete(l.files, fd)
	if e.f != nil {
		if err := e.f.Close(); err != nil {
			return -errno(err), desc
		}
	}
	return 0, desc
}

func (l *Linux) sysLseek(fd int, off int64, whence int) (int64, string) {
	desc := fmt.Sprintf("lseek(%d, %d, %d)", fd, off, whence)
	e := l.files[fd]
	if e == nil {
		return -EBADF, desc
	}
	if e.f == nil {
		return -ESPIPE, desc
	}
	if whence < 0 || whence > 2 {
		return -EINVAL, desc
	}
	pos, err := e.f.Seek(off, whence)
	if err != nil {
		return -EINVAL, desc
	}
	return pos, desc
}

// InitBrk sets the start of the program break (end of .bss, page aligned).
func (l *Linux) InitBrk(c *CPU, start uint64) {
	l.brkStart, l.brkCur = start, start
	l.heap = c.Mem.Map(start, 0, PermR|PermW, "heap", KindHeap)
}

func (l *Linux) sysBrk(c *CPU, addr uint64) uint64 {
	if l.heap == nil || addr < l.brkStart || addr > l.brkStart+brkLimit {
		return l.brkCur
	}
	if addr > l.heap.End {
		c.Mem.Grow(l.heap, addr)
	}
	l.brkCur = addr
	return l.brkCur
}

func (l *Linux) sysTime(c *CPU, tloc uint64, int80 bool) (int64, string) {
	t := time.Now().Unix()
	desc := fmt.Sprintf("time(%#x)", tloc)
	if tloc != 0 {
		size := 8
		if int80 {
			size = 4
		}
		if err := c.Mem.Write(c.wrap(tloc), size, uint64(t)); err != nil {
			return -EFAULT, desc
		}
	}
	if int80 {
		t = int64(int32(t))
	}
	return t, desc
}
