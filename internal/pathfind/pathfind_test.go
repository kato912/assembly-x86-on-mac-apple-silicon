//go:build unix

package pathfind

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOtherSkipsSelf(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Skip(err)
	}
	mine, real := t.TempDir(), t.TempDir()
	os.Symlink(self, filepath.Join(mine, "faketool")) // nasmx86 standing in for faketool
	tool := filepath.Join(real, "faketool")
	os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o755)
	t.Setenv("PATH", mine+string(filepath.ListSeparator)+real)
	if got, ok := Other("faketool"); !ok || got != tool {
		t.Errorf("Other = %q %v, want %q", got, ok, tool)
	}
	if !IsSelf(filepath.Join(mine, "faketool")) || IsSelf(tool) {
		t.Error("IsSelf")
	}
	t.Setenv("PATH", mine)
	if got, ok := Other("faketool", real); !ok || got != tool {
		t.Errorf("Other with dirs = %q %v", got, ok)
	}
	if _, ok := Other("faketool"); ok {
		t.Error("found only nasmx86 itself")
	}
}
