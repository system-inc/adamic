package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Review witnesses are independent top-level units. Register the same sources in
// main's oracle/count families; the module's two imported files are dependencies.
var asyncReviewCases = []struct{ path, output, reason, location string }{
	{"refused/computed_then.a", "1\n", "a computed field name", "computed_then.a:3:"},
	{
		"refused/hidden_mutated.a",
		"number\n",
		"Object.assign on a shape not proven by a plain literal or its const binding",
		"hidden_mutated.a:3:",
	},
	{"accepted/parallel_catch.a", "caught:parallel1\n", "", ""},
	{"accepted/conditional_throw.a", "caught:condition\n", "", ""},
	{"accepted/loop.a", "total:3|caught:loop:0\n", "", ""},
	{"accepted/branch.a", "value:7|caught:branch\n", "", ""},
	{"accepted/closure.a", "seed1:3|seed1:caught:closure\n", "", ""},
	{"accepted/method.a", "worker2:4|worker2:caught:method\n", "", ""},
	{"accepted/concurrent.a", "one:start\ntwo:start\none:resume\ntwo:resume\none:ok|caught:two\n", "", ""},
	{"accepted/recursion.a", "2:1:leaf|caught:recursive\n", "", ""},
	{"accepted/snapshot_reference.a", "before1right\n", "", ""},
	{
		"refused/unawaited_rejection.a",
		"done\n",
		"a promise not awaited in the same function",
		"unawaited_rejection.a:5:",
	},
	{"refused/hidden_await.a", "number\n", "await of a non-primitive structurally typed value", "hidden_await.a:4:"},
	{
		"refused/hidden_resolve.a",
		"number\n",
		"Promise.resolve of a non-primitive structurally typed value",
		"hidden_resolve.a:4:",
	},
	{
		"refused/hidden_return.a",
		"number\n",
		"async return of a non-primitive structurally typed value",
		"hidden_return.a:4:",
	},
	{"refused/modules/main.a", "a1\nb\na2\n", "top-level await outside the entry module", "suspended.a:2:"},
	{
		"refused/maybe_boolean_and.a",
		"undefined\n",
		"await inside maybe-boolean logical operands",
		"maybe_boolean_and.a:3:",
	},
}

func init() {
	for _, each := range asyncReviewCases {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/async_review/" + each.path, each.reason == "", false})
	}
}

func asyncReview(test *testing.T, name string) {
	test.Helper()
	for _, each := range asyncReviewCases {
		if each.path != name {
			continue
		}
		path, absolutePathError := filepath.Abs(
			filepath.Join(repository, "internal/oracle/testdata/async_review", each.path),
		)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		source := onNode(test, path)
		exit, stderr := 0, ""
		if name == "refused/unawaited_rejection.a" {
			exit, stderr = 70, "adamic: panic: Error: lost rejection\n"
		}
		if source.exitCode != exit || string(source.stdout) != each.output || string(source.stderr) != stderr {
			test.Fatalf("Node witness: %+v", source)
		}
		program, loweredError := lowered(test, path)
		if each.reason != "" {
			var notYet *lower.NotYet
			if !errors.As(loweredError, &notYet) ||
				!strings.Contains(loweredError.Error(), each.reason) ||
				!strings.Contains(notYet.Where, each.location) {
				test.Fatalf("want located refusal %q at %s: %v", each.reason, each.location, loweredError)
			}
			test.Logf(
				"Node stdout %q, stderr %q, exit %d; %v",
				source.stdout,
				source.stderr,
				source.exitCode,
				loweredError,
			)
			return
		}
		if loweredError != nil {
			test.Fatal(loweredError)
		}
		actual, leaked, comparable := nativelyCheckingLeaks(test, program, true)
		if leaked != "" {
			test.Fatal(leaked)
		}
		if comparable && disagreement(source, actual) != "" {
			test.Fatalf("native: %s: %+v", disagreement(source, actual), actual)
		}
		if actual := onJavaScriptBackend(test, program); disagreement(source, actual) != "" {
			test.Fatalf("JavaScript: %s: %+v", disagreement(source, actual), actual)
		}
		if actual := released(test, program); disagreement(source, actual) != "" {
			test.Fatalf("release: %s: %+v", disagreement(source, actual), actual)
		}
		return
	}
	test.Fatalf("unregistered review case %s", name)
}

