package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRCBlock(t *testing.T) {
	for _, orig := range []string{"", "export A=1\n", "export A=1", "# mine\n\n"} {
		added := withBlock(orig)
		if !strings.Contains(added, rcBlock) || !strings.HasPrefix(added, orig) {
			t.Errorf("withBlock(%q) = %q", orig, added)
		}
		if again := withBlock(added); again != added {
			t.Errorf("withBlock twice:\n%q\n%q", added, again)
		}
		want := orig
		if want != "" && !strings.HasSuffix(want, "\n") {
			want += "\n"
		}
		if removed := withoutBlock(added); removed != want {
			t.Errorf("withoutBlock(withBlock(%q)) = %q", orig, removed)
		}
	}
	// Lines the user added after the block stay.
	text := withBlock("a\n") + "b\n"
	if got := withoutBlock(text); got != "a\nb\n" {
		t.Errorf("got %q", got)
	}
}

func TestSetupFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("NASMX86_HOME", home)
	t.Setenv("SHELL", "/bin/zsh")
	zshrc := filepath.Join(home, ".zshrc")
	os.WriteFile(zshrc, []byte("export A=1\n"), 0o644)
	if setupSupported() {
		if code := setupMain(nil); code != 0 {
			t.Fatalf("setup: %d", code)
		}
		if data, _ := os.ReadFile(zshrc); !strings.Contains(string(data), rcBlock) {
			t.Fatalf("~/.zshrc:\n%s", data)
		}
		if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
			t.Error("setup created ~/.bash_profile for a zsh user")
		}
	} else {
		os.WriteFile(zshrc, []byte(withBlock("export A=1\n")), 0o644)
	}
	if code := setupMain([]string{"--remove"}); code != 0 {
		t.Fatalf("setup --remove: %d", code)
	}
	if data, _ := os.ReadFile(zshrc); string(data) != "export A=1\n" {
		t.Errorf("after --remove:\n%q", data)
	}
	if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
		t.Error("--remove created ~/.bash_profile")
	}
}
