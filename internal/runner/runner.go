// Package runner glues the pieces together: assemble with nasm, link with
// the mini linker, load into the emulator, run, and report the outcome.
package runner

import (
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/libc"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/loader"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/progfile"
)

// DefaultMaxSteps is the default instruction limit (infinite-loop guard).
const DefaultMaxSteps = 50_000_000

// Exit codes used by nasmx86 itself (the program's own exit code is
// passed through unchanged).
const (
	ExitUsage    = 2
	ExitBuild    = 1
	ExitNoNasm   = 127
	ExitMaxSteps = 124
	signalOffset = 128
)

// Mode says how a program is linked.
type Mode int

const (
	// Auto is nasmx86's own way: start at _start, else run main like a C
	// program, and let any file call the built-in C library.
	Auto Mode = iota
	// LD links like GNU ld: start at _start (or Config.Entry), and use the
	// C library only with Config.CLib (ld -lc).
	LD
	// GCC links like gcc: start the C runtime, which calls main, with the
	// C library.
	GCC
)

// Config describes one run.
type Config struct {
	Source   string   // the main source file (.asm, a nasm .o object, or a program file from ld/gcc)
	Extra    []string // more source/object files linked with it
	Bits     int      // 32, 64, or 0 to auto-detect
	NasmArgs []string // extra options passed to nasm
	Args     []string // program arguments (argv[1:])
	MaxSteps uint64
	ShowRegs bool // print the registers when the program ends
	Debug    bool // keep DWARF debug info so WriteExecutable can give it to gdb

	Mode  Mode   // how to link (a program file sets it)
	Entry string // LD: the symbol to start at (ld -e); default _start
	CLib  bool   // LD: link the built-in C library too (ld -lc)
	Argv0 string // the program's argv[0]; default ./NAME
	Tool  string // name to start error messages with, such as "ld"; default nasmx86

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Program is a built program loaded into a fresh emulator.
type Program struct {
	Cfg       Config
	Bits      int
	Image     *loader.Image
	CPU       *emu.CPU
	Linux     *emu.Linux
	Lib       *libc.Lib    // the built-in C library, when the program calls it
	Files     []SourceFile // input files in link order; Files[0] is the main one
	InitialSP uint64
	ExitStub  uint64
	CRuntime  bool // main is called by a C runtime (returning from main exits)

	tmpDir string
}

// SourceFile is one input file of the program.
type SourceFile struct {
	Path   string
	Source []string      // text lines (1-based line N = Source[N-1]); nil for .o files
	Lines  *nasm.LineMap // address -> line map from the nasm listing; nil for .o files
}

// IsObject reports whether path names an already assembled object file:
// a .o or .obj file, or any ELF file.
func IsObject(path string) bool {
	if strings.EqualFold(filepath.Ext(path), ".o") || strings.EqualFold(filepath.Ext(path), ".obj") {
		return true
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 4)
	_, err = io.ReadFull(f, head)
	return err == nil && string(head) == "\x7fELF"
}

// ObjectBits returns 32 or 64 for an x86 ELF object, or 0.
func ObjectBits(path string) int {
	f, err := elf.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	switch f.Machine {
	case elf.EM_386:
		return 32
	case elf.EM_X86_64:
		return 64
	}
	return 0
}

func (p *Program) prefix() string {
	if p.Cfg.Tool != "" {
		return p.Cfg.Tool + ": "
	}
	return i18n.T("err.prefix")
}

func (p *Program) errorf(key string, args ...any) {
	fmt.Fprintln(p.Cfg.Stderr, p.prefix()+i18n.T(key, args...))
}

func (p *Program) hintf(key string, args ...any) {
	fmt.Fprintln(p.Cfg.Stderr, "    "+i18n.T("hint.prefix")+i18n.T(key, args...))
}

// Build assembles, links and loads the program. On failure it prints a
// message and returns a nil program and the exit code to use.
func Build(cfg Config) (*Program, int) {
	if cfg.Stdin == nil {
		cfg.Stdin = consoleStdin()
	}
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}
	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}
	p := &Program{}
	paths := append([]string{cfg.Source}, cfg.Extra...)
	names := paths // what to call each file in messages
	if progfile.Is(cfg.Source) {
		// A program file from nasmx86's ld or gcc: link its objects again.
		prog, err := progfile.Open(cfg.Source)
		if err != nil {
			fmt.Fprintln(cfg.Stderr, i18n.T("err.prefix")+err.Error())
			return nil, ExitBuild
		}
		if p.tmpDir, err = os.MkdirTemp("", "nasmx86-"); err != nil {
			fmt.Fprintln(cfg.Stderr, i18n.T("err.prefix")+i18n.T("err.internal", err))
			return nil, ExitBuild
		}
		if paths, err = prog.Extract(p.tmpDir); err != nil {
			p.Close()
			fmt.Fprintln(cfg.Stderr, i18n.T("err.prefix")+i18n.T("err.internal", err))
			return nil, ExitBuild
		}
		names = prog.Objects
		cfg.Mode, cfg.Bits, cfg.Entry, cfg.CLib = LD, prog.Bits, prog.Entry, prog.CLib
		if prog.Linker == progfile.GCC {
			cfg.Mode = GCC
		}
		if cfg.Argv0 == "" {
			cfg.Argv0 = cfg.Source // as the shell ran it, e.g. ./hello
		}
	}
	p.Cfg = cfg

	// Read the sources and work out 32/64-bit mode from them.
	srcs := make([][]byte, len(paths))
	for i, path := range paths {
		p.Files = append(p.Files, SourceFile{Path: names[i]})
		if IsObject(path) {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			p.errorf("err.readSource", path, err)
			return nil, ExitBuild
		}
		srcs[i] = src
		p.Files[i].Source = strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n")
	}
	bits := cfg.Bits
	if bits == 0 && cfg.Mode != Auto {
		bits = 64 // ld and gcc make 64-bit programs unless told -m elf_i386 / -m32
	}
	if bits == 0 {
		bits = 64
		found := false
		for i, path := range paths {
			if srcs[i] == nil {
				continue
			}
			b, why := nasm.DetectBitsFile(path, srcs[i], includeDirs(cfg.NasmArgs))
			if b == 16 {
				p.errorf("err.bits16", why)
				return nil, ExitBuild
			}
			if why != "default" {
				bits, found = b, true
				break
			}
		}
		if !found {
			for i, path := range paths {
				if srcs[i] == nil {
					if b := ObjectBits(path); b != 0 {
						bits = b
						break
					}
				}
			}
		}
	}
	if cfg.Mode != Auto {
		// Like ld, refuse objects of the other size, and say how to fix it.
		for i, path := range paths {
			if b := ObjectBits(path); b != 0 && b != bits {
				p.errorf("err.objBits", names[i], b, bits)
				p.hintf(fmt.Sprintf("hint.objBits%d.%s", b, cfg.Tool))
				p.Close()
				return nil, ExitBuild
			}
		}
	}

	var objects []string
	var nasmPath string
	for i, path := range paths {
		if srcs[i] == nil {
			objects = append(objects, path)
			// Source lines from the object's debug information, if any
			// (yasm -g dwarf2).
			if lm, err := nasm.DWARFLines(path); err == nil {
				p.Files[i].Lines = lm
				if data, err := os.ReadFile(lm.Source); err == nil {
					p.Files[i].Source = strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
				}
			}
			continue
		}
		if nasmPath == "" {
			var err error
			if nasmPath, err = nasm.Find(); err != nil {
				p.errorf("err.nasmMissing")
				fmt.Fprintln(cfg.Stderr, InstallHint())
				return nil, ExitNoNasm
			}
			if p.tmpDir, err = os.MkdirTemp("", "nasmx86-"); err != nil {
				p.errorf("err.internal", err)
				return nil, ExitBuild
			}
		}
		dir := filepath.Join(p.tmpDir, fmt.Sprint(i))
		if err := os.Mkdir(dir, 0o755); err != nil {
			p.Close()
			p.errorf("err.internal", err)
			return nil, ExitBuild
		}
		srcPath, extra := path, cfg.NasmArgs
		if cfg.Debug {
			// DWARF records the path nasm was given: make it absolute so gdb
			// finds the source from any directory.
			if abs, err := filepath.Abs(path); err == nil {
				srcPath = abs
			}
			extra = append([]string{"-F", "dwarf"}, extra...)
		}
		res, err := nasm.Assemble(nasmPath, srcPath, bits, extra, dir, cfg.Stderr)
		if err != nil {
			p.Close()
			var ee *nasm.ExitError
			if errors.As(err, &ee) {
				return nil, ee.Code
			}
			p.errorf("err.internal", err)
			return nil, ExitBuild
		}
		if f, err := os.Open(res.Listing); err == nil {
			p.Files[i].Lines, _ = nasm.ParseListing(f)
			f.Close()
		}
		objects = append(objects, res.Object)
	}

	opts := loader.Options{Externs: libc.Symbols(bits), Debug: cfg.Debug}
	switch cfg.Mode {
	case LD:
		opts.Entry = cfg.Entry
		if opts.Entry == "" {
			opts.Entry = "_start"
		}
		if !cfg.CLib {
			opts.Externs = nil // ld links the C library only when asked (-lc)
		}
	case GCC:
		opts.Entry = "main"
	}
	img, err := loader.LinkOpts(opts, objects...)
	if err != nil {
		p.Close()
		var ue *loader.UndefinedError
		var de *loader.DuplicateError
		switch {
		case errors.As(err, &ue):
			if cfg.Mode == LD && !cfg.CLib {
				p.errorf("err.undefinedLD", strings.Join(ue.Names, ", "))
			} else {
				p.errorf("err.undefined", strings.Join(ue.Names, ", "))
			}
			hint := i18n.T("hint.undefined", libc.Summary())
			switch {
			case cfg.Mode == LD && !cfg.CLib && libc.Symbols(bits)[ue.Names[0]] != 0:
				hint = i18n.T("hint.undefinedUseGCC", ue.Names[0], outName(cfg), strings.Join(names, " "))
			case len(paths) > 1:
				hint = i18n.T("hint.undefinedMulti", ue.Names[0])
			}
			fmt.Fprintln(cfg.Stderr, "    "+i18n.T("hint.prefix")+hint)
		case errors.As(err, &de):
			names := make([]string, len(de.Files))
			for i, f := range de.Files {
				names[i] = p.sourceFor(objects, f)
			}
			p.errorf("err.duplicate", de.Name, strings.Join(names, ", "))
		default:
			p.errorf("err.link", err)
		}
		return nil, ExitBuild
	}
	switch {
	case cfg.Mode == GCC && img.EntryMissing:
		p.Close()
		p.errorf("err.noMain")
		if _, ok := img.Lookup("_start"); ok {
			p.hintf("hint.startUseLD", outName(cfg), strings.Join(names, " "))
		}
		return nil, ExitBuild
	case cfg.Mode == GCC:
		if s, ok := img.Lookup("_start"); ok && s.Global {
			p.Close()
			p.errorf("err.startWithGCC")
			p.hintf("hint.startUseLD", outName(cfg), strings.Join(names, " "))
			return nil, ExitBuild
		}
	case img.EntryMissing:
		// GNU ld only warns, and the program starts at the top of .text.
		fmt.Fprintln(cfg.Stderr, p.prefix()+i18n.T("warn.noEntry", opts.Entry, img.Entry))
	}
	p.Image = img
	p.Bits = img.Bits
	p.load()
	return p, 0
}

