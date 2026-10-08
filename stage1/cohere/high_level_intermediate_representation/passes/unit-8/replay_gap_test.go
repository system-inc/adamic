package unit8

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Remove this gap test when hir-01 supplies the three terminal records and codecs.
// A successful decode must turn this test red, rather than silently retaining a stop.
func TestReplayTerminalSchemaRequests(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) []byte {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, output)
		}
		return output
	}
	entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-8/replay_gap.a")
	scratch := t.TempDir()
	binary := filepath.Join(scratch, "probe")
	javascript := filepath.Join(scratch, "probe.mjs")
	run("go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	if err := os.WriteFile(javascript, run("go", "run", "./cmd/adamic", "js", entry), 0600); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"Scope", "Sequence", "MaybeThrow"} {
		for _, runtime := range []struct {
			name string
			args []string
		}{
			{"source Node", []string{"node", "--no-warnings", "oracle/node.mjs", entry, kind}},
			{"emitted JavaScript", []string{"node", "--no-warnings", "oracle/node.mjs", javascript, kind}},
			{"sanitized native", []string{binary, kind}},
		} {
			cmd := exec.Command(runtime.args[0], runtime.args[1:]...)
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			want := []byte("adamic: panic: unknown terminal " + kind + "\n")
			if !ok || exit.ExitCode() != 70 || !bytes.Equal(output, want) {
				t.Fatalf("%s %s: want missing-schema panic exit 70; got %v, %q", runtime.name, kind, err, output)
			}
			t.Logf("%s: %s shared schema missing, exit 70", runtime.name, kind)
		}
	}
}
