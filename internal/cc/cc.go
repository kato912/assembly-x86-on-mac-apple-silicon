// Package cc compiles C source files into x86 Linux object files for
// nasmx86's gcc (gcc main.c stats.o -o cstats), with the Mac's own clang
// as a cross compiler: clang's x86 back end writes the ELF object, and the
// headers here declare the functions of nasmx86's built-in C library.
//
// nasmx86 does not run floating-point or SSE instructions, so the code is
// compiled without SSE: integer C works, while float and double are
// reported by clang.
package cc

import (
	"bytes"
	"embed"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/pathfind"
)

//go:embed include/*.h
var headers embed.FS

// ErrNoCompiler means clang is not installed.
var ErrNoCompiler = errors.New("clang not found")

// ErrFloat means the C code uses floating point, which nasmx86 cannot run.
var ErrFloat = errors.New("floating point")

// Find returns the path of clang.
func Find() (string, error) {
	if p, ok := pathfind.Other("clang", "/usr/bin"); ok {
		return p, nil
	}
	return "", ErrNoCompiler
}

// passFlags reports whether a gcc option is passed on to the compiler.
func passFlags(a string) bool {
	for _, p := range []string{"-O", "-W", "-w", "-D", "-U", "-I", "-std=", "-pedantic", "-ansi", "-fsigned-char", "-funsigned-char"} {
		if strings.HasPrefix(a, p) {
			return true
		}
	}
	return false
}

// Compile compiles src for a 32- or 64-bit Linux program into dir and
// returns the object's path. flags are gcc's options from the command
// line (those that matter for compiling are passed on); debug adds DWARF
// line information (gcc -g). Compiler messages go to stderr.
func Compile(src string, bits int, flags []string, debug bool, dir string, stderr io.Writer) (string, error) {
	clang, err := Find()
	if err != nil {
		return "", err
	}
	inc := filepath.Join(dir, "include")
	if err := writeHeaders(inc); err != nil {
		return "", err
	}
	target := "x86_64-linux-gnu"
	if bits == 32 {
		target = "i386-linux-gnu"
	}
	obj := filepath.Join(dir, strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))+".o")
	if _, err := os.Stat(obj); err == nil {
		obj = filepath.Join(dir, filepath.Base(src)+".o") // two files with the same name
	}
	args := []string{"--target=" + target, "-nostdlibinc", "-isystem", inc,
		"-mno-sse", "-mno-mmx", "-fno-pic", "-fno-pie", "-fno-stack-protector",
		"-fno-asynchronous-unwind-tables", "-fno-unwind-tables", "-fno-addrsig",
		"-Wno-unused-command-line-argument"}
	if debug {
		args = append(args, "-gdwarf-4")
	}
	for i := 0; i < len(flags); i++ {
		a := flags[i]
		switch {
		case (a == "-I" || a == "-D" || a == "-U") && i+1 < len(flags):
			args = append(args, a, flags[i+1])
			i++
		case passFlags(a):
			args = append(args, a)
		}
	}
	args = append(args, "-c", src, "-o", obj)
	var msgs bytes.Buffer
	cmd := exec.Command(clang, args...)
	cmd.Stdout = stderr
	cmd.Stderr = io.MultiWriter(stderr, &msgs)
	if err := cmd.Run(); err != nil {
		if strings.Contains(msgs.String(), "SSE") || strings.Contains(msgs.String(), "x87") {
			return "", ErrFloat
		}
		return "", err
	}
	return obj, nil
}

func writeHeaders(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(headers, "include", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := headers.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, filepath.Base(path)), data, 0o644)
	})
}
