package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		args     []string
		file     string
		bits     int
		nasm     []string
		progArgs []string
	}{
		{[]string{"hello.asm"}, "hello.asm", 0, nil, nil},
		{[]string{"-f", "elf32", "-o", "hello.o", "hello.asm"}, "hello.asm", 32, nil, nil},
		{[]string{"-felf64", "-g", "-F", "dwarf", "x.asm"}, "x.asm", 64, nil, nil},
		{[]string{"-32", "-DDEBUG", "-I", "inc/", "x.asm", "a1", "a2"}, "x.asm", 32, []string{"-DDEBUG", "-I", "inc/"}, []string{"a1", "a2"}},
		{[]string{"x.asm", "--max-steps", "10", "arg"}, "x.asm", 0, nil, []string{"arg"}},
		{[]string{"x.asm", "--", "--max-steps"}, "x.asm", 0, nil, []string{"--max-steps"}},
		{[]string{"--", "-weird.asm", "p"}, "-weird.asm", 0, nil, []string{"p"}},
		{[]string{"-64", "-w+all", "x.asm"}, "x.asm", 64, []string{"-w+all"}, nil},
	}
	for _, c := range cases {
		o, err := parseArgs(c.args)
		if err != nil {
			t.Errorf("%v: %v", c.args, err)
			continue
		}
		if o.cfg.Source != c.file || o.cfg.Bits != c.bits || !reflect.DeepEqual(o.cfg.NasmArgs, c.nasm) || !reflect.DeepEqual(o.cfg.Args, c.progArgs) {
			t.Errorf("%v: got file=%q bits=%d nasm=%q args=%q", c.args, o.cfg.Source, o.cfg.Bits, o.cfg.NasmArgs, o.cfg.Args)
		}
	}
	if o, _ := parseArgs([]string{"--max-steps=5", "x.asm"}); o.cfg.MaxSteps != 5 {
		t.Errorf("max steps %d", o.cfg.MaxSteps)
	}
	for _, bad := range [][]string{{}, {"-f", "win64", "x.asm"}, {"--max-steps", "lots", "x.asm"}, {"-f"}} {
		if _, err := parseArgs(bad); err == nil {
			t.Errorf("%v: expected an error", bad)
		}
	}
}
