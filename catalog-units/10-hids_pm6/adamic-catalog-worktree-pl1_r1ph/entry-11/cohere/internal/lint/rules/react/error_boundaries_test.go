package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The upstream corpus is four fixtures and only two of them reach this rule.
//
// React ships `invalid-jsx-in-try-with-catch` and `invalid-jsx-in-catch-in-outer-try-with-catch`,
// which report, and `error.todo-invalid-jsx-in-try-with-finally` and
// `error.todo-invalid-jsx-in-catch-in-outer-try-with-finally`, which do not: those two are named
// `invalid-...` and produce a lowering Todo instead of this rule's finding, which is what the
// `error.todo` prefix records. A porter counting `error.`-named files finds the two that are silent
// and misses both that report, because the reporting pair runs under `@outputMode:"lint"` and lands
// its diagnostic in a `## Logs` block rather than an `## Error` one.
//
// All four are transcribed here verbatim, with the compiler pragma comment removed, and the two
// silent ones are pinned as silent with the reasoning at `ErrorBoundaries`.
//
// Every case beyond those four was measured by running React's own `react-hooks/error-boundaries`
// through the ESLint Linter API on that exact input before it was written down, rather than derived
// from reading the implementation. Reading gave the wrong answer on the finally clauses, on nested
// functions, and on the whole component gate.

type errorBoundariesCase struct {
	name   string
	source string
	count  int
}

