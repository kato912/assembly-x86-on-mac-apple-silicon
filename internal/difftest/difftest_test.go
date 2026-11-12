// Package difftest runs the differential test programs in testdata/diff
// (and the examples) through nasmx86 and, on x86-64 Linux, natively with
// nasm + ld, requiring identical stdout and exit status. Programs in
// testdata/diff/libc32 and libc64 call the C library; natively they are
// linked with gcc against glibc.
//
// Environment:
//
//	NASMX86_NATIVE=0          never run natively (emulator-only smoke test)
//	NASMX86_REQUIRE_NATIVE=1  fail instead of skipping the native comparison
package difftest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/nasm"
	"github.com/kato912/assembly-x86-on-mac-apple-silicon/internal/runner"
)

const (
	diffDir    = "../../testdata/diff"
	exampleDir = "../../examples"
	inputDir   = "../../testdata/examples"
)

type program struct {
	name     string
	path     string   // main source file
	extra    []string // more source files linked with it (multi-file programs)
	bits     int      // 0 = let nasmx86 auto-detect (examples)
	stdin    []byte
	wantExit int
	incDir   string
	libc     bool // link with gcc + glibc natively
}

var exitRE = regexp.MustCompile(`(?m)^;\s*exit:\s*(\d+)`)

// usesLibcRE spots examples that call the C library (linked with gcc natively).
var usesLibcRE = regexp.MustCompile(`(?mi)^\s*extern\s+[^;\n]*\b(?:printf|scanf|puts|malloc)\b`)

// collect finds the test programs: every *.asm file in testdata/diff/32,
// testdata/diff/64 and examples, and every subdirectory there, which is one
// program made of all its .asm files (main.asm first).
func collect(t *testing.T) []program {
	var progs []program
	inc, _ := filepath.Abs(filepath.Join(diffDir, "inc"))
	libc := false
	add := func(name string, files []string, bits int, inPath string) {
		src, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		p := program{name: name, path: files[0], extra: files[1:], bits: bits, incDir: inc, libc: libc || usesLibcRE.Match(src)}
		if m := exitRE.FindSubmatch(src); m != nil {
			p.wantExit, _ = strconv.Atoi(string(m[1]))
		}
		if b, err := os.ReadFile(inPath); err == nil {
			p.stdin = b
		}
		progs = append(progs, p)
	}
	scan := func(dir string, bits int, inputFor func(base string, path string) string) {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			path := filepath.Join(dir, e.Name())
			label := filepath.Base(dir) + "/" + strings.TrimSuffix(e.Name(), ".asm")
			switch {
			case e.IsDir():
				files, _ := filepath.Glob(filepath.Join(path, "*.asm"))
				sort.SliceStable(files, func(i, j int) bool {
					return filepath.Base(files[i]) == "main.asm" && filepath.Base(files[j]) != "main.asm"
				})
				if len(files) > 0 {
					add(label, files, bits, inputFor(e.Name(), filepath.Join(path, "main")))
				}
			case strings.HasSuffix(e.Name(), ".asm"):
				add(label, []string{path}, bits, inputFor(strings.TrimSuffix(e.Name(), ".asm"), strings.TrimSuffix(path, ".asm")))
			}
		}
	}
	for _, bits := range []int{32, 64} {
		scan(filepath.Join(diffDir, strconv.Itoa(bits)), bits, func(_, stem string) string { return stem + ".in" })
	}
	libc = true
	for _, bits := range []int{32, 64} {
		scan(filepath.Join(diffDir, "libc"+strconv.Itoa(bits)), bits, func(_, stem string) string { return stem + ".in" })
	}
	libc = false
	scan(exampleDir, 0, func(base, _ string) string { return filepath.Join(inputDir, base+".in") })
	if len(progs) == 0 {
		t.Fatal("no test programs found")
	}
	return progs
}

func TestDifferential(t *testing.T) {
	if _, err := nasm.Find(); err != nil {
		t.Skip("nasm not installed")
	}
	native, why := nativeAvailable()
	if !native {
		if os.Getenv("NASMX86_REQUIRE_NATIVE") == "1" {
			t.Fatalf("native comparison required but unavailable: %s", why)
		}
		t.Logf("native comparison skipped (%s); checking nasmx86 runs only", why)
	}
	for _, p := range collect(t) {
		p := p
		t.Run(p.name, func(t *testing.T) {
			emuOut, emuErr, emuCode := runEmu(p)
			if emuCode != p.wantExit {
				t.Errorf("nasmx86 exit status %d, want %d\nstderr:\n%s", emuCode, p.wantExit, emuErr)
			}
			if !native {
				return
			}
			if p.libc {
				bits := p.bits
				if bits == 0 {
					src, _ := os.ReadFile(p.path)
					bits, _ = nasm.DetectBitsFile(p.path, src, []string{p.incDir})
				}
				if ok, why := libcAvailable(bits); !ok {
					if os.Getenv("NASMX86_REQUIRE_NATIVE") == "1" {
						t.Fatalf("native comparison required but unavailable: %s", why)
					}
					t.Skipf("native comparison skipped: %s", why)
				}
			}
			natOut, natCode, err := runNative(t, p)
			if err != nil {
				t.Fatalf("native build/run: %v", err)
			}
			if natCode != p.wantExit {
				t.Errorf("native exit status %d, want %d (the test program itself is wrong)", natCode, p.wantExit)
			}
			if natCode != emuCode {
				t.Errorf("exit status: native %d, nasmx86 %d\nnasmx86 stderr:\n%s", natCode, emuCode, emuErr)
			}
			if !bytes.Equal(natOut, emuOut) {
				t.Errorf("stdout differs from native run:\n%s", describeDiff(natOut, emuOut))
			}
		})
	}
}

func runEmu(p program) (stdout []byte, stderr string, code int) {
	var out, errb bytes.Buffer
	code = runner.Run(runner.Config{
		Source:   p.path,
		Extra:    p.extra,
		Bits:     p.bits,
		NasmArgs: []string{"-I" + p.incDir + string(filepath.Separator)},
		MaxSteps: runner.DefaultMaxSteps,
		Stdin:    bytes.NewReader(p.stdin),
		Stdout:   &out,
		Stderr:   &errb,
	})
	return out.Bytes(), errb.String(), code
}

// describeDiff shows the first few differing lines, highlighting the
// space-separated fields (registers, flags) that differ.
func describeDiff(native, emu []byte) string {
	a := strings.Split(string(native), "\n")
	b := strings.Split(string(emu), "\n")
	var sb strings.Builder
	shown := 0
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y string
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x == y {
			continue
		}
		fmt.Fprintf(&sb, "line %d:\n  native:  %q\n  nasmx86: %q\n", i+1, x, y)
		xf, yf := strings.Fields(x), strings.Fields(y)
		var diffs []string
		for j := 0; j < len(xf) && j < len(yf); j++ {
			if xf[j] != yf[j] {
				diffs = append(diffs, xf[j]+" vs "+yf[j])
			}
		}
		if len(diffs) > 0 {
			fmt.Fprintf(&sb, "  fields:  %s\n", strings.Join(diffs, ", "))
		}
		if shown++; shown == 5 {
			sb.WriteString("  ...\n")
			break
		}
	}
	return sb.String()
}
