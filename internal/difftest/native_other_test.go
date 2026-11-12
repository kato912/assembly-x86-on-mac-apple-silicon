//go:build !linux

package difftest

import (
	"errors"
	"testing"
)

func nativeAvailable() (bool, string) { return false, "native x86 Linux execution needs linux/amd64" }

func libcAvailable(bits int) (bool, string) { return false, "needs linux/amd64" }

func runNative(t *testing.T, p program) ([]byte, int, error) {
	return nil, 0, errors.New("unsupported")
}