// outName is the program name for hints: what -o will be called.
func outName(cfg Config) string {
	if cfg.Argv0 != "" {
		return filepath.Base(cfg.Argv0)
	}
	return strings.TrimSuffix(filepath.Base(cfg.Source), filepath.Ext(cfg.Source))
}

// sourceFor maps an object path back to the file the user named.
func (p *Program) sourceFor(objects []string, obj string) string {
	for i, o := range objects {
		if o == obj {
			return filepath.Base(p.Files[i].Path)
		}
	}
	return filepath.Base(obj)
}

// WriteExecutable writes the linked program as an ELF executable with its
// symbols and (when built with Config.Debug) DWARF line information, for
// gdb to read while it debugs the program running in nasmx86. The file is
// placed in nasmx86's temporary directory and removed by Close.
func (p *Program) WriteExecutable() (string, error) {
	dir := p.tmpDir
	if dir == "" {
		var err error
		if dir, err = os.MkdirTemp("", "nasmx86-"); err != nil {
			return "", err
		}
		p.tmpDir = dir
	}
	name := strings.TrimSuffix(filepath.Base(p.Cfg.Source), filepath.Ext(p.Cfg.Source))
	path := filepath.Join(dir, name)
	var extra []loader.ExtraRegion
	if p.Lib != nil {
		base, size, funcs := libc.Region(p.Bits)
		r := loader.ExtraRegion{Name: ".libc", Addr: base, Size: size}
		for _, f := range funcs {
			r.Symbols = append(r.Symbols, loader.ExtraSymbol{Name: f, Addr: libc.Symbols(p.Bits)[f]})
		}
		extra = append(extra, r)
	}
	return path, p.Image.WriteELF(path, extra)
}

