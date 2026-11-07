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
func jsxSpansOracle(test *testing.T) string {
	test.Helper()
	root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	side, _ := filepath.Abs("testdata/jsx_spans.go")
	virtual := filepath.Join(root, "adamic_jsx_spans.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	directory := test.TempDir()
	path := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(path, overlay, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	binary := filepath.Join(directory, "jsx-spans")
	execute(test, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}

// jsxSources returns the captured upstream cases of every registry rule whose source holds JSX.
func jsxSources(test *testing.T) []string {
	test.Helper()
	spans := jsxSpansOracle(test)
	var paths []string
	byRule := map[string]int{}
	for _, row := range upstream(test) {
		fields := strings.Split(row, "\t")
		if len(bytes.TrimSpace(execute(test, "", spans, fields[0]).output)) > 0 {
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
		test.Fatalf("captured JSX cases by rule %v, want %v", byRule, want)
	}
	return paths
}

func buildNative(test *testing.T, entry string, sanitize bool) string {
	test.Helper()
	program, loadError := load.Load([]string{entry})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := filepath.Join(test.TempDir(), "native")
	if buildError := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); buildError != nil {
		test.Fatal(buildError)
	}
	return binary
}

// Not parallel: throughput measurements share CPU capacity and the same source corpus.
func TestJsxLintReleaseAndThroughput(test *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		test.Skip("set ADAMIC_LINT_BENCH=1 for JSX throughput")
	}
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var rows []string
	for _, path := range jsxSources(test) {
		rows = append(rows, path+"\tall")
	}
	oracle := lintGoOracleProduct(test)
	binary := buildPort(test, directory, false)
	// Captured JSX includes parser-recovery witnesses, just like TestRulesAgree.
	// Mark those rows before asking the oracle to compare or count their findings.
	input := manifest(test, recoveryRows(test, oracle, rows))
	compare(test, oracle, binary, directory, input)
	best := map[string]time.Duration{}
	var answer []byte
	names := []string{"Go", "native", "Node"}
	for round := 0; round < 5; round++ {
		for offset := 0; offset < len(names); offset++ {
			name := names[(round+offset)%len(names)]
			var result execution
			switch name {
			case "Go":
				result = execute(test, "", oracle, "--manifest", input, "--count")
			case "native":
				result = execute(test, "", binary, "--manifest", input, "--count")
			case "Node":
				result = node(test, directory, input, true)
			}
			if answer == nil {
				answer = result.output
			}
			if !bytes.Equal(answer, result.output) {
				test.Fatalf("JSX finding count differs on %s", name)
			}
			if best[name] == 0 || result.duration < best[name] {
				best[name] = result.duration
			}
			test.Logf("round %d %s %s findings=%s", round+1, name, result.duration, strings.TrimSpace(string(answer)))
		}
	}
	var count int
	if _, scanError := fmt.Sscan(string(answer), &count); scanError != nil || count == 0 {
		test.Fatalf("positive count missing: %s %v", answer, scanError)
	}
	for _, name := range names {
		test.Logf(
			"best of 5 JSX %s: %.6fs, %.2f findings/s (%d files, %d findings)",
			name,
			best[name].Seconds(),
			float64(count)/best[name].Seconds(),
			len(rows),
			count,
		)
	}
}

// Fixed top-level shards are visible to go test -list. ADAMIC_TEST_SHARD=i/n
// selects indices congruent to i modulo n; unset runs all. Each shard retains
// Go, Node, sanitized native, and leak checks on its entire assigned slice.
const testJsxLintTreesShards = 16

func TestJsxLintTrees_000(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 0) }
func TestJsxLintTrees_001(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 1) }
func TestJsxLintTrees_002(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 2) }
func TestJsxLintTrees_003(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 3) }
func TestJsxLintTrees_004(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 4) }
func TestJsxLintTrees_005(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 5) }
func TestJsxLintTrees_006(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 6) }
func TestJsxLintTrees_007(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 7) }
func TestJsxLintTrees_008(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 8) }
func TestJsxLintTrees_009(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 9) }
func TestJsxLintTrees_010(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 10) }
func TestJsxLintTrees_011(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 11) }
func TestJsxLintTrees_012(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 12) }
func TestJsxLintTrees_013(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 13) }
func TestJsxLintTrees_014(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 14) }
func TestJsxLintTrees_015(test *testing.T) { test.Parallel(); jsxRunTreeShard(test, 15) }
