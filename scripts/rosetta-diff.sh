#!/bin/sh
# Developer tool (Apple Silicon only): check nasmx86 against real x86-64
# execution under Rosetta 2. Each 64-bit differential test is also built as
# a macOS x86-64 program (dump64.inc switches to macOS system calls when
# MACOS is defined) and both runs' stdout and exit codes are compared.
# Tests whose first line contains "linux-only" are skipped. A test must exit
# with 0, or with the code given in a "; exit: N" line.
#
#   scripts/rosetta-diff.sh [test.asm ...]
set -u
cd "$(dirname "$0")/.."
INC=testdata/diff/inc/
SDK=$(xcrun --show-sdk-path)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
go build -o "$WORK/nasmx86" ./cmd/nasmx86 || exit 1
[ $# -eq 0 ] && set -- testdata/diff/64/*.asm
pass=0 fail=0 skip=0
for t in "$@"; do
    name=$(basename "$t" .asm)
    if head -1 "$t" | grep -q linux-only; then skip=$((skip+1)); continue; fi
    in=/dev/null
    [ -f "${t%.asm}.in" ] && in="${t%.asm}.in"
    if ! nasm -f macho64 -DMACOS -I"$INC" "$t" -o "$WORK/$name.o" ||
       ! ld -arch x86_64 -e _start -platform_version macos 11.0 11.0 -syslibroot "$SDK" -lSystem \
            "$WORK/$name.o" -o "$WORK/$name" 2>/dev/null; then
        echo "BUILD FAIL $name"; fail=$((fail+1)); continue
    fi
    arch -x86_64 "$WORK/$name" < "$in" > "$WORK/$name.native" 2>/dev/null; nc=$?
    "$WORK/nasmx86" -64 -I"$INC" "$t" < "$in" > "$WORK/$name.emu" 2>"$WORK/$name.err"; ec=$?
    want=$(sed -n 's/^; *exit: *\([0-9]*\).*/\1/p' "$t" | head -1)
    if [ "$nc" -ne "${want:-0}" ]; then
        echo "FAIL $name: native run exited with $nc, test expects ${want:-0}"; fail=$((fail+1)); continue
    fi
    if [ $nc -eq $ec ] && cmp -s "$WORK/$name.native" "$WORK/$name.emu"; then
        echo "ok   $name"; pass=$((pass+1))
    else
        echo "FAIL $name (exit native=$nc nasmx86=$ec)"; fail=$((fail+1))
        diff "$WORK/$name.native" "$WORK/$name.emu" | head -20
        head -5 "$WORK/$name.err"
    fi
done
echo "passed $pass, failed $fail, skipped $skip"
[ $fail -eq 0 ]
