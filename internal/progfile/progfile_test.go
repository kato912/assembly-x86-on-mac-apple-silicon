package progfile

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello")
	m := Manifest{Linker: GCC, Bits: 32, CLib: true, Objects: []string{"main.o", "lib/io.o"},
		Command: []string{"gcc", "-m32", "-o", "hello", "main.o", "lib/io.o"}}
	data := [][]byte{[]byte("\x7fELF main"), []byte("\x7fELF io")}
	if err := Write(path, m, data); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil || (runtime.GOOS != "windows" && st.Mode().Perm() != 0o755) {
		t.Fatalf("mode %v %v", st.Mode(), err)
	}
	if !Is(path) {
		t.Fatal("Is = false")
	}
	p, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	m.Format = format
	if !reflect.DeepEqual(p.Manifest, m) {
		t.Errorf("manifest %+v, want %+v", p.Manifest, m)
	}
	paths, err := p.Extract(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for i, path := range paths {
		got, _ := os.ReadFile(path)
		if !bytes.Equal(got, data[i]) || filepath.Base(path) != filepath.Base(m.Objects[i]) {
			t.Errorf("object %d: %s %q", i, path, got)
		}
	}
	head, _ := os.ReadFile(path)
	if !bytes.HasPrefix(head, []byte("#!/usr/bin/env nasmx86\n# nasmx86 program: a Linux i386 program made by: gcc -m32 -o hello main.o lib/io.o\n")) {
		t.Errorf("header:\n%s", head[:120])
	}
}

func TestNotAProgram(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"script.asm": "#!/usr/bin/env nasmx86\nsection .text\n", // a source file run as a script
		"empty":      "",
		"elf":        "\x7fELF\x02\x01\x01",
	} {
		path := filepath.Join(dir, name)
		os.WriteFile(path, []byte(text), 0o644)
		if Is(path) {
			t.Errorf("%s: Is = true", name)
		}
	}
	path := filepath.Join(dir, "broken")
	os.WriteFile(path, []byte(shebang+magic+"x\nnot a zip"), 0o755)
	if !Is(path) {
		t.Error("broken: Is = false")
	}
	if _, err := Open(path); err == nil {
		t.Error("broken: Open succeeded")
	}
	if Is(filepath.Join(dir, "missing")) {
		t.Error("missing: Is = true")
	}
}
