package main

// nasmx86 as ld and gcc: on a Mac, the course's commands
//
//	yasm -f elf64 -g dwarf2 -o hello.o hello.s
//	ld -o hello hello.o          (or: gcc -o hello hello.o for main + printf)
//	./hello
//
// work as on Linux. nasmx86 links the x86 objects into a program file that
// ./hello runs in nasmx86 (see internal/progfile). Anything else, such as
// a Mac program's objects or C sources, goes to the Mac's own ld or gcc.

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/cc"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/pathfind"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/progfile"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// linkCmd is a parsed ld or gcc command line.
type linkCmd struct {
	tool    string // "ld" or "gcc"
	out     string
	bits    int      // 32 or 64 (-m elf_i386 / -m32), 0 = not given
	entry   string   // ld -e
	clib    bool     // ld -lc
	noCRT   bool     // gcc -nostartfiles / -nostdlib: no C runtime, start at _start
	noLibc  bool     // gcc -nostdlib / -nodefaultlibs
	inputs  []string // object files (and, for gcc, C files), in command-line order
	csrc    []string // gcc: the C source files among inputs
	cflags  []string // gcc: options, for compiling the C files
	debug   bool     // gcc -g
	errs    []string // problems with the options, reported unless the Mac's tool takes over
	command []string
}

func (c *linkCmd) fail(key string, args ...any) { c.errs = append(c.errs, i18n.T(key, args...)) }

// ldValueOpts are GNU ld options whose value is the next argument and that
// nasmx86 can ignore.
var ldValueOpts = map[string]bool{"-L": true, "-z": true, "-dynamic-linker": true, "--dynamic-linker": true,
	"-I": true, "-soname": true, "-h": true, "-rpath": true, "-rpath-link": true, "--hash-style": true,
	"-Map": true, "--sysroot": true, "-plugin": true, "-plugin-opt": true, "-y": true, "-u": true}

// ldFlags are GNU ld options without a value that make no difference here.
var ldFlags = map[string]bool{"-g": true, "-s": true, "-S": true, "--strip-all": true, "--strip-debug": true,
	"-static": true, "-Bstatic": true, "-Bdynamic": true, "-no-pie": true, "--no-pie": true, "-pie": true,
	"-n": true, "-N": true, "-x": true, "-X": true, "-E": true, "--export-dynamic": true, "-export-dynamic": true,
	"--as-needed": true, "--no-as-needed": true, "--eh-frame-hdr": true, "--gc-sections": true,
	"--no-gc-sections": true, "--verbose": true, "-nostdlib": true, "--build-id": true,
	"-O0": true, "-O1": true, "-O2": true, "-O3": true, "--start-group": true, "--end-group": true,
	"-(": true, "-)": true, "--warn-common": true, "--fatal-warnings": true, "-M": true, "--print-map": true}

func parseLD(args []string) *linkCmd {
	c := &linkCmd{tool: "ld", command: append([]string{"ld"}, args...)}
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() string {
			if i+1 < len(args) {
				i++
				return args[i]
			}
			c.fail("err.missingValue", a)
			return ""
		}
		name, val, hasVal := strings.Cut(a, "=")
		switch {
		case !strings.HasPrefix(a, "-") || a == "-":
			c.inputs = append(c.inputs, a)
		case ldFlags[a]:
		case ldValueOpts[a]:
			next()
		case a == "-o" || a == "--output":
			c.out = next()
		case name == "--output" && hasVal:
			c.out = val
		case strings.HasPrefix(a, "-o") && !strings.HasPrefix(a, "--"):
			c.out = a[2:]
		case a == "-m":
			c.emulation(next())
		case strings.HasPrefix(a, "-m") && !strings.HasPrefix(a, "--"):
			c.emulation(a[2:])
		case a == "-e" || a == "--entry":
			c.entry = next()
		case name == "--entry" && hasVal:
			c.entry = val
		case strings.HasPrefix(a, "-e") && !strings.HasPrefix(a, "--"):
			c.entry = a[2:]
		case a == "-l" || a == "--library":
			c.library(next())
		case name == "--library" && hasVal:
			c.library(val)
		case strings.HasPrefix(a, "-l") && !strings.HasPrefix(a, "--"):
			c.library(a[2:])
		case hasVal && (ldValueOpts[name] || ldFlags[name]), strings.HasPrefix(a, "-L"),
			strings.HasPrefix(a, "--build-id"), strings.HasPrefix(a, "-z"):
		case a == "-T" || strings.HasPrefix(a, "-T") || strings.HasPrefix(a, "--script"):
			c.fail("err.unsupportedOption", "ld", a)
		default:
			c.fail("err.badOption", a)
		}
	}
	return c
}

