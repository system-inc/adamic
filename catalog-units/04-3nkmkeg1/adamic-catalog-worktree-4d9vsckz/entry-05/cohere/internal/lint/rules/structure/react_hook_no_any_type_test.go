package structure

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// This rule has no upstream corpus on either side. It is a house rule Kirk asked for directly, so
// every fixture below was written from reasoning about our own code rather than imported from a
// reference implementation, which makes this a weaker artifact than a ported rule with an imported
// floor. The next reader should know which kind they are holding.
//
// # How both a resolvable and an unresolvable React live in one harness
//
// `rule_testing`'s tsconfig sets `types: []` and points at a fresh temp directory, so `@types/react`
// can never resolve in a fixture and there is no `node_modules` to reach. The passing case
// therefore supplies its own React through a sibling declaration file using `declare module
// 'react'`, which is a global augmentation the program picks up from any file in the include set;
// `reactHookNoAnyTypeReactDeclarations` below is the minimum of `@types/react` this rule reads. The
// failing cases simply omit that file, which is what makes `import React from 'react'` resolve to
// nothing and every hook call come back `any`.
//
// So the two states are one map entry apart, and the same fixture text can be run both ways. That
// is used deliberately in `TestReactHookAnyTypeSeesTheDifferenceTheDeclarationsMake`, which is the
// only test here that proves the rule is measuring resolution rather than measuring syntax.
const reactHookNoAnyTypeReactDeclarations = `declare module 'react' {
  export type SetStateAction<S> = S | ((prev: S) => S);
  export type Dispatch<A> = (value: A) => void;
  export function useState<S>(initial?: S): [S, Dispatch<SetStateAction<S>>];
  export function useEffect(effect: () => void, deps?: unknown[]): void;
  export function useMemo<T>(compute: () => T, deps: unknown[]): T;
  export function useRef<T>(initial: T): { current: T };
  const React: {
    useState: typeof useState;
    useEffect: typeof useEffect;
    useMemo: typeof useMemo;
    useRef: typeof useRef;
  };
  export default React;
}`

// runReactHookAnyTypeWithReact runs a fixture with React's declarations resolvable.
func runReactHookAnyTypeWithReact(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, ReactHookNoAnyType, map[string]string{
		"react.d.ts":  reactHookNoAnyTypeReactDeclarations,
		"fixture.tsx": source,
	}, "fixture.tsx")
}

// runReactHookAnyTypeWithoutReact runs a fixture with nothing declaring React at all.
func runReactHookAnyTypeWithoutReact(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, ReactHookNoAnyType, map[string]string{
		"fixture.tsx": source,
	}, "fixture.tsx")
}

