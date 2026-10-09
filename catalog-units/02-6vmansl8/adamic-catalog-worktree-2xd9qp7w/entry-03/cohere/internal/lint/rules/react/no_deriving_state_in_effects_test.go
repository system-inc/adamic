package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is 22 fixtures and every verdict below was MEASURED before it was written down.
//
// React ships 21 fixtures in `__tests__/fixtures/compiler/effect-derived-computations/` and exactly
// one more, `error.invalid-derived-computation-in-effect.js`, at the top of the fixture tree. Those
// two groups test DIFFERENT passes and that is the single most important fact about this file.
//
// All 21 in the directory carry `@validateNoDerivedComputationsInEffects_exp`, the experimental
// 842-line pass. The one at the top carries `@validateNoDerivedComputationsInEffects`, the 229-line
// pass the shipped ESLint plugin actually runs. So 21 of 22 fixture NAMES describe a verdict that is
// not this rule's, and five of them invert: three named `-no-error` report under the shipped rule,
// and `invalid-derived-computation-in-effect` is clean.
//
// Every case was therefore run through React 7.1.1's own `react-hooks/no-deriving-state-in-effects`
// via the ESLint Linter API before being written here, and the measured verdict is what each case
// asserts. The names are kept as upstream wrote them, inversions and all, because renaming them to
// match would hide the fact that these fixtures came from the other pass.
//
// Two parsers were run over the whole corpus, espree and typescript-eslint, and every verdict was
// identical. That was not cosmetic: one fixture is typed TypeScript and produced a parse error
// rather than a verdict under the default parser, which reads as `clean` and is not a measurement.
//
// # Why these fixtures carry a react.d.ts and upstream's do not
//
// Upstream recognises `useEffect` and a `useState` setter through a builtin shape registry keyed on
// an import from the literal module `react`. This rule reads the type checker instead, exactly as
// `set-state-in-effect` does, so its fixtures need real declarations. The CASE TEXT is upstream's
// byte for byte; only the module specifier moves, from 'react' to "./react", and a script verified
// that the rewrite touched specifier lines and nothing else on all 22 files.
//
// That rewrite is also a MEASUREMENT rather than a formality, and it is this port's headline
// divergence. Running the shipped rule over the corpus twice, once with each specifier, four of the
// five reporting fixtures go silent under "./react": upstream keys on the specifier and this rule
// keys on the resolved declaration. The verdicts asserted below are the ones measured with
// upstream's own specifier, which is the judgment being ported.

type derivedEffectCase struct {
	name     string
	file     string
	source   string
	findings int
}

// `RunTypedFiles` rather than `RunTyped` because both of this rule's predicates are type questions,
// and the harness pins `types: []` in its tsconfig so a real `@types/react` on disk can never leak
// in and make a fixture pass for a reason this file did not state.
//
// `reactStub` and `otherModuleStub` are `set_state_in_effect_test.go`'s, shared rather than copied:
// the two rules ask the checker the same two questions through the same two helpers, so a stub that
// drifted between them would make one rule's fixtures stop testing the other's predicate.
func runNoDerivingStateInEffects(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, NoDerivingStateInEffects, map[string]string{
		"react.d.ts": reactStub,
		"other.d.ts": otherModuleStub,
		"a.tsx":      source,
	}, "a.tsx")
}

