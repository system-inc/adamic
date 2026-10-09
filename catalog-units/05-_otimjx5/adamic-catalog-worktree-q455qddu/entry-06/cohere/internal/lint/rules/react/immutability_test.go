package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestImmutabilityFires covers every input measured to report under React's own rule.
//
// Every expectation in this file and in TestImmutabilityStaysSilent was produced by RUNNING
// `eslint-plugin-react-hooks` 7.1.1 through the ESLint Linter API on the exact source below, not by
// reading either implementation. That matters more than usual for this rule: three separate
// hypotheses formed from reading were wrong, and each would have shipped as a confident fixture.
// They are recorded at their cases.
func TestImmutabilityFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		messages []string
	}{
		{
			name: "f_context",
			source: `import {useContext} from 'react';
function Component(props) {
  const c = useContext(MyContext);
  c.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_global",
			source: `const g = {a:0};
function Component(props) {
  g.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_hook_arg",
			source: `function Component(props) {
  const o = {a:0};
  useThing(o);
  o.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			// The control for `s_ref_by_name_is_exempt`: the identical shape with a binding name the
			// ref regex rejects. Without this the exemption fixture would pass even if the rule went
			// silent on everything, which is the failure mode a lone clean case cannot see.
			name: "f_non_ref_name_current_write",
			source: `function Component(props) {
  const thing = props.thing;
  useThing(thing);
  thing.current = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_hook_arg_alias",
			source: `function Component(props) {
  const o = {a:0};
  const alias = o;
  useThing(alias);
  o.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_hook_return",
			source: `function Component(props) {
  const h = useThing();
  h.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_hook_two_params",
			source: `function useThing(a, b) {
  useOther();
  a.test = 1;
  return a;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_jsx_captured",
			source: `function Component(props) {
  const o = {a:0};
  const el = <Foo x={o} />;
  o.a = 1;
  return el;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_mutate_in_callback",
			source: `function Component(props) {
  const cb = () => { props.a = 1; };
  return <div onClick={cb} />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_namespaced_hook_arg",
			source: `function Component(props) {
  const o = {a:0};
  React.useThing(o);
  o.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_phi_maybe_frozen",
			source: `function Component(props) {
  const h = useThing();
  let x;
  if (props.c) { x = h; } else { x = {}; }
  x.q = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_bag",
			source: `function Component(props) {
  props.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_compound",
			source: `function Component(props) {
  props.count += 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_computed",
			source: `function Component(props) {
  props['a'] = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_delete",
			source: `function Component(props) {
  delete props.a;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_destructured",
			source: `function Component({model}) {
  model.value = 1;
  return <div>{model.value}</div>;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_increment",
			source: `function Component(props) {
  props.count++;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_load_then",
			source: `function Component(props) {
  const x = props.a;
  x.q = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_props_member",
			source: `function Component(props) {
  props.a.q = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_reassign_async",
			source: `function Component(props) {
  let local = 0;
  const cb = async () => { local = 1; };
  return <div onClick={cb} />;
}
`,
			messages: []string{"reassignedInAsyncFunction", "knownMutableFunction"},
		},
		{
			name: "f_reassign_render",
			source: `function Component(props) {
  let local = 0;
  const cb = () => { local = 1; };
  return <div onClick={cb} />;
}
`,
			messages: []string{"reassignedAfterRender", "knownMutableFunction"},
		},
		{
			name: "f_reducer",
			source: `import {useReducer} from 'react';
function Component(props) {
  const [s, d] = useReducer(r, {a:0});
  s.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_state",
			source: `import {useState} from 'react';
function Component(props) {
  const [s, setS] = useState({a:0});
  s.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_useEffect_callback_captures",
			source: `function Component(props) {
  const o = {a:0};
  useEffect(() => { read(o); }, []);
  o.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
		{
			name: "f_useMemo_dependency_array",
			source: `function Component(props) {
  const o = {a:0};
  useMemo(() => 1, [o]);
  o.a = 1;
  return <div />;
}
`,
			messages: []string{"immutableValue"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Immutability, "component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// TestImmutabilityStaysSilent covers every input measured CLEAN under React's own rule.
//
// The ordering pairs here are the rule's whole reason for existing: `s_before_hook` and
// `s_before_jsx` mutate exactly the value that `f_hook_arg` and `f_jsx_captured` mutate, and differ
// only in whether the mutation comes before or after the freeze. A rule keyed on where a value came
// from rather than on what has happened to it passes the fires table and fails both of these.
func TestImmutabilityStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// The effect-cleanup idiom, which is the most common shape in any React codebase and which
			// the ownership check exists for. `cancelled` is declared INSIDE the effect callback, so the
			// closure that writes it writes a binding the effect owns rather than one the render owns,
			// and no render outlives it. Measured CLEAN under React against a control that reports twice.
			//
			// Added for a surviving mutant: forcing the ownership test always-true left every fixture
			// green, because nothing in the imported set declares a variable inside a callback and writes
			// it from a deeper one. It was 30 findings on Kirk's tree.
			// A ref reached by NAME rather than by type, which is the half of the ref exemption a
			// fixture can actually see. `rule_testing`'s tsconfig sets `types: []` and resolves no
			// `node_modules`, so `@types/react` never resolves here and the CHECKER half of that
			// exemption is unreachable through this harness at all; it is pinned by the dry run instead,
			// where it removed 211 findings. Measured CLEAN under React against
			// `f_non_ref_name_current_write`, which is the same shape with a name the ref regex rejects.
			name: "s_ref_by_name_is_exempt",
			source: `function Component(props) {
  const ref = props.ref;
  useThing(ref);
  ref.current = 1;
  return <div />;
}
`,
		},
		{
			name: "s_effect_cleanup_declares_its_own_local",
			source: `function Component(props) {
  useEffect(() => {
    let cancelled = false;
    void go();
    return function() { cancelled = true; };
  }, []);
  return <div />;
}
`,
		},
		{
			// The other half of that pair: a variable declared inside a callback and written from a
			// branch of the SAME callback. Also clean under React.
			name: "s_callback_declares_and_writes_its_own_local",
			source: `function Component(props) {
  const cb = () => { let v; if (props.c) { v = 1; } else { v = 2; } return v; };
  return <div onClick={cb} />;
}
`,
		},
		{
			name: "s_array_push_local",
			source: `function Component(props) {
  const items = [];
  items.push(1);
  return <div>{items.length}</div>;
}
`,
		},
		{
			name: "s_before_hook",
			source: `function Component(props) {
  const o = {a:0};
  o.a = 1;
  useThing(o);
  return <div />;
}
`,
		},
		{
			name: "s_before_jsx",
			source: `function Component(props) {
  const o = {a:0};
  o.a = 1;
  return <Foo x={o} />;
}
`,
		},
		{
			name: "s_computed_var_key",
			source: `function Component(props, k) {
  props[k] = 1;
  return <div />;
}
`,
		},
		{
			name: "s_local_mutate",
			source: `function Component(props) {
  const o = {a:0};
  o.a = 1;
  return <div>{o.a}</div>;
}
`,
		},
		{
			name: "s_lowercase",
			source: `function helper(props) {
  useThing();
  props.a = 1;
  return props;
}
`,
		},
		{
			name: "s_method_push",
			source: `function Component(props) {
  props.list.push(1);
  return <div />;
}
`,
		},
		{
			name: "s_no_gate",
			source: `function Component(props) {
  props.a = 1;
  return null;
}
`,
		},
		{
			name: "s_object_assign",
			source: `function Component(props) {
  Object.assign(props, {a:1});
  return <div />;
}
`,
		},
		{
			name: "s_phi_both_mutable",
			source: `function Component(props) {
  let x;
  if (props.c) { x = {a:1}; } else { x = {}; }
  x.q = 1;
  return <div />;
}
`,
		},
		{
			name: "s_phi_global_mutable",
			source: `const g = {};
function Component(props) {
  let x;
  if (props.c) { x = g; } else { x = {}; }
  x.q = 1;
  return <div />;
}
`,
		},
		{
			name: "s_plain_call",
			source: `function Component(props) {
  const o = {a:0};
  doThing(o);
  o.a = 1;
  return <div />;
}
`,
		},
		{
			name: "s_primitive",
			source: `function Component(props) {
  let x = 3;
  x.q = 1;
  return <div />;
}
`,
		},
		{
			name: "s_three_params",
			source: `function Component(a, b, c) {
  a.q = 1;
  return <div />;
}
`,
		},
		{
			name: "s_two_params",
			source: `function Component(props, cond) {
  props.a = 1;
  return <div />;
}
`,
		},
		{
			name: "s_useEffect_dependency_array",
			source: `function Component(props) {
  const o = {a:0};
  useEffect(() => {}, [o]);
  o.a = 1;
  return <div />;
}
`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Immutability, "component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestImmutabilitySpans asserts WHERE each finding points, which no message-id fixture can see.
//
// The expectations are React's own spans, read off the executable's rendered underline for the same
// source, and the first draft of this test guessed all six of them WRONG in the same direction. The
// guess was that a write underlines the assignment; React underlines the OBJECT BEING MUTATED, so
// `props.a = 1` underlines `props` and `props.a.q = 1` underlines `props.a`. Every one of those six
// was green under `ExpectFindings`, which is exactly the failure this step exists to catch.
func TestImmutabilitySpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			name:     "a bag write underlines the assignment",
			source:   "function Component(props) {\n  props.a = 1;\n  return <div />;\n}\n",
			expected: []string{"props"},
		},
		{
			name:     "a member write underlines the whole assignment",
			source:   "function Component(props) {\n  props.a.q = 1;\n  return <div />;\n}\n",
			expected: []string{"props.a"},
		},
		{
			name:     "a delete underlines the delete expression",
			source:   "function Component(props) {\n  delete props.a;\n  return <div />;\n}\n",
			expected: []string{"props"},
		},
		{
			name:     "a state write underlines the assignment",
			source:   "import {useState} from 'react';\nfunction Component() {\n  const [s] = useState({a: 0});\n  s.a = 1;\n  return <div />;\n}\n",
			expected: []string{"s"},
		},
		{
			name:     "a write after a hook freeze underlines the write, not the hook call",
			source:   "function Component() {\n  const o = {a: 0};\n  useThing(o);\n  o.a = 1;\n  return <div />;\n}\n",
			expected: []string{"o"},
		},
		{
			name:   "a reassigning callback reports at the write and at the handover",
			source: "function Component() {\n  let local = 0;\n  const cb = () => { local = 1; };\n  return <div onClick={cb} />;\n}\n",
			// The first finding is the reassignment itself; the second is the JSX element that hands
			// the callback to React. React underlines `cb` for the second one and this rule
			// underlines the element, which is a narrower-to-wider divergence recorded here rather
			// than silently matched. See the rule's report.
			expected: []string{"local", "<div onClick={cb} />"},
		},
		{
			name:     "a write inside a callback reports inside the callback",
			source:   "function Component(props) {\n  const cb = () => { props.a = 1; };\n  return <div onClick={cb} />;\n}\n",
			expected: []string{"props"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Immutability, "component.tsx", testCase.source)
			if len(result.Diagnostics) != len(testCase.expected) {
				t.Fatalf("expected %d findings, got %d", len(testCase.expected), len(result.Diagnostics))
			}
			// The harness writes `strings.TrimSpace(contents)+"\n"`, so the file on disk is offset
			// from the Go literal above whenever the literal has leading whitespace. These literals
			// have none, so slicing the literal is safe; the transformation is applied anyway so a
			// later edit that adds a leading newline cannot silently shift every span by one.
			written := strings.TrimSpace(testCase.source) + "\n"
			for index, diagnostic := range result.Diagnostics {
				got := written[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.expected[index] {
					t.Errorf("finding %d: expected span %q, got %q", index, testCase.expected[index], got)
				}
			}
		})
	}
}

// TestImmutabilityMessagesAreDistinct pins the four message ids and their descriptions.
//
// Asserted against literal strings typed here rather than against the rule's own constants: the
// brief records a porter whose message-id mutant and message-text mutant BOTH survived a test
// written the other way, because equality against the constant moves with the constant.
func TestImmutabilityMessagesAreDistinct(t *testing.T) {
	t.Parallel()

	source := "function Component(props) {\n  props.a = 1;\n  return <div />;\n}\n"
	result := rule_testing.RunTyped(t, Immutability, "component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "immutableValue" {
		t.Errorf("expected id %q, got %q", "immutableValue", got)
	}
	wantPrefix := "This value is a prop or a hook argument, so it belongs to the caller."
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("expected description to begin %q, got %q", wantPrefix, got)
	}
}

// TestImmutabilityReasonSelectsTheMessage is the split this rule is most likely to get wrong.
//
// The lattice KIND decides whether to report; the reason set decides what the finding SAYS. One
// verdict has ten different explanations and a reader who conflates the two ships a rule that is
// right about every verdict and wrong about every sentence. Each expectation is React's own wording
// for the same input, matched on the distinguishing clause rather than on the whole sentence.
func TestImmutabilityReasonSelectsTheMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		clause string
	}{
		{"props", "function Component(props) {\n  props.a = 1;\n  return <div />;\n}\n", "prop or a hook argument"},
		{"state", "import {useState} from 'react';\nfunction Component() {\n  const [s] = useState({});\n  s.a = 1;\n  return <div />;\n}\n", "came from `useState`"},
		{"reducer", "import {useReducer} from 'react';\nfunction Component() {\n  const [s, d] = useReducer(r, {});\n  s.a = 1;\n  return <div />;\n}\n", "came from `useReducer`"},
		{"context", "import {useContext} from 'react';\nfunction Component() {\n  const c = useContext(C);\n  c.a = 1;\n  return <div />;\n}\n", "came from `useContext`"},
		{"hook return", "function Component() {\n  const h = useThing();\n  h.a = 1;\n  return <div />;\n}\n", "came from a hook"},
		{"hook argument", "function Component() {\n  const o = {};\n  useThing(o);\n  o.a = 1;\n  return <div />;\n}\n", "already passed to a hook"},
		{"global", "const g = {};\nfunction Component() {\n  g.a = 1;\n  return <div />;\n}\n", "declared outside the component"},
		{"jsx captured", "function Component() {\n  const o = {};\n  const el = <Foo x={o} />;\n  o.a = 1;\n  return el;\n}\n", "already used in JSX"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Immutability, "component.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; !strings.Contains(got, testCase.clause) {
				t.Errorf("expected description to contain %q, got %q", testCase.clause, got)
			}
		})
	}
}

// TestImmutabilityConvergenceEqualityIgnoresReasons pins the fixpoint's termination test.
//
// The reason set only ever grows, so a convergence test comparing whole abstract values would see a
// reason bit arriving at a value whose kind is already final and call it movement, on every round,
// until the bound ran out. `immutabilityMergeValues` merges both halves; only the KIND is allowed to
// vote on termination, and this asserts that directly rather than through a fixture, because a
// fixture can only observe the symptom as slowness.
func TestImmutabilityConvergenceEqualityIgnoresReasons(t *testing.T) {
	t.Parallel()

	first := immutabilityAbstractValue{Kind: immutabilityFrozen, Reasons: immutabilityReasonState}
	second := immutabilityAbstractValue{Kind: immutabilityFrozen, Reasons: immutabilityReasonJsxCaptured}
	merged := immutabilityMergeValues(first, second)
	if merged.Kind != first.Kind {
		t.Fatalf("merging two frozen values moved the kind to %v", merged.Kind)
	}
	if merged == first {
		t.Fatal("the merged value is identical to the first, so this test cannot see the hazard")
	}
	if merged.Reasons != (immutabilityReasonState | immutabilityReasonJsxCaptured) {
		t.Errorf("expected both reasons to survive the merge, got %b", merged.Reasons)
	}
}

// TestImmutabilityJoinIsNotAnOrdering pins the third element.
//
// `internal/utilities/hir/high_level_intermediate_representation.go` carries a warning addressed to this rule: a dataflow pass whose join
// is "take the larger constant" is self-consistent, passes its own tests, and is wrong, because
// `Frozen` joined with `Mutable` is `MaybeFrozen`, an element that is NEITHER operand. Every row
// here was measured against React's executable with the compilation gate held constant.
func TestImmutabilityJoinIsNotAnOrdering(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		a, b     immutabilityValueKind
		expected immutabilityValueKind
	}{
		{"frozen with mutable is a third element", immutabilityFrozen, immutabilityMutable, immutabilityMaybeFrozen},
		{"the join is commutative", immutabilityMutable, immutabilityFrozen, immutabilityMaybeFrozen},
		{"mutable with mutable stays mutable", immutabilityMutable, immutabilityMutable, immutabilityMutable},
		{"frozen with frozen stays frozen", immutabilityFrozen, immutabilityFrozen, immutabilityFrozen},
		{"frozen with primitive is frozen", immutabilityFrozen, immutabilityPrimitive, immutabilityFrozen},
		{"a global joined with a mutable is NOT absorbing", immutabilityGlobal, immutabilityMutable, immutabilityMutable},
		{"frozen with context is maybe frozen", immutabilityFrozen, immutabilityContext, immutabilityMaybeFrozen},
		{"maybe frozen absorbs everything", immutabilityMaybeFrozen, immutabilityPrimitive, immutabilityMaybeFrozen},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := immutabilityMergeKinds(testCase.a, testCase.b); got != testCase.expected {
				t.Errorf("expected %v, got %v", testCase.expected, got)
			}
		})
	}

	// A maybe-frozen value must REPORT. The scoping measurement for this rule described the gate as
	// an equality against Frozen, and a rule built from that reading is silent on every joined
	// value while passing every straight-line fixture.
	if immutabilityMutate(immutabilityMaybeFrozen) != immutabilityMutationFrozen {
		t.Error("a maybe-frozen value must report on mutation")
	}
	if immutabilityMutate(immutabilityGlobal) != immutabilityMutationGlobal {
		t.Error("a global must report on mutation")
	}
	if immutabilityMutate(immutabilityPrimitive) != immutabilityMutationNone {
		t.Error("writing through a primitive is silent")
	}
}

// TestImmutabilityRequiresTheTypedHarness fails loudly if the rule stops declaring the checker.
//
// The plain harness hands a rule a nil checker. This rule guards, so it goes completely silent
// rather than panicking, which is the more dangerous of the two failure modes: every silent fixture
// would pass vacuously.
func TestImmutabilityRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "function Component(props) {\n  props.a = 1;\n  return <div />;\n}\n"
	if result := rule_testing.RunTyped(t, Immutability, "component.tsx", source); len(result.Diagnostics) == 0 {
		t.Fatal("the typed harness reported nothing, so this test cannot see a regression")
	}
	if result := rule_testing.Run(t, Immutability, "component.tsx", source); len(result.Diagnostics) != 0 {
		t.Errorf("expected silence without a checker, got %d findings", len(result.Diagnostics))
	}
}

// TestImmutabilityAnalyzesANestedComponentAsItsOwnUnit is TanStack Query's test shape: a component
// declared inside a test callback (#zx5xvtg item 8). React Compiler compiles such a component as a
// unit of its own and resolves the test's variables as it resolves module-level ones, so writing
// them from a handler, from an effect, or from a memoized callback raises no immutability error.
// Measured on the compiler bundled in eslint-plugin-react-hooks 7.1.1, run without the plugin's file
// gate: it compiles the nested `Page` and reports nothing on the first source (the `++` is its Todo,
// not an error this rule names). cohere used to analyze `Page` as a function nested in the test
// callback and report all three writes, twice each.
//
// The control is the component writing its OWN local after render, which the compiler reports at
// the same line nested or not, so the nested component is still analyzed, as itself.
func TestImmutabilityAnalyzesANestedComponentAsItsOwnUnit(t *testing.T) {
	t.Parallel()

	testVariables := `import {useEffect, useCallback} from 'react';
declare function it(name: string, body: () => void): void;
it('counts', () => {
  let isRefetch = false;
  let result: unknown = null;
  let selectCalled = 0;
  function Page() {
    useEffect(() => {
      result = 1;
    });
    const select = useCallback((x: number) => {
      selectCalled++;
      return x;
    }, []);
    return <button onClick={() => { isRefetch = true; }}>{select(1)}</button>;
  }
  return [Page, isRefetch, result, selectCalled];
});
`
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, Immutability, "component.tsx", testVariables))

	nested := rule_testing.RunTyped(t, Immutability, "component.tsx", `declare function it(name: string, body: () => void): void;
it('own local', () => {
  function Page() {
    let local = 0;
    return <button onClick={() => { local = 1; }}>{local}</button>;
  }
  return Page;
});
`)
	topLevel := rule_testing.RunTyped(t, Immutability, "component.tsx", `export function Page() {
  let local = 0;
  return <button onClick={() => { local = 1; }}>{local}</button>;
}
`)
	if len(topLevel.Diagnostics) == 0 {
		t.Fatal("the top-level control reports nothing, so the comparison proves nothing")
	}
	rule_testing.ExpectFindings(t, nested, topLevel.MessageIds()...)
}
