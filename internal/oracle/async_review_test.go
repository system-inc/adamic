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
	{"refused/hidden_mutated.a", "number\n", "Object.assign on a shape not proven by a plain literal or its const binding", "hidden_mutated.a:3:"},
	{"accepted/parallel_catch.a", "caught:parallel1\n", "", ""},
	{"accepted/conditional_throw.a", "caught:condition\n", "", ""},
	{"accepted/loop.a", "total:3|caught:loop:0\n", "", ""},
	{"accepted/branch.a", "value:7|caught:branch\n", "", ""},
	{"accepted/closure.a", "seed1:3|seed1:caught:closure\n", "", ""},
	{"accepted/method.a", "worker2:4|worker2:caught:method\n", "", ""},
	{"accepted/concurrent.a", "one:start\ntwo:start\none:resume\ntwo:resume\none:ok|caught:two\n", "", ""},
	{"accepted/recursion.a", "2:1:leaf|caught:recursive\n", "", ""},
	{"accepted/snapshot_reference.a", "before1right\n", "", ""},
	{"refused/unawaited_rejection.a", "done\n", "a promise not awaited in the same function", "unawaited_rejection.a:5:"},
	{"refused/hidden_await.a", "number\n", "await of a non-primitive structurally typed value", "hidden_await.a:4:"},
	{"refused/hidden_resolve.a", "number\n", "Promise.resolve of a non-primitive structurally typed value", "hidden_resolve.a:4:"},
	{"refused/hidden_return.a", "number\n", "async return of a non-primitive structurally typed value", "hidden_return.a:4:"},
	{"refused/modules/main.a", "a1\nb\na2\n", "top-level await outside the entry module", "suspended.a:2:"},
	{"refused/maybe_boolean_and.a", "undefined\n", "await inside maybe-boolean logical operands", "maybe_boolean_and.a:3:"},
}

func init() {
	for _, each := range asyncReviewCases {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/async_review/" + each.path, each.reason == "", false})
	}
}

func asyncReview(t *testing.T, name string) {
	t.Helper()
	for _, each := range asyncReviewCases {
		if each.path != name {
			continue
		}
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_review", each.path))
		if err != nil {
			t.Fatal(err)
		}
		source := onNode(t, path)
		exit, stderr := 0, ""
		if name == "refused/unawaited_rejection.a" {
			exit, stderr = 70, "adamic: panic: Error: lost rejection\n"
		}
		if source.exitCode != exit || string(source.stdout) != each.output || string(source.stderr) != stderr {
			t.Fatalf("Node witness: %+v", source)
		}
		program, err := lowered(t, path)
		if each.reason != "" {
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), each.reason) || !strings.Contains(notYet.Where, each.location) {
				t.Fatalf("want located refusal %q at %s: %v", each.reason, each.location, err)
			}
			t.Logf("Node stdout %q, stderr %q, exit %d; %v", source.stdout, source.stderr, source.exitCode, err)
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
		if leaked != "" {
			t.Fatal(leaked)
		}
		if comparable && disagreement(source, actual) != "" {
			t.Fatalf("native: %s: %+v", disagreement(source, actual), actual)
		}
		if actual := onJavaScriptBackend(t, program); disagreement(source, actual) != "" {
			t.Fatalf("JavaScript: %s: %+v", disagreement(source, actual), actual)
		}
		if actual := released(t, program); disagreement(source, actual) != "" {
			t.Fatalf("release: %s: %+v", disagreement(source, actual), actual)
		}
		return
	}
	t.Fatalf("unregistered review case %s", name)
}

