package loader

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"sort"
	"strings"
)

// ExtraSymbol is a symbol added to a written executable that is not part
// of the image, such as the built-in C library's functions.
type ExtraSymbol struct {
	Name string
	Addr uint64
}

// ExtraRegion is an address range written as an allocated NOBITS section
// so a debugger can name addresses in it (the C library).
type ExtraRegion struct {
	Name    string
	Addr    uint64
	Size    uint64
	Symbols []ExtraSymbol
}

// WriteELF writes the image as an ELF executable (ET_EXEC) with its
// symbols and linked DWARF sections, for a debugger such as gdb to read
// while it debugs the program running in nasmx86. The file is laid out
// like GNU ld output: one PT_LOAD per section, at the same addresses the
// emulator uses.
func (img *Image) WriteELF(path string, extra []ExtraRegion) error {
	b, err := img.elfBytes(extra)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o755)
}

type outSection struct {
	name             string
	typ              elf.SectionType
	flags            elf.SectionFlag
	addr, size, off  uint64
	align            uint64
	data             []byte
	link, info, ents uint32
}

func (img *Image) elfBytes(extra []ExtraRegion) ([]byte, error) {
	is64 := img.Bits == 64
	le := binary.LittleEndian
	var ehSize, phEnt, shEnt uint64 = 52, 32, 40
	if is64 {
		ehSize, phEnt, shEnt = 64, 56, 64
	}

	// Sections: null, the image's sections, extra regions, debug, symtab,
	// strtab, shstrtab.
	secs := []*outSection{{}}
	secIndex := map[*Section]int{}
	for _, s := range img.Sections {
		o := &outSection{name: s.Name, typ: elf.SHT_PROGBITS, flags: elf.SHF_ALLOC, addr: s.Addr, size: s.Size, align: max(s.align, 1), data: s.Data}
		if s.Write {
			o.flags |= elf.SHF_WRITE
		}
		if s.Exec {
			o.flags |= elf.SHF_EXECINSTR
		}
		if s.NoBits || s.Data == nil {
			o.typ, o.data = elf.SHT_NOBITS, nil
		}
		secIndex[s] = len(secs)
		secs = append(secs, o)
	}
	extraIndex := make([]int, len(extra))
	for i, r := range extra {
		extraIndex[i] = len(secs)
		secs = append(secs, &outSection{name: r.Name, typ: elf.SHT_NOBITS, flags: elf.SHF_ALLOC | elf.SHF_EXECINSTR, addr: r.Addr, size: r.Size, align: 16})
	}
	for _, d := range img.Debug {
		secs = append(secs, &outSection{name: d.Name, typ: elf.SHT_PROGBITS, size: uint64(len(d.Data)), align: 1, data: d.Data})
	}

	// Symbol table: locals first, as ELF requires.
	var strtab bytes.Buffer
	strtab.WriteByte(0)
	str := func(s string) uint32 {
		off := strtab.Len()
		strtab.WriteString(s)
		strtab.WriteByte(0)
		return uint32(off)
	}
	type sym struct {
		name   uint32
		value  uint64
		size   uint64
		info   byte
		shndx  uint16
		global bool
	}
	var syms []sym
	for _, s := range img.Symbols {
		shndx := uint16(elf.SHN_ABS)
		typ := elf.STT_NOTYPE
		if s.Section != "" {
			if sec := img.SectionAt(s.Addr); sec != nil {
				shndx = uint16(secIndex[sec])
				if sec.Exec && !strings.Contains(s.Name, ".") {
					typ = elf.STT_FUNC
				} else if !sec.Exec {
					typ = elf.STT_OBJECT
				}
			} else {
				// a label just past the end of a section (msg_end: after the data)
				for _, sec := range img.Sections {
					if sec.End() == s.Addr {
						shndx = uint16(secIndex[sec])
					}
				}
			}
		}
		bind := elf.STB_LOCAL
		if s.Global {
			bind = elf.STB_GLOBAL
		}
		syms = append(syms, sym{name: str(s.Name), value: s.Addr, info: elf.ST_INFO(bind, typ), shndx: shndx, global: s.Global})
	}
	for i, r := range extra {
		for _, e := range r.Symbols {
			syms = append(syms, sym{name: str(e.Name), value: e.Addr, size: 16, info: elf.ST_INFO(elf.STB_GLOBAL, elf.STT_FUNC), shndx: uint16(extraIndex[i]), global: true})
		}
	}
	sort.SliceStable(syms, func(i, j int) bool { return !syms[i].global && syms[j].global })
	firstGlobal := uint32(1)
	for _, s := range syms {
		if !s.global {
			firstGlobal++
		}
	}
	var symtab bytes.Buffer
	symEnt := 16
	if is64 {
		symEnt = 24
	}
	symtab.Write(make([]byte, symEnt)) // the null symbol
	for _, s := range syms {
		if is64 {
			binary.Write(&symtab, le, elf.Sym64{Name: s.name, Info: s.info, Shndx: s.shndx, Value: s.value, Size: s.size})
		} else {
			binary.Write(&symtab, le, elf.Sym32{Name: s.name, Value: uint32(s.value), Size: uint32(s.size), Info: s.info, Shndx: s.shndx})
		}
	}
	symIdx := len(secs)
	secs = append(secs, &outSection{name: ".symtab", typ: elf.SHT_SYMTAB, size: uint64(symtab.Len()), align: 8, data: symtab.Bytes(),
		link: uint32(symIdx + 1), info: firstGlobal, ents: uint32(symEnt)})
	secs = append(secs, &outSection{name: ".strtab", typ: elf.SHT_STRTAB, size: uint64(strtab.Len()), align: 1, data: strtab.Bytes()})
	var shstr bytes.Buffer
	shstr.WriteByte(0)
	names := make([]uint32, len(secs)+1)
	for i, s := range secs {
		if i > 0 {
			names[i] = uint32(shstr.Len())
			shstr.WriteString(s.name)
			shstr.WriteByte(0)
		}
	}
	names[len(secs)] = uint32(shstr.Len())
	shstr.WriteString(".shstrtab")
	shstr.WriteByte(0)
	shstrIdx := len(secs)
	secs = append(secs, &outSection{name: ".shstrtab", typ: elf.SHT_STRTAB, size: uint64(shstr.Len()), align: 1, data: shstr.Bytes()})

	// Program headers: one PT_LOAD per allocated section with contents.
	var loads []*outSection
	for _, s := range secs[1:] {
		if s.flags&elf.SHF_ALLOC != 0 && s.size > 0 && s.name != "" && !isExtra(s, extra) {
			loads = append(loads, s)
		}
	}

	// File layout.
	off := ehSize + uint64(len(loads))*phEnt
	for _, s := range secs[1:] {
		if s.typ == elf.SHT_NOBITS {
			s.off = off
			continue
		}
		if s.flags&elf.SHF_ALLOC != 0 {
			// file offset congruent to the address modulo the page size
			for off%page != s.addr%page {
				off++
			}
		} else {
			off = alignUp(off, s.align)
		}
		s.off = off
		off += uint64(len(s.data))
	}
	shOff := alignUp(off, 8)

	var out bytes.Buffer
	if is64 {
		h := elf.Header64{Type: uint16(elf.ET_EXEC), Machine: uint16(elf.EM_X86_64), Version: uint32(elf.EV_CURRENT), Entry: img.Entry,
			Phoff: ehSize, Shoff: shOff, Ehsize: uint16(ehSize), Phentsize: uint16(phEnt), Phnum: uint16(len(loads)),
			Shentsize: uint16(shEnt), Shnum: uint16(len(secs)), Shstrndx: uint16(shstrIdx)}
		copy(h.Ident[:], []byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)})
		binary.Write(&out, le, h)
	} else {
		h := elf.Header32{Type: uint16(elf.ET_EXEC), Machine: uint16(elf.EM_386), Version: uint32(elf.EV_CURRENT), Entry: uint32(img.Entry),
			Phoff: uint32(ehSize), Shoff: uint32(shOff), Ehsize: uint16(ehSize), Phentsize: uint16(phEnt), Phnum: uint16(len(loads)),
			Shentsize: uint16(shEnt), Shnum: uint16(len(secs)), Shstrndx: uint16(shstrIdx)}
		copy(h.Ident[:], []byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS32), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)})
		binary.Write(&out, le, h)
	}
	for _, s := range loads {
		flags := elf.PF_R
		if s.flags&elf.SHF_WRITE != 0 {
			flags |= elf.PF_W
		}
		if s.flags&elf.SHF_EXECINSTR != 0 {
			flags |= elf.PF_X
		}
		filesz := uint64(len(s.data))
		if is64 {
			binary.Write(&out, le, elf.Prog64{Type: uint32(elf.PT_LOAD), Flags: uint32(flags), Off: s.off, Vaddr: s.addr, Paddr: s.addr,
				Filesz: filesz, Memsz: s.size, Align: page})
		} else {
			binary.Write(&out, le, elf.Prog32{Type: uint32(elf.PT_LOAD), Off: uint32(s.off), Vaddr: uint32(s.addr), Paddr: uint32(s.addr),
				Filesz: uint32(filesz), Memsz: uint32(s.size), Flags: uint32(flags), Align: page})
		}
	}
	for _, s := range secs[1:] {
		if s.typ == elf.SHT_NOBITS {
			continue
		}
		for uint64(out.Len()) < s.off {
			out.WriteByte(0)
		}
		out.Write(s.data)
	}
	for uint64(out.Len()) < shOff {
		out.WriteByte(0)
	}
	for i, s := range secs {
		if is64 {
			binary.Write(&out, le, elf.Section64{Name: names[i], Type: uint32(s.typ), Flags: uint64(s.flags), Addr: s.addr, Off: s.off,
				Size: s.size, Link: s.link, Info: s.info, Addralign: s.align, Entsize: uint64(s.ents)})
		} else {
			binary.Write(&out, le, elf.Section32{Name: names[i], Type: uint32(s.typ), Flags: uint32(s.flags), Addr: uint32(s.addr), Off: uint32(s.off),
				Size: uint32(s.size), Link: s.link, Info: s.info, Addralign: uint32(s.align), Entsize: s.ents})
		}
	}
	return out.Bytes(), nil
}

func isExtra(s *outSection, extra []ExtraRegion) bool {
	for _, r := range extra {
		if s.addr == r.Addr && s.name == r.Name && s.typ == elf.SHT_NOBITS {
			return true
		}
	}
	return false
}
