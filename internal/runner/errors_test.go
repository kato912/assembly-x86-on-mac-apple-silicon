package runner

import (
	"strings"
	"testing"
)

func TestFriendlyErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code int
		want []string
	}{
		{"null read", `
section .text
global _start
_start:
    xor ebx, ebx
    mov eax, [ebx]
`, 139, []string{"line 6: mov eax, [ebx]  →  cannot read memory at 0x00000000 (ebx = 0x00000000)", "hint:"}},
		{"write to text", `
section .text
global _start
_start:
    mov dword [counter], 1
    int 0x80
counter: dd 0
`, 139, []string{"line 5:", "cannot write memory", ".text is read-only"}},
		{"div overflow", `
section .text
global _start
_start:
    mov eax, 10
    mov edx, 5
    mov ebx, 2
    div ebx
`, 136, []string{"line 8: div ebx", "edx:eax = 0x00000005:0x0000000a divided by 2 does not fit in eax", "xor edx, edx"}},
		{"div zero", `
section .text
global _start
_start:
    mov eax, 1
    xor ecx, ecx
    idiv ecx
    int 0x80
`, 136, []string{"line 7: idiv ecx", "division by zero (the divisor ecx is 0)"}},
		{"stack overflow", `
section .text
global _start
_start:
    call _start
`, 139, []string{"line 5: call _start", "stack overflow", "recursion"}},
		{"ret from start", `
section .text
global _start
_start:
    mov eax, 1
    ret
`, 139, []string{"line 6: ret", "ret in _start", "exit system call"}},
		{"bad ret", `
section .text
global _start
_start:
    call f
    mov eax, 1
    int 0x80
f:  push 5
    ret
`, 139, []string{"line 9: ret", "ret jumped to 0x00000005", "matching pop"}},
		{"run off end", `
section .text
global _start
_start:
    mov eax, 1
`, 139, []string{"line 5: mov eax, 1", "ran past the end", "exit system call"}},
		{"run off end 32", `
section .text
global _start
_start:
    mov eax, 4
    int 0x80
`, 139, []string{"line 6: int 0x80", "ran past the end", "mov eax, 1 / mov ebx, 0 / int 0x80"}},
		{"dos interrupt", `
section .text
global _start
_start:
    mov ah, 9
    int 21h
`, 139, []string{"line 6: int 21h", "int 0x21", "MS-DOS"}},
		{"syscall in 32-bit", `
bits 32
section .text
global _start
_start:
    mov eax, 1
    syscall
`, 132, []string{"line 7: syscall", "int 0x80"}},
		{"fpu", `
section .text
global _start
_start:
    fld1
    int 0x80
`, 132, []string{"line 5: fld1", "unsupported instruction fld1", "floating-point"}},
		{"jump into data", `
section .data
x dd 0
section .text
global _start
_start:
    jmp x
`, 139, []string{"line 7: jmp x", "jumped to", "not part of the code"}},
		{"macro line", `
%macro load 1
    mov eax, [%1]
%endmacro
section .text
global _start
_start:
    xor esi, esi
    load esi
`, 139, []string{"line 9: load esi (in macro: mov eax, [%1])", "esi = 0x00000000"}},
	}
	for _, c := range cases {
		_, errs, code := runSource(t, c.src, "")
		if code != c.code {
			t.Errorf("%s: exit code %d, want %d\n%s", c.name, code, c.code, errs)
		}
		for _, w := range c.want {
			if !strings.Contains(errs, w) {
				t.Errorf("%s: stderr lacks %q:\n%s", c.name, w, errs)
			}
		}
	}
}

func TestMaxStepsMessage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	_, errs, code := runSourceCfg(t, `
section .text
global _start
_start:
    jmp _start
`, func(c *Config) { c.MaxSteps = 1000 })
	if code != ExitMaxSteps || !strings.Contains(errs, "line 5: jmp _start") || !strings.Contains(errs, "infinite loop") {
		t.Errorf("code=%d\n%s", code, errs)
	}
}
