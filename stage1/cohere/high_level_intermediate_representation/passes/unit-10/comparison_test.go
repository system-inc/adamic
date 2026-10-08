package unit10

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, root string, env []string, args ...string) []byte {
	t.Helper()
	command := exec.Command(args[0], args[1:]...)
	command.Dir = root
	command.Env = append(os.Environ(), env...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, output)
	}
	return output
}
func location(t *testing.T) (string, string) {
	t.Helper()
	own, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root, own
}
func oracle(t *testing.T, root, own string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	mapping := map[string]string{
		filepath.Join(upstream, "stage1_unit10_test.go"):     filepath.Join(own, "oracle_adapter.go.txt"),
		filepath.Join(upstream, "stage1_hir_oracle_test.go"): filepath.Join(lane, "testdata/oracle_test.go"),
	}
	data, err := json.Marshal(map[string]any{"Replace": mapping})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(own, "testdata/comparison.input.json")
	before := filepath.Join(dir, "before.json")
	after := filepath.Join(dir, "after.txt")
	scope := filepath.Join(dir, "scope.graph.txt")
	environment := []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "HIR_UNIT10_INPUT=" + input, "HIR_UNIT10_BEFORE=" + before, "HIR_UNIT10_AFTER=" + after, "HIR_UNIT10_SCOPE=" + scope}
	output := run(t, filepath.Join(root, "cohere"), environment, "go", "test", "-tags=lintoracle", "-overlay="+overlay, "-run=^TestUnit10", "-count=1", "./internal/lint/ecmascript/high_level_intermediate_representation")
	if err := os.WriteFile(filepath.Join(dir, "go.log"), output, 0600); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(before)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("Go before-state changed")
	}
	return after, scope
}
func emitted(t *testing.T, root, entry, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "emitted.mjs")
	data := run(t, root, nil, "go", "run", "./cmd/adamic", "js", entry)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestDependencyComparison(t *testing.T) {
	t.Parallel()
	root, own := location(t)
	after, _ := oracle(t, root, own)
	want, err := os.ReadFile(after)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join(own, "testdata/comparison.output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(golden, want) {
		t.Fatal("Go after-state differs from committed output fixture")
	}
	if strings.Count(string(want), "\n") != 46 {
		t.Fatal("comparison inventory changed")
	}
	input := filepath.Join(own, "testdata/comparison.input.json")
	entry := filepath.Join(own, "comparison_main.a")
	node := run(t, root, nil, "node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), entry, input)
	if !bytes.Equal(node, want) {
		t.Fatalf("Node comparison differs\n%s", node)
	}
	t.Log("Node: 20 dependency cases, 25 merges, 1 message case match Go; census certificate 0/1465")
	binary := filepath.Join(t.TempDir(), "comparison")
	run(t, root, nil, "go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	native := run(t, root, nil, binary, input)
	if !bytes.Equal(native, want) {
		t.Fatalf("sanitized native comparison differs\n%s", native)
	}
	t.Log("sanitized native: 20 dependency cases, 25 merges, 1 message case match Go; census certificate 0/1465")
	javascript := emitted(t, root, entry, t.TempDir())
	js := run(t, root, nil, "node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), javascript, input)
	if !bytes.Equal(js, want) {
		t.Fatalf("emitted JavaScript comparison differs\n%s", js)
	}
	t.Log("emitted JavaScript: 46/46 comparison probes")

	// A successful semantic mutant accepts a missing manual dependency path.
	mutant := t.TempDir()
	for _, name := range []string{"comparison.a", "comparison_main.a"} {
		data, err := os.ReadFile(filepath.Join(own, name))
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ReplaceAll(string(data), "../../", filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")+"/")
		if name == "comparison.a" {
			if strings.Count(text, "return 3;") != 1 {
				t.Fatal("mutant anchor changed")
			}
			text = strings.Replace(text, "return 3;", "return 0;", 1)
		}
		if err := os.WriteFile(filepath.Join(mutant, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	badEntry := filepath.Join(mutant, "comparison_main.a")
	badNode := run(t, root, nil, "node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), badEntry, input)
	badBinary := filepath.Join(mutant, "comparison")
	run(t, root, nil, "go", "run", "./cmd/adamic", "build", badEntry, "-o", badBinary, "--sanitize")
	badNative := run(t, root, nil, badBinary, input)
	badJS := run(t, root, nil, "node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), emitted(t, root, badEntry, mutant), input)
	for name, actual := range map[string][]byte{"Node": badNode, "native": badNative, "emitted JavaScript": badJS} {
		if bytes.Equal(actual, want) || !strings.Contains(string(actual), "shallower\t0\tok\n") {
			t.Fatalf("%s missing-dependency mutant was not caught by comparison", name)
		}
		t.Logf("%s: successful missing-dependency semantic mutant caught by byte comparison", name)
	}
}
func TestScopeReplayBlocked(t *testing.T) {
	t.Parallel()
	root, own := location(t)
	_, scope := oracle(t, root, own)
	want, err := os.ReadFile(scope)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join(own, "testdata/scope.graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	var graph string
	if err := json.Unmarshal(golden, &graph); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(graph), want) {
		t.Fatal("actual Go scope producer differs from blocker fixture")
	}
	entry := filepath.Join(own, "scope_replay_main.a")
	binary := filepath.Join(t.TempDir(), "scope")
	run(t, root, nil, "go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	js := emitted(t, root, entry, t.TempDir())
	for _, args := range [][]string{{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), entry, scope}, {binary, scope}, {"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), js, scope}} {
		command := exec.Command(args[0], args[1:]...)
		command.Dir = root
		output, err := command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "unknown terminal Scope") {
			t.Fatalf("shared Scope blocker changed: %v %s", err, output)
		}
	}
	t.Log("shared decoder declines the actual Go Scope terminal on Node, emitted JavaScript and sanitized native; no validation census cases certified")
}
