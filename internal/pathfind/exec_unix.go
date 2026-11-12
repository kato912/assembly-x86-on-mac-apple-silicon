//go:build unix

package pathfind

import "syscall"

// Exec replaces nasmx86 with the program at path (argv[0] = path), so the
// Mac's own tool runs exactly as if it had been called directly. It
// returns only on failure.
func Exec(path string, args, env []string) error {
	return syscall.Exec(path, append([]string{path}, args...), env)
}
