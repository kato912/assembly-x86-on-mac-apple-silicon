package cc

import (
	"debug/elf"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/libc"
)

func TestCompile(t *testing.T) {
	if _, err := Find(); err != nil {
		t.Skip("clang not installed")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "main.c")
	os.WriteFile(src, []byte(`#include <stdio.h>
#include <stdlib.h>
#include <string.h>
int main(void) { char *s = malloc(8); strcpy(s, "hi"); printf("%s %d\n", s, (int)strlen(s)); return 0; }
`), 0o644)
	for bits, machine := range map[int]elf.Machine{64: elf.EM_X86_64, 32: elf.EM_386} {
		obj, err := Compile(src, bits, []string{"-Wall", "-O1", "-no-pie"}, true, t.TempDir(), io.Discard)
		if err != nil {
			t.Fatalf("%d-bit: %v", bits, err)
		}
		f, err := elf.Open(obj)
		if err != nil {
			t.Fatal(err)
		}
		if f.Machine != machine || f.Type != elf.ET_REL || f.Section(".debug_line") == nil {
			t.Errorf("%d-bit: machine %v type %v", bits, f.Machine, f.Type)
		}
		f.Close()
	}
	bad := filepath.Join(dir, "bad.c")
	os.WriteFile(bad, []byte("int main(void) { return x; }\n"), 0o644)
	if _, err := Compile(bad, 64, nil, false, t.TempDir(), io.Discard); err == nil {
		t.Error("expected a compile error")
	}
}

// The headers declare exactly the functions of the built-in C library.
func TestHeadersMatchLibrary(t *testing.T) {
	declared := map[string]bool{}
	decl := regexp.MustCompile(`(?m)^[A-Za-z_ ]+[ *]+([a-z_][a-z_0-9]*)\(`)
	entries, _ := headers.ReadDir("include")
	for _, e := range entries {
		data, _ := headers.ReadFile("include/" + e.Name())
		for _, m := range decl.FindAllStringSubmatch(string(data), -1) {
			declared[m[1]] = true
		}
	}
	lib := map[string]bool{}
	for _, n := range libc.Names() {
		lib[n] = true
	}
	for n := range lib {
		if !declared[n] && n != "stdin" && n != "stdout" && n != "stderr" {
			t.Errorf("%s is in the C library but in no header", n)
		}
	}
	for n := range declared {
		if !lib[n] {
			t.Errorf("%s is declared but not in the C library", n)
		}
	}
}