// TestReactHookAnyTypeFires covers every arrangement that costs the type-based rules their sight.
func TestReactHookAnyTypeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
		// withReact runs the case with the declarations present, for the arrangements that are
		// degraded even though React itself resolves.
		withReact bool
	}{
		{
			// Kirk's motivating case, verbatim in shape. `tsc` is satisfied by the `declare`, the
			// setter is called during render, and `set-state-in-render` is silent because `setCount`
			// is `any` rather than a `Dispatch`. The gate passes on a genuine render loop.
			name: "an ambient shim declaring useState as returning any",
			source: `declare function useState<T>(initial: T): any;
export function Widget() {
    const [count, setCount] = useState(0);
    setCount(count + 1);
    return count;
}`,
			want: []string{"reactHookNoAnyType"},
		},
		{
			// No import statement anywhere, which is what makes this rule and
			// `react-import-no-destructuring` non-overlapping: that rule has nothing to look at.
			name: "a bare hook call with nothing declaring it",
			source: `export function Widget() {
    const [count, setCount] = useState(0);
    return count;
}`,
			want: []string{"reactHookNoAnyType"},
		},
		{
			// House style. The import is written and looks correct; the module does not resolve, so
			// `React` is `any` and every member access off it is too. This is the shape a stale
			// `@types` package or a half-resolving path mapping produces.
			name: "the namespaced house style with react unresolvable",
			source: `import React from 'react';
export function Widget() {
    const [count, setCount] = React.useState(0);
    return count;
}`,
			want: []string{"reactHookNoAnyType"},
		},
		{
			// Two hooks, two findings. A file is not reported once and then let alone: each call is
			// its own lost analysis and a reader fixing this needs to see how many.
			name: "two degraded hook calls report twice",
			source: `import React from 'react';
export function Widget() {
    const [count, setCount] = React.useState(0);
    const memo = React.useMemo(function () { return count; }, [count]);
    return memo;
}`,
			want: []string{"reactHookNoAnyType", "reactHookNoAnyType"},
		},
		{
			// The parenthesized receiver. `react.IsNamespacedMember` skips parentheses, so this is
			// the same call, and a hand-rolled receiver test would decline it while looking correct.
			name: "a parenthesized React receiver is the same call",
			source: `import React from 'react';
export function Widget() {
    const [count, setCount] = (React).useState(0);
    return count;
}`,
			want: []string{"reactHookNoAnyType"},
		},
		{
			// React resolves, but a local ambient declaration shadows the hook and destroys its
			// return type. This is the arrangement the message's advice names, and it proves the
			// rule is reading the call rather than reading whether an import resolved.
			name: "a local shim shadowing a hook while react itself resolves",
			source: `declare function useThing(): any;
export function Widget() {
    const value = useThing();
    return value;
}`,
			want:      []string{"reactHookNoAnyType"},
			withReact: true,
		},
		{
			// A digit after the prefix. React's own pattern is `/^use[A-Z0-9]/` and
			// `react.IsHookName` on the shelf rejects this, which is why the rule carries its own
			// predicate. Kirk's tree has no such name today; this pins the behaviour before one
			// arrives.
			name: "a hook whose fourth character is a digit",
			source: `declare function use2Things(): any;
export function Widget() {
    return use2Things();
}`,
			want:      []string{"reactHookNoAnyType"},
			withReact: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.withReact {
				result = runReactHookAnyTypeWithReact(t, testCase.source)
			} else {
				result = runReactHookAnyTypeWithoutReact(t, testCase.source)
			}
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// TestReactHookAnyTypeStaysSilent covers healthy code and the near neighbours that must not report.
func TestReactHookAnyTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		withReact bool
	}{
		{
			// The whole point of the rule: a file whose React types are real reports nothing.
			name: "house style with react resolvable",
			source: `import React from 'react';
export function Widget() {
    const [count, setCount] = React.useState(0);
    React.useEffect(function () {}, []);
    return count;
}`,
			withReact: true,
		},
		{
			// `useEffect` returns `void`. A rule keyed on anything looser than `TypeFlagsAny` would
			// report this healthy call, so it is the negative control for the flag test.
			name: "a healthy useEffect returning void",
			source: `import React from 'react';
export function Widget() {
    React.useEffect(function () {}, []);
    return 1;
}`,
			withReact: true,
		},
		{
			// A deliberate `unknown` is an author refusing to guess, which is the opposite failure
			// from the checker giving up. Reported would make the rule a type-cleanliness nit.
			name: "a hook deliberately annotated as returning unknown",
			source: `declare function useDeliberate(): unknown;
export function Widget() {
    return useDeliberate();
}`,
			withReact: true,
		},
		{
			// `never` is deliberate too.
			name: "a hook annotated as returning never",
			source: `declare function useNeverReturns(): never;
export function Widget() {
    useNeverReturns();
    return 1;
}`,
			withReact: true,
		},
		{
			// The precision claim. `any` density on production trees is 1.8% to 2.9%, so a rule
			// reporting every `any`-typed call would be a noise generator. This call produces `any`
			// and is silent because it is not hook-shaped.
			name: "a non-hook call returning any",
			source: `declare function used(x: number): any;
export function Widget() {
    used(1);
    return 1;
}`,
			withReact: true,
		},
		{
			// A method on an unrelated object. Treating every namespaced `use*` as a React hook
			// would report code that has nothing to do with React, which is the judgment
			// `react_detection.go`'s `isHookCall` makes for the same reason.
			name: "a use-prefixed method on a non-React receiver",
			source: `declare const somethingElse: any;
export function Widget() {
    somethingElse.useCache('key');
    return 1;
}`,
			withReact: true,
		},
		{
			// A non-hook member on the React namespace itself, with react unresolvable so the member
			// resolves to `any` exactly as a hook would. The receiver is right and the name is not,
			// so only the name test can decline it.
			//
			// Added for a surviving mutant: replacing the predicate handed to
			// `react.IsNamespacedMember` with an always-true function passed the entire suite,
			// because every other namespaced fixture here carries a real hook name and could not
			// tell the two versions apart. Measured before writing: pristine reports 0 on this
			// input and the mutant reports 1.
			name: "a non-hook member on the React namespace",
			source: `import React from 'react';
export function Widget() {
    return React.createElement('div');
}`,
		},
		{
			// `user` is not a hook: the fourth character is lowercase. React's pattern requires an
			// uppercase letter or a digit, so this is somebody's ordinary function.
			name: "a use-prefixed name whose fourth character is lowercase",
			source: `declare function user(): any;
export function Widget() {
    user();
    return 1;
}`,
			withReact: true,
		},
		{
			// A name starting `us` but not `use`, with an uppercase fourth character so every part
			// of the pattern except the prefix is satisfied. Added for a surviving mutant that
			// shortened the prefix comparison to `us`: every other silent fixture here fails the
			// pattern on the fourth character rather than on the prefix, so none could see it.
			// Measured before writing: pristine reports 0 and the mutant reports 1.
			name: "a name starting us rather than use",
			source: `declare function usbPort(): any;
export function Widget() {
    usbPort();
    return 1;
}`,
			withReact: true,
		},
		{
			// A bare `use` is too short for the pattern. React's `use` hook itself is spelled `use`
			// and does not match `/^use[A-Z0-9]/`, which is upstream's own behaviour rather than an
			// omission here.
			name: "a bare use call",
			source: `declare function use(): any;
export function Widget() {
    use();
    return 1;
}`,
			withReact: true,
		},
		{
			// Non-ASCII uppercase. React's character class is `[A-Z0-9]`, so `useÉ` is not a hook
			// name, and a `unicode.IsUpper` test would answer that it is.
			name: "a hook name with a non-ascii fourth character",
			source: `declare function useÉ(): any;
export function Widget() {
    useÉ();
    return 1;
}`,
			withReact: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.withReact {
				result = runReactHookAnyTypeWithReact(t, testCase.source)
			} else {
				result = runReactHookAnyTypeWithoutReact(t, testCase.source)
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestReactHookAnyTypeSpan pins where the finding points.
//
// `ExpectFindings` asserts message ids and count and nothing else, so a rule that reports the right
// number of findings at the wrong node passes a complete fixture pair while being wrong. The finding
// is anchored on the whole call expression rather than on the callee, because the call is what
// produced the untyped value and it is the range a reader has to look at to see the shape.
//
// The source is sliced with the finding's own range rather than compared against a literal offset,
// and the source is taken from the same constant the harness was handed. That matters here:
// `rule_testing.RunTypedFiles` writes `strings.TrimSpace(contents)+"\n"`, so a fixture with leading
// whitespace sits one byte off its Go literal and a hand-computed offset reports the wrong text.
// Trimming the same way is what keeps the two aligned.
func TestReactHookAnyTypeSpan(t *testing.T) {
	t.Parallel()

	source := `import React from 'react';
export function Widget() {
    const [count, setCount] = React.useState(0);
    return count;
}`
	result := runReactHookAnyTypeWithoutReact(t, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
	}

	written := strings.TrimSpace(source) + "\n"
	found := result.Diagnostics[0]
	reported := written[found.Range.Pos():found.Range.End()]
	if reported != "React.useState(0)" {
		t.Errorf("the finding covers %q, want %q", reported, "React.useState(0)")
	}
}

// TestReactHookAnyTypeSpanOnABareCall pins the other callee spelling.
func TestReactHookAnyTypeSpanOnABareCall(t *testing.T) {
	t.Parallel()

	source := `declare function useState<T>(initial: T): any;
export function Widget() {
    const [count, setCount] = useState(0);
    return count;
}`
	result := runReactHookAnyTypeWithoutReact(t, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
	}

	written := strings.TrimSpace(source) + "\n"
	found := result.Diagnostics[0]
	reported := written[found.Range.Pos():found.Range.End()]
	if reported != "useState(0)" {
		t.Errorf("the finding covers %q, want %q", reported, "useState(0)")
	}
}

// TestReactHookAnyTypeMessageNamesTheHook asserts the rendered text rather than the id.
//
// The description is built with concatenation per finding, so the id assertion cannot see anything
// the builder does. A rule rendering the wrong hook name, or dropping the cost clause entirely,
// passes every fixture above. The assertions here are equality and exact prefixes rather than
// `strings.Contains` on an interpolated value, because a weaker predicate than the property it
// guards is not a guard: a doubled backtick or a wrong name still satisfies a substring test.
//
// Asserted against literal strings typed here rather than against the rule's own builder, so that a
// mutation moving the text fails rather than moving both sides together.
func TestReactHookAnyTypeMessageNamesTheHook(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     string
		wantOpener string
	}{
		{
			name: "a namespaced hook is named with its receiver",
			source: `import React from 'react';
export function Widget() {
    return React.useState(0);
}`,
			wantOpener: "`React.useState` here resolves to `any`, so the type checker cannot see what this hook returns. ",
		},
		{
			name: "a bare hook is named as written",
			source: `declare function useThing(): any;
export function Widget() {
    return useThing();
}`,
			wantOpener: "`useThing` here resolves to `any`, so the type checker cannot see what this hook returns. ",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runReactHookAnyTypeWithoutReact(t, testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
			}
			got := result.Diagnostics[0].Message.Description
			if !strings.HasPrefix(got, testCase.wantOpener) {
				t.Errorf("message opens with %q,\nwant it to open with %q", got, testCase.wantOpener)
			}
			if result.Diagnostics[0].Message.Id != "reactHookNoAnyType" {
				t.Errorf("message id = %q, want %q", result.Diagnostics[0].Message.Id, "reactHookNoAnyType")
			}
		})
	}
}

// reactHookNoAnyTypeStandInBlindedRule is a rule registered by this test file for one purpose: to put
// something in the live catalog that declares `ResolvesReactValueTypes`, so the populated branch of
// the message can be proven.
//
// It exists because of a real difference between two binaries rather than to fake one.
// `rule.Registered()` returns what the linked binary registered, and this package's test binary
// links only `structure` rules, none of which declare the flag. The rules that DO declare it —
// `set-state-in-render` and `set-state-in-effect` — live in `internal/rules/react`, and importing
// that package here would both create a dependency this package does not want and be refused by the
// leaf guard's spirit. So the branch is proven with a stand-in, and the empty branch is proven
// separately below by asserting on what this binary genuinely has.
//
// The name is deliberately not one of the real rules'. A stand-in wearing a real rule's name would
// make this test pass whether or not the derivation reads the flag at all.
var reactHookNoAnyTypeStandInBlindedRule = rule.Rule{
	Name:                    "react-hook-any-type-stand-in-blinded",
	NeedsTypeChecker:        true,
	ResolvesReactValueTypes: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Never registered with the real catalog and never offered a file; this body exists only so
		// the struct is a rule rather than a shape.
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{ast.KindCallExpression: func(node *ast.Node) {}}
	},
}

func init() {
	rule.Register(rule.Registration{Rule: reactHookNoAnyTypeStandInBlindedRule})
}

// TestReactHookAnyTypeNamesTheBlindedRulesFromTheCatalog proves the list is derived, not written.
//
// The assertion that matters is not "the message contains a rule name" but "the message contains
// the name of a rule that declared the flag, and does not contain the name of one that did not".
// A hardcoded string would satisfy the first and fail the second.
//
// `react-hook-no-destructuring` is the negative control: it is registered in this same binary, in
// this same package, and declares neither `NeedsTypeChecker` nor the flag. If the derivation were
// reading the whole catalog rather than the flag, its name would appear.
func TestReactHookAnyTypeNamesTheBlindedRulesFromTheCatalog(t *testing.T) {
	t.Parallel()

	blinded := reactHookNoAnyTypeBlindedRules()
	if len(blinded) == 0 {
		t.Fatal("the stand-in rule declaring ResolvesReactValueTypes is not in the catalog, so " +
			"this test would prove the empty branch twice rather than the populated one")
	}

	source := `declare function useState<T>(initial: T): any;
export function Widget() {
    return useState(0);
}`
	result := runReactHookAnyTypeWithoutReact(t, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
	}
	got := result.Diagnostics[0].Message.Description

	if !strings.Contains(got, "react-hook-any-type-stand-in-blinded") {
		t.Errorf("the message does not name the rule that declared ResolvesReactValueTypes;\ngot %q", got)
	}
	// The negative control. Registered in this binary, declares nothing, must not be named.
	if strings.Contains(got, "react-hook-no-destructuring") {
		t.Errorf("the message names a rule that declared nothing, so the list is not being derived "+
			"from the flag;\ngot %q", got)
	}
	if !strings.Contains(got, "Suppressing this finding turns those off for the file") {
		t.Errorf("the message does not name the cost of suppression;\ngot %q", got)
	}
}

// TestReactHookAnyTypeBlindedRuleDerivationReadsTheFlag asserts the helper directly.
//
// Separate from the message test because the two can fail for different reasons and reporting one
// defect under two names wastes a reader's time. This one says the derivation reads the flag; that
// one says the message renders what the derivation returned.
func TestReactHookAnyTypeBlindedRuleDerivationReadsTheFlag(t *testing.T) {
	t.Parallel()

	blinded := reactHookNoAnyTypeBlindedRules()

	named := map[string]bool{}
	for _, name := range blinded {
		named[name] = true
	}

	for _, registration := range rule.Registered() {
		want := registration.Rule.ResolvesReactValueTypes
		if got := named[registration.Rule.Name]; got != want {
			t.Errorf("rule %q declares ResolvesReactValueTypes=%v but the derivation %s it",
				registration.Rule.Name, want, map[bool]string{true: "named", false: "did not name"}[got])
		}
	}
}

// TestReactHookAnyTypeRequiresTheTypedHarness asserts the rule is useless without a checker.
//
// A typed rule handed a nil checker goes silent, and every StaysSilent case above would then pass
// vacuously while proving nothing. This makes that state a failure rather than a green suite, so a
// later revert of `NeedsTypeChecker` or of the guard fails loudly instead of quietly.
func TestReactHookAnyTypeRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := `declare function useState<T>(initial: T): any;
export function Widget() {
    return useState(0);
}`

	typed := runReactHookAnyTypeWithoutReact(t, source)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness produced %d findings, want 1", len(typed.Diagnostics))
	}

	// The untyped harness hands the rule a nil checker, which it must decline rather than crash on.
	untyped := rule_testing.Run(t, ReactHookNoAnyType, "fixture.tsx", source)
	if len(untyped.Diagnostics) != 0 {
		t.Errorf("the untyped harness produced %d findings, want 0; this rule cannot answer without "+
			"a checker and must decline rather than guess", len(untyped.Diagnostics))
	}
}

