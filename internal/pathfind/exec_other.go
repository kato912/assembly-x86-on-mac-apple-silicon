//go:build !unix

package pathfind

import (
	"errors"
	"os"
	"os/exec"
)

// Exec runs the program at path in nasmx86's place and exits with its exit
// code. It returns only on failure to start it.
func Exec(path string, args, env []string) error {
	cmd := exec.Command(path, args...)
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		os.Exit(ee.ExitCode())
	}
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
