// Command nasmx86 assembles, links and runs x86 (32/64-bit) Linux NASM
// programs on any machine — no VM, no Docker, no Rosetta.
package main

import (
	"fmt"
	"os"
)

// version is overwritten at release time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-V") {
		fmt.Println("nasmx86", version)
		return
	}
	fmt.Fprintln(os.Stderr, "usage: nasmx86 [options] file.asm [program args...]")
	os.Exit(2)
}
