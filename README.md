# nasmx86

**Run x86 assembly with one command — on a Mac (Apple Silicon or Intel), Windows or Linux. No VM, no Docker, no Rosetta.**

[ภาษาไทย](README.th.md) · English

```
$ nasmx86 hello.asm
Hello, World!
```

`nasmx86` assembles your NASM program with the real `nasm`, links it with a
built-in mini linker, and runs it on a built-in x86 CPU emulator with Linux
system calls. It supports both **32-bit** (`int 0x80`) and **64-bit**
(`syscall`) Linux programs — exactly what introductory x86 assembly courses use.

When something goes wrong it tells you *which line* and *why*, instead of
`Segmentation fault`:

```
line 14: div ecx  →  division overflow: edx:eax = 0x00000009:0x00000064 divided by 4 does not fit in eax
    hint: div divides the double-width register pair; clear the high half first (xor edx, edx)
```

---

## Install

### macOS with Homebrew — recommended

```bash
brew install kato912/tap/nasmx86
```

This also installs `nasm`. No Homebrew account or GitHub login is needed, and
it works even if Xcode is outdated or not installed. Update later with
`brew upgrade nasmx86`.

### Linux (or macOS without Homebrew)

```bash
curl -fsSL https://raw.githubusercontent.com/kato912/assembly-x86-on-mac-apple-silicon/main/install.sh | sh
```

The script downloads the right binary for your OS/CPU into `/usr/local/bin`
(or `~/.local/bin`), verifies its checksum, adds it to your `PATH`, and installs
`nasm` if it can (`brew`, `apt`, `dnf`, `pacman`, ...). Example programs go to
`~/.local/share/nasmx86/examples`.

### Windows

