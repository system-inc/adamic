package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

// replaceProcess runs the cohere binary and exits with its exit code, the nearest Windows has to exec:
// it has no call that replaces a process, and syscall.Exec there fails every time.
//
// The binary inherits the console's three handles, so it reads and writes the terminal itself. Ctrl+C
// reaches every process on the console, the binary included, so this one ignores it and waits for the
// binary to end however it chooses to. Its exit code is passed on as it is, since a wrapper that lost a
// non-zero exit is how a gate goes quietly green.
func replaceProcess(binaryPath string, arguments []string) error {
	command := exec.Command(binaryPath, arguments...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	signal.Ignore(os.Interrupt)
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return fmt.Errorf("running %s: %w", binaryPath, err)
	}
	os.Exit(command.ProcessState.ExitCode())
	return nil
}
