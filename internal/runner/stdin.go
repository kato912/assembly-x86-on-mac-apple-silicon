package runner

import (
	"io"
	"os"
	"runtime"
)

// consoleStdin returns the program's standard input. On a Windows console
// typed lines end in "\r\n"; Linux programs expect "\n", so the carriage
// returns are dropped there.
func consoleStdin() io.Reader {
	if runtime.GOOS == "windows" {
		if st, err := os.Stdin.Stat(); err == nil && st.Mode()&os.ModeCharDevice != 0 {
			return &crlfReader{r: os.Stdin}
		}
	}
	return os.Stdin
}

// crlfReader turns "\r\n" into "\n".
type crlfReader struct {
	r       io.Reader
	pending bool // a '\r' was held back at the end of the previous read
}

func (c *crlfReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	buf := make([]byte, len(p))
	n, err := c.r.Read(buf)
	out := 0
	if c.pending {
		c.pending = false
		if n == 0 || buf[0] != '\n' {
			p[out] = '\r'
			out++
		}
	}
	for i := 0; i < n; i++ {
		b := buf[i]
		if b == '\r' {
			if i+1 < n {
				if buf[i+1] == '\n' {
					continue
				}
			} else if err == nil {
				c.pending = true
				continue
			}
		}
		if out < len(p) {
			p[out] = b
			out++
		}
	}
	return out, err
}
