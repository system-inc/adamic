package wave10

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func run(t *testing.T, directory, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = directory
	out, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	cmd.Stdout = out
	stderr, err := os.CreateTemp(t.TempDir(), "stderr-")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	cmd.Stderr = stderr
	err = cmd.Run()
	diagnostics, _ := os.ReadFile(stderr.Name())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("%s %v: %v stderr=%s", name, args, err, diagnostics)
	}
	data, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
func absolute(t *testing.T, path string) string {
	t.Helper()
	result, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func oracle(t *testing.T) string {
	t.Helper()
	root := absolute(t, "../../../../../cohere")
	scratch := t.TempDir()
	main := filepath.Join(root, "adamic_math_whitespace.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{main: absolute(t, "testdata/oracle.go.txt"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_math_whitespace.go"): absolute(t, "testdata/exports.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(scratch, "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(scratch, "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, main)
	return binary
}
func compile(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	binary := filepath.Join(scratch, "native")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(scratch, "emitted.mjs")
	write(t, module, []byte(javascript.JavaScript(ir)))
	return binary, module
}
func commands(t *testing.T, entry, path string) [][]string {
	t.Helper()
	binary, module := compile(t, entry)
	runner := absolute(t, "../../../../../oracle/node.mjs")
	return [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, entry, path}, {"node", "--disable-warning=ExperimentalWarning", runner, module, path}, {binary, path}}
}
func checkCoverage(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile("testdata/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Rule, File string
		Literals   int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	ledger, err := os.ReadFile("../readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	var readiness struct {
		Remaining []struct {
			Rule             string
			RemainingHelpers []string `json:"remaining_helpers"`
		}
	}
	if err = json.Unmarshal(ledger, &readiness); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.addWhitespaceAroundMathOperators"
	for _, rule := range readiness.Remaining {
		for _, helper := range rule.RemainingHelpers {
			if helper == symbol {
				want[rule.Rule] = true
			}
		}
	}
	for _, row := range rows {
		if !want[row.Rule] || row.Literals == 0 {
			t.Fatalf("unexpected or empty consumer %+v", row)
		}
		delete(want, row.Rule)
	}
	if len(want) != 0 {
		t.Fatalf("missing consumers %v", want)
	}
}

// Not parallel: bounded parser/clang memory and deterministic mutation evidence.
func TestMathWhitespace(t *testing.T) {
	checkCoverage(t)
	path := absolute(t, "testdata/cases.json")
	want := run(t, "", oracle(t), path)
	entry := absolute(t, "main.a")
	for i, cmd := range commands(t, entry, path) {
		got := run(t, "", cmd[0], cmd[1:]...)
		if !bytes.Equal(got, want) {
			t.Fatalf("runtime %d differs: got %d bytes Go %d", i, len(got), len(want))
		}
		t.Logf("runtime %d matched %d bytes", i, len(got))
	}
	t.Log("all six consumers covered; 49162 scanner inputs")
}
func TestMathWhitespaceMutant(t *testing.T) {
	checkCoverage(t)
	path := absolute(t, "testdata/cases.json")
	want := run(t, "", oracle(t), path)
	directory := t.TempDir()
	main, err := os.ReadFile("main.a")
	if err != nil {
		t.Fatal(err)
	}
	options := absolute(t, "../options_json.ts")
	main = []byte(strings.ReplaceAll(string(main), "../options_json.ts", filepath.ToSlash(options)))
	write(t, filepath.Join(directory, "main.a"), main)
	source, err := os.ReadFile("math_operator_whitespace.a")
	if err != nil {
		t.Fatal(err)
	}
	old := "else if(character === ',' && active) { result += ', '; }"
	replacement := "else if(character === ',' && active) { result += ','; }"
	if strings.Count(string(source), old) != 1 {
		t.Fatal("mutant anchor changed")
	}
	write(t, filepath.Join(directory, "math_operator_whitespace.a"), []byte(strings.Replace(string(source), old, replacement, 1)))
	for i, cmd := range commands(t, filepath.Join(directory, "main.a"), path) {
		got := run(t, "", cmd[0], cmd[1:]...)
		if bytes.Equal(got, want) {
			t.Fatalf("runtime %d mutant survived", i)
		}
		t.Logf("runtime %d compiling comma_spacing_disabled mutant caught only by output comparison", i)
	}
}
