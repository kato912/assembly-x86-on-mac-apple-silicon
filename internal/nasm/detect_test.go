package nasm

import (
	"os"
	"path/filepath"
	"testing"
)

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
		{"extern printf\npush msg\ncall printf\nadd esp, 4", 32}, // cdecl call
		{"push ebp\nmov ebp, esp\nsyscall", 64},
		{"mov rbp, rsp\nmov eax, [ebp]", 64},
	}
	for _, c := range cases {
		if got, why := DetectBits([]byte(c.src)); got != c.bits {
			t.Errorf("%q: got %d (%s), want %d", c.src, got, why, c.bits)
		}
	}
}

func TestDetectBitsFollowsInclude(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "io.inc"), []byte("print:\n  mov eax, 4\n  int 0x80\n  ret\n"), 0o644)
	main := filepath.Join(dir, "main.asm")
	src := []byte("%include \"io.inc\"\n_start: mov eax, 1\n call print\n")
	if bits, why := DetectBitsFile(main, src, nil); bits != 32 {
		t.Errorf("got %d (%s)", bits, why)
	}
	// The main file's own evidence still wins over a weaker hint in the include.
	src = []byte("%include \"io.inc\"\nmov rax, 60\nsyscall\n")
	if bits, _ := DetectBitsFile(main, src, nil); bits != 64 {
		t.Errorf("got %d", bits)
	}
}
