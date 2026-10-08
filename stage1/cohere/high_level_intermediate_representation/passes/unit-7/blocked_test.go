package unit7

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// This is a selector for an owner schema gap, not a unit-7 pass certificate.
// When Scope replay lands, replace this refusal expectation with pass comparisons.
func TestScopeReplayGap(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	owned := filepath.Join(lane, "passes/unit-7")
	temp := t.TempDir()
	run := func(dir string, env []string, args ...string) []byte {
		t.Helper()
		command := exec.Command(args[0], args[1:]...)
		command.Dir = dir
		command.Env = append(os.Environ(), env...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, output)
		}
		t.Logf("%v: successful (%d output bytes)", args, len(output))
		return output
	}
	target := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	overlay, err := json.Marshal(map[string]map[string]string{"Replace": {
		filepath.Join(target, "unit7_input_test.go"): filepath.Join(owned, "oracle_test.go"),
		filepath.Join(target, "unit7_dump_test.go"):  filepath.Join(lane, "testdata/oracle_test.go"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(temp, "scope-terminal.txt")
	run(filepath.Join(root, "cohere"), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "UNIT7_SCOPE_TERMINAL_OUTPUT=" + inputPath},
		"go", "test", "-overlay", overlayPath, "-tags=lintoracle", "-count=1", "-v", "-run=^TestUnit7ScopeTerminalInput$", "./internal/lint/ecmascript/high_level_intermediate_representation")
	got, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(owned, "testdata/scope-terminal.txt"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Go input: %s", got)
	if !bytes.Equal(got, want) {
		t.Fatalf("Go Scope input changed: %s", got)
	}
	entry := filepath.Join(owned, "testdata/scope-replay.a")
	native := filepath.Join(temp, "scope-native")
	run(root, nil, "go", "run", "./cmd/adamic", "build", entry, "-o", native, "--sanitize")
	javascript := filepath.Join(temp, "scope.mjs")
	javascriptOutput := run(root, nil, "go", "run", "./cmd/adamic", "js", entry)
	if err := os.WriteFile(javascript, javascriptOutput, 0600); err != nil {
		t.Fatal(err)
	}
	for _, runtime := range []struct {
		name string
		args []string
	}{
		{"Node", []string{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), entry, inputPath}},
		{"emitted JavaScript", []string{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), javascript, inputPath}},
		{"sanitized native", []string{native, inputPath}},
	} {
		command := exec.Command(runtime.args[0], runtime.args[1:]...)
		command.Dir = root
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		err := command.Run()
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || stdout.Len() != 0 || stderr.String() != "adamic: panic: unknown terminal Scope\n" {
			t.Fatalf("%s: exit=%v stdout=%q stderr=%q", runtime.name, err, stdout.String(), stderr.String())
		}
		t.Logf("%s: Go Scope input refused, exit=70, %s", runtime.name, stderr.String())
	}
}