func TestNoDerivingStateInEffectsFires(t *testing.T) {
	t.Parallel()

	// Written as a literal rather than referenced through the rule's own message constant, so the
	// assertion cannot move together with the code it guards.
	const noDerivingStateInEffects = "noDerivingStateInEffects"

	cases := []derivedEffectCase{
		{
			name:     "derivedStateFromPropSetterCallOutsideEffectNoError",
			file:     "derived-state-from-prop-setter-call-outside-effect-no-error.jsx",
			source:   "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({initialName}) {\n  const [name, setName] = useState('');\n\n  useEffect(() => {\n    setName(initialName);\n  }, [initialName]);\n\n  return (\n    <div>\n      <input value={name} onChange={e => setName(e.target.value)} />\n    </div>\n  );\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{initialName: 'John'}],\n};\n",
			findings: 1,
		},
		{
			name:     "derivedStateFromPropSetterUsedOutsideEffectNoError",
			file:     "derived-state-from-prop-setter-used-outside-effect-no-error.jsx",
			source:   "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction MockComponent({onSet}) {\n  return <div onClick={() => onSet('clicked')}>Mock Component</div>;\n}\n\nfunction Component({propValue}) {\n  const [value, setValue] = useState(null);\n  useEffect(() => {\n    setValue(propValue);\n  }, [propValue]);\n\n  return <MockComponent onSet={setValue} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{propValue: 'test'}],\n};\n",
			findings: 1,
		},
		{
			name:     "effectWithGlobalFunctionCallNoError",
			file:     "effect-with-global-function-call-no-error.jsx",
			source:   "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({propValue}) {\n  const [value, setValue] = useState(null);\n  useEffect(() => {\n    setValue(propValue);\n    globalCall();\n  }, [propValue]);\n\n  return <div>{value}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{propValue: 'test'}],\n};\n",
			findings: 1,
		},
		{
			name:     "baseinvalidDerivedComputationInEffect",
			file:     "zz.base-error.invalid-derived-computation-in-effect.jsx",
			source:   "// @validateNoDerivedComputationsInEffects\nimport {useEffect, useState} from \"./react\";\n\nfunction BadExample() {\n  const [firstName, setFirstName] = useState('Taylor');\n  const [lastName, setLastName] = useState('Swift');\n\n  // 🔴 Avoid: redundant state and unnecessary Effect\n  const [fullName, setFullName] = useState('');\n  useEffect(() => {\n    setFullName(firstName + ' ' + lastName);\n  }, [firstName, lastName]);\n\n  return <div>{fullName}</div>;\n}\n",
			findings: 1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoDerivingStateInEffects(t, testCase.source)
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = noDerivingStateInEffects
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

func TestNoDerivingStateInEffectsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []derivedEffectCase{
		{
			name:   "derivedStateConditionallyInEffect",
			file:   "derived-state-conditionally-in-effect.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({value, enabled}) {\n  const [localValue, setLocalValue] = useState('');\n\n  useEffect(() => {\n    if (enabled) {\n      setLocalValue(value);\n    } else {\n      setLocalValue('disabled');\n    }\n  }, [value, enabled]);\n\n  return <div>{localValue}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{value: 'test', enabled: true}],\n};\n",
		},
		{
			name:   "derivedStateFromDefaultProps",
			file:   "derived-state-from-default-props.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nexport default function Component({input = 'empty'}) {\n  const [currInput, setCurrInput] = useState(input);\n  const localConst = 'local const';\n\n  useEffect(() => {\n    setCurrInput(input + localConst);\n  }, [input, localConst]);\n\n  return <div>{currInput}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{input: 'test'}],\n};\n",
		},
		{
			name:   "derivedStateFromLocalStateInEffect",
			file:   "derived-state-from-local-state-in-effect.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\n\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({shouldChange}) {\n  const [count, setCount] = useState(0);\n\n  useEffect(() => {\n    if (shouldChange) {\n      setCount(count + 1);\n    }\n  }, [count]);\n\n  return <div>{count}</div>;\n}\n",
		},
		{
			name:   "derivedStateFromPropLocalStateAndComponentScope",
			file:   "derived-state-from-prop-local-state-and-component-scope.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({firstName}) {\n  const [lastName, setLastName] = useState('Doe');\n  const [fullName, setFullName] = useState('John');\n\n  const middleName = 'D.';\n\n  useEffect(() => {\n    setFullName(firstName + ' ' + middleName + ' ' + lastName);\n  }, [firstName, middleName, lastName]);\n\n  return (\n    <div>\n      <input value={lastName} onChange={e => setLastName(e.target.value)} />\n      <div>{fullName}</div>\n    </div>\n  );\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{firstName: 'John'}],\n};\n",
		},
		{
			name:   "derivedStateFromPropSetterTernary",
			file:   "derived-state-from-prop-setter-ternary.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @outputMode:\"lint\"\n\nfunction Component({value}) {\n  const [checked, setChecked] = useState('');\n\n  useEffect(() => {\n    setChecked(value === '' ? [] : value.split(','));\n  }, [value]);\n\n  return <div>{checked}</div>;\n}\n",
		},
		{
			name:   "derivedStateFromPropWithSideEffect",
			file:   "derived-state-from-prop-with-side-effect.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({value}) {\n  const [localValue, setLocalValue] = useState('');\n\n  useEffect(() => {\n    setLocalValue(value);\n    document.title = `Value: ${value}`;\n  }, [value]);\n\n  return <div>{localValue}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{value: 'test'}],\n};\n",
		},
		{
			name:   "derivedStateFromRefAndStateNoError",
			file:   "derived-state-from-ref-and-state-no-error.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState, useRef} from \"./react\";\n\nexport default function Component({test}) {\n  const [local, setLocal] = useState('');\n\n  const myRef = useRef(null);\n\n  useEffect(() => {\n    setLocal(myRef.current + test);\n  }, [test]);\n\n  return <>{local}</>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{test: 'testString'}],\n};\n",
		},
		{
			name:   "effectContainsLocalFunctionCall",
			file:   "effect-contains-local-function-call.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({propValue}) {\n  const [value, setValue] = useState(null);\n\n  function localFunction() {\n    console.log('local function');\n  }\n\n  useEffect(() => {\n    setValue(propValue);\n    localFunction();\n  }, [propValue]);\n\n  return <div>{value}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{propValue: 'test'}],\n};\n",
		},
		{
			name:   "effectContainsPropFunctionCallNoError",
			file:   "effect-contains-prop-function-call-no-error.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component({propValue, onChange}) {\n  const [value, setValue] = useState(null);\n  useEffect(() => {\n    setValue(propValue);\n    onChange();\n  }, [propValue]);\n\n  return <div>{value}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{propValue: 'test', onChange: () => {}}],\n};\n",
		},
		{
			name:   "effectUsedInDepArrayStillErrors",
			file:   "effect-used-in-dep-array-still-errors.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\n\nfunction Component({prop}) {\n  const [s, setS] = useState(0);\n  useEffect(() => {\n    setS(prop);\n  }, [prop, setS]);\n\n  return <div>{prop}</div>;\n}\n",
		},
		{
			name:   "effectWithCleanupFunctionDependingOnDerivedComputationValue",
			file:   "effect-with-cleanup-function-depending-on-derived-computation-value.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\n\nimport {useEffect, useState} from \"./react\";\n\nfunction Component(file: File) {\n  const [imageUrl, setImageUrl] = useState(null);\n\n  /*\n   * Cleaning up the variable or a source of the variable used to setState\n   * inside the effect communicates that we always need to clean up something\n   * which is a valid use case for useEffect. In which case we want to\n   * avoid an throwing\n   */\n  useEffect(() => {\n    const imageUrlPrepared = URL.createObjectURL(file);\n    setImageUrl(imageUrlPrepared);\n    return () => URL.revokeObjectURL(imageUrlPrepared);\n  }, [file]);\n\n  return <Image src={imageUrl} xstyle={styles.imageSizeLimits} />;\n}\n",
		},
		{
			name:   "fromPropsSetstateInEffectNoError",
			file:   "from-props-setstate-in-effect-no-error.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @enableTreatSetIdentifiersAsStateSetters @loggerTestOnly @outputMode:\"lint\"\n\nfunction Component({setParentState, prop}) {\n  useEffect(() => {\n    setParentState(prop);\n  }, [prop]);\n\n  return <div>{prop}</div>;\n}\n",
		},
		{
			name:   "functionExpressionMutationEdgeCase",
			file:   "function-expression-mutation-edge-case.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\n\nfunction Component() {\n  const [foo, setFoo] = useState({});\n  const [bar, setBar] = useState(new Set());\n\n  /*\n   * isChanged is considered context of the effect's function expression,\n   * if we don't bail out of effect mutation derivation tracking, isChanged\n   * will inherit the sources of the effect's function expression.\n   *\n   * This is innacurate and with the multiple passes ends up causing an infinite loop.\n   */\n  useEffect(() => {\n    let isChanged = false;\n\n    const newData = foo.map(val => {\n      bar.someMethod(val);\n      isChanged = true;\n    });\n\n    if (isChanged) {\n      setFoo(newData);\n    }\n  }, [foo, bar]);\n\n  return (\n    <div>\n      {foo}, {bar}\n    </div>\n  );\n}\n",
		},
		{
			name:   "invalidDerivedComputationInEffect",
			file:   "invalid-derived-computation-in-effect.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [firstName, setFirstName] = useState('Taylor');\n  const lastName = 'Swift';\n\n  // 🔴 Avoid: redundant state and unnecessary Effect\n  const [fullName, setFullName] = useState('');\n  useEffect(() => {\n    setFullName(firstName + ' ' + lastName);\n  }, [firstName, lastName]);\n\n  return <div>{fullName}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [],\n};\n",
		},
		{
			name:   "invalidDerivedStateFromComputedProps",
			file:   "invalid-derived-state-from-computed-props.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nexport default function Component(props) {\n  const [displayValue, setDisplayValue] = useState('');\n\n  useEffect(() => {\n    const computed = props.prefix + props.value + props.suffix;\n    setDisplayValue(computed);\n  }, [props.prefix, props.value, props.suffix]);\n\n  return <div>{displayValue}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{prefix: '[', value: 'test', suffix: ']'}],\n};\n",
		},
		{
			name:   "invalidDerivedStateFromDestructuredProps",
			file:   "invalid-derived-state-from-destructured-props.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState} from \"./react\";\n\nexport default function Component({props}) {\n  const [fullName, setFullName] = useState(\n    props.firstName + ' ' + props.lastName\n  );\n\n  useEffect(() => {\n    setFullName(props.firstName + ' ' + props.lastName);\n  }, [props.firstName, props.lastName]);\n\n  return <div>{fullName}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{props: {firstName: 'John', lastName: 'Doe'}}],\n};\n",
		},
		{
			name:   "refConditionalInEffectNoError",
			file:   "ref-conditional-in-effect-no-error.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\nimport {useEffect, useState, useRef} from \"./react\";\n\nexport default function Component({test}) {\n  const [local, setLocal] = useState(0);\n\n  const myRef = useRef(null);\n\n  useEffect(() => {\n    if (myRef.current) {\n      setLocal(test);\n    } else {\n      setLocal(test + test);\n    }\n  }, [test]);\n\n  return <>{local}</>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{test: 4}],\n};\n",
		},
		// Upstream REPORTS this one and this rule is silent, and the cause is above the rule rather
		// than in it: the fixture has NO import line at all, so `useState`, `useEffect` and `setS`
		// are undeclared globals that the checker types as `any` with no symbol. Both of this rule's
		// predicates correctly decline.
		//
		// Measured rather than assumed, by adding the one missing import line and running the shipped
		// rule and this one on the result: both report, at the same span. So the shape is one this
		// rule handles and the silence is the corpus omitting its declarations.
		//
		// Recorded here as silent, which is what this rule DOES, rather than deleted or forced green.
		// It is the same trade `set-state-in-effect` documents: a dependency on `@types/react`
		// instead of a builtin shape registry, better on application code and worse on a corpus
		// written for a compiler that needed no types.
		{
			name:   "usestateDerivedFromPropNoShowInDataFlowTree",
			file:   "usestate-derived-from-prop-no-show-in-data-flow-tree.jsx",
			source: "// @validateNoDerivedComputationsInEffects_exp @loggerTestOnly @outputMode:\"lint\"\n\nfunction Component({prop}) {\n  const [s, setS] = useState();\n  const [second, setSecond] = useState(prop);\n\n  /*\n   * `second` is a source of state. It will inherit the value of `prop` in\n   * the first render, but after that it will no longer be updated when\n   * `prop` changes. So we shouldn't consider `second` as being derived from\n   * `prop`\n   */\n  useEffect(() => {\n    setS(second);\n  }, [second]);\n\n  return <div>{s}</div>;\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoDerivingStateInEffects(t, testCase.source))
		})
	}
}

// Cases this rule's own shape demands and the corpus does not cover, each MEASURED against React
// 7.1.1 on that exact input before being written here.
//
// The corpus is 22 fixtures written for a different pass, so it says nothing about most of the
// distinctions this rule actually makes. Two of these were surprises that reading had got wrong.
//
// **`useLayoutEffect` is silent**, and that is the one that would have shipped a false positive.
// This rule's first draft reused `effectHookKind` from `set_state_in_effect.go`, which accepts
// `useEffect`, `useLayoutEffect` and `useInsertionEffect` because ITS pass calls all three
// predicates. This pass calls `isUseEffectHookType` alone. Two rules in one family, one hook apart,
// and no imported fixture writes a layout effect into a derived computation, so nothing in the
// corpus could see it.
//
// **An inline callback and an inline array still report.** Upstream reads both arguments out of maps
// keyed by identifier, which reads as though a literal written at the call site cannot be a
// candidate. Lowering assigns both to temporaries first, so by the time the pass runs they are
// identifiers. The judgment is not syntactic and this pins it.
func TestNoDerivingStateInEffectsMeasuredCases(t *testing.T) {
	t.Parallel()

	const noDerivingStateInEffects = "noDerivingStateInEffects"

	cases := []derivedEffectCase{
		{name: "branchOnDep", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    if (a) { setV(a + b); }\n  }, [a, b]);\n  return v;\n}\n", findings: 0},
		{name: "dispatchNotSetter", source: "import {useEffect, useReducer} from \"./react\";\nfunction Component({a, b}) {\n  const [v, dispatch] = useReducer((s, x) => s, '');\n  useEffect(() => {\n    dispatch(a + b);\n  }, [a, b]);\n  return v;\n}\n", findings: 0},
		{name: "emptyDeps", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a);\n  }, []);\n  return v;\n}\n", findings: 0},
		{name: "inlineBoth", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n", findings: 1},
		{name: "layoutEffect", source: "import {useLayoutEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useLayoutEffect(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n", findings: 0},
		{name: "namespace", source: "import * as React from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = React.useState('');\n  React.useEffect(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n", findings: 1},
		{name: "noDepArray", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a);\n  });\n  return v;\n}\n", findings: 0},
		{name: "oneOfTwoDeps", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a);\n  }, [a, b]);\n  return v;\n}\n", findings: 0},
		{name: "renamedImport", source: "import {useEffect as useFx, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useFx(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n", findings: 1},
		{name: "twoSetters", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  const [w, setW] = useState('');\n  useEffect(() => {\n    setV(a + b);\n    setW(a + b);\n  }, [a, b]);\n  return v + w;\n}\n", findings: 2},
		// A SPREAD callback argument. Upstream tests `value.args[0].kind === 'Identifier'`, which a
		// spread element is not, and this rule tests `args[0].Spread`. The two agree and both are
		// silent.
		//
		// This case exists because the guard's mutant SURVIVED the whole suite, and the first two
		// hypotheses about why were both wrong. `useEffect(...args)` lowers to a call with ONE
		// argument, so the length test already declines it; `useEffect(cb, ...[a, b])` reaches the
		// guard but the mutant then misses `functions`, because `cb` is a StoreLocal and the map is
		// keyed by the FunctionExpression's own lvalue. Both versions reach silence by different
		// routes, which is why no fixture over those shapes could ever see the mutation.
		//
		// The shape below is the one that separates them, found by printing the lowered form rather
		// than by reasoning a third time: `Call $28(...$29, $32)` where `$29` is the
		// FunctionExpression's lvalue, so `functions` hits AND `args[0].Spread` is true. Measured
		// both ways: pristine 0 findings, mutant 1. React 7.1.1 is silent on it.
		{name: "spreadCallbackArgument", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(...(() => { setV(a + b); }), [a, b]);\n  return v;\n}\n", findings: 0},
		// The taint COUNT test, and the fixture that finally reaches it.
		//
		// `oneOfTwoDeps` above reads as though it covers this and does not: with `b` declared as a
		// dep but never read, the callback does not capture it and the "every dep is captured"
		// check declines the effect before the count is ever compared. The lowered form says so
		// directly, `captures=[setV a]`.
		//
		// To reach the count test every dep must be captured while the SETTER's argument carries
		// taint from only some of them, which is what `log(b)` does here. Measured against React
		// 7.1.1: the partial case is clean and the full case reports, and the count mutant reports
		// on BOTH.
		{name: "everyDepCapturedButSetterReadsOnlyOne", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => { log(b); setV(a); }, [a, b]);\n  return v;\n}\n", findings: 0},
		{name: "everyDepCapturedAndSetterReadsBoth", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => { log(b); setV(a + b); }, [a, b]);\n  return v;\n}\n", findings: 1},
		// A LOOP inside the effect body, which is what the back-edge guard is for.
		//
		// Upstream refuses to analyse a block whose predecessor it has not yet seen, so an effect
		// containing any loop is abandoned rather than judged. Both of these are clean under React
		// 7.1.1 and clean here.
		//
		// Two loop kinds rather than one, because the guard's mutant survives a `for` loop and dies
		// on `while` and `do-while`: the counted loop declines earlier for its own reasons, in the
		// allowed-instruction switch, so a fixture using only `for` would read as coverage while
		// testing nothing. Measured both ways, pristine 0 and mutant 1 on these two.
		{name: "whileLoopInEffectBody", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a}) {\n  const [v, setV] = useState('');\n  useEffect(() => { while (g()) { h(); } setV(a); }, [a]);\n  return v;\n}\n", findings: 0},
		{name: "doWhileLoopInEffectBody", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a}) {\n  const [v, setV] = useState('');\n  useEffect(() => { do { h(); } while (g()); setV(a); }, [a]);\n  return v;\n}\n", findings: 0},
		// A local dependency whose initializer is NOT a literal. This is the other side of
		// `dependencyIsFoldedConstant`: upstream folds `const x = 'lit'` away and leaves
		// `const x = compute()` alone, so this reports where the constant cases are silent.
		// Measured against React 7.1.1, which reports it. Without this the primitive test in that
		// helper could be replaced by `return true` and nothing would notice, which would silence
		// every effect with a local dep.
		{name: "localDependencyFromANonLiteralInitializer", source: "import {useEffect, useState} from \"./react\";\nfunction Component({firstName}) {\n  const lastName = compute();\n  const [fullName, setFullName] = useState('');\n  useEffect(() => {\n    setFullName(firstName + ' ' + lastName);\n  }, [firstName, lastName]);\n  return <div>{fullName}</div>;\n}\n", findings: 1},
		// The setter's ARGUMENT COUNT. Upstream tests `value.args.length === 1`, so a setter called
		// with two arguments is not the shape it judges and the effect is abandoned. Both of these
		// are clean under React 7.1.1 and clean here, and each separately kills the mutant that
		// drops the argument test.
		{name: "setterCalledWithTwoArguments", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a + b, a);\n  }, [a, b]);\n  return v;\n}\n", findings: 0},
		{name: "setterCalledWithASpreadArgument", source: "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(...[a + b]);\n  }, [a, b]);\n  return v;\n}\n", findings: 0},
		// A component declared INSIDE another function. Lowering puts it in the enclosing function's
		// arena rather than walking into it, so the rule recurses through `function.Functions` to
		// reach it. That recursion's mutant survived every other fixture in this file, because
		// nothing else here nests a component. React 7.1.1 reports this one.
		{name: "effectInsideANestedComponent", source: "import {useEffect, useState} from \"./react\";\nfunction Outer() {\n  function Inner({a, b}) {\n    const [v, setV] = useState('');\n    useEffect(() => {\n      setV(a + b);\n    }, [a, b]);\n    return v;\n  }\n  return Inner;\n}\n", findings: 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoDerivingStateInEffects(t, testCase.source)
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = noDerivingStateInEffects
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestNoDerivingStateInEffectsPointsAtTheSetterCall pins WHERE the finding lands.
//
// Upstream reports at `instr.value.callee.loc`, the setter's reference inside the effect. Three
// anchors were plausible and all three produce the same message id, so no `ExpectFindings` fixture
// in this file can tell them apart: the setter's reference, the setter's BINDING in the `useState`
// destructure, and the `useEffect` call itself. Each expectation below was read off React 7.1.1's
// own caret span on that input.
//
// The expectation is sliced from the source the HARNESS wrote rather than from the Go literal.
// `rule_testing.RunTypedFiles` writes `strings.TrimSpace(contents)+"\n"`, so a fixture carrying a
// leading newline sits one byte off from its literal and an assertion built from the literal reports
// a span shifted by one while the rule is correct.
func TestNoDerivingStateInEffectsPointsAtTheSetterCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		wantText []string
	}{
		{
			name:     "directCall",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [fullName, setFullName] = useState('');\n  useEffect(() => {\n    setFullName(a + b);\n  }, [a, b]);\n  return fullName;\n}\n",
			wantText: []string{"setFullName"},
		},
		// Two setters in one effect report twice, each at its own call. Upstream collects every
		// location rather than returning the first, which is the opposite of the sibling rule
		// `set-state-in-effect`, whose `getSetStateCall` stops at the first setter it finds.
		// Measured: React reports both.
		{
			name:     "twoSettersReportAtTheirOwnCalls",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  const [w, setW] = useState('');\n  useEffect(() => {\n    setV(a + b);\n    setW(a + b);\n  }, [a, b]);\n  return v + w;\n}\n",
			wantText: []string{"setV", "setW"},
		},
		// A namespace member is reported at the member name, not at `React`.
		{
			name:     "namespaceMember",
			source:   "import * as React from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = React.useState('');\n  React.useEffect(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n",
			wantText: []string{"setV"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoDerivingStateInEffects(t, testCase.source)
			if len(result.Diagnostics) != len(testCase.wantText) {
				t.Fatalf("expected %d findings, got %d", len(testCase.wantText), len(result.Diagnostics))
			}
			written := strings.TrimSpace(testCase.source) + "\n"
			for index, want := range testCase.wantText {
				span := result.Diagnostics[index].Range
				if span.Pos() < 0 || span.End() > len(written) || span.Pos() >= span.End() {
					t.Fatalf("finding %d range %d..%d is outside the %d byte file", index, span.Pos(), span.End(), len(written))
				}
				if got := written[span.Pos():span.End()]; got != want {
					t.Errorf("finding %d points at %q, want %q", index, got, want)
				}
			}
		})
	}
}