// TestReactHookAnyTypeSeesTheDifferenceTheDeclarationsMake is the test that proves the rule measures
// resolution rather than syntax.
//
// One source, run twice, differing only in whether React's declarations are in the program. Every
// other test here fixes one side of that and could pass over a rule keyed on something else
// entirely — a rule reporting on the mere absence of a sibling file, or on the string `useState`,
// would satisfy them all. This one cannot be satisfied that way, because the bytes of the subject
// file are identical in both runs.
func TestReactHookAnyTypeSeesTheDifferenceTheDeclarationsMake(t *testing.T) {
	t.Parallel()

	source := `import React from 'react';
export function Widget() {
    const [count, setCount] = React.useState(0);
    return count;
}`

	withReact := runReactHookAnyTypeWithReact(t, source)
	if len(withReact.Diagnostics) != 0 {
		t.Errorf("with React resolvable the rule produced %d findings, want 0", len(withReact.Diagnostics))
	}

	withoutReact := runReactHookAnyTypeWithoutReact(t, source)
	if len(withoutReact.Diagnostics) != 1 {
		t.Errorf("with React unresolvable the rule produced %d findings, want 1", len(withoutReact.Diagnostics))
	}
}

// TestReactHookAnyTypeJoinRendersEveryArity tests the joiner directly rather than through the rule.
//
// It has to be direct, and the reason is a reachability limit rather than a preference. The joiner's
// three-or-more arm can only be exercised by a catalog holding three rules that declare
// `ResolvesReactValueTypes`, and today the tree holds two. Reached through the rule, that arm is
// unreachable, and a mutation removing its `and` separator SURVIVED the entire suite for exactly
// that reason — not because the fixtures were blind, but because no input could reach the code.
//
// A third such rule is expected rather than hypothetical, which is what makes testing the arm now
// worth doing: the alternative is an arm that silently renders wrong prose on the day it first runs.
// Asserted against literal strings typed here rather than against the joiner's own output, so a
// mutation moves one side and not both.
func TestReactHookAnyTypeJoinRendersEveryArity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		names []string
		want  string
	}{
		{name: "none", names: nil, want: ""},
		{name: "one", names: []string{"a"}, want: "`a`"},
		{name: "two", names: []string{"a", "b"}, want: "`a` and `b`"},
		{name: "three", names: []string{"a", "b", "c"}, want: "`a`, `b`, and `c`"},
		{name: "four", names: []string{"a", "b", "c", "d"}, want: "`a`, `b`, `c`, and `d`"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := reactHookNoAnyTypeJoin(testCase.names); got != testCase.want {
				t.Errorf("reactHookNoAnyTypeJoin(%q) = %q, want %q", testCase.names, got, testCase.want)
			}
		})
	}
}

