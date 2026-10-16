// Package checkplanted holds planted mutants for ahra loom check's proof (#cg02c0f): one per step. It never lands.
package checkplanted

import (
	"fmt"
	"os/exec"
)

// Compile runs clang itself, outside native, which the compiler census and the plan's compilerDependencies refuse.
func Compile(path string) error {
	fmt.Printf("compiling %d\n", path)
	return exec.Command("clang", "-c", path).Run()
}

func   Misformatted() int { return 1 }
