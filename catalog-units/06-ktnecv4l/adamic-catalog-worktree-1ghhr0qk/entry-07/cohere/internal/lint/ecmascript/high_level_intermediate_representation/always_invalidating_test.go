package high_level_intermediate_representation

import (
	"path/filepath"
	"testing"

	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestIsAlwaysInvalidatingTypeSeparatesTheFourShapes is the whole claim of this predicate.
//
// Upstream answers from a shape registry it maintains because Babel has no checker. This asks the
// resident checker instead, and the only thing that makes that substitution legitimate is that the
// checker separates the same set. So the set is asserted directly, by name, in both directions --
// every invalidating shape true and every cheap-to-compare one false.
//
// A test that only checked the true cases would pass on a predicate that returns true for
// everything, which is the failure this file is most likely to have.
func TestIsAlwaysInvalidatingTypeSeparatesTheFourShapes(t *testing.T) {
	t.Parallel()

	// The callable forms are here because the predicate is ONE flag check rather than upstream's
	// two arms. A separate call-signature arm was written, then removed when a mutation sweep
	// showed deleting it changed no answer: every callable form TypeScript has carries the Object
	// flag. These five hold that reduction in place -- if a callable form is ever found without the
	// flag, one of them fails and the arm comes back.
	answers := invalidatingAnswers(t, `
		interface CallableValue { (x: number): number; }
		type AliasedFunction = (x: number) => number;
		function f(flag, aliasedValue: AliasedFunction, callableValue: CallableValue,
				overloadedValue: {(): void; (x: number): void}) {
			const arrayValue = [1, 2, 3];
			const objectValue = {a: 1};
			const arrowValue = () => 1;
			function declaredValue() { return 1; }
			const numberValue = 42;
			const stringValue = "s";
			const booleanValue = true;
			const nullValue = null;
			return [arrayValue, objectValue, arrowValue, declaredValue, numberValue,
				stringValue, booleanValue, nullValue, flag, aliasedValue, callableValue,
				overloadedValue];
		}
	`)

	for _, testCase := range []struct {
		name string
		want bool
		why  string
	}{
		{name: "arrayValue", want: true, why: "an array literal allocates on every evaluation"},
		{name: "objectValue", want: true, why: "an object literal allocates on every evaluation"},
		{name: "arrowValue", want: true, why: "a closure allocates on every evaluation"},
		{name: "declaredValue", want: true, why: "a function declaration is still a function type"},
		{name: "aliasedValue", want: true, why: "a type-alias callback is callable and allocates"},
		{name: "callableValue", want: true, why: "a callable interface carries the Object flag"},
		{name: "overloadedValue", want: true, why: "an overloaded signature is still one object type"},
		{name: "numberValue", want: false, why: "a number is comparable with Object.is"},
		{name: "stringValue", want: false, why: "a string is comparable with Object.is"},
		{name: "booleanValue", want: false, why: "a boolean is comparable with Object.is"},
		{name: "nullValue", want: false, why: "null is comparable with Object.is"},
	} {
		got, found := answers[testCase.name]
		if !found {
			t.Errorf("%q never reached the predicate; the lowering did not give it an identifier "+
				"with a node, so this case asserts nothing", testCase.name)
			continue
		}
		if got != testCase.want {
			t.Errorf("%s: got %t, want %t (%s)", testCase.name, got, testCase.want, testCase.why)
		}
	}
}

// TestIsAlwaysInvalidatingTypeDeclinesWithoutAChecker pins the conservative direction.
//
// Both upstream call sites use a TRUE answer to permit a merge, so a missing type must answer false
// rather than true: declining to merge on missing information is safe, merging on it is not. A
// predicate that defaulted the other way would still pass the separation test above.
func TestIsAlwaysInvalidatingTypeDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `function f() { const a = [1]; return a; }`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	reached := 0
	for _, identifier := range function.Identifiers {
		if identifier == nil {
			continue
		}
		reached++
		if IsAlwaysInvalidatingType(function, identifier.Id, nil) {
			t.Errorf("identifier %q answered true with a nil checker; a missing type must decline",
				identifier.Name)
		}
	}
	if reached == 0 {
		t.Fatal("the function holds no identifiers, so the loop asserted nothing")
	}

	if IsAlwaysInvalidatingType(nil, 0, nil) {
		t.Error("a nil function answered true")
	}
}

