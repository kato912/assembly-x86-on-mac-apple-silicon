package emu

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	testCode  = 0x10000
	testStack = 0x200000
)

// asmBin assembles src as a flat binary loaded at testCode.
func asmBin(t *testing.T, bits int, src string) []byte {
	t.Helper()
	nasmPath, err := exec.LookPath("nasm")
	if err != nil {
		for _, p := range []string{"/opt/homebrew/bin/nasm", "/usr/local/bin/nasm"} {
			if _, e := os.Stat(p); e == nil {
				nasmPath = p
			}
		}
		if nasmPath == "" {
			t.Skip("nasm not installed")
		}
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "t.asm")
	out := filepath.Join(dir, "t.bin")
	head := "bits 64\n"
	if bits == 32 {
		head = "bits 32\n"
	}
	head += "org 0x10000\n"
	if err := os.WriteFile(in, []byte(head+src+"\nhlt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, err := exec.Command(nasmPath, "-f", "bin", "-o", out, in).CombinedOutput(); err != nil {
		t.Fatalf("nasm: %v\n%s", err, b)
	}
	code, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

// machine builds a CPU with code at testCode (rwx, so tests can keep data
// next to code) and a stack below testStack.
func machine(t *testing.T, bits int, src string) *CPU {
	code := asmBin(t, bits, src)
	m := NewMemory()
	m.Map(testCode, uint64(len(code)), PermR|PermW|PermX, ".text", KindSection)
	m.Poke(testCode, code)
	m.Map(testStack-0x10000, 0x10000, PermR|PermW, "stack", KindStack)
	c := New(bits, m, &fakeSys{})
	c.Code = []CodeRange{{testCode, testCode + uint64(len(code))}}
	c.RIP = testCode
	c.R[RSP] = testStack
	c.MaxSteps = 1_000_000
	return c
}

// run executes src until hlt and returns the CPU.
func run(t *testing.T, bits int, src string) *CPU {
	t.Helper()
	c := machine(t, bits, src)
	err := c.Run()
	var ex *Exit
	if !errors.As(err, &ex) || !ex.ByHalt {
		t.Fatalf("program did not reach hlt: %v (rip=%#x)", err, c.RIP)
	}
	return c
}

// runFault executes src and returns the fault it stops with.
func runFault(t *testing.T, bits int, src string) *Fault {
	t.Helper()
	c := machine(t, bits, src)
	err := c.Run()
	var f *Fault
	if !errors.As(err, &f) {
		t.Fatalf("expected a fault, got %v", err)
	}
	return f
}

type fakeSys struct{ calls []uint64 }

func (s *fakeSys) Syscall(c *CPU, int80 bool) error {
	s.calls = append(s.calls, c.R[RAX])
	c.R[RAX] = 0
	return nil
}

func expect(t *testing.T, c *CPU, reg int, want uint64) {
	t.Helper()
	if got := c.R[reg]; got != want {
		t.Errorf("%s = %#x, want %#x", RegName(reg, 64), got, want)
	}
}

func expectFlags(t *testing.T, c *CPU, which, want uint64) {
	t.Helper()
	if got := c.RFLAGS & which; got != want {
		t.Errorf("flags %s, want %s", FormatFlags(got), FormatFlags(want))
	}
}
