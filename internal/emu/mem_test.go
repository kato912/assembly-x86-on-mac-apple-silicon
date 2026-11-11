package emu

import (
	"errors"
	"testing"
)

func TestMemoryReadWrite(t *testing.T) {
	m := NewMemory()
	m.Map(0x1000, 0x2000, PermR|PermW, ".data", KindSection)

	for _, size := range []int{1, 2, 4, 8} {
		if err := m.Write(0x1100, size, 0x1122334455667788); err != nil {
			t.Fatal(err)
		}
		got, err := m.Read(0x1100, size)
		if err != nil {
			t.Fatal(err)
		}
		want := uint64(0x1122334455667788) & (^uint64(0) >> (64 - 8*size))
		if got != want {
			t.Errorf("size %d: got %#x want %#x", size, got, want)
		}
	}
	// Little-endian byte order.
	b, _ := m.ReadBytes(0x1100, 4)
	if b[0] != 0x88 || b[3] != 0x55 {
		t.Errorf("bytes % x", b)
	}
	// Across a page boundary.
	if err := m.Write(0x1ffd, 8, 0x0102030405060708); err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Read(0x1ffd, 8); v != 0x0102030405060708 {
		t.Errorf("cross-page read %#x", v)
	}
}

func TestMemoryFaults(t *testing.T) {
	m := NewMemory()
	m.Map(0x1000, 0x1000, PermR|PermX, ".text", KindSection)
	m.Map(0x3000, 0x1000, PermR|PermW, ".data", KindSection)

	var mf *MemFault
	if _, err := m.Read(0, 4); !errors.As(err, &mf) || mf.Region != nil || mf.Access != AccessRead {
		t.Errorf("read of NULL: %v", err)
	}
	if err := m.Write(0x1000, 4, 1); !errors.As(err, &mf) || mf.Region == nil || mf.Region.Name != ".text" {
		t.Errorf("write to .text: %v", err)
	}
	if _, err := m.Fetch(0x3000, make([]byte, 15)); !errors.As(err, &mf) || mf.Access != AccessExec {
		t.Errorf("exec of .data: %v", err)
	}
	// A 4-byte read straddling the end of .data must fault at the first bad byte.
	if _, err := m.Read(0x3ffe, 4); !errors.As(err, &mf) || mf.Addr != 0x4000 {
		t.Errorf("straddling read: %v", err)
	}
	// Fetch stops at the end of executable memory.
	m.Poke(0x1ff8, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	buf := make([]byte, 15)
	if n, err := m.Fetch(0x1ff8, buf); err != nil || n != 8 {
		t.Errorf("fetch near end: n=%d err=%v", n, err)
	}
}