func TestAsyncReviewParallelCatch(t *testing.T) {
	t.Parallel()
	asyncReview(t, "accepted/parallel_catch.a")
}
func TestAsyncReviewConditionalThrow(t *testing.T) {
	t.Parallel()
	asyncReview(t, "accepted/conditional_throw.a")
}
func TestAsyncReviewLoop(t *testing.T)       { t.Parallel(); asyncReview(t, "accepted/loop.a") }
func TestAsyncReviewBranch(t *testing.T)     { t.Parallel(); asyncReview(t, "accepted/branch.a") }
func TestAsyncReviewClosure(t *testing.T)    { t.Parallel(); asyncReview(t, "accepted/closure.a") }
func TestAsyncReviewMethod(t *testing.T)     { t.Parallel(); asyncReview(t, "accepted/method.a") }
func TestAsyncReviewConcurrent(t *testing.T) { t.Parallel(); asyncReview(t, "accepted/concurrent.a") }
func TestAsyncReviewRecursion(t *testing.T)  { t.Parallel(); asyncReview(t, "accepted/recursion.a") }
func TestAsyncReviewSnapshotReference(t *testing.T) {
	t.Parallel()
	asyncReview(t, "accepted/snapshot_reference.a")
}
func TestAsyncReviewUnawaitedRejection(t *testing.T) {
	t.Parallel()
	asyncReview(t, "refused/unawaited_rejection.a")
}
func TestAsyncReviewHiddenAwait(t *testing.T) { t.Parallel(); asyncReview(t, "refused/hidden_await.a") }
func TestAsyncReviewHiddenResolve(t *testing.T) {
	t.Parallel()
	asyncReview(t, "refused/hidden_resolve.a")
}
func TestAsyncReviewHiddenReturn(t *testing.T) {
	t.Parallel()
	asyncReview(t, "refused/hidden_return.a")
}
func TestAsyncReviewImportedAwait(t *testing.T) {
	t.Parallel()
	asyncReview(t, "refused/modules/main.a")
}
func TestAsyncReviewMaybeBooleanAnd(t *testing.T) {
	t.Parallel()
	asyncReview(t, "refused/maybe_boolean_and.a")
}

func asyncReviewProgram(t *testing.T, name string) (string, run) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_review/accepted", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return native.C(program), onNode(t, path)
}

func TestAsyncReviewNativeSuspensionRemovedMutant(t *testing.T) {
	t.Parallel()
	control, source := asyncReviewProgram(t, "concurrent")
	suspension := regexp.MustCompile(`adamic_async_await\(base, [^;]+\);`)
	location := suspension.FindStringIndex(control)
	if location == nil {
		t.Fatal("missing void suspension")
	}
	mutant := control[:location[0]] + "base->resume(base, (adamic_value){.number = 0}, false);" + control[location[1]:]
	binary := filepath.Join(t.TempDir(), "synchronous-resume")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if disagreement(source, actual) != "stdout differs" || actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("only Node scheduling order must catch removed suspension: %+v", actual)
	}
	t.Logf("Node %q; immediate resume %q", source.stdout, actual.stdout)
}

func TestAsyncReviewJavaScriptSuspensionRemovedMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_review/accepted/concurrent.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := javascript.JavaScript(program)
	suspension := regexp.MustCompile(`\(await (local_[0-9]+_suspension)\);`)
	location := suspension.FindStringSubmatchIndex(control)
	if location == nil {
		t.Fatal("missing void suspension")
	}
	mutant := control[:location[0]] + control[location[2]:location[3]] + ";" + control[location[1]:]
	file := filepath.Join(t.TempDir(), "synchronous.mjs")
	if err := os.WriteFile(file, []byte(mutant), 0o644); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), file)
	source := onNode(t, path)
	if disagreement(source, actual) != "stdout differs" || actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("only Node scheduling order must catch removed JS suspension: %+v", actual)
	}
	t.Logf("Node %q; removed JS suspension %q", source.stdout, actual.stdout)
}

func TestAsyncReviewSnapshotRetainDroppedMutant(t *testing.T) {
	t.Parallel()
	control, source := asyncReviewProgram(t, "snapshot_reference")
	retain := regexp.MustCompile(`adamic_retain\((adamic_global_[A-Za-z0-9_]*_text)\)`)
	location := retain.FindStringSubmatchIndex(control)
	if location == nil {
		t.Fatal("missing global snapshot retain")
	}
	mutant := control[:location[0]] + control[location[2]:location[3]] + control[location[1]:]
	binary := filepath.Join(t.TempDir(), "borrowed-snapshot")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if actual.exitCode == 0 || !strings.Contains(string(actual.stderr), "heap-use-after-free") {
		t.Fatalf("snapshot retain mutant survived: Node %q; native %+v", source.stdout, actual)
	}
	t.Log("snapshot retain dropped: ASan heap-use-after-free")
}

func TestAsyncReviewRejectionSwallowedMutant(t *testing.T) {
	t.Parallel()
	control, source := asyncReviewProgram(t, "branch")
	mutant := strings.ReplaceAll(control, "if (rejected)", "if (false && rejected)")
	if mutant == control {
		t.Fatal("missing rejection dispatch")
	}
	binary := filepath.Join(t.TempDir(), "swallowed-rejection")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if disagreement(source, actual) != "stdout differs" || actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("only Node's catch output must catch swallowed rejection: %+v", actual)
	}
	if report := leakcheck.Report(t, mutant, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("Node %q; swallowed rejection %q", source.stdout, actual.stdout)
}

func TestAsyncReviewHiddenMutated(t *testing.T) {
	t.Parallel()
	asyncReview(t, "refused/hidden_mutated.a")
}
