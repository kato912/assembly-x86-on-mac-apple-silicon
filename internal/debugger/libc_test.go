package debugger

import (
	"bytes"
	"strings"
	"testing"
)

const libcSrc = `
default rel
extern printf, stdout
section .data
fmt db "n=%d", 10, 0
section .text
global main
main:
    push rbp
    mov edi, 5
    call show
    xor eax, eax
    pop rbp
    ret
show:
    push rbx
    mov esi, edi
    lea rdi, [fmt]
    xor eax, eax
    call printf
    pop rbx
    ret
`

// Calls into the C library run as one step, like gdb's next, and leave
// no frame behind in bt.
func TestDebuggerLibcCalls(t *testing.T) {
	p, out, _ := build(t, libcSrc, "")
	var dbg bytes.Buffer
	Interactive(NewSession(p), strings.NewReader("b show\nc\ns\ns\ns\ns\nx/i $pc\ns\nbt\np/x stdout\nc\n"), &dbg)
	got := dbg.String()
	for _, w := range []string{
		"call printf",
		`→ printf("n=%d\n", 5) = 4`,
		"#0  ",
		"show",
		"main",
		"= 0x7f010008", // p/x stdout: the variable's address, like any label
	} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in:\n%s", w, got)
		}
	}
	if strings.Contains(got, "#2") {
		t.Errorf("bt shows a frame for printf:\n%s", got)
	}
	if out.String() != "n=5\n" {
		t.Errorf("program output %q", out.String())
	}
}
