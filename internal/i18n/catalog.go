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

	// source locations
	"loc.line":  {en: "line %d: %s"},
	"loc.macro": {en: "(in macro: %s)"},
	"loc.stub":  {en: "(after returning from main)"},

	// runtime errors
	"hint.prefix":        {en: "hint: "},
	"err.read":           {en: "cannot read memory at %s"},
	"err.write":          {en: "cannot write memory at %s"},
	"hint.null":          {en: "the address is (almost) 0, so the register was probably never set to a valid address (e.g. mov ebx, msg)"},
	"hint.readonly":      {en: "%s is read-only. Put variables that change in section .data or section .bss"},
	"hint.smallAddr":     {en: "that address is not part of your program. Did you mean the value itself? [x] reads the memory AT address x"},
	"hint.unmapped":      {en: "that address is not part of your program's memory (check the pointer and the index/offset)"},
	"err.stackOverflow":  {en: "stack overflow"},
	"hint.stackOverflow": {en: "the stack ran out of space: is there a recursion without a base case, or a loop that pushes without popping?"},
	"err.exec":           {en: "jumped to %s, which is not part of the code"},
	"hint.exec":          {en: "check the target of the jmp/call (a register or memory value used as an address)"},
	"hint.execZero":      {en: "a call or jmp through a register that was never set (it is 0)"},
	"err.runoff":         {en: "the program ran past the end of its code"},
	"hint.exit32":        {en: "did you forget the exit system call?  mov eax, 1 / mov ebx, 0 / int 0x80"},
	"hint.exit64":        {en: "did you forget the exit system call?  mov rax, 60 / mov rdi, 0 / syscall"},
	"err.retStart":       {en: "ret in _start: there is no function to return to"},
	"err.badRet":         {en: "ret jumped to %s, which is not part of the code"},
	"hint.badRet":        {en: "the value on top of the stack was not a return address: make sure every push has a matching pop before ret"},
	"err.divZero":        {en: "division by zero (the divisor %s is 0)"},
	"err.divOverflow":    {en: "division overflow: %s divided by %d does not fit in %s"},
	"hint.divUnsigned":   {en: "div divides the double-width register pair; clear the high half first (%s)"},
	"hint.divSigned":     {en: "idiv divides the double-width register pair; sign-extend first (%s)"},
	"err.unsupported":    {en: "unsupported instruction %s"},
	"hint.fpu":           {en: "floating-point (x87) and SSE/AVX instructions are not supported by nasmx86"},
	"err.privileged":     {en: "%s is a privileged instruction (only the operating system kernel may use it)"},
	"err.invalid":        {en: "invalid instruction (bytes %s)"},
	"hint.invalid":       {en: "the CPU is executing data: did a jump land in the middle of data, or is data placed inside section .text?"},
	"err.syscall32":      {en: "syscall is not available in 32-bit programs: use int 0x80 with 32-bit system call numbers, or write a 64-bit program"},
	"err.badInt":         {en: "int %#x is not available to Linux programs"},
	"hint.dos":           {en: "this is an MS-DOS/BIOS interrupt. On Linux use int 0x80 (32-bit) or syscall (64-bit)"},
	"hint.int":           {en: "Linux programs make system calls with int 0x80 (32-bit) or syscall (64-bit)"},
	"err.int3":           {en: "stopped at int3 (breakpoint)"},
	"hint.int3":          {en: "run with --step to debug interactively"},
	"err.maxSteps":       {en: "stopped after %d instructions: the program may be stuck in an infinite loop"},
	"hint.maxSteps":      {en: "use --max-steps N to allow more (0 = no limit), or --trace to see what repeats"},
}