// InstallHint explains how to install nasm on the current OS.
func InstallHint() string {
	switch runtime.GOOS {
	case "darwin", "windows":
		return i18n.T("install." + runtime.GOOS)
	}
	return i18n.T("install.linux")
}

// includeDirs extracts -I/-i directories from nasm arguments.
func includeDirs(args []string) []string {
	var dirs []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-I" || a == "-i" {
			if i+1 < len(args) {
				dirs = append(dirs, args[i+1])
				i++
			}
		} else if strings.HasPrefix(a, "-I") || strings.HasPrefix(a, "-i") {
			dirs = append(dirs, a[2:])
		}
	}
	return dirs
}

// Close releases the program's resources: files it left open (as process
// exit would) and nasmx86's temporary files.
func (p *Program) Close() {
	if p.Linux != nil {
		p.Linux.CloseAll()
	}
	if p.tmpDir != "" {
		os.RemoveAll(p.tmpDir)
		p.tmpDir = ""
	}
}

// Reload puts the program back in its initial state (the debugger's
// "run" restarts it): fresh memory, registers and open files.
func (p *Program) Reload() {
	if p.Linux != nil {
		p.Linux.CloseAll()
	}
	p.ExitStub = 0
	p.load()
}

func (p *Program) load() {
	mem := emu.NewMemory()
	p.Linux = emu.NewLinux(p.Cfg.Stdin, p.Cfg.Stdout, p.Cfg.Stderr)
	p.Linux.Warn = func(key string, args ...any) {
		fmt.Fprintln(p.Cfg.Stderr, "nasmx86: "+i18n.T(key, args...))
	}
	c := emu.New(p.Bits, mem, p.Linux)
	p.CPU = c
	p.Linux.InitBrk(c, p.Image.BrkStart)
	p.Lib = nil
	if len(p.Image.Externs) > 0 {
		p.Lib = libc.Install(c, p.Linux)
		p.Lib.Warn = p.Linux.Warn
	}
	for _, s := range p.Image.Sections {
		perm := emu.PermR
		if s.Write {
			perm |= emu.PermW
		}
		if s.Exec {
			perm |= emu.PermX
			c.Code = append(c.Code, emu.CodeRange{Start: s.Addr, End: s.End()})
		}
		size := s.Size
		if size == 0 {
			continue
		}
		mem.Map(s.Addr, size, perm, s.Name, emu.KindSection)
		if s.Data != nil {
			mem.Poke(s.Addr, s.Data)
		}
	}

	argv0 := p.Cfg.Argv0
	if argv0 == "" {
		argv0 = "./" + strings.TrimSuffix(filepath.Base(p.Cfg.Source), filepath.Ext(p.Cfg.Source)) // like ./hello
	}
	argv := append([]string{argv0}, p.Cfg.Args...)
	sp, argvPtr, envpPtr := emu.SetupStack(c, argv, os.Environ(), p.Image.Entry)
	c.RIP = p.Image.Entry
	p.CRuntime = p.Cfg.Mode == GCC || (p.Cfg.Mode == Auto && p.Image.EntryName == "main")
	if p.CRuntime {
		// Behave like a C runtime: main(argc, argv, envp), and returning
		// from main exits with the value in eax.
		p.ExitStub = emu.InstallExitStub(c)
		w := uint64(c.Width())
		if p.Bits == 64 {
			c.R[emu.RDI] = uint64(len(argv))
			c.R[emu.RSI] = argvPtr
			c.R[emu.RDX] = envpPtr
			sp -= 8 // keep the ABI's 16-byte alignment at function entry
		} else {
			sp -= 3 * w
			mem.Poke(sp, le32(uint64(len(argv))))
			mem.Poke(sp+4, le32(argvPtr))
			mem.Poke(sp+8, le32(envpPtr))
			sp -= w
		}
		mem.Poke(sp, leN(p.ExitStub, int(w)))
		c.R[emu.RSP] = sp
		if p.Lib != nil {
			p.Lib.ExitStub = p.ExitStub
		}
	}
	p.InitialSP = c.SP()
	if p.Cfg.MaxSteps == 0 {
		c.MaxSteps = 0
	} else {
		c.MaxSteps = p.Cfg.MaxSteps
	}
}

