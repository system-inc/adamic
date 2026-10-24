package worker

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratorsThroughSymlink(t *testing.T) {
	t.Parallel()
	for _, script := range []string{"generate-crossing.mjs", "generate.mjs"} {
		t.Run(script, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			template, err := filepath.Abs("wasm")
			if err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(directory, "linked")
			if err := os.Symlink(template, link); err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(directory, "output")
			arguments := []string{filepath.Join(link, script), output}
			generated := filepath.Join(output, "worker.mjs")
			if script == "generate-crossing.mjs" {
				abi := filepath.Join(directory, "abi.json")
				write(t, abi, `{"version":1,"exports":[]}`)
				generated = filepath.Join(directory, "crossing.mjs")
				arguments = []string{filepath.Join(link, script), abi, generated}
			}
			command := exec.Command("node", arguments...)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("%s through symlink: %v\n%s", script, err, output)
			}
			if _, err := os.Stat(generated); err != nil {
				t.Fatalf("%s exited successfully without output through symlink: %v", script, err)
			}
		})
	}
}

func TestGenerateCrossingThroughSymlink(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	link := filepath.Join(t.TempDir(), "linked")
	if err := os.Symlink(directory, link); err != nil {
		t.Fatal(err)
	}
	abi := filepath.Join(link, "abi.json")
	write(t, abi, `{"version":1,"exports":[]}`)
	output := filepath.Join(link, "crossing.mjs")
	if err := generateCrossing(abi, output); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("embedded generator exited successfully without output through symlink: %v", err)
	}
}

func TestGenerateCrossingMissingOutput(t *testing.T) {
	// Not parallel: PATH selects a generator process that exits successfully without writing.
	directory := t.TempDir()
	node := filepath.Join(directory, "node")
	write(t, node, "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(node, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	abi := filepath.Join(directory, "abi.json")
	write(t, abi, `{"version":1,"exports":[]}`)
	err := generateCrossing(abi, filepath.Join(directory, "crossing.mjs"))
	if err == nil || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "generate-crossing.mjs") {
		t.Fatalf("missing output must name the generator and preserve the file error: %v", err)
	}
}
