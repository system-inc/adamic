package json

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// An explicit repository-relative key keeps ownership stable as the corpus grows.
// Agreement belongs to _000; _001 adds only the independent leak pass.
const jsonPortClassOrderKey = "cohere/internal/lint/rules/tailwind/collapse/testdata/classorder_fixtures.json"

func jsonPortClassOrderPartition(all []textCase) (ordinary []textCase, shards []nativeChunk, dedicated []textCase) {
	for _, item := range all {
		if item.Name == jsonPortClassOrderKey {
			dedicated = append(dedicated, item)
		} else {
			ordinary = append(ordinary, item)
		}
	}
	shards = jsonHashShards(ordinary, testPortMatchesGoCohereShards)
	return
}

func jsonPortClassOrderUnion(all, ordinary []textCase, shards []nativeChunk, dedicated []textCase) error {
	if len(dedicated) > 1 {
		return fmt.Errorf("duplicate dedicated class-order case")
	}
	combined := append(append([]textCase(nil), ordinary...), dedicated...)
	owners := append([]nativeChunk(nil), shards...)
	owners = append(owners, nativeChunk{start: len(ordinary), end: len(combined)})
	if err := jsonPortUnion(combined, owners); err != nil {
		return err
	}
	if len(combined) != len(all) {
		return fmt.Errorf("dedicated union has %d of %d cases", len(combined), len(all))
	}
	expected := make(map[string]textCase, len(all))
	for _, item := range all {
		if _, exists := expected[item.Name]; exists {
			return fmt.Errorf("duplicate input %s", item.Name)
		}
		expected[item.Name] = item
	}
	for i, item := range combined {
		if want, exists := expected[item.Name]; !exists || want != item {
			return fmt.Errorf("dedicated union changed %s", item.Name)
		}
		if i < len(ordinary) && item.Name == jsonPortClassOrderKey {
			return fmt.Errorf("dedicated case retained in ordinary shard")
		}
		if i >= len(ordinary) && item.Name != jsonPortClassOrderKey {
			return fmt.Errorf("unexpected dedicated case %s", item.Name)
		}
	}
	return nil
}

func jsonPortClassOrderRun(t *testing.T, leak bool) {
	t.Helper()
	setup := time.Now()
	portMatchesPrepare(t)
	all, _ := jsonTopCorpus(t)
	ordinary, shards, items := jsonPortClassOrderPartition(all)
	if err := jsonPortClassOrderUnion(all, ordinary, shards, items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("dedicated fixture count %d, want 1", len(items))
	}
	t.Logf("shared setup: %.3fs; exact union: %d ordinary + %d dedicated", time.Since(setup).Seconds(), len(ordinary), len(items))
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	start := time.Now()
	input, _ := protocol(items, make([]answer, len(items)))
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	reference := execute(t, nil, portMatchesShared.oracle, "--cases", path)
	if reference.exitCode != 0 || len(reference.stderr) != 0 {
		t.Fatalf("Go oracle: exit %d stderr %s", reference.exitCode, reference.stderr)
	}
	expected := string(reference.stdout)
	if _, err := jsonPortAnswers(expected, len(items)); err != nil {
		t.Fatal(err)
	}
	compare(t, "release", execute(t, nil, portMatchesShared.release, "--cases", path), expected, items)
	if !leak {
		compare(t, "Node", onNode(t, portMatchesShared.entry, "--cases", path), expected, items)
		compare(t, "JavaScript backend", onNode(t, portMatchesShared.script, "--cases", path), expected, items)
	}
	mode := "ASan/UBSan"
	options := "ASAN_OPTIONS=detect_leaks=0"
	if leak {
		mode = "LeakSanitizer"
		options = "ASAN_OPTIONS=detect_leaks=1"
	}
	if leak && runtime.GOOS != "linux" {
		result := execute(t, nil, "leaks", "--atExit", "--", portMatchesShared.release, "--cases", path)
		if result.exitCode != 0 {
			t.Fatalf("leaks: %s", result.stderr)
		}
	} else {
		began := time.Now()
		compare(t, mode, execute(t, []string{options}, portMatchesShared.sanitized, "--cases", path), expected, items)
		t.Logf("%s: %.3fs", mode, time.Since(began).Seconds())
	}
	t.Logf("%s: own work %.3fs", jsonPortClassOrderKey, time.Since(start).Seconds())
}

func TestPortMatchesGoCohereClassOrder_000(t *testing.T) {
	t.Parallel()
	jsonPortClassOrderRun(t, false)
}
func TestPortMatchesGoCohereClassOrder_001(t *testing.T) {
	t.Parallel()
	jsonPortClassOrderRun(t, true)
}

func TestJSONClassOrderShardOwnership(t *testing.T) {
	t.Parallel()
	all := []textCase{{Name: "small.json", Text: "{}"}, {Name: jsonPortClassOrderKey, Text: "[]"}}
	ordinary, shards, dedicated := jsonPortClassOrderPartition(all)
	if err := jsonPortClassOrderUnion(all, ordinary, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	if len(dedicated) != 1 || len(ordinary) != 1 {
		t.Fatal("wrong dedicated ownership")
	}
	owner := jsonCaseShard("small.json", testPortMatchesGoCohereShards)
	if shards[owner].end-shards[owner].start != 1 || ordinary[shards[owner].start].Name != "small.json" {
		t.Fatal("ordinary hash owner changed")
	}
	if err := jsonPortClassOrderUnion(all, ordinary, shards, nil); err == nil {
		t.Fatal("missing dedicated case survived")
	}
	if err := jsonPortClassOrderUnion(all, ordinary, shards, append(dedicated, dedicated...)); err == nil {
		t.Fatal("duplicate dedicated case survived")
	}
	// Agreement ownership includes the dedicated ASan shard exactly once. The
	// second shard repeats its Go/release controls solely for the leak pass.
	owners := [][]textCase{}
	for _, shard := range shards {
		owners = append(owners, ordinary[shard.start:shard.end])
	}
	owners = append(owners, dedicated)
	caught := 0
	for _, items := range owners {
		answers := make([]answer, len(items))
		_, want := protocol(items, answers)
		for i, item := range items {
			if item.Name == jsonPortClassOrderKey {
				answers[i].Output = "planted disagreement"
			}
		}
		_, got := protocol(items, answers)
		if comparisonError("planted", run{stdout: []byte(got)}, want, items) != nil {
			caught++
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d owners", caught)
	}
	tree, err := parser.ParseFile(token.NewFileSet(), "class_order_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]bool{}
	for _, decl := range tree.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestPortMatchesGoCohereClassOrder_") {
			declared[f.Name.Name] = true
		}
	}
	if len(declared) != 2 || !declared["TestPortMatchesGoCohereClassOrder_000"] || !declared["TestPortMatchesGoCohereClassOrder_001"] {
		t.Fatalf("dedicated shard declarations: %v", declared)
	}
	t.Log("dedicated union intact; planted disagreement caught by exactly one agreement shard")
}
