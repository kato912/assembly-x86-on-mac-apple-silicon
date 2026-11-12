// Package progfile reads and writes nasmx86 program files: what nasmx86's
// ld and gcc write where Linux's would write an ELF executable. macOS
// cannot run a Linux program, so the file is a script for nasmx86
// (#!/usr/bin/env nasmx86) and ./hello runs it in nasmx86's emulator, as a
// student would on Linux. The script carries the object files and the link
// options in a zip archive; nasmx86 links them again each time it runs or
// debugs the program.
package progfile

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	shebang  = "#!/usr/bin/env nasmx86\n"
	magic    = "# nasmx86 program: "
	manifest = "nasmx86.json"
	format   = 1
)

// Linker names in Manifest.Linker.
const (
	LD  = "ld"
	GCC = "gcc"
)

// Manifest describes how the program is linked.
type Manifest struct {
	Format  int      `json:"format"`
	Linker  string   `json:"linker"`          // LD or GCC
	Bits    int      `json:"bits"`            // 32 or 64
	Entry   string   `json:"entry,omitempty"` // ld -e SYMBOL
	CLib    bool     `json:"clib,omitempty"`  // ld -lc: link with the C library
	Objects []string `json:"objects"`         // the object files, as named on the command line
	Command []string `json:"command,omitempty"`
}

// Is reports whether path is a nasmx86 program file.
func Is(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, len(shebang)+len(magic))
	if _, err := io.ReadFull(f, head); err != nil {
		return false
	}
	return string(head) == shebang+magic
}

// Write writes a program file to path (mode 0755, like ld's output).
// data holds the contents of each object file named in m.Objects.
func Write(path string, m Manifest, data [][]byte) error {
	if len(data) != len(m.Objects) {
		return errors.New("progfile: objects and data differ in length")
	}
	m.Format = format
	var buf bytes.Buffer
	buf.WriteString(shebang)
	machine := "x86-64"
	if m.Bits == 32 {
		machine = "i386"
	}
	fmt.Fprintf(&buf, "%sa Linux %s program made by: %s\n", magic, machine, strings.Join(m.Command, " "))
	fmt.Fprintf(&buf, "# macOS cannot run Linux programs itself, so nasmx86 runs it: ./%s\n", filepath.Base(path))
	zw := zip.NewWriter(&buf)
	zw.SetOffset(int64(buf.Len()))
	js, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	w, err := zw.Create(manifest)
	if err != nil {
		return err
	}
	w.Write(append(js, '\n'))
	for i, name := range m.Objects {
		w, err := zw.Create(objectEntry(i, name))
		if err != nil {
			return err
		}
		w.Write(data[i])
	}
	if err := zw.Close(); err != nil {
		return err
	}
	// Replace the file rather than writing into it: a program that is
	// running (or open in a debugger) keeps its old contents.
	os.Remove(path)
	return os.WriteFile(path, buf.Bytes(), 0o755)
}

func objectEntry(i int, name string) string {
	return fmt.Sprintf("objects/%d/%s", i, filepath.Base(name))
}

// Program is an opened program file.
type Program struct {
	Manifest
	zr *zip.Reader
}

// Open reads a program file.
func Open(path string) (*Program, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(raw, []byte(shebang+magic)) {
		return nil, fmt.Errorf("%s is not a nasmx86 program", path)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("%s is damaged: %v", path, err)
	}
	p := &Program{zr: zr}
	js, err := p.read(manifest)
	if err != nil {
		return nil, fmt.Errorf("%s is damaged: %v", path, err)
	}
	if err := json.Unmarshal(js, &p.Manifest); err != nil {
		return nil, fmt.Errorf("%s is damaged: %v", path, err)
	}
	if p.Format > format {
		return nil, fmt.Errorf("%s was made by a newer nasmx86: upgrade nasmx86 or link it again", path)
	}
	if len(p.Objects) == 0 {
		return nil, fmt.Errorf("%s is damaged: no object files", path)
	}
	return p, nil
}

func (p *Program) read(name string) ([]byte, error) {
	f, err := p.zr.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(bufio.NewReader(f))
}

// Extract writes the object files into dir and returns their paths, in
// link order. Each keeps its own name, in a subdirectory of its own.
func (p *Program) Extract(dir string) ([]string, error) {
	var paths []string
	for i, name := range p.Objects {
		data, err := p.read(objectEntry(i, name))
		if err != nil {
			return nil, fmt.Errorf("object %s: %v", name, err)
		}
		sub := filepath.Join(dir, fmt.Sprintf("obj%d", i))
		if err := os.MkdirAll(sub, 0o755); err != nil {
			return nil, err
		}
		path := filepath.Join(sub, filepath.Base(name))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}
