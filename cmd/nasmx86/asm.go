package main

// nasmx86 as nasm-x86 and yasm-x86: the real nasm and yasm, under names
// that say the code is x86 (a Mac's own processor is Arm). They take the
// command lines of course sheets for either assembler:
//
//	nasm-x86 -f elf64 -g -F dwarf -o hello.o hello.asm
//	yasm-x86 -f elf64 -g dwarf2 -o hello.o hello.s
//
// The other assembler's debug options are translated (-g dwarf2 for nasm,
// -F dwarf for yasm), and without -f the program's own size picks elf32
// or elf64 (nasm alone would write a flat binary, not an object file).

import (
	"fmt"
	"os"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/pathfind"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// asmValueOpts are each assembler's options whose value is the next argument.
var asmValueOpts = map[string]map[string]bool{
	"nasm": {"-o": true, "-l": true, "-f": true, "-F": true, "-I": true, "-i": true, "-P": true, "-p": true,
		"-D": true, "-d": true, "-U": true, "-u": true, "-L": true, "-X": true, "-Z": true,
		"-MF": true, "-MT": true, "-MQ": true, "--before": true, "--prefix": true, "--suffix": true},
	"yasm": {"-o": true, "-l": true, "-f": true, "-p": true, "-r": true, "-m": true, "-L": true,
		"-I": true, "-D": true, "-U": true, "-P": true, "-X": true, "-E": true},
}

// asmArgs turns a course command line into the given assembler's, and
// returns the source file it names ("" if none).
func asmArgs(assembler string, args []string) (out []string, src string, hasFormat bool) {
	values := asmValueOpts[assembler]
	debug := func() []string {
		if assembler == "nasm" {
			return []string{"-g", "-F", "dwarf"}
		}
		return []string{"-g", "dwarf2"}
	}
	addDebug := func() {
		for _, a := range out {
			if a == "-g" {
				return // already asked for (e.g. -g dwarf2 and -F dwarf together)
			}
		}
		out = append(out, debug()...)
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-g":
			// yasm's -g takes the format (-g dwarf2); nasm's -g does not.
			if i+1 < len(args) && debugFormats[strings.ToLower(args[i+1])] {
				i++
			}
			addDebug()
		case strings.HasPrefix(a, "-g") && debugFormats[strings.ToLower(a[2:])]: // -gdwarf2
			addDebug()
		case a == "-F" || (strings.HasPrefix(a, "-F") && len(a) > 2): // nasm's debug format
			if a == "-F" && i+1 < len(args) {
				i++
			}
			addDebug()
		case a == "-f" || a == "--oformat" || (strings.HasPrefix(a, "-f") && len(a) > 2) || strings.HasPrefix(a, "--oformat="):
			hasFormat = true
			out = append(out, a)
			if (a == "-f" || a == "--oformat") && i+1 < len(args) {
				i++
				out = append(out, args[i])
			}
		case values[a]:
			out = append(out, a)
			if i+1 < len(args) {
				i++
				out = append(out, args[i])
			}
		case !strings.HasPrefix(a, "-") && src == "":
			src = a
			out = append(out, a)
		default:
			out = append(out, a)
		}
	}
	return out, src, hasFormat
}

func asmMain(assembler string, args []string) int {
	out, src, hasFormat := asmArgs(assembler, args)
	if src != "" && !hasFormat {
		// An object file for a Linux program of the source's own size.
		bits := 64
		if data, err := os.ReadFile(src); err == nil {
			if b, _ := nasm.DetectBitsFile(src, data, includeDirs(out)); b == 32 {
				bits = 32
			}
		}
		out = append([]string{"-f", fmt.Sprintf("elf%d", bits)}, out...)
	}
	path, ok := pathfind.Other(assembler, toolDirs()...)
	if !ok {
		fmt.Fprintln(os.Stderr, assembler+"-x86: "+i18n.T("err.asmMissing", assembler))
		return runner.ExitNoNasm
	}
	err := pathfind.Exec(path, out, os.Environ())
	fmt.Fprintf(os.Stderr, "%s-x86: %v\n", assembler, err)
	return 126
}

// toolDirs are where Homebrew puts programs, in case PATH lacks them.
func toolDirs() []string {
	return []string{"/opt/homebrew/bin", "/usr/local/bin", "/home/linuxbrew/.linuxbrew/bin"}
}

// includeDirs extracts -I directories from assembler arguments.
func includeDirs(args []string) []string {
	var dirs []string
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case (a == "-I" || a == "-i") && i+1 < len(args):
			dirs = append(dirs, args[i+1])
			i++
		case strings.HasPrefix(a, "-I") && len(a) > 2:
			dirs = append(dirs, a[2:])
		}
	}
	return dirs
}
