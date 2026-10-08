package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// jsxSpansOracle builds testdata/jsx_spans.go inside cohere, which says whether a file holds JSX.
func jsxSpansOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	side, _ := filepath.Abs("testdata/jsx_spans.go")
	virtual := filepath.Join(root, "adamic_jsx_spans.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	directory := t.TempDir()
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "jsx-spans")
	execute(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}

// jsxSources returns the captured upstream cases of every registry rule whose source holds JSX.
func jsxSources(t *testing.T) []string {
	t.Helper()
	spans := jsxSpansOracle(t)
	var paths []string
	byRule := map[string]int{}
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(bytes.TrimSpace(execute(t, "", spans, fields[0]).output)) > 0 {
			paths = append(paths, fields[0])
			byRule[fields[1]]++
		}
	}
	// Each rule's captured JSX cases, exactly. A rule losing some, or a capture losing a rule, fails here
	// rather than shrinking the parser check silently, and a new rule that brings JSX cases adds its row.
	// Batch 8's three rules held the original 54.
	want := map[string]int{
		"@next/next/no-sync-scripts":                  11,
		"react/jsx-no-comment-textnodes":              40,
		"react/no-find-dom-node":                      9,
		"react/no-is-mounted":                         5,
		"nexus/consistency-no-abbreviated-identifier": 5,
		"nexus/consistency-no-ambiguous-identifier":   4,
	}
	if !reflect.DeepEqual(byRule, want) {
		t.Fatalf("captured JSX cases by rule %v, want %v", byRule, want)
	}
	return paths
}

func buildNative(t *testing.T, entry string, sanitize bool) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	return binary
}

// Not parallel: fresh-process throughput is measured after correctness on the same sources.
func TestJsxLintReleaseAndThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		t.Skip("set ADAMIC_LINT_BENCH=1 for JSX throughput")
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	for _, path := range jsxSources(t) {
		rows = append(rows, path+"\tall")
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, false)
	input := manifest(t, recoveryRows(t, oracle, rows))
	compare(t, oracle, binary, directory, input)
	best := map[string]time.Duration{}
	var answer []byte
	names := []string{"Go", "native", "Node"}
	for round := 0; round < 5; round++ {
		for offset := 0; offset < len(names); offset++ {
			name := names[(round+offset)%len(names)]
			var result execution
			switch name {
			case "Go":
				result = execute(t, "", oracle, "--manifest", input, "--count")
			case "native":
				result = execute(t, "", binary, "--manifest", input, "--count")
			case "Node":
				result = node(t, directory, input, true)
			}
			if answer == nil {
				answer = result.output
			}
			if !bytes.Equal(answer, result.output) {
				t.Fatalf("JSX finding count differs on %s", name)
			}
			if best[name] == 0 || result.duration < best[name] {
				best[name] = result.duration
			}
			t.Logf("round %d %s %s findings=%s", round+1, name, result.duration, strings.TrimSpace(string(answer)))
		}
	}
	var count int
	if _, err := fmt.Sscan(string(answer), &count); err != nil || count == 0 {
		t.Fatalf("positive count missing: %s %v", answer, err)
	}
	for _, name := range names {
		t.Logf("best of 5 JSX %s: %.6fs, %.2f findings/s (%d files, %d findings)", name, best[name].Seconds(), float64(count)/best[name].Seconds(), len(rows), count)
	}
}

// Not parallel: this replays all JSX fixture trees before native throughput.
func TestJsxLintTrees(t *testing.T) {
	paths := jsxSources(t)
	root, _ := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
	side, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go"))
	virtual := filepath.Join(root, "adamic_jsx_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(t.TempDir(), "parser-oracle")
	execute(t, root, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	path := manifest(t, paths)
	want := execute(t, "", oracle, "--manifest", path, "--whole", "--jsx-recovery")
	directory, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser"))
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	node := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--whole")
	if diff := difference(node.output, want.output); diff != "" {
		t.Fatal(diff)
	}
	binary := buildNative(t, filepath.Join(directory, "main.ts"), true)
	got := execute(t, "", binary, "--manifest", path, "--whole")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("%d captured cohere JSX sources: %d identical whole-tree bytes", len(paths), len(want.output))
}
