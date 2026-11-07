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
	{"search_shrink_nullable.a", "findIndex", "", "", 0},
}

func init() {
	for _, fixture := range searchShrinkFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + fixture.name, true, fixture.element != ""})
	}
}

func TestSearchShrinkNumberElementMaybe(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 0)
}

func TestSearchShrinkReferenceElementMaybe(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 1)
}

func TestSearchShrinkFindLastIndexChecked(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 2)
}

func TestSearchShrinkFindLastChecked(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 3)
}

func TestSearchShrinkFindIndexChecked(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 4)
}

func TestSearchShrinkFindLastIndexMaybe(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 5)
}

func TestSearchShrinkFindMaybe(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 6)
}

func TestSearchShrinkFindLastMaybe(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 7)
}

func TestSearchShrinkFindChecked(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 8)
}

func TestSearchShrinkUnknown(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 9)
}

func TestSearchShrinkBoolean(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 10)
}

func TestSearchShrinkDefault(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 11)
}

func TestSearchShrinkOmitted(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 12)
}

func TestSearchShrinkFound(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 13)
}

func TestSearchShrinkNullableReference(test *testing.T) {
	test.Parallel()
	checkSearchShrink(test, 14)
}

// These focused witnesses run Node directly. The general observation cache's
// identity builds unrelated full-runtime variants before its first observation.
func searchShrinkOnNode(test *testing.T, path string) run {
	test.Helper()
	return execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "node.mjs"),
		path,
	)
}

func searchShrinkJavaScript(test *testing.T, script string) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		test.Fatal(err)
	}
	return searchShrinkOnNode(test, path)
}

func checkSearchShrink(test *testing.T, index int) {
	test.Helper()
	fixture := searchShrinkFixtures[index]
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture.name))
	if err != nil {
		test.Fatal(err)
	}
	truth := searchShrinkOnNode(test, path)
	if truth.exitCode != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := truth
	if fixture.element != "" {
		want = run{stdout: []byte(fixture.stdout), stderr: []byte(fmt.Sprintf(
			"adamic: panic: %s: index %d is undefined; element type %s does not admit undefined\n",
			fixture.method,
			fixture.index,
			fixture.element,
		)), exitCode: 70}
	}
	sanitized, binary := nativelyUncached(test, program)
	for backend, got := range map[string]run{
		"sanitized":  sanitized,
		"release":    releasedUncached(test, program),
		"javascript": searchShrinkJavaScript(test, javascript.JavaScript(program)),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s, want %+v got %+v", backend, difference, want, got)
		}
	}
	test.Logf("Node: exit=%d stdout=%q", truth.exitCode, truth.stdout)
	if sanitized.exitCode == 0 {
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatal(report)
		}
	}

}

// Both mutations compile and run. Only the fixture's behavioral assertion kills
// them: the missing check returns normally, while the old stop rejects Node's run.
func TestSearchShrinkMutantSkipCheck(test *testing.T) {
	test.Parallel()
	checkSearchShrinkMutant(test, "skip-check")
}

func TestSearchShrinkMutantOldStop(test *testing.T) {
	test.Parallel()
	checkSearchShrinkMutant(test, "old-stop")
}

func checkSearchShrinkMutant(test *testing.T, mutant string) {
	test.Helper()
	name := "search_shrink_findIndex_checked.a"
	if mutant == "old-stop" {
		name = "047cb0d_f_maybe.a"
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
	if err != nil {
		test.Fatal(err)
	}
	truth := searchShrinkOnNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := native.C(program)
	script := javascript.JavaScript(program)
	if mutant == "skip-check" {
		pattern := regexp.MustCompile(`if \(!\((adamic_temporary_[A-Za-z0-9_]+)\)\)`)
		source = pattern.ReplaceAllString(source, `if (false && !($1))`)
		script = strings.Replace(
			script,
			"index >= array.length && !admitsUndefined",
			"index >= array.length && false",
			1,
		)
	} else {
		pattern := regexp.MustCompile(`bool (\w+) = (\w+) < (\w+)->length;`)
		source = pattern.ReplaceAllString(source, `bool $1 = $2 < $3->length;
    if (!$1) { static const char message[] = "findIndex: the array shrank while it was being searched"; adamic_panic(message, sizeof message - 1); }`)
		script = strings.Replace(
			script,
			"if (index >= array.length && !admitsUndefined) panic(`${method}: index ${index} is undefined; element type ${elementType} does not admit undefined`);",
			"if (index >= array.length) panic(`${method}: the array shrank while it was being searched`);",
			1,
		)
	}
	if source == native.C(program) || script == javascript.JavaScript(program) {
		test.Fatal("mutant changed nothing")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	nativeRun := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	jsPath := filepath.Join(test.TempDir(), "mutant.mjs")
	if err := os.WriteFile(jsPath, []byte(script), 0600); err != nil {
		test.Fatal(err)
	}
	for backend, got := range map[string]run{"native": nativeRun, "javascript": searchShrinkOnNode(test, jsPath)} {
		if mutant == "skip-check" {
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("%s mutation failed outside the exit-70 assertion: %+v", backend, got)
			}
			want := run{stdout: []byte("0 1\n1 2\n"), stderr: []byte(
				"adamic: panic: findIndex: index 2 is undefined; element type number does not admit undefined\n",
			), exitCode: 70}
			if disagreement(want, got) != "exit codes differ" {
				test.Fatalf("%s missing check escaped: %+v", backend, got)
			}
		} else {
			if got.exitCode != 70 || disagreement(truth, got) != "exit codes differ" {
				test.Fatalf("%s old stop escaped: %+v", backend, got)
			}
		}
		test.Logf(
			"%s caught %s by exit code: exit=%d stdout=%q stderr=%q",
			backend,
			mutant,
			got.exitCode,
			got.stdout,
			got.stderr,
		)
	}

}
