package libc

import "strconv"

const eof = ^uint64(0) // -1 as an int return value

func defStdio() {
	// output
	def("printf", retInt, printfLike, func(k *call) uint64 {
		f := k.fmtArg(0)
		k.startVarargs(1)
		return k.lib.emit(k.lib.stdStream(1), k.format(f))
	})
	def("fprintf", retInt, printfLike, func(k *call) uint64 {
		s := k.fileArg(0)
		f := k.fmtArg(1)
		k.startVarargs(2)
		return k.lib.emit(s, k.format(f))
	})
	def("dprintf", retInt, printfLike, func(k *call) uint64 {
		fd := int(k.argInt(0))
		k.describe(strconv.Itoa(fd))
		f := k.fmtArg(1)
		k.startVarargs(2)
		out := k.format(f)
		if _, err := k.lib.os.WriteFD(fd, out); err != nil {
			return eof
		}
		return uint64(len(out))
	})
	def("sprintf", retInt, printfLike, func(k *call) uint64 {
		buf := k.argPtr(0)
		k.describePtr(buf)
		f := k.fmtArg(1)
		k.startVarargs(2)
		out := k.format(f)
		k.write(buf, append(out, 0))
		return uint64(len(out))
	})
	def("snprintf", retInt, printfLike, func(k *call) uint64 {
		buf := k.argPtr(0)
		n := uint64(k.argLong(1))
		if k.c.Mode == 32 {
			n &= 0xffffffff
		}
		k.describePtr(buf)
		k.describe(strconv.FormatUint(n, 10))
		f := k.fmtArg(2)
		k.startVarargs(3)
		out := k.format(f)
		if n > 0 {
			m := min(uint64(len(out)), n-1)
			k.write(buf, append(append([]byte(nil), out[:m]...), 0))
		}
		return uint64(len(out))
	})
	def("puts", retInt, fixedArgs, func(k *call) uint64 {
		s := k.strArg(0)
		return k.lib.emitOK(k.lib.stdStream(1), append(s, '\n'), uint64(len(s)+1))
	})
	def("fputs", retInt, fixedArgs, func(k *call) uint64 {
		s := k.strArg(0)
		return k.lib.emitOK(k.fileArg(1), s, 1)
	})
	def("putchar", retInt, fixedArgs, func(k *call) uint64 {
		c := byte(k.arg(0))
		k.describe(strconv.QuoteRune(rune(c)))
		return k.lib.emitOK(k.lib.stdStream(1), []byte{c}, uint64(c))
	})
	putc := func(k *call) uint64 {
		c := byte(k.arg(0))
		k.describe(strconv.QuoteRune(rune(c)))
		return k.lib.emitOK(k.fileArg(1), []byte{c}, uint64(c))
	}
	def("fputc", retInt, fixedArgs, putc)
	def("putc", retInt, fixedArgs, putc)
	def("fwrite", retLong, fixedArgs, func(k *call) uint64 {
		p, size, n := k.argPtr(0), k.size(1), k.size(2)
		k.describePtr(p)
		k.describe(strconv.FormatUint(size, 10))
		k.describe(strconv.FormatUint(n, 10))
		s := k.fileArg(3)
		if size == 0 || n == 0 {
			return 0
		}
		if !k.lib.put(s, k.read(p, int(size*n))) {
			return 0
		}
		return n
	})
	def("fflush", retInt, fixedArgs, func(k *call) uint64 {
		a := k.argPtr(0)
		if a == 0 {
			k.describe("NULL")
			k.lib.FlushAll()
			return 0
		}
		if !k.lib.flush(k.fileArg(0)) {
			return eof
		}
		return 0
	})

	// input
	def("scanf", retInt, scanfLike, func(k *call) uint64 {
		f := k.fmtArg(0)
		k.startVarargs(1)
		return uint64(k.scan(streamSource{k.lib, k.lib.stdStream(0)}, f))
	})
	def("fscanf", retInt, scanfLike, func(k *call) uint64 {
		s := k.fileArg(0)
		f := k.fmtArg(1)
		k.startVarargs(2)
		return uint64(k.scan(streamSource{k.lib, s}, f))
	})
	def("sscanf", retInt, scanfLike, func(k *call) uint64 {
		str := k.strArg(0)
		f := k.fmtArg(1)
		k.startVarargs(2)
		return uint64(k.scan(&stringSource{b: str}, f))
	})
	def("getchar", retInt, fixedArgs, func(k *call) uint64 {
		return uint64(int64(k.lib.getc(k.lib.stdStream(0))))
	})
	getc := func(k *call) uint64 { return uint64(int64(k.lib.getc(k.fileArg(0)))) }
	def("fgetc", retInt, fixedArgs, getc)
	def("getc", retInt, fixedArgs, getc)
	def("ungetc", retInt, fixedArgs, func(k *call) uint64 {
		c := int32(k.arg(0))
		k.describe(strconv.Itoa(int(c)))
		s := k.fileArg(1)
		if c == -1 {
			return eof
		}
		k.lib.ungetc(s, byte(c))
		return uint64(byte(c))
	})
	def("fgets", retLong, fixedArgs, func(k *call) uint64 {
		buf := k.argPtr(0)
		n := int(k.argInt(1))
		k.describePtr(buf)
		k.describe(strconv.Itoa(n))
		s := k.fileArg(2)
		if n <= 0 {
			return 0
		}
		var line []byte
		for len(line) < n-1 {
			c := k.lib.getc(s)
			if c < 0 {
				break
			}
			line = append(line, byte(c))
			if c == '\n' {
				break
			}
		}
		if len(line) == 0 {
			return 0
		}
		k.write(buf, append(line, 0))
		return buf
	})
	def("gets", retLong, fixedArgs, func(k *call) uint64 {
		buf := k.argPtr(0)
		k.describePtr(buf)
		s := k.lib.stdStream(0)
		var line []byte
		c := k.lib.getc(s)
		if c < 0 {
			return 0
		}
		for ; c >= 0 && c != '\n'; c = k.lib.getc(s) {
			line = append(line, byte(c))
		}
		k.write(buf, append(line, 0))
		return buf
	})

	// files
	def("fopen", retLong, fixedArgs, func(k *call) uint64 {
		path := string(k.strArg(0))
		mode := string(k.strArg(1))
		s := k.lib.openFile(path, mode)
		if s == nil {
			return 0
		}
		return s.addr
	})
	def("fclose", retInt, fixedArgs, func(k *call) uint64 {
		return uint64(int64(k.lib.closeFile(k.fileArg(0))))
	})
	def("feof", retInt, fixedArgs, func(k *call) uint64 {
		s := k.fileArg(0)
		if s.eof && len(s.in) == 0 {
			return 1
		}
		return 0
	})
}