func (c *linkCmd) emulation(m string) {
	switch m {
	case "elf_i386":
		c.bits = 32
	case "elf_x86_64":
		c.bits = 64
	default:
		c.fail("err.badEmulation", m)
	}
}

func (c *linkCmd) library(name string) {
	if name == "c" {
		c.clib = true
		return
	}
	if c.tool == "gcc" && name == "m" {
		return // math functions, if the program uses any, are reported as undefined
	}
	c.fail("err.noLib", name)
}

// gccValueOpts are gcc options whose value is the next argument.
var gccValueOpts = map[string]bool{"-L": true, "-I": true, "-D": true, "-U": true, "-include": true,
	"-isystem": true, "-Xlinker": true, "-MF": true, "-MT": true, "-MQ": true, "-x": true, "-e": true}

func parseGCC(args []string) *linkCmd {
	c := &linkCmd{tool: "gcc", command: append([]string{"gcc"}, args...)}
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() string {
			if i+1 < len(args) {
				i++
				return args[i]
			}
			c.fail("err.missingValue", a)
			return ""
		}
		switch {
		case !strings.HasPrefix(a, "-") || a == "-":
			c.inputs = append(c.inputs, a)
			if isCSource(a) {
				c.csrc = append(c.csrc, a)
			}
		case a == "-o":
			c.out = next()
		case strings.HasPrefix(a, "-o"):
			c.out = a[2:]
		case a == "-m32":
			c.bits = 32
		case a == "-m64":
			c.bits = 64
		case a == "-nostartfiles":
			c.noCRT = true
		case a == "-nostdlib":
			c.noCRT, c.noLibc = true, true
		case a == "-nodefaultlibs":
			c.noLibc = true
		case a == "-l":
			c.library(next())
		case strings.HasPrefix(a, "-l"):
			c.library(a[2:])
		case a == "-c" || a == "-S" || a == "-E":
			c.fail("err.unsupportedOption", "gcc", a)
		case a == "-x":
			next()
			c.fail("err.unsupportedOption", "gcc", a)
		case a == "-I" || a == "-D" || a == "-U":
			c.cflags = append(c.cflags, a, next())
		case gccValueOpts[a]:
			next()
		case strings.HasPrefix(a, "-g"):
			c.debug = true
		default:
			c.cflags = append(c.cflags, a)
			// -g, -O2, -Wall, -std=c99, -no-pie, -fno-pie, -static,
			// -march=..., -Wl,..., -I/-D/-L with the value attached, ...:
			// nothing nasmx86 needs to change for linking objects.
		}
	}
	return c
}

func isCSource(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".c")
}

// objectKind classifies an input file for choosing between nasmx86 and the
// Mac's own tool.
type objectKind int

const (
	kindMissing objectKind = iota
	kindELF                // an ELF object: nasmx86's
	kindMac                // Mach-O, a universal binary or an archive: the Mac's
	kindOther              // a source file or anything else
)

func classify(path string) objectKind {
	f, err := os.Open(path)
	if err != nil {
		return kindMissing
	}
	defer f.Close()
	head := make([]byte, 8)
	n, _ := f.Read(head)
	head = head[:n]
	switch {
	case bytes.HasPrefix(head, []byte("\x7fELF")):
		return kindELF
	case bytes.HasPrefix(head, []byte("!<arch>")), bytes.HasPrefix(head, []byte("--- !tapi")),
		bytes.HasPrefix(head, []byte{0xcf, 0xfa, 0xed, 0xfe}), bytes.HasPrefix(head, []byte{0xce, 0xfa, 0xed, 0xfe}),
		bytes.HasPrefix(head, []byte{0xca, 0xfe, 0xba, 0xbe}):
		return kindMac
	}
	return kindOther
}

// forMac reports whether a command line is for the Mac's own tool. nasmx86
// takes it when an input is an x86 ELF object. Otherwise gcc is the Mac's
// (gcc -o prog prog.c, gcc --version, gcc -E - ...) unless it is asked for
// 32-bit x86 code (-m32). ld is the Mac's when it has Mac objects or no
// inputs at all (ld -v); for missing files or sources nasmx86 explains.
func (c *linkCmd) forMac() bool {
	mac := len(c.inputs) == 0
	for _, in := range c.inputs {
		switch {
		case in == "-":
			mac = true // standard input: a build tool probing the compiler
		case classify(in) == kindELF:
			return false
		case classify(in) == kindMac:
			mac = true
		}
	}
	if c.tool == "gcc" {
		return c.bits != 32 || len(c.inputs) == 0
	}
	return mac
}

