package nasm

import (
	"debug/dwarf"
	"debug/elf"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DWARFLines builds a LineMap from the DWARF line table of an object file
// assembled with debug information (yasm -g dwarf2, nasm -g -F dwarf), so
// programs linked from object files still report source lines. The map's
// Source is the main source file; lines from other files (%include) carry
// their file name in Loc.File.
func DWARFLines(objPath string) (*LineMap, error) {
	f, err := elf.Open(objPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if f.Section(".debug_line") == nil {
		return nil, errors.New("no DWARF line table")
	}
	d, err := f.DWARF() // applies the object's relocations to the debug sections
	if err != nil {
		return nil, err
	}
	// Line addresses in an object file are offsets into its code section.
	section := ""
	for _, s := range f.Sections {
		if s.Flags&elf.SHF_EXECINSTR != 0 && (section == "" || s.Name == ".text") {
			section = s.Name
		}
	}
	if section == "" {
		return nil, errors.New("no code section")
	}
	lm := &LineMap{sections: map[string][]entry{}, lines: map[int][]SecOff{}, texts: map[int]string{}}
	texts := map[string][]string{} // source file -> its lines
	text := func(name string, line int) string {
		lines, ok := texts[name]
		if !ok {
			lines = readLines(name, objPath)
			texts[name] = lines
		}
		if line >= 1 && line <= len(lines) {
			return strings.TrimSpace(lines[line-1])
		}
		return ""
	}
	r := d.Reader()
	for {
		cu, err := r.Next()
		if err != nil || cu == nil {
			break
		}
		if cu.Tag != dwarf.TagCompileUnit {
			r.SkipChildren()
			continue
		}
		main, _ := cu.Val(dwarf.AttrName).(string)
		if dir, _ := cu.Val(dwarf.AttrCompDir).(string); dir != "" && main != "" && !filepath.IsAbs(main) {
			main = filepath.Join(dir, main)
		}
		if lm.Source == "" {
			lm.Source = main
		}
		lr, err := d.LineReader(cu)
		if err != nil || lr == nil {
			r.SkipChildren()
			continue
		}
		var e dwarf.LineEntry
		for {
			if err := lr.Next(&e); err != nil {
				if err != io.EOF {
					return nil, err
				}
				break
			}
			if e.EndSequence || e.File == nil || e.Line <= 0 {
				continue
			}
			loc := Loc{Line: e.Line, Text: text(e.File.Name, e.Line)}
			if filepath.Base(e.File.Name) != filepath.Base(lm.Source) {
				loc.File = filepath.Base(e.File.Name)
			} else if loc.Text != "" {
				lm.texts[e.Line] = loc.Text
			}
			lm.add(section, e.Address, loc)
		}
		r.SkipChildren()
	}
	if len(lm.sections) == 0 {
		return nil, errors.New("empty DWARF line table")
	}
	for s := range lm.sections {
		es := lm.sections[s]
		sort.SliceStable(es, func(i, j int) bool { return es[i].off < es[j].off })
	}
	return lm, nil
}

// readLines reads a source file named in debug information: as given,
// else next to the object file, else in the current directory.
func readLines(name, objPath string) []string {
	for _, p := range []string{name, filepath.Join(filepath.Dir(objPath), filepath.Base(name)), filepath.Base(name)} {
		if data, err := os.ReadFile(p); err == nil {
			return strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		}
	}
	return nil
}
