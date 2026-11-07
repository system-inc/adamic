package worker

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratorsThroughSymlink(test *testing.T) {
	test.Parallel()
	for _, script := range []string{"generate-crossing.mjs", "generate.mjs"} {
		test.Run(script, func(test *testing.T) {
			test.Parallel()
			directory := test.TempDir()
			template, absolutePathError := filepath.Abs("wasm")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			link := filepath.Join(directory, "linked")
			if symlinkError := os.Symlink(template, link); symlinkError != nil {
				test.Fatal(symlinkError)
			}
			output := filepath.Join(directory, "output")
			arguments := []string{filepath.Join(link, script), output}
			generated := filepath.Join(output, "worker.mjs")
			if script == "generate-crossing.mjs" {
				abi := filepath.Join(directory, "abi.json")
				write(test, abi, `{"version":1,"exports":[]}`)
				generated = filepath.Join(directory, "crossing.mjs")
				arguments = []string{filepath.Join(link, script), abi, generated}
			}
			command := exec.Command("node", arguments...)
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				test.Fatalf("%s through symlink: %v\n%s", script, combinedOutputError, output)
			}
			if _, statError := os.Stat(generated); statError != nil {
				test.Fatalf("%s exited successfully without output through symlink: %v", script, statError)
			}
		})
	}
}

func TestGenerateCrossingThroughSymlink(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	link := filepath.Join(test.TempDir(), "linked")
	if symlinkError := os.Symlink(directory, link); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	abi := filepath.Join(link, "abi.json")
	write(test, abi, `{"version":1,"exports":[]}`)
	output := filepath.Join(link, "crossing.mjs")
	if generateCrossingError := generateCrossing(abi, output); generateCrossingError != nil {
		test.Fatal(generateCrossingError)
	}
	if _, statError := os.Stat(output); statError != nil {
		test.Fatalf("embedded generator exited successfully without output through symlink: %v", statError)
	}
}

// Not parallel: changes PATH for a generator process that writes no output.
func TestGenerateCrossingMissingOutput(test *testing.T) {
	// Not parallel: PATH selects a generator process that exits successfully without writing.
	directory := test.TempDir()
	node := filepath.Join(directory, "node")
	write(test, node, "#!/bin/sh\nexit 0\n")
	if changeModeError := os.Chmod(node, 0755); changeModeError != nil {
		test.Fatal(changeModeError)
	}
	test.Setenv("PATH", directory)
	abi := filepath.Join(directory, "abi.json")
	write(test, abi, `{"version":1,"exports":[]}`)
	generateCrossingError := generateCrossing(abi, filepath.Join(directory, "crossing.mjs"))
	if generateCrossingError == nil ||
		!errors.Is(generateCrossingError, os.ErrNotExist) ||
		!strings.Contains(generateCrossingError.Error(), "generate-crossing.mjs") {
		test.Fatalf("missing output must name the generator and preserve the file error: %v", generateCrossingError)
	}
}
