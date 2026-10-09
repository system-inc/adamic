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

// A test its owner split into shards declares the count; the planner names each shard a unit, and Go runs
// exactly the shards a pattern selects. A count that isn't a positive literal refuses, and widths grow past 999.
func TestShardChildren(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	files := map[string]string{
		"go.mod": "module shardprobe\n\ngo 1.27.0\n",
		"probe_test.go": `package shardprobe
import (
	"fmt"
	"testing"
)
const testWideShards = 12
const testHugeShards = 1001
const shardCount = 2
const testBadShards = shardCount
func TestWide(t *testing.T) {
	for index := 0; index < testWideShards; index++ {
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {})
	}
}
func TestWideExtra(t *testing.T) {}
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	wide, err := shardChildren(directory, "TestWide")
	if err != nil || len(wide) != 12 || wide[0] != "shard-000" || wide[11] != "shard-011" {
		t.Fatalf("TestWide's shards: %v, %v", wide, err)
	}
	if extra, err := shardChildren(directory, "TestWideExtra"); err != nil || extra != nil {
		t.Fatalf("an unsharded test got shards: %v, %v", extra, err)
	}
	if huge, err := shardChildren(directory, "TestHuge"); err != nil || len(huge) != 1001 || huge[0] != "shard-0000" || huge[1000] != "shard-1000" {
		t.Fatalf("1,001 shards: first %q, last %q, %v", huge[0], huge[len(huge)-1], err)
	}
	if _, err := shardChildren(directory, "TestBad"); err == nil || !strings.Contains(err.Error(), "positive integer literal") {
		t.Fatalf("a count that isn't a literal was accepted: %v", err)
	}
	asked := []unit{{Package: "shardprobe", Test: "TestWide/" + wide[3]}, {Package: "shardprobe", Test: "TestWide/" + wide[10]}}
	var leaves []string
	for _, pattern := range patterns(asked) {
		command := exec.Command("go", "test", "-count=1", "-v", "-run", pattern, ".")
		command.Dir = directory
		command.Env = append(os.Environ(), "GOWORK=off")
		out, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "--- PASS: TestWide/") {
				leaves = append(leaves, strings.Fields(line)[2])
			}
		}
	}
	if want := []string{"TestWide/shard-003", "TestWide/shard-010"}; !reflect.DeepEqual(leaves, want) {
		t.Fatalf("Go ran %v, want exactly %v", leaves, want)
	}
}
