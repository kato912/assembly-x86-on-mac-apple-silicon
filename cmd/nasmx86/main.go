// Command nasmx86 assembles, links and runs x86 (32/64-bit) Linux NASM
// programs on any machine — no VM, no Docker, no Rosetta.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/debugger"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// version is overwritten at release time via -ldflags "-X main.version=...".
var version = "dev"

type options struct {
	cfg     runner.Config
	help    bool
	version bool
	trace   bool
	step    bool
}

// usageError is a command-line mistake; it is printed with the usage hint.
type usageError string

func (e usageError) Error() string { return string(e) }

func main() {
	i18n.Init()
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("err.prefix")+err.Error())
		fmt.Fprintln(os.Stderr, "try: nasmx86 --help")
		os.Exit(runner.ExitUsage)
	}
	switch {
	case opts.version:
		fmt.Println(versionString(version))
		return
	case opts.help:
		fmt.Print(i18n.T("usage"))
		return
	}
	os.Exit(run(opts))
}

// versionString formats the version; 0.x releases are marked as beta.
func versionString(v string) string {
	if strings.HasPrefix(v, "0.") {
		return "nasmx86 " + v + " (beta)"
	}
	return "nasmx86 " + v
}

func run(o *options) int {
	p, code := runner.Build(o.cfg)
	if p == nil {
		return code
	}
	defer p.Close()
	var err error
	switch {
	case o.step:
		in := io.Reader(os.Stdin)
		if tty, terr := debugger.OpenTerminal(); terr == nil {
			defer tty.Close()
			in = tty
		}
		return debugger.Interactive(debugger.NewSession(p), in, p.Cfg.Stderr)
	case o.trace:
		err = debugger.Trace(debugger.NewSession(p), p.Cfg.Stderr)
	default:
		err = p.CPU.Run()
	}
	return p.Finish(err)
}

// nasm options that take their value as the next argument.
var nasmValueOpts = map[string]bool{"-D": true, "-d": true, "-U": true, "-u": true, "-I": true,
	"-i": true, "-P": true, "-p": true, "-X": true, "-Z": true, "-E": false}

// parseArgs understands nasmx86's own options plus nasm-style options
// copied from class sheets. Everything after the .asm file is passed to
// the program, except nasmx86's own long options (use -- to pass those).
func parseArgs(args []string) (*options, error) {
	o := &options{cfg: runner.Config{MaxSteps: runner.DefaultMaxSteps}}
	var file string
	afterDashes := false
	value := func(i *int, name string) (string, error) {
		if *i+1 >= len(args) {
			return "", usageError(i18n.T("err.missingValue", name))
		}
		*i++
		return args[*i], nil
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if file != "" && afterDashes {
			o.cfg.Args = append(o.cfg.Args, a)
			continue
		}
		if a == "--" {
			afterDashes = true
			if file == "" && i+1 < len(args) {
				i++
				file = args[i]
			}
			continue
		}
		// nasmx86's own options are recognised anywhere.
		if handled, err := o.ownOption(args, &i, value); handled || err != nil {
			if err != nil {
				return nil, err
			}
			continue
		}
		if file != "" {
			// More source/object files directly after the first one are
			// linked with it; the first other argument starts the program's
			// own arguments.
			if len(o.cfg.Args) == 0 && isInputFile(a) {
				o.cfg.Extra = append(o.cfg.Extra, a)
			} else {
				o.cfg.Args = append(o.cfg.Args, a)
			}
			continue
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			file = a
			continue
		}
		// nasm-style options.
		switch {
		case a == "-f" || a == "--format" || (strings.HasPrefix(a, "-f") && len(a) > 2):
			f := strings.TrimPrefix(a, "-f")
			if a == "-f" || a == "--format" {
				v, err := value(&i, a)
				if err != nil {
					return nil, err
				}
				f = v
			}
			switch strings.ToLower(strings.TrimPrefix(f, "=")) {
			case "elf32", "elf":
				o.cfg.Bits = 32
			case "elf64", "elfx32":
				o.cfg.Bits = 64
			default:
				return nil, usageError(i18n.T("err.unknownFormat", f))
			}
		case a == "-o" || a == "-l" || a == "-F" || a == "-L" || a == "-MF" || a == "-MT" || a == "-MQ":
			if _, err := value(&i, a); err != nil { // ignored: nasmx86 manages its own files
				return nil, err
			}
		case strings.HasPrefix(a, "-o") || strings.HasPrefix(a, "-l") || strings.HasPrefix(a, "-F") ||
			a == "-g" || strings.HasPrefix(a, "-L") || strings.HasPrefix(a, "-M"):
			// ignored
		case nasmValueOpts[a]:
			v, err := value(&i, a)
			if err != nil {
				return nil, err
			}
			o.cfg.NasmArgs = append(o.cfg.NasmArgs, a, v)
		default:
			o.cfg.NasmArgs = append(o.cfg.NasmArgs, a)
		}
	}
	if o.help || o.version {
		return o, nil
	}
	if file == "" {
		return nil, usageError(i18n.T("err.noFile"))
	}
	o.cfg.Source = file
	return o, nil
}

// isInputFile reports whether a command-line argument names an assembly
// source or object file.
func isInputFile(a string) bool {
	switch strings.ToLower(filepath.Ext(a)) {
	case ".asm", ".nasm", ".s", ".o", ".obj":
		return !strings.HasPrefix(a, "-")
	}
	return false
}

// ownOption handles nasmx86's own flags. It reports whether args[*i] was one.
func (o *options) ownOption(args []string, i *int, value func(*int, string) (string, error)) (bool, error) {
	a := args[*i]
	name, val, hasVal := strings.Cut(a, "=")
	switch name {
	case "-h", "--help", "-help":
		o.help = true
	case "--version", "-v", "-V":
		o.version = true
	case "-32", "--32", "-m32":
		o.cfg.Bits = 32
	case "-64", "--64", "-m64":
		o.cfg.Bits = 64
	case "--trace":
		o.trace = true
	case "--step", "--debug":
		o.step = true
	case "--regs":
		o.cfg.ShowRegs = true
	case "--th", "--thai":
		i18n.Lang = "th"
	case "--en":
		i18n.Lang = "en"
	case "--lang":
		if !hasVal {
			v, err := value(i, name)
			if err != nil {
				return true, err
			}
			val = v
		}
		if strings.HasPrefix(strings.ToLower(val), "th") {
			i18n.Lang = "th"
		} else {
			i18n.Lang = "en"
		}
	case "--max-steps":
		if !hasVal {
			v, err := value(i, name)
			if err != nil {
				return true, err
			}
			val = v
		}
		n, err := strconv.ParseUint(strings.ReplaceAll(val, "_", ""), 10, 64)
		if err != nil {
			return true, usageError(i18n.T("err.badSteps", val))
		}
		o.cfg.MaxSteps = n
	default:
		return false, nil
	}
	return true, nil
}
