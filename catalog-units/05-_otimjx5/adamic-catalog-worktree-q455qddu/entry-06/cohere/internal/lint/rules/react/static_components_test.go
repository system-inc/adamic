package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The six invalid cases are React's own, vendored at
// `oxc_react_compiler/fixtures/static-components/`, copied byte for byte and verified against the
// originals by hash rather than by reading. Each keeps its upstream pragma comment so a reader can
// find the source file. oxc's linter tester carries only one of these plus one clean case, which
// badly understates the corpus; the fixture directory is the real floor.
//
// The clean cases below the corpus are measured rather than invented. Upstream ships NO valid
// fixtures for this rule at all, so every pass case here was established by running the release
// `oxlint` binary on the input and recording what it decided. That is the port brief's rule about
// taking truth only from the corpus, applied where the corpus has a hole: a pass case written from
// belief would encode the same belief as the rule and pass for exactly the wrong reason.

// Upstream: invalid-dynamically-construct-component-in-render.js
const scDirectCall = `// @loggerTestOnly @validateStaticComponents @outputMode:"lint"
function Example(props) {
  const Component = createComponent();
  return <Component />;
}
`

// Upstream: invalid-dynamically-constructed-component-new.js
const scNewExpression = `// @loggerTestOnly @validateStaticComponents @outputMode:"lint"
function Example(props) {
  const Component = new ComponentFactory();
  return <Component />;
}
`

// Upstream: invalid-dynamically-constructed-component-method-call.js
const scMethodCall = `// @loggerTestOnly @validateStaticComponents @outputMode:"lint"
function Example(props) {
  const Component = props.foo.bar();
  return <Component />;
}
`

// Upstream: invalid-dynamically-constructed-component-function.js
const scFunctionDeclaration = `// @loggerTestOnly @validateStaticComponents @outputMode:"lint"
function Example(props) {
  function Component() {
    return <div />;
  }
  return <Component />;
}
`

// Upstream: invalid-dynamically-constructed-component-arrow-destructured-param.js
const scArrowDestructured = `// @loggerTestOnly @validateStaticComponents @outputMode:"lint"
function Example() {
  const Component = ({x}) => <div>{x}</div>;
  return <Component x="value" />;
}
`

// Upstream: invalid-conditionally-assigned-dynamically-constructed-component-in-render.js
//
// The case that decides whether this rule can exist at all: `Component` is dynamic on one branch
// only, so reporting it requires merging taint where the branches join. Nothing syntactic reaches
// it.
const scConditionalAssignment = `// @loggerTestOnly @validateStaticComponents @outputMode:"lint"
function Example(props) {
  let Component;
  if (props.cond) {
    Component = createComponent();
  } else {
    Component = DefaultComponent;
  }
  return <Component />;
}
`

// The single fail case oxc's linter tester carries, kept separate because it is the input the
// published snapshot's spans were measured against.
const scTesterFail = `
function Example(props) {
  const Component = createComponent();
  return <Component />;
}
`

// The single pass case oxc's linter tester carries.
const scTesterPass = `
function Inner(props) {
  return <div>{props.text}</div>;
}
function Outer() {
  return <Inner text='hello' />;
}
`

func TestStaticComponentsFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"directCall", scDirectCall},
		{"newExpression", scNewExpression},
		{"methodCall", scMethodCall},
		{"functionDeclaration", scFunctionDeclaration},
		{"arrowDestructured", scArrowDestructured},
		{"conditionalAssignment", scConditionalAssignment},
		{"testerFail", scTesterFail},

		// The tag family beyond the plain case. An identifier tag reports wherever it sits, so
		// nesting inside an element or a fragment changes nothing; a member tag never reports
		// however dynamic its base. Measured across six shapes against the release binary, because
		// one shape is not the family and `<c.x/>` alone could have been declining by accident.
		{"tagInsideFragment", `function D() {
  const T = mk();
  return <><T /></>;
}
`},
		{"tagNestedInElement", `function E() {
  const T = mk();
  return <div><T /></div>;
}
`},

		// A component created inside a NESTED function still reports, which is what makes the
		// recursion into `function.Functions` load-bearing rather than tidy. Measured: upstream
		// reports at the inner `<Component />` on the arrow's own line, so each function is
		// analysed as its own graph rather than the outer one being the only subject. Added after
		// a mutant that deleted the recursion survived every other fixture here.
		{"componentCreatedInsideNestedFunction", `function Standalone() {
  const Inner = () => {
    const Component = createComponent();
    return <Component />;
  };
  return Inner;
}
`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, StaticComponents, "component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "staticComponents")
		})
	}
}