func TestErrorBoundariesFires(t *testing.T) {
	t.Parallel()

	cases := []errorBoundariesCase{
		{
			name:   "upstreamJsxInTryWithCatch",
			source: "function Component(props) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "upstreamJsxInCatchInOuterTryWithCatch",
			source: "import {identity} from 'shared-runtime';\n\nfunction Component(props) {\n  let el;\n  try {\n    let value;\n    try {\n      value = identity(props.foo);\n    } catch {\n      el = <div value={value} />;\n    }\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "jsxReturnedDirectlyFromTry",
			source: "function Component() {\n  try {\n    return <div />;\n  } catch {\n    return null;\n  }\n}\n",
			count:  1,
		},
		{
			name:   "fragmentInTry",
			source: "function Component() {\n  let el;\n  try {\n    el = <>hi</>;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "twoJsxInTryReportTwice",
			source: "function Component() {\n  let a, b;\n  try {\n    a = <div />;\n    b = <span />;\n  } catch {\n    return null;\n  }\n  return [a, b];\n}\n",
			count:  2,
		},
		{
			name:   "nestedJsxChildrenReportSeparately",
			source: "function Component() {\n  let el;\n  try {\n    el = <div><span /></div>;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  2,
		},
		{
			name:   "finallyAlongsideCatchStillReports",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  } finally {\n    log();\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "jsxInLoopInTry",
			source: "function Component() {\n  let el;\n  try {\n    for (const x of xs) {\n      el = <div />;\n    }\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "jsxInIfInTry",
			source: "function Component() {\n  let el;\n  try {\n    if (x) {\n      el = <div />;\n    }\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "jsxInSwitchInTry",
			source: "function Component() {\n  let el;\n  try {\n    switch (x) {\n      case 1:\n        el = <div />;\n    }\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "jsxInInnerTryInsideCatch",
			source: "function Component() {\n  let el;\n  try {\n    foo();\n  } catch {\n    try {\n      el = <div />;\n    } catch {\n      el = null;\n    }\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "tripleNestedInnermostCatch",
			source: "function Component() {\n  let el;\n  try {\n    try {\n      try {\n        foo();\n      } catch {\n        el = <div />;\n      }\n    } catch {\n      bar();\n    }\n  } catch {\n    baz();\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "conditionalExpressionInTryReportsBothArms",
			source: "function Component() {\n  let el;\n  try {\n    el = x ? <div /> : <span />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  2,
		},
		{
			name:   "logicalExpressionInTry",
			source: "function Component() {\n  let el;\n  try {\n    el = x && <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "hookNamedFunctionReports",
			source: "function useThing() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "arrowAssignedToComponentName",
			source: "const Component = (props) => {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n};\n",
			count:  1,
		},
		{
			name:   "functionExpressionAssignedToComponentName",
			source: "const Component = function (props) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n};\n",
			count:  1,
		},
		{
			name:   "exportedDefaultNamedComponent",
			source: "export default function Component(props) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			// "Ref" strictly in the middle, which is what separates a substring test from a prefix
			// one AND from a suffix one. Every other ref fixture here ends in the needle, so a
			// mutant narrowing `Contains` to `HasSuffix` survived them all.
			name:   "secondParameterWithRefInTheMiddle",
			source: "function Component(props, theRefValue) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
			count:  1,
		},
		{
			// Upstream's looseness, pinned rather than tidied: a parameter named `prefix` contains
			// `ref`, so it counts as a ref and the function is a component. Measured REPORTING on
			// React. It reads like a bug and it is upstream's, so it is reproduced.
			name:   "secondParameterNamedPrefixCountsAsARef",
			source: "function Component(props, prefix) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
			count:  1,
		},
		{
			// Upstream's ref test is a SUBSTRING test, `name.includes('ref') || name.includes('Ref')`,
			// not equality, and these two are what pin that. Both measured REPORTING on React.
			name:   "secondParameterNamedForwardedRef",
			source: "function Component(props, forwardedRef) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "secondParameterNamedMyref",
			source: "function Component(props, myref) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "secondParameterNamedRef",
			source: "function Component(props, ref) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "destructuredPropsParameter",
			source: "function Component({a, b}) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "noParameters",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
			count:  1,
		},
		{
			name:   "jsxUsedAsCallArgumentInTry",
			source: "function Component() {\n  try {\n    foo(<div />);\n  } catch {\n    return null;\n  }\n  return <span />;\n}\n",
			count:  1,
		},
		{
			// Upstream is CLEAN here and this rule reports, deliberately. `el` is assigned and
			// never read, so React's lowering drops the JSX instruction as a value nothing
			// consumes before the validator runs. That elimination is not reproduced; see the
			// divergence recorded at `ErrorBoundaries`. Measured on the real rule rather than
			// assumed: adding `log(el)` to this same body makes upstream report.
			name:   "assignedJsxNeverReadIsADocumentedDivergence",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  return 1;\n}\n",
			count:  1,
		},
		{
			// A non-node return visited before a node one does not decide the answer, because
			// upstream assigns rather than accumulates. Measured reporting.
			name:   "nonNodeReturnBeforeNodeReturn",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  if (x) return {};\n  return el;\n}\n",
			count:  1,
		},
		{
			// A nested function's returns say nothing about the enclosing one, and the nested
			// function has to be written INSIDE the component's own return for a fixture to see
			// that. A nested function written earlier cannot: an unguarded walk visits its return
			// first and the component's own return last, and since upstream assigns rather than
			// accumulates, the last one decides in both versions. Two fixtures written that way
			// left the mutant alive, and the walk was instrumented rather than guessed at a third
			// time, which is what the brief asks for after a second failed hypothesis.
			//
			// Here the nested function sits in the return's own argument list, so an unguarded walk
			// visits `return {}` LAST and answers non-component. Measured REPORTING on React.
			name:   "nestedFunctionReturningANonNodeInsideTheComponentsOwnReturn",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  return log(el, function () {\n    return {};\n  });\n}\n",
			count:  1,
		},
		{
			name:   "nestedArrowReturningANonNodeInsideTheComponentsOwnReturn",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  return wrap(el, () => {\n    return {};\n  });\n}\n",
			count:  1,
		},
		{
			// No return statement at all leaves upstream's flag false, so the function is a
			// component. Measured reporting, and it is the case that separates "assigns" from
			// "any return is a non-node".
			name:   "noReturnStatementAtAll",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n}\n",
			count:  1,
		},
		{
			name:   "innerComponentElementInOuterTry",
			source: "function Component() {\n  try {\n    function Inner() {\n      return <div />;\n    }\n    return <Inner />;\n  } catch {\n    return null;\n  }\n}\n",
			count:  1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ErrorBoundaries, "component.tsx", testCase.source)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "jsxInTryStatement"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestErrorBoundariesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []errorBoundariesCase{
		{
			name:   "upstreamTodoTryWithFinallyNoCatch",
			source: "function Component(props) {\n  let el;\n  try {\n    el = <div />;\n  } finally {\n    console.log(el);\n  }\n  return el;\n}\n",
		},
		{
			name:   "upstreamTodoCatchInOuterTryWithFinally",
			source: "import {identity} from 'shared-runtime';\n\nfunction Component(props) {\n  let el;\n  try {\n    let value;\n    try {\n      value = identity(props.foo);\n    } catch {\n      el = <div value={value} />;\n    }\n  } finally {\n    console.log(el);\n  }\n  return el;\n}\n",
		},
		{
			name:   "jsxInCatchWithNoOuterTry",
			source: "function Component() {\n  try {\n    doSomething();\n  } catch {\n    return <ErrorMessage />;\n  }\n  return null;\n}\n",
		},
		{
			name:   "jsxInFinallyBlock",
			source: "function Component() {\n  let el;\n  try {\n    foo();\n  } catch {\n    bar();\n  } finally {\n    el = <div />;\n  }\n  return el;\n}\n",
		},
		{
			name:   "jsxInCatchWithFinallySameStatement",
			source: "function Component() {\n  let el;\n  try {\n    foo();\n  } catch {\n    el = <div />;\n  } finally {\n    log();\n  }\n  return el;\n}\n",
		},
		{
			name:   "jsxAfterTryStatement",
			source: "function Component() {\n  try {\n    foo();\n  } catch {\n    bar();\n  }\n  return <div />;\n}\n",
		},
		{
			// A METHOD is a function boundary too, and these are what prove it. The arrow and
			// function-expression cases nearby cannot: a mutant that dropped methods and accessors
			// from the boundary set survived every one of them, because those shapes are bounded by
			// an arm the mutant kept.
			//
			// All three measured CLEAN on React, which is the same reason the arrow cases are
			// clean: the inner function is its own compilation unit, so its JSX sits in no try.
			name:   "jsxInObjectMethodInsideTry",
			source: "function Component() {\n  let el;\n  try {\n    const o = {\n      render() {\n        return <div />;\n      },\n    };\n    el = o.render();\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
		},
		{
			name:   "jsxInGetterInsideTry",
			source: "function Component() {\n  let el;\n  try {\n    const o = {\n      get x() {\n        return <div />;\n      },\n    };\n    el = o.x;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
		},
		{
			name:   "jsxInClassMethodInsideTry",
			source: "function Component() {\n  let el;\n  try {\n    class K {\n      render() {\n        return <div />;\n      }\n    }\n    el = new K().render();\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
		},
		{
			name:   "jsxInArrowInsideTry",
			source: "function Component() {\n  try {\n    const f = () => <div />;\n    return f();\n  } catch {\n    return null;\n  }\n}\n",
		},
		{
			name:   "jsxInFunctionExpressionInsideTry",
			source: "function Component() {\n  try {\n    const f = function () {\n      return <div />;\n    };\n    return f();\n  } catch {\n    return null;\n  }\n}\n",
		},
		{
			name:   "jsxInCallbackInsideTry",
			source: "function Component() {\n  try {\n    return items.map((i) => <li key={i} />);\n  } catch {\n    return null;\n  }\n}\n",
		},
		{
			name:   "lowercaseFunctionName",
			source: "function widget() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "underscorePrefixedName",
			source: "function _Private() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "dollarPrefixedName",
			source: "function $Dollar() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "bareUsePrefixLowercase",
			source: "function usething() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "anonymousDefaultExport",
			source: "export default function (props) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "classMethodNamedComponent",
			source: "class Foo {\n  Component() {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      return null;\n    }\n    return el;\n  }\n}\n",
		},
		{
			// The class arm of the candidacy walk, isolated. An arrow inside a class STATIC BLOCK
			// is bound to a variable, so it has a readable name and passes every later gate; the
			// only thing declining it is its class ancestor. The property-arrow cases below cannot
			// see that arm at all, because a property holds no name the rule can read and the name
			// lookup declines them first.
			//
			// Found by reading after a fixture written on a wrong hypothesis failed to kill the
			// mutant a second time, which is the brief's own instruction. Both spellings of the
			// class are pinned, because the mutant kept the declaration arm and dropped only the
			// expression one. Measured CLEAN on React before either was written.
			name:   "componentArrowInsideAClassExpressionStaticBlock",
			source: "const K = class {\n  static {\n    const Component = () => {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      el = null;\n    }\n    log(el);\n    return el;\n    };\n  }\n};\n",
		},
		{
			name:   "componentArrowInsideAClassDeclarationStaticBlock",
			source: "class K {\n  static {\n    const Component = () => {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      el = null;\n    }\n    log(el);\n    return el;\n    };\n  }\n}\n",
		},
		{
			// A class PROPERTY holding an arrow reaches the class node with no function-like
			// ancestor in between, which is the only shape where the class arm of the candidacy
			// walk decides anything on its own. The method cases nearby are declined by their
			// method ancestor first, so they cannot see it.
			//
			// Written for a surviving mutant that dropped `KindClassExpression` from that arm, and
			// measured CLEAN on React first. Both spellings of the class are pinned, because the
			// mutant kept the declaration arm and dropped only the expression one.
			name:   "classExpressionPropertyArrowNamedComponent",
			source: "const K = class {\n  Component = () => {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      el = null;\n    }\n    log(el);\n    return el;\n  };\n};\n",
		},
		{
			name:   "classDeclarationPropertyArrowNamedComponent",
			source: "class K {\n  Component = () => {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      el = null;\n    }\n    log(el);\n    return el;\n  };\n}\n",
		},
		{
			name:   "objectMethodNamedComponent",
			source: "const o = {\n  Component() {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      return null;\n    }\n    return el;\n  },\n};\n",
		},
		{
			name:   "moduleScopeTry",
			source: "let el;\ntry {\n  el = <div />;\n} catch {\n  el = null;\n}\n",
		},
		{
			// The upward walk must stop at the enclosing function, and this is the input that
			// proves it. The declaration sits lexically inside a module-scope try block and it
			// PASSES the component gate, so nothing else declines it: only the boundary does.
			//
			// The nested-function cases above cannot see that, because their inner functions are
			// declined by the gate first and both a stopping walk and a running one reach silence
			// by different routes. Found by a surviving mutant that removed the boundary test, and
			// measured CLEAN on React before this was written down rather than after.
			name:   "componentDeclaredInsideAModuleScopeTryBlock",
			source: "try {\n  function Component() {\n    return <div />;\n  }\n} catch {\n  log();\n}\n",
		},
		{
			// The same shape with the JSX bound to a local first, so the walk crosses one more
			// parent before reaching the function boundary.
			name:   "componentDeclaredInsideAModuleScopeTryBlockWithLocal",
			source: "try {\n  function Component() {\n    let el = <div />;\n    return el;\n  }\n} catch {\n  log();\n}\n",
		},
		{
			name:   "componentNestedInsideAnotherFunction",
			source: "function Outer() {\n  function Component() {\n    let el;\n    try {\n      el = <div />;\n    } catch {\n      return null;\n    }\n    return el;\n  }\n  return <Component />;\n}\n",
		},
		{
			// Three parameters whose SECOND one does mention `ref`. The plain three-parameter case
			// below cannot see a mutant that widens the arity test, because its second parameter
			// fails the ref test anyway and both versions reach silence by different routes.
			// Measured CLEAN on React before it was written.
			name:   "threeParametersWithARefInSecondPosition",
			source: "function Component(props, ref, third) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return el;\n}\n",
		},
		{
			name:   "threeParameters",
			source: "function Component(a, b, c) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "restParameter",
			source: "function Component(...props) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			name:   "secondParameterNotNamedRef",
			source: "function Component(props, other) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n",
		},
		{
			// `returnsNonNode`, isolated. The JSX value is consumed by `log(el)`, so upstream's
			// dead-value elimination cannot be what silences this, and the last return visited
			// hands back an object literal, which is on upstream's non-node list. Measured clean
			// on the real rule; the same body ending `return el` reports.
			name:   "lastReturnIsObjectLiteralWhileValueIsConsumed",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return {};\n}\n",
		},
		{
			// The same discrimination in the other direction: a non-node return visited BEFORE a
			// node one does not decide the answer, because upstream assigns rather than
			// accumulates. Measured reporting, so this belongs among the silent cases only as its
			// mirror; it lives in the Fires table.
			name:   "lastReturnIsNewExpressionWhileValueIsConsumed",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return new X();\n}\n",
		},
		{
			// A bare `return;` is `isNonNode(null)`, which is true. Measured clean.
			name:   "lastReturnIsBareReturnWhileValueIsConsumed",
			source: "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    el = null;\n  }\n  log(el);\n  return;\n}\n",
		},
		{
			// The concise-body arrow branch of `returnsNonNode`, isolated, and the fixture is here
			// because a confident unreachability argument was wrong. The reasoning ran: a concise
			// arrow's body is an expression, a try is a statement, so no try can sit between the
			// JSX and the arrow and the branch cannot be reached. A class STATIC BLOCK inside a
			// class EXPRESSION is a statement position inside an expression, which breaks it, and
			// the rule's own two predicates were run over this input rather than the argument being
			// trusted: the walk does find the try and the gate is asked.
			//
			// It is clean because the arrow's concise body is a class expression, which is on
			// upstream's non-node list, so the arrow is not a component. Measured CLEAN on React.
			// Deleting the branch makes this report.
			name:   "conciseArrowBodyIsAClassExpressionHoldingAStaticBlockTry",
			source: "const Component = () => class {\n  static {\n    try {\n      foo(<div />);\n    } catch {\n      bar();\n    }\n  }\n};\n",
		},
		{
			name:   "lowercaseArrowAssignment",
			source: "const widget = (props) => {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n};\n",
		},
		{
			name:   "tryWithFinallyOnlyNoCatchNestedJsx",
			source: "function Component() {\n  let el;\n  try {\n    if (x) {\n      el = <div />;\n    }\n  } finally {\n    log();\n  }\n  return el;\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ErrorBoundaries, "component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestErrorBoundariesSpan pins WHERE the finding points, which no message-id assertion can see.
//
// The span is byte-verified against React's own golden rather than chosen. React's
// `invalid-jsx-in-try-with-catch.expect.md` logs the finding at `"index":123` through
// `"index":130`; slicing that fixture's bytes at those offsets yields `<div />`, seven bytes. The
// source below is that fixture with its pragma comment removed, so the offsets differ while the
// reported text does not, and the text is what this asserts.
func TestErrorBoundariesSpan(t *testing.T) {
	t.Parallel()

	source := "function Component(props) {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n"
	result := rule_testing.Run(t, ErrorBoundaries, "component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	// A literal typed here rather than anything derived from the rule, so the assertion cannot
	// move together with the code it guards.
	if reported != "<div />" {
		t.Errorf("finding points at %q, want %q", reported, "<div />")
	}
}

// TestErrorBoundariesSpanOnNestedElements pins that a nested element reports separately from its
// parent, and that each finding points at its own element rather than both at the outer one.
//
// Measured on React first: `<div><span /></div>` inside a try produces two findings, one spanning
// the whole outer element and one spanning the inner.
func TestErrorBoundariesSpanOnNestedElements(t *testing.T) {
	t.Parallel()

	source := "function Component() {\n  let el;\n  try {\n    el = <div><span /></div>;\n  } catch {\n    return null;\n  }\n  return el;\n}\n"
	result := rule_testing.Run(t, ErrorBoundaries, "component.tsx", source)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("expected two findings, got %d", len(result.Diagnostics))
	}
	got := []string{}
	for _, diagnostic := range result.Diagnostics {
		got = append(got, source[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	want := []string{"<div><span /></div>", "<span />"}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("finding %d points at %q, want %q", index, got[index], want[index])
		}
	}
}

// TestErrorBoundariesMessage asserts the message identity and its text against literals typed here.
//
// `rule.Message` is `{Id, Description}` with no interpolation, so there is no rendered text to
// guard and equality on both fields is the whole assertion. Asserting against the rule's own
// constant would move with it under mutation and guard nothing.
func TestErrorBoundariesMessage(t *testing.T) {
	t.Parallel()

	if messageJsxInTryStatement.Id != "jsxInTryStatement" {
		t.Errorf("message id is %q", messageJsxInTryStatement.Id)
	}
	if !strings.HasPrefix(messageJsxInTryStatement.Description, "This JSX is constructed inside a try block") {
		t.Errorf("message description opens %q", messageJsxInTryStatement.Description[:40])
	}
	if !strings.Contains(messageJsxInTryStatement.Description, "error boundary") {
		t.Error("the description never names the mechanism that does catch rendering errors")
	}
}

// TestErrorBoundariesRunsWithoutTheTypeChecker pins that this rule needs no checker.
//
// It resolves nothing: the try question is answered from the parent chain and the component gate
// from names, parameters and returns. A later revert adding a checker call would go silent under
// the plain harness rather than failing, which is why this is asserted rather than assumed.
func TestErrorBoundariesRunsWithoutTheTypeChecker(t *testing.T) {
	t.Parallel()

	if ErrorBoundaries.NeedsTypeChecker {
		t.Error("the rule declares a type checker it never asks anything")
	}
	source := "function Component() {\n  let el;\n  try {\n    el = <div />;\n  } catch {\n    return null;\n  }\n  return el;\n}\n"
	result := rule_testing.Run(t, ErrorBoundaries, "component.tsx", source)
	rule_testing.ExpectFindings(t, result, "jsxInTryStatement")
}
