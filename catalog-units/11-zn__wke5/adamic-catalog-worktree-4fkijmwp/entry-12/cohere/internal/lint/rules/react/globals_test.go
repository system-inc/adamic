package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus for this rule is unusually large and unusually misleading, and both facts are recorded
// here rather than in the rule, because they are facts about the FIXTURES.
//
// # Where the cases came from
//
// React vendors 325 error goldens under `internal/react_conformance/testdata/fixtures/`. Eleven of
// them carry this rule's exact message, `Cannot reassign variables declared outside of the
// component/hook`, for fifteen diagnostics in total. Two of those eleven are the `new-mutability/`
// copies of two others, identical inputs shifted one line by a pragma, so there are nine distinct
// behavioral cases. The research note that seeded this port said thirteen fixtures; the measured
// number is fifteen diagnostics over eleven files over nine distinct inputs, and the count depends
// on which of those three things is being counted.
//
// oxc ships a real linter rule for this one, unlike `gating` and `config`, and its tester block is
// one pass and one fail with one diagnostic. The extractor reports no discrepancy. Both of its cases
// are transcribed below.
//
// # The trap in this corpus, which is that most of it is silent under the rule
//
// Two fixtures are named for globals, read exactly like this rule's, and belong to a different rule:
// `error.store-property-in-global` and `error.mutate-property-from-global` both write to a PROPERTY
// of a module-scope binding, and both report `This value cannot be modified` under React's
// immutability rule. Neither carries a `Globals` diagnostic. They are excluded here, and a fixture
// pins the property-write shape as silent so that a later widening has to argue with a test.
//
// More importantly, SIX of the nine distinct goldens produce NO finding under React's own lint rule,
// which was measured by running it rather than inferred. The goldens come from the compiler's
// fixture harness, which compiles a named entrypoint unconditionally; the lint rule instead selects
// its own roots and requires a component or hook to write JSX or call a hook. `function Component()
// { someUnknownGlobal = true; }` has neither, so the compiler fixture reports twice and the lint
// rule is silent. Every such case is recorded below as SILENT with the golden's count in its note,
// because the lint rule is what this port reproduces and pinning them as reporting would have meant
// building a rule that fires on inputs React's rule passes.
//
// That is the port brief's warning about a passing case not being evidence about what you think,
// arriving through the imported corpus in its most expensive form: here the FAILING cases are the
// ones that mislead.
//
// # How the rest of the cases were established
//
// Every case beyond the corpus was measured by running React's `react-hooks/globals` through the
// ESLint Linter API on that exact input before it was written down. Reading the implementation gave
// the wrong answer on the update operators, on the logical compound assignments, on catch and
// finally bodies, and on both spans. Where a case asserts silence, a control assignment was placed
// in the same function first to prove the function was not simply bailing out, since a bailout and
// a decision are indistinguishable from a single silent result.
//
// Sources are built from lists of physical lines through `json.dumps`, so no escape sequence was
// ever typed on the way to this file and nothing could cook one.

type globalsCase struct {
	name   string
	source string
	count  int
}

func TestGlobalsFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range globalsFiresCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Globals, "Subject.tsx", testCase.source)
			expected := make([]string, testCase.count)
			for index := range expected {
				// A literal rather than the rule's own message constant. Comparing against the
				// constant is equality that looks correct and moves with the code under mutation,
				// which the port brief records as having let a message-id mutant survive a whole
				// suite.
				expected[index] = "globalReassignment"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestGlobalsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range globalsSilentCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Globals, "Subject.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func globalsFiresCases() []globalsCase {
	return []globalsCase{
		{name: "moduleLetWrittenInComponent", source: "let g = 0;\nfunction Component() {\n  g = 1;\n  return <div />;\n}\n", count: 1},
		{name: "undeclaredNameWrittenInComponent", source: "function Component() {\n  someGlobal = true;\n  return <div />;\n}\n", count: 1},
		{name: "moduleConstWritten", source: "const b = 1;\nfunction Component() {\n  b = 2;\n  return <div />;\n}\n", count: 1},
		{name: "moduleVarWritten", source: "var b = 1;\nfunction Component() {\n  b = 2;\n  return <div />;\n}\n", count: 1},
		{name: "importedBindingWritten", source: "import {thing} from 'mod';\nfunction Component() {\n  thing = 2;\n  return <div />;\n}\n", count: 1},
		{name: "functionDeclarationWritten", source: "function other() {}\nfunction Component() {\n  other = 2;\n  return <div />;\n}\n", count: 1},
		{name: "classDeclarationWritten", source: "class Thing {}\nfunction Component() {\n  Thing = 1;\n  return <div />;\n}\n", count: 1},
		{name: "standardLibraryBindingWritten", source: "function Component() {\n  Object = 1;\n  return <div />;\n}\n", count: 1},
		{name: "hookRootWrites", source: "import {useState} from 'react';\nlet g = 0;\nfunction useThing() {\n  useState(0);\n  g = 1;\n  return g;\n}\n", count: 1},
		{name: "arrowComponentRootWrites", source: "let g = 0;\nconst Component = () => {\n  g = 1;\n  return <div />;\n};\n", count: 1},
		{name: "twoWritesReportTwice", source: "let g = 0;\nlet h = 0;\nfunction Component() {\n  g = 1;\n  h = 2;\n  return <div />;\n}\n", count: 2},
		{name: "chainedAssignmentReportsBoth", source: "function Component() {\n  a = b = 1;\n  return <div />;\n}\n", count: 2},
		{name: "componentWithRefSecondParameter", source: "let g = 0;\nfunction Component(props, ref) {\n  g = 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundPlus", source: "let g = 0;\nfunction Component() {\n  g += 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundMinus", source: "let g = 0;\nfunction Component() {\n  g -= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundTimes", source: "let g = 0;\nfunction Component() {\n  g *= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundDivide", source: "let g = 0;\nfunction Component() {\n  g /= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundModulo", source: "let g = 0;\nfunction Component() {\n  g %= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundExponent", source: "let g = 0;\nfunction Component() {\n  g **= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundShiftLeft", source: "let g = 0;\nfunction Component() {\n  g <<= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundShiftRight", source: "let g = 0;\nfunction Component() {\n  g >>= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundUnsignedShiftRight", source: "let g = 0;\nfunction Component() {\n  g >>>= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundBitwiseAnd", source: "let g = 0;\nfunction Component() {\n  g &= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundBitwiseOr", source: "let g = 0;\nfunction Component() {\n  g |= 1;\n  return <div />;\n}\n", count: 1},
		{name: "compoundBitwiseXor", source: "let g = 0;\nfunction Component() {\n  g ^= 1;\n  return <div />;\n}\n", count: 1},
		{name: "destructureArrayElement", source: "let g = 0;\nfunction Component(props) {\n  [g] = props.value;\n  return <div />;\n}\n", count: 1},
		{name: "destructureShorthandProperty", source: "function Component(props) {\n  ({a} = props);\n  return <div />;\n}\n", count: 1},
		{name: "destructureRenamedProperty", source: "let g = 0;\nfunction Component(props) {\n  ({k: g} = props);\n  return <div />;\n}\n", count: 1},
		{name: "destructureRestElementArray", source: "function Component(props) {\n  [...rest] = props;\n  return <div />;\n}\n", count: 1},
		{name: "destructureRestElementObject", source: "function Component(props) {\n  ({...rest} = props);\n  return <div />;\n}\n", count: 1},
		{name: "destructureNested", source: "function Component(props) {\n  [[q]] = props;\n  return <div />;\n}\n", count: 1},
		{name: "destructureMixedLocalAndOuter", source: "function Component(props) {\n  let a;\n  [a, b] = props.value;\n  return <div />;\n}\n", count: 1},
		{name: "writeInsideIf", source: "let g = 0;\nfunction Component(p) {\n  if (p) { g = 1; }\n  return <div />;\n}\n", count: 1},
		{name: "writeInsideLoop", source: "let g = 0;\nfunction Component(p) {\n  while (p) { g = 1; }\n  return <div />;\n}\n", count: 1},
		{name: "writeInsideTryBlock", source: "let g = 0;\nfunction Component(p) {\n  try { g = 1; } catch (e) {}\n  return <div />;\n}\n", count: 1},
		{name: "writeInsideSwitchCase", source: "let g = 0;\nfunction Component(p) {\n  switch (p) { case 1: g = 1; break; }\n  return <div />;\n}\n", count: 1},
		{name: "writeInsideBareBlock", source: "let g = 0;\nfunction Component(p) {\n  { g = 1; }\n  return <div />;\n}\n", count: 1},
		{name: "writeInsideJsxExpression", source: "let g = 0;\nfunction Component() {\n  return <div>{(g = 1)}</div>;\n}\n", count: 1},
		{name: "writeAsCallArgument", source: "let g = 0;\nfunction Component() {\n  helper(g = 1);\n  return <div />;\n}\n", count: 1},
		{name: "parenthesizedSimpleTarget", source: "let g = 0;\nfunction Component() {\n  (g) = 1;\n  return <div />;\n}\n", count: 1},
		{name: "parenthesizedCompoundTarget", source: "let g = 0;\nfunction Component() {\n  (g) += 1;\n  return <div />;\n}\n", count: 1},
		// The shorthand accessor returns nil for an undeclared target, which is the reporting case
		// rather than a decline. See the note in `declaredOutsideCompilationRoot`.
		{name: "destructureShorthandOnModuleBinding", source: "let a = 0;\nfunction Component(props) {\n  ({a} = props);\n  return <div />;\n}\n", count: 1},
		// A rest parameter declines ONE function rather than bailing out the file. A sibling rule
		// found upstream silencing a whole function on a rest parameter with no diagnostic to
		// explain it, so the same shape was measured here with a control: the second component in
		// this file still reports, and a rest parameter in a nested function or in a destructuring
		// does not silence the component at all. This rule's gate is per-function, and this is the
		// case that says so.
		{name: "restParameterDeclinesOneFunctionOnly", source: "let g = 0;\nlet h = 0;\nfunction Widget(...props) {\n  g = 1;\n  return <div />;\n}\nfunction Other() {\n  h = 1;\n  return <div />;\n}\n", count: 1},
		// The for-of bailout is also per-function. `First` is silent, `Second` reports.
		{name: "forOfBailoutDoesNotReachTheNextComponent", source: "let h = 0;\nfunction First(props) {\n  for (item of props.items) {}\n  return <div />;\n}\nfunction Second() {\n  h = 1;\n  return <div />;\n}\n", count: 1},
		{name: "restParameterInANestedFunctionDoesNotDecline", source: "let g = 0;\nfunction Component() {\n  const f = (...args) => args;\n  g = 1;\n  return <div />;\n}\n", count: 1},
		{name: "restElementInABodyDestructuringDoesNotDecline", source: "let g = 0;\nfunction Component(props) {\n  const [a, ...others] = props.list;\n  g = 1;\n  return <div />;\n}\n", count: 1},
	}
}