func TestAsyncReviewParallelCatch(test *testing.T) {
	test.Parallel()
	asyncReview(test, "accepted/parallel_catch.a")
}
func TestAsyncReviewConditionalThrow(test *testing.T) {
	test.Parallel()
	asyncReview(test, "accepted/conditional_throw.a")
}
func TestAsyncReviewLoop(test *testing.T)   { test.Parallel(); asyncReview(test, "accepted/loop.a") }
func TestAsyncReviewBranch(test *testing.T) { test.Parallel(); asyncReview(test, "accepted/branch.a") }
func TestAsyncReviewClosure(test *testing.T) {
	test.Parallel()
	asyncReview(test, "accepted/closure.a")
}
func TestAsyncReviewMethod(test *testing.T) { test.Parallel(); asyncReview(test, "accepted/method.a") }
func TestAsyncReviewConcurrent(test *testing.T) {
	test.Parallel()
	asyncReview(test, "accepted/concurrent.a")
}
func TestAsyncReviewRecursion(test *testing.T) {
	test.Parallel()
	asyncReview(test, "accepted/recursion.a")
}
func TestAsyncReviewSnapshotReference(test *testing.T) {
	test.Parallel()
	asyncReview(test, "accepted/snapshot_reference.a")
}
func TestAsyncReviewUnawaitedRejection(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/unawaited_rejection.a")
}
func TestAsyncReviewHiddenAwait(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/hidden_await.a")
}
func TestAsyncReviewHiddenResolve(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/hidden_resolve.a")
}
func TestAsyncReviewHiddenReturn(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/hidden_return.a")
}
func TestAsyncReviewImportedAwait(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/modules/main.a")
}
func TestAsyncReviewMaybeBooleanAnd(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/maybe_boolean_and.a")
}

func asyncReviewProgram(test *testing.T, name string) (string, run) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/async_review/accepted", name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	return native.C(program), onNode(test, path)
}

func TestAsyncReviewNativeSuspensionRemovedMutant(test *testing.T) {
	test.Parallel()
	control, source := asyncReviewProgram(test, "concurrent")
	suspension := regexp.MustCompile(`adamic_async_await\(base, [^;]+\);`)
	location := suspension.FindStringIndex(control)
	if location == nil {
		test.Fatal("missing void suspension")
	}
	mutant := control[:location[0]] + "base->resume(base, (adamic_value){.number = 0}, false);" + control[location[1]:]
	binary := filepath.Join(test.TempDir(), "synchronous-resume")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if disagreement(source, actual) != "stdout differs" || actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("only Node scheduling order must catch removed suspension: %+v", actual)
	}
	test.Logf("Node %q; immediate resume %q", source.stdout, actual.stdout)
}

func TestAsyncReviewJavaScriptSuspensionRemovedMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/async_review/accepted/concurrent.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	control := javascript.JavaScript(program)
	suspension := regexp.MustCompile(`\(await (local_[0-9]+_suspension)\);`)
	location := suspension.FindStringSubmatchIndex(control)
	if location == nil {
		test.Fatal("missing void suspension")
	}
	mutant := control[:location[0]] + control[location[2]:location[3]] + ";" + control[location[1]:]
	file := filepath.Join(test.TempDir(), "synchronous.mjs")
	if writeError := os.WriteFile(file, []byte(mutant), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	actual := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		file,
	)
	source := onNode(test, path)
	if disagreement(source, actual) != "stdout differs" || actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("only Node scheduling order must catch removed JS suspension: %+v", actual)
	}
	test.Logf("Node %q; removed JS suspension %q", source.stdout, actual.stdout)
}

func TestAsyncReviewSnapshotRetainDroppedMutant(test *testing.T) {
	test.Parallel()
	control, source := asyncReviewProgram(test, "snapshot_reference")
	retain := regexp.MustCompile(`adamic_retain\((adamic_global_[A-Za-z0-9_]*_text)\)`)
	location := retain.FindStringSubmatchIndex(control)
	if location == nil {
		test.Fatal("missing global snapshot retain")
	}
	mutant := control[:location[0]] + control[location[2]:location[3]] + control[location[1]:]
	binary := filepath.Join(test.TempDir(), "borrowed-snapshot")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if actual.exitCode == 0 || !strings.Contains(string(actual.stderr), "heap-use-after-free") {
		test.Fatalf("snapshot retain mutant survived: Node %q; native %+v", source.stdout, actual)
	}
	test.Log("snapshot retain dropped: ASan heap-use-after-free")
}

func TestAsyncReviewRejectionSwallowedMutant(test *testing.T) {
	test.Parallel()
	control, source := asyncReviewProgram(test, "branch")
	mutant := strings.ReplaceAll(control, "if (rejected)", "if (false && rejected)")
	if mutant == control {
		test.Fatal("missing rejection dispatch")
	}
	binary := filepath.Join(test.TempDir(), "swallowed-rejection")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if disagreement(source, actual) != "stdout differs" || actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("only Node's catch output must catch swallowed rejection: %+v", actual)
	}
	if report := leakcheck.Report(test, mutant, binary); report != "" {
		test.Fatal(report)
	}
	test.Logf("Node %q; swallowed rejection %q", source.stdout, actual.stdout)
}

func TestAsyncReviewHiddenMutated(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/hidden_mutated.a")
}

func TestAsyncReviewComputedThen(test *testing.T) {
	test.Parallel()
	asyncReview(test, "refused/computed_then.a")
}