func le32(v uint64) []byte { return leN(v, 4) }

func leN(v uint64, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(v >> (8 * i))
	}
	return b
}

// Finish interprets the error that stopped the program, prints any
// message, and returns nasmx86's exit code.
func (p *Program) Finish(err error) int {
	code := p.finish(err)
	if p.Cfg.ShowRegs {
		fmt.Fprint(p.Cfg.Stderr, p.FormatRegs())
	}
	return code
}

func (p *Program) finish(err error) int {
	var ex *emu.Exit
	if errors.As(err, &ex) {
		if ex.ByHalt {
			fmt.Fprintln(p.Cfg.Stderr, "nasmx86: "+i18n.T("warn.halt"))
		}
		return ex.Code
	}
	var f *emu.Fault
	if errors.As(err, &f) {
		fmt.Fprintln(p.Cfg.Stderr, p.Describe(f))
		if p.Lib != nil {
			p.Lib.Crashed() // a crashed process never prints its buffered output
		}
		if f.Kind == emu.FaultMaxSteps {
			return ExitMaxSteps
		}
		return signalOffset + f.Signal()
	}
	p.errorf("err.internal", err)
	return ExitBuild
}

// Run builds and runs a program, returning the exit code.
func Run(cfg Config) int {
	p, code := Build(cfg)
	if p == nil {
		return code
	}
	defer p.Close()
	return p.Finish(p.CPU.Run())
}