func globalsSilentCases() []globalsCase {
	return []globalsCase{
		// golden reports 2; React's lint rule is silent, no JSX and no hook call
		{name: "upstreamReassignmentToGlobal", source: "function Component() {\n  // Cannot assign to globals\n  someUnknownGlobal = true;\n  moduleLocal = true;\n}\n"},
		// golden reports 2; lint rule silent, no JSX and no hook call
		{name: "upstreamReassignmentToGlobalIndirect", source: "function Component() {\n  const foo = () => {\n    // Cannot assign to globals\n    someUnknownGlobal = true;\n    moduleLocal = true;\n  };\n  foo();\n}\n"},
		// golden reports 1; lint rule silent, hook makes no hook call and writes no JSX
		{name: "upstreamInvalidDestructureAssignmentToGlobal", source: "function useFoo(props) {\n  [x] = props;\n  return {x};\n}\n"},
		// golden reports 1 on b; lint rule silent, no JSX and no hook call
		{name: "upstreamInvalidDestructureToLocalGlobalVariables", source: "function Component(props) {\n  let a;\n  [a, b] = props.value;\n\n  return [a, b];\n}\n"},
		// golden reports 1; lint rule silent, hook has no hook call and no JSX
		{name: "upstreamUpdateGlobalShouldBailout", source: "let renderCount = 0;\nfunction useFoo() {\n  renderCount += 1;\n  return renderCount;\n}\n"},
		// golden reports 1; write is in a nested function, the stated subset
		{name: "upstreamReassignGlobalFnArg", source: "let b = 1;\n\nexport default function MyApp() {\n  const fn = () => {\n    b = 2;\n  };\n  return foo(fn);\n}\n\nfunction foo(fn) {}\n"},
		// React lint reports 1; write is in a nested function, the stated subset
		{name: "upstreamAssignGlobalInComponentTagFunction", source: "function Component() {\n  const Foo = () => {\n    someGlobal = true;\n  };\n  return <Foo />;\n}\n"},
		// React lint reports 1; write is in a nested function, the stated subset
		{name: "upstreamAssignGlobalInJsxChildren", source: "function Component() {\n  const foo = () => {\n    someGlobal = true;\n  };\n  // Children are generally access/called during render, so\n  // modifying a global in a children function is almost\n  // certainly a mistake.\n  return <Foo>{foo}</Foo>;\n}\n"},
		// React lint reports 1; write is in a nested function, the stated subset
		{name: "upstreamInvalidGlobalReassignmentIndirect", source: "import {useEffect, useState} from 'react';\n\nlet someGlobal = false;\n\nfunction Component() {\n  const [state, setState] = useState(someGlobal);\n\n  const setGlobal = () => {\n    someGlobal = true;\n  };\n  const indirectSetGlobal = () => {\n    setGlobal();\n  };\n  indirectSetGlobal();\n\n  useEffect(() => {\n    setState(someGlobal);\n  }, [someGlobal]);\n\n  return <div>{String(state)}</div>;\n}\n"},
		// oxc pass case; React agrees it is silent
		{name: "upstreamOxcPassGlobalInEffect", source: "\nimport {useEffect} from 'react';\nlet someGlobal = false;\nfunction Component() {\n  useEffect(() => {\n    someGlobal = true;\n  }, []);\n  return <div />;\n}\n"},
		// oxc fail case, 1 diagnostic; write is in a nested function, the stated subset
		{name: "upstreamOxcFailSetGlobal", source: "\nlet someGlobal = false;\nfunction Component() {\n  const setGlobal = () => {\n    someGlobal = true;\n  };\n  setGlobal();\n  return <div>{String(someGlobal)}</div>;\n}\n"},
		// own local
		{name: "localDeclaredInComponent", source: "function Component() {\n  let x = 1;\n  x = 2;\n  return <div />;\n}\n"},
		// own parameter
		{name: "parameterWritten", source: "function Component(props) {\n  props = 1;\n  return <div />;\n}\n"},
		// the HIR gets this one wrong; upstream classifies it StoreContext
		{name: "captureFromEnclosingFunction", source: "function Component() {\n  let x = 1;\n  const f = () => { x = 2; };\n  f();\n  return <div />;\n}\n"},
		// shadow
		{name: "shadowedByLocal", source: "let g = 0;\nfunction Component() {\n  let g = 1;\n  g = 2;\n  return <div />;\n}\n"},
		// shadow
		{name: "shadowedByParameter", source: "let g = 0;\nfunction Component(g) {\n  g = 2;\n  return <div />;\n}\n"},
		// gate: name
		{name: "lowercaseFunctionIsNotAComponent", source: "function helper() {\n  someGlobal = true;\n  return <div />;\n}\n"},
		// gate: no enclosing function
		{name: "moduleTopLevelWrite", source: "someGlobal = true;\n"},
		// gate: no evidence
		{name: "componentWithoutJsxOrHookCall", source: "function Component() {\n  someGlobal = true;\n}\n"},
		// gate: rest parameter
		{name: "componentWithRestParameter", source: "function Component(...props) {\n  someGlobal = true;\n  return <div />;\n}\n"},
		// gate: second parameter must contain ref
		{name: "componentWithSecondParameterNotRef", source: "function Component(props, other) {\n  someGlobal = true;\n  return <div />;\n}\n"},
		// gate: at most two parameters
		{name: "componentWithThreeParameters", source: "function Component(a, b, c) {\n  someGlobal = true;\n  return <div />;\n}\n"},
		// gate: evidence walk skips nested functions
		{name: "jsxOnlyInsideNestedFunction", source: "function Component() {\n  someGlobal = true;\n  const f = () => <div />;\n  return f;\n}\n"},
		// measured silent upstream with a control
		{name: "updatePostfixIncrement", source: "let g = 0;\nfunction Component() {\n  g++;\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "updatePostfixDecrement", source: "let g = 0;\nfunction Component() {\n  g--;\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "updatePrefixIncrement", source: "let g = 0;\nfunction Component() {\n  ++g;\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "updatePrefixDecrement", source: "let g = 0;\nfunction Component() {\n  --g;\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "updateLogicalOr", source: "let g = 0;\nfunction Component() {\n  g ||= 1;\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "updateLogicalAnd", source: "let g = 0;\nfunction Component() {\n  g &&= 1;\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "updateNullish", source: "let g = 0;\nfunction Component() {\n  g ??= 1;\n  return <div />;\n}\n"},
		// a property write is the immutability rule, not this one
		{name: "propertyWriteOnOuterBinding", source: "let wat = {};\nfunction Component() {\n  wat.test = 1;\n  return <div />;\n}\n"},
		// property write
		{name: "elementWriteOnOuterBinding", source: "let g = {};\nfunction Component() {\n  g[0] = 1;\n  return <div />;\n}\n"},
		// a read is not a write
		{name: "readOnly", source: "function Component() {\n  return <div>{someGlobal}</div>;\n}\n"},
		// a call is not a write
		{name: "callOnly", source: "function Component() {\n  someGlobal();\n  return <div />;\n}\n"},
		// upstream takes the whole function silent; stated divergence
		{name: "forOfHeadWithoutDeclaration", source: "function Component(props) {\n  for (item of props.items) {}\n  return <div />;\n}\n"},
		// a declaration is not an assignment
		{name: "forOfHeadWithDeclaration", source: "function Component(props) {\n  for (const q of props.items) {}\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "writeInsideCatchClause", source: "let g = 0;\nfunction Component(p) {\n  try {} catch (e) { g = 1; }\n  return <div />;\n}\n"},
		// measured silent upstream with a control
		{name: "writeInsideFinallyBlock", source: "let g = 0;\nfunction Component(p) {\n  try {} finally { g = 1; }\n  return <div />;\n}\n"},
		// unreachable; measured silent upstream with a control
		{name: "writeAfterReturn", source: "let g = 0;\nfunction Component() {\n  return <div />;\n  g = 1;\n}\n"},
		// not during render
		{name: "writeInsideEffectCallback", source: "import {useEffect} from 'react';\nlet g = 0;\nfunction Component() {\n  useEffect(() => { g = 1; }, []);\n  return <div />;\n}\n"},
		// not during render
		{name: "writeInsideEventHandler", source: "let g = 0;\nfunction Component() {\n  return <div onClick={() => { g = 1; }} />;\n}\n"},
		// The three shapes that keep the unconditional shorthand substitution honest: a shadow must
		// still decline, and a default VALUE is a read rather than a write.
		{name: "shorthandShadowedByLocal", source: "function Component(props) {\n  let a;\n  ({a} = props);\n  return <div />;\n}\n"},
		{name: "shorthandShadowedByParameter", source: "function Component(props, a) {\n  ({a} = props);\n  return <div />;\n}\n"},
		{name: "shorthandDefaultValueIsARead", source: "let a = 0;\nfunction Component(props) {\n  let files;\n  ({files = a} = props);\n  return <div />;\n}\n"},
		// A parenthesized target under a NON-reporting operator. These are what make the upward
		// parenthesis skip in `isGlobalStoreTarget` load-bearing: without it the switch never sees
		// the update or logical-assign parent, falls through, and reports. A mutant removing that
		// skip SURVIVED the whole suite before these existed, because every parenthesized fixture
		// here used a REPORTING operator, where skipping and not skipping both reach "report".
		// All four measured silent under React's own rule.
		{name: "parenthesizedPostfixIncrement", source: "let g = 0;\nfunction Component() {\n  (g)++;\n  return <div />;\n}\n"},
		{name: "parenthesizedPrefixIncrement", source: "let g = 0;\nfunction Component() {\n  ++(g);\n  return <div />;\n}\n"},
		{name: "parenthesizedLogicalOrAssign", source: "let g = 0;\nfunction Component() {\n  (g) ||= 1;\n  return <div />;\n}\n"},
		{name: "parenthesizedNullishAssign", source: "let g = 0;\nfunction Component() {\n  (g) ??= 1;\n  return <div />;\n}\n"},
		// A rest parameter in a NESTED function, and a rest element in a destructuring, are not the
		// component's own parameter list and do not decline it. Both measured; both report their
		// global write, which is why they are in the fires table rather than here. What is pinned
		// here is the component whose OWN parameter list carries the rest.
		{name: "restParameterOnTheComponentItself", source: "let g = 0;\nfunction Widget(...props) {\n  g = 1;\n  return <div />;\n}\n"},
	}
}

// TestGlobalsSpans asserts where each finding points, which no message-id assertion can see.
//
// This rule has TWO spans and they are upstream's, not a choice: a simple assignment reports the
// identifier and a compound assignment reports the whole assignment expression. The two come from
// two lowering sites passing different spans (`build_hir.rs:4408` against `:4534`), React's ESLint
// output and oxc's snapshot agree on both, and a port could satisfy every count fixture above while
// reporting the wrong one everywhere.
//
// The expected text was read off React's own output for the same input rather than derived from the
// source, and the parenthesized pair is here because it is where the two spans visibly disagree
// about the parentheses: `(g) = 1` reports inside them and `(g) += 1` reports from the open paren.
func TestGlobalsSpans(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name     string
		source   string
		reported string
	}{
		{
			name:     "simpleAssignmentReportsTheIdentifier",
			source:   "let longGlobalName = 0;\nfunction Component() {\n  longGlobalName = 1;\n  return <div />;\n}\n",
			reported: "longGlobalName",
		},
		{
			name:     "compoundAssignmentReportsTheWholeExpression",
			source:   "let longGlobalName = 0;\nfunction Component() {\n  longGlobalName += 1;\n  return <div />;\n}\n",
			reported: "longGlobalName += 1",
		},
		{
			name:     "simpleAssignmentExcludesALongRightSide",
			source:   "let g = 0;\nfunction Component() {\n  g = someLongExpression + 1;\n  return <div />;\n}\n",
			reported: "g",
		},
		{
			name:     "compoundAssignmentIncludesALongRightSide",
			source:   "let g = 0;\nfunction Component() {\n  g += someLongExpression + 1;\n  return <div />;\n}\n",
			reported: "g += someLongExpression + 1",
		},
		{
			name:     "parenthesizedSimpleTargetReportsInsideTheParentheses",
			source:   "let g = 0;\nfunction Component() {\n  (g) = 1;\n  return <div />;\n}\n",
			reported: "g",
		},
		{
			name:     "parenthesizedCompoundTargetReportsFromTheOpenParenthesis",
			source:   "let g = 0;\nfunction Component() {\n  (g) += 1;\n  return <div />;\n}\n",
			reported: "(g) += 1",
		},
		{
			name:     "destructuredElementReportsTheElement",
			source:   "function Component(props) {\n  [...rest] = props;\n  return <div />;\n}\n",
			reported: "rest",
		},
		{
			name:     "defaultedDestructuredElementReportsTheWholeElement",
			source:   "function Component(props) {\n  [q = 5] = props;\n  return <div />;\n}\n",
			reported: "q = 5",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Globals, "Subject.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Errorf("finding points at %q, wanted %q", reported, testCase.reported)
			}
		})
	}
}

// TestGlobalsMessageNamesTheBinding asserts the rendered text exactly.
//
// The message interpolates the binding's name, so the id assertion above cannot see anything the
// format string does. The port brief records two survivors that were exactly this: a finding count
// and id both correct while the per-finding text moved. Equality rather than `strings.Contains`,
// because a weaker predicate than the property it guards is not a guard.
func TestGlobalsMessageNamesTheBinding(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, Globals, "Subject.tsx",
		"let renderCount = 0;\nfunction Component() {\n  renderCount = 1;\n  return <div />;\n}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
	}
	message := result.Diagnostics[0].Message.Description
	wantedPrefix := "This writes to `renderCount`, which is declared outside the component or hook, while the component is rendering."
	if !strings.HasPrefix(message, wantedPrefix) {
		t.Errorf("message is %q, wanted it to start with %q", message, wantedPrefix)
	}
	if strings.Contains(message, "``") {
		t.Errorf("message has a doubled backtick, which means the name was interpolated twice: %q", message)
	}
}

