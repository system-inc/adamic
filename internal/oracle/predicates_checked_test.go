package oracle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// guard, where given, is the native stderr of the removed-check mutant: it reads a field the false
// side's claim says is there, and native stops at that read's own guard where Node, and the
// JavaScript backend, read undefined.
var checkedPredicateFixtures = []struct {
	name, nodeOut, checkedOut, branch, source, target, guard string
	checked, unobservable                                    int
}{
	{"true", "claimed\n", "", "true", "string | number", "number", "", 2, 0},
	{"true_unused", "continued\n", "", "true", "\"text\"", "number", "", 1, 1},
	{"false", "continued\n", "", "false", "string | number", "number", "", 2, 0},
	{"false_alias", "continued\n", "", "false", "string | number", "number", "", 2, 0},
	{"false_compound_unchanged", "continued\n", "", "", "", "", "", 1, 1},
	{"asserts", "called\nasserted\n", "called\n", "asserts", "string | number", "number", "", 1, 0},
	{"valid", "argument\ncalled\nnumber\nargument\ncalled\nstring\n", "", "", "", "", "", 1, 1},
	{"asserts_valid", "called\n7\n", "", "", "", "", "", 1, 0},
	{"condition_valid", "continued\n", "", "", "\"truthy\"", "truthy", "", 1, 0},
	{"condition_lie", "continued\n", "", "asserts", "0", "truthy", "", 1, 0},
	{"condition_once", "argument\ncontinued\n", "", "", "number", "truthy", "", 1, 0},
	{"condition_once_lie", "argument\ncontinued\n", "argument\n", "asserts", "number", "truthy", "", 1, 0},
	{"false_unchanged", "continued\n", "", "", "", "", "", 1, 1},
	// tsc's Debug.assertNotNode (debug.ts:307): the bodyless overload claims Exclude<T, U> from the
	// call's resolved signature, and the implementation, gated by an assertion level as tsc's is,
	// never runs the test. Its own claim is checked at the call, so a Keyword stops it there.
	{"assert_not_node", "asserting\nnarrowed\n", "asserting\n", "asserts", "Node", "Identifier", "", 1, 0},
	// Language's false-branch forms (#7v4rjr7): the checker narrows each one's false side, and isS
	// says false to a Statement, so the false check must stop it before n.e is read.
	{
		"false_switch_true",
		"undefined\n",
		"",
		"false",
		"Node",
		"Statement",
		"adamic: panic: compiler bug: a field the checker proved is there is missing\n",
		2,
		0,
	},
	{
		"false_eq_false",
		"undefined\n",
		"",
		"false",
		"Node",
		"Statement",
		"adamic: panic: compiler bug: a field the checker proved is there is missing\n",
		2,
		0,
	},
	{
		"false_ne_true",
		"undefined\n",
		"",
		"false",
		"Node",
		"Statement",
		"adamic: panic: compiler bug: a field the checker proved is there is missing\n",
		2,
		0,
	},
	{
		"false_false_eq",
		"undefined\n",
		"",
		"false",
		"Node",
		"Statement",
		"adamic: panic: compiler bug: a field the checker proved is there is missing\n",
		2,
		0,
	},
	{
		"false_comma",
		"undefined\n",
		"",
		"false",
		"Node",
		"Statement",
		"adamic: panic: compiler bug: a field the checker proved is there is missing\n",
		2,
		0,
	},
}

func predicateCheckedInput(test *testing.T, name string) (string, string) {
	test.Helper()
	return predicateInputFrom(test, "internal/lower/testdata/predicates_checked", name)
}

