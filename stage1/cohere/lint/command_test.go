package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The go command's module downloads are the only stderr execute forgives, and only from go.
func TestCommandDiagnosticsDropOnlyModuleDownloads(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, command, stderr, want string
	}{
		{"one download", "go", "go: downloading github.com/dlclark/regexp2 v1.11.5\n", ""},
		{"two downloads", "go", "go: downloading github.com/dlclark/regexp2/v2 v2.5.2\ngo: downloading github.com/dlclark/regexp2 v1.11.5\n", ""},
		{"a warning stays", "go", "go: warning: ignoring symlink\n", "go: warning: ignoring symlink\n"},
		{"a compile error stays", "go", "go: downloading a.example/m v1.0.0\n./x.go:1: undefined: y\n", "./x.go:1: undefined: y\n"},
		{"a download without a version stays", "go", "go: downloading github.com/dlclark/regexp2\n", "go: downloading github.com/dlclark/regexp2\n"},
		{"a download with more after it stays", "go", "go: downloading a.example/m v1.0.0 (cached)\n", "go: downloading a.example/m v1.0.0 (cached)\n"},
		{"a download not at a line start stays", "go", "note go: downloading a.example/m v1.0.0\n", "note go: downloading a.example/m v1.0.0\n"},
		{"only go is forgiven", "node", "go: downloading a.example/m v1.0.0\n", "go: downloading a.example/m v1.0.0\n"},
	} {
		if got := string(commandDiagnostics(row.command, []byte(row.stderr))); got != row.want {
			t.Errorf("%s: got %q, want %q", row.name, got, row.want)
		}
	}
}

// Not parallel: each case runs this test binary again with a fake go as the toolchain's.
// execute passes a go command whose only stderr is module downloads, and fails one that writes anything else.
// Not parallel: t.Setenv changes the process-wide ADAMIC_TOOLS in the stderr probe.
func TestExecuteFailsOnStderrOtherThanModuleDownloads(t *testing.T) {
	if lines := os.Getenv("ADAMIC_LINT_STDERR_PROBE"); lines != "" {
		// execute resolves go through toolchain, under ADAMIC_TOOLS and toolchain's environment, so the fake is the tools
		// directory's go and carries its stderr in a file of its own rather than in an environment variable.
		tools := t.TempDir()
		if err := os.MkdirAll(filepath.Join(tools, "bin"), 0755); err != nil {
			t.Fatal(err)
		}
		stderr := filepath.Join(tools, "stderr")
		if err := os.WriteFile(stderr, []byte(lines), 0644); err != nil {
			t.Fatal(err)
		}
		script := "#!/bin/sh\ncat '" + stderr + "' >&2\n"
		if err := os.WriteFile(filepath.Join(tools, "bin", "go"), []byte(script), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("ADAMIC_TOOLS", tools)
		execute(t, "", "go", "version")
		return
	}
	for _, row := range []struct {
		name, stderr string
		passes       bool
	}{
		{"downloads only", "go: downloading github.com/dlclark/regexp2 v1.11.5\ngo: downloading github.com/dlclark/regexp2/v2 v2.5.2\n", true},
		{"a download and a warning", "go: downloading github.com/dlclark/regexp2 v1.11.5\ngo: warning: planted stderr\n", false},
		{"a warning alone", "go: warning: planted stderr\n", false},
	} {
		command := exec.Command(os.Args[0], "-test.run=^TestExecuteFailsOnStderrOtherThanModuleDownloads$", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_LINT_STDERR_PROBE="+row.stderr)
		output, err := command.CombinedOutput()
		if row.passes && err != nil {
			t.Errorf("%s: execute failed on module downloads alone: %v\n%s", row.name, err, output)
		}
		if !row.passes && (err == nil || !strings.Contains(string(output), "planted stderr")) {
			t.Errorf("%s: execute passed, or failed without naming the stderr: %v\n%s", row.name, err, output)
		}
	}
}
