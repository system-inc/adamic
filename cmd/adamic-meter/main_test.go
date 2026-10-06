package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFixtureCorpusCountsEveryDiagnosticKind(t *testing.T) {
	t.Parallel()
	r, err := measure("testdata/corpus")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, reason := range r.Reasons {
		got[reason.Kind] += reason.Count
	}
	if got["Refused"] != 1 || got["NotYet"] != 1 {
		t.Fatalf("counts = %#v, want one Refused and one NotYet", got)
	}
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
	if got.FilesExamined != 2 || len(got.Reasons) != 2 {
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