// TestReactHookAnyTypeFlagReadAgreesWithTheCheckersOwnName guards the unsafe field read.
//
// The whole rule turns on `shimchecker.Type_flags`, which is not a method call: it casts a
// `*checker.Type` to a hand-mirrored struct in `shim/checker/shim.go` and reads a field by offset.
// That is correct only while the mirror reproduces upstream's layout exactly, and it does not fail
// loudly when it stops. `Checker_numberType` once returned the NULL type because a substituted
// stand-in was 24 bytes short at field 99 of 319, and every package compiled and every test passed;
// the only symptom was `no-for-in-array` silently declining its entire corpus.
//
// Two properties of that defect set the bar this test has to clear. A nil check cannot see it,
// because the wrong value was non-nil and a perfectly valid value of the right Go type. And a
// field-name comparison cannot see it either, because all 319 names aligned and only a type was
// wrong. So this asserts what the value IS: for every subject the flag read is compared against the
// checker's own name for the same type, which is derived through a completely different path
// (`TypeToString`, a linkname to an upstream method, no offset arithmetic anywhere).
//
// `flags` happens to be field zero of the mirror, which makes it the least likely field in the file
// to drift. That is a reason to expect this test to pass, not a reason to skip writing it: an
// upstream reordering that moves `flags` off zero would silently turn this rule into one that
// reports nothing, or reports on everything, and nothing else in the tree would notice.
func TestReactHookAnyTypeFlagReadAgreesWithTheCheckersOwnName(t *testing.T) {
	t.Parallel()

	subjects := []struct {
		name string
		// source declares a hook whose return type is spelled in `wantTypeName`.
		source string
		// wantTypeName is what the checker calls the type the hook call produces.
		wantTypeName string
		// wantAny is whether TypeFlagsAny should be set, which must agree with the name.
		wantAny bool
	}{
		{name: "any", source: `declare function useSubject(): any;`, wantTypeName: "any", wantAny: true},
		{name: "unknown", source: `declare function useSubject(): unknown;`, wantTypeName: "unknown", wantAny: false},
		{name: "void", source: `declare function useSubject(): void;`, wantTypeName: "void", wantAny: false},
		{name: "never", source: `declare function useSubject(): never;`, wantTypeName: "never", wantAny: false},
		{name: "number", source: `declare function useSubject(): number;`, wantTypeName: "number", wantAny: false},
	}

	for _, subject := range subjects {
		t.Run(subject.name, func(t *testing.T) {
			t.Parallel()
			var gotName string
			var gotAny bool
			var visited int

			// A throwaway rule rather than the real one, so this measures the shim read directly
			// instead of measuring it through everything else the rule decides.
			probe := rule.Rule{
				Name:             "react-hook-any-type-shim-probe",
				NeedsTypeChecker: true,
				Run: func(ctx rule.Context, options any) rule.Listeners {
					if ctx.TypeChecker == nil {
						return nil
					}
					return rule.Listeners{
						ast.KindCallExpression: func(node *ast.Node) {
							resultType := ctx.TypeChecker.GetTypeAtLocation(node)
							if resultType == nil {
								return
							}
							visited++
							gotName = ctx.TypeChecker.TypeToString(resultType)
							gotAny = shimchecker.Type_flags(resultType)&shimchecker.TypeFlagsAny != 0
						},
					}
				},
			}

			source := subject.source + `
export function Widget() {
    return useSubject();
}`
			rule_testing.RunTypedFiles(t, probe, map[string]string{"fixture.tsx": source}, "fixture.tsx")

			if visited != 1 {
				// A measurement that visited nothing reads exactly like agreement, which is the
				// failure this whole file is about.
				t.Fatalf("the probe visited %d call expressions, want 1; nothing was measured", visited)
			}
			if gotName != subject.wantTypeName {
				t.Errorf("the checker calls this type %q, want %q; the fixture is not producing "+
					"the type this case is about", gotName, subject.wantTypeName)
			}
			if gotAny != subject.wantAny {
				t.Errorf("shimchecker.Type_flags says any=%v for a type the checker itself calls "+
					"%q, want any=%v; the mirrored struct in shim/checker/shim.go is reading a "+
					"field other than the one it names", gotAny, gotName, subject.wantAny)
			}
		})
	}
}

// The branch no test binary reaches through the catalog: a build that linked none of the rules says so
// instead of naming an empty list. Both of the message's options are exercised, this one here.
func TestReactHookAnyTypeSaysWhenNoBlindedRuleIsLinked(t *testing.T) {
	t.Parallel()

	got := reactHookNoAnyTypeDescriptionFor("useThing", nil)
	want := "`useThing` here resolves to `any`, so the type checker cannot see what this hook returns. " +
		"Every rule that identifies a React value by its type is disabled for this file, and this build linked " +
		"none of them so none can be named. React's types are probably not resolving here: check that " +
		"`@types/react` is installed and current, that no ambient declaration is shadowing it, and that the " +
		"path mapping for `react` resolves."
	if got != want {
		t.Errorf("reads %q, want %q", got, want)
	}
}
