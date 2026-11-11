package debugger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

const factSrc = `
section .data
msg db "done", 10
section .text
global _start
_start:
    mov eax, 4
    call fact          ; eax = 4! = 24
    mov ebx, eax
    mov eax, 1
    int 0x80

fact:
    cmp eax, 1
    jbe .base
    push eax
    dec eax
    call fact
    pop ecx
    mul ecx
    ret
.base:
    mov eax, 1
    ret
`

func session(t *testing.T, commands string) (string, int) {
	p, _, errb := build(t, factSrc, "")
	var out bytes.Buffer
	code := Interactive(NewSession(p), strings.NewReader(commands), &out)
	return out.String() + errb.String(), code
}

func TestStepCommands(t *testing.T) {
	out, code := session(t, "s\n\nb fact\nb 22\nc\nx 4\nr\nm msg 8\nl\nd\nq\n")
	if code != 0 {
		t.Errorf("quit before the end should return 0, got %d", code)
	}
	for _, w := range []string{
		"→ line 7: mov eax, 4",
		"[1]    line 7: mov eax, 4", // Enter/s steps
		"esp=0x",                    // call pushes
		"breakpoint #1 at line 14: cmp eax, 1",
		"breakpoint #2 at line 23: mov eax, 1", // label-only line -> next instruction
		"breakpoint #1\n",                      // hit after c
		"return address → line 9",
		"argc",
		"eax = 0x00000003 (3)", // r: second entry into fact
		"64 6f 6e 65 0a",       // m msg
		"<msg>",
		"→   14      cmp eax, 1", // l marks the current line
		"all breakpoints and watchpoints deleted",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}

func TestNextAndContinue(t *testing.T) {
	out, code := session(t, "s\nn\nr\nc\ns\nq\n")
	if code != 24 {
		t.Errorf("exit code %d\n%s", code, out)
	}
	// "n" ran the whole recursive call: we are back on line 9 with eax = 24.
	if !strings.Contains(out, "→ line 9: mov ebx, eax") || !strings.Contains(out, "eax = 0x00000018 (24)") {
		t.Errorf("next did not step over the call:\n%s", out)
	}
	if !strings.Contains(out, "program exited with code 24") || !strings.Contains(out, "already finished") {
		t.Errorf("missing end-of-program messages:\n%s", out)
	}
}

func TestStepIntoCrash(t *testing.T) {
	p, _, errb := build(t, `
section .text
global _start
_start:
    xor ebx, ebx
    mov eax, [ebx]
`, "")
	var out bytes.Buffer
	code := Interactive(NewSession(p), strings.NewReader("c\nr\nq\n"), &out)
	if code != 139 || !strings.Contains(errb.String(), "cannot read memory at 0x00000000") ||
		!strings.Contains(out.String(), "stopped with an error") {
		t.Errorf("code=%d\n%s%s", code, out.String(), errb.String())
	}
}

func TestEval(t *testing.T) {
	p, _, _ := build(t, factSrc, "")
	d := &stepper{p: p, c: p.CPU}
	msg, _ := p.Image.Lookup("msg")
	cases := map[string]uint64{"msg": msg.Addr, "msg+2": msg.Addr + 2, "0x10": 16, "20": 20,
		"10h": 16, "esp-4": p.CPU.SP() - 4, ".base": mustSym(t, p, "fact.base")}
	for expr, want := range cases {
		got, err := d.eval(expr)
		if err != nil || got != want {
			t.Errorf("eval(%q) = %#x, %v; want %#x", expr, got, err, want)
		}
	}
	if _, err := d.eval("nosuch"); err == nil {
		t.Errorf("expected an error")
	}
}

func mustSym(t *testing.T, p *runner.Program, name string) uint64 {
	t.Helper()
	s, ok := p.Image.Lookup(name)
	if !ok {
		t.Fatalf("no symbol %s", name)
	}
	return s.Addr
}

func TestMultiFileBreakpoints(t *testing.T) {
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.asm")
	ioPath := filepath.Join(dir, "io.asm")
	os.WriteFile(mainPath, []byte("extern twice\nglobal _start\nsection .text\n_start:\n    mov eax, 21\n    call twice\n    mov ebx, eax\n    mov eax, 1\n    int 0x80\n"), 0o644)
	os.WriteFile(ioPath, []byte("global twice\nsection .text\ntwice:\n    add eax, eax\n    ret\n"), 0o644)
	var errb bytes.Buffer
	p, code := runner.Build(runner.Config{Source: mainPath, Extra: []string{ioPath}, MaxSteps: 1000,
		Stdin: strings.NewReader(""), Stdout: io.Discard, Stderr: &errb})
	if p == nil {
		t.Fatalf("build %d: %s", code, errb.String())
	}
	defer p.Close()
	var out bytes.Buffer
	code = Interactive(NewSession(p), strings.NewReader("b io.asm:4\nc\nl\nc\n"), &out)
	for _, w := range []string{
		"breakpoint #1 at io.asm line 4: add eax, eax",
		"→ io.asm line 4: add eax, eax",
		"  io.asm\n", // l shows which file it lists
		"→    4      add eax, eax",
		"program exited with code 42",
	} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("output lacks %q:\n%s", w, out.String())
		}
	}
	if code != 42 {
		t.Errorf("exit code %d", code)
	}
}

