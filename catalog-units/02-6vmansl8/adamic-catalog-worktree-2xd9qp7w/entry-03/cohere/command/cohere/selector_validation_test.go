package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPipelineRefusesAZeroMatchOverrideSelector proves the validation is in the live command, not
// merely available as an unused configuration helper.
//
// The second run is the mutation control: changing only the selector from an impossible path to the
// file that exists must turn the configuration failure into a successful lint run. Without the call
// in run(), the first invocation exits cleanly and this test catches the original vacuous pass.
func TestPipelineRefusesAZeroMatchOverrideSelector(t *testing.T) {
	t.Parallel()

	binary := buildCohere(t)

	directory := t.TempDir()
	writeSelectorValidationProject(t, directory, "missing/**")

	command := exec.Command(binary, "--lint", "--no-fix", "--directory", directory)
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("the pipeline exited successfully with a zero-match override selector:\n%s", output)
	}
	if !strings.Contains(string(output), "missing/**") {
		t.Fatalf("the pipeline failure does not name the selector to repair:\n%s", output)
	}

	writeSelectorValidationProject(t, directory, "**/*.ts")
	command = exec.Command(binary, "--lint", "--no-fix", "--directory", directory)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("the same project failed after its selector was changed to match index.ts: %v\n%s", err, output)
	}
}

func writeSelectorValidationProject(t *testing.T, directory string, selector string) {
	t.Helper()

	files := map[string]string{
		"tsconfig.json": `{"compilerOptions":{"strict":true},"include":["*.ts"]}`,
		"index.ts":      "export const value = 1;\n",
		"CohereSettings.json": `{"rules":{},"overrides":[{"files":["` + selector +
			`"],"rules":{"no-debugger":"off"}}]}`,
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
}
