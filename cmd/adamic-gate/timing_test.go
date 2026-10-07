package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func timingLog(t *testing.T, events ...event) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(f)
	for _, e := range events {
		if err := encoder.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestTimingsRetainRepeatedParentSetupOnce(t *testing.T) {
	paths := []string{
		timingLog(t, event{Action: "pass", Package: "p", Test: "TestParent/a", Elapsed: 2}, event{Action: "pass", Package: "p", Test: "TestParent", Elapsed: 12}),
		timingLog(t, event{Action: "pass", Package: "p", Test: "TestParent/b", Elapsed: 3}, event{Action: "pass", Package: "p", Test: "TestParent", Elapsed: 23}),
	}
	weights, audit, err := calibrateTimings(paths)
	if err != nil {
		t.Fatal(err)
	}
	if weights["p::TestParent"] != 20 || len(audit.Observations["p::TestParent"]) != 2 {
		t.Fatalf("lost repeated setup/children: %v", weights)
	}
	reverse, _, err := calibrateTimings([]string{paths[1], paths[0]})
	if err != nil || reverse["p::TestParent"] != 20 {
		t.Fatal("log order changed calibration", reverse, err)
	}
}
func TestTimingsPriceLayoutSetupOnce(t *testing.T) {
	g := knownAffinities()[1]
	paths := []string{}
	start := time.Date(2026, 10, 7, 15, 0, 0, 0, time.UTC)
	for _, name := range g.Tests {
		paths = append(paths, timingLog(t, event{Action: "cont", Package: g.Package, Test: name, Time: start}, event{Action: "output", Package: g.Package, Test: name, Time: start.Add(10 * time.Second), Output: "original fork full Markdown parsing/layout off agrees on all 5175 source documents"}, event{Action: "pass", Package: g.Package, Test: name, Elapsed: 12}))
	}
	weights, _, err := calibrateTimings(paths)
	if err != nil {
		t.Fatal(err)
	}
	if weights[g.key()] != 28 {
		t.Fatalf("fixture setup repeated: got %v want 28", weights[g.key()])
	}
}
func TestTimingsPreserveParallelParentSpan(t *testing.T) {
	path := timingLog(t, event{Action: "pass", Package: "p", Test: "TestParent/a", Elapsed: 5}, event{Action: "pass", Package: "p", Test: "TestParent/b", Elapsed: 5}, event{Action: "pass", Package: "p", Test: "TestParent", Elapsed: 6})
	weights, _, err := calibrateTimings([]string{path})
	if err != nil || weights["p::TestParent"] != 6 {
		t.Fatal(weights, err)
	}
}
func TestTimingsRejectMalformedLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.jsonl")
	if err := os.WriteFile(path, []byte("not json\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := calibrateTimings([]string{path}); err == nil {
		t.Fatal("accepted corrupt timing evidence")
	}
}

func TestTimingsMergedLogPreservesInvocations(t *testing.T) {
	path := timingLog(t,
		event{Action: "start", Package: "p"}, event{Action: "pass", Package: "p", Test: "TestParent/a", Elapsed: 2}, event{Action: "pass", Package: "p", Test: "TestParent", Elapsed: 12}, event{Action: "pass", Package: "p"},
		event{Action: "start", Package: "p"}, event{Action: "pass", Package: "p", Test: "TestParent/b", Elapsed: 3}, event{Action: "pass", Package: "p", Test: "TestParent", Elapsed: 23}, event{Action: "pass", Package: "p"})
	weights, audit, err := calibrateTimings([]string{path})
	if err != nil || weights["p::TestParent"] != 20 || len(audit.Observations["p::TestParent"]) != 2 {
		t.Fatal("merged invocations overwritten", weights, err)
	}
}

func TestTimingsReorderingIsByteIdentical(t *testing.T) {
	paths := []string{}
	for i := 1; i <= 15; i++ {
		paths = append(paths, timingLog(t, event{Action: "pass", Package: "p", Test: "TestParent/" + string(rune('a'+i)), Elapsed: float64(i) / 10}, event{Action: "pass", Package: "p", Test: "TestParent", Elapsed: 10 + float64(i)/10}))
	}
	first, _, err := calibrateTimings(paths)
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		paths = append(paths[1:], paths[0])
		weights, _, err := calibrateTimings(paths)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(weights)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatal("reordered evidence changed timing bytes", string(want), string(got))
		}
	}
}
