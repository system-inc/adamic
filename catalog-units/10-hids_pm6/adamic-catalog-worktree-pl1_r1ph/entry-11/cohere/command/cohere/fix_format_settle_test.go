package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestOneFixFormatRunSettlesAPrintedArrow is the case that found the gap, through the real rule and the
// real printer: a one-line block-bodied arrow that the printer breaks across lines, under
// nexus/consistency-no-multiline-arrow-function, which repairs a multi-line arrow. Found as 51 arrows
// in 23 of www-phi-health's files, which one `--fix --format` formatted and left for the next run's
// fixer, so the run's own `--no-fix` check flagged them.
//
// One writing run must leave nothing for the check: the fix the printed text triggers lands in the
// same run, and the fixed text is formatted again.
func TestOneFixFormatRunSettlesAPrintedArrow(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"nexus/consistency-no-multiline-arrow-function":"error"}}`,
		"NexusCohereSettings.json": `{"format":{"tabWidth":4,"singleQuote":true,"printWidth":120,"ignore":[]}}`,
		"Probe.ts":                 "export function register(listener: () => void): void {\n    listener();\n}\n\nregister(() => { console.log('ready'); });\n",
	})

	output, code := runCohere(t, binary, root, "--fix", "--format", "Probe.ts")
	if code != 0 {
		t.Fatalf("the writing run exited %d:\n%s", code, output)
	}
	written := readForTest(t, filepath.Join(root, "Probe.ts"))
	if !strings.Contains(written, "register(function() {\n    console.log('ready');\n});") {
		t.Fatalf("the arrow the printer broke across lines was not repaired in the same run:\n%s\noutput:\n%s", written, output)
	}

	output, code = runCohere(t, binary, root, "--no-fix", "--format", "Probe.ts")
	if code != 0 || !strings.Contains(output, "0 files would change") {
		t.Fatalf("one --fix --format run left work for the next, exit %d:\n%s", code, output)
	}
}

// TestAFixToARecordedFileIsFormatted: a file on the format record is out of the default scope, but the
// record vouches only for the bytes it hashed. A fix that rewrites the file hands the formatter text the
// record never saw, and that text is formatted in the same run rather than written as the fixer left it.
//
// Found through the nested drift check, which records a library's files with only the formatter asked:
// a library file holding a `debugger` was on record, so the library's own `--fix --format` removed the
// statement and wrote the result unformatted. A rule switched on after a file was recorded is the same
// case in one repository.
func TestAFixToARecordedFileIsFormatted(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"off"}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":[]}}`,
		"Producer.ts":              "export function value(): number {\n  debugger;\n  return 1;\n}\n",
	})

	// No rule asks for a fix, so the formatted file is recorded at its own bytes (the settings files are
	// formatted alongside it).
	output, code := runCohere(t, binary, root, "--fix", "--format")
	if code != 0 || readForTest(t, filepath.Join(root, "Producer.ts")) != "export function value(): number {\n  debugger;\n  return 1;\n}\n" {
		t.Fatalf("the run before the rule changed the formatted file, exit %d:\n%s", code, output)
	}
	output, _ = runCohere(t, binary, root, "--no-fix", "--format")
	if !strings.Contains(output, "0 of 4 files not on record as formatted") {
		t.Fatalf("the file was not on record before the rule:\n%s", output)
	}

	writeTree(t, root, map[string]string{
		"CohereSettings.json": `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"error"}}`,
	})
	output, code = runCohere(t, binary, root, "--fix", "--format")
	if code != 0 {
		t.Fatalf("the writing run exited %d:\n%s", code, output)
	}
	if got := readForTest(t, filepath.Join(root, "Producer.ts")); got != "export function value(): number {\n  return 1;\n}\n" {
		t.Fatalf("the fix to a recorded file was written unformatted:\n%q\noutput:\n%s", got, output)
	}

	output, code = runCohere(t, binary, root, "--no-fix", "--format")
	if code != 0 || !strings.Contains(output, "0 files would change") {
		t.Fatalf("one --fix --format run left a recorded file's fix for the next, exit %d:\n%s", code, output)
	}
}
