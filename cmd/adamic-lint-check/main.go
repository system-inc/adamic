// adamic-lint-check runs the selected-rule certification harness from the repository root.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: adamic-lint-check <slug>")
		os.Exit(2)
	}
	command := exec.Command("go", "test", "./stage1/cohere/lint", "-run", "^TestRule$", "-count=1", "-timeout", "30m", "-v", "-args", "-rule", os.Args[1])
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
