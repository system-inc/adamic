package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
)

// compiler/lower-agree had no pushed branch when this revision started.
// Keep the local helper limited to the source and JavaScript observations.
func lowersAndAgreesWithNode(t *testing.T, source string) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	type observation struct {
		stdout, stderr []byte
		exit           int
	}
	run := func(path string) observation {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", runner, path)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.Output()
		result := observation{stdout: stdout, stderr: stderr.Bytes()}
		if err != nil {
			var exit *exec.ExitError
			if ctx.Err() != nil || !errors.As(err, &exit) {
				t.Fatalf("Node %s: %v", path, err)
			}
			result.exit = exit.ExitCode()
		}
		return result
	}
	want := run(path)
	if want.exit != 0 || len(want.stdout) == 0 || len(want.stderr) != 0 {
		t.Fatalf("source must finish and print its enum results: exit %d, stdout %q, stderr %q", want.exit, want.stdout, want.stderr)
	}
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	if program == nil {
		t.Fatal("lowering returned empty IR")
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	got := run(generated)
	if got.exit != want.exit || !bytes.Equal(got.stdout, want.stdout) || !bytes.Equal(got.stderr, want.stderr) {
		t.Errorf("JavaScript backend: exit %d, stdout %q, stderr %q; source Node: exit %d, stdout %q, stderr %q", got.exit, got.stdout, got.stderr, want.exit, want.stdout, want.stderr)
	}
}
