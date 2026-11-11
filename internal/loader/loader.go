// Package loader is a tiny static linker: it takes the relocatable ELF
// object produced by nasm, lays its sections out at Linux-like addresses,
// resolves symbols and applies relocations, producing a ready-to-run image.
package loader

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// Linux-like load addresses (what GNU ld uses by default).
const (
	Base32 = 0x08048000
	Base64 = 0x400000
	page   = 0x1000
)

// Section is an allocated section placed in memory.
type Section struct {
	Name   string
	Addr   uint64
	Size   uint64
	Data   []byte // nil for .bss-like (NOBITS) sections
	Exec   bool
	Write  bool
	NoBits bool
	index  int
}

// End returns the first address past the section.
func (s *Section) End() uint64 { return s.Addr + s.Size }

// Symbol is a resolved symbol (label).
type Symbol struct {
	Name    string
	Addr    uint64
	Section string // "" for absolute symbols (equ constants)
	Global  bool
}

// Image is a linked program ready to be loaded into memory.
type Image struct {
	Bits      int
	Entry     uint64
	EntryName string // "_start", "main" or "" (start of .text)
	Sections  []*Section
	Symbols   []Symbol // sorted by address
	BrkStart  uint64   // first address after the program (page aligned)
}

// UndefinedError lists symbols used but never defined (e.g. extern printf).
type UndefinedError struct{ Names []string }

func (e *UndefinedError) Error() string {
	return "undefined symbols: " + strings.Join(e.Names, ", ")
}