func TestExpressions(t *testing.T) {
	p, _, _ := build(t, `
section .data
arr: dd 10, 20, 30, 40
msg: db "Hi!", 0
section .text
global _start
_start:
    mov ecx, 2
    mov eax, 1
    int 0x80
`, "")
	d := &stepper{p: p, c: p.CPU}
	d.s = NewSession(p)
	for i := 0; i < 2; i++ { // run the two movs
		d.s.Step()
	}
	arr := mustSym(t, p, "arr")
	cases := map[string]uint64{
		"-5":                  ^uint64(4),
		"2*3+1":               7,
		"(1+2)*3":             9,
		"'A'":                 65,
		"0b101":               5,
		"arr + 4*2":           arr + 8,
		"[arr]":               10,
		"dword [arr + ecx*4]": 30,
		"byte [msg+1]":        'i',
		"word [msg]":          0x6948,
		"ecx * 10 - 1":        19,
		"[esp]":               1, // argc
	}
	for expr, want := range cases {
		got, err := d.eval(expr)
		if err != nil || got != want {
			t.Errorf("eval(%q) = %#x, %v; want %#x", expr, got, err, want)
		}
	}
	for _, bad := range []string{"", "1 +", "[arr", "nosuch", "2 3", "dword arr"} {
		if _, err := d.eval(bad); err == nil {
			t.Errorf("eval(%q) should fail", bad)
		}
	}
}

func TestPrintCommand(t *testing.T) {
	out, _ := session(t, "s\np eax\np/s msg\np msg\np byte [msg]\np [esp]\np -1\np al\np/s 0\nq\n")
	for _, w := range []string{
		"eax = 0x00000004 = 4",
		`= "done\n" (5 bytes)`,
		"(msg is an address; to see the value stored there use  p [msg]",
		"byte [msg] = 0x64 = 100 'd'",
		"[esp] = 0x00000001 = 1",
		"-1 = 0xffffffff = 4294967295 (signed -1)",
		"al = 0x04 = 4",
		"memory at 0x00000000 is not readable",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}

func TestSetCommand(t *testing.T) {
	// Set eax before "call fact" so the program computes 3! instead of 4!.
	out, code := session(t, "s\nset eax 3\nset al 'A'\nset al 3\nset zf 1\nset dword [msg] 0x454e4f44\np/s msg\nset byte [fact] 0x90\nset eip 99\nc\nq\n")
	for _, w := range []string{
		"eax = 0x00000003 = 3",
		"al = 0x41 = 65 'A'",
		"ZF = 1",
		"dword [0x0804a000] = 0x454e4f44",
		`"DONE\n"`,
		"cannot write memory at", // .text is read-only
		"is not an instruction address",
		"program exited with code 6",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
	if code != 6 {
		t.Errorf("exit code %d, want 6 (3!)", code)
	}
}

func TestBacktraceAndFinish(t *testing.T) {
	// Break at the base case of the recursion: three nested calls to fact.
	out, code := session(t, "b 23\nc\nbt\nfinish\nbt\nfinish\nfinish\nfinish\nfinish\nq\n")
	for _, w := range []string{
		"#0   fact    line 23: mov eax, 1",
		"#1   fact    line 18: call fact",
		"#3   fact    line 18: call fact",
		"#4   _start  line 8: call fact",
		"returned from fact: eax = 0x00000001 (1)",
		"returned from fact: eax = 0x00000002 (2)",
		"returned from fact: eax = 0x00000018 (24)",
		"not inside a function",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
	if code != 0 {
		t.Errorf("quit before the end should return 0, got %d", code)
	}
}

func TestBacktraceAfterCrash(t *testing.T) {
	p, _, _ := build(t, `
section .text
global _start
_start:
    call outer
outer:
    call inner
inner:
    xor ebx, ebx
    mov eax, [ebx]
`, "")
	var out bytes.Buffer
	Interactive(NewSession(p), strings.NewReader("c\nbt\nq\n"), &out)
	for _, w := range []string{"#0   inner   line 10: mov eax, [ebx]", "#1   outer   line 7: call inner", "#2   _start  line 5: call outer"} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("output lacks %q:\n%s", w, out.String())
		}
	}
}

func TestWatchpoints(t *testing.T) {
	p, _, _ := build(t, `
section .data
counter dd 0
section .text
global _start
_start:
    mov ecx, 3
again:
    add dword [counter], 10
    loop again
    mov ebx, [counter]
    mov eax, 1
    int 0x80
`, "")
	var out bytes.Buffer
	code := Interactive(NewSession(p), strings.NewReader("watch counter\nwatch ecx\nwatch\nc\nc\nd 2\nc\nc\nc\nq\n"), &out)
	for _, w := range []string{
		"watchpoint #1: dword [counter] (now 0x00000000 = 0)",
		"watchpoint #2: ecx (now 0x00000000 = 0)",
		"watchpoint #2: ecx changed 0x00000000 (0) → 0x00000003 (3)\n    by line 7: mov ecx, 3",
		"watchpoint #1: dword [counter] changed 0x00000000 (0) → 0x0000000a (10)\n    by line 9: add dword [counter], 10",
		"watchpoint #2 deleted",
		"changed 0x0000000a (10) → 0x00000014 (20)",
		"changed 0x00000014 (20) → 0x0000001e (30)",
		"program exited with code 30",
	} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("output lacks %q:\n%s", w, out.String())
		}
	}
	if code != 30 {
		t.Errorf("exit code %d", code)
	}
}

