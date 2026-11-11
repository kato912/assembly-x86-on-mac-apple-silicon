package loader

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
)

func assemble(t *testing.T, bits int, src string) string {
	t.Helper()
	path, err := nasm.Find()
	if err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	asm := filepath.Join(dir, "t.asm")
	if err := os.WriteFile(asm, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := nasm.Assemble(path, asm, bits, nil, dir, io.Discard)
	if err != nil {
		t.Fatalf("nasm: %v", err)
	}
	return res.Object
}

func TestLink32(t *testing.T) {
	obj := assemble(t, 32, `
section .data
msg db "hi", 10
ptr dd msg
section .bss
buf resb 16
section .text
global _start
_start:
    mov ecx, msg
    call f
f:  ret
`)
	img, err := Link(obj)
	if err != nil {
		t.Fatal(err)
	}
	text := img.SectionByName(".text")
	data := img.SectionByName(".data")
	bss := img.SectionByName(".bss")
	if text.Addr != Base32+0x1000 {
		t.Errorf(".text at %#x", text.Addr)
	}
	if data.Addr%0x1000 != 0 || data.Addr <= text.Addr {
		t.Errorf(".data at %#x", data.Addr)
	}
	if bss.Addr < data.End() {
		t.Errorf(".bss at %#x overlaps .data", bss.Addr)
	}
	if img.Entry != text.Addr || img.EntryName != "_start" {
		t.Errorf("entry %#x %q", img.Entry, img.EntryName)
	}
	// mov ecx, msg  => B9 <abs32 msg>
	if got := binary.LittleEndian.Uint32(text.Data[1:]); uint64(got) != data.Addr {
		t.Errorf("R_386_32 in .text: %#x want %#x", got, data.Addr)
	}
	// ptr dd msg
	if got := binary.LittleEndian.Uint32(data.Data[3:]); uint64(got) != data.Addr {
		t.Errorf("R_386_32 in .data: %#x", got)
	}
	// call f is PC-relative to a local label: already resolved by nasm (rel 0).
	if s, ok := img.Lookup("buf"); !ok || s.Addr != bss.Addr {
		t.Errorf("buf symbol %+v", s)
	}
}

func TestLink64(t *testing.T) {
	obj := assemble(t, 64, `
section .data
x dq 5
p dq x
section .text
global _start
_start:
    mov rax, [rel x]     ; R_X86_64_PC32
    mov ebx, x           ; R_X86_64_32
    mov rcx, [x]         ; R_X86_64_32S
    mov rdx, x           ; R_X86_64_64 (mov r64, imm64)
`)
	img, err := Link(obj)
	if err != nil {
		t.Fatal(err)
	}
	text := img.SectionByName(".text")
	data := img.SectionByName(".data")
	le := binary.LittleEndian
	// mov rax,[rip+d]: 48 8B 05 d32 (7 bytes)
	if rel := int32(le.Uint32(text.Data[3:])); text.Addr+7+uint64(int64(rel)) != data.Addr {
		t.Errorf("PC32: rel=%#x", rel)
	}
	// mov ebx, imm32: BB imm32 at offset 7
	if v := le.Uint32(text.Data[8:]); uint64(v) != data.Addr {
		t.Errorf("R_X86_64_32: %#x", v)
	}
	// mov rcx,[abs32]: 48 8B 0C 25 d32 at offset 12
	if v := le.Uint32(text.Data[16:]); uint64(v) != data.Addr {
		t.Errorf("R_X86_64_32S: %#x", v)
	}
	// mov rdx, imm64: 48 BA imm64 at offset 20
	if v := le.Uint64(text.Data[22:]); v != data.Addr {
		t.Errorf("R_X86_64_64: %#x", v)
	}
	if v := le.Uint64(data.Data[8:]); v != data.Addr {
		t.Errorf("data R_X86_64_64: %#x", v)
	}
}

func TestUndefined(t *testing.T) {
	obj := assemble(t, 64, `
extern printf
global main
section .text
main: call printf
`)
	_, err := Link(obj)
	var ue *UndefinedError
	if !errors.As(err, &ue) || len(ue.Names) != 1 || ue.Names[0] != "printf" {
		t.Fatalf("got %v", err)
	}
}

func TestEntryFallback(t *testing.T) {
	obj := assemble(t, 32, "section .text\nnop\nmain: nop\n")
	img, err := Link(obj)
	if err != nil {
		t.Fatal(err)
	}
	if img.EntryName != "main" || img.Entry != img.SectionByName(".text").Addr+1 {
		t.Errorf("entry %#x %q", img.Entry, img.EntryName)
	}
	obj = assemble(t, 32, "section .text\nnop\nnop\n")
	img, err = Link(obj)
	if err != nil {
		t.Fatal(err)
	}
	if img.EntryName != "" || img.Entry != img.SectionByName(".text").Addr {
		t.Errorf("entry %#x %q", img.Entry, img.EntryName)
	}
}
