package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestFixtureCorpusCountsEveryDiagnosticKind(t *testing.T) {
	t.Parallel()
	r, err := measure("testdata/corpus", false)
	if err != nil {
		t.Fatal(err)
	}
	// Optional fields now have a named representation gap. Keep that witness
	// and the async witness; non_null.a supplies the soundness refusal.
	got := map[string]int{}
	for _, reason := range r.Reasons {
		got[reason.Kind] += reason.Count
	}
	if r.FilesReachingLowering != 3 {
		t.Fatalf("entries reaching lowering = %d, want 3", r.FilesReachingLowering)
	}
	if got["Refused"] != 1 || got["NotYet"] != 2 {
		t.Fatalf("counts = %#v, want one Refused and two NotYet", got)
	}
}

func TestAdaptRewritesTypeOnlyImportInMemory(t *testing.T) {
	t.Parallel()
	const path = "testdata/adapt/main.ts"
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := measure("testdata/adapt", false)
	if err != nil {
		t.Fatal(err)
	}
	adapted, err := measure("testdata/adapt", true)
	if err != nil {
		t.Fatal(err)
	}
	if countKind(plain, "mechanical") != 1 {
		t.Fatalf("plain mechanical diagnostics = %d, want 1", countKind(plain, "mechanical"))
	}
	if countKind(adapted, "mechanical") != 0 || len(adapted.Adaptations) != 1 || adapted.Adaptations[0].Removed != 1 {
		t.Fatalf("adapted report = %#v", adapted)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("--adapt changed the fixture on disk")
	}
}

func countKind(r *report, kind string) int {
	count := 0
	for _, reason := range r.Reasons {
		if reason.Kind == kind {
			count += reason.Count
		}
	}
	return count
}

func TestJSONIsMachineReadable(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--json", "testdata/corpus"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run = %d, stderr = %s", code, stderr.String())
	}
	var got report
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if got.FilesExamined != 3 || len(got.Reasons) != 3 {
		t.Fatalf("report = %#v", got)
	}
}

func TestNormalizeDropsIncidentalDetails(t *testing.T) {
	t.Parallel()
	got := normalize("/tmp/place.ts:4:2: a value of type Dog seen as Cat (detail)")
	if got != "<location>: a value of type <type>" {
		t.Fatalf("normalize = %q", got)
	}
}
