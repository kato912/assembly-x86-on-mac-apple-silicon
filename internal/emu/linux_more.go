package emu

import (
	"fmt"
	"io/fs"
	"os"
	"time"
)

// Time and file-information system calls. Structures are written with the
// exact Linux layout of the ABI in use: the 32-bit layouts for `int 0x80`
// (even from 64-bit code, like the kernel's compat layer) and the 64-bit
// layouts for `syscall`.

const ENOTEMPTY = 39

var processStart = time.Now()

// putWords writes consecutive words (4 bytes for the 32-bit ABI, 8 for the
// 64-bit one) at addr.
func putWords(c *CPU, addr uint64, int80 bool, vals ...int64) error {
	w := 8
	if int80 {
		w = 4
	}
	for i, v := range vals {
		if err := c.Mem.Write(c.wrap(addr+uint64(i*w)), w, uint64(v)); err != nil {
			return err
		}
	}
	return nil
}

func getWords(c *CPU, addr uint64, int80 bool, n int) ([]int64, error) {
	w := 8
	if int80 {
		w = 4
	}
	out := make([]int64, n)
	for i := range out {
		v, err := c.Mem.Read(c.wrap(addr+uint64(i*w)), w)
		if err != nil {
			return nil, err
		}
		out[i] = signExtend(v, w)
	}
	return out, nil
}

func (l *Linux) sysNanosleep(c *CPU, req uint64, int80 bool) (int64, string) {
	ts, err := getWords(c, req, int80, 2)
	if err != nil {
		l.warn("warn.efault", "nanosleep", req)
		return -EFAULT, fmt.Sprintf("nanosleep(%#x)", req)
	}
	desc := fmt.Sprintf("nanosleep({%d s, %d ns})", ts[0], ts[1])
	if ts[0] < 0 || ts[1] < 0 || ts[1] >= 1e9 {
		return -EINVAL, desc
	}
	time.Sleep(time.Duration(ts[0])*time.Second + time.Duration(ts[1]))
	return 0, desc
}

func (l *Linux) sysClockGettime(c *CPU, clk, tp uint64, int80 bool) (int64, string) {
	desc := fmt.Sprintf("clock_gettime(%d, %#x)", int32(clk), tp)
	var sec, nsec int64
	switch int32(clk) {
	case 0, 5, 8, 11: // REALTIME, REALTIME_COARSE, REALTIME_ALARM, TAI
		now := time.Now()
		sec, nsec = now.Unix(), int64(now.Nanosecond())
	case 1, 2, 3, 4, 6, 7, 9: // MONOTONIC (+RAW/COARSE), BOOTTIME, CPU time
		d := time.Since(processStart) + time.Hour // never 0, like system uptime
		sec, nsec = int64(d/time.Second), int64(d%time.Second)
	default:
		return -EINVAL, desc
	}
	if err := putWords(c, tp, int80, sec, nsec); err != nil {
		l.warn("warn.efault", "clock_gettime", tp)
		return -EFAULT, desc
	}
	return 0, desc
}

func (l *Linux) sysGettimeofday(c *CPU, tv, tz uint64, int80 bool) (int64, string) {
	desc := fmt.Sprintf("gettimeofday(%#x, %#x)", tv, tz)
	now := time.Now()
	if tv != 0 {
		if err := putWords(c, tv, int80, now.Unix(), int64(now.Nanosecond()/1000)); err != nil {
			l.warn("warn.efault", "gettimeofday", tv)
			return -EFAULT, desc
		}
	}
	if tz != 0 { // struct timezone: two ints, always zero here (UTC)
		if err := c.Mem.Write(c.wrap(tz), 8, 0); err != nil {
			return -EFAULT, desc
		}
	}
	return 0, desc
}

// Linux file type bits for st_mode.
const (
	sIFIFO = 0o010000
	sIFCHR = 0o020000
	sIFDIR = 0o040000
	sIFREG = 0o100000
	sIFLNK = 0o120000
)

func linuxMode(fi fs.FileInfo) uint64 {
	m := fi.Mode()
	t := uint64(sIFREG)
	switch {
	case m.IsDir():
		t = sIFDIR
	case m&fs.ModeSymlink != 0:
		t = sIFLNK
	case m&fs.ModeNamedPipe != 0:
		t = sIFIFO
	case m&fs.ModeCharDevice != 0, m&fs.ModeDevice != 0:
		t = sIFCHR
	}
	return t | uint64(m.Perm())
}

