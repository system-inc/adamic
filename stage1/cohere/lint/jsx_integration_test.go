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

// Not parallel: throughput measurements share CPU capacity and the same source corpus.
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
	// Captured JSX includes parser-recovery witnesses, just like TestRulesAgree.
	// Mark those rows before asking the oracle to compare or count their findings.
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

// Fixed top-level shards are visible to go test -list. ADAMIC_TEST_SHARD=i/n
// selects indices congruent to i modulo n; unset runs all. Each shard retains
// Go, Node, sanitized native, and leak checks on its entire assigned slice.
const testJsxLintTreesShards = 16

func TestJsxLintTrees_000(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 0) }
func TestJsxLintTrees_001(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 1) }
func TestJsxLintTrees_002(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 2) }
func TestJsxLintTrees_003(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 3) }
func TestJsxLintTrees_004(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 4) }
func TestJsxLintTrees_005(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 5) }
func TestJsxLintTrees_006(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 6) }
func TestJsxLintTrees_007(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 7) }
func TestJsxLintTrees_008(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 8) }
func TestJsxLintTrees_009(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 9) }
func TestJsxLintTrees_010(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 10) }
func TestJsxLintTrees_011(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 11) }
func TestJsxLintTrees_012(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 12) }
func TestJsxLintTrees_013(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 13) }
func TestJsxLintTrees_014(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 14) }
func TestJsxLintTrees_015(t *testing.T) { t.Parallel(); jsxRunTreeShard(t, 15) }
