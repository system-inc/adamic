package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Keep the slice's new checked fixtures registered without editing the shared oracle list.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/library_array_find_last_shrinks.a",
		"internal/oracle/testdata/library_array_find_last_value_shrinks.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, true})
	}
	for _, path := range []string{
		"internal/oracle/testdata/library_array_dense_construction.a",
		"internal/oracle/testdata/library_array_reductions.a",
		"internal/oracle/testdata/library_array_default_sort.a",
		"internal/oracle/testdata/library_array_receiver_calls.a",
		"internal/oracle/testdata/library_array_from_string.a",
		"internal/oracle/testdata/library_array_string_mutations.a",
		"internal/oracle/testdata/library_array_find_shrinks_optional_numbers.a",
		"internal/oracle/testdata/library_array_find_shrinks_optional_leaves.a",
		"internal/oracle/testdata/library_array_find_shrinks_optional_methods.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// Raw Node visits removed indexes with undefined. Adamic protects a callback whose
// element type cannot hold that value with an inserted check, on both backends.
func TestArrayShrinkingSearchChecksAndMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ fixture, method, result string }{
		{"library_array_find_last_shrinks.a", "findLastIndex", "-1"},
		{"library_array_find_last_value_shrinks.a", "findLast", "absent"},
	} {
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "3 dd\n2 undefined\n1 undefined\n0 dd\n"+test.result+"\n" || len(truth.stderr) != 0 {
				t.Fatalf("unexpected raw Node observation: %+v", truth)
			}
			check := func(got run) bool {
				return got.exitCode == 70 && string(got.stdout) == "3 dd\n" && string(got.stderr) == "adamic: panic: "+test.method+": the array shrank while it was being searched\n"
			}
			original, _ := natively(t, program)
			if !check(original) || !check(onJavaScriptBackend(t, program)) {
				t.Fatal("original inserted check failed")
			}
			t.Logf("Node: exit 0, stdout %q; both backends: exit 70 after 3 dd", truth.stdout)

			// Mutate the native backend's emitted check alone. This C must still compile
			// and finish cleanly; the intended check, not a sanitizer, catches it.
			c := native.C(program)
			pattern := regexp.MustCompile(`static const char message\[\] = "` + test.method + `: the array shrank while it was being searched";\s*adamic_panic\(message, sizeof message - 1\);`)
			mutant := pattern.ReplaceAllString(c, "continue;")
			if mutant == c {
				t.Fatal("native mutant changed nothing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("native mutant failed outside comparison: %+v", got)
			}
			if check(got) || string(got.stdout) == string(truth.stdout) {
				t.Fatal("native skip mutant escaped comparisons")
			}
			t.Logf("native skip mutant caught: exit %d, stdout %q; no sanitizer or leak report", got.exitCode, got.stdout)

			js := javascript.JavaScript(program)
			mutatedJS := strings.Replace(js, "if (!allowsUndefined && index >= array.length) panic(`${method}: the array shrank while it was being searched`);", "if (index >= array.length) continue;", 1)
			if mutatedJS == js {
				t.Fatal("JavaScript mutant changed nothing")
			}
			module := filepath.Join(t.TempDir(), "mutant.mjs")
			if err := os.WriteFile(module, []byte(mutatedJS), 0644); err != nil {
				t.Fatal(err)
			}
			got = onNode(t, module)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("JavaScript mutant failed outside comparison: %+v", got)
			}
			if check(got) || string(got.stdout) == string(truth.stdout) {
				t.Fatal("JavaScript skip mutant escaped comparisons")
			}
			t.Logf("JavaScript skip mutant caught: exit %d, stdout %q", got.exitCode, got.stdout)
			// The original artifacts remain the baseline after each independent mutation.
			if native.C(program) != c || javascript.JavaScript(program) != js {
				t.Fatal("mutant was not restored")
			}
		})
	}
}

// Reintroduce the old unconditional stop separately in each backend. The raw
// Node run, not an internal expected IR, decides whether these mutants fail.
func TestArrayOptionalShrinkingSearchMutants(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"library_array_find_shrinks_optional_numbers.a", "library_array_find_shrinks_optional_leaves.a"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "1\n2\nhole\n-1\n" || len(truth.stderr) != 0 {
				t.Fatalf("unexpected Node observation: %+v", truth)
			}
			original, _ := natively(t, program)
			for _, got := range []run{original, onJavaScriptBackend(t, program)} {
				if got.exitCode != truth.exitCode || string(got.stdout) != string(truth.stdout) || len(got.stderr) != 0 {
					t.Fatalf("original differs from Node: %+v", got)
				}
			}
			t.Logf("Node and both backends: exit 0, stdout %q", truth.stdout)
			c := native.C(program)
			pattern := regexp.MustCompile(`(?m)^\s*\w+ = \(adamic_value\)\{\.(?:number|reference) = [^\n]+\};$`)
			mutant := pattern.ReplaceAllString(c, `static const char message[] = "findIndex: the array shrank while it was being searched"; adamic_panic(message, sizeof message - 1);`)
			if mutant == c {
				t.Fatal("native unconditional-stop mutant changed nothing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			caught := func(backend string, got run) {
				t.Helper()
				if got.exitCode != 70 || string(got.stdout) != "1\n2\n" || string(got.stderr) != "adamic: panic: findIndex: the array shrank while it was being searched\n" {
					t.Fatalf("%s mutant failed outside the intended comparison: %+v", backend, got)
				}
				if got.exitCode == truth.exitCode && string(got.stdout) == string(truth.stdout) {
					t.Fatal("unconditional-stop mutant escaped Node comparison")
				}
				t.Logf("%s unconditional-stop mutant caught by Node comparison: exit 70, stdout %q; no sanitizer report", backend, got.stdout)
			}
			caught("native", executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary))
			js := javascript.JavaScript(program)
			mutatedJS := strings.Replace(js, "if (!allowsUndefined && index >= array.length) panic", "if (index >= array.length) panic", 1)
			if mutatedJS == js {
				t.Fatal("JavaScript unconditional-stop mutant changed nothing")
			}
			module := filepath.Join(t.TempDir(), "mutant.mjs")
			if err := os.WriteFile(module, []byte(mutatedJS), 0644); err != nil {
				t.Fatal(err)
			}
			caught("JavaScript", onNode(t, module))
			if native.C(program) != c || javascript.JavaScript(program) != js {
				t.Fatal("mutant was not restored")
			}
		})
	}
}