// TestIsAlwaysInvalidatingTypeHandlesNodelessIdentifiers exercises the guard the hand-written cases
// cannot reach.
//
// Every identifier in a small fixture has a syntactic source, so the nil-node guard is unexercised
// there -- a mutation sweep confirmed it: deleting that guard passes every other test in this file.
// The corpus is where the case actually lives: 645 of 40,241 identifiers over 100 files carried no
// node, which is the 1.5% that `high_level_intermediate_representation.go` documents as having no direct syntactic source.
//
// The assertion is that the predicate answers rather than panicking, and that it answers false. A
// nodeless identifier has no type to ask about, and both upstream call sites use a true answer to
// PERMIT a merge, so the missing-information direction must be false.
func TestIsAlwaysInvalidatingTypeHandlesNodelessIdentifiers(t *testing.T) {
	t.Parallel()

	nodeless, answeredTrue := 0, 0

	forEachCorpusFunctionWithChecker(t, 100, func(function *Function, checker *shimchecker.Checker) {
		for _, identifier := range function.Identifiers {
			if identifier == nil || identifier.Node != nil {
				continue
			}
			nodeless++
			// The REAL checker is passed, which is the whole point. An earlier spelling passed nil
			// here and returned at the checker guard without ever reaching the node guard -- so it
			// asserted nothing about nodeless identifiers, and a mutation deleting the node guard
			// survived it. The mutant is caught only when the lookup is actually attempted.
			if IsAlwaysInvalidatingType(function, identifier.Id, checker) {
				answeredTrue++
			}
		}
	})

	if nodeless == 0 {
		// The count is not written into this message. A hardcoded figure beside the code that
		// computes it drifts as the corpus moves, and a stale number in the failure message of a
		// coverage guard is the worst place for one: it hands a future reader a phantom
		// discrepancy inside the very assertion meant to prove the guard is exercised.
		t.Fatal("no nodeless identifier was found in the corpus, so this guard is unexercised " +
			"and the test proves nothing")
	}
	if answeredTrue != 0 {
		t.Errorf("%d of %d nodeless identifiers answered true; a value with no type must decline, "+
			"because both call sites use true to permit a merge", answeredTrue, nodeless)
	}

	t.Logf("nodelessIdentifiers=%d answeredTrue=%d", nodeless, answeredTrue)
}

// invalidatingAnswers runs the predicate over a source and returns its answer per named identifier.
func invalidatingAnswers(t *testing.T, source string) map[string]bool {
	t.Helper()

	answers := map[string]bool{}
	probe := rule.Rule{
		Name:             "always-invalidating-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the typed harness handed this probe a nil checker, so every " +
							"answer below would be the nil-checker default rather than a type")
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						for _, identifier := range function.Identifiers {
							if identifier == nil || identifier.Name == "" {
								continue
							}
							answers[identifier.Name] = IsAlwaysInvalidatingType(
								function, identifier.Id, ctx.TypeChecker)
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"/invalidating.ts": source}, "/invalidating.ts")
	return answers
}

// forEachCorpusFunctionWithChecker is forEachCorpusFunction, keeping the checker.
//
// The shared helper drops it, and every consumer so far only needed the graph. This one needs the
// checker itself, because the predicate under test takes one and passing nil short-circuits before
// the code being asserted runs.
func forEachCorpusFunctionWithChecker(t *testing.T, limit int,
	visit func(function *Function, checker *shimchecker.Checker)) {
	t.Helper()

	// Frozen, so a count asserted over it moves only when the analysis does. See pinnedCorpusFiles.
	files, contentsOf := pinnedCorpusFiles(t, limit)

	for _, path := range files {
		contents := contentsOf[path]
		fileName := "/corpus/" + filepath.Base(path)
		probe := rule.Rule{
			Name:             "invalidating-corpus",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindSourceFile: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							t.Fatal("the typed harness handed this probe a nil checker")
						}
						forEachFunctionLike(node, func(functionNode *ast.Node) {
							function := Lower(functionNode, ctx.TypeChecker)
							if function == nil {
								return
							}
							Construct(function)
							visit(function, ctx.TypeChecker)
						})
					},
				}
			},
		}
		rule_testing.RunTypedFiles(t, probe, map[string]string{fileName: contents}, fileName)
	}
}
