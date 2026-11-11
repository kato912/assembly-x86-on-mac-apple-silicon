; linux-only: stat/fstat/mkdir/rmdir/rename/nanosleep/clock_gettime/
; gettimeofday. Only deterministic results (return values, sizes, file
; types) are compared, never the clock values themselves. Results are kept
; in memory because syscall overwrites rcx and r11.
%include "dump64.inc"

%macro SYS 1-4                  ; SYS nr, rdi, rsi, rdx  (result in rax)
    mov eax, %1
  %if %0 > 1
    mov rdi, %2
  %endif
  %if %0 > 2
    mov rsi, %3
  %endif
  %if %0 > 3
    mov rdx, %4
  %endif
    syscall
%endmacro

%macro KEEP 1                   ; res[%1] = rax
    mov [res + 8*%1], rax
%endmacro

%macro SHOW 0                   ; dump res[0..5] in rbx rdx rsi rdi r8 r9
    mov rbx, [res]
    mov rdx, [res + 8]
    mov rsi, [res + 16]
    mov rdi, [res + 24]
    mov r8,  [res + 32]
    mov r9,  [res + 40]
    xor eax, eax
    xor ecx, ecx
    xor r11d, r11d
    DUMP FL_NONE
    mov qword [res], 0
    mov qword [res + 8], 0
    mov qword [res + 16], 0
    mov qword [res + 24], 0
    mov qword [res + 32], 0
    mov qword [res + 40], 0
%endmacro

section .data
fname:  db "nasmx86_sc2_64.tmp", 0
dname:  db "nasmx86_sc2_64.dir", 0
dname2: db "nasmx86_sc2_64.dir2", 0
text:   db "thirteen byte"
nap:    dq 0, 1000000
badnap: dq 0, 1000000000
section .bss
res:    resq 6
st:     resb 144
ts:     resq 2
fd:     resq 1

section .text
global _start
_start:
    INIT
    LABEL "fstat / stat"
    SYS 2, fname, 0x241, 0o644  ; open(fname, O_WRONLY|O_CREAT|O_TRUNC)
    mov [fd], rax
    SYS 1, [fd], text, 13       ; write 13 bytes
    SYS 5, [fd], st             ; fstat(fd, st)
    KEEP 0
    mov rax, [st + 48]          ; st_size = 13
    KEEP 1
    mov eax, [st + 24]
    and eax, 0o170000           ; S_IFREG
    KEEP 2
    SYS 3, [fd]                 ; close
    SYS 4, fname, st            ; stat(fname, st)
    KEEP 3
    mov rax, [st + 48]
    KEEP 4
    SYS 4, dname, st            ; stat of a missing file -> -ENOENT
    KEEP 5
    SHOW
    SYS 5, 0, st                ; fstat(0): stdin is a pipe here
    KEEP 0
    mov eax, [st + 24]
    and eax, 0o170000           ; S_IFIFO
    KEEP 1
    SHOW

    LABEL "mkdir / rename / rmdir"
    SYS 83, dname, 0o755
    KEEP 0                      ; 0
    SYS 83, dname, 0o755
    KEEP 1                      ; -EEXIST
    SYS 82, dname, dname2
    KEEP 2                      ; 0
    SYS 84, dname2
    KEEP 3                      ; 0
    SYS 84, dname2
    KEEP 4                      ; -ENOENT
    SYS 84, fname
    KEEP 5                      ; -ENOTDIR
    SYS 87, fname               ; unlink
    SHOW

    LABEL "time"
    SYS 35, nap, 0              ; nanosleep 1 ms -> 0
    KEEP 0
    SYS 35, badnap, 0           ; tv_nsec = 1e9 -> -EINVAL
    KEEP 1
    SYS 228, 1, ts              ; clock_gettime(CLOCK_MONOTONIC) -> 0
    KEEP 2
    SYS 228, 0, ts              ; clock_gettime(CLOCK_REALTIME) -> 0
    KEEP 3
    xor eax, eax
    cmp qword [ts], 1700000000  ; after Nov 2023
    seta al
    KEEP 4
    SYS 96, ts, 0               ; gettimeofday(ts, NULL) -> 0
    KEEP 5
    SHOW
    EXIT 0
