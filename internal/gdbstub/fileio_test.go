package gdbstub

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"

	"testing"
)

// The program's standard I/O goes through gdb's File-I/O requests.
func TestConsoleFileIO(t *testing.T) {
	p, _ := build(t, `
section .data
ask db "name? "
hi  db "hi "
section .bss
buf resb 16
section .text
global _start
_start:
    mov rax, 1
    mov rdi, 1
    mov rsi, ask
    mov rdx, 6
    syscall
again:
    mov rax, 0          ; read(0, buf, 16)
    mov rdi, 0
    mov rsi, buf
    mov rdx, 16
    syscall
    mov r12, rax
    mov rax, 1
    mov rdi, 1
    mov rsi, hi
    mov rdx, 3
    syscall
    mov rax, 1
    mov rdi, 1
    mov rsi, buf
    mov rdx, r12
    syscall
    mov rax, 60
    mov rdi, 0
    syscall
`)
	a, b := net.Pipe()
	s := New(p)
	s.UseConsole()
	done := make(chan int)
	go func() { done <- s.Serve(b) }()
	c := &client{t: t, c: a, r: bufio.NewReader(a)}
	c.ask("QStartNoAckMode")
	buf := fioAddr(64)

	// write(1, "name? ", 6): gdb reads the bytes from the buffer and prints them.
	want := fmt.Sprintf("Fwrite,1,%x,6", buf)
	if r := c.ask("c"); r != want {
		t.Fatalf("continue = %q, want %q", r, want)
	}
	if r := c.ask(fmt.Sprintf("m%x,6", buf)); r != hex.EncodeToString([]byte("name? ")) {
		t.Errorf("buffer = %q", r)
	}
	// read(0, buf, 16): Ctrl-C while waiting stops the program at the read.
	if r := c.ask("F6"); r != fmt.Sprintf("Fread,0,%x,10", buf) {
		t.Fatalf("after write = %q", r)
	}
	if r := c.ask("F-1,4,C"); r != "T02thread:01;" {
		t.Fatalf("interrupted read = %q", r)
	}
	if in, err := p.CPU.Decode(p.CPU.RIP); err != nil || p.Disasm(in, p.CPU.RIP) != "syscall" {
		t.Errorf("stopped at %#x, want the read's syscall", p.CPU.RIP)
	}
	// Continuing reads again; gdb writes the input into the buffer.
	if r := c.ask("c"); r != fmt.Sprintf("Fread,0,%x,10", buf) {
		t.Fatalf("continue after Ctrl-C = %q", r)
	}
	if r := c.ask(fmt.Sprintf("M%x,4:%s", buf, hex.EncodeToString([]byte("Ann\n")))); r != "OK" {
		t.Errorf("M = %q", r)
	}
	if r := c.ask("F4"); r != fmt.Sprintf("Fwrite,1,%x,3", buf) {
		t.Fatalf("after read = %q", r)
	}
	if r := c.ask("F3"); r != fmt.Sprintf("Fwrite,1,%x,4", buf) {
		t.Fatalf("after hi = %q", r)
	}
	if r := c.ask(fmt.Sprintf("m%x,4", buf)); r != hex.EncodeToString([]byte("Ann\n")) {
		t.Errorf("echo buffer = %q", r)
	}
	if r := c.ask("F4"); r != "W00" {
		t.Errorf("exit = %q", r)
	}
	c.send("k")
	a.Close()
	<-done
}

func TestParseFileReply(t *testing.T) {
	for body, want := range map[string]string{
		"6": "6 0 false", "-1,4,C": "-1 4 true", "-1,9": "-1 9 false", "0;attach": "0 0 false", "1a": "26 0 false",
	} {
		ret, errno, ctrlC, err := parseFileReply(body)
		if got := fmt.Sprint(ret, errno, ctrlC); err != nil || got != want {
			t.Errorf("F%s: %s %v, want %s", body, got, err, want)
		}
	}
	if _, _, _, err := parseFileReply("x"); err == nil {
		t.Error("Fx: no error")
	}
}