// A component declared inside a component is NOT a second compilation unit.
//
// Upstream reports once here, at `<Inner />`, because `Inner` is a value created during `Outer`'s
// render. It never reports the `<Component />` inside `Inner`, because `Inner` is lowered into
// `Outer`'s graph rather than compiled on its own.
//
// This fixture asserted two findings when it was written, from reading rather than measuring, and
// the rule obligingly produced two. Both were wrong. Running the release oxlint binary on this
// exact input gave one finding at line 6, which is what it asserts now. It is the port brief's
// warning about taking truth from anywhere but the corpus, caught by the dry run rather than here.
func TestStaticComponentsTreatsTheOutermostComponentAsTheUnit(t *testing.T) {
	t.Parallel()

	source := `function Outer() {
  function Inner() {
    const Component = createComponent();
    return <Component />;
  }
  return <Inner />;
}
`
	parsed := strings.TrimSpace(source) + "\n"
	result := rule_testing.RunTyped(t, StaticComponents, "component.tsx", source)
	rule_testing.ExpectFindings(t, result, "staticComponents")
	reported := parsed[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "Inner" {
		t.Errorf("span points at %q, want %q", reported, "Inner")
	}
}

// The span points at the tag identifier and not at the trivia before it.
//
// This exists because a mutant that reported the representation's raw `Place.Range` instead of the
// tag's own node survived every other fixture in this file. The two spans are byte-identical for
// every case in the corpus, so nothing there could distinguish them; they diverge only when
// something sits between the `<` and the name, where the raw range is `" /*c*/ C"` and the node
// range is `"C"`. Upstream points at the identifier too, measured at column 18 of this input.
//
// The general shape is the port brief's warning that a fixture whose predicate is weaker than the
// property it guards is not a guard: every message-id assertion in this file stayed green while the
// reporting layer was wrong.
func TestStaticComponentsSpanSkipsTriviaBeforeTheTag(t *testing.T) {
	t.Parallel()

	source := "function O() {\n  const C = mk();\n  return < /*c*/ C />;\n}\n"
	parsed := strings.TrimSpace(source) + "\n"
	result := rule_testing.RunTyped(t, StaticComponents, "component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	reported := parsed[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "C" {
		t.Errorf("span points at %q, want %q", reported, "C")
	}
}

func TestStaticComponentsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"testerPass", scTesterPass},

		// The gate, and the seven false positives it prevents. A render callback defined inside a
		// component is not itself a compilation unit, so upstream never analyses it on its own and
		// reports nothing here. Every finding this rule produced on the real tree before the gate
		// existed was this exact shape, and reading them they look like true positives, which is
		// what makes the divergence worth a fixture rather than a comment.
		{"renderCallbackInsideComponent", `function Widget() {
  const render = function (item) {
    const Icon = iconFor(item);
    return <Icon />;
  };
  return <div>{render(1)}</div>;
}
`},

		// Capitalized but creates no JSX and calls no hook, so upstream does not compile it. The
		// name alone is not the test, which is the half of `getReactFunctionType` that is easy to
		// miss when porting from the name check alone.
		{"capitalizedButNotAComponent", `function Widget() {
  const C = mk();
  return C;
}
`},

		// A lowercase function is never a unit however plainly it renders.
		{"lowercaseFunction", `function widget() {
  const C = mk();
  return <C />;
}
`},

		// Measured against the release oxlint binary, all silent there. Upstream ships no valid
		// fixtures, so these are the pass side of the corpus rather than additions to it.
		{"moduleScopeAlias", `const Hoisted = () => <div />;
function Outer() {
  const C = Hoisted;
  return <C />;
}
`},

		// A member-expression tag is not a JsxTag Place, so no taint can attach to it however
		// dynamic its base is. Measured across the family: `<c.x/>`, `<obj.Inner.Deep/>` and
		// `<T.x.y/>` are all silent upstream, and a lowercase base changes nothing, because this
		// rule never consults a component-name predicate.
		{"memberTag", `function Outer() {
  const c = makeIt();
  return <c.x />;
}
`},
		{"memberTagDeep", `function Outer() {
  const obj = makeIt();
  return <obj.Inner.Deep />;
}
`},

		// The back edge. `C` is dynamic only on the loop's second and later iterations, and a
		// single forward pass over blocks in reverse postorder cannot see taint that arrives on an
		// edge it has already passed. Upstream has the identical limitation, confirmed by running
		// it: taint written BEFORE the loop reports, taint written INSIDE reaching a use after the
		// loop does not. This is fidelity, not a gap, and the fixture exists so a later reader who
		// notices the silence does not helpfully repair it into a divergence.
		{"loopBackEdge", `function Outer(items) {
  let C = Known;
  for (const i of items) { C = makeIt(); }
  return <C />;
}
`},

		// A tag inside a nested function, where the enclosing function is not a compilation unit.
		// Upstream is silent here too, measured on this exact input, so this is agreement rather
		// than the under-reporting the comment here used to claim.
		//
		// The claim was that `FunctionExpression.Captures` was empty so taint could not cross a
		// function boundary. That was true when this fixture was written and stopped being true one
		// commit later, when lowering learned to resolve captures. Re-measured rather than left:
		// captures now populate, this input is still silent, and upstream is still silent, because
		// what decides it is the compilation gate and not the capture gap.
		{"tagInsideNonUnitNestedFunction", `function Outer() {
  const C = makeIt();
  const render = () => <C />;
  return render();
}
`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, StaticComponents, "component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The rule points at the JSX tag rather than at the creation site, matching upstream's primary
// span. Upstream additionally carries a secondary label at the creation site; our Diagnostic holds
// one Range, so that detail moves into the message instead. Asserted on the sliced source rather
// than on a line number, so a change to where the finding points fails here.
//
// # Slice what was PARSED, not what was passed
//
// `rule_testing.RunTyped` writes each fixture to disk as `strings.TrimSpace(contents)+"\n"`, so a case
// carrying leading or trailing whitespace is parsed as a DIFFERENT string from the one handed to
// it. oxc's tester writes every case with a leading newline and this file copies them verbatim, so
// every diagnostic offset here is shifted one byte from the Go literal. Slicing the literal reports
// `<Componen` for a finding that is correctly on `Component`, which reads exactly like an
// off-by-one in the rule and is not one. `rule_testing.Run`, the untyped harness, does not trim, so the
// trap is invisible to any rule that does not need the checker.
func TestStaticComponentsSpan(t *testing.T) {
	t.Parallel()

	parsed := strings.TrimSpace(scTesterFail) + "\n"
	result := rule_testing.RunTyped(t, StaticComponents, "component.tsx", scTesterFail)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	reported := parsed[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "Component" {
		t.Errorf("span points at %q, want %q", reported, "Component")
	}
}

// The message names the construct that created the component, which is the information upstream
// puts in its secondary span. Asserted by equality on the rendered text rather than with
// strings.Contains, because a Contains predicate is weaker than the property it guards and would
// stay green through a wrong interpolation.
func TestStaticComponentsMessageNamesCreationSite(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, StaticComponents, "component.tsx", scTesterFail)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	got := result.Diagnostics[0].Message.Description
	if !strings.HasPrefix(got, "This component is created during render, at `createComponent`,") {
		t.Errorf("message does not name the creation site, got %q", got)
	}
}

// The rule declares NeedsTypeChecker, and it is load-bearing rather than decorative: lowering
// without a checker resolves every reference as a global, so no value is ever named, no taint can
// attach, and the rule goes silent on every input while looking healthy. This asserts the typed
// harness is required, so a later revert to `rule_testing.Run` fails loudly here rather than turning
// every fixture above into a vacuous pass.
func TestStaticComponentsRequiresTypeChecker(t *testing.T) {
	t.Parallel()

	if !StaticComponents.NeedsTypeChecker {
		t.Fatal("rule must declare NeedsTypeChecker; lowering without one names no values and the rule silently reports nothing")
	}
	untyped := rule_testing.Run(t, StaticComponents, "component.tsx", scTesterFail)
	if len(untyped.Diagnostics) != 0 {
		t.Fatalf("expected the untyped harness to produce nothing, got %d", len(untyped.Diagnostics))
	}
}
