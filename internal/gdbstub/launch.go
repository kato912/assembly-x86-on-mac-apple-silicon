package gdbstub

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/i18n"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

// Options configure Run.
type Options struct {
	Addr string    // listen address, e.g. "localhost:1234"; "" picks a free port
	DDD  bool      // start DDD (with gdb) connected to the server
	Log  io.Writer // messages for the user (the terminal)
}

// Run serves gdb for program p, whose executable (for gdb's symbols) is
// at exe. It returns nasmx86's exit code.
func Run(p *runner.Program, exe string, o Options) int {
	if o.Log == nil {
		o.Log = os.Stderr
	}
	say := func(key string, args ...any) { fmt.Fprintln(o.Log, "nasmx86: "+i18n.T(key, args...)) }

	var gdb, ddd string
	if o.DDD {
		var ok bool
		if ddd, ok = FindDDD(); !ok {
			say("ddd.missing")
			return runner.ExitNoNasm
		}
		if gdb, ok = FindGDB(); !ok {
			say("gdb.missing")
			return runner.ExitNoNasm
		}
	}
	addr := o.Addr
	if addr == "" {
		addr = "localhost:0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		say("gdb.listen", addr, err)
		return runner.ExitBuild
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	target := fmt.Sprintf("localhost:%d", port)

	conns := make(chan net.Conn, 1)
	go func() {
		if c, err := ln.Accept(); err == nil {
			conns <- c
		}
		close(conns)
	}()

	var dddDone chan error
	var dddOut bytes.Buffer
	if o.DDD {
		// DDD runs "gdb" = a wrapper that starts the x86 gdb with a script
		// connecting it to nasmx86 (the name must contain "gdb" for DDD).
		dir := filepath.Dir(exe)
		script := filepath.Join(dir, "nasmx86.gdb")
		wrapper := filepath.Join(dir, "nasmx86-gdb")
		os.WriteFile(script, []byte(gdbScript(target)), 0o644)
		os.WriteFile(wrapper, []byte("#!/bin/sh\nexec "+shellQuote(gdb)+" -x "+shellQuote(script)+" \"$@\"\n"), 0o755)
		env, ok := x11Env()
		if !ok {
			say("ddd.noXQuartz")
			return runner.ExitNoNasm
		}
		cmd := exec.Command(ddd, "--gdb", "--debugger", wrapper, exe)
		cmd.Env = utf8Locale(env)
		cmd.Stdout, cmd.Stderr = &dddOut, &dddOut
		if err := cmd.Start(); err != nil {
			say("ddd.failed", err, "")
			return runner.ExitBuild
		}
		say("ddd.started", target)
		dddDone = make(chan error, 1)
		go func() { dddDone <- cmd.Wait() }()
		defer func() {
			if cmd.ProcessState == nil {
				cmd.Process.Kill()
			}
		}()
	} else {
		g, ok := FindGDB()
		if !ok {
			g = "x86_64-elf-gdb"
		}
		say("gdb.waiting", target, g, exe, target)
	}

	select {
	case c, ok := <-conns:
		if !ok {
			return runner.ExitBuild
		}
		s := New(p)
		if os.Getenv("NASMX86_GDB_LOG") != "" {
			s.Log = o.Log
		}
		code := s.Serve(c)
		if dddDone != nil {
			<-dddDone
		}
		return code
	case err := <-dddDone:
		// DDD ended before gdb connected: show why (no X server, ...).
		out := strings.TrimSpace(dddOut.String())
		if strings.Contains(out, "display") || strings.Contains(out, "DISPLAY") {
			say("ddd.noDisplay")
		}
		say("ddd.failed", err, out)
		return runner.ExitBuild
	}
}

func gdbScript(target string) string {
	return `# written by nasmx86: connect to the program running in nasmx86
set confirm off
set pagination off
set disassembly-flavor intel
target extended-remote ` + target + `
`
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// utf8Locale makes DDD read sources as UTF-8. In the C locale (a shell
// with no LANG) DDD 3.4 counts each byte of a non-ASCII character, such as
// a Thai comment or an em dash, as a character of its own, so its
// execution arrow and stop signs land on the wrong line or vanish.
func utf8Locale(env []string) []string {
	get := func(key string) string {
		v := ""
		for _, kv := range env {
			if k, val, ok := strings.Cut(kv, "="); ok && k == key {
				v = val
			}
		}
		return v
	}
	ctype := get("LC_ALL")
	if ctype == "" {
		ctype = get("LC_CTYPE")
	}
	if ctype == "" {
		ctype = get("LANG")
	}
	norm := strings.ToLower(strings.ReplaceAll(ctype, "-", ""))
	if strings.Contains(norm, "utf8") {
		return env
	}
	key, loc := "LC_CTYPE", "C.UTF-8"
	if runtime.GOOS == "darwin" {
		loc = "en_US.UTF-8"
	}
	if get("LC_ALL") != "" {
		key = "LC_ALL" // it overrides LC_CTYPE
	}
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if !strings.HasPrefix(kv, key+"=") {
			out = append(out, kv)
		}
	}
	return append(out, key+"="+loc)
}

// x11Env returns the environment for an X11 program such as DDD. On macOS
// it starts XQuartz when no X display is set (a fresh install sets
// DISPLAY only after logging out and in again).
func x11Env() ([]string, bool) {
	env := os.Environ()
	if os.Getenv("DISPLAY") != "" || runtime.GOOS != "darwin" {
		return env, true
	}
	app := ""
	for _, a := range []string{"/Applications/Utilities/XQuartz.app", "/Applications/XQuartz.app"} {
		if _, err := os.Stat(a); err == nil {
			app = a
		}
	}
	if app == "" {
		return nil, false
	}
	exec.Command("open", "-g", "-a", app).Run()
	for i := 0; i < 100; i++ { // wait up to 10 s for its display socket
		if _, err := os.Stat("/tmp/.X11-unix/X0"); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return append(env, "DISPLAY=:0"), true
}

// toolDirs are where Homebrew puts programs, in case PATH lacks them.
var toolDirs = []string{"/opt/homebrew/bin", "/usr/local/bin", "/home/linuxbrew/.linuxbrew/bin"}

func find(name string) (string, bool) {
	if p, err := exec.LookPath(name); err == nil {
		return p, true
	}
	for _, d := range toolDirs {
		p := filepath.Join(d, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}
	return "", false
}

// FindDDD finds the DDD front-end.
func FindDDD() (string, bool) {
	if p := os.Getenv("NASMX86_DDD"); p != "" {
		return p, true
	}
	return find("ddd")
}

// FindGDB finds a gdb that can debug x86 code: a cross gdb such as
// Homebrew's x86_64-elf-gdb, or a gdb built for several architectures.
func FindGDB() (string, bool) {
	if p := os.Getenv("NASMX86_GDB"); p != "" {
		return p, true
	}
	for _, n := range []string{"x86_64-elf-gdb", "x86_64-linux-gnu-gdb", "gdb-multiarch"} {
		if p, ok := find(n); ok {
			return p, true
		}
	}
	if p, ok := find("gdb"); ok {
		// A native gdb on an Arm Mac may not know x86 at all.
		out, err := exec.Command(p, "-nx", "--batch", "-ex", "set architecture i386:x86-64").CombinedOutput()
		if err == nil && !bytes.Contains(out, []byte("Undefined")) {
			return p, true
		}
	}
	return "", false
}
