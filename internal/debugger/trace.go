package debugger

import (
	"fmt"
	"io"
)

// Trace runs the program to completion, printing every executed
// instruction with its source line and the registers, flags, memory and
// system calls it changed. It returns the error that stopped the program.
func Trace(s *Session, out io.Writer) error {
	for {
		r := s.Step()
		if r.Ran {
			fmt.Fprintln(out, r.Format(s.C.Steps))
		}
		if r.Err != nil {
			return r.Err
		}
	}
}