// Link loads and links the relocatable ELF object at path.
func Link(path string) (*Image, error) {
	f, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if f.Type != elf.ET_REL {
		return nil, fmt.Errorf("%s is not a relocatable object", path)
	}
	img := &Image{}
	switch f.Machine {
	case elf.EM_386:
		img.Bits = 32
	case elf.EM_X86_64:
		img.Bits = 64
	default:
		return nil, fmt.Errorf("unsupported ELF machine %v", f.Machine)
	}

	// 1. Collect allocated sections into three groups like ld does:
	//    code (r-x), read-only data (r--), writable data then bss (rw-).
	var code, rodata, data, bss []*Section
	byIndex := map[int]*Section{}
	for i, s := range f.Sections {
		if s.Flags&elf.SHF_ALLOC == 0 || (s.Type != elf.SHT_PROGBITS && s.Type != elf.SHT_NOBITS) {
			continue
		}
		sec := &Section{
			Name:   s.Name,
			Size:   s.Size,
			Exec:   s.Flags&elf.SHF_EXECINSTR != 0,
			Write:  s.Flags&elf.SHF_WRITE != 0,
			NoBits: s.Type == elf.SHT_NOBITS,
			index:  i,
		}
		if !sec.NoBits {
			sec.Data, err = s.Data()
			if err != nil {
				return nil, fmt.Errorf("reading section %s: %v", s.Name, err)
			}
		}
		byIndex[i] = sec
		switch {
		case sec.Exec:
			code = append(code, sec)
		case sec.NoBits:
			bss = append(bss, sec)
		case sec.Write:
			data = append(data, sec)
		default:
			rodata = append(rodata, sec)
		}
	}

	// Common symbols (NASM `common`) become bss space.
	syms, err := f.Symbols()
	if err != nil && len(f.Sections) > 0 {
		syms = nil // no symbol table at all is fine
	}
	var commonSec *Section
	for _, s := range syms {
		if s.Section == elf.SHN_COMMON {
			if commonSec == nil {
				commonSec = &Section{Name: "COMMON", Write: true, NoBits: true, index: -1}
				bss = append(bss, commonSec)
			}
		}
	}
	commonOff := map[string]uint64{}
	if commonSec != nil {
		for _, s := range syms {
			if s.Section == elf.SHN_COMMON {
				align := s.Value
				if align == 0 {
					align = 1
				}
				commonSec.Size = alignUp(commonSec.Size, align)
				commonOff[s.Name] = commonSec.Size
				commonSec.Size += s.Size
			}
		}
	}

	// 2. Lay out.
	base := uint64(Base64)
	if img.Bits == 32 {
		base = Base32
	}
	addr := base + page
	elfAlign := func(s *Section) uint64 {
		if s.index >= 0 {
			if a := f.Sections[s.index].Addralign; a > 1 {
				return a
			}
		}
		return 1
	}
	for gi, group := range [][]*Section{code, rodata, append(data, bss...)} {
		if len(group) == 0 {
			continue
		}
		if gi > 0 {
			addr = alignUp(addr, page)
		}
		for _, s := range group {
			addr = alignUp(addr, elfAlign(s))
			s.Addr = addr
			addr += s.Size
		}
		img.Sections = append(img.Sections, group...)
	}
	img.BrkStart = alignUp(addr, page)

	// 3. Resolve symbols.
	symAddr := make([]uint64, len(syms))
	var undefined []string
	seenUndef := map[string]bool{}
	for i, s := range syms {
		st := elf.ST_TYPE(s.Info)
		switch {
		case s.Section == elf.SHN_UNDEF:
			if s.Name != "" && !seenUndef[s.Name] {
				seenUndef[s.Name] = true
				undefined = append(undefined, s.Name)
			}
		case s.Section == elf.SHN_ABS:
			symAddr[i] = s.Value
		case s.Section == elf.SHN_COMMON:
			symAddr[i] = commonSec.Addr + commonOff[s.Name]
		default:
			sec := byIndex[int(s.Section)]
			if sec == nil {
				continue // symbol in a non-allocated (debug) section
			}
			symAddr[i] = sec.Addr + s.Value
		}
		if st == elf.STT_SECTION || st == elf.STT_FILE || s.Name == "" || s.Section == elf.SHN_UNDEF {
			continue
		}
		if strings.HasPrefix(s.Name, "..@") { // NASM macro-local labels
			continue
		}
		secName := ""
		if sec := byIndex[int(s.Section)]; sec != nil {
			secName = sec.Name
		} else if s.Section == elf.SHN_COMMON {
			secName = "COMMON"
		}
		img.Symbols = append(img.Symbols, Symbol{
			Name:    s.Name,
			Addr:    symAddr[i],
			Section: secName,
			Global:  elf.ST_BIND(s.Info) == elf.STB_GLOBAL,
		})
	}
	sort.SliceStable(img.Symbols, func(i, j int) bool { return img.Symbols[i].Addr < img.Symbols[j].Addr })

	// 4. Apply relocations.
	for _, rs := range f.Sections {
		if rs.Type != elf.SHT_REL && rs.Type != elf.SHT_RELA {
			continue
		}
		target := byIndex[int(rs.Info)]
		if target == nil {
			continue // relocations for debug sections
		}
		raw, err := rs.Data()
		if err != nil {
			return nil, err
		}
		if err := relocate(img.Bits, rs.Type == elf.SHT_RELA, raw, target, syms, symAddr, seenUndef); err != nil {
			return nil, err
		}
	}
	if len(undefined) > 0 {
		return nil, &UndefinedError{Names: undefined}
	}

	// 5. Entry point.
	for _, name := range []string{"_start", "main"} {
		if s, ok := img.Lookup(name); ok && s.Section != "" {
			img.Entry, img.EntryName = s.Addr, name
			return img, nil
		}
	}
	if len(code) == 0 || code[0].Size == 0 {
		return nil, fmt.Errorf("no code found: the program has no instructions in section .text")
	}
	for _, s := range code {
		if s.Name == ".text" {
			img.Entry = s.Addr
			return img, nil
		}
	}
	img.Entry = code[0].Addr
	return img, nil
}

