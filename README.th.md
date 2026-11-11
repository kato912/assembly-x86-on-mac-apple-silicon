# nasmx86

**รันโปรแกรมแอสเซมบลี x86 ได้ในคำสั่งเดียว — บน Mac (Apple Silicon หรือ Intel), Windows และ Linux โดยไม่ต้องใช้ VM, Docker หรือ Rosetta**

ภาษาไทย · [English](README.md)

```
$ nasmx86 hello.asm
Hello, World!
```

`nasmx86` จะแปลงโปรแกรม NASM ด้วย `nasm` ตัวจริง ลิงก์ด้วยลิงเกอร์ขนาดเล็กที่มีในตัว
แล้วรันบนตัวจำลอง CPU x86 ที่รองรับ system call ของ Linux ใช้ได้ทั้งโปรแกรม **32 บิต**
(`int 0x80`) และ **64 บิต** (`syscall`) ซึ่งเป็นแบบที่วิชาแอสเซมบลี x86 เบื้องต้นใช้กัน

เมื่อโปรแกรมผิดพลาด จะบอกว่าผิดที่ *บรรทัดไหน* และ *เพราะอะไร* แทนที่จะขึ้นแค่
`Segmentation fault`:

```
$ nasmx86 --th avg.asm
บรรทัด 14: div ecx  →  ผลหารล้น: edx:eax = 0x00000009:0x00000064 หารด้วย 4 ได้ผลหารใหญ่เกินกว่าจะเก็บใน eax
    คำแนะนำ: div จะหารค่าจากรีจิสเตอร์คู่ (ขนาดสองเท่า) ต้องเคลียร์ครึ่งบนก่อน (xor edx, edx)
```

---

## การติดตั้ง

### macOS (และ Linux) ด้วย Homebrew — แนะนำ

```bash
brew install kato912/tap/nasmx86
```

คำสั่งนี้ติดตั้ง `nasm` ให้ด้วย ไม่ต้องมีบัญชี Homebrew หรือล็อกอิน GitHub

### macOS / Linux แบบไม่ใช้ Homebrew

```bash
curl -fsSL https://raw.githubusercontent.com/kato912/assembly-x86-on-mac-apple-silicon/main/install.sh | sh
```

สคริปต์จะดาวน์โหลดไฟล์ที่ตรงกับระบบปฏิบัติการ/CPU ของเครื่องไปไว้ที่ `/usr/local/bin`
(หรือ `~/.local/bin`) ตรวจ checksum เพิ่มเข้า `PATH` ให้ และติดตั้ง `nasm` ให้ถ้าทำได้
(`brew`, `apt`, `dnf`, `pacman` ...) ตัวอย่างโปรแกรมจะอยู่ที่ `~/.local/share/nasmx86/examples`

### Windows

