package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestFixtureCorpusCountsEveryDiagnosticKind(test *testing.T) {
	test.Parallel()
	r, err := measure("testdata/corpus", false)
	if err != nil {
		test.Fatal(err)
	}
	got := map[string]int{}
	for _, reason := range r.Reasons {
		got[reason.Kind] += reason.Count
	}
	if r.FilesReachingLowering != 2 {
		test.Fatalf("entries reaching lowering = %d, want 2", r.FilesReachingLowering)
	}
	if got["Refused"] != 1 || got["NotYet"] != 1 {
		test.Fatalf("counts = %#v, want one Refused and one NotYet", got)
	}
}

func TestAdaptRewritesTypeOnlyImportInMemory(test *testing.T) {
	test.Parallel()
	const path = "testdata/adapt/main.ts"
	before, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	plain, err := measure("testdata/adapt", false)
	if err != nil {
		test.Fatal(err)
	}
	adapted, err := measure("testdata/adapt", true)
	if err != nil {
		test.Fatal(err)
	}
	if countKind(plain, "mechanical") != 1 {
		test.Fatalf("plain mechanical diagnostics = %d, want 1", countKind(plain, "mechanical"))
	}
	if countKind(adapted, "mechanical") != 0 || len(adapted.Adaptations) != 1 || adapted.Adaptations[0].Removed != 1 {
		test.Fatalf("adapted report = %#v", adapted)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		test.Fatal("--adapt changed the fixture on disk")
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

func TestJSONIsMachineReadable(test *testing.T) {
	test.Parallel()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--json", "testdata/corpus"}, &stdout, &stderr); code != 0 {
		test.Fatalf("run = %d, stderr = %s", code, stderr.String())
	}
	var got report
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		test.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if got.FilesExamined != 2 || len(got.Reasons) != 2 {
		test.Fatalf("report = %#v", got)
	}
}

func TestNormalizeDropsIncidentalDetails(test *testing.T) {
	test.Parallel()
	got := normalize("/tmp/place.ts:4:2: a value of type Dog seen as Cat (detail)")
	if got != "<location>: a value of type <type>" {
		test.Fatalf("normalize = %q", got)
	}
}

// The former NotYet corpus witness is admitted by optional boolean field lowering.
func TestUnionClassFieldHasNoDiagnostic(test *testing.T) {
	test.Parallel()
	r, err := measure("testdata/compiled", false)
	if err != nil {
		test.Fatal(err)
	}
	if r.FilesExamined != 1 || r.FilesReachingLowering != 1 || len(r.Reasons) != 0 {
		test.Fatalf("compiled union field report = %#v", r)
	}
}
