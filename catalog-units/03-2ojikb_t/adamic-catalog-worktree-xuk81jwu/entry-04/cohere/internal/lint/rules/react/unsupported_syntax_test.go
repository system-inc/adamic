package react

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// A `.tsx` name, because the gate this rule turns on is "does this function create JSX", and
// every case below either writes JSX or deliberately withholds it. Parsing these as plain
// TypeScript would read `<div>` as a type assertion, which would not merely change a span, it
// would remove the evidence the rule reads and silence most of the reporting cases.
const unsupportedSyntaxFile = "unsupported_syntax.tsx"

// upstreamEvalFixture is `error.invalid-eval-unsupported.js`, React's only error-named fixture
// for this category, transcribed verbatim.
//
// Its golden reads "Compilation Skipped: The 'eval' function is not supported" at 2:2 in
// upstream's zero-based column numbering, with a four-caret span under `eval`.
// TestUnsupportedSyntaxTranscriptionMatchesTheVendoredCorpus compares these bytes against the
// vendored file rather than trusting that this was copied correctly.
const upstreamEvalFixture = `function Component(props) {
  eval('props.x = true');
  return <div />;
}
`

// TestUnsupportedSyntaxFiresOnUpstreamCorpus runs React's own error fixture.
//
// One fixture, which is the whole imported reporting corpus. React ships no fixture writing a
// `with` statement or an inline class declaration inside a component, checked by grepping the
// full 1,835-input fixture list at the pinned sha rather than only the 453 vendored here.
func TestUnsupportedSyntaxFiresOnUpstreamCorpus(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, UnsupportedSyntax, unsupportedSyntaxFile, upstreamEvalFixture)
	rule_testing.ExpectFindings(t, result, "unsupportedEval")

	// The span is asserted against a literal typed here rather than against the rule's own
	// message constant, so a mutation moving both sides together cannot stay green.
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	reported := upstreamEvalFixture[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "eval" {
		t.Errorf("finding points at %q, want %q", reported, "eval")
	}
	if result.Diagnostics[0].Message.Id != "unsupportedEval" {
		t.Errorf("message id is %q, want %q", result.Diagnostics[0].Message.Id, "unsupportedEval")
	}
}

// TestUnsupportedSyntaxStaysSilentOnUpstreamCorpus runs the one imported case that looks like a
// violation and is not.
//
// `error.todo-kitchensink.js` declares `class Bar {}` inside `function foo`. The declaration is
// exactly the shape the inline-class arm reports, and it is clean because `foo` is lowercase, so
// the gate declines the whole function. Its own golden reports an `Invariant` diagnostic about a
// `for` statement, a different category entirely, and running it through React's rule produces
// nothing at all for this one.
//
// The file is read from the vendored corpus rather than transcribed. It is 1,061 bytes carrying
// template literals and escape sequences, and the port brief's rule is that the tool writing a
// fixture can cook an escape without anything going red. Reading the bytes removes the
// transcription step instead of verifying it.
func TestUnsupportedSyntaxStaysSilentOnUpstreamCorpus(t *testing.T) {
	t.Parallel()

	path := filepath.Join("conformance", "testdata", "fixtures", "error.todo-kitchensink.js")
	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the vendored fixture: %v", err)
	}
	// A control alongside the zero this asserts: the fixture must actually contain the
	// declaration, or the clean verdict is about an empty file rather than about the gate.
	if !strings.Contains(string(sourceBytes), "class Bar {") {
		t.Fatal("the vendored kitchensink fixture no longer declares an inline class, so this case tests nothing")
	}
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, UnsupportedSyntax, unsupportedSyntaxFile, string(sourceBytes)))
}

