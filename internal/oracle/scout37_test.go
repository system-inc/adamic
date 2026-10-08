package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These reductions exercise existing task rules. A source mutant must still
// compile and agree with its own Node execution; only comparison with the
// pristine Node witness rejects its wrong list, intern pool, or partition.
func TestScout37ShapesAndMutants(t *testing.T) {
	cases := []struct {
		name, before, after string
	}{
		{"scout37_parse_list.a", "list.push({ pos, end: position, name, value });", "list.push({ pos, end: position, name, value: value + 1 });"},
		{"scout37_intern.a", "identifiers.set(text, identifier);", "// mutant: omit insertion"},
		{"scout37_diagnostics.a", "return file.diagnostics.slice();", "return file.diagnostics.slice(1);"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted", test.name))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), test.before) != 1 {
				t.Fatal("mutant lost its unique source anchor")
			}
			want := onNode(t, path)
			if want.exitCode != 0 || len(want.stdout) == 0 || len(want.stderr) != 0 {
				t.Fatalf("bad Node witness: %+v", want)
			}
			changed := filepath.Join(t.TempDir(), test.name)
			if err := os.WriteFile(changed, []byte(strings.Replace(string(source), test.before, test.after, 1)), 0o644); err != nil {
				t.Fatal(err)
			}
			mutant := onNode(t, changed)
			if mutant.exitCode != 0 || len(mutant.stderr) != 0 || bytes.Equal(want.stdout, mutant.stdout) {
				t.Fatalf("source mutant did not complete and disagree: %+v", mutant)
			}
			for _, side := range []struct {
				name, path string
				node       run
			}{{"control", path, want}, {"mutant", changed, mutant}} {
				t.Run(side.name, func(t *testing.T) {
					program, err := lowered(t, side.path)
					if err != nil {
						t.Fatal(err)
					}
					code := native.C(program)
					if !usesParallelMap(program) {
						t.Fatal("shape did not emit parallel work")
					}
					if diff := disagreement(side.node, onJavaScriptBackend(t, program)); diff != "" {
						t.Fatal("backend: " + diff)
					}
					binary := filepath.Join(t.TempDir(), "sanitized")
					if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
						t.Fatal(err)
					}
					for _, threads := range []string{"1", "16"} {
						got := executeParallel(t, threads, true, binary)
						if diff := disagreement(side.node, got); diff != "" {
							t.Fatalf("ASan/UBSan/leaks threads=%s: %s: %s", threads, diff, got.stderr)
						}
						if side.name == "mutant" && disagreement(want, got) == "" {
							t.Fatal("native mutant survived pristine Node comparison")
						}
					}
					if side.name == "control" && runtime.GOOS == "linux" {
						race := filepath.Join(t.TempDir(), "tsan")
						if err := native.Build(code, race, native.Options{ThreadSanitize: true}); err != nil {
							t.Fatal(err)
						}
						for _, threads := range []string{"1", "16"} {
							for attempt := 0; attempt < 3; attempt++ {
								got := executeParallel(t, threads, false, race)
								if diff := disagreement(want, got); diff != "" {
									t.Fatalf("TSan threads=%s: %s: %s", threads, diff, got.stderr)
								}
							}
						}
					}
				})
			}
			if !t.Failed() {
				t.Log("control matches Node; compiling, leak-clean mutant rejected by pristine Node stdout")
			}
		})
	}
}
