package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

// buildTest262Compiler has the buildcache.Product callback signature and takes no test state.
// Its inputs are the Go module/workspace files, cmd/adamic, internal, bridge, cohere, and the
// Go toolchain/build environment. The product contains only the compiler executable.
func buildTest262Compiler(directory string) error {
	command := exec.Command("go", "build", "-o", filepath.Join(directory, "adamic"), "./cmd/adamic")
	command.Dir = "../.."
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("building test262 compiler: %w\n%s", err, output)
	}
	return nil
}

var splitCompilerDirectory string
var splitCompiler = sync.OnceValues(func() (string, error) {
	directory, err := os.MkdirTemp("", "test262-compiler-")
	if err != nil {
		return "", err
	}
	splitCompilerDirectory = directory
	if err := buildTest262Compiler(directory); err != nil {
		return "", err
	}
	return filepath.Join(directory, "adamic"), nil
})

func TestMain(m *testing.M) {
	status := m.Run()
	if splitCompilerDirectory != "" {
		os.RemoveAll(splitCompilerDirectory)
	}
	os.Exit(status)
}

func prepareSplitTest262(t *testing.T, corpus string) (*engine, error) {
	t.Helper()
	compiler, err := splitCompiler()
	if err != nil {
		return nil, err
	}
	return prepareModeWithCompiler("../..", corpus, t.TempDir(), nil, false, compiler)
}
