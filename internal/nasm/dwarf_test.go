package nasm

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDWARFLines(t *testing.T) {
	nasmPath, err := Find()
	if err != nil {
		t.Skip("nasm not installed")
	}
	for _, bits := range []int{32, 64} {
		dir := t.TempDir()
		src := filepath.Join(dir, "t.asm")
		os.WriteFile(src, []byte("section .text\nglobal _start\n_start:\n    nop        ; first\n\n    mov eax, 1\n"), 0o644)
		res, err := Assemble(nasmPath, src, bits, []string{"-F", "dwarf"}, dir, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		lm, err := DWARFLines(res.Object)
		if err != nil {
			t.Fatalf("%d-bit: %v", bits, err)
		}
		if filepath.Base(lm.Source) != "t.asm" {
			t.Errorf("%d-bit: source %q", bits, lm.Source)
		}
		for off, want := range map[uint64]Loc{0: {Line: 4, Text: "nop        ; first"}, 1: {Line: 6, Text: "mov eax, 1"}} {
			if got, ok := lm.Lookup(".text", off); !ok || got != want {
				t.Errorf("%d-bit: offset %d: %+v %v, want %+v", bits, off, got, ok, want)
			}
		}
		if places := lm.Places(6); len(places) != 1 || places[0] != (SecOff{".text", 1}) {
			t.Errorf("%d-bit: line 6 at %v", bits, places)
		}
	}
	// Without debug information there is no line table.
	dir := t.TempDir()
	src := filepath.Join(dir, "t.asm")
	os.WriteFile(src, []byte("section .text\nnop\n"), 0o644)
	obj := filepath.Join(dir, "t.o")
	if out, err := exec.Command(nasmPath, "-f", "elf64", "-o", obj, src).CombinedOutput(); err != nil {
		t.Fatalf("nasm: %v\n%s", err, out)
	}
	if _, err := DWARFLines(obj); err == nil {
		t.Error("expected no line table without -g")
	}
}
