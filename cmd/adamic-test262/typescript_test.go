package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These controls run the real Adamic checker and stock tsc on the same bytes.
// RegExp split has a stricter result type in Adamic; stock tsc only sees the
// missing name, so the disagreement is TS2322 versus TS2304.
func TestTypescriptControls(t *testing.T) {
	work := t.TempDir()
	oracle, err := startTypescript("../..", work)
	if err != nil {
		t.Fatal(err)
	}
	defer oracle.close()
	binary := filepath.Join(t.TempDir(), "adamic")
	build := exec.Command("go", "build", "-o", binary, "./cmd/adamic")
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	cases := []struct {
		name     string
		kind     outcomeKind
		contains string
	}{
		{"argument", outcomeNotTypescript, "TS2345"},
		{"missing", outcomeRefused, "not yet"},
		{"disagree", outcomeRefused, "tsc: TS2304"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("testdata/typescript", test.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			codes, err := oracle.check(string(source))
			if err != nil {
				t.Fatal(err)
			}
			engine := &engine{adamic: binary, work: work, oracle: oracle}
			got := engine.attempt(classified{Path: test.name, Program: string(source)})
			if got.Kind != test.kind || !strings.Contains(got.Reason, test.contains) {
				t.Fatalf("got %+v, want %s containing %s (tsc %v)", got, test.kind, test.contains, codes)
			}
			if test.name == "missing" && len(codes) != 0 {
				t.Fatalf("tsc must accept missing feature: %v", codes)
			}
			again, err := oracle.check(string(source))
			if err != nil || strings.Join(again, ",") != strings.Join(codes, ",") {
				t.Fatalf("cache: %v %v", again, err)
			}
		})
	}
	if oracle.stats.Checks != 3 || oracle.stats.Hits != 5 {
		t.Fatalf("cache stats: %+v", oracle.stats)
	}
}

func TestNotTypescriptTable(t *testing.T) {
	report := filterReport{Path: "built-ins/Date"}
	report.add(result{Directory: "built-ins/Date", Kind: outcomeNotTypescript, Reason: "TS2345"})
	report.add(result{Directory: "built-ins/Date", Kind: outcomeRefused, Reason: "not yet"})
	report.finish()
	if report.NotTypescript != 1 || report.Refused != 1 || report.Total != 2 || report.Directories[0].NotTypescript != 1 || report.Directories[0].Refused != 1 || len(report.RefusalReasons) != 1 || len(report.NotTypescriptReasons) != 1 || report.NotTypescriptReasons[0].Reason != "TS2345" {
		t.Fatalf("separate totals: %+v", report)
	}
	var output bytes.Buffer
	printTables(&output, reportDocument{Filters: []filterReport{report}})
	if !strings.Contains(output.String(), "refused 1   not-typescript 1") || !strings.Contains(output.String(), "not-typescript") {
		t.Fatalf("table: %s", output.String())
	}
}
