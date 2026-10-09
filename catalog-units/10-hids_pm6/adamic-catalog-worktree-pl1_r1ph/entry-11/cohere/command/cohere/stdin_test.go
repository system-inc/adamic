package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCohereWithStdin runs the binary from a directory with text on stdin, and returns stdout, stderr
// and the exit code separately: stdout is the file's text in this mode, and mixing a note into it would
// corrupt the save.
func runCohereWithStdin(t *testing.T, binary string, directory string, stdin string, arguments ...string) (string, string, int) {
	t.Helper()
	command := exec.Command(binary, arguments...)
	command.Dir = directory
	command.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	exitError, isExit := err.(*exec.ExitError)
	if !isExit {
		t.Fatalf("running cohere: %v\n%s", err, stderr.String())
	}
	return stdout.String(), stderr.String(), childExitCode(t, exitError)
}

// TestStdinAnswersWhatTheGateWrites is the editor save's contract: for a buffer at a path, the stdin
// mode prints exactly what `cohere --fix --format` writes when the same text is on disk at that path,
// and writes nothing itself.
//
// The TypeScript buffer differs from the file on disk and carries a fixable finding the disk version
// does not have, so the answer has to come from a graph that holds the buffer: computed from the disk,
// the fix's offsets would land in the wrong text. The markdown buffer has no program behind it, which
// is the format-only path, and the one the gate itself used to bail on.
func TestStdinAnswersWhatTheGateWrites(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)

	for _, testCase := range []struct {
		name     string
		buffer   string
		contains string
	}{
		{
			name:     "Producer.ts",
			buffer:   "export function value(): number {\n  const   first = 1\n  debugger;\n    return first }\n",
			contains: "const first = 1;\n",
		},
		{
			name:     "Notes.md",
			buffer:   "Title\n=====\n\n*  one\n*  two\n",
			contains: "- one\n- two\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			fixScopeProject(t, root, map[string]string{"Notes.md": "# Notes\n"})
			path := filepath.Join(root, testCase.name)
			onDisk := readForTest(t, path)

			arguments := []string{"--fix", "--format"}
			saved, stderr, code := runCohereWithStdin(t, binary, root, testCase.buffer,
				append(arguments, "--stdin-filepath", testCase.name)...)
			if code != 0 {
				t.Fatalf("exit %d:\n%s", code, stderr)
			}
			if after := readForTest(t, path); after != onDisk {
				t.Fatalf("the stdin mode wrote to disk:\n%s", after)
			}
			if !strings.Contains(saved, testCase.contains) || strings.Contains(saved, "debugger") {
				t.Fatalf("the buffer was not fixed and formatted, wanted %q in:\n%s", testCase.contains, saved)
			}

			writeTree(t, root, map[string]string{testCase.name: testCase.buffer})
			output, gateCode := runCohere(t, binary, root, append(arguments, testCase.name)...)
			if gateCode != 0 {
				t.Fatalf("the gate exited %d:\n%s", gateCode, output)
			}
			if written := readForTest(t, path); written != saved {
				t.Fatalf("the save and the gate disagree\n--- gate\n%s--- save\n%s", written, saved)
			}
		})
	}
}

// TestStdinDeclinesWithoutAComplaint covers the two quiet answers: a buffer that does not parse comes
// back as typed, and a file type nothing formats comes back as typed. A save is not where a typo is
// reported, and a file cohere does not handle saves untouched.
func TestStdinDeclinesWithoutAComplaint(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	fixScopeProject(t, root, map[string]string{"Data.txt": "plain\n"})

	for name, buffer := range map[string]string{
		"Producer.ts": "export function value(: number {\n",
		"Data.txt":    "plain   text\n",
	} {
		saved, stderr, code := runCohereWithStdin(t, binary, root, buffer,
			"--fix", "--format", "--stdin-filepath", name)
		if code != 0 || saved != buffer {
			t.Errorf("%s: exit %d, wanted the buffer back unchanged:\n%s\nstderr:\n%s", name, code, saved, stderr)
		}
	}
}

// TestStdinNeedsFix: the mode answers what --fix writes, so without --fix there is no question.
func TestStdinNeedsFix(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	fixScopeProject(t, root, nil)
	for _, arguments := range [][]string{
		{"--format", "--stdin-filepath", "Producer.ts"},
		{"--fix", "--no-fix", "--stdin-filepath", "Producer.ts"},
		{"--fix", "--stdin-filepath", "Producer.ts", "Sibling.ts"},
	} {
		if saved, stderr, code := runCohereWithStdin(t, binary, root, "", arguments...); code == 0 || saved != "" || stderr == "" {
			t.Errorf("%v: exit %d, wanted a refusal on stderr and nothing on stdout:\n%s", arguments, code, stderr)
		}
	}
}