// emit writes formatted output and returns printf's result.
func (lib *Lib) emit(s *stream, b []byte) uint64 {
	if !lib.put(s, b) {
		return eof
	}
	return uint64(len(b))
}

func (lib *Lib) emitOK(s *stream, b []byte, ret uint64) uint64 {
	if !lib.put(s, b) {
		return eof
	}
	return ret
}

// fmtArg reads a format string argument.
func (k *call) fmtArg(n int) []byte {
	a := k.argPtr(n)
	if k.c.Mode == 32 {
		a &= 0xffffffff
	}
	f := k.cstr(a, -1)
	k.describe(quote(f))
	return f
}

// strArg reads a string argument.
func (k *call) strArg(n int) []byte {
	a := k.argPtr(n)
	if k.c.Mode == 32 {
		a &= 0xffffffff
	}
	s := k.cstr(a, -1)
	k.describe(quote(s))
	return s
}

// fileArg reads a FILE* argument.
func (k *call) fileArg(n int) *stream {
	a := k.argPtr(n)
	if k.c.Mode == 32 {
		a &= 0xffffffff
	}
	s := k.file(a)
	k.describe(s.name)
	return s
}

// size reads a size_t argument.
func (k *call) size(n int) uint64 {
	v := k.arg(n)
	if k.c.Mode == 32 {
		v &= 0xffffffff
	}
	return v
}
