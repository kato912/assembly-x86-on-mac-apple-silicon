package runner

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestCRLFReader(t *testing.T) {
	r := &crlfReader{r: iotest.OneByteReader(strings.NewReader("12\r\n34\r\nx\ry\r"))}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "12\n34\nx\ry\r" {
		t.Errorf("got %q", b)
	}
}