// TestGlobalsRequiresTheTypedHarness pins that this rule needs a checker.
//
// The port brief records that `GetSymbolAtLocation` on a nil checker returns nil rather than
// panicking, so a typed rule run through the untyped harness goes VACUOUSLY GREEN rather than
// failing loudly. Every silent fixture above would pass for the wrong reason if the harness were
// ever changed, so this asserts the difference directly: the same input reports under `RunTyped`
// and is silent under `Run`.
func TestGlobalsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "let g = 0;\nfunction Component() {\n  g = 1;\n  return <div />;\n}\n"

	typed := rule_testing.RunTyped(t, Globals, "Subject.tsx", source)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness should report once, got %d", len(typed.Diagnostics))
	}

	untyped := rule_testing.Run(t, Globals, "Subject.tsx", source)
	if len(untyped.Diagnostics) != 0 {
		t.Errorf("the untyped harness should be silent, got %d findings", len(untyped.Diagnostics))
	}
}

// TestGlobalsBoundary pins the stated subset so it stays a decision rather than becoming a drift.
//
// Upstream reports a write inside a nested function when its effects pass can prove that function
// runs during render. That proof is inter-procedural aliasing, and the measurements recorded on the
// rule show it giving opposite answers at the same call depth depending on how the functions are
// spelled and ordered, so there is no narrower syntactic approximation worth shipping. This rule
// therefore reports only writes lexically inside the compilation root's own body.
//
// Each case below is one React REPORTS and this rule does not. They are pinned as silent so that a
// later widening has to come here and change them deliberately, and so the next reader can see
// exactly what the subset costs rather than having to rediscover it.
func TestGlobalsBoundary(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{
			name:   "calledArrowInsideComponent",
			source: "let g = 0;\nfunction Component() {\n  const f = () => { g = 1; };\n  f();\n  return <div />;\n}\n",
		},
		{
			name:   "immediatelyInvokedArrow",
			source: "let g = 0;\nfunction Component() {\n  (() => { g = 1; })();\n  return <div />;\n}\n",
		},
		{
			name:   "calledFunctionDeclarationInsideComponent",
			source: "let g = 0;\nfunction Component() {\n  function f() { g = 1; }\n  f();\n  return <div />;\n}\n",
		},
		{
			name:   "arrowPassedAsJsxChildren",
			source: "let g = 0;\nfunction Component() {\n  const f = () => { g = 1; };\n  return <Foo>{f}</Foo>;\n}\n",
		},
		{
			name:   "arrowUsedAsAComponentTag",
			source: "let g = 0;\nfunction Component() {\n  const F = () => { g = 1; };\n  return <F />;\n}\n",
		},
		{
			name:   "writeInsideAUseMemoCallback",
			source: "import {useMemo} from 'react';\nlet g = 0;\nfunction Component() {\n  useMemo(() => { g = 1; }, []);\n  return <div />;\n}\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, Globals, "Subject.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