// TestUnsupportedSyntaxFires covers the reporting cases upstream does not ship.
//
// Every source below was run through React 7.1.1's own `unsupported-syntax` rule via the ESLint
// Linter API before being written here, and the expected findings are that run's output rather
// than a prediction from reading the bundle. Three of them contradicted what the source alone
// suggested, and those are called out at the case.
func TestUnsupportedSyntaxFires(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			// The `with` arm. Our parser produces a live KindWithStatement in .tsx where oxc's refuses
			// the file outright, so this is reachable here and was not reachable through oxlint on
			// TypeScript input. React reports it at columns 3 through 18.
			name: "withInComponent",
			sourceText: `function Component(props) {
  with (props) {}
  return <div />;
}
`,
			wantIds: []string{"unsupportedWith"},
		},
		{
			// The inline-class arm, upstream's `case 'ClassDeclaration'` in the statement dispatch.
			name: "inlineClassInComponent",
			sourceText: `function Component(props) {
  class Foo {}
  return <div />;
}
`,
			wantIds: []string{"unsupportedInlineClass"},
		},
		{
			// `eval` fires on the lowered identifier rather than on a call, so a bare reference reports
			// with no invocation anywhere. A rule keyed on KindCallExpression would miss this.
			name: "evalBareReference",
			sourceText: `function Component(props) {
  const runner = eval;
  return <div>{String(runner)}</div>;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// The hook branch of the gate. A `use`-prefixed name needs evidence but not component-shaped
			// parameters, which is why a single non-props parameter does not disqualify it.
			name: "evalInHookWithJsx",
			sourceText: `function useThing(a) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// React's hook pattern is `/^use[A-Z0-9]/`, so a digit after the prefix is a hook. The shelf's
			// `react.IsHookName` rejects this, which is why this rule does not use it.
			name: "evalInDigitHook",
			sourceText: `function use2Things(a) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// The variable-declarator naming path, the second of the two forms this port recognizes.
			name: "evalInArrowComponent",
			sourceText: `const Component = (props) => {
  eval('x');
  return <div />;
};
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// React compiles the outermost qualifying function and lowers its whole body, so a nested
			// plain function's `eval` still reports and the finding points inside the nest.
			name: "evalInNestedFunction",
			sourceText: `function Component(props) {
  function inner() {
    eval('x');
  }
  return <div onClick={inner} />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// A second parameter is allowed when its name contains `ref`, which is upstream's own
			// `name.includes('ref') || name.includes('Ref')`.
			name: "evalWithRefSecondParameter",
			sourceText: `function Component(props, ref) {
  eval('x');
  return <div ref={ref} />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// An object type is a valid props annotation, so the parameter shape does not disqualify.
			// Paired with evalWithPrimitivePropsAnnotation below, which is the same shape and silent.
			name: "evalWithObjectPropsAnnotation",
			sourceText: `function Component(props: {a: number}) {
  eval('x');
  return <div>{props.a}</div>;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// Upstream's `returnsNonNode` reads as a component disqualifier and measurably is not one
			// here: this reports under React's own rule. Implementing that check would have made this
			// port silent on an input React reports, which is why it is deliberately absent.
			name: "evalReturningObjectLiteral",
			sourceText: `function Component(props) {
  eval('x');
  if (props.a) {
    return {};
  }
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// A hook call is evidence on its own, with no JSX anywhere in the function.
			name: "evalWithHookCallEvidence",
			sourceText: `function Component(props) {
  eval('x');
  const value = useState(0);
  return value;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// A plain function may sit BETWEEN the compilation root and the finding. `Outer` is the
			// root, `mid` is simply part of what it lowers, and the eval inside the innermost arrow
			// still reports. This is the case that rules out a "every enclosing function must
			// qualify" formulation of the gate, which an earlier version of this rule used.
			name: "componentRootWithPlainFunctionBetween",
			sourceText: `function Outer(props) {
  function mid() {
    const Inner = (p) => {
      eval('x');
      return <b />;
    };
    return Inner;
  }
  return <div>{mid()}</div>;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// The root may be a hook rather than a component, and then a component nested inside it
			// is judged too. Paired with componentNestedInsideHookWithoutEvidence below, which is
			// the same shape with the hook's evidence removed and is silent.
			name: "componentNestedInsideQualifyingHook",
			sourceText: `function useOuter() {
  const Component = (p) => {
    eval('x');
    return <div />;
  };
  return Component;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// An exported top-level arrow is still program scope. The export modifier wraps the
			// declaration without moving it, which is why the scope walk treats the variable
			// statement chain as transparent rather than as a boundary.
			name: "exportedArrowComponent",
			sourceText: `export const Component = (props) => {
  eval('x');
  return <div />;
};
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// A redundant parenthesis around the arrow. Our parser keeps a KindParenthesizedExpression
			// between the declarator and the function where Babel's tree has already dropped it, so
			// the name lookup has to walk out through it. Measured on React's rule, this reports,
			// and it was the ONE disagreement in a 111-input differential run before the skip was
			// added in both directions.
			name: "parenthesizedArrowComponent",
			sourceText: `const Component = ((props) => {
  eval('x');
  return <div />;
});
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// A type reference is a VALID props annotation, which is what makes the disqualifying
			// list a list rather than an "is it an object type" test. Paired with the annotation
			// cases in the silent suite, this is the half proving the list is not simply
			// rejecting everything annotated.
			name: "typeReferencePropsAnnotation",
			sourceText: `interface Props {
  a: number;
}
function Component(props: Props) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// The hook branch ignores parameter shape entirely, which upstream gets by asking
			// isValidComponentParams only inside the component branch. A rest parameter, three
			// parameters, and a primitive-annotated first parameter each disqualify a COMPONENT
			// and none of them disqualifies a hook. A mutation adding the parameter test to the
			// hook branch survived the whole suite until these four cases were added.
			name: "hookWithRestParameter",
			sourceText: `function useThing(...args) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			name: "hookWithThreeParameters",
			sourceText: `function useThing(a, b, c) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			name: "hookWithPrimitiveAnnotatedParameter",
			sourceText: `function useThing(a: string) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			name: "hookWithNonRefSecondParameter",
			sourceText: `function useThing(a, b) {
  eval('x');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// A namespaced hook call is evidence when the receiver is capitalized. Paired with
			// lowercaseNamespaceHookCall in the silent suite, which is the same call on a lowercase
			// receiver and is not evidence. A mutation dropping the receiver test survived until
			// both existed, because neither case alone can see it.
			name: "capitalizedNamespaceHookCall",
			sourceText: `function Component(props) {
  eval('x');
  Helpers.useThing(0);
  return 1;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// All three diagnostics in one body, in source order. The only fixture that proves the arms
			// are independent rather than one arm answering for all of them.
			name: "allThreeInOneComponent",
			sourceText: `function Component(props) {
  eval('x');
  with (props) {}
  class Foo {}
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval", "unsupportedWith", "unsupportedInlineClass"},
		},
		{
			// Two findings for two references. The rule reports per identifier, not per function.
			name: "twoEvalsInOneComponent",
			sourceText: `function Component(props) {
  eval('x');
  eval('y');
  return <div />;
}
`,
			wantIds: []string{"unsupportedEval", "unsupportedEval"},
		},
		{
			// A fragment is JSX evidence, which upstream's `JSX` traversal alias covers and an
			// element-kinds-only test would miss.
			name: "evalInFragmentComponent",
			sourceText: `function Component(props) {
  eval('x');
  return <></>;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
		{
			// The inner arrow is not named through a path this port recognizes, and the finding still
			// appears because the OUTER function qualifies and the walk goes outward until something
			// does. Without the outward walk this would be a false negative.
			name: "nestedComponentInsideComponent",
			sourceText: `function Outer(props) {
  const Inner = (p) => {
    eval('x');
    return <span />;
  };
  return <div>{Inner(props)}</div>;
}
`,
			wantIds: []string{"unsupportedEval"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, UnsupportedSyntax, unsupportedSyntaxFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// TestUnsupportedSyntaxStaysSilent covers the clean cases, which are where this rule's real
// difficulty lives.
//
// The three diagnostics are trivial to detect and the gate deciding whether to look at all is
// the entire rule, so most of these assert the gate rather than the diagnostics. Each was run
// through React's own rule and came back silent.
func TestUnsupportedSyntaxStaysSilent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
	}{
		{
			// A component-shaped function inside a bare block is not at program scope, so it is
			// never chosen as a compilation root. The identical function at program scope reports,
			// which is what makes this a measurement of position rather than of shape.
			name: "componentInsideBareBlock",
			sourceText: `{
  function Component(props) {
    eval('x');
    return <div />;
  }
}
`,
		},
		{
			// The same, through an `if` rather than a bare block.
			name: "componentInsideIfBlock",
			sourceText: `if (true) {
  function Component(props) {
    eval('x');
    return <div />;
  }
}
`,
		},
		{
			// The same, through a class method body. A method is not a program-scope position and
			// is also not one of the two naming forms this port recognizes.
			name: "componentInsideClassMethod",
			sourceText: `class Container {
  render() {
    function Component(props) {
      eval('x');
      return <div />;
    }
    return Component;
  }
}
`,
		},
		{
			// A hook with no evidence of its own is not a root, so a component nested inside it is
			// not reached either. Paired with componentNestedInsideQualifyingHook above: the only
			// difference is that the hook there returns something the evidence walk counts.
			name: "componentNestedInsideHookWithoutEvidence",
			sourceText: `function useOuter() {
  const Component = (p) => {
    eval('x');
    return 1;
  };
  return Component;
}
`,
		},
		{
			// The gate on the `with` arm specifically. Every other clean case here reaches the gate
			// through the eval arm, so a mutation removing the gate from THIS arm survived the
			// whole suite until this case was added. Measured silent under React's own rule.
			name: "withInPlainFunction",
			sourceText: `function helper(props) {
  with (props) {}
  return props.a;
}
`,
		},
		{
			// The same blind spot on the inline-class arm, found by the same mutation sweep.
			name: "inlineClassInPlainFunction",
			sourceText: `function helper(props) {
  class Foo {}
  return Foo;
}
`,
		},
		{
			// A tuple annotation. Each kind in upstream's disqualifying list is a separate
			// discrimination, and a mutation deleting only this one survived a suite whose only
			// annotation fixture was `string`, so the list is covered kind by kind below.
			name: "tuplePropsAnnotation",
			sourceText: `function Component(props: [number, string]) {
  eval('x');
  return <div />;
}
`,
		},
		{
			name: "arrayPropsAnnotation",
			sourceText: `function Component(props: number[]) {
  eval('x');
  return <div />;
}
`,
		},
		{
			name: "functionTypePropsAnnotation",
			sourceText: `function Component(props: () => void) {
  eval('x');
  return <div />;
}
`,
		},
		{
			name: "booleanPropsAnnotation",
			sourceText: `function Component(props: boolean) {
  eval('x');
  return <div />;
}
`,
		},
		{
			name: "literalTypePropsAnnotation",
			sourceText: `function Component(props: "a") {
  eval('x');
  return <div />;
}
`,
		},
		{
			// A hook-named property on a lowercase receiver is not a hook call, so this function
			// has no evidence and is not a component. See capitalizedNamespaceHookCall above.
			name: "lowercaseNamespaceHookCall",
			sourceText: `function Component(props) {
  eval('x');
  helpers.useThing(0);
  return 1;
}
`,
		},
		{
			// The gate, stated at its simplest. A lowercase function is not a component, so nothing in
			// it is judged. Without this the rule would report every `eval` in the tree.
			name: "evalInPlainFunction",
			sourceText: `function helper(a) {
  eval('x');
  return a;
}
`,
		},
		{
			// Module scope has no enclosing function at all.
			name: "evalAtModuleScope",
			sourceText: `eval('x');
`,
		},
		{
			// A capitalized name is a claim, not evidence. Without JSX or a hook call the function is
			// not inferred as a component, so this is silent even though the name qualifies.
			name: "evalInCapitalizedNonComponent",
			sourceText: `function Component(props) {
  eval('x');
  return 1;
}
`,
		},
		{
			// The same requirement on the hook branch: a `use` name with no evidence is not a hook.
			name: "evalInHookWithoutEvidence",
			sourceText: `function useThing(a) {
  eval('x');
  return a;
}
`,
		},
		{
			// `eval` resolves to the parameter rather than to the global, so nothing is lowered as a
			// global load. This is why the rule asks the checker instead of matching the name.
			name: "evalShadowedByParameter",
			sourceText: `function Component(eval) {
  eval('x');
  return <div />;
}
`,
		},
		{
			// The same resolution question through a local binding rather than a parameter.
			name: "evalShadowedByLocal",
			sourceText: `function Component(props) {
  const eval = (s: string) => s;
  eval('x');
  return <div />;
}
`,
		},
		{
			// Both upstreams dispatch on the STATEMENT, so a class expression is untouched. This is the
			// pass case that separates the declaration arm from a naive `class` keyword search.
			name: "classExpressionInComponent",
			sourceText: `function Component(props) {
  const Foo = class {};
  return <div>{String(Foo)}</div>;
}
`,
		},
		{
			// Module scope is exactly where upstream's message tells the author to move the class to.
			name: "classAtModuleScope",
			sourceText: `class Foo {}
`,
		},
		{
			// React's component pattern is the ASCII `/^[A-Z]/`, so an accented capital is not a
			// component name. The shelf's `react.IsLikelyComponentName` uses unicode.IsUpper and answers
			// true here, which is why this rule does not use it.
			name: "evalInAccentedName",
			sourceText: `function Émile(props) {
  eval('x');
  return <div />;
}
`,
		},
		{
			// A rest first parameter cannot be props.
			name: "evalWithRestParameter",
			sourceText: `function Component(...props) {
  eval('x');
  return <div />;
}
`,
		},
		{
			// Three or more parameters is never a component shape.
			name: "evalWithThreeParameters",
			sourceText: `function Component(a, b, c) {
  eval('x');
  return <div />;
}
`,
		},
		{
			// A second parameter whose name carries no `ref` disqualifies. Paired with
			// evalWithRefSecondParameter above, which is the same arity and reports.
			name: "evalWithNonRefSecondParameter",
			sourceText: `function Component(props, other) {
  eval('x');
  return <div>{String(other)}</div>;
}
`,
		},
		{
			// A primitive props annotation disqualifies, from upstream's own list of annotation kinds.
			name: "evalWithPrimitivePropsAnnotation",
			sourceText: `function Component(props: string) {
  eval('x');
  return <div>{props}</div>;
}
`,
		},
		{
			// Upstream's evidence walk installs skip handlers on nested functions, so JSX that only
			// appears inside a closure is not evidence for the enclosing function.
			name: "jsxOnlyInNestedFunction",
			sourceText: `function Component(props) {
  eval('x');
  const inner = () => <div />;
  return inner;
}
`,
		},
		{
			// The same skip, through the hook-call half of the evidence test rather than the JSX half.
			name: "hookCallOnlyInNestedFunction",
			sourceText: `function Component(props) {
  eval('x');
  const inner = () => useState(0);
  return inner;
}
`,
		},
		{
			// Bare `use` fails `/^use[A-Z0-9]/`, so it is not a hook call to the React Compiler. The
			// neighbouring `isHookCallee` in rules_of_hooks.go accepts it, correctly for that rule's own
			// upstreams, which is why this rule carries its own predicate.
			name: "bareUseCallIsNotEvidence",
			sourceText: `function Component(props) {
  eval('x');
  use(props.p);
  return 1;
}
`,
		},
		{
			// The same divergence through the function's own name rather than through a call it makes.
			name: "functionNamedUse",
			sourceText: `function use(a) {
  eval('x');
  return <div />;
}
`,
		},
		{
			// A property named `eval` is not a reference to the global, in either the object literal or
			// the member access.
			name: "evalAsPropertyName",
			sourceText: `function Component(props) {
  const table = {eval: 1};
  return <div>{table.eval}</div>;
}
`,
		},
		{
			// A component-shaped arrow nested inside a plain function is silent upstream, so the outward
			// walk qualifying an outer function does not mean an inner one is judged on its own.
			name: "componentNestedInPlainFunction",
			sourceText: `function outer() {
  const Component = (props) => {
    eval('x');
    return <div />;
  };
  return Component;
}
`,
		},
		{
			// Upstream has a `memo` callback path and it measurably does not produce this diagnostic, so
			// the path is not reproduced. Recorded as a fixture rather than left as an assumption.
			name: "memoWrappedArrow",
			sourceText: `const Component = memo((props) => {
  eval('x');
  return <div />;
});
`,
		},
		{
			// The object-property naming path, likewise measured silent and likewise not reproduced.
			name: "objectPropertyComponent",
			sourceText: `const namespace = {
  Component: (props) => {
    eval('x');
    return <div />;
  },
};
`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, UnsupportedSyntax, unsupportedSyntaxFile, testCase.sourceText))
		})
	}
}

// TestUnsupportedSyntaxTranscriptionMatchesTheVendoredCorpus compares the one transcribed upstream
// fixture against its source bytes.
//
// The port brief's rule is that reading a copied fixture is not enough, because the tool writing it
// can cook an escape and the result still compiles and still goes green. This compares bytes.
//
// The comparison is possible at all only because the corpus is vendored into this repository. A
// rule whose upstream fixtures were fetched at write time would have nothing to compare against
// later, which is the argument for the vendoring rather than an incidental convenience of it.
func TestUnsupportedSyntaxTranscriptionMatchesTheVendoredCorpus(t *testing.T) {
	t.Parallel()

	path := filepath.Join("conformance", "testdata", "fixtures", "error.invalid-eval-unsupported.js")
	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the vendored fixture: %v", err)
	}
	if string(sourceBytes) != upstreamEvalFixture {
		t.Errorf("transcription drifted from the vendored corpus\n vendored: %q\n     here: %q",
			string(sourceBytes), upstreamEvalFixture)
	}
}

// TestUnsupportedSyntaxRequiresTheTypedHarness pins that the checker is genuinely load-bearing.
//
// The port brief records that a rule guarding on a nil checker goes vacuously silent under the
// plain harness, which looks identical to a rule that is simply not firing, and that three shipped
// typed rules were found with no guard at all. This asserts the guard is present by running an
// input that reports under `RunTyped` through `Run` and requiring silence, so a later revert that
// drops `NeedsTypeChecker` fails here rather than quietly halving the rule.
//
// The `with` and inline-class arms are deliberately not covered by this: neither consults the
// checker, so both still report under the plain harness, and asserting silence for them would
// assert the opposite of the truth.
func TestUnsupportedSyntaxRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	// The control first. Under the typed harness this input reports, which is what makes the
	// silence below a measurement of the harness rather than of the input.
	typed := rule_testing.RunTyped(t, UnsupportedSyntax, unsupportedSyntaxFile, upstreamEvalFixture)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("typed harness: want 1 diagnostic, got %d", len(typed.Diagnostics))
	}

	untyped := rule_testing.Run(t, UnsupportedSyntax, unsupportedSyntaxFile, upstreamEvalFixture)
	if len(untyped.Diagnostics) != 0 {
		t.Errorf("plain harness: want 0 diagnostics from the guarded eval arm, got %d",
			len(untyped.Diagnostics))
	}
}

// TestUnsupportedSyntaxPointsAtTheRightNode asserts where each of the three findings lands.
//
// `ExpectFindings` asserts message ids and count and nothing else, so a rule reporting the correct
// finding at the wrong node passes a complete fixture pair while being wrong. That is not
// hypothetical here: a mutation reporting the `with` statement's object expression instead of the
// statement survived the entire suite until this test existed, because every case covering `with`
// asserted only that a finding appeared.
//
// The expected text on each row is a literal typed here rather than anything read off the rule, so
// a mutation moving both the rule and its constant together cannot stay green.
//
// The three spans reproduce upstream exactly. React's own golden for the eval case draws four
// carets under `eval`, and running React's rule on the other two reports columns 3 through 18 for
// `with (props) {}` and 3 through 15 for `class Foo {}`, which are the two spans below.
func TestUnsupportedSyntaxPointsAtTheRightNode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{
			name: "the eval identifier alone, not the call around it",
			sourceText: `function Component(props) {
  eval('x');
  return <div />;
}
`,
			wantText: "eval",
		},
		{
			name: "the whole with statement, not its object",
			sourceText: `function Component(props) {
  with (props) {}
  return <div />;
}
`,
			wantText: "with (props) {}",
		},
		{
			name: "the whole class declaration",
			sourceText: `function Component(props) {
  class Foo {}
  return <div />;
}
`,
			wantText: "class Foo {}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UnsupportedSyntax, unsupportedSyntaxFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Errorf("finding points at %q, want %q", reported, testCase.wantText)
			}
		})
	}
}

// TestUnsupportedSyntaxMessagesAreDistinct asserts the three messages carry the ids and the
// descriptions this rule means, against literals typed here.
//
// A `rule.Message` is `{Id, Description}` with no interpolation layer, so there is no rendered text
// to guard and nothing a format string can get wrong. What is still worth pinning is that the three
// arms report three DIFFERENT messages: a mutation routing all three to one message would leave
// every count assertion green while telling a reader the wrong thing about their code.
func TestUnsupportedSyntaxMessagesAreDistinct(t *testing.T) {
	t.Parallel()

	if messageUnsupportedEval.Id != "unsupportedEval" {
		t.Errorf("eval message id is %q", messageUnsupportedEval.Id)
	}
	if messageUnsupportedWith.Id != "unsupportedWith" {
		t.Errorf("with message id is %q", messageUnsupportedWith.Id)
	}
	if messageUnsupportedInlineClass.Id != "unsupportedInlineClass" {
		t.Errorf("inline class message id is %q", messageUnsupportedInlineClass.Id)
	}

	descriptions := map[string]bool{
		messageUnsupportedEval.Description:        true,
		messageUnsupportedWith.Description:        true,
		messageUnsupportedInlineClass.Description: true,
	}
	if len(descriptions) != 3 {
		t.Error("two of the three messages share a description, so a reader cannot tell which syntax was found")
	}
	if !strings.Contains(messageUnsupportedEval.Description, "eval") {
		t.Error("the eval message does not name eval")
	}
	if !strings.Contains(messageUnsupportedWith.Description, "with") {
		t.Error("the with message does not name with")
	}
	if !strings.Contains(messageUnsupportedInlineClass.Description, "class") {
		t.Error("the inline class message does not name class")
	}
}
