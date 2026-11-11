package debugger

import (
	"bytes"
	"strings"
	"testing"

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
		"breakpoint #1\n",                       // hit after c
		"return address → line 9",
		"argc",
		"eax = 0x00000003 (3)", // r: second entry into fact
		"64 6f 6e 65 0a",       // m msg
		"<msg>",
		"→   14      cmp eax, 1", // l marks the current line
		"all breakpoints deleted",
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