1. Install NASM: `winget install NASM.NASM`
   (or the installer from <https://www.nasm.us/pub/nasm/releasebuilds/>).
2. Download `nasmx86_<version>_windows_amd64.zip` from the
   [Releases page](https://github.com/kato912/assembly-x86-on-mac-apple-silicon/releases),
   unzip it and put `nasmx86.exe` in a folder on your `PATH`
   (see `INSTALL-WINDOWS.txt` inside the zip).
3. Open a **new** terminal and run `nasmx86 examples\hello32.asm`.

nasmx86 finds `nasm` in `C:\Program Files\NASM` even when it is not on `PATH`.

### From source

```bash
go install github.com/kato912/assembly-x86-on-mac-apple-silicon/cmd/nasmx86@latest
```

---

## Quick start

`hello.asm` (32-bit):

```nasm
section .data
    msg db "Hello, World!", 10
    len equ $ - msg

section .text
    global _start
_start:
    mov eax, 4          ; sys_write
    mov ebx, 1          ; stdout
    mov ecx, msg
    mov edx, len
    int 0x80

    mov eax, 1          ; sys_exit
    mov ebx, 0
    int 0x80
```

```bash
nasmx86 hello.asm                 # assemble + link + run
echo 5 | nasmx86 factorial32.asm  # stdin works like on Linux
nasmx86 prog.asm arg1 arg2        # arguments go to the program (argc/argv on the stack)
nasmx86 prog.asm; echo $?         # the exit code is the program's exit code
nasmx86 main.asm io.asm           # a program split over several files
```

More examples are in [`examples/`](examples): reading input, loops, arrays,
functions with stack arguments, recursion (factorial), a program split over
two files ([`examples/multifile`](examples/multifile)) — in 32 and 64 bits.

### Programs with several files

List all the `.asm` files, main program first; labels shared between them are
declared `global` where they are defined and `extern` where they are used,
exactly as with `nasm` + `ld`:

```bash
nasmx86 main.asm io.asm mathlib.asm   # arguments for the program go after the files
nasmx86 main.asm io.o                 # files already assembled with nasm -f elf32/elf64 work too
```

Errors and `--trace` name the file (`io.asm line 12: ...`), and the debugger
accepts `b io.asm:12`. To pass an argument that ends in `.asm` to the program,
put `--` before it.

### Calling C functions (`printf`, `scanf`, ...)

Programs that call the C library — the "calling C from assembly" chapter of
many courses — run as they are. Declare the functions `extern`, write `main`,
and use the normal calling conventions (64-bit: arguments in
`rdi, rsi, rdx, rcx, r8, r9`, then the stack, and `al` = 0; 32-bit: push the
arguments right to left and pop them after the call):

```nasm
extern printf
section .data
    fmt db "%d + %d = %d", 10, 0
section .text
global main
main:
    push rbp            ; keeps the stack 16-byte aligned for the call
    mov rdi, fmt
    mov rsi, 2
    mov rdx, 3
    lea rcx, [rsi+rdx]
    xor eax, eax        ; no floating-point arguments
    call printf
    pop rbp
    xor eax, eax        ; return 0
    ret
```

nasmx86 has its own C library built in (no gcc or glibc needed) that behaves
like glibc, so what works here works on Linux with `gcc -no-pie prog.o`
([`examples/printf64.asm`](examples/printf64.asm)):

- output is buffered like glibc: line by line on a terminal, in blocks when
  piped. `exit` and returning from `main` write it; ending with the exit
  *system call* (or crashing) loses it — on Linux too — and nasmx86 tells you;
- a call may change `rax rcx rdx rsi rdi r8-r11` (`eax ecx edx` in 32-bit),
  as real C functions do: a loop counter kept in `rcx` across `call printf`
  breaks here just as it does on Linux;
- a misaligned stack, passing a value where `scanf` needs an address, `free`
  twice, `mov rdi, stdout` instead of `mov rdi, [stdout]`... are reported at
  the `call` line with a hint;
- `--trace` and the debugger show each call: `call printf → printf("x = %d\n", 5) = 6`.

Available functions: `printf fprintf sprintf snprintf dprintf puts fputs
putchar putc fputc fwrite fflush scanf fscanf sscanf getchar getc fgetc ungetc
fgets gets fopen fclose feof` (and `stdin stdout stderr`), `exit _exit abort
atoi atol atoll strtol strtoul abs labs malloc calloc realloc free rand srand`,
`strlen strcpy strncpy strcat strncat strcmp strncmp strchr strrchr strstr
strdup memset memcpy memmove memcmp`, `toupper tolower isdigit isalpha isalnum
isspace isupper islower isxdigit ispunct isprint`, `time sleep usleep`.
In 64-bit mode `printf("%f")` is not available (doubles go in SSE registers,
which nasmx86 does not emulate); in 32-bit mode doubles pushed on the stack
print fine. `call printf wrt ..plt` and `call [rel printf wrt ..got]` work too.

### YASM

Course material written for YASM runs too: nasmx86 accepts yasm's command
line (`nasmx86 -f elf64 -g dwarf2 prog.asm`, `-o`/`-l` are ignored), `.s`
files, and YASM's `ddq`/`resdq` directives. The source is assembled with
NASM, whose syntax YASM shares.

### 32 or 64 bit?

nasmx86 decides automatically (it also looks inside `%include` files):

1. a `bits 32` / `bits 64` directive wins;
2. 64-bit registers (`rax`, `r8`, ...) or `default rel` → 64-bit;
3. `syscall` → 64-bit, `int 0x80` → 32-bit;
4. `esp`/`ebp` (32-bit stack code such as `push x` / `call printf` / `add esp, 8`) → 32-bit;
5. otherwise 64-bit.

Force it with `-32` / `-64`, or with the nasm flag from your class sheet:
`nasmx86 -f elf32 hello.asm`.

---

## Options

| Option | Meaning |
|---|---|
| `--trace` | print every executed instruction with its source line and what changed (registers, flags, memory, system calls) — to stderr |
| `--step` | interactive debugger (see below) |
| `--ddd` | debug in DDD, the graphical debugger, through gdb ([below](#ddd-and-gdb)) |
| `--gdb[=PORT]` | let gdb connect: `target extended-remote localhost:1234` |
| `--regs` | print all registers and flags when the program ends |
| `-32`, `-64` | force 32- or 64-bit mode |
| `-f elf32`, `-f elf64` | same as `-32`/`-64` (nasm style). `-o`, `-g`, `-F`, `-l` are accepted and ignored; other nasm options (`-D`, `-I`, `-w`...) are passed to nasm |
| `--max-steps N` | stop after N instructions (default 50,000,000; `0` = no limit) — catches infinite loops |
| `--th` | Thai messages (or set `NASMX86_LANG=th`) |
| `--version`, `-h` | version / help |

Everything after the `.asm` file is passed to your program, except nasmx86's
own long options; put `--` before arguments that look like them.

### `--trace`

```
$ nasmx86 --trace examples/function32.asm
[1]    line 10: push 7                        esp=0xfffdc10c  [0xfffdc10c]=0x00000007
[2]    line 11: push 5                        esp=0xfffdc108  [0xfffdc108]=0x00000005
[3]    line 12: call add_mul                  esp=0xfffdc104  [0xfffdc104]=0x08049009
[4]    line 21: push ebp                      esp=0xfffdc100  [0xfffdc100]=0x00000000
[5]    line 22: mov ebp, esp                  ebp=0xfffdc100
[6]    line 23: mov eax, [ebp + 8]            eax=0x00000005 (5)
[7]    line 24: add eax, [ebp + 12]           eax=0x0000000c (12)  PF=1
[8]    line 25: shl eax, 1                    eax=0x00000018 (24)
...
[14]   line 17: int 0x80                      → exit(24)
```

Trace output goes to stderr, so `nasmx86 --trace prog.asm 2> trace.txt` keeps
the program's own output separate.

### `--step` debugger

```
$ nasmx86 --step examples/factorial32.asm
nasmx86 debugger: Enter = step one instruction, h = help, q = quit
→ line 19: mov eax, 4
(nasmx86) b factorial
breakpoint #1 at line 63: push ebp
(nasmx86) c
```

| Command | |
|---|---|
| Enter, `s [n]` | step one (or n) instructions, showing what changed |
| `n` | next: step over a `call` (runs the whole function) |
| `finish` | run until the current function returns, and show its `eax`/`rax` |
| `c` | continue until a breakpoint, a watchpoint or the end (Ctrl-C interrupts) |
| `b <line\|label>` | set a breakpoint (`b` alone lists them; `b io.asm:12` in another file) |
| `watch <reg\|addr>` | stop when it changes: `watch ecx`, `watch counter`, `watch byte [buf+3]` |
| `d [n\|line]` | delete breakpoint/watchpoint n (`d` alone deletes all) |
| `r` | registers and flags |
| `p <expr>` | print a value: `p eax`, `p [count]`, `p dword [arr+ecx*4]`, `p/s msg` (string) |
| `set <what> <value>` | change a register, flag or memory: `set eax 5`, `set zf 1`, `set dword [count] 0` |
| `m <addr\|label\|reg> [len]` | hex dump of memory: `m msg`, `m esp 32`, `m buf+4 8` |
| `x [n]` / `stack` | top n stack entries, with return addresses mapped back to lines |
| `bt` | the call stack: which function called which (handy for recursion) |
| `l` | list the source around the current line |
| `q` | quit |

The program's own input still works while debugging (commands are read from
the terminal, even when the program's stdin is redirected).

#### Coming from gdb?

On Apple Silicon Macs gdb cannot debug x86 programs, so `--step` also
understands the gdb commands lab sheets use — follow them as written:

| gdb (lab sheet) | works in `nasmx86 --step` |
|---|---|
| `break _start`, `break *0x8049005`, `tbreak`, `run`, `start`, `kill` | ✅ (`run` again restarts the program) |
| `stepi`, `nexti`, `stepi 5`, `continue`, `finish`, `until 20`, `advance loop` | ✅ |
| `info registers`, `info registers eax`, `i r` | ✅ same layout, incl. `eflags [ PF ZF IF ]` |
| `print $eax`, `p/x $eax`, `p/t $al` (binary), `p/c`, `p (int)count`, `p *(int*)&arr` | ✅ |
| `x/4xw $esp`, `x/s &msg`, `x/16xb msg`, `x/5i $pc` | ✅ |
| `display/i $pc`, `display $eax`, `undisplay`, `info display` | ✅ shown after every step |
| `disassemble`, `backtrace`, `info breakpoints`, `enable`/`disable`, `delete` | ✅ |
| `set $eax = 5`, `set var (int)count = 0`, `set {char}&msg = 'H'` | ✅ |
| `set disassembly-flavor intel` | accepted (output is always Intel/NASM syntax) |
| `layout regs` / `layout asm` (TUI) | ❌ use `display/i $pc` and `display $eax` instead |

The short commands above (`s`, `n`, `c`, `r` for registers, `m`, `x` for the
stack) keep working. Note `r` shows registers here; gdb's `r` (run) is `run`.

Example: why does my recursion return the wrong value?

```
(nasmx86) b 23              ← the base case of factorial
(nasmx86) c
(nasmx86) bt
#0   fact    line 23: mov eax, 1
#1   fact    line 18: call fact
#2   fact    line 18: call fact
#3   _start  line 8: call fact
(nasmx86) finish
returned from fact: eax = 0x00000001 (1)
(nasmx86) watch eax         ← stop each time eax changes
```

### DDD and gdb

`nasmx86 --ddd prog.asm` opens your program in **DDD**, the graphical
debugger many assembly courses use — the real DDD with the real gdb
underneath, so lab sheets work as written (`break`, `run`, `step`, `next`,
`display`, `x/4dw &arr`, `info registers`...). nasmx86 plays the part of
the Linux machine: it runs the program, and gdb controls it over gdb's
remote protocol, as it does with `gdbserver`. Your program's own input and
output stay in the terminal where you typed the command.

One-time setup on a Mac (Homebrew):

```bash
brew install ddd x86_64-elf-gdb
brew install --cask xquartz       # the X11 window system DDD draws with
```

`nasmx86 --gdb prog.asm` does the same for gdb on its own (or any front-end
that speaks gdb's remote protocol): it waits on `localhost:1234` and prints
the gdb command to run in a second terminal. `--gdb=4444` picks the port.

As on Linux, gdb knows no types for assembly variables: use `x/dw &count`,
`p (int)count` or `display (int)count`. `break printf` and `finish` work for
the [C functions](#calling-c-functions-printf-scanf-) too, and a crash is
reported to gdb as the Linux signal (SIGSEGV...) together with nasmx86's
explanation.

### Errors explained

| Situation | What nasmx86 says |
|---|---|
| reading/writing a bad address | `line 12: mov eax, [ebx]  →  cannot read memory at 0x00000000 (ebx = 0x00000000)` + hint |
| writing to `section .text` | `.text is read-only. Put variables that change in section .data or section .bss` |
| `div` with `edx` not cleared / divide by zero | the dividend, divisor and the fix (`xor edx, edx` / `cdq`) |
| no exit system call | `the program ran past the end of its code` + the exact exit code to add |
| `ret` in `_start`, unbalanced `push`/`ret` | explains the stack problem |
| infinite recursion | `stack overflow` |
| infinite loop | stops after `--max-steps` and says so |
| `scanf` given a value instead of an address, misaligned stack at `call printf`, `free` twice | explains the mistake at the `call` line |
| `int 21h` (DOS), x87/SSE | says why it cannot work on Linux / in nasmx86 |

The exit status mimics Linux: 139 for a segmentation fault, 136 for a divide
error, etc.

---

## What is supported

**Instructions** (8/16/32/64-bit forms, all flags CF ZF SF OF PF AF DF):
`mov movzx movsx movsxd lea xchg xadd cmpxchg push pop pushad popad pushf popf
add adc sub sbb inc dec neg cmp mul imul div idiv cbw cwde cdqe cwd cdq cqo
and or xor not test shl sal shr sar rol ror rcl rcr shld shrd
jmp jcc jcxz jecxz jrcxz loop loope loopne call ret enter leave
setcc cmovcc bt bts btr btc bsf bsr bswap
movs stos lods cmps scas (with rep/repe/repne) cld std clc stc cmc
xlat lahf sahf nop hlt int 0x80 syscall`, and the BCD adjustments
`daa das aaa aas aam aad` (32-bit; like real CPUs, not available in 64-bit mode).

**Addressing modes:** `[disp]`, `[base]`, `[base+disp]`,
`[base+index*scale+disp]`, RIP-relative (`default rel`, `[rel x]`).

**System calls** (Linux numbers, both ABIs): `read write open openat creat close
lseek unlink stat fstat mkdir rmdir rename brk getpid time gettimeofday
clock_gettime nanosleep exit exit_group`. Others return `-ENOSYS` with a
warning.

**Program start** is like Linux: `argc/argv/envp` on the stack, registers zero,
`.text` at `0x08049000` (32-bit) / `0x401000` (64-bit). If your entry point is
`main` instead of `_start`, returning from `main` exits with `eax` as the status
(like a C program).

## Limitations

- The built-in C library has the functions listed [above](#calling-c-functions-printf-scanf-)
  (no `%f` in 64-bit mode, no locales or wide characters). C source files
  (calling your assembly from `main.c`) are not compiled — nasmx86 runs assembly.
- NASM syntax for Linux only: not MASM/TASM, GAS (AT&T), DOS (`int 21h`) or Windows programs.
- No floating point (x87), SSE/AVX, or privileged instructions.
- Single-threaded; no signals, `mmap`, `fork`, sockets.
- `hlt` ends the program (real Linux would crash) — nasmx86 prints a note.
- Flags that the CPU manual calls *undefined* (e.g. SF/ZF after `mul`) may differ
  from a particular real CPU — real CPUs differ from each other there too.
- Stack addresses differ from a real run (Linux randomizes them anyway).

## FAQ

**Is it accurate?** Every instruction group is covered by
[differential tests](testdata/diff): the same programs run natively on x86-64
Linux in CI and through nasmx86, and their output (register and flag dumps
after each operation) must be byte-for-byte identical.

**Why not Rosetta?** Rosetta 2 only runs macOS x86-64 apps; it cannot run
Linux programs or any 32-bit code.

**`nasm: command not found`?** Install nasm (`brew install nasm`,
`sudo apt install nasm`, `winget install NASM.NASM`). nasmx86 prints the
right command for your system.

**My program waits forever.** It is probably reading stdin — type the input
(Ctrl-D ends input), or pipe it: `echo 5 | nasmx86 prog.asm`.

**Can I use gdb or DDD?** Yes: `nasmx86 --gdb` or `--ddd` (see [DDD and gdb](#ddd-and-gdb)). `--step` is a built-in alternative that needs nothing else installed.

---

## Development

```bash
go build ./cmd/nasmx86
go test ./...                     # unit + emulator-only differential tests
scripts/rosetta-diff.sh           # Apple Silicon: check 64-bit tests against real x86-64 (Rosetta 2)
```

CI (`.github/workflows/ci.yml`) runs the unit tests on Linux, macOS and
Windows, and the native-vs-nasmx86 differential tests on x86-64 Linux.
Pushing a tag `v*` builds and publishes release binaries and updates the
Homebrew tap (`.goreleaser.yaml`, `.github/workflows/release.yml`).

Layout: `cmd/nasmx86` (CLI) · `internal/nasm` (running nasm, bit detection,
listing → source lines) · `internal/loader` (mini linker) · `internal/emu`
(CPU, memory, Linux system calls) · `internal/runner` (build/run/friendly
errors) · `internal/debugger` (`--trace`, `--step`) · `internal/i18n`
(English/Thai messages) · `testdata/diff` (differential tests).
