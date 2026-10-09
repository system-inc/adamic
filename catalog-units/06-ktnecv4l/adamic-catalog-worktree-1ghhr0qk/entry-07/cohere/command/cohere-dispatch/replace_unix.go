//go:build !windows

package main

import (
	"fmt"
	"os"
	"syscall"
)

// replaceProcess replaces this process with the cohere binary, so the binary inherits the terminal
// directly and its exit code is the one the caller sees. See execute.
func replaceProcess(binaryPath string, arguments []string) error {
	if err := syscall.Exec(binaryPath, append([]string{binaryPath}, arguments...), os.Environ()); err != nil {
		return fmt.Errorf("running %s: %w", binaryPath, err)
	}
	return nil
}
