package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/progfile"
)

func TestParseLD(t *testing.T) {
	cases := []struct {
		args   []string
		out    string
		bits   int
		entry  string
		clib   bool
		inputs []string
	}{
		{[]string{"-o", "hello", "hello.o"}, "hello", 0, "", false, []string{"hello.o"}},
		{[]string{"-ohello", "a.o", "b.o"}, "hello", 0, "", false, []string{"a.o", "b.o"}},
		{[]string{"-m", "elf_i386", "-o", "h", "h.o"}, "h", 32, "", false, []string{"h.o"}},
		{[]string{"-melf_x86_64", "--output=h", "h.o"}, "h", 64, "", false, []string{"h.o"}},
		{[]string{"-g", "-e", "main", "-o", "x", "x.o"}, "x", 0, "main", false, []string{"x.o"}},
		{[]string{"--entry=go", "x.o"}, "", 0, "go", false, []string{"x.o"}},
		{[]string{"-dynamic-linker", "/lib64/ld-linux-x86-64.so.2", "-lc", "-o", "p", "p.o"}, "p", 0, "", true, []string{"p.o"}},
		{[]string{"-static", "-nostdlib", "-z", "noexecstack", "-L/usr/lib", "-s", "x.o"}, "", 0, "", false, []string{"x.o"}},
		{[]string{"-export-dynamic", "x.o"}, "", 0, "", false, []string{"x.o"}},
	}
	for _, c := range cases {
		l := parseLD(c.args)
		if len(l.errs) > 0 || l.out != c.out || l.bits != c.bits || l.entry != c.entry || l.clib != c.clib || !reflect.DeepEqual(l.inputs, c.inputs) {
			t.Errorf("%v: %+v", c.args, l)
		}
	}
	for _, bad := range [][]string{{"-m", "i386pe", "x.o"}, {"-T", "link.ld", "x.o"}, {"--frobnicate", "x.o"}, {"-lm", "x.o"}, {"-o"}} {
		if l := parseLD(bad); len(l.errs) == 0 {
			t.Errorf("%v: expected an error", bad)
		}
	}
}

func TestParseGCC(t *testing.T) {
	g := parseGCC([]string{"-g", "-Wall", "-O0", "-no-pie", "main.c", "stats.o", "-o", "cstats", "-DN=3", "-I", "inc", "-lm"})
	if len(g.errs) > 0 || g.out != "cstats" || !g.debug || !reflect.DeepEqual(g.inputs, []string{"main.c", "stats.o"}) ||
		!reflect.DeepEqual(g.csrc, []string{"main.c"}) || !reflect.DeepEqual(g.cflags, []string{"-Wall", "-O0", "-no-pie", "-DN=3", "-I", "inc"}) {
		t.Errorf("%+v", g)
	}
	g = parseGCC([]string{"-m32", "-nostartfiles", "-o", "p", "p.o"})
	if g.bits != 32 || !g.noCRT || g.noLibc || g.out != "p" {
		t.Errorf("%+v", g)
	}
	if g := parseGCC([]string{"-nostdlib", "p.o"}); !g.noCRT || !g.noLibc {
		t.Errorf("-nostdlib: %+v", g)
	}
	for _, bad := range [][]string{{"-c", "x.o"}, {"-lpthread", "x.o"}} {
		if g := parseGCC(bad); len(g.errs) == 0 {
			t.Errorf("%v: expected an error", bad)
		}
	}
}

func TestForMac(t *testing.T) {
	dir := t.TempDir()
	write := func(name, data string) string {
		p := filepath.Join(dir, name)
		os.WriteFile(p, []byte(data), 0o644)
		return p
	}
	elf := write("x86.o", "\x7fELF\x02\x01\x01\x00")
	mac := write("mac.o", "\xcf\xfa\xed\xfe\x0c\x00\x00\x01")
	arch := write("libfoo.a", "!<arch>\n")
	src := write("hello.s", "section .text\n")
	cases := []struct {
		tool string
		args []string
		mac  bool
	}{
		{"ld", []string{"-o", "h", elf}, false},
		{"ld", []string{"-o", "h", mac}, true},
		{"ld", []string{"-arch", "arm64", "-o", "h", mac, arch}, true},
		{"ld", []string{"-v"}, true},
		{"ld", []string{"-o", "h", src}, false},                          // nasmx86 explains: assemble it first
		{"ld", []string{"-o", "h", filepath.Join(dir, "nope.o")}, false}, // and: cannot find nope.o
		{"gcc", []string{"-o", "m", "m.c"}, true},
		{"gcc", []string{"main.c", elf, "-o", "c"}, false},
		{"gcc", []string{"-m32", "-o", "m", "m.c"}, false},
		{"gcc", []string{"-o", "p", elf}, false},
		{"gcc", []string{"--version"}, true},
		{"gcc", []string{"-E", "-dM", "-"}, true},
		{"gcc", []string{"-o", "p", filepath.Join(dir, "nope.o")}, true}, // the Mac's gcc reports it
		{"gcc", []string{"-m32"}, true},
		{"ld", []string{"-"}, true},
	}
	for _, c := range cases {
		l := parseLD(c.args)
		if c.tool == "gcc" {
			l = parseGCC(c.args)
		}
		if l.forMac() != c.mac {
			t.Errorf("%s %v: forMac = %v", c.tool, c.args, !c.mac)
		}
	}
}

func TestProgramFileArgs(t *testing.T) {
	prog := filepath.Join(t.TempDir(), "hello")
	if err := progfile.Write(prog, progfile.Manifest{Linker: progfile.LD, Bits: 64, Objects: []string{"h.o"}}, [][]byte{nil}); err != nil {
		t.Fatal(err)
	}
	// ./hello --trace x runs nasmx86 ./hello --trace x: the options are the program's.
	o, err := parseArgs([]string{prog, "--trace", "x.asm", "--", "-v"})
	if err != nil || o.trace || o.cfg.Source != prog || !reflect.DeepEqual(o.cfg.Args, []string{"--trace", "x.asm", "--", "-v"}) {
		t.Errorf("%+v %v", o, err)
	}
	if o, _ := parseArgs([]string{"--trace", prog, "a"}); !o.trace || !reflect.DeepEqual(o.cfg.Args, []string{"a"}) {
		t.Errorf("nasmx86 --trace ./hello a: %+v", o)
	}
}
