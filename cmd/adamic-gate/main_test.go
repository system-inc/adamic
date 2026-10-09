package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Ask Go itself which tests ran. A regex-only check would miss Go's slash splitting rules.
func TestAnchoredSelectors(t *testing.T) {
	directory := t.TempDir()
	files := map[string]string{
		"go.mod": "module selectorprobe\n\ngo 1.27.0\n",
		"probe_test.go": `package selectorprobe
import "testing"
func TestParent(t *testing.T) {
 for _,name:=range []string{"internal/oracle/testdata/a.a", "internal/oracle/testdata/aXa", "internal/oracle/testdata/b+.a", "internal/load/testdata/0.1/compile/main.a", "internal/load/testdata/0X1/compile/extra.a"} {
  t.Run(name,func(t *testing.T){})
 }
}
func TestParentExtra(t *testing.T) {}
func TestOther(t *testing.T) {}
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	asked := []unit{{Package: "selectorprobe", Test: "TestParent/internal/oracle/testdata/a.a"}, {Package: "selectorprobe", Test: "TestParent/internal/oracle/testdata/b+.a"}, {Package: "selectorprobe", Test: "TestParent/internal/load/testdata/0.1/compile/main.a"}}
	var all []result
	for index, pattern := range patterns(asked) {
		log := filepath.Join(directory, "run-"+string(rune('a'+index))+".jsonl")
		f, err := os.Create(log)
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("go", "test", "-count=1", "-json", "-run", pattern, ".")
		cmd.Dir = directory
		cmd.Env = append(os.Environ(), "GOWORK=off")
		cmd.Stdout = f
		cmd.Stderr = f
		err = cmd.Run()
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		events, _, _, err := readLog(log)
		if err != nil {
			t.Fatal(err)
		}
		all = append(all, events...)
	}
	var leaves []string
	for _, r := range all {
		if r.Test != "" && r.Test != "TestParent" {
			leaves = append(leaves, r.Test)
		}
	}
	want := []string{"TestParent/internal/load/testdata/0.1/compile/main.a", "TestParent/internal/oracle/testdata/a.a", "TestParent/internal/oracle/testdata/b+.a"}
	if !reflect.DeepEqual(leaves, want) {
		t.Fatalf("Go executed %v, want exactly %v", leaves, want)
	}
}

func TestCoverageRejectsOverlapAndUnplannedTests(t *testing.T) {
	p := plan{Count: 2, Units: []unit{{Package: "p", Test: "TestParent/one", Shard: 0}, {Package: "p", Test: "TestParent/two", Shard: 1}, {Package: "p", Test: "TestWhole", Shard: 0}}}
	produced := map[string]bool{}
	seen := map[string]int{}
	first := []result{{Package: "p", Test: "TestParent", Action: "pass"}, {Package: "p", Test: "TestParent/one", Action: "pass"}, {Package: "p", Test: "TestWhole", Action: "pass"}, {Package: "p", Test: "TestWhole/child", Action: "pass"}}
	second := []result{{Package: "p", Test: "TestParent", Action: "pass"}, {Package: "p", Test: "TestParent/two", Action: "pass"}}
	if problems := validateResults(p, 0, first, produced, seen); len(problems) > 0 {
		t.Fatal(problems)
	}
	if problems := validateResults(p, 1, second, produced, seen); len(problems) > 0 {
		t.Fatal(problems)
	}
	for _, u := range p.Units {
		if !produced[u.key()] {
			t.Fatalf("missing evidence for %s", u.key())
		}
	}
	overlap := validateResults(p, 0, []result{first[1]}, produced, seen)
	if len(overlap) != 1 || !strings.Contains(strings.Join(overlap, "\n"), "ran twice") {
		t.Fatalf("overlap accepted: %v", overlap)
	}
	unknown := validateResults(p, 0, []result{{Package: "p", Test: "TestUnexpected", Action: "pass"}}, produced, seen)
	if len(unknown) != 1 || !strings.Contains(unknown[0], "unplanned") {
		t.Fatal(unknown)
	}
	missing := map[string]bool{}
	validateResults(p, 0, []result{first[0]}, missing, map[string]int{})
	if missing[p.Units[0].key()] {
		t.Fatal("an ancestor satisfied missing child coverage")
	}
}

func TestRawEvidenceKeepsSkipReasonsAndFailures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.jsonl")
	log := `{"Action":"output","Package":"p","Test":"TestSkip","Output":"    probe_test.go:7: needs external corpus\n"}
{"Action":"skip","Package":"p","Test":"TestSkip","Elapsed":0.1}
{"Action":"output","Package":"p","Test":"TestFail","Output":"wrong byte 7\n"}
{"Action":"fail","Package":"p","Test":"TestFail","Elapsed":0.2}
{"Action":"output","Package":"p","Output":"gate cache: native hits=1 misses=0\n"}
`
	if err := os.WriteFile(path, []byte(log), 0600); err != nil {
		t.Fatal(err)
	}
	results, cache, count, err := readLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || len(results) != 2 || !strings.Contains(results[0].Reason, "needs external corpus") || results[1].Output != "wrong byte 7\n" {
		t.Fatalf("lost raw evidence: %v", results)
	}
	if len(cache) != 1 || cacheHits.FindStringSubmatch(cache[0])[2] != "1" {
		t.Fatal("lost hit evidence")
	}
	if err := os.WriteFile(path, []byte("truncated JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := readLog(path); err == nil {
		t.Fatal("corrupt log accepted")
	}
}

// Not parallel: t.Chdir moves the process to the repository root, where children reads internal/oracle.
func TestNativeOracleGateShardCoverage(t *testing.T) {
	t.Chdir("../..")
	pkg := "github.com/system-inc/adamic/internal/oracle"
	names, err := children(pkg, "TestNativeAgreesWithNode")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 34 {
		t.Fatalf("discovered %d shard names, want 34", len(names))
	}
	expected := plan{Count: 1}
	for i, name := range names {
		want := []string{"shard-000", "shard-001", "shard-002", "shard-003-0", "shard-003-1", "shard-004", "shard-005", "shard-006", "shard-007", "shard-008", "shard-009", "shard-010", "shard-011", "shard-012", "shard-013", "shard-014", "shard-015", "shard-016", "shard-017", "shard-018", "shard-019-0", "shard-019-1", "shard-020", "shard-021", "shard-022", "shard-023", "shard-024", "shard-025", "shard-026", "shard-027", "shard-028", "shard-029", "shard-030", "shard-031"}[i]
		if name != want {
			t.Fatalf("child %d: %q, want %q", i, name, want)
		}
		expected.Units = append(expected.Units, unit{Package: pkg, Test: "TestNativeAgreesWithNode/" + name, Shard: 0})
	}
	// A parent PASS plus all other shards cannot green the shard that did not run.
	records := []result{{Package: pkg, Test: "TestNativeAgreesWithNode", Action: "pass"}}
	for _, u := range expected.Units[1:] {
		records = append(records, result{Package: pkg, Test: u.Test, Action: "pass"})
	}
	produced := map[string]bool{}
	if problems := validateResults(expected, 0, records, produced, map[string]int{}); len(problems) != 0 {
		t.Fatal(problems)
	}
	missing := []string{}
	for _, u := range expected.Units {
		if !produced[u.key()] {
			missing = append(missing, u.Test)
		}
	}
	if len(missing) != 1 || missing[0] != "TestNativeAgreesWithNode/shard-000" {
		t.Fatalf("missing shard accepted: %v", missing)
	}
	t.Logf("missing evidence: %s; parent pass cannot satisfy it", missing[0])
}
