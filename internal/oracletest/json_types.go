// Package oracletest prepares the source oracle for Go integration tests.
package oracletest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Run builds the JSON descriptor tool once for a package's TestMain, before any
// fixture's execution deadline starts. Children inherit its path, including
// tests that run Node through another harness. The binary lasts until run ends.
func Run(run func() int) int {
	root, err := filepath.Abs("../..")
	if err != nil {
		fmt.Fprintln(os.Stderr, "oracle json_types:", err)
		return 1
	}
	directory, err := os.MkdirTemp("", "adamic-json-types-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "oracle json_types:", err)
		return 1
	}
	defer os.RemoveAll(directory)
	binary := filepath.Join(directory, "json_types")
	command := exec.Command("go", "build", "-o", binary, "./oracle/json_types.go")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "oracle json_types build: %v\n%s", err, output)
		return 1
	}
	if err := os.Setenv("ADAMIC_JSON_TYPES", binary); err != nil {
		fmt.Fprintln(os.Stderr, "oracle json_types:", err)
		return 1
	}
	fmt.Println("oracle json_types: built once for this package run")
	return run()
}