1. ติดตั้ง NASM: `winget install NASM.NASM`
   (หรือดาวน์โหลดตัวติดตั้งจาก <https://www.nasm.us/pub/nasm/releasebuilds/>)
2. ดาวน์โหลด `nasmx86_<เวอร์ชัน>_windows_amd64.zip` จาก
   [หน้า Releases](https://github.com/kato912/assembly-x86-on-mac-apple-silicon/releases)
   แตกไฟล์ แล้วนำ `nasmx86.exe` ไปไว้ในโฟลเดอร์ที่อยู่ใน `PATH`
   (ดู `INSTALL-WINDOWS.txt` ในไฟล์ zip)
3. เปิดเทอร์มินัล **ใหม่** แล้วลอง `nasmx86 examples\hello32.asm`

nasmx86 หา `nasm` ใน `C:\Program Files\NASM` ให้เองแม้จะไม่ได้อยู่ใน `PATH`

### คอมไพล์จากซอร์สโค้ด

```bash
go install github.com/kato912/assembly-x86-on-mac-apple-silicon/cmd/nasmx86@latest
```

---

## เริ่มต้นใช้งาน

`hello.asm` (32 บิต):

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
nasmx86 hello.asm                 # แปลง + ลิงก์ + รัน
echo 5 | nasmx86 factorial32.asm  # รับ input ทาง stdin ได้เหมือนบน Linux
nasmx86 prog.asm arg1 arg2        # อาร์กิวเมนต์ส่งให้โปรแกรม (argc/argv อยู่บนสแตก)
nasmx86 prog.asm; echo $?         # exit code คือ exit code ของโปรแกรม
```

ดูตัวอย่างเพิ่มเติมได้ที่ [`examples/`](examples): รับ input, ลูป, อาร์เรย์,
ฟังก์ชันที่ส่งพารามิเตอร์ทางสแตก, การเรียกซ้ำ (factorial) — มีทั้ง 32 และ 64 บิต

### 32 หรือ 64 บิต?

nasmx86 เลือกให้อัตโนมัติ (ดูในไฟล์ที่ `%include` ด้วย):

1. ถ้ามี `bits 32` / `bits 64` จะใช้ตามนั้น
2. ใช้รีจิสเตอร์ 64 บิต (`rax`, `r8`, ...) หรือ `default rel` → 64 บิต
3. ใช้ `syscall` → 64 บิต, ใช้ `int 0x80` → 32 บิต
4. ถ้าไม่มีอะไรบอก → 64 บิต

บังคับเองได้ด้วย `-32` / `-64` หรือใช้ตัวเลือกแบบ nasm ตามเอกสารในห้องเรียน:
`nasmx86 -f elf32 hello.asm`

---

## ตัวเลือก

| ตัวเลือก | ความหมาย |
|---|---|
| `--trace` | แสดงทุกคำสั่งที่รัน พร้อมบรรทัดในซอร์สโค้ด และสิ่งที่เปลี่ยน (รีจิสเตอร์ แฟล็ก หน่วยความจำ system call) — ออกทาง stderr |
| `--step` | ดีบักเกอร์แบบโต้ตอบ (ดูด้านล่าง) |
| `--regs` | แสดงรีจิสเตอร์และแฟล็กทั้งหมดตอนโปรแกรมจบ |
| `-32`, `-64` | บังคับโหมด 32 หรือ 64 บิต |
| `-f elf32`, `-f elf64` | เหมือน `-32`/`-64` (แบบ nasm) ส่วน `-o`, `-g`, `-F`, `-l` ใส่ได้แต่จะถูกข้าม ตัวเลือกอื่นของ nasm (`-D`, `-I`, `-w` ...) ส่งต่อให้ nasm |
| `--max-steps N` | หยุดเมื่อรันครบ N คำสั่ง (ค่าเริ่มต้น 50,000,000; `0` = ไม่จำกัด) — กันลูปไม่รู้จบ |
| `--th` | ข้อความภาษาไทย (หรือตั้งค่า `NASMX86_LANG=th`) |
| `--version`, `-h` | เวอร์ชัน / วิธีใช้ |

ทุกอย่างที่อยู่หลังชื่อไฟล์ `.asm` จะส่งให้โปรแกรม ยกเว้นตัวเลือกยาวของ nasmx86 เอง
ถ้าต้องการส่งอาร์กิวเมนต์ที่หน้าตาเหมือนตัวเลือกเหล่านั้น ให้ใส่ `--` นำหน้า

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

ผลของ trace ออกทาง stderr จึงแยกจาก output ของโปรแกรมได้ เช่น
`nasmx86 --trace prog.asm 2> trace.txt`

### ดีบักเกอร์ `--step`

```
$ nasmx86 --step --th examples/factorial32.asm
ดีบักเกอร์ nasmx86: กด Enter = รันทีละคำสั่ง, h = ดูวิธีใช้, q = ออก
→ บรรทัด 19: mov eax, 4
(nasmx86) b factorial
ตั้ง breakpoint #1 ที่ บรรทัด 63: push ebp
(nasmx86) c
```

| คำสั่ง | |
|---|---|
| Enter, `s [n]` | รันทีละ 1 (หรือ n) คำสั่ง พร้อมแสดงสิ่งที่เปลี่ยน |
| `n` | next: ถ้าเป็น `call` จะรันทั้งฟังก์ชันจนกลับมา |
| `c` | รันต่อจนเจอ breakpoint หรือโปรแกรมจบ (กด Ctrl-C เพื่อหยุด) |
| `b <บรรทัด\|label>` | ตั้ง breakpoint (`b` เฉย ๆ = ดูรายการ) |
| `d [n\|บรรทัด]` | ลบ breakpoint หมายเลข n (`d` เฉย ๆ = ลบทั้งหมด) |
| `r` | ดูรีจิสเตอร์และแฟล็ก |
| `m <แอดเดรส\|label\|รีจิสเตอร์> [len]` | ดูหน่วยความจำ: `m msg`, `m esp 32`, `m buf+4 8` |
| `x [n]` / `stack` | ดูสแตก n ช่องบนสุด พร้อมบอกว่า return address จะกลับไปบรรทัดไหน |
| `l` | แสดงซอร์สโค้ดรอบบรรทัดปัจจุบัน |
| `q` | ออก |

ระหว่างดีบัก โปรแกรมยังรับ input ได้ตามปกติ (คำสั่งดีบักอ่านจากเทอร์มินัลโดยตรง
แม้ stdin ของโปรแกรมจะถูก redirect)

### ข้อผิดพลาดที่อธิบายให้

| สถานการณ์ | nasmx86 บอกว่า |
|---|---|
| อ่าน/เขียนแอดเดรสที่ไม่ถูกต้อง | `บรรทัด 12: mov eax, [ebx]  →  อ่านหน่วยความจำที่ 0x00000000 ไม่ได้ (ebx = 0x00000000)` + คำแนะนำ |
| เขียนลงใน `section .text` | `.text เป็นส่วนที่อ่านได้อย่างเดียว ให้ย้ายตัวแปรไปไว้ใน section .data หรือ .bss` |
| `div` โดยไม่เคลียร์ `edx` / หารด้วยศูนย์ | แสดงตัวตั้ง ตัวหาร และวิธีแก้ (`xor edx, edx` / `cdq`) |
| ไม่ได้เรียก system call exit | `โปรแกรมรันเลยจุดสิ้นสุดของโค้ด` + โค้ด exit ที่ต้องเพิ่ม |
| ใช้ `ret` ใน `_start`, `push`/`ret` ไม่สมดุล | อธิบายปัญหาของสแตก |
| เรียกซ้ำไม่รู้จบ | `สแตกล้น (stack overflow)` |
| ลูปไม่รู้จบ | หยุดเมื่อครบ `--max-steps` และบอกสาเหตุ |
| `int 21h` (DOS), `printf` (ไลบรารี C), x87/SSE | อธิบายว่าทำไมใช้บน Linux / ใน nasmx86 ไม่ได้ |

exit status เลียนแบบ Linux: 139 เมื่อ segmentation fault, 136 เมื่อหารผิดพลาด เป็นต้น

---

## สิ่งที่รองรับ

**คำสั่ง** (ทั้งแบบ 8/16/32/64 บิต และแฟล็ก CF ZF SF OF PF AF DF ครบ):
`mov movzx movsx movsxd lea xchg xadd cmpxchg push pop pushad popad pushf popf
add adc sub sbb inc dec neg cmp mul imul div idiv cbw cwde cdqe cwd cdq cqo
and or xor not test shl sal shr sar rol ror rcl rcr shld shrd
jmp jcc jcxz jecxz jrcxz loop loope loopne call ret enter leave
setcc cmovcc bt bts btr btc bsf bsr bswap
movs stos lods cmps scas (พร้อม rep/repe/repne) cld std clc stc cmc
xlat lahf sahf nop hlt int 0x80 syscall`

**Addressing mode:** `[disp]`, `[base]`, `[base+disp]`,
`[base+index*scale+disp]`, RIP-relative (`default rel`, `[rel x]`)

**System call** (หมายเลขของ Linux ทั้งสอง ABI): `read write open openat creat close
lseek unlink brk getpid time exit exit_group` ตัวอื่นจะคืนค่า `-ENOSYS` พร้อมคำเตือน

**การเริ่มโปรแกรม** เหมือน Linux: มี `argc/argv/envp` บนสแตก รีจิสเตอร์เป็นศูนย์
`.text` เริ่มที่ `0x08049000` (32 บิต) / `0x401000` (64 บิต) ถ้า entry point เป็น `main`
แทน `_start` การ `ret` ออกจาก `main` จะจบโปรแกรมโดยใช้ `eax` เป็น exit code (เหมือนโปรแกรม C)

## ข้อจำกัด

- ไม่มีไลบรารี C: ใช้ `extern printf` ฯลฯ ไม่ได้ — ให้ใช้ system call
- ไม่รองรับทศนิยม (x87), SSE/AVX, โปรแกรม DOS 16 บิต และคำสั่งสิทธิ์พิเศษ
- ทำงานเธรดเดียว ไม่มี signal, `mmap`, `fork`
- `hlt` จะจบโปรแกรม (บน Linux จริงจะ crash) — nasmx86 แสดงหมายเหตุให้
- แฟล็กที่คู่มือ CPU ระบุว่า *undefined* (เช่น SF/ZF หลัง `mul`) อาจไม่ตรงกับ CPU บางรุ่น
  — ซึ่ง CPU จริงแต่ละรุ่นก็ให้ค่าต่างกันอยู่แล้ว
- แอดเดรสของสแตกไม่ตรงกับการรันจริง (Linux สุ่มแอดเดรสอยู่แล้ว)

## คำถามที่พบบ่อย

**แม่นยำแค่ไหน?** ทุกกลุ่มคำสั่งมี [differential test](testdata/diff): โปรแกรมเดียวกันถูกรัน
บน x86-64 Linux จริงใน CI และรันผ่าน nasmx86 แล้วผลลัพธ์ (ค่ารีจิสเตอร์และแฟล็กหลังทุกคำสั่ง)
ต้องตรงกันทุกไบต์

**ทำไมไม่ใช้ Rosetta?** Rosetta 2 รันได้แค่แอป macOS แบบ x86-64 ไม่สามารถรันโปรแกรม Linux
หรือโค้ด 32 บิตได้

**ขึ้นว่า `nasm: command not found`?** ติดตั้ง nasm ก่อน (`brew install nasm`,
`sudo apt install nasm`, `winget install NASM.NASM`) nasmx86 จะบอกคำสั่งที่ถูกต้องสำหรับเครื่องนั้น ๆ

**โปรแกรมค้างไม่จบ** น่าจะกำลังรอ input จาก stdin — พิมพ์ข้อมูลเข้าไป (Ctrl-D เพื่อจบ input)
หรือส่งผ่าน pipe: `echo 5 | nasmx86 prog.asm`

**ใช้ gdb ได้ไหม?** ให้ใช้ `nasmx86 --step` หรือ `--trace` แทน

---

## สำหรับนักพัฒนา

```bash
go build ./cmd/nasmx86
go test ./...                     # unit test + differential test (เฉพาะฝั่ง nasmx86)
scripts/rosetta-diff.sh           # Apple Silicon: เทียบเทสต์ 64 บิตกับ x86-64 จริงผ่าน Rosetta 2
```

CI (`.github/workflows/ci.yml`) รัน unit test บน Linux, macOS และ Windows และรัน differential test
(รันจริงเทียบกับ nasmx86) บน x86-64 Linux เมื่อ push tag `v*` จะ build และเผยแพร่ไฟล์บน
Releases พร้อมอัปเดต Homebrew tap ให้อัตโนมัติ (`.goreleaser.yaml`, `.github/workflows/release.yml`)