// macTool runs the Mac's own ld or gcc with the same arguments.
func macTool(name string, args []string) int {
	path, ok := pathfind.Other(name, "/usr/bin")
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", name)
		return 127
	}
	err := pathfind.Exec(path, args, os.Environ())
	fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
	return 126
}

func ldMain(args []string) int {
	c := parseLD(args)
	if c.forMac() {
		return macTool("ld", args)
	}
	return c.link()
}

func gccMain(args []string) int {
	c := parseGCC(args)
	if c.forMac() {
		return macTool("gcc", args)
	}
	return c.link()
}

// link checks the command line and the inputs, links the objects to see
// that they make a program, and writes the program file.
func (c *linkCmd) link() int {
	say := func(key string, args ...any) { fmt.Fprintln(os.Stderr, c.tool+": "+i18n.T(key, args...)) }
	if len(c.errs) > 0 {
		fmt.Fprintln(os.Stderr, c.tool+": "+c.errs[0])
		return runner.ExitBuild
	}
	if len(c.inputs) == 0 {
		say("err.noInput")
		return runner.ExitBuild
	}
	bits := c.bits
	if bits == 0 {
		bits = 64
	}
	objs := make([]string, len(c.inputs))  // the object files to link
	names := make([]string, len(c.inputs)) // and what to call them
	copy(objs, c.inputs)
	copy(names, c.inputs)
	if c.tool == "gcc" && len(c.csrc) > 0 {
		// Compile the C files for x86 Linux first, as gcc does.
		dir, err := os.MkdirTemp("", "nasmx86-cc-")
		if err != nil {
			say("err.internal", err)
			return runner.ExitBuild
		}
		defer os.RemoveAll(dir)
		for i, in := range c.inputs {
			if !isCSource(in) {
				continue
			}
			if _, err := os.Stat(in); err != nil {
				say("err.cantFind", in)
				return runner.ExitBuild
			}
			obj, err := cc.Compile(in, bits, c.cflags, c.debug, dir, os.Stderr)
			switch {
			case errors.Is(err, cc.ErrNoCompiler):
				say("err.noClang")
				return runner.ExitBuild
			case errors.Is(err, cc.ErrFloat):
				fmt.Fprintln(os.Stderr, "    "+i18n.T("hint.prefix")+i18n.T("hint.cFloat"))
				return runner.ExitBuild
			case err != nil:
				return runner.ExitBuild // clang printed why
			}
			objs[i], names[i] = obj, filepath.Base(obj)
		}
	}
	for i, in := range objs {
		if c.tool == "gcc" && isCSource(c.inputs[i]) {
			continue
		}
		switch classify(in) {
		case kindMissing:
			say("err.cantFind", in)
			return runner.ExitBuild
		case kindELF:
		default:
			say("err.notObject", in)
			if ext := strings.ToLower(filepath.Ext(in)); ext == ".s" || ext == ".asm" || ext == ".nasm" {
				fmt.Fprintln(os.Stderr, "    "+i18n.T("hint.prefix")+i18n.T("hint.assembleFirst", strings.TrimSuffix(in, filepath.Ext(in)), in))
			}
			return runner.ExitBuild
		}
	}
	if c.out == "" {
		c.out = "a.out"
	}
	cfg := runner.Config{
		Source: objs[0], Extra: objs[1:], Bits: bits, Tool: c.tool,
		Mode: runner.LD, Entry: c.entry, CLib: c.clib, Argv0: c.out, MaxSteps: runner.DefaultMaxSteps,
	}
	m := progfile.Manifest{Linker: progfile.LD, Entry: c.entry, CLib: c.clib, Objects: names, Command: c.command}
	if c.tool == "gcc" {
		if c.noCRT {
			cfg.CLib = !c.noLibc
		} else {
			cfg.Mode = runner.GCC
			m.Linker = progfile.GCC
		}
		m.CLib = cfg.CLib
	}
	p, code := runner.Build(cfg)
	if p == nil {
		return code
	}
	m.Bits = p.Bits
	p.Close()

	data := make([][]byte, len(objs))
	for i, in := range objs {
		var err error
		if data[i], err = os.ReadFile(in); err != nil {
			say("err.cantFind", in)
			return runner.ExitBuild
		}
	}
	if err := progfile.Write(c.out, m, data); err != nil {
		say("err.writeOutput", c.out, err)
		return runner.ExitBuild
	}
	return 0
}
