// Package runner glues the pieces together: assemble with nasm, link with
// the mini linker, load into the emulator, run, and report the outcome.
package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/emu"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/loader"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
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

// Config describes one run.
type Config struct {
	Source   string
	Bits     int      // 32, 64, or 0 to auto-detect
	NasmArgs []string // extra options passed to nasm
	Args     []string // program arguments (argv[1:])
	MaxSteps uint64
	ShowRegs bool // print the registers when the program ends

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
	Source    []string // source file lines (1-based index = line-1)
	Lines     *nasm.LineMap
	InitialSP uint64
	ExitStub  uint64

	tmpDir string
}

func (p *Program) errorf(key string, args ...any) {
	fmt.Fprintln(p.Cfg.Stderr, i18n.T("err.prefix")+i18n.T(key, args...))
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
	p := &Program{Cfg: cfg}

	src, err := os.ReadFile(cfg.Source)
	if err != nil {
		p.errorf("err.readSource", cfg.Source, err)
		return nil, ExitBuild
	}
	p.Source = strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n")

	bits := cfg.Bits
	if bits == 0 {
		var why string
		bits, why = nasm.DetectBitsFile(cfg.Source, src, includeDirs(cfg.NasmArgs))
		if bits == 16 {
			p.errorf("err.bits16", why)
			return nil, ExitBuild
		}
	}

	nasmPath, err := nasm.Find()
	if err != nil {
		p.errorf("err.nasmMissing")
		fmt.Fprintln(cfg.Stderr, InstallHint())
		return nil, ExitNoNasm
	}
	p.tmpDir, err = os.MkdirTemp("", "nasmx86-")
	if err != nil {
		p.errorf("err.internal", err)
		return nil, ExitBuild
	}
	res, err := nasm.Assemble(nasmPath, cfg.Source, bits, cfg.NasmArgs, p.tmpDir, cfg.Stderr)
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
		p.Lines, _ = nasm.ParseListing(f)
		f.Close()
	}

	img, err := loader.Link(res.Object)
	if err != nil {
		p.Close()
		var ue *loader.UndefinedError
		if errors.As(err, &ue) {
			p.errorf("err.undefined", strings.Join(ue.Names, ", "))
		} else {
			p.errorf("err.link", err)
		}
		return nil, ExitBuild
	}
	p.Image = img
	p.Bits = img.Bits
	p.load()
	return p, 0
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

func (p *Program) load() {
	mem := emu.NewMemory()
	p.Linux = emu.NewLinux(p.Cfg.Stdin, p.Cfg.Stdout, p.Cfg.Stderr)
	p.Linux.Warn = func(key string, args ...any) {
		fmt.Fprintln(p.Cfg.Stderr, "nasmx86: "+i18n.T(key, args...))
	}
	c := emu.New(p.Bits, mem, p.Linux)
	p.CPU = c
	p.Linux.InitBrk(c, p.Image.BrkStart)
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

	argv0 := "./" + strings.TrimSuffix(filepath.Base(p.Cfg.Source), filepath.Ext(p.Cfg.Source))
	argv := append([]string{argv0}, p.Cfg.Args...)
	sp, argvPtr, envpPtr := emu.SetupStack(c, argv, os.Environ(), p.Image.Entry)
	c.RIP = p.Image.Entry
	if p.Image.EntryName == "main" {
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
