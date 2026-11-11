package nasm

import "testing"

func TestDetectBits(t *testing.T) {
	cases := []struct {
		src  string
		bits int
	}{
		{"bits 32\nmov rax, 1\nsyscall", 32},
		{"[BITS 64]\nmov eax, 1\nint 0x80", 64},
		{"mov eax, 4\nmov ebx, 1\nint 0x80", 32},
		{"mov eax, 4\nint 80h", 32},
		{"mov eax, 60\nxor edi, edi\nsyscall", 64},
		{"mov rax, 1\nint 0x80", 64},
		{"mov r8d, 1\nint 0x80", 64},
		{"; mov rax, 60 in a comment\nmov eax, 1\nint 0x80", 32},
		{"msg db 'rax syscall', 0\nmov eax, 1\nint 0x80", 32},
		{"default rel\nmov eax, 1", 64},
		{"mov eax, 1", 64},
		{"use32\nnop", 32},
	}
	for _, c := range cases {
		if got, why := DetectBits([]byte(c.src)); got != c.bits {
			t.Errorf("%q: got %d (%s), want %d", c.src, got, why, c.bits)
		}
	}
}
