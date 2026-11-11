package i18n

var catalog = map[string]msg{
	// command line
	"usage": {en: `usage: nasmx86 [options] file.asm [more.asm ...] [program arguments...]

Assemble, link and run a 32- or 64-bit Linux NASM program in one step.
Several .asm (or nasm .o) files are linked together: the first one is the
main program; use global/extern to share labels between them.

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
  nasmx86 main.asm io.asm
  echo 5 | nasmx86 fact.asm
`, th: `วิธีใช้: nasmx86 [ตัวเลือก] ไฟล์.asm [ไฟล์อื่น.asm ...] [อาร์กิวเมนต์ของโปรแกรม...]

แปลง (assemble) ลิงก์ และรันโปรแกรม NASM ของ Linux แบบ 32 หรือ 64 บิต ในคำสั่งเดียว
ใส่ไฟล์ .asm (หรือ .o ของ nasm) หลายไฟล์ได้ ไฟล์แรกคือโปรแกรมหลัก
ใช้ global/extern เพื่อเรียก label ข้ามไฟล์

ตัวเลือก:
  --trace            แสดงทุกคำสั่งที่รัน พร้อมบรรทัดในซอร์สโค้ด
                     และรีจิสเตอร์/แฟล็กที่เปลี่ยน (ออกทาง stderr)
  --step             ดีบักเกอร์แบบโต้ตอบ (พิมพ์ h ที่พรอมต์เพื่อดูคำสั่ง)
  -32, -64           บังคับโหมด 32 หรือ 64 บิต (ปกติจะตรวจจับให้อัตโนมัติ)
  -f elf32|elf64     เหมือน -32/-64 (แบบ nasm; -o และ -g จะถูกข้าม
                     ตัวเลือกอื่นของ nasm เช่น -D หรือ -I ส่งต่อให้ nasm)
  --max-steps N      หยุดเมื่อรันครบ N คำสั่ง (ค่าเริ่มต้น 50000000, 0 = ไม่จำกัด)
  --regs             แสดงค่ารีจิสเตอร์ตอนโปรแกรมจบ
  --th               แสดงข้อความภาษาไทย (หรือตั้ง NASMX86_LANG=th)
  --version          แสดงเวอร์ชัน
  -h, --help         แสดงวิธีใช้นี้

ตัวอย่าง:
  nasmx86 hello.asm
  nasmx86 --trace loop.asm
  nasmx86 -f elf32 hello.asm
  nasmx86 main.asm io.asm
  echo 5 | nasmx86 fact.asm
`},
	"err.prefix":        {en: "nasmx86: "},
	"err.noFile":        {en: "no .asm file given", th: "ไม่ได้ระบุไฟล์ .asm"},
	"err.missingValue":  {en: "option %s needs a value", th: "ตัวเลือก %s ต้องมีค่าตามหลัง"},
	"err.badSteps":      {en: "--max-steps needs a number, got %q", th: "--max-steps ต้องเป็นตัวเลข แต่ได้ %q"},
	"err.unknownFormat": {en: "unsupported output format %q: nasmx86 runs Linux programs, use -f elf32 or -f elf64", th: "ไม่รองรับรูปแบบ %q: nasmx86 รันโปรแกรมของ Linux ให้ใช้ -f elf32 หรือ -f elf64"},
	"err.readSource":    {en: "cannot read %s: %v", th: "อ่านไฟล์ %s ไม่ได้: %v"},
	"err.nasmMissing":   {en: "nasm (the assembler) is not installed.", th: "ยังไม่ได้ติดตั้ง nasm (ตัวแปลงภาษาแอสเซมบลี)"},
	"install.darwin": {en: "Install it with Homebrew:\n    brew install nasm\n(no Homebrew yet? see https://brew.sh)",
		th: "ติดตั้งด้วย Homebrew:\n    brew install nasm\n(ยังไม่มี Homebrew? ดูที่ https://brew.sh)"},
	"install.windows": {en: "Install it with winget:\n    winget install NASM.NASM\nor download the installer from https://www.nasm.us/pub/nasm/releasebuilds/\n(then open a new terminal; nasmx86 also looks in C:\\Program Files\\NASM)",
		th: "ติดตั้งด้วย winget:\n    winget install NASM.NASM\nหรือดาวน์โหลดตัวติดตั้งจาก https://www.nasm.us/pub/nasm/releasebuilds/\n(แล้วเปิดเทอร์มินัลใหม่ nasmx86 จะหาใน C:\\Program Files\\NASM ให้ด้วย)"},
	"install.linux": {en: "Install it with your package manager, e.g.\n    sudo apt install nasm      (Debian/Ubuntu)\n    sudo dnf install nasm      (Fedora)\n    sudo pacman -S nasm        (Arch)",
		th: "ติดตั้งด้วยตัวจัดการแพ็กเกจ เช่น\n    sudo apt install nasm      (Debian/Ubuntu)\n    sudo dnf install nasm      (Fedora)\n    sudo pacman -S nasm        (Arch)"},
	"err.bits16": {en: "this looks like 16-bit (DOS) code (%s). nasmx86 runs 32-bit and 64-bit Linux programs only.", th: "โค้ดนี้ดูเหมือนเป็นโค้ด 16 บิต (DOS) (%s) nasmx86 รันได้เฉพาะโปรแกรม Linux แบบ 32 และ 64 บิต"},
	"err.link":   {en: "link error: %v", th: "ลิงก์ไม่สำเร็จ: %v"},
	"err.undefined": {en: `your program uses %s, which is not defined in this file.
nasmx86 runs pure assembly programs without the C library, so C functions
like printf/scanf/puts are not available. Use Linux system calls instead,
e.g. write (eax=4, int 0x80 in 32-bit / rax=1, syscall in 64-bit).`, th: `โปรแกรมเรียกใช้ %s ซึ่งไม่ได้นิยามไว้ในไฟล์นี้
nasmx86 รันโปรแกรมแอสเซมบลีล้วน ๆ โดยไม่มีไลบรารีภาษา C จึงใช้ฟังก์ชัน C
อย่าง printf/scanf/puts ไม่ได้ ให้ใช้ system call ของ Linux แทน
เช่น write (eax=4 แล้ว int 0x80 ในโหมด 32 บิต / rax=1 แล้ว syscall ในโหมด 64 บิต)`},
	"err.internal":        {en: "internal error: %v", th: "ข้อผิดพลาดภายใน: %v"},
	"err.duplicate":       {en: "the label %s is defined (global) in more than one file: %s", th: "label %s ถูกประกาศเป็น global ในหลายไฟล์: %s"},
	"hint.undefinedMulti": {en: "if %[1]s is in one of your other .asm files, declare it there with:  global %[1]s", th: "ถ้า %[1]s อยู่ในไฟล์ .asm อื่นของคุณ ต้องประกาศในไฟล์นั้นด้วย:  global %[1]s"},

	// runtime warnings
	"warn.efault": {en: "warning: %s: the buffer at %#x is not valid memory, so the call failed (returned -EFAULT)", th: "คำเตือน: %s: บัฟเฟอร์ที่ %#x ไม่ใช่หน่วยความจำที่ใช้ได้ การเรียกจึงล้มเหลว (คืนค่า -EFAULT)"},
	"warn.enosys": {en: "warning: system call %d is not supported by nasmx86 (returned -ENOSYS)", th: "คำเตือน: nasmx86 ยังไม่รองรับ system call หมายเลข %d (คืนค่า -ENOSYS)"},
	"warn.halt":   {en: "note: the program stopped at hlt. (On real Linux, hlt in a user program crashes with a segmentation fault — use the exit system call.)", th: "หมายเหตุ: โปรแกรมหยุดที่คำสั่ง hlt (บน Linux จริง hlt ในโปรแกรมผู้ใช้จะทำให้เกิด segmentation fault — ควรใช้ system call exit)"},
	"warn.badfd":  {en: "warning: %s: file descriptor %d is not open (returned -EBADF)", th: "คำเตือน: %s: file descriptor %d ไม่ได้เปิดอยู่ (คืนค่า -EBADF)"},

	// source locations
	"loc.line":  {en: "line %d: %s", th: "บรรทัด %d: %s"},
	"loc.macro": {en: "(in macro: %s)", th: "(ในแมโคร: %s)"},
	"loc.stub":  {en: "(after returning from main)", th: "(หลังจาก return ออกจาก main)"},

	// runtime errors
	"hint.prefix":        {en: "hint: ", th: "คำแนะนำ: "},
	"err.read":           {en: "cannot read memory at %s", th: "อ่านหน่วยความจำที่ %s ไม่ได้"},
	"err.write":          {en: "cannot write memory at %s", th: "เขียนหน่วยความจำที่ %s ไม่ได้"},
	"hint.null":          {en: "the address is (almost) 0, so the register was probably never set to a valid address (e.g. mov ebx, msg)", th: "แอดเดรสเป็น 0 (หรือเกือบ 0) แปลว่ารีจิสเตอร์น่าจะยังไม่ได้ใส่แอดเดรสที่ถูกต้อง (เช่น mov ebx, msg)"},
	"hint.readonly":      {en: "%s is read-only. Put variables that change in section .data or section .bss", th: "%s เป็นส่วนที่อ่านได้อย่างเดียว ให้ย้ายตัวแปรที่ต้องแก้ค่าไปไว้ใน section .data หรือ section .bss"},
	"hint.smallAddr":     {en: "that address is not part of your program. Did you mean the value itself? [x] reads the memory AT address x", th: "แอดเดรสนี้ไม่ใช่ของโปรแกรม ตั้งใจจะใช้ค่าตัวเลขตรง ๆ หรือเปล่า? [x] คือการอ่านหน่วยความจำ ณ แอดเดรส x"},
	"hint.unmapped":      {en: "that address is not part of your program's memory (check the pointer and the index/offset)", th: "แอดเดรสนี้อยู่นอกหน่วยความจำของโปรแกรม (ตรวจสอบพอยน์เตอร์และค่า index/offset)"},
	"err.stackOverflow":  {en: "stack overflow", th: "สแตกล้น (stack overflow)"},
	"hint.stackOverflow": {en: "the stack ran out of space: is there a recursion without a base case, or a loop that pushes without popping?", th: "สแตกเต็ม: มีการเรียกซ้ำ (recursion) ที่ไม่มีเงื่อนไขหยุด หรือลูปที่ push โดยไม่ pop หรือเปล่า?"},
	"err.exec":           {en: "jumped to %s, which is not part of the code", th: "กระโดดไปที่ %s ซึ่งไม่ใช่ส่วนของโค้ด"},
	"hint.exec":          {en: "check the target of the jmp/call (a register or memory value used as an address)", th: "ตรวจสอบเป้าหมายของ jmp/call (ค่าในรีจิสเตอร์หรือหน่วยความจำที่ใช้เป็นแอดเดรส)"},
	"hint.execZero":      {en: "a call or jmp through a register that was never set (it is 0)", th: "call หรือ jmp ผ่านรีจิสเตอร์ที่ยังไม่ได้กำหนดค่า (ค่าเป็น 0)"},
	"err.runoff":         {en: "the program ran past the end of its code", th: "โปรแกรมรันเลยจุดสิ้นสุดของโค้ด"},
	"hint.exit32":        {en: "did you forget the exit system call?  mov eax, 1 / mov ebx, 0 / int 0x80", th: "ลืมเรียก system call exit หรือเปล่า?  mov eax, 1 / mov ebx, 0 / int 0x80"},
	"hint.exit64":        {en: "did you forget the exit system call?  mov rax, 60 / mov rdi, 0 / syscall", th: "ลืมเรียก system call exit หรือเปล่า?  mov rax, 60 / mov rdi, 0 / syscall"},
	"err.retStart":       {en: "ret in _start: there is no function to return to", th: "ใช้ ret ใน _start: ไม่มีฟังก์ชันให้ return กลับไป"},
	"err.badRet":         {en: "ret jumped to %s, which is not part of the code", th: "ret กระโดดไปที่ %s ซึ่งไม่ใช่ส่วนของโค้ด"},
	"hint.badRet":        {en: "the value on top of the stack was not a return address: make sure every push has a matching pop before ret", th: "ค่าบนสุดของสแตกไม่ใช่ return address: ตรวจสอบว่าทุก push มี pop คู่กันก่อนถึง ret"},
	"err.divZero":        {en: "division by zero (the divisor %s is 0)", th: "หารด้วยศูนย์ (ตัวหาร %s มีค่าเป็น 0)"},
	"err.divOverflow":    {en: "division overflow: %s divided by %d does not fit in %s", th: "ผลหารล้น: %s หารด้วย %d ได้ผลหารใหญ่เกินกว่าจะเก็บใน %s"},
	"hint.divUnsigned":   {en: "div divides the double-width register pair; clear the high half first (%s)", th: "div จะหารค่าจากรีจิสเตอร์คู่ (ขนาดสองเท่า) ต้องเคลียร์ครึ่งบนก่อน (%s)"},
	"hint.divSigned":     {en: "idiv divides the double-width register pair; sign-extend first (%s)", th: "idiv จะหารค่าจากรีจิสเตอร์คู่ (ขนาดสองเท่า) ต้องขยายเครื่องหมายก่อน (%s)"},
	"err.unsupported":    {en: "unsupported instruction %s", th: "ไม่รองรับคำสั่ง %s"},
	"hint.fpu":           {en: "floating-point (x87) and SSE/AVX instructions are not supported by nasmx86", th: "nasmx86 ไม่รองรับคำสั่งทศนิยม (x87) และคำสั่ง SSE/AVX"},
	"err.privileged":     {en: "%s is a privileged instruction (only the operating system kernel may use it)", th: "%s เป็นคำสั่งสิทธิ์พิเศษ (ใช้ได้เฉพาะเคอร์เนลของระบบปฏิบัติการ)"},
	"err.invalid":        {en: "invalid instruction (bytes %s)", th: "คำสั่งไม่ถูกต้อง (ไบต์ %s)"},
	"hint.invalid":       {en: "the CPU is executing data: did a jump land in the middle of data, or is data placed inside section .text?", th: "CPU กำลังรันข้อมูลแทนโค้ด: มีการกระโดดไปกลางข้อมูล หรือวางข้อมูลไว้ใน section .text หรือเปล่า?"},
	"err.syscall32":      {en: "syscall is not available in 32-bit programs: use int 0x80 with 32-bit system call numbers, or write a 64-bit program", th: "โปรแกรม 32 บิตใช้คำสั่ง syscall ไม่ได้: ให้ใช้ int 0x80 กับหมายเลข system call แบบ 32 บิต หรือเขียนเป็นโปรแกรม 64 บิต"},
	"err.badInt":         {en: "int %#x is not available to Linux programs", th: "โปรแกรมบน Linux ใช้ int %#x ไม่ได้"},
	"hint.dos":           {en: "this is an MS-DOS/BIOS interrupt. On Linux use int 0x80 (32-bit) or syscall (64-bit)", th: "นี่คืออินเทอร์รัปต์ของ MS-DOS/BIOS บน Linux ให้ใช้ int 0x80 (32 บิต) หรือ syscall (64 บิต)"},
	"hint.int":           {en: "Linux programs make system calls with int 0x80 (32-bit) or syscall (64-bit)", th: "โปรแกรมบน Linux เรียก system call ด้วย int 0x80 (32 บิต) หรือ syscall (64 บิต)"},
	"err.int3":           {en: "stopped at int3 (breakpoint)", th: "หยุดที่ int3 (breakpoint)"},
	"hint.int3":          {en: "run with --step to debug interactively", th: "รันด้วย --step เพื่อดีบักทีละคำสั่ง"},
	"err.maxSteps":       {en: "stopped after %d instructions: the program may be stuck in an infinite loop", th: "หยุดหลังจากรันไป %d คำสั่ง: โปรแกรมอาจติดอยู่ในลูปไม่รู้จบ"},
	"hint.maxSteps":      {en: "use --max-steps N to allow more (0 = no limit), or --trace to see what repeats", th: "ใช้ --max-steps N เพื่อเพิ่มจำนวน (0 = ไม่จำกัด) หรือใช้ --trace เพื่อดูว่าส่วนไหนวนซ้ำ"},

	// --step debugger
	"dbg.banner": {en: "nasmx86 debugger: Enter = step one instruction, h = help, q = quit",
		th: "ดีบักเกอร์ nasmx86: กด Enter = รันทีละคำสั่ง, h = ดูวิธีใช้, q = ออก"},
	"dbg.help": {en: `commands:
  Enter, s [n]     step one (or n) instructions
  n                next: like s, but runs a whole call until it returns
  c                continue until a breakpoint or the end of the program
  b <line|label>   set a breakpoint (b alone lists them; b util.asm:5
                   for a line in another file)
  d [n|line]       delete breakpoint n (d alone deletes all)
  r                show registers and flags
  m <addr> [len]   hex dump of memory, e.g.  m msg   m esp 32   m buf+4 8
  x [n], stack     show the top n stack entries (default 8)
  l                list the source around the current line
  q                quit
(Ctrl-C stops a running "c" or "n")
`, th: `คำสั่ง:
  Enter, s [n]     รันทีละ 1 (หรือ n) คำสั่ง
  n                next: เหมือน s แต่ถ้าเป็น call จะรันทั้งฟังก์ชันจนกลับมา
  c                รันต่อจนเจอ breakpoint หรือโปรแกรมจบ
  b <บรรทัด|label>  ตั้ง breakpoint (พิมพ์ b เฉย ๆ เพื่อดูรายการ;
                   b util.asm:5 สำหรับบรรทัดในไฟล์อื่น)
  d [n|บรรทัด]      ลบ breakpoint หมายเลข n (d เฉย ๆ = ลบทั้งหมด)
  r                แสดงรีจิสเตอร์และแฟล็ก
  m <แอดเดรส> [len] แสดงหน่วยความจำแบบเลขฐานสิบหก เช่น  m msg   m esp 32   m buf+4 8
  x [n], stack     แสดงข้อมูลบนสแตก n ช่องบนสุด (ค่าเริ่มต้น 8)
  l                แสดงซอร์สโค้ดรอบบรรทัดปัจจุบัน
  q                ออก
(กด Ctrl-C เพื่อหยุดการรันของ "c" หรือ "n")
`},
	"dbg.unknown":       {en: "unknown command %q (type h for help)", th: "ไม่รู้จักคำสั่ง %q (พิมพ์ h เพื่อดูวิธีใช้)"},
	"dbg.finished":      {en: "the program has already finished (exit code %d). Use r, m, x to look around, q to quit.", th: "โปรแกรมจบไปแล้ว (exit code %d) ใช้ r, m, x เพื่อตรวจดู หรือ q เพื่อออก"},
	"dbg.exited":        {en: "program exited with code %d", th: "โปรแกรมจบด้วย exit code %d"},
	"dbg.crashed":       {en: "the program stopped with an error. You can still inspect it with r, m, x; q quits.", th: "โปรแกรมหยุดเพราะเกิดข้อผิดพลาด ยังตรวจดูได้ด้วย r, m, x และ q เพื่อออก"},
	"dbg.interrupted":   {en: "interrupted", th: "หยุดการรันแล้ว"},
	"dbg.hitBreak":      {en: "breakpoint #%d", th: "ถึง breakpoint #%d"},
	"dbg.noCode":        {en: "no instructions at or after line %d", th: "ไม่มีคำสั่งที่บรรทัด %d หรือหลังจากนั้น"},
	"dbg.notCode":       {en: "%s is not an instruction address", th: "%s ไม่ใช่แอดเดรสของคำสั่ง"},
	"dbg.noFile":        {en: "no input file named %q", th: "ไม่มีไฟล์ชื่อ %q ในโปรแกรมนี้"},
	"dbg.noBreaks":      {en: "no breakpoints", th: "ยังไม่มี breakpoint"},
	"dbg.breakSet":      {en: "breakpoint #%d at %s", th: "ตั้ง breakpoint #%d ที่ %s"},
	"dbg.breaksCleared": {en: "all breakpoints deleted", th: "ลบ breakpoint ทั้งหมดแล้ว"},
	"dbg.breakDeleted":  {en: "breakpoint #%d deleted", th: "ลบ breakpoint #%d แล้ว"},
	"dbg.noSuchBreak":   {en: "no breakpoint %s", th: "ไม่มี breakpoint %s"},
	"dbg.badExpr":       {en: "cannot understand %q (use a label, register, or number like 0x804a000)", th: "ไม่เข้าใจ %q (ใช้ label, ชื่อรีจิสเตอร์ หรือตัวเลข เช่น 0x804a000)"},
	"dbg.memUsage":      {en: "usage: m <address|label|register> [length]", th: "วิธีใช้: m <แอดเดรส|label|รีจิสเตอร์> [ความยาว]"},
	"dbg.unreadable":    {en: "memory at %s is not readable", th: "อ่านหน่วยความจำที่ %s ไม่ได้"},
	"dbg.retAddr":       {en: "return address → line %d", th: "return address → บรรทัด %d"},
	"dbg.retMain":       {en: "return address → exit", th: "return address → exit"},
}