func relocate(bits int, rela bool, raw []byte, target *Section, syms []elf.Symbol, symAddr []uint64, undef map[string]bool) error {
	le := binary.LittleEndian
	entSize := 8
	switch {
	case bits == 64 && rela:
		entSize = 24
	case bits == 64:
		entSize = 16
	case rela:
		entSize = 12
	}
	for off := 0; off+entSize <= len(raw); off += entSize {
		var r uint64
		var symIdx int
		var typ uint32
		var addend int64
		if bits == 64 {
			r = le.Uint64(raw[off:])
			info := le.Uint64(raw[off+8:])
			symIdx, typ = int(info>>32), uint32(info)
			if rela {
				addend = int64(le.Uint64(raw[off+16:]))
			}
		} else {
			r = uint64(le.Uint32(raw[off:]))
			info := le.Uint32(raw[off+4:])
			symIdx, typ = int(info>>8), info&0xff
			if rela {
				addend = int64(int32(le.Uint32(raw[off+8:])))
			}
		}
		var S uint64
		symName := ""
		if symIdx > 0 && symIdx <= len(syms) {
			S = symAddr[symIdx-1]
			symName = syms[symIdx-1].Name
			if syms[symIdx-1].Section == elf.SHN_UNDEF {
				continue // reported as undefined symbol
			}
		}
		if target.Data == nil || r >= uint64(len(target.Data)) {
			return fmt.Errorf("relocation outside section %s", target.Name)
		}
		P := target.Addr + r
		d := target.Data[r:]
		width, pcrel, signed := 0, false, false
		if bits == 64 {
			switch elf.R_X86_64(typ) {
			case elf.R_X86_64_NONE:
				continue
			case elf.R_X86_64_64:
				width = 8
			case elf.R_X86_64_PC64:
				width, pcrel = 8, true
			case elf.R_X86_64_32:
				width = 4
			case elf.R_X86_64_32S:
				width, signed = 4, true
			case elf.R_X86_64_PC32, elf.R_X86_64_PLT32:
				width, pcrel, signed = 4, true, true
			case elf.R_X86_64_16:
				width = 2
			case elf.R_X86_64_PC16:
				width, pcrel, signed = 2, true, true
			case elf.R_X86_64_8:
				width = 1
			case elf.R_X86_64_PC8:
				width, pcrel, signed = 1, true, true
			default:
				return fmt.Errorf("unsupported relocation %v (symbol %q) in %s", elf.R_X86_64(typ), symName, target.Name)
			}
		} else {
			switch elf.R_386(typ) {
			case elf.R_386_NONE:
				continue
			case elf.R_386_32:
				width = 4
			case elf.R_386_PC32, elf.R_386_PLT32:
				width, pcrel = 4, true
			case elf.R_386_16:
				width = 2
			case elf.R_386_PC16:
				width, pcrel, signed = 2, true, true
			case elf.R_386_8:
				width = 1
			case elf.R_386_PC8:
				width, pcrel, signed = 1, true, true
			default:
				return fmt.Errorf("unsupported relocation %v (symbol %q) in %s", elf.R_386(typ), symName, target.Name)
			}
		}
		if len(d) < width {
			return fmt.Errorf("truncated relocation in %s", target.Name)
		}
		if !rela { // implicit addend stored in place
			switch width {
			case 1:
				addend = int64(int8(d[0]))
			case 2:
				addend = int64(int16(le.Uint16(d)))
			case 4:
				addend = int64(int32(le.Uint32(d)))
			case 8:
				addend = int64(le.Uint64(d))
			}
		}
		v := S + uint64(addend)
		if pcrel {
			v -= P
		}
		if width < 8 && bits == 64 {
			sv := int64(v)
			lim := int64(1) << (8*width - 1)
			ok := (signed && sv >= -lim && sv < lim) || (!signed && v < uint64(lim)*2)
			if !ok {
				return fmt.Errorf("address of %q does not fit in a %d-bit field (in %s)", symName, 8*width, target.Name)
			}
		}
		switch width {
		case 1:
			d[0] = byte(v)
		case 2:
			le.PutUint16(d, uint16(v))
		case 4:
			le.PutUint32(d, uint32(v))
		case 8:
			le.PutUint64(d, v)
		}
	}
	return nil
}

// Lookup finds a symbol by name.
func (img *Image) Lookup(name string) (Symbol, bool) {
	for _, s := range img.Symbols {
		if s.Name == name {
			return s, true
		}
	}
	return Symbol{}, false
}

// SectionAt returns the section containing addr, or nil.
func (img *Image) SectionAt(addr uint64) *Section {
	for _, s := range img.Sections {
		if addr >= s.Addr && addr < s.End() {
			return s
		}
	}
	return nil
}

// SectionByName returns the named section, or nil.
func (img *Image) SectionByName(name string) *Section {
	for _, s := range img.Sections {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// SymbolFor returns the closest label at or before addr (for messages like
// "in loop+0x4"). Only labels in sections are considered.
func (img *Image) SymbolFor(addr uint64) (Symbol, bool) {
	i := sort.Search(len(img.Symbols), func(i int) bool { return img.Symbols[i].Addr > addr })
	for i--; i >= 0; i-- {
		s := img.Symbols[i]
		if s.Section != "" {
			if sec := img.SectionAt(addr); sec != nil && sec.Name == s.Section {
				return s, true
			}
			return Symbol{}, false
		}
	}
	return Symbol{}, false
}

func alignUp(v, a uint64) uint64 {
	if a <= 1 {
		return v
	}
	return (v + a - 1) / a * a
}
