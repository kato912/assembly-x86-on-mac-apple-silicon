package emu

import (
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

func TestSubRegisters64(t *testing.T) {
	c := run(t, 64, `
	mov rax, 0x1122334455667788
	mov al, 0xff        ; only the low byte changes
	mov rbx, rax
	mov ah, 0x01
	mov rcx, rax
	mov ax, 0xabcd
	mov rdx, rax
	mov eax, 1          ; 32-bit write zero-extends
	mov r8, -1
	mov r8b, 0
	mov r9, -1
	mov r9w, 0
	mov r10, -1
	mov r10d, 2
	mov rsi, -1
	mov sil, 5
`)
	expect(t, c, RBX, 0x11223344556677ff)
	expect(t, c, RCX, 0x11223344556601ff)
	expect(t, c, RDX, 0x112233445566abcd)
	expect(t, c, RAX, 1)
	expect(t, c, R8, 0xffffffffffffff00)
	expect(t, c, R9, 0xffffffffffff0000)
	expect(t, c, R10, 2)
	expect(t, c, RSI, 0xffffffffffffff05)
}

func TestAddSubFlags(t *testing.T) {
	cases := []struct {
		name      string
		a, b      uint64
		size      int
		sub       bool
		res, flag uint64
	}{
		{"add 8 carry+zero", 0xff, 0x01, 1, false, 0x00, CF | ZF | PF | AF},
		{"add 8 overflow", 0x7f, 0x01, 1, false, 0x80, OF | SF | AF},
		{"add 32 plain", 2, 3, 4, false, 5, PF},
		{"add 32 carry", 0xffffffff, 2, 4, false, 1, CF | AF},
		{"add 64 carry", ^uint64(0), 1, 8, false, 0, CF | ZF | PF | AF},
		{"add 64 overflow", 0x7fffffffffffffff, 1, 8, false, 0x8000000000000000, OF | SF | PF | AF},
		{"sub 8 borrow", 0, 1, 1, true, 0xff, CF | SF | PF | AF},
		{"sub 32 equal", 5, 5, 4, true, 0, ZF | PF},
		{"sub 32 borrow", 1, 2, 4, true, 0xffffffff, CF | SF | PF | AF},
		{"sub 8 overflow", 0x80, 1, 1, true, 0x7f, OF | AF},
		{"sub 16", 0x1000, 0x0001, 2, true, 0x0fff, PF | AF},
		{"sub 64 borrow", 0, 1, 8, true, ^uint64(0), CF | SF | PF | AF},
	}
	for _, tc := range cases {
		var r, f uint64
		if tc.sub {
			r, f = Sub(tc.a, tc.b, 0, tc.size)
		} else {
			r, f = Add(tc.a, tc.b, 0, tc.size)
		}
		if r != tc.res || f != tc.flag {
			t.Errorf("%s: got %#x [%s], want %#x [%s]", tc.name, r, FormatFlags(f), tc.res, FormatFlags(tc.flag))
		}
	}
}

func TestALUAndJumps32(t *testing.T) {
	c := run(t, 32, `
	xor eax, eax
	mov ecx, 10
top:
	add eax, ecx
	dec ecx
	jnz top          ; eax = 55
	mov ebx, eax
	sub ebx, 60      ; negative
	js neg
	mov edx, 1
	jmp done
neg:
	mov edx, 2
done:
	cmp eax, 55
	je ok
	mov esi, 0xbad
ok:
	and eax, 0x0f    ; 55 & 15 = 7
	or eax, 0x100
	test eax, 0x100
`)
	expect(t, c, RAX, 0x107)
	expect(t, c, RBX, 0xfffffffb)
	expect(t, c, RDX, 2)
	expect(t, c, RSI, 0)
	expectFlags(t, c, ZF|CF|OF, 0)
}

func TestIncDecKeepCarry(t *testing.T) {
	c := run(t, 32, `
	mov eax, 0xffffffff
	add eax, 1       ; CF=1
	inc eax          ; CF unchanged
`)
	expect(t, c, RAX, 1)
	expectFlags(t, c, CF|ZF, CF)
}

func TestStackCallRet(t *testing.T) {
	c := run(t, 64, `
	mov rax, 3
	push rax
	push -2          ; sign-extended to 64 bits
	call f
	pop rbx
	pop rcx
	jmp out
f:
	mov rdx, [rsp + 8]
	ret
out:
`)
	expect(t, c, RBX, ^uint64(1))
	expect(t, c, RCX, 3)
	expect(t, c, RDX, ^uint64(1))
	expect(t, c, RSP, testStack)

	c = run(t, 32, `
	mov eax, 3
	push eax
	push 7
	call f
	pop ebx
	pop ecx
	jmp out
f:
	mov edx, [esp + 4]
	ret 0
out:
`)
	expect(t, c, RBX, 7)
	expect(t, c, RCX, 3)
	expect(t, c, RDX, 7)
	expect(t, c, RSP, testStack)
}

func TestAddressingModes(t *testing.T) {
	c := run(t, 32, `
	mov ebx, arr
	mov esi, 2
	mov eax, [ebx + esi*4]        ; arr[2] = 30
	mov ecx, [arr + 4]            ; disp only = 20
	mov edx, [ebx + esi*4 - 4]    ; 20
	lea edi, [ebx + esi*4 + 8]    ; address of arr[4]
	sub edi, ebx                  ; 16
	mov dword [ebx + 12], 99
	mov ebp, [arr + 12]
	jmp over
arr: dd 10, 20, 30, 40, 50
over:
`)
	expect(t, c, RAX, 30)
	expect(t, c, RCX, 20)
	expect(t, c, RDX, 20)
	expect(t, c, RDI, 16)
	expect(t, c, RBP, 99)

	c = run(t, 64, `
	default rel
	mov rax, [val]               ; RIP-relative
	lea rbx, [val]
	mov rcx, [rbx]
	mov r8, 1
	mov r9, [rbx + r8*8]
	jmp over
val: dq 0x1234, 0x5678
over:
`)
	expect(t, c, RAX, 0x1234)
	expect(t, c, RCX, 0x1234)
	expect(t, c, R9, 0x5678)
}

func TestMemoryFaultInfo(t *testing.T) {
	f := runFault(t, 32, `
	xor ebx, ebx
	mov eax, [ebx]
`)
	if f.Kind != FaultMemory || f.Mem.Addr != 0 || len(f.Regs) != 1 || f.Regs[0].Name != "ebx" {
		t.Fatalf("fault %+v regs %+v", f, f.Regs)
	}
	if f.PC != testCode+2 {
		t.Errorf("fault pc %#x", f.PC)
	}
}

func TestRunOffEnd(t *testing.T) {
	c := machine(t, 32, "nop")
	c.Code[0].End-- // pretend hlt is not part of the code
	err := c.Run()
	f, ok := err.(*Fault)
	if !ok || f.Kind != FaultRunOff {
		t.Fatalf("got %v", err)
	}
}

func TestMaxSteps(t *testing.T) {
	c := machine(t, 64, "l: jmp l")
	c.MaxSteps = 1000
	err := c.Run()
	if f, ok := err.(*Fault); !ok || f.Kind != FaultMaxSteps {
		t.Fatalf("got %v", err)
	}
}

func TestNegativeDisp32(t *testing.T) {
	// Backward RIP-relative references and large negative displacements
	// are encoded as disp32 and must be sign-extended.
	c := run(t, 64, `
	jmp start
before: dq 0x1111
start:
	lea rax, [rel before]
	mov rbx, [rel before]
	mov rcx, rsp
	mov qword [rcx - 0x200], 0x2222
	mov rdx, [rcx - 0x200]
	lea rsi, [rcx - 0x12345]
	sub rcx, rsi
`)
	expect(t, c, RAX, testCode+2)
	expect(t, c, RBX, 0x1111)
	expect(t, c, RDX, 0x2222)
	expect(t, c, RCX, 0x12345)
}

func TestDisasmLabelImmediate(t *testing.T) {
	in, err := x86asm.Decode([]byte{0xb9, 0x00, 0xa0, 0x04, 0x08}, 32) // mov ecx, 0x804a000
	if err != nil {
		t.Fatal(err)
	}
	sym := func(a uint64) (string, uint64) {
		if a == 0x804a000 {
			return "msg", a
		}
		return "", 0
	}
	if got := Disasm(&in, 0x8049000, sym); got != "mov ecx, msg" {
		t.Errorf("got %q", got)
	}
}