// writeStat fills a Linux struct stat at buf from host file information
// (fi == nil describes a pipe, e.g. a redirected stdin).
func writeStat(c *CPU, buf uint64, int80 bool, fi fs.FileInfo) error {
	mode, size, nlink := uint64(sIFIFO|0o600), int64(0), uint64(1)
	var mtime time.Time
	if fi != nil {
		mode, size, mtime = linuxMode(fi), fi.Size(), fi.ModTime()
		if fi.IsDir() {
			nlink = 2
		}
	} else {
		mtime = time.Now()
	}
	uid, gid := uint64(max(os.Getuid(), 0)), uint64(max(os.Getgid(), 0))
	sec, nsec := uint64(mtime.Unix()), uint64(mtime.Nanosecond())
	blocks := uint64((size + 511) / 512)
	var b []byte
	put := func(off, width int, v uint64) {
		for i := 0; i < width; i++ {
			b[off+i] = byte(v >> (8 * i))
		}
	}
	if int80 { // struct stat from arch/x86 (i386): 64 bytes
		b = make([]byte, 64)
		put(8, 2, mode)
		put(10, 2, nlink)
		put(12, 2, uid)
		put(14, 2, gid)
		put(20, 4, uint64(size))
		put(24, 4, 4096)
		put(28, 4, blocks)
		for _, off := range []int{32, 40, 48} {
			put(off, 4, sec)
			put(off+4, 4, nsec)
		}
	} else { // struct stat for x86-64: 144 bytes
		b = make([]byte, 144)
		put(16, 8, nlink)
		put(24, 4, mode)
		put(28, 4, uid)
		put(32, 4, gid)
		put(48, 8, uint64(size))
		put(56, 8, 4096)
		put(64, 8, blocks)
		for _, off := range []int{72, 88, 104} {
			put(off, 8, sec)
			put(off+8, 8, nsec)
		}
	}
	return c.Mem.WriteBytes(c.wrap(buf), b)
}

func (l *Linux) sysStat(c *CPU, pathPtr, buf uint64, int80 bool) (int64, string) {
	path, err := cString(c, pathPtr)
	if err != nil {
		l.warn("warn.efault", "stat", pathPtr)
		return -EFAULT, fmt.Sprintf("stat(%#x, %#x)", pathPtr, buf)
	}
	desc := fmt.Sprintf("stat(%q, %#x)", path, buf)
	fi, err := os.Stat(path)
	if err != nil {
		return -errno(err), desc
	}
	if err := writeStat(c, buf, int80, fi); err != nil {
		l.warn("warn.efault", "stat", buf)
		return -EFAULT, desc
	}
	return 0, desc
}

func (l *Linux) sysFstat(c *CPU, fd int, buf uint64, int80 bool) (int64, string) {
	desc := fmt.Sprintf("fstat(%d, %#x)", fd, buf)
	e := l.files[fd]
	if e == nil {
		return -EBADF, desc
	}
	var fi fs.FileInfo
	f := e.f
	if f == nil { // stdio: look at the real stream if it is a file/terminal
		if osf, ok := e.r.(*os.File); ok {
			f = osf
		} else if osf, ok := e.w.(*os.File); ok {
			f = osf
		}
	}
	if f != nil {
		var err error
		if fi, err = f.Stat(); err != nil {
			return -errno(err), desc
		}
	}
	if err := writeStat(c, buf, int80, fi); err != nil {
		l.warn("warn.efault", "fstat", buf)
		return -EFAULT, desc
	}
	return 0, desc
}

func (l *Linux) sysMkdir(c *CPU, pathPtr, mode uint64) (int64, string) {
	path, err := cString(c, pathPtr)
	if err != nil {
		return -EFAULT, fmt.Sprintf("mkdir(%#x)", pathPtr)
	}
	desc := fmt.Sprintf("mkdir(%q, %#o)", path, mode&0o7777)
	if err := os.Mkdir(path, os.FileMode(mode&0o777)); err != nil {
		return -errno(err), desc
	}
	return 0, desc
}

func (l *Linux) sysRmdir(c *CPU, pathPtr uint64) (int64, string) {
	path, err := cString(c, pathPtr)
	if err != nil {
		return -EFAULT, fmt.Sprintf("rmdir(%#x)", pathPtr)
	}
	desc := fmt.Sprintf("rmdir(%q)", path)
	fi, err := os.Lstat(path)
	switch {
	case err != nil:
		return -errno(err), desc
	case !fi.IsDir():
		return -ENOTDIR, desc
	}
	if err := os.Remove(path); err != nil {
		if entries, rerr := os.ReadDir(path); rerr == nil && len(entries) > 0 {
			return -ENOTEMPTY, desc
		}
		return -errno(err), desc
	}
	return 0, desc
}

func (l *Linux) sysRename(c *CPU, oldPtr, newPtr uint64) (int64, string) {
	oldp, err1 := cString(c, oldPtr)
	newp, err2 := cString(c, newPtr)
	if err1 != nil || err2 != nil {
		return -EFAULT, fmt.Sprintf("rename(%#x, %#x)", oldPtr, newPtr)
	}
	desc := fmt.Sprintf("rename(%q, %q)", oldp, newp)
	if err := os.Rename(oldp, newp); err != nil {
		return -errno(err), desc
	}
	return 0, desc
}
