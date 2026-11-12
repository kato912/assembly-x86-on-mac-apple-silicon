package main

import (
	"reflect"
	"testing"
)

func TestAsmArgs(t *testing.T) {
	cases := []struct {
		asm       string
		args, out []string
		src       string
		hasFormat bool
	}{
		// Each assembler's own course command lines pass through.
		{"yasm", []string{"-f", "elf64", "-g", "dwarf2", "-o", "hello.o", "hello.s"},
			[]string{"-f", "elf64", "-g", "dwarf2", "-o", "hello.o", "hello.s"}, "hello.s", true},
		{"yasm", []string{"-f", "elf64", "-o", "hello.o", "-l", "hello.lst", "hello.s"},
			[]string{"-f", "elf64", "-o", "hello.o", "-l", "hello.lst", "hello.s"}, "hello.s", true},
		{"nasm", []string{"-f", "elf64", "-g", "-F", "dwarf", "-o", "hello.o", "hello.asm"},
			[]string{"-f", "elf64", "-g", "-F", "dwarf", "-o", "hello.o", "hello.asm"}, "hello.asm", true},
		{"nasm", []string{"-felf32", "-l", "x.lst", "x.asm"}, []string{"-felf32", "-l", "x.lst", "x.asm"}, "x.asm", true},
		// The other assembler's debug options are translated.
		{"nasm", []string{"-f", "elf64", "-g", "dwarf2", "-o", "hello.o", "hello.s"},
			[]string{"-f", "elf64", "-g", "-F", "dwarf", "-o", "hello.o", "hello.s"}, "hello.s", true},
		{"nasm", []string{"-gdwarf2", "x.asm"}, []string{"-g", "-F", "dwarf", "x.asm"}, "x.asm", false},
		{"yasm", []string{"-f", "elf64", "-g", "-F", "dwarf", "x.asm"}, []string{"-f", "elf64", "-g", "dwarf2", "x.asm"}, "x.asm", true},
		{"yasm", []string{"-g", "x.s"}, []string{"-g", "dwarf2", "x.s"}, "x.s", false},
		// Values are not taken for the source file.
		{"nasm", []string{"-I", "inc/", "-D", "N=3", "-o", "out.o", "main.asm"},
			[]string{"-I", "inc/", "-D", "N=3", "-o", "out.o", "main.asm"}, "main.asm", false},
		{"yasm", []string{"--version"}, []string{"--version"}, "", false},
	}
	for _, c := range cases {
		out, src, hasFormat := asmArgs(c.asm, c.args)
		if !reflect.DeepEqual(out, c.out) || src != c.src || hasFormat != c.hasFormat {
			t.Errorf("%s-x86 %v:\n got %v %q %v\nwant %v %q %v", c.asm, c.args, out, src, hasFormat, c.out, c.src, c.hasFormat)
		}
	}
}