// predicateInputFrom is a stored .a witness in directory and its .ts copy, which checked mode reads.
func predicateInputFrom(test *testing.T, directory, name string) (string, string) {
	test.Helper()
	witness, absolutePathError := filepath.Abs(filepath.Join(repository, directory, name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(witness)
	if readError != nil {
		test.Fatal(readError)
	}
	checked := filepath.Join(test.TempDir(), name+".ts")
	if writeError := os.WriteFile(checked, source, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return witness, checked
}

func TestCheckedPredicateOracle(test *testing.T) {
	test.Parallel()
	for _, probe := range checkedPredicateFixtures {
		test.Run(probe.name, func(test *testing.T) {
			witness, path := predicateCheckedInput(test, probe.name)
			node := onNode(test, witness)
			expectedNode := run{stdout: []byte(probe.nodeOut)}
			if difference := disagreement(expectedNode, node); difference != "" {
				test.Fatal("source Node: " + difference)
			}
			if strings.HasPrefix(probe.name, "condition_") {
				var refused *lower.Refused
				if _, loweredError := lowered(test, witness); !errors.As(loweredError, &refused) {
					test.Fatalf(".a condition assertion: want refusal, got %v", loweredError)
				}
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			counts := program.PredicateChecks
			if counts.Checked != probe.checked ||
				counts.Unobservable != probe.unobservable ||
				counts.Proven != 0 ||
				len(counts.Sites) != 1 {
				test.Fatalf("direction counts %+v", counts)
			}
			if probe.target == "truthy" {
				directions := counts.Sites[0].Directions
				if len(directions) != 1 ||
					directions[0].Direction != "asserts" ||
					directions[0].Status != "checked" ||
					directions[0].Reason != probe.source+" -> truthy" {
					test.Fatalf("condition assertion counts: %+v", directions)
				}
			}
			want := node
			if probe.branch != "" {
				site := counts.Sites[0]
				// The expected source and target strings are pins, not emitted diagnostics.
				message := fmt.Sprintf(
					"adamic: panic: predicate %s at %s %s branch: source %s, target %s\n",
					site.Function,
					site.Where,
					probe.branch,
					probe.source,
					probe.target,
				)
				want = run{stdout: []byte(probe.checkedOut), stderr: []byte(message), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(test, program)
			for _, got := range []run{sanitized, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatal(difference)
				}
			}
			if want.exitCode == 0 {
				if report := leaksUncached(test, program, binary); report != "" {
					test.Fatal(report)
				}
				countedBinary := filepath.Join(test.TempDir(), "counted")
				if buildError := native.Build(
					native.C(program),
					countedBinary,
					native.Options{Count: true},
				); buildError != nil {
					test.Fatal(buildError)
				}
				if report := leakcheck.Unbalanced(leakRun(execute(test, countedBinary))); report != "" {
					test.Fatal(report)
				}
			}
			if probe.branch != "" {
				changes := 0
				for i := range program.Functions {
					if strings.HasSuffix(program.Functions[i].Name, "_predicate_check") {
						program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
							if panic, ok := node.(ir.Panic); ok {
								constant, ok := panic.Message.(ir.StringConstant)
								if ok &&
									strings.Contains(program.Strings[constant.Index], " "+probe.branch+" branch:") {
									changes++
									return ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
								}
							}
							return node
						})
					}
				}
				if changes != 1 {
					test.Fatalf("check mutant removed %d guards", changes)
				}
				mutant, _ := nativelyUncached(test, program)
				for index, got := range []run{mutant, onJavaScriptBackend(test, program)} {
					if disagreement(want, got) == "" {
						test.Fatal("removed-check mutant escaped exit-70 contract")
					}
					expected := node
					if index == 0 && probe.guard != "" {
						expected = run{stderr: []byte(probe.guard), exitCode: 70}
					}
					if difference := disagreement(expected, got); difference != "" {
						test.Fatal("mutant failed to execute source semantics: " + difference)
					}
				}
				test.Logf(
					"%s lying body exits 70; removed-%s-check mutant executes Node result %q and is caught",
					probe.name,
					probe.branch,
					node.stdout,
				)
			}
		})
	}
}

// tsc's Node hierarchy tags on an open SyntaxKind over a non-union base, so a kind proves no layout:
// another member can hold the same value. The true branch reads the target as a checked view, as a
// .ts cast does, so an honest program agrees with Node and a lying tag stops at its first field read
// with the view's own located message, never a compiler-bug panic or another guard (#k1t2wp9).
func TestCheckedPredicateOpenTagViews(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, nodeOut, stderr string }{
		{"open_tag_view", "call f\nlike f\nlike G\nother\n", ""},
		// Identifier and Keyword share SyntaxKind.Identifier with different layouts, and isIdentifier
		// says yes to a Keyword: Node reads width as undefined and prints NaN.
		{
			"open_tag_view_lie",
			"NaN\n",
			"adamic: panic: field read failed: node.width is not initialized; expected number, found missing\n",
		},
		// The same lie with the field present at another representation: nothing but the view sees it.
		// Without the view, native reads a string's slot as a number where Node prints wide1.
		{
			"open_tag_view_retype",
			"wide1\n",
			"adamic: panic: field read failed: node.width is not a number; expected number, found string\n",
		},
		// The lie one level down: Keyword's name is a Width where Identifier's is a Name. Both name fields
		// hold an object, so only the view's nested contract checks the read of text; Node prints
		// "name undefined" and stops nowhere.
		{
			"open_tag_view_nested",
			"name undefined\nname x\n",
			"adamic: panic: field read failed: node.name.text is not initialized; expected string, found missing\n",
		},
		// Kind.Identifier and Kind.Keyword share 1, so a Keyword passes isIdentifier's tag test: Node reads
		// text as undefined; the view stops at that read.
		{
			"shared_kind",
			"undefined\nshown\n",
			"adamic: panic: field read failed: node.text is not initialized; expected string, found missing\n",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			witness, path := predicateCheckedInput(test, probe.name)
			node := onNode(test, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				test.Fatal("source Node: " + difference)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want := node
			if probe.stderr != "" {
				want = run{stderr: []byte(probe.stderr), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(test, program)
			for index, got := range []run{
				sanitized,
				releasedUncached(test, program),
				onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf(
						"%s on %s: exit %d, stdout %q, stderr %q",
						difference,
						[]string{"sanitized native", "release native", "the JavaScript backend"}[index],
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
			if want.exitCode == 0 {
				if report := leaksUncached(test, program, binary); report != "" {
					test.Fatal(report)
				}
			}
		})
	}
}

// Each row is keyed as its .ts checked mode copy, read from the stored .a witness.
func checkedPredicateCountKey(name string) string {
	return "internal/lower/testdata/predicates_checked/" + name + ".a (.ts checked mode)"
}

func init() {
	for _, probe := range checkedPredicateFixtures {
		key := checkedPredicateCountKey(probe.name)
		countFixtureSources[key] = "internal/lower/testdata/predicates_checked/" + probe.name + ".a"
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{key, func(test *testing.T) (string, func() string) {
				return checkedPredicateCounted(test, key, probe.name)
			}},
		)
	}
}

// Give the .ts copy its logical name before lowering: a check's panic message carries its
// path, so replacing a scratch path in the emitted C afterward leaves that path's string lengths behind.
func checkedPredicateCounted(test *testing.T, key, name string) (string, func() string) {
	test.Helper()
	_, typed := predicateCheckedInput(test, name)
	logical := "<repository>/internal/lower/testdata/predicates_checked/" + name + ".ts"
	checked, loadDisplayedError := load.LoadDisplayed([]string{typed}, func(path string) string {
		if path == typed {
			return logical
		}
		return path
	})
	if loadDisplayedError != nil {
		test.Fatal(loadDisplayedError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := native.C(program)
	return countPortable(test, source, ""), func() string {
		return countedRun(test, key, typed, source, false, nil, false, false)
	}
}

// TestCountsAreRecorded's shards hold these rows with every other; this focused verifier
// measures only the checked predicate rows against their committed records.
func TestCheckedPredicateCountsAreRecorded(test *testing.T) {
	test.Parallel()
	// TestMain owns committed writes during -update-counts.
	if *updateCounts {
		return
	}
	for _, probe := range checkedPredicateFixtures {
		key := checkedPredicateCountKey(probe.name)
		_, measure := checkedPredicateCounted(test, key, probe.name)
		want, readCountRowError := readCountRow(countsDirectory, "runtime", key)
		if readCountRowError != nil {
			test.Fatal(readCountRowError)
		}
		if got := measure(); got != want {
			test.Errorf(
				"checked predicate counts changed in %s; update with -update-counts:\nrecorded: %s\nmeasured: %s",
				countFile("runtime", key),
				want,
				got,
			)
		}
	}
}

// The forwarder checks its saved parameter too, with its declaration's location.
func TestCheckedPredicateConditionForwarderOracle(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name  string
		lying bool
	}{
		{"condition_value_valid", false}, {"condition_value_lie", true},
	} {
		test.Run(probe.name, func(test *testing.T) {
			witness, path := predicateCheckedInput(test, probe.name)
			node := onNode(test, witness)
			if difference := disagreement(run{stdout: []byte("argument\ncontinued\n")}, node); difference != "" {
				test.Fatal("source Node: " + difference)
			}
			var refused *lower.Refused
			if _, loweredError := lowered(test, witness); !errors.As(loweredError, &refused) {
				test.Fatalf(".a condition assertion: want refusal, got %v", loweredError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			if len(program.PredicateChecks.Sites) != 0 {
				test.Fatalf("forwarder counted as direct call: %+v", program.PredicateChecks)
			}
			want := node
			if probe.lying {
				want = run{stdout: []byte("argument\n"), stderr: []byte(fmt.Sprintf(
					"adamic: panic: predicate assert used as a value, declared at %s:2:1, asserts branch: source unknown, target truthy\n",
					path,
				)), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(test, program)
			for _, got := range []run{sanitized, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatal(difference)
				}
			}
			if !probe.lying {
				if report := leaksUncached(test, program, binary); report != "" {
					test.Fatal(report)
				}
			}
		})
	}
}