// TestNoDerivingStateInEffectsMessage asserts the rendered message exactly.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so the two fields are compared
// directly. Asserted against literals typed here rather than against the rule's own constant:
// comparing a diagnostic to the constant it was reported with is equality that moves on both sides
// under mutation, which is how a message-text mutant survives a test that looks correct.
func TestNoDerivingStateInEffectsMessage(t *testing.T) {
	t.Parallel()

	result := runNoDerivingStateInEffects(t, "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noDerivingStateInEffects" {
		t.Errorf("message id is %q, want %q", got, "noDerivingStateInEffects")
	}
	const wantDescription = "This effect exists only to copy a value into state that was already computable " +
		"from the values it depends on. React renders once with the stale state, runs the effect, " +
		"then renders a second time with the new state, so the user can briefly see the old value " +
		"and every dependent effect runs twice. Compute the value during render instead, as a " +
		"plain constant beside the values it derives from. If the derivation is expensive, wrap " +
		"that constant in `useMemo` rather than storing it in state."
	if got := result.Diagnostics[0].Message.Description; got != wantDescription {
		t.Errorf("message description is:\n%q\nwant:\n%q", got, wantDescription)
	}
}

// TestNoDerivingStateInEffectsRequiresTheTypedHarness pins that this rule is useless without a
// checker, so a later revert of `NeedsTypeChecker` fails loudly rather than going vacuously green.
//
// Both of this rule's predicates are type questions, and `GetSymbolAtLocation` on a nil checker
// returns nil rather than crashing, so a rule that lost its guard would go SILENT rather than
// panic. Silence is the more dangerous failure because every StaysSilent case in this file would
// still pass.
func TestNoDerivingStateInEffectsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "import {useEffect, useState} from \"./react\";\nfunction Component({a, b}) {\n  const [v, setV] = useState('');\n  useEffect(() => {\n    setV(a + b);\n  }, [a, b]);\n  return v;\n}\n"

	typed := runNoDerivingStateInEffects(t, source)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("typed harness: expected 1 finding, got %d", len(typed.Diagnostics))
	}

	// The untyped harness hands the rule a nil checker. The guard makes that silence rather than a
	// panic, and this asserts the silence so the guard's own mutant has something to fail.
	untyped := rule_testing.Run(t, NoDerivingStateInEffects, "a.tsx", source)
	rule_testing.ExpectClean(t, untyped)
}

// TestNoDerivingStateInEffectsJudgesOnlyWhatReactCompiles pins the component-or-hook gate
// (#1pmwkmv), with the same two rows as set-state-in-effect's: a plain function is not compiled, so
// it is silent, and a component nested inside one still is.
func TestNoDerivingStateInEffectsJudgesOnlyWhatReactCompiles(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, runNoDerivingStateInEffects(t, "import {useEffect, useState} from \"./react\";\n\nexport function deriveName(initialName: string) {\n  const [name, setName] = useState('');\n  useEffect(() => {\n    setName(initialName);\n  }, [initialName]);\n  return name;\n}\n"))

	rule_testing.ExpectFindings(t, runNoDerivingStateInEffects(t, "import {useEffect, useState} from \"./react\";\n\nexport function wrapper() {\n  function Component({initialName}: {initialName: string}) {\n    const [name, setName] = useState('');\n    useEffect(() => {\n      setName(initialName);\n    }, [initialName]);\n    return <div>{name}</div>;\n  }\n  return Component;\n}\n"), "noDerivingStateInEffects")
}
