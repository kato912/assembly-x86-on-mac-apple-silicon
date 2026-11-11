package emu

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func linuxMachine(t *testing.T, bits int, src string) (*CPU, *Linux, *bytes.Buffer) {
	c := machine(t, bits, src)
	var out bytes.Buffer
	l := NewLinux(strings.NewReader(""), &out, &out)
	c.Sys = l
	return c, l, &out
}

func TestWriteExit32(t *testing.T) {
	c, _, out := linuxMachine(t, 32, `
	mov eax, 4
	mov ebx, 1
	mov ecx, msg
	mov edx, 3
	int 0x80
	mov esi, eax     ; bytes written
	mov eax, 1
	mov ebx, 7
	int 0x80
msg: db "hi", 10
`)
	err := c.Run()
	var ex *Exit
	if !errors.As(err, &ex) || ex.Code != 7 {
		t.Fatalf("got %v", err)
	}
	if out.String() != "hi\n" {
		t.Errorf("stdout %q", out.String())
	}
	expect(t, c, RSI, 3)
}

func TestWriteExit64(t *testing.T) {
	c, _, out := linuxMachine(t, 64, `
	default rel
	mov rax, 1
	mov rdi, 1
	lea rsi, [msg]
	mov rdx, 6
	syscall
	mov r12, rcx     ; syscall stores the return address in rcx
	mov rax, 60
	mov rdi, 300     ; exit status is truncated to 8 bits
	syscall
msg: db "hello", 10
`)
	err := c.Run()
	var ex *Exit
	if !errors.As(err, &ex) || ex.Code != 300&0xff {
		t.Fatalf("got %v", err)
	}
	if out.String() != "hello\n" {
		t.Errorf("stdout %q", out.String())
	}
	if c.R[R12] != testCode+0x18 {
		t.Errorf("rcx after syscall = %#x", c.R[R12])
	}
}

func TestBadBufferAndUnknownSyscall(t *testing.T) {
	c, l, _ := linuxMachine(t, 32, `
	mov eax, 4
	mov ebx, 1
	mov ecx, 0       ; NULL buffer
	mov edx, 5
	int 0x80
	mov esi, eax
	mov eax, 9999
	int 0x80
	mov edi, eax
`)
	var warnings []string
	l.Warn = func(key string, args ...any) { warnings = append(warnings, key) }
	err := c.Run()
	if ex, ok := err.(*Exit); !ok || !ex.ByHalt {
		t.Fatalf("got %v", err)
	}
	expect(t, c, RSI, uint64(uint32(-EFAULT&0xffffffff)))
	expect(t, c, RDI, uint64(uint32(-ENOSYS&0xffffffff)))
	if strings.Join(warnings, ",") != "warn.efault,warn.enosys" {
		t.Errorf("warnings %v", warnings)
	}
}

func TestInitialStack(t *testing.T) {
	for _, bits := range []int{32, 64} {
		c := New(bits, NewMemory(), nil)
		sp, argv, envp := SetupStack(c, []string{"./prog", "arg1"}, []string{"A=1"}, 0x1234)
		w := c.Width()
		if sp%16 != 0 {
			t.Errorf("%d: sp %#x not 16-byte aligned", bits, sp)
		}
		argc, _ := c.Mem.Read(sp, w)
		if argc != 2 || argv != sp+uint64(w) || envp != argv+3*uint64(w) {
			t.Errorf("%d: argc=%d argv=%#x envp=%#x", bits, argc, argv, envp)
		}
		p1, _ := c.Mem.Read(argv+uint64(w), w)
		s, _ := c.Mem.ReadBytes(p1, 5)
		if string(s) != "arg1\x00" {
			t.Errorf("%d: argv[1] = %q", bits, s)
		}
		if null, _ := c.Mem.Read(argv+2*uint64(w), w); null != 0 {
			t.Errorf("%d: argv not NULL terminated", bits)
		}
	}
}
