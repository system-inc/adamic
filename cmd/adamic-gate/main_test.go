package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Ask Go itself which tests ran. A regex-only check would miss Go's slash splitting rules. The probe
// (testdata/selectorprobe, a module of its own, built with GOWORK off) is a test binary built ahead, and
// test2json is too, so the unit runs Go's own test selection without running go.
func TestAnchoredSelectors(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	probe, test2json := selectorProbe(t)
	asked := []unit{{Package: "selectorprobe", Test: "TestParent/internal/oracle/testdata/a.a"}, {Package: "selectorprobe", Test: "TestParent/internal/oracle/testdata/b+.a"}, {Package: "selectorprobe", Test: "TestParent/internal/load/testdata/0.1/compile/main.a"}}
	var all []result
	for index, pattern := range patterns(asked) {
		log := filepath.Join(directory, "run-"+string(rune('a'+index))+".jsonl")
		f, err := os.Create(log)
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(test2json, "-t", "-p", "selectorprobe", probe, "-test.v=test2json", "-test.timeout=90s", "-test.count=1", "-test.run="+pattern)
		cmd.Dir = filepath.Dir(probe)
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
	sort.Strings(leaves)
	want := []string{"TestParent/internal/load/testdata/0.1/compile/main.a", "TestParent/internal/oracle/testdata/a.a", "TestParent/internal/oracle/testdata/b+.a"}
	if !reflect.DeepEqual(leaves, want) {
		t.Fatalf("Go executed %v, want exactly %v", leaves, want)
	}
}

// selectorProbe is testdata/selectorprobe's test binary and go's test2json, both products built ahead.
func selectorProbe(t *testing.T) (string, string) {
	t.Helper()
	probe := buildcache.GoTest(t, "cmd/adamic-gate/testdata/selectorprobe", "selectorprobe.test", ".", nil, "GOWORK=off")
	return probe, buildcache.GoBuild(t, "test2json", "cmd/test2json", nil)
}

func TestProduct_SelectorProbe(t *testing.T) {
	t.Parallel()
	selectorProbe(t)
}

func TestCoverageRejectsOverlapAndUnplannedTests(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
