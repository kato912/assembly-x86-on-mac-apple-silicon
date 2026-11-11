package i18n

var catalog = map[string]msg{
	// command line
	"usage": {en: `usage: nasmx86 [options] file.asm [program arguments...]

Assemble, link and run a 32- or 64-bit Linux NASM program in one step.

options:
  --trace            print every executed instruction with its source line
                     and the registers/flags it changed (to stderr)
  --step             interactive debugger (type h at the prompt for help)
  -32, -64           force 32-bit or 64-bit mode (default: auto-detect)
  -f elf32|elf64     same as -32/-64 (nasm style; -o and -g are ignored,
                     other nasm options such as -D or -I are passed to nasm)
  --max-steps N      stop after N instructions (default 50000000, 0 = no limit)
  --regs             print the registers when the program ends
  --th               Thai messages (or set NASMX86_LANG=th)
  --version          print the version
  -h, --help         show this help

examples:
  nasmx86 hello.asm
  nasmx86 --trace loop.asm
  nasmx86 -f elf32 hello.asm
  echo 5 | nasmx86 fact.asm
`},
	"err.prefix":        {en: "nasmx86: "},
	"err.noFile":        {en: "no .asm file given"},
	"err.missingValue":  {en: "option %s needs a value"},
	"err.badSteps":      {en: "--max-steps needs a number, got %q"},
	"err.unknownFormat": {en: "unsupported output format %q: nasmx86 runs Linux programs, use -f elf32 or -f elf64"},
	"err.readSource":    {en: "cannot read %s: %v"},
	"err.nasmMissing":   {en: "nasm (the assembler) is not installed."},
	"err.bits16":        {en: "this looks like 16-bit (DOS) code (%s). nasmx86 runs 32-bit and 64-bit Linux programs only."},
	"err.link":          {en: "link error: %v"},
	"err.undefined": {en: `your program uses %s, which is not defined in this file.
nasmx86 runs pure assembly programs without the C library, so C functions
like printf/scanf/puts are not available. Use Linux system calls instead,
e.g. write (eax=4, int 0x80 in 32-bit / rax=1, syscall in 64-bit).`},
	"err.internal": {en: "internal error: %v"},

	// runtime warnings
	"warn.efault": {en: "warning: %s: the buffer at %#x is not valid memory, so the call failed (returned -EFAULT)"},
	"warn.enosys": {en: "warning: system call %d is not supported by nasmx86 (returned -ENOSYS)"},
	"warn.halt":   {en: "note: the program stopped at hlt. (On real Linux, hlt in a user program crashes with a segmentation fault — use the exit system call.)"},
	"warn.badfd":  {en: "warning: %s: file descriptor %d is not open (returned -EBADF)"},
}
