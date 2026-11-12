package loader

import (
	"debug/dwarf"
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The written executable has the image's sections at their addresses,
// the symbols, and DWARF line tables that map addresses to source lines.
func TestWriteELFWithDWARF(t *testing.T) {
	nasm, err := exec.LookPath("nasm")
	if err != nil {
		t.Skip("nasm not installed")
	}
	for _, bits := range []int{32, 64} {
		dir := t.TempDir()
		src := filepath.Join(dir, "p.asm")
		os.WriteFile(src, []byte(`section .data
msg db "hi", 10
msg_end:
section .bss
buf resb 8
section .text
global _start
_start:
    nop
    mov eax, 1
.loop:
    dec eax
    jnz .loop
    nop
`), 0o644)
		obj := filepath.Join(dir, "p.o")
		format := "elf64"
		if bits == 32 {
			format = "elf32"
		}
		if out, err := exec.Command(nasm, "-f", format, "-g", "-F", "dwarf", "-o", obj, src).CombinedOutput(); err != nil {
			t.Fatalf("nasm: %v\n%s", err, out)
		}
		img, err := LinkOpts(Options{Debug: true}, obj)
		if err != nil {
			t.Fatal(err)
		}
		exe := filepath.Join(dir, "p")
		if err := img.WriteELF(exe, []ExtraRegion{{Name: ".libc", Addr: 0x7f000000, Size: 32, Symbols: []ExtraSymbol{{"printf", 0x7f000000}}}}); err != nil {
			t.Fatal(err)
		}
		f, err := elf.Open(exe)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if f.Type != elf.ET_EXEC || f.Entry != img.Entry {
			t.Errorf("%d-bit: type %v entry %#x, want ET_EXEC %#x", bits, f.Type, f.Entry, img.Entry)
		}
		text := f.Section(".text")
		if text == nil || text.Addr != img.Sections[0].Addr {
			t.Fatalf("%d-bit: .text missing or misplaced: %+v", bits, text)
		}
		code, _ := text.Data()
		if code[0] != 0x90 {
			t.Errorf("%d-bit: .text starts with %#x, want nop", bits, code[0])
		}
		syms, _ := f.Symbols()
		found := map[string]uint64{}
		for _, s := range syms {
			found[s.Name] = s.Value
		}
		for _, n := range []string{"_start", "_start.loop", "msg", "msg_end", "buf", "printf"} {
			if _, ok := found[n]; !ok {
				t.Errorf("%d-bit: symbol %s missing", bits, n)
			}
		}
		d, err := f.DWARF()
		if err != nil {
			t.Fatalf("%d-bit: DWARF: %v", bits, err)
		}
		r := d.Reader()
		cu, _ := r.Next()
		lr, err := d.LineReader(cu)
		if err != nil || lr == nil {
			t.Fatalf("%d-bit: line table: %v", bits, err)
		}
		var e dwarf.LineEntry
		lines := map[uint64]int{}
		for lr.Next(&e) == nil {
			lines[e.Address] = e.Line
		}
		if got := lines[found["_start.loop"]]; got != 12 {
			t.Errorf("%d-bit: .loop at %#x maps to line %d, want 12 (table %v)", bits, found["_start.loop"], got, lines)
		}
	}
}
