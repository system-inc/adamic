package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

// Not parallel: these startup and PATH controls change process environment.
func TestTypescriptRequiresPinnedSource(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv("ADAMIC_TYPESCRIPT_SOURCE", "")
		oracle, err := startTypescript("../..", t.TempDir())
		if oracle != nil {
			oracle.close()
			t.Fatal("unset source started an oracle")
		}
		if err == nil || !strings.Contains(err.Error(), "ADAMIC_TYPESCRIPT_SOURCE") || !strings.Contains(err.Error(), "6.0.3") {
			t.Fatalf("want clear missing pin error: %v", err)
		}
	})
	t.Run("wrong version", func(t *testing.T) {
		module, err := filepath.Abs(filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "lib/typescript.js"))
		if err != nil {
			t.Fatal(err)
		}
		wrong := fakeTypescript(t, module)
		// An old API may lack the enums used by the helper: version refusal
		// must happen before touching those APIs.
		if err := os.WriteFile(filepath.Join(wrong, "lib/typescript.js"), []byte("module.exports = {version: '5.9.3'};\n"), 0644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("ADAMIC_TYPESCRIPT_SOURCE", wrong)
		oracle, err := startTypescript("../..", t.TempDir())
		if oracle != nil {
			oracle.close()
			t.Fatal("wrong version started an oracle")
		}
		if err == nil || !strings.Contains(err.Error(), "6.0.3") || !strings.Contains(err.Error(), "5.9.3") || !strings.Contains(err.Error(), "ADAMIC_TYPESCRIPT_SOURCE") {
			t.Fatalf("want clear version pin error: %v", err)
		}
	})
}

func fakeTypescript(t *testing.T, actual string) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{"bin", "lib"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0755); err != nil {
			t.Fatal(err)
		}
	}
	module, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	// A different stock API is available beside the shadowing executable. Its
	// version makes accidental selection visible before any diagnostic is used.
	if err := os.WriteFile(filepath.Join(root, "lib/typescript.js"), []byte("module.exports = {...require("+string(module)+"), version: '5.9.3'};\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin/tsc"), []byte("#!/bin/sh\nprintf 'Version 5.9.3\\n'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	return root
}

// Not parallel: deliberately shadows tsc on PATH for the mutant control.
func TestTypescriptIgnoresPathTsc(t *testing.T) {
	baseline, err := startTypescript("../..", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer baseline.close()
	const source = "const value: number = 'wrong';"
	want, err := baseline.check(source)
	if err != nil || !reflect.DeepEqual(want, []string{"TS2322"}) {
		t.Fatalf("baseline: %v %v", want, err)
	}
	module, err := filepath.Abs(filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "lib/typescript.js"))
	if err != nil {
		t.Fatal(err)
	}
	wrong := fakeTypescript(t, module)
	t.Setenv("PATH", filepath.Join(wrong, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	path, err := exec.LookPath("tsc")
	if err != nil || path != filepath.Join(wrong, "bin/tsc") {
		t.Fatalf("mutant did not shadow tsc: %s %v", path, err)
	}
	mutated, err := startTypescript("../..", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer mutated.close()
	got, err := mutated.check(source)
	if err != nil || mutated.stats.Version != baseline.stats.Version || !reflect.DeepEqual(got, want) {
		t.Fatalf("PATH changed pinned oracle: version %s codes %v error %v", mutated.stats.Version, got, err)
	}
	t.Logf("shadowed tsc 5.9.3; pinned oracle remains %s with %v", mutated.stats.Version, got)
}