func TestGdbExpressions(t *testing.T) {
	p, _, _ := build(t, `
section .data
arr: dd 10, -20, 30, 40
msg: db "Hi!", 0
section .text
global _start
_start:
    mov ecx, 2
    mov eax, 1
    int 0x80
`, "")
	d := &stepper{p: p, c: p.CPU}
	d.s = NewSession(p)
	d.s.Step()
	d.s.Step()
	arr := mustSym(t, p, "arr")
	cases := map[string]uint64{
		"$ecx":                       2,
		"$pc":                        p.CPU.RIP,
		"$sp":                        p.CPU.SP(),
		"&arr":                       arr,
		"(int)arr":                   10,
		"*(int*)&arr":                10,
		"*(int*)(&arr + 4)":          0xffffffec, // -20 as a 4-byte value
		"*(short*)&arr":              10,
		"{int}(&arr + 8)":            30,
		"*(char*)&msg":               'H',
		"(char)0x141":                0x41,
		"(unsigned char)$ecx":        2,
		"*(int*)($sp)":               1, // argc
		"$ecx * 4 + 1":               9,
		"*(unsigned long long*)&arr": 0xffffffec0000000a,
	}
	for expr, want := range cases {
		got, err := d.eval(expr)
		if err != nil || got != want {
			t.Errorf("eval(%q) = %#x, %v; want %#x", expr, got, err, want)
		}
	}
	if _, err := d.eval("$nosuch"); err == nil {
		t.Errorf("$nosuch should fail")
	}
}

func TestGdbInfoAndPrint(t *testing.T) {
	out, _ := session(t, "s\ninfo registers\ni r eax eflags al\np/x $eax\np/t $eax\np/d -1\np/c 65\np $eflags\np/x (int)msg\nbreak *0x08049005\nb fact\ninfo breakpoints\nq\n")
	for _, w := range []string{
		"eax            0x4                4\n",
		"esp            0xfffd",
		"eip            0x8049005          0x8049005 <_start+5>",
		"eflags         0x202              [ IF ]",
		"cs             0x23               35",
		"al             0x4                4",
		"$1 = 0x4",
		"$2 = 100",
		"$3 = -1",
		"$4 = 65 'A'",
		"$5 = [ IF ]",
		"$6 = 0x656e6f64", // "done" read as an int
		"breakpoint #1 at line 8: call fact",
		"Num     Type           Disp Enb Address    What",
		"1       breakpoint     keep y   0x08049005 line 8: call fact",
		"2       breakpoint     keep y",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}

func TestGdbExamineAndDisassemble(t *testing.T) {
	out, _ := session(t, "s\ns\nx/4xw $esp\nx/s &msg\nx/5xb msg\nx\nx/2i $pc\nx/dw $esp\nx 2\ndisassemble\ndisas fact\nx/xw 0\nq\n")
	for _, w := range []string{
		"\t0x0804900a\t0x00000001", // return address of "call fact", then argc
		"0x804a000 <msg>:\t\"done\\n\"",
		"0x804a000 <msg>:\t0x64\t0x6f\t0x6e\t0x65\t0x0a",
		"=> 0x8049013 <fact>:\tcmp    eax, 0x1",
		"Dump of assembler code for function fact:",
		"=> 0x08049013 <+0>:\tcmp    eax, 0x1",
		"; line 14",
		"End of assembler dump.",
		"Cannot access memory at address 0x0",
		"← esp", // "x 2" is still the stack view
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}
