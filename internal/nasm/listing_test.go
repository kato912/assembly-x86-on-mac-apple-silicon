package nasm

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A listing as produced by NASM (format shared by 2.x and 3.x).
const sampleListing = `     1                                  %include "inc.inc"
     1                              <1> %macro exit 1
     2                              <1>     mov eax, 1
     3                              <1>     mov ebx, %1
     4                              <1>     int 0x80
     5                              <1> %endmacro
     2                                  section .data
     3 00000000 48656C6C6F2C20576F-     msg db "Hello, World! long string here", 10
     3 00000009 726C6421206C6F6E67-
     3 00000012 20737472696E672068-
     3 0000001B 6572650A
     4                                  len equ $ - msg
     5 0000001F 00<rep 4h>              buf: times 4 db 0
     6                                  section .bss
     7 00000000 <res 28h>               x resd 10
     8                                  section .text
     9                                  global _start
    10                                  _start:
    11 00000000 B804000000                  mov eax, 4
    12 00000005 BB01000000                  mov ebx, 1
    13 0000000A B9[00000000]                mov ecx, msg
    14 0000000F BA1F000000                  mov edx, len
    15 00000014 CD80                        int 0x80
    16 00000016 A1[00000000]                mov eax, [x]
    17 0000001B 90<rep 2h>                  times 2 nop
    18                                      exit 0
     2 0000001D B801000000          <1>  mov eax, 1
     3 00000022 BB00000000          <1>  mov ebx, %1
     4 00000027 CD80                <1>  int 0x80
`

func TestParseListing(t *testing.T) {
	lm, err := ParseListing(strings.NewReader(sampleListing))
	if err != nil {
		t.Fatal(err)
	}
	check := func(sec string, off uint64, line int, text, macro string) {
		t.Helper()
		loc, ok := lm.Lookup(sec, off)
		if !ok || loc.Line != line || loc.Text != text || loc.Macro != macro || loc.File != "" {
			t.Errorf("%s+%#x: got %+v ok=%v, want line %d %q macro %q", sec, off, loc, ok, line, text, macro)
		}
	}
	check(".text", 0, 11, "mov eax, 4", "")
	check(".text", 0x0a, 13, "mov ecx, msg", "")
	check(".text", 0x0c, 13, "mov ecx, msg", "") // inside an instruction: closest preceding line
	check(".text", 0x1c, 17, "times 2 nop", "")
	check(".text", 0x22, 18, "exit 0", "mov ebx, %1") // macro expansion maps to the invocation line
	check(".data", 0x0c, 3, `msg db "Hello, World! long string here", 10`, "")
	check(".data", 0x1f, 5, "buf: times 4 db 0", "")
	check(".bss", 0, 7, "x resd 10", "")
	if p := lm.Places(12); len(p) != 1 || p[0] != (SecOff{".text", 5}) {
		t.Errorf("places(12) = %v", p)
	}
	if p := lm.Places(18); len(p) != 3 {
		t.Errorf("places(18) = %v", p)
	}
}

func TestParseListingFromNasm(t *testing.T) {
	path, err := Find()
	if err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "util.inc"), []byte("print_nl:\n    mov eax, 4\n    ret\n"), 0o644)
	src := filepath.Join(dir, "p.asm")
	os.WriteFile(src, []byte(`struc point
  .x resd 1
  .y resd 1
endstruc
%macro twice 1
  %1
  %1
%endmacro
section .data
p: istruc point
  at point.x, dd 1
  at point.y, dd 2
iend
section .text
global _start
_start:
  mov eax, [p + point.y]
%rep 2
  inc eax
%endrep
  twice nop
  call print_nl
  mov ebx, 0
%include "util.inc"
`), 0o644)
	res, err := Assemble(path, src, 32, nil, dir, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := os.Open(res.Listing)
	defer f.Close()
	lm, err := ParseListing(f)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		sec  string
		off  uint64
		line int
		file string
	}{
		{".data", 0, 11, ""},  // at point.x (the struc fields must not leak into .data/.text)
		{".data", 4, 12, ""},  // at point.y
		{".text", 0, 17, ""},  // mov eax, [p + point.y]  (5 bytes)
		{".text", 5, 19, ""},  // %rep body
		{".text", 6, 19, ""},  // second repetition
		{".text", 7, 21, ""},  // macro twice -> invocation line
		{".text", 8, 21, ""},  //
		{".text", 9, 22, ""},  // call
		{".text", 14, 23, ""}, // mov ebx, 0
		{".text", 19, 2, "util.inc"},
		{".text", 24, 3, "util.inc"},
	}
	for _, w := range want {
		loc, ok := lm.Lookup(w.sec, w.off)
		if !ok || loc.Line != w.line || loc.File != w.file {
			t.Errorf("%s+%d: got %+v, want line %d file %q", w.sec, w.off, loc, w.line, w.file)
		}
	}
}
