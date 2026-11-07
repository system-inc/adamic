package lint

import (
	"github.com/system-inc/adamic/stage1/cohere/executable"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The go command's module downloads are the only stderr execute forgives, and only from go.
func TestCommandDiagnosticsDropOnlyModuleDownloads(test *testing.T) {
	test.Parallel()
	for _, row := range []struct {
		name, command, stderr, want string
	}{
		{"one download", "go", "go: downloading github.com/dlclark/regexp2 v1.11.5\n", ""},
		{
			"two downloads",
			"go",
			"go: downloading github.com/dlclark/regexp2/v2 v2.5.2\ngo: downloading github.com/dlclark/regexp2 v1.11.5\n",
			"",
		},
		{"a warning stays", "go", "go: warning: ignoring symlink\n", "go: warning: ignoring symlink\n"},
		{
			"a compile error stays",
			"go",
			"go: downloading a.example/m v1.0.0\n./x.go:1: undefined: y\n",
			"./x.go:1: undefined: y\n",
		},
		{
			"a download without a version stays",
			"go",
			"go: downloading github.com/dlclark/regexp2\n",
			"go: downloading github.com/dlclark/regexp2\n",
		},
		{
			"a download with more after it stays",
			"go",
			"go: downloading a.example/m v1.0.0 (cached)\n",
			"go: downloading a.example/m v1.0.0 (cached)\n",
		},
		{
			"a download not at a line start stays",
			"go",
			"note go: downloading a.example/m v1.0.0\n",
			"note go: downloading a.example/m v1.0.0\n",
		},
		{"only go is forgiven", "node", "go: downloading a.example/m v1.0.0\n", "go: downloading a.example/m v1.0.0\n"},
	} {
		if got := string(commandDiagnostics(row.command, []byte(row.stderr))); got != row.want {
			test.Errorf("%s: got %q, want %q", row.name, got, row.want)
		}
	}
}

// Not parallel: each case runs this test binary again with a fake go as the toolchain's.
// execute passes a go command whose only stderr is module downloads, and fails one that writes anything else.
// Not parallel: test.Setenv changes the process-wide ADAMIC_TOOLS in the stderr probe.
func TestExecuteFailsOnStderrOtherThanModuleDownloads(test *testing.T) {
	if lines := os.Getenv("ADAMIC_LINT_STDERR_PROBE"); lines != "" {
		// execute resolves go through toolchain, under ADAMIC_TOOLS and toolchain's environment, so the fake is the
		// tools directory's go and carries its stderr in a file of its own rather than in an environment variable.
		tools := test.TempDir()
		if createDirectoryError := os.MkdirAll(filepath.Join(tools, "bin"), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		stderr := filepath.Join(tools, "stderr")
		if writeError := os.WriteFile(stderr, []byte(lines), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		script := "#!/bin/sh\ncat '" + stderr + "' >&2\n"
		if writeError := executable.Write(filepath.Join(tools, "bin", "go"), []byte(script)); writeError != nil {
			test.Fatal(writeError)
		}
		test.Setenv("ADAMIC_TOOLS", tools)
		execute(test, "", "go", "version")
		return
	}
	for _, row := range []struct {
		name, stderr string
		passes       bool
	}{
		{
			"downloads only",
			"go: downloading github.com/dlclark/regexp2 v1.11.5\ngo: downloading github.com/dlclark/regexp2/v2 v2.5.2\n",
			true,
		},
		{
			"a download and a warning",
			"go: downloading github.com/dlclark/regexp2 v1.11.5\ngo: warning: planted stderr\n",
			false,
		},
		{"a warning alone", "go: warning: planted stderr\n", false},
	} {
		command := exec.Command(os.Args[0], "-test.run=^TestExecuteFailsOnStderrOtherThanModuleDownloads$", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_LINT_STDERR_PROBE="+row.stderr)
		output, combinedOutputError := command.CombinedOutput()
		if row.passes && combinedOutputError != nil {
			test.Errorf("%s: execute failed on module downloads alone: %v\n%s", row.name, combinedOutputError, output)
		}
		if !row.passes && (combinedOutputError == nil || !strings.Contains(string(output), "planted stderr")) {
			test.Errorf(
				"%s: execute passed, or failed without naming the stderr: %v\n%s",
				row.name,
				combinedOutputError,
				output,
			)
		}
	}
}
