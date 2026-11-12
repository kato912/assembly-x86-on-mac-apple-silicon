package gdbstub

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

const testSrc = `
section .data
value dq 0x1122334455667788
section .text
global _start
_start:
    mov eax, 5
    call twice
    mov edi, eax
    mov eax, 60
    syscall
twice:
    add eax, eax
    ret
`

func build(t *testing.T, src string) (*runner.Program, string) {
	t.Helper()
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "prog.asm")
	os.WriteFile(path, []byte(src), 0o644)
	var out, errb bytes.Buffer
	p, code := runner.Build(runner.Config{Source: path, Debug: true, Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb})
	if p == nil {
		t.Fatalf("build failed (%d): %s", code, errb.String())
	}
	t.Cleanup(p.Close)
	exe, err := p.WriteExecutable()
	if err != nil {
		t.Fatal(err)
	}
	return p, exe
}

// client speaks gdb's side of the protocol.
type client struct {
	t *testing.T
	c net.Conn
	r *bufio.Reader
}

func (c *client) send(pkt string) {
	var sum byte
	for i := 0; i < len(pkt); i++ {
		sum += pkt[i]
	}
	fmt.Fprintf(c.c, "$%s#%02x", pkt, sum)
}

func (c *client) recv() string {
	for {
		b, err := c.r.ReadByte()
		if err != nil {
			c.t.Fatalf("read: %v", err)
		}
		if b != '$' {
			continue // acks
		}
		s, _ := c.r.ReadString('#')
		c.r.ReadByte()
		c.r.ReadByte()
		return strings.TrimSuffix(s, "#")
	}
}

func (c *client) ask(pkt string) string {
	c.send(pkt)
	r := c.recv()
	for strings.HasPrefix(r, "O") && r != "OK" { // console output before a stop reply
		r = c.recv()
	}
	return r
}

func TestProtocol(t *testing.T) {
	p, _ := build(t, testSrc)
	a, b := net.Pipe()
	done := make(chan int)
	go func() { done <- New(p).Serve(b) }()
	c := &client{t: t, c: a, r: bufio.NewReader(a)}

	if r := c.ask("qSupported:swbreak+"); !strings.Contains(r, "qXfer:features:read+") {
		t.Errorf("qSupported = %q", r)
	}
	if r := c.ask("QStartNoAckMode"); r != "OK" {
		t.Errorf("QStartNoAckMode = %q", r)
	}
	if r := c.ask("qXfer:features:read:target.xml:0,fff"); !strings.Contains(r, "i386:x86-64") || !strings.Contains(r, `name="rip"`) {
		t.Errorf("target.xml = %.80q", r)
	}
	if r := c.ask("?"); !strings.HasPrefix(r, "T05") {
		t.Errorf("? = %q", r)
	}
	// rip is register 16 in amd64 numbering: the entry point, little-endian
	entry := p.Image.Entry
	if r := c.ask("p10"); r != fmt.Sprintf("%016x", swap64(entry)) {
		t.Errorf("rip = %s, want entry %#x", r, entry)
	}
	g := c.ask("g")
	if len(g) != 2*536 {
		t.Errorf("g packet is %d bytes, want 536", len(g)/2)
	}
	sym, _ := p.Image.Lookup("value")
	if r := c.ask(fmt.Sprintf("m%x,8", sym.Addr)); r != "8877665544332211" {
		t.Errorf("m value = %q", r)
	}
	if r := c.ask(fmt.Sprintf("M%x,1:ff", sym.Addr)); r != "OK" {
		t.Errorf("M = %q", r)
	}
	twice, _ := p.Image.Lookup("twice")
	if r := c.ask(fmt.Sprintf("Z0,%x,1", twice.Addr)); r != "OK" {
		t.Errorf("Z0 = %q", r)
	}
	if r := c.ask("vCont;c"); !strings.Contains(r, "swbreak") {
		t.Errorf("continue to breakpoint = %q", r)
	}
	if r := c.ask("p10"); r != fmt.Sprintf("%016x", swap64(twice.Addr)) {
		t.Errorf("stopped at %s, want twice %#x", r, twice.Addr)
	}
	c.ask(fmt.Sprintf("z0,%x,1", twice.Addr))
	if r := c.ask("vCont;s:1"); !strings.HasPrefix(r, "T05") {
		t.Errorf("step = %q", r)
	}
	if r := c.ask("p0"); r != fmt.Sprintf("%016x", swap64(10)) { // rax = 5+5
		t.Errorf("rax after add = %s", r)
	}
	if r := c.ask("vCont;c"); r != "W0a" {
		t.Errorf("exit = %q, want W0a", r)
	}
	if r := c.ask("qfThreadInfo"); r != "l" {
		t.Errorf("threads after exit = %q", r)
	}
	if r := c.ask("vRun;"); !strings.HasPrefix(r, "T05") {
		t.Errorf("vRun = %q", r)
	}
	if r := c.ask("p10"); r != fmt.Sprintf("%016x", swap64(entry)) {
		t.Errorf("rip after vRun = %s", r)
	}
	c.ask("vKill;1")
	c.send("D")
	c.recv()
	a.Close()
	<-done
}

func swap64(v uint64) uint64 {
	var r uint64
	for i := 0; i < 8; i++ {
		r = r<<8 | (v>>(8*i))&0xff
	}
	return r
}

// With a real gdb installed, a session finds source lines and labels.
func TestWithRealGDB(t *testing.T) {
	gdb, ok := FindGDB()
	if !ok {
		t.Skip("no x86 gdb installed")
	}
	p, exe := build(t, testSrc)
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		if c, err := ln.Accept(); err == nil {
			New(p).Serve(c)
		}
	}()
	out, err := exec.Command(gdb, "-nx", "--batch", "-ex", "set confirm off", exe,
		"-ex", "target extended-remote "+ln.Addr().String(),
		"-ex", "break twice", "-ex", "continue", "-ex", "bt", "-ex", "p $eax", "-ex", "continue").CombinedOutput()
	got := string(out)
	for _, w := range []string{"Breakpoint 1, twice () at", "prog.asm:13", "#1  ", "_start", "$1 = 5", "exited with code 012"} {
		if !strings.Contains(got, w) {
			t.Errorf("gdb output lacks %q (err %v):\n%s", w, err, got)
		}
	}
}
