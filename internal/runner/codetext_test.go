package runner

import "testing"

func TestCodeText(t *testing.T) {
	cases := map[string]string{
		"mov     eax, 4      ; sys_write": "mov eax, 4",
		"\tmsg db  'a  ;b', 10 ; c":       "msg db 'a  ;b', 10",
		"  ret  ":                         "ret",
		"; only comment":                  "",
	}
	for in, want := range cases {
		if got := CodeText(in); got != want {
			t.Errorf("CodeText(%q) = %q, want %q", in, got, want)
		}
	}
}
