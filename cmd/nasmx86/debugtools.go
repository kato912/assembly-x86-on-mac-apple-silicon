package main

// nasmx86 as ddd and gdb: ddd ./hello & and gdb ./hello debug a program
// made by nasmx86's ld or gcc with the real DDD and gdb, the program
// running in nasmx86. Other uses go to the real DDD or gdb.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/gdbstub"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/pathfind"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/progfile"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// debuggable reports whether nasmx86 debugs path itself: a program file
// from its ld or gcc, or an assembly source file.
func debuggable(path string) bool {
	if progfile.Is(path) {
		return true
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".asm", ".s", ".nasm":
		st, err := os.Stat(path)
		return err == nil && !st.IsDir()
	}
	return false
}

func dddMain(args []string) int {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		if debuggable(a) {
			o := &options{cfg: runner.Config{Source: a, Debug: true, MaxSteps: runner.DefaultMaxSteps}, ddd: true}
			return run(o)
		}
		break
	}
	ddd, ok := gdbstub.FindDDD()
	if !ok {
		fmt.Fprintln(os.Stderr, "nasmx86: "+i18n.T("ddd.missing"))
		return runner.ExitNoNasm
	}
	env, ok := gdbstub.DDDEnv(os.Stderr)
	if !ok {
		fmt.Fprintln(os.Stderr, "nasmx86: "+i18n.T("ddd.noXQuartz"))
		return runner.ExitNoNasm
	}
	err := pathfind.Exec(ddd, args, env)
	fmt.Fprintf(os.Stderr, "ddd: %v\n", err)
	return 126
}

// gdbValueOpts are gdb options whose value is the next argument.
var gdbValueOpts = map[string]bool{"-x": true, "-command": true, "-ex": true, "-eval-command": true,
	"-ix": true, "-init-command": true, "-iex": true, "-init-eval-command": true, "-cd": true,
	"-d": true, "-directory": true, "-s": true, "-symbols": true, "-e": true, "-exec": true, "-se": true,
	"-c": true, "-core": true, "-p": true, "-pid": true, "-b": true, "-l": true, "-D": true,
	"-data-directory": true, "-tty": true, "-t": true}

func gdbMain(args []string) int {
	var opts, progArgs []string
	target := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		norm := a
		if strings.HasPrefix(a, "--") {
			norm = a[1:]
		}
		switch {
		case norm == "-args":
			if i+1 < len(args) {
				target, progArgs = args[i+1], args[i+2:]
			}
			i = len(args)
		case gdbValueOpts[norm]:
			opts = append(opts, a)
			if i+1 < len(args) {
				i++
				opts = append(opts, args[i])
			}
		case strings.HasPrefix(a, "-"):
			opts = append(opts, a)
		case target == "":
			target = a
		}
	}
	if target != "" && debuggable(target) {
		cfg := runner.Config{Source: target, Args: progArgs, Debug: true, MaxSteps: runner.DefaultMaxSteps}
		p, code := runner.Build(cfg)
		if p == nil {
			return code
		}
		defer p.Close()
		exe, err := p.WriteExecutable()
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.T("err.prefix")+i18n.T("err.internal", err))
			return runner.ExitBuild
		}
		return gdbstub.Run(p, exe, gdbstub.Options{GDB: true, GDBArgs: opts, Log: os.Stderr})
	}
	gdb, ok := pathfind.Other("gdb")
	if !ok {
		if gdb, ok = gdbstub.FindGDB(); !ok {
			fmt.Fprintln(os.Stderr, "nasmx86: "+i18n.T("gdb.missing"))
			return runner.ExitNoNasm
		}
	}
	err := pathfind.Exec(gdb, args, os.Environ())
	fmt.Fprintf(os.Stderr, "gdb: %v\n", err)
	return 126
}
