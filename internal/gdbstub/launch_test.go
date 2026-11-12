package gdbstub

import (
	"runtime"
	"slices"
	"testing"
)

func TestUTF8Locale(t *testing.T) {
	loc := "C.UTF-8"
	if runtime.GOOS == "darwin" {
		loc = "en_US.UTF-8"
	}
	cases := []struct {
		env, want []string
	}{
		{[]string{"HOME=/h"}, []string{"HOME=/h", "LC_CTYPE=" + loc}},
		{[]string{"LANG=en_US.UTF-8"}, []string{"LANG=en_US.UTF-8"}},
		{[]string{"LANG=th_TH.utf8"}, []string{"LANG=th_TH.utf8"}},
		{[]string{"LC_CTYPE=UTF-8"}, []string{"LC_CTYPE=UTF-8"}},
		{[]string{"LANG=th_TH.UTF-8", "LC_CTYPE=C"}, []string{"LANG=th_TH.UTF-8", "LC_CTYPE=" + loc}},
		{[]string{"LC_ALL=C", "LANG=en_US.UTF-8"}, []string{"LANG=en_US.UTF-8", "LC_ALL=" + loc}},
		{[]string{"LC_ALL=en_US.UTF-8", "LC_CTYPE=C"}, []string{"LC_ALL=en_US.UTF-8", "LC_CTYPE=C"}},
	}
	for _, c := range cases {
		if got := utf8Locale(c.env); !slices.Equal(got, c.want) {
			t.Errorf("utf8Locale(%q) = %q, want %q", c.env, got, c.want)
		}
	}
}
