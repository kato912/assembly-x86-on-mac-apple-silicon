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
  --ddd              debug in DDD, the graphical debugger (uses gdb)
  --gdb[=PORT]       let gdb connect (target extended-remote localhost:1234)
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
  nasmx86 --ddd loop.asm
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
  --ddd              ดีบักด้วย DDD ดีบักเกอร์แบบกราฟิก (ทำงานผ่าน gdb)
  --gdb[=PORT]       ให้ gdb เชื่อมต่อเข้ามา (target extended-remote localhost:1234)
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
  nasmx86 --ddd loop.asm
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
	"err.undefined": {en: "your program uses %s, which is not defined in your files or in nasmx86's built-in C library",
		th: "โปรแกรมเรียกใช้ %s ซึ่งไม่ได้นิยามไว้ในไฟล์ของคุณ และไม่มีในไลบรารี C ที่ติดมากับ nasmx86"},
	"hint.undefined": {en: "check the spelling (names are case-sensitive). C functions you can call with extern: %s",
		th: "ตรวจตัวสะกด (ตัวพิมพ์เล็ก/ใหญ่มีผล) ฟังก์ชัน C ที่เรียกได้ด้วย extern: %s"},
	"err.internal":        {en: "internal error: %v", th: "ข้อผิดพลาดภายใน: %v"},
	"err.duplicate":       {en: "the label %s is defined (global) in more than one file: %s", th: "label %s ถูกประกาศเป็น global ในหลายไฟล์: %s"},
	"hint.undefinedMulti": {en: "if %[1]s is in one of your other .asm files, declare it there with:  global %[1]s", th: "ถ้า %[1]s อยู่ในไฟล์ .asm อื่นของคุณ ต้องประกาศในไฟล์นั้นด้วย:  global %[1]s"},

	// runtime warnings
	"warn.efault": {en: "warning: %s: the buffer at %#x is not valid memory, so the call failed (returned -EFAULT)", th: "คำเตือน: %s: บัฟเฟอร์ที่ %#x ไม่ใช่หน่วยความจำที่ใช้ได้ การเรียกจึงล้มเหลว (คืนค่า -EFAULT)"},
	"warn.enosys": {en: "warning: system call %d is not supported by nasmx86 (returned -ENOSYS)", th: "คำเตือน: nasmx86 ยังไม่รองรับ system call หมายเลข %d (คืนค่า -ENOSYS)"},
	"warn.halt":   {en: "note: the program stopped at hlt. (On real Linux, hlt in a user program crashes with a segmentation fault — use the exit system call.)", th: "หมายเหตุ: โปรแกรมหยุดที่คำสั่ง hlt (บน Linux จริง hlt ในโปรแกรมผู้ใช้จะทำให้เกิด segmentation fault — ควรใช้ system call exit)"},
	"warn.badfd":  {en: "warning: %s: file descriptor %d is not open (returned -EBADF)", th: "คำเตือน: %s: file descriptor %d ไม่ได้เปิดอยู่ (คืนค่า -EBADF)"},

	"warn.libcAlign": {en: "warning: %s was called with a misaligned stack (rsp = %s before the call). It works here, but printf and scanf can crash on real Linux: keep rsp a multiple of 16 at every call (e.g. push rbp at the start of main)",
		th: "คำเตือน: เรียก %s ตอนที่สแตกไม่ได้จัดแนว (rsp = %s ก่อน call) ในนี้ยังทำงานได้ แต่บน Linux จริง printf และ scanf อาจพังได้: ให้ rsp หารด้วย 16 ลงตัวทุกครั้งที่ call (เช่น push rbp ที่ต้น main)"},
	"warn.libcLostCrash": {en: "warning: %d bytes of printf/puts output were still in the C library's buffer when the program crashed, so they were never printed: %s (on real Linux they are lost too)",
		th: "คำเตือน: ตอนโปรแกรมพังยังมีข้อความจาก printf/puts ค้างอยู่ในบัฟเฟอร์ของไลบรารี C %d ไบต์ จึงไม่ได้แสดงออกมา: %s (บน Linux จริงก็หายเหมือนกัน)"},
	"warn.libcLost": {en: "warning: the program ended with %d bytes of printf/puts output still in the C library's buffer, so they were never printed: %s\n    hint: buffered output is written by exit — end with call exit (or return from main) instead of the exit system call, or call fflush first. On real Linux this output is lost too",
		th: "คำเตือน: โปรแกรมจบไปโดยยังมีข้อความจาก printf/puts ค้างอยู่ในบัฟเฟอร์ของไลบรารี C %d ไบต์ จึงไม่ได้แสดงออกมา: %s\n    คำแนะนำ: ข้อความในบัฟเฟอร์จะถูกเขียนออกตอน exit — ให้จบด้วย call exit (หรือ return ออกจาก main) แทน system call exit หรือเรียก fflush ก่อน บน Linux จริงข้อความนี้ก็หายเหมือนกัน"},

	"gdb.waiting": {en: "waiting for gdb on %s. In another terminal run:\n    %s %s -ex 'target extended-remote %s'\n(the program's input and output stay in this terminal; Ctrl-C here quits)",
		th: "กำลังรอ gdb ที่ %s ให้เปิดอีกเทอร์มินัลแล้วรัน:\n    %s %s -ex 'target extended-remote %s'\n(อินพุตและเอาต์พุตของโปรแกรมอยู่ในเทอร์มินัลนี้ กด Ctrl-C ที่นี่เพื่อออก)"},
	"gdb.listen": {en: "cannot listen on %s: %v", th: "เปิดพอร์ต %s ไม่ได้: %v"},
	"gdb.missing": {en: "no gdb that can debug x86 was found. Install one with Homebrew:\n    brew install x86_64-elf-gdb",
		th: "ไม่พบ gdb ที่ดีบักโค้ด x86 ได้ ติดตั้งด้วย Homebrew:\n    brew install x86_64-elf-gdb"},
	"ddd.missing": {en: "DDD is not installed. Install it (and a gdb for x86, and the X11 window system DDD needs) with Homebrew:\n    brew install ddd x86_64-elf-gdb\n    brew install --cask xquartz      (then log out and log in once)",
		th: "ยังไม่ได้ติดตั้ง DDD ติดตั้ง (พร้อม gdb สำหรับ x86 และระบบหน้าต่าง X11 ที่ DDD ต้องใช้) ด้วย Homebrew:\n    brew install ddd x86_64-elf-gdb\n    brew install --cask xquartz      (แล้วออกจากระบบและเข้าใหม่หนึ่งครั้ง)"},
	"ddd.started": {en: "starting DDD; its gdb connects to nasmx86 at %s. The program's input and output are in this terminal.",
		th: "กำลังเปิด DDD โดย gdb ของ DDD เชื่อมต่อกับ nasmx86 ที่ %s อินพุตและเอาต์พุตของโปรแกรมอยู่ในเทอร์มินัลนี้"},
	"ddd.noXQuartz": {en: "DDD needs XQuartz (the X11 window system for macOS) to open its window. Install it once with:\n    brew install --cask xquartz\nthen run this command again (nasmx86 starts XQuartz by itself).",
		th: "DDD ต้องใช้ XQuartz (ระบบหน้าต่าง X11 สำหรับ macOS) เพื่อเปิดหน้าต่าง ติดตั้งครั้งเดียวด้วย:\n    brew install --cask xquartz\nแล้วรันคำสั่งนี้อีกครั้ง (nasmx86 จะเปิด XQuartz ให้เอง)"},
	"ddd.noDisplay": {en: "DDD could not open its window. Start XQuartz (open -a XQuartz) and try again; if it still fails, log out and log in once.",
		th: "DDD เปิดหน้าต่างไม่ได้ ให้เปิด XQuartz (open -a XQuartz) แล้วลองอีกครั้ง ถ้ายังไม่ได้ให้ออกจากระบบแล้วเข้าใหม่หนึ่งครั้ง"},
	"ddd.failed": {en: "DDD stopped: %v\n%s", th: "DDD หยุดทำงาน: %v\n%s"},

	// source locations
	"loc.line":  {en: "line %d: %s", th: "บรรทัด %d: %s"},
	"loc.macro": {en: "(in macro: %s)", th: "(ในแมโคร: %s)"},
	"loc.stub":  {en: "(after returning from main)", th: "(หลังจาก return ออกจาก main)"},
	"loc.libc":  {en: "%s (C library)", th: "%s (ไลบรารี C)"},

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
	"err.libcPtr": {en: "%s could not access memory at %s, so it crashed (segmentation fault inside the C library)",
		th: "%s เข้าถึงหน่วยความจำที่ %s ไม่ได้ จึงพัง (segmentation fault ภายในไลบรารี C)"},
	"hint.libcNull": {en: "that is a NULL pointer (0): did you forget to put an address in the argument (e.g. mov rdi, msg in 64-bit, push msg in 32-bit)?",
		th: "นั่นคือพอยน์เตอร์ NULL (0): ลืมใส่แอดเดรสเป็นอาร์กิวเมนต์หรือเปล่า (เช่น mov rdi, msg ในโหมด 64 บิต, push msg ในโหมด 32 บิต)"},
	"hint.libcValue": {en: "that looks like a number, not an address: pass the variable's address (num), not its value ([num]) — e.g. mov rsi, num in 64-bit or push num in 32-bit. scanf needs addresses to store the values in",
		th: "ค่านี้ดูเหมือนตัวเลข ไม่ใช่แอดเดรส: ให้ส่งแอดเดรสของตัวแปร (num) ไม่ใช่ค่าของมัน ([num]) — เช่น mov rsi, num ในโหมด 64 บิต หรือ push num ในโหมด 32 บิต scanf ต้องการแอดเดรสไว้เก็บค่า"},
	"hint.libcPtr": {en: "the pointer points outside the program's memory or past the end of a buffer (is the string terminated with 0?)",
		th: "พอยน์เตอร์ชี้ออกนอกหน่วยความจำของโปรแกรม หรือเลยท้ายบัฟเฟอร์ (สตริงปิดท้ายด้วย 0 หรือยัง?)"},
	"err.libcAlign": {en: "%s was called with a misaligned stack (rsp = %s before the call), which crashes it on Linux (segmentation fault inside the C library)",
		th: "เรียก %s ตอนที่สแตกไม่ได้จัดแนว (rsp = %s ก่อน call) ซึ่งบน Linux จะทำให้พัง (segmentation fault ภายในไลบรารี C)"},
	"hint.libcAlign": {en: "the x86-64 ABI requires rsp to be a multiple of 16 at every call. At the start of main, push rbp (or sub rsp, 8), and keep pushes and pops paired",
		th: "ABI ของ x86-64 กำหนดให้ rsp หารด้วย 16 ลงตัวทุกครั้งที่ call — ที่ต้น main ให้ push rbp (หรือ sub rsp, 8) และให้ push กับ pop สมดุลกัน"},
	"err.libcFloat": {en: "%s: %s prints a floating-point value, which 64-bit code passes in the xmm registers; nasmx86 does not emulate floating point (SSE/x87) yet",
		th: "%s: %s ใช้แสดงค่าทศนิยม ซึ่งโค้ด 64 บิตส่งผ่านรีจิสเตอร์ xmm — nasmx86 ยังไม่รองรับทศนิยม (SSE/x87)"},
	"hint.libcFloat": {en: "print whole numbers with %d instead", th: "ใช้ %d แสดงจำนวนเต็มแทน"},
	"err.libcFile": {en: "%s: %s is not a FILE pointer (it should be the value of stdin/stdout/stderr, or what fopen returned)",
		th: "%s: %s ไม่ใช่ FILE pointer (ต้องเป็นค่าของ stdin/stdout/stderr หรือค่าที่ fopen คืนมา)"},
	"hint.libcFile": {en: "declare extern stdout and load its value: mov rdi, [stdout] (64-bit) or push dword [stdout] (32-bit)",
		th: "ประกาศ extern stdout แล้วโหลดค่าของมัน: mov rdi, [stdout] (64 บิต) หรือ push dword [stdout] (32 บิต)"},
	"err.libcFileVar": {en: "%s was given the address of the variable %s instead of its value",
		th: "%s ได้รับแอดเดรสของตัวแปร %s แทนที่จะเป็นค่าของมัน"},
	"hint.libcFileVar": {en: "use [%[1]s], not %[1]s: mov rdi, [%[1]s] (64-bit) or push dword [%[1]s] (32-bit)",
		th: "ใช้ [%[1]s] ไม่ใช่ %[1]s: mov rdi, [%[1]s] (64 บิต) หรือ push dword [%[1]s] (32 บิต)"},
	"err.libcAbort": {en: "the program called abort() (SIGABRT)", th: "โปรแกรมเรียก abort() (SIGABRT)"},
	"err.libcDoubleFree": {en: "free(): double free detected — the memory at %s was already freed (Linux aborts the program)",
		th: "free(): ตรวจพบการ free ซ้ำ — หน่วยความจำที่ %s ถูก free ไปแล้ว (Linux จะยกเลิกโปรแกรม)"},
	"hint.libcDoubleFree": {en: "free each block once; set the pointer to 0 after freeing it",
		th: "free แต่ละบล็อกครั้งเดียว และตั้งพอยน์เตอร์เป็น 0 หลัง free"},
	"err.libcBadFree": {en: "%s(): invalid pointer %s — it was not returned by malloc/calloc/realloc (Linux aborts the program)",
		th: "%s(): พอยน์เตอร์ %s ไม่ถูกต้อง — ไม่ใช่ค่าที่ได้จาก malloc/calloc/realloc (Linux จะยกเลิกโปรแกรม)"},
	"hint.libcBadFree": {en: "pass exactly the address malloc returned, not an address inside the block or of a variable",
		th: "ส่งแอดเดรสที่ malloc คืนมาตรง ๆ ไม่ใช่แอดเดรสกลางบล็อกหรือแอดเดรสของตัวแปร"},
	"err.privileged": {en: "%s is a privileged instruction (only the operating system kernel may use it)", th: "%s เป็นคำสั่งสิทธิ์พิเศษ (ใช้ได้เฉพาะเคอร์เนลของระบบปฏิบัติการ)"},
	"err.invalid":    {en: "invalid instruction (bytes %s)", th: "คำสั่งไม่ถูกต้อง (ไบต์ %s)"},
	"hint.invalid":   {en: "the CPU is executing data: did a jump land in the middle of data, or is data placed inside section .text?", th: "CPU กำลังรันข้อมูลแทนโค้ด: มีการกระโดดไปกลางข้อมูล หรือวางข้อมูลไว้ใน section .text หรือเปล่า?"},
	"err.syscall32":  {en: "syscall is not available in 32-bit programs: use int 0x80 with 32-bit system call numbers, or write a 64-bit program", th: "โปรแกรม 32 บิตใช้คำสั่ง syscall ไม่ได้: ให้ใช้ int 0x80 กับหมายเลข system call แบบ 32 บิต หรือเขียนเป็นโปรแกรม 64 บิต"},
	"err.badInt":     {en: "int %#x is not available to Linux programs", th: "โปรแกรมบน Linux ใช้ int %#x ไม่ได้"},
	"hint.dos":       {en: "this is an MS-DOS/BIOS interrupt. On Linux use int 0x80 (32-bit) or syscall (64-bit)", th: "นี่คืออินเทอร์รัปต์ของ MS-DOS/BIOS บน Linux ให้ใช้ int 0x80 (32 บิต) หรือ syscall (64 บิต)"},
	"hint.int":       {en: "Linux programs make system calls with int 0x80 (32-bit) or syscall (64-bit)", th: "โปรแกรมบน Linux เรียก system call ด้วย int 0x80 (32 บิต) หรือ syscall (64 บิต)"},
	"err.int3":       {en: "stopped at int3 (breakpoint)", th: "หยุดที่ int3 (breakpoint)"},
	"hint.int3":      {en: "run with --step to debug interactively", th: "รันด้วย --step เพื่อดีบักทีละคำสั่ง"},
	"err.maxSteps":   {en: "stopped after %d instructions: the program may be stuck in an infinite loop", th: "หยุดหลังจากรันไป %d คำสั่ง: โปรแกรมอาจติดอยู่ในลูปไม่รู้จบ"},
	"hint.maxSteps":  {en: "use --max-steps N to allow more (0 = no limit), or --trace to see what repeats", th: "ใช้ --max-steps N เพื่อเพิ่มจำนวน (0 = ไม่จำกัด) หรือใช้ --trace เพื่อดูว่าส่วนไหนวนซ้ำ"},

	// --step debugger
	"dbg.banner": {en: "nasmx86 debugger: Enter = step one instruction, h = help, q = quit",
		th: "ดีบักเกอร์ nasmx86: กด Enter = รันทีละคำสั่ง, h = ดูวิธีใช้, q = ออก"},
	"dbg.help": {en: `commands:
  Enter, s [n]       step one (or n) instructions
  n                  next: like s, but runs a whole call until it returns
  finish             run until the current function returns (shows eax/rax)
  c                  continue until a breakpoint, a watchpoint or the end
  b <line|label>     set a breakpoint (b alone lists them; b util.asm:5
                     for a line in another file)
  watch <what>       stop when it changes: watch ecx, watch counter,
                     watch byte [buf+3] (watch alone lists them)
  d [n|line]         delete breakpoint/watchpoint n (d alone deletes all)
  r                  show registers and flags
  p <expr>           print a value: p eax, p [count], p dword [arr+ecx*4],
                     p/s msg (a string), p ecx*4+1
  set <what> <val>   change it: set eax 5, set zf 1, set dword [count] 0
  m <addr> [len]     hex dump of memory, e.g.  m msg   m esp 32   m buf+4 8
  x [n], stack       show the top n stack entries (default 8)
  bt                 show the call stack (which function called which)
  l                  list the source around the current line
  q                  quit
(Ctrl-C stops a running c, n or finish)

gdb commands work too, so lab sheets written for gdb can be followed as is:
  break/tbreak <loc|*addr>, run, start, kill, stepi/nexti [n], continue,
  until/advance <loc>, info registers [eax ...], info breakpoints,
  p/x p/d p/u p/t p/c p/o <expr>  ($eax, $pc, &msg, (int)count, *(int*)p),
  x/NFU <addr> (x/4xw $esp, x/s &msg, x/16xb msg, x/5i $pc),
  display[/FMT] <expr>, undisplay, info display, disassemble [func],
  enable/disable [n], set $eax = 5, set var (int)count = 0, backtrace, finish
`, th: `คำสั่ง:
  Enter, s [n]       รันทีละ 1 (หรือ n) คำสั่ง
  n                  next: เหมือน s แต่ถ้าเป็น call จะรันทั้งฟังก์ชันจนกลับมา
  finish             รันจนออกจากฟังก์ชันปัจจุบัน (แสดงค่า eax/rax ที่คืนมา)
  c                  รันต่อจนเจอ breakpoint, watchpoint หรือโปรแกรมจบ
  b <บรรทัด|label>    ตั้ง breakpoint (พิมพ์ b เฉย ๆ เพื่อดูรายการ;
                     b util.asm:5 สำหรับบรรทัดในไฟล์อื่น)
  watch <สิ่งที่ดู>     หยุดเมื่อค่าเปลี่ยน: watch ecx, watch counter,
                     watch byte [buf+3] (watch เฉย ๆ = ดูรายการ)
  d [n|บรรทัด]        ลบ breakpoint/watchpoint หมายเลข n (d เฉย ๆ = ลบทั้งหมด)
  r                  แสดงรีจิสเตอร์และแฟล็ก
  p <นิพจน์>          แสดงค่า: p eax, p [count], p dword [arr+ecx*4],
                     p/s msg (ข้อความ), p ecx*4+1
  set <สิ่งที่แก้> <ค่า> แก้ค่า: set eax 5, set zf 1, set dword [count] 0
  m <แอดเดรส> [len]   แสดงหน่วยความจำ เช่น  m msg   m esp 32   m buf+4 8
  x [n], stack       แสดงข้อมูลบนสแตก n ช่องบนสุด (ค่าเริ่มต้น 8)
  bt                 แสดงลำดับการเรียกฟังก์ชัน (ใครเรียกใคร)
  l                  แสดงซอร์สโค้ดรอบบรรทัดปัจจุบัน
  q                  ออก
(กด Ctrl-C เพื่อหยุดการรันของ c, n หรือ finish)

คำสั่งแบบ gdb ก็ใช้ได้ ทำตามชีทแลปที่เขียนไว้สำหรับ gdb ได้เลย:
  break/tbreak <ตำแหน่ง|*แอดเดรส>, run, start, kill, stepi/nexti [n], continue,
  until/advance <ตำแหน่ง>, info registers [eax ...], info breakpoints,
  p/x p/d p/u p/t p/c p/o <นิพจน์>  ($eax, $pc, &msg, (int)count, *(int*)p),
  x/NFU <แอดเดรส> (x/4xw $esp, x/s &msg, x/16xb msg, x/5i $pc),
  display[/FMT] <นิพจน์>, undisplay, info display, disassemble [ฟังก์ชัน],
  enable/disable [n], set $eax = 5, set var (int)count = 0, backtrace, finish
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
	"dbg.breaksCleared": {en: "all breakpoints and watchpoints deleted", th: "ลบ breakpoint และ watchpoint ทั้งหมดแล้ว"},
	"dbg.breakDeleted":  {en: "breakpoint #%d deleted", th: "ลบ breakpoint #%d แล้ว"},
	"dbg.noSuchBreak":   {en: "no breakpoint %s", th: "ไม่มี breakpoint %s"},
	"dbg.badExpr":       {en: "cannot understand %q (use a label, register, or number like 0x804a000)", th: "ไม่เข้าใจ %q (ใช้ label, ชื่อรีจิสเตอร์ หรือตัวเลข เช่น 0x804a000)"},
	"dbg.pUsage":        {en: "usage: p <expression>   e.g.  p eax   p [count]   p dword [arr + ecx*4]   p/s msg", th: "วิธีใช้: p <นิพจน์>   เช่น  p eax   p [count]   p dword [arr + ecx*4]   p/s msg"},
	"dbg.pLabel":        {en: "(%[1]s is an address; to see the value stored there use  p [%[1]s]  or  p dword [%[1]s])", th: "(%[1]s คือแอดเดรส ถ้าจะดูค่าที่เก็บไว้ให้ใช้  p [%[1]s]  หรือ  p dword [%[1]s])"},
	"dbg.setUsage":      {en: "usage: set <register|flag|eip> <value>  or  set [byte|word|dword|qword] [address] <value>\n  e.g.  set eax 5   set zf 1   set dword [count] 0   set byte [msg] 'h'", th: "วิธีใช้: set <รีจิสเตอร์|แฟล็ก|eip> <ค่า>  หรือ  set [byte|word|dword|qword] [แอดเดรส] <ค่า>\n  เช่น  set eax 5   set zf 1   set dword [count] 0   set byte [msg] 'h'"},
	"dbg.setMemFail":    {en: "cannot write memory at %s (not writable — e.g. section .text is read-only)", th: "เขียนหน่วยความจำที่ %s ไม่ได้ (ส่วนนั้นเขียนไม่ได้ เช่น section .text อ่านได้อย่างเดียว)"},
	"dbg.finishTop":     {en: "not inside a function (no call to return from); use c to continue", th: "ตอนนี้ไม่ได้อยู่ในฟังก์ชัน (ไม่มี call ให้ return กลับ) ใช้ c เพื่อรันต่อ"},
	"dbg.returned":      {en: "returned from %s: %s = %s", th: "ออกจาก %s แล้ว: %s = %s"},
	"dbg.watchUsage":    {en: "usage: watch <register>  or  watch [byte|word|dword|qword] <address>   e.g.  watch ecx   watch counter   watch byte [buf+3]", th: "วิธีใช้: watch <รีจิสเตอร์>  หรือ  watch [byte|word|dword|qword] <แอดเดรส>   เช่น  watch ecx   watch counter   watch byte [buf+3]"},
	"dbg.watchSet":      {en: "watchpoint #%d: %s (now %s)", th: "ตั้ง watchpoint #%d: %s (ตอนนี้ %s)"},
	"dbg.watchHit":      {en: "watchpoint #%d: %s changed %s → %s\n    by %s", th: "watchpoint #%d: %s เปลี่ยนจาก %s → %s\n    โดย %s"},
	"dbg.watchDeleted":  {en: "watchpoint #%d deleted", th: "ลบ watchpoint #%d แล้ว"},
	"dbg.noWatches":     {en: "no watchpoints", th: "ยังไม่มี watchpoint"},
	"dbg.badRegister":   {en: "no register named %q", th: "ไม่มีรีจิสเตอร์ชื่อ %q"},
	"dbg.infoUsage":     {en: "info registers [names] | info breakpoints | info display | info frame", th: "info registers [ชื่อ] | info breakpoints | info display | info frame"},
	"dbg.cannotAccess":  {en: "Cannot access memory at address %s", th: "เข้าถึงหน่วยความจำที่แอดเดรส %s ไม่ได้"},
	"dbg.disasHeader":   {en: "Dump of assembler code for function %s:", th: "โค้ดแอสเซมบลีของฟังก์ชัน %s:"},
	"dbg.disasEnd":      {en: "End of assembler dump.", th: "จบรายการโค้ด"},
	"dbg.noDisplays":    {en: "There are no auto-display expressions now.", th: "ยังไม่มีนิพจน์ที่ตั้ง display ไว้"},
	"dbg.restarting":    {en: "Starting program from the beginning.", th: "เริ่มรันโปรแกรมใหม่ตั้งแต่ต้น"},
	"dbg.killed":        {en: "Program killed. (run starts it again)", th: "หยุดโปรแกรมแล้ว (พิมพ์ run เพื่อเริ่มใหม่)"},
	"dbg.noTui":         {en: "TUI layouts are not available; try  display/i $pc  and  display $eax  (shown after every step), or  info registers", th: "ไม่มีหน้าจอแบบ TUI ให้ใช้  display/i $pc  และ  display $eax  (แสดงทุกครั้งที่รัน) หรือ  info registers  แทน"},
	"dbg.attSyntax":     {en: "note: nasmx86 always shows Intel/NASM syntax", th: "หมายเหตุ: nasmx86 แสดงโค้ดแบบ Intel/NASM เสมอ"},
	"dbg.memUsage":      {en: "usage: m <address|label|register> [length]", th: "วิธีใช้: m <แอดเดรส|label|รีจิสเตอร์> [ความยาว]"},
	"dbg.unreadable":    {en: "memory at %s is not readable", th: "อ่านหน่วยความจำที่ %s ไม่ได้"},
	"dbg.retAddr":       {en: "return address → line %d", th: "return address → บรรทัด %d"},
	"dbg.retMain":       {en: "return address → exit", th: "return address → exit"},
}
