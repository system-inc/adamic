package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var searchShrinkFixtures = []struct {
	name, method, stdout, element string
	index                         int
}{
	{"047cb0d_f_maybe.a", "findIndex", "", "", 0},
	{"047cb0d_f_maybe_ref.a", "findIndex", "", "", 0},
	{"57f2d04_find_last_shrinks.a", "findLastIndex", "3 dd\n", "string", 2},
	{"57f2d04_find_last_shrinks2.a", "findLast", "3 dd\n", "string", 2},
	{"search_shrink_findIndex_checked.a", "findIndex", "0 1\n1 2\n", "number", 2},
	{"search_shrink_findLastIndex_maybe.a", "findLastIndex", "", "", 0},
	{"search_shrink_find_maybe.a", "find", "", "", 0},
	{"search_shrink_findLast_maybe.a", "findLast", "", "", 0},
	{"search_shrink_find_checked.a", "find", "0 1\n1 2\n", "number", 2},
	{"search_shrink_unknown.a", "find", "", "", 0},
	{"search_shrink_boolean.a", "find", "", "", 0},
	{"search_shrink_default.a", "findIndex", "", "", 0},
	{"search_shrink_omitted.a", "findIndex", "", "", 0},
	{"search_shrink_found.a", "find", "", "", 0},
}

func init() {
	for _, fixture := range searchShrinkFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + fixture.name, true, fixture.element != ""})
	}
}

func TestSearchShrink(t *testing.T) {
	t.Parallel()
	for _, fixture := range searchShrinkFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture.name))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := truth
			if fixture.element != "" {
				want = run{stdout: []byte(fixture.stdout), stderr: []byte(fmt.Sprintf("adamic: panic: %s: index %d is undefined; element type %s does not admit undefined\n", fixture.method, fixture.index, fixture.element)), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"sanitized": sanitized, "release": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s, want %+v got %+v", backend, difference, want, got)
				}
			}
			t.Logf("Node: exit=%d stdout=%q", truth.exitCode, truth.stdout)
			if sanitized.exitCode == 0 {
				if report := leakSanitizer(t, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// Both mutations compile and run. Only the fixture's behavioral assertion kills
// them: the missing check returns normally, while the old stop rejects Node's run.
func TestSearchShrinkMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []string{"skip-check", "old-stop"} {
		t.Run(mutant, func(t *testing.T) {
			t.Parallel()
			name := "search_shrink_findIndex_checked.a"
			if mutant == "old-stop" {
				name = "047cb0d_f_maybe.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			script := javascript.JavaScript(program)
			if mutant == "skip-check" {
				pattern := regexp.MustCompile(`if \(!\((adamic_temporary_[0-9]+)\)\)`)
				source = pattern.ReplaceAllString(source, `if (false && !($1))`)
				script = strings.Replace(script, "index >= array.length && !admitsUndefined", "index >= array.length && false", 1)
			} else {
				pattern := regexp.MustCompile(`bool (\w+) = (\w+) < (\w+)->length;`)
				source = pattern.ReplaceAllString(source, `bool $1 = $2 < $3->length;
    if (!$1) { static const char message[] = "findIndex: the array shrank while it was being searched"; adamic_panic(message, sizeof message - 1); }`)
				script = strings.Replace(script, "if (index >= array.length && !admitsUndefined) panic(`${method}: index ${index} is undefined; element type ${elementType} does not admit undefined`);", "if (index >= array.length) panic(`${method}: the array shrank while it was being searched`);", 1)
			}
			if source == native.C(program) || script == javascript.JavaScript(program) {
				t.Fatal("mutant changed nothing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			nativeRun := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			jsPath := filepath.Join(t.TempDir(), "mutant.mjs")
			if err := os.WriteFile(jsPath, []byte(script), 0600); err != nil {
				t.Fatal(err)
			}
			for backend, got := range map[string]run{"native": nativeRun, "javascript": onNode(t, jsPath)} {
				if mutant == "skip-check" {
					if got.exitCode != 0 || len(got.stderr) != 0 {
						t.Fatalf("%s mutation failed outside the exit-70 assertion: %+v", backend, got)
					}
					want := run{stdout: []byte("0 1\n1 2\n"), stderr: []byte("adamic: panic: findIndex: index 2 is undefined; element type number does not admit undefined\n"), exitCode: 70}
					if disagreement(want, got) != "exit codes differ" {
						t.Fatalf("%s missing check escaped: %+v", backend, got)
					}
				} else {
					if got.exitCode != 70 || disagreement(truth, got) != "exit codes differ" {
						t.Fatalf("%s old stop escaped: %+v", backend, got)
					}
				}
				t.Logf("%s caught %s by exit code: exit=%d stdout=%q stderr=%q", backend, mutant, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}
