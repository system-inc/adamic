package react

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus for this rule is React's own conformance fixtures, not oxc's tester block.
//
// oxc's `rules/react/refs.rs` carries 1 pass and 3 fail cases and no snapshot at all, because the
// rule file is a four-line dispatcher and the real corpus lives with the compiler. React ships 42
// fixtures whose headline is `Cannot access refs during render`, vendored at
// `internal/react_conformance/testdata/fixtures`. Twelve of those are Flow (`//@flow` plus
// `component C()` syntax, which our parser cannot accept), leaving 30 that are scorable here. That
// 30 is the same number `internal/react_conformance/verdict_test.go` independently asserts for this
// rule, which is the cross-check that the set below is the right set.
//
// Every source string is copied byte-for-byte from the fixture on disk, emitted through
// `json.dumps` so that no escape sequence was ever typed on the way here and nothing could cook a
// `\n` into a real newline. The expected message ids are read from the label under each caret run
// in the fixture's own `.expect.md`, in order, so a fixture reporting twice carries two ids.
//
// # Why every fixture gets a React import prepended
//
// `rule_testing`'s tsconfig sets `types: []` and resolves no `node_modules`, deliberately, so
// `@types/react` can never resolve in a fixture. Most of these fixtures also call a bare `useRef`
// with no import at all, because the compiler harness seeds a global table naming it. This port
// asks the type checker instead, so an unimported `useRef` is an undeclared global and the file is
// `any`.
//
// The declarations below are the minimum of `@types/react` this rule reads, and one detail is
// deliberate rather than incidental: `RefObject` is written as an INTERFACE. This rule's checker
// predicate keys on the type's SYMBOL, which is populated for an interface and nil for an alias.
// A stub that made `RefObject` a type alias would test a checker behaviour that does not exist, and
// the rule would pass its fixtures while reporting nothing on real code. `Dispatch` is written as
// an alias for the same reason in reverse: it is the shape `set_state_in_render.go` keys on, and
// keeping both here means a change to either predicate has something to fail against.
const refsReactDeclarations = `declare module 'react' {
  export interface RefObject<T> { current: T }
  export interface MutableRefObject<T> { current: T }
  export type RefCallback<T> = (instance: T | null) => void;
  export type Ref<T> = RefCallback<T> | RefObject<T> | null;
  export type Dispatch<A> = (value: A) => void;
  export type SetStateAction<S> = S | ((prev: S) => S);
  export function useRef<T>(initial: T): RefObject<T>;
  export function useState<S>(initial?: S): [S, Dispatch<SetStateAction<S>>];
  export function useEffect(effect: () => void, deps?: unknown[]): void;
  export function useCallback<T>(callback: T, deps: unknown[]): T;
  export function useMemo<T>(compute: () => T, deps: unknown[]): T;
  export function useReducer<S, A>(reducer: (state: S, action: A) => S, initial: S): [S, (action: A) => void];
}
declare module 'shared-runtime' {
  export function addOne(value: number): number;
  export function useHook(...args: unknown[]): unknown;
  export function identity<T>(value: T): T;
}`

// refsImport is the one line prepended to each verbatim fixture so its ref has a type.
//
// Nineteen of the thirty fixtures call a bare `useRef` with no import, because the compiler harness
// seeds a global table naming it and never asks a checker. This port asks the checker, so without
// this line `useRef` is an undeclared global, the file is `any`, and the type signal cannot fire.
// The fixture BODY is untouched; the prepend lives here rather than in the constant so the copied
// text stays byte-comparable against upstream.
//
// This is the same trade `set_state_in_render.go` documents and states plainly: the rule depends on
// `@types/react` resolving, which is right for application code and wrong for a corpus written for
// a compiler that infers its own types. Note that the NAME signal needs no import at all, which is
// why the `props.ref` fixtures score with or without this line.
const refsImport = "import {useRef, useState, useCallback, useMemo, useReducer, useEffect} from 'react';\n"

// refsFixture is one imported case plus the ids its `.expect.md` asserts, in order.
type refsFixture struct {
	Name   string
	Source string
	Want   []string
}

// refsCorpus is the 30 non-Flow fixtures, verbatim.
var refsCorpus = []refsFixture{
	{
		Name:   "error.capture-ref-for-mutation",
		Source: "import {useRef} from 'react';\nimport {addOne} from 'shared-runtime';\n\nfunction useKeyCommand() {\n  const currentPosition = useRef(0);\n  const handleKey = direction => () => {\n    const position = currentPosition.current;\n    const nextPosition = direction === 'left' ? addOne(position) : position;\n    currentPosition.current = nextPosition;\n  };\n  const moveLeft = {\n    handler: handleKey('left')(),\n  };\n  const moveRight = {\n    handler: handleKey('right')(),\n  };\n  return [moveLeft, moveRight];\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: useKeyCommand,\n  params: [],\n};\n",
		Want:   []string{messageRefsFunctionAccessesRefId, messageRefsFunctionAccessesRefId},
	},
	{
		Name:   "error.fault-tolerance-reports-multiple-errors",
		Source: "// @validateRefAccessDuringRender\n/**\n * This fixture tests fault tolerance: the compiler should report\n * multiple independent errors rather than stopping at the first one.\n *\n * Error 1: Ref access during render (ref.current)\n * Error 2: Mutation of frozen value (props)\n */\nfunction Component(props) {\n  const ref = useRef(null);\n\n  // Error: reading ref during render\n  const value = ref.current;\n\n  // Error: mutating frozen value (props, which is frozen after hook call)\n  props.items = [];\n\n  return <div>{value}</div>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.hook-ref-value",
		Source: "import {useEffect, useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef();\n  useEffect(() => {}, [ref.current]);\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [],\n};\n",
		Want:   []string{messageRefsValueAccessId, messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-access-ref-during-render",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef(null);\n  const value = ref.current;\n  return value;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-access-ref-in-reducer-init",
		Source: "import {useReducer, useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef(props.value);\n  const [state] = useReducer(\n    (state, action) => state + action,\n    0,\n    init => ref.current\n  );\n\n  return <Stringify state={state} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{value: 42}],\n};\n",
		Want:   []string{messageRefsPassedToFunctionId},
	},
	{
		Name:   "error.invalid-access-ref-in-reducer",
		Source: "import {useReducer, useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef(props.value);\n  const [state] = useReducer(() => ref.current, null);\n\n  return <Stringify state={state} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{value: 42}],\n};\n",
		Want:   []string{messageRefsPassedToFunctionId},
	},
	{
		Name:   "error.invalid-access-ref-in-render-mutate-object-with-ref-function",
		Source: "import {useRef} from 'react';\n\nfunction Component() {\n  const ref = useRef(null);\n  const object = {};\n  object.foo = () => ref.current;\n  const refValue = object.foo();\n  return <div>{refValue}</div>;\n}\n",
		Want:   []string{messageRefsFunctionAccessesRefId},
	},
	{
		Name:   "error.invalid-access-ref-in-state-initializer",
		Source: "import {useRef, useState} from 'react';\n\nfunction Component(props) {\n  const ref = useRef(props.value);\n  const [state] = useState(() => ref.current);\n\n  return <Stringify state={state} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{value: 42}],\n};\n",
		Want:   []string{messageRefsPassedToFunctionId},
	},
	{
		Name:   "error.invalid-aliased-ref-in-callback-invoked-during-render-",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef(null);\n  const renderItem = item => {\n    const aliasedRef = ref;\n    const current = aliasedRef.current;\n    return <Foo item={item} current={current} />;\n  };\n  return <Items>{props.items.map(item => renderItem(item))}</Items>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-disallow-mutating-ref-in-render",
		Source: "// @validateRefAccessDuringRender\nfunction Component() {\n  const ref = useRef(null);\n  ref.current = false;\n\n  return <button ref={ref} />;\n}\n",
		Want:   []string{messageRefsUpdateId},
	},
	{
		Name:   "error.invalid-disallow-mutating-refs-in-render-transitive",
		Source: "// @validateRefAccessDuringRender\nfunction Component() {\n  const ref = useRef(null);\n\n  const setRef = () => {\n    ref.current = false;\n  };\n  const changeRef = setRef;\n  changeRef();\n\n  return <button ref={ref} />;\n}\n",
		Want:   []string{messageRefsFunctionAccessesRefId},
	},
	{
		Name:   "error.invalid-pass-ref-to-function",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef(null);\n  const x = foo(ref);\n  return x.current;\n}\n",
		Want:   []string{messageRefsPassedToFunctionId},
	},
	{
		Name:   "error.invalid-read-ref-prop-in-render-destructure",
		Source: "// @validateRefAccessDuringRender @compilationMode:\"infer\"\nfunction Component({ref}) {\n  const value = ref.current;\n  return <div>{value}</div>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-read-ref-prop-in-render-property-load",
		Source: "// @validateRefAccessDuringRender @compilationMode:\"infer\"\nfunction Component(props) {\n  const value = props.ref.current;\n  return <div>{value}</div>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-ref-in-callback-invoked-during-render",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef(null);\n  const renderItem = item => {\n    const current = ref.current;\n    return <Foo item={item} current={current} />;\n  };\n  return <Items>{props.items.map(item => renderItem(item))}</Items>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-ref-value-as-props",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef(null);\n  return <Foo ref={ref.current} />;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-set-and-read-ref-during-render",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef(null);\n  ref.current = props.value;\n  return ref.current;\n}\n",
		Want:   []string{messageRefsUpdateId, messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-set-and-read-ref-nested-property-during-render",
		Source: "// @validateRefAccessDuringRender\nfunction Component(props) {\n  const ref = useRef({inner: null});\n  ref.current.inner = props.value;\n  return ref.current.inner;\n}\n",
		Want:   []string{messageRefsUpdateId, messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-use-ref-added-to-dep-without-type-info",
		Source: "// @validateRefAccessDuringRender\nfunction Foo({a}) {\n  const ref = useRef();\n  // type information is lost here as we don't track types of fields\n  const val = {ref};\n  // without type info, we don't know that val.ref.current is a ref value so we\n  // *would* end up depending on val.ref.current\n  // however, this is an instance of accessing a ref during render and is disallowed\n  // under React's rules, so we reject this input\n  const x = {a, val: val.ref.current};\n\n  return <VideoList videos={x} />;\n}\n",
		Want:   []string{messageRefsValueAccessId, messageRefsValueAccessId},
	},
	{
		Name:   "error.invalid-write-but-dont-read-ref-in-render",
		Source: "// @validateRefAccessDuringRender\nfunction useHook({value}) {\n  const ref = useRef(null);\n  // Writing to a ref in render is against the rules:\n  ref.current = value;\n  // returning a ref is allowed, so this alone doesn't trigger an error:\n  return ref;\n}\n",
		Want:   []string{messageRefsUpdateId},
	},
	{
		Name:   "error.invalid-write-ref-prop-in-render",
		Source: "// @validateRefAccessDuringRender @compilationMode:\"infer\"\nfunction Component(props) {\n  const ref = props.ref;\n  ref.current = true;\n  return <div>{value}</div>;\n}\n",
		Want:   []string{messageRefsUpdateId},
	},
	{
		Name:   "error.ref-optional",
		Source: "import {useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef();\n  return ref?.current;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [],\n};\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.todo-useCallback-set-ref-nested-property-ref-modified-later-preserve-memoization",
		Source: "// @enablePreserveExistingMemoizationGuarantees @validateRefAccessDuringRender\nimport {useCallback, useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef({inner: null});\n\n  const onChange = useCallback(event => {\n    // The ref should still be mutable here even though function deps are frozen in\n    // @enablePreserveExistingMemoizationGuarantees mode\n    ref.current.inner = event.target.value;\n  });\n\n  // The ref is modified later, extending its range and preventing memoization of onChange\n  ref.current.inner = null;\n\n  return <input onChange={onChange} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{}],\n};\n",
		Want:   []string{messageRefsUpdateId},
	},
	{
		Name:   "error.useCallback-accesses-ref-mutated-later-via-function-preserve-memoization",
		Source: "// @enablePreserveExistingMemoizationGuarantees @validateRefAccessDuringRender\nimport {useCallback, useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef({inner: null});\n\n  const onChange = useCallback(event => {\n    // The ref should still be mutable here even though function deps are frozen in\n    // @enablePreserveExistingMemoizationGuarantees mode\n    ref.current.inner = event.target.value;\n  });\n\n  // The ref is modified later, extending its range and preventing memoization of onChange\n  const reset = () => {\n    ref.current.inner = null;\n  };\n  reset();\n\n  return <input onChange={onChange} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{}],\n};\n",
		Want:   []string{messageRefsFunctionAccessesRefId},
	},
	{
		Name:   "error.useCallback-set-ref-nested-property-dont-preserve-memoization",
		Source: "// @enablePreserveExistingMemoizationGuarantees:false\nimport {useCallback, useRef} from 'react';\n\nfunction Component(props) {\n  const ref = useRef({inner: null});\n\n  const onChange = useCallback(event => {\n    // The ref should still be mutable here even though function deps are frozen in\n    // @enablePreserveExistingMemoizationGuarantees mode\n    ref.current.inner = event.target.value;\n  });\n\n  ref.current.inner = null;\n\n  return <input onChange={onChange} />;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [{}],\n};\n",
		Want:   []string{messageRefsUpdateId},
	},
	{
		Name:   "error.validate-mutate-ref-arg-in-render",
		Source: "// @validateRefAccessDuringRender:true\nimport {mutate} from 'shared-runtime';\n\nfunction Foo(props, ref) {\n  mutate(ref.current);\n  return <div>{props.bar}</div>;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Foo,\n  params: [{bar: 'foo'}, {ref: {cuurrent: 1}}],\n  isComponent: true,\n};\n",
		Want:   []string{messageRefsPassedToFunctionId},
	},
	{
		Name:   "error.try-finally-and-ref-access",
		Source: "// @validateRefAccessDuringRender\n/**\n * Fault tolerance test: two independent errors should both be reported.\n *\n * Error 1 (BuildHIR): `try/finally` is not supported\n * Error 2 (ValidateNoRefAccessInRender): reading ref.current during render\n */\nfunction Component() {\n  const ref = useRef(null);\n\n  // Error: try/finally (Todo from BuildHIR)\n  try {\n    doSomething();\n  } finally {\n    cleanup();\n  }\n\n  // Error: reading ref during render\n  const value = ref.current;\n\n  return <div>{value}</div>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.try-finally-ref-access-and-mutation",
		Source: "// @validateRefAccessDuringRender\n/**\n * Fault tolerance test: three independent errors should all be reported.\n *\n * Error 1 (BuildHIR): `try/finally` is not supported\n * Error 2 (ValidateNoRefAccessInRender): reading ref.current during render\n * Error 3 (InferMutationAliasingEffects): Mutation of frozen props\n */\nfunction Component(props) {\n  const ref = useRef(null);\n\n  // Error: try/finally (Todo from BuildHIR)\n  try {\n    doWork();\n  } finally {\n    cleanup();\n  }\n\n  // Error: reading ref during render\n  const value = ref.current;\n\n  // Error: mutating frozen props\n  props.items = [];\n\n  return <div>{value}</div>;\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.var-declaration-and-ref-access",
		Source: "// @validateRefAccessDuringRender\n/**\n * Fault tolerance test: two independent errors should both be reported.\n *\n * Error 1 (BuildHIR): `var` declarations are not supported (treated as `let`)\n * Error 2 (ValidateNoRefAccessInRender): reading ref.current during render\n */\nfunction Component() {\n  const ref = useRef(null);\n\n  // Error: var declaration (Todo from BuildHIR)\n  var items = [1, 2, 3];\n\n  // Error: reading ref during render\n  const value = ref.current;\n\n  return (\n    <div>\n      {value}\n      {items.length}\n    </div>\n  );\n}\n",
		Want:   []string{messageRefsValueAccessId},
	},
	{
		Name:   "error.maybe-mutable-ref-not-preserved",
		Source: "// @validatePreserveExistingMemoizationGuarantees:true\n\nimport {useRef, useMemo} from 'react';\nimport {makeArray} from 'shared-runtime';\n\nfunction useFoo() {\n  const r = useRef();\n  return useMemo(() => makeArray(r), []);\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: useFoo,\n  params: [],\n};\n",
		Want:   []string{messageRefsPassedToFunctionId},
	},
}

// runRefsFixture builds the two-file program a typed rule needs.
//
// The fixture keeps its own extension so that a `.tsx` case parses as JSX. `RunTypedFiles` writes
// each file as `strings.TrimSpace(contents)+"\n"`, so the bytes on disk are one newline short of
// the literal where a fixture began with one; every span assertion below slices the source the
// harness actually wrote rather than the Go literal, which is the documented way this bites.
func runRefsFixture(t *testing.T, fixture refsFixture) rule_testing.Result {
	t.Helper()
	name := "component.tsx"
	return rule_testing.RunTypedFiles(t, Refs, map[string]string{
		name:         refsImport + fixture.Source,
		"react.d.ts": refsReactDeclarations,
	}, name)
}

// TestRefsMatchesTheImportedCorpus asserts every imported fixture, id by id and in order.
//
// This is the whole score and it is an EXACT assertion rather than a count: 30 of 30 fixtures, with
// each fixture's message ids compared as an ordered sequence. An exact assertion is what makes the
// number honest, because a fix and a regression landing together cannot net to zero the way they
// can under a threshold.
func TestRefsMatchesTheImportedCorpus(t *testing.T) {
	t.Parallel()

	if len(refsCorpus) != 30 {
		t.Fatalf("the corpus holds %d fixtures, want the 30 non-Flow refs fixtures", len(refsCorpus))
	}
	for _, fixture := range refsCorpus {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			result := runRefsFixture(t, fixture)
			// Routed through the harness rather than compared by hand: `ExpectFindings` asserts the
			// ids AND the count in the order the rule produced them, which is the same assertion
			// this made itself, and `TestEveryRuleShipsAFixturePair` in the registry reads for this
			// call textually to prove the rule can be shown to fire at all.
			rule_testing.ExpectFindings(t, result, fixture.Want...)
		})
	}
}

// TestRefsReportsExactlyThisManyFindingsOnTheCorpus pins the TOTAL, so a fix and a regression
// landing in the same change cannot cancel out.
//
// The brief calls an exact-count assertion the instrument that makes a subset honest. This rule is
// not a subset, and the count is asserted anyway for the same reason: 35 findings over 30 fixtures,
// which is the number of caret runs in the vendored `.expect.md` files.
func TestRefsReportsExactlyThisManyFindingsOnTheCorpus(t *testing.T) {
	t.Parallel()

	total := 0
	for _, fixture := range refsCorpus {
		total += len(runRefsFixture(t, fixture).Diagnostics)
	}
	if total != 35 {
		t.Errorf("the corpus produced %d findings, want exactly 35", total)
	}
}

// refsCleanCases are inputs that must NOT report.
//
// Upstream's `refs` corpus is all error fixtures, so there is no imported clean set for this rule
// and every case here was written from reading the validator and then CHECKED against the running
// `eslint-plugin-react-hooks` 7.1.1 rather than against my reading. The brief's warning about
// invented fixtures applies with full force here: a clean case I invented encodes the same belief
// as the rule, so each one below names the distinction it exists to protect and each was measured.
var refsCleanCases = []struct {
	Name   string
	Source string
	Why    string
}{
	{
		Name:   "read in an effect",
		Source: "function Component() {\n  const ref = useRef(null);\n  useEffect(() => {\n    ref.current;\n  }, []);\n  return <div ref={ref} />;\n}\n",
		Why:    "the whole point of the rule: outside render is where a ref may be read",
	},
	{
		Name:   "read in an event handler",
		Source: "function Component() {\n  const ref = useRef(null);\n  return <div onClick={() => ref.current} />;\n}\n",
		Why:    "a handler runs after commit, so the ref is attached",
	},
	{
		Name:   "guarded initialization",
		Source: "function Component() {\n  const ref = useRef(null);\n  if (ref.current == null) {\n    ref.current = 1;\n  }\n  return <div ref={ref} />;\n}\n",
		Why:    "the sanctioned lazy-init idiom, and the reason the guard element exists at all",
	},
	{
		Name:   "the ref object passed as a JSX ref prop",
		Source: "function Component() {\n  const ref = useRef(null);\n  return <div ref={ref} />;\n}\n",
		Why:    "handing the ref to React is what a ref is for; only reading current is a finding",
	},
	{
		Name:   "returning the ref from a hook",
		Source: "function useThing() {\n  const ref = useRef(null);\n  return ref;\n}\n",
		Why:    "returning a ref is explicitly allowed; only a ref VALUE may not escape",
	},
	{
		Name:   "a lowercase function is not a component",
		Source: "function helper(props) {\n  return props.ref.current;\n}\n",
		Why:    "the compilation gate: upstream never analyses a non-component, measured clean",
	},
	{
		Name:   "a bare Ref name does not match",
		Source: "function Component(props) {\n  const Ref = props.thing;\n  return <div>{Ref.current}</div>;\n}\n",
		Why:    "upstream's regex requires a character before Ref; `Ref` alone is clean, measured",
	},
	{
		Name:   "an unrelated name with a current property",
		Source: "function Component(props) {\n  const thing = props.thing;\n  return <div>{thing.current}</div>;\n}\n",
		Why:    "the negative control for the name signal: no ref type, no ref-like name",
	},
	{
		Name:   "a property load off props does not inherit a ref-like name",
		Source: "function Component(props) {\n  return <div>{props.fooRef.current}</div>;\n}\n",
		Why:    "measured clean upstream: the binding is props, and only LoadLocal carries a name",
	},
	{
		Name:   "a renamed alias of a ref-like prop",
		Source: "function Component(props) {\n  const q = props.fooRef;\n  return <div>{q.current}</div>;\n}\n",
		Why:    "measured clean upstream: the binding is q, which is not a ref-like name",
	},
}

// TestRefsStaysSilent asserts the clean cases.
func TestRefsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range refsCleanCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			result := runRefsFixture(t, refsFixture{Name: testCase.Name, Source: testCase.Source})
			if len(result.Diagnostics) != 0 {
				t.Logf("this case exists because: %s", testCase.Why)
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestRefsPointsAtTheAccess asserts the SPAN, which no message-id assertion can see.
//
// The source is sliced with the finding's own range and compared against a literal typed here, not
// against anything the rule computed. Note the harness writes each fixture as
// `strings.TrimSpace(contents)+"\n"`, so the text the rule saw is the trimmed text; the slice below
// is taken from that same string rather than from the Go literal, which is the documented way this
// assertion goes wrong by one byte.
func TestRefsPointsAtTheAccess(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		Name   string
		Source string
		Want   string
	}{
		{
			Name:   "a read points at the whole member access",
			Source: "function Component() {\n  const ref = useRef(null);\n  const value = ref.current;\n  return value;\n}\n",
			Want:   "ref.current",
		},
		{
			Name:   "a write points at the target of the assignment",
			Source: "function Component() {\n  const ref = useRef(null);\n  ref.current = 1;\n  return null;\n}\n",
			Want:   "ref.current",
		},
		{
			Name:   "a ref passed to a function points at the argument",
			Source: "function Component() {\n  const ref = useRef(null);\n  const x = foo(ref);\n  return x;\n}\n",
			Want:   "ref",
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			source := refsImport + testCase.Source
			result := runRefsFixture(t, refsFixture{Name: testCase.Name, Source: testCase.Source})
			if len(result.Diagnostics) == 0 {
				t.Fatalf("reported nothing, want a finding spanning %q", testCase.Want)
			}
			written := strings.TrimSpace(source) + "\n"
			span := result.Diagnostics[0].Range
			if span.Pos() < 0 || span.End() > len(written) || span.Pos() >= span.End() {
				t.Fatalf("finding has range [%d,%d) outside the %d byte source", span.Pos(), span.End(), len(written))
			}
			if got := written[span.Pos():span.End()]; got != testCase.Want {
				t.Errorf("finding spans %q, want %q", got, testCase.Want)
			}
		})
	}
}

// TestRefsMessagesReadCorrectly asserts the rendered text exactly.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer for this rule, so these are
// equality assertions on constants rather than on a format string's output. They exist because a
// message id assertion cannot see the text, and a Description that drifted into restating the rule
// name would pass every other test in this file.
func TestRefsMessagesReadCorrectly(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		Kind         refsFindingKind
		WantId       string
		WantPrefix   string
		WantMentions string
	}{
		{refsFindingValueAccess, "refValueAccess", "Cannot access ref value during render.", "event handler"},
		{refsFindingUpdate, "refUpdate", "Cannot update ref during render.", "event handler"},
		{refsFindingPassedToFunction, "refPassedToFunction", "Passing a ref to a function may read its value during render.", "event handler"},
		{refsFindingFunctionAccessesRef, "functionAccessesRef", "This function accesses a ref value, and calling it during render reads the ref.", "event handler"},
	} {
		message := refsMessageFor(testCase.Kind)
		if message.Id != testCase.WantId {
			t.Errorf("kind %d has id %q, want %q", testCase.Kind, message.Id, testCase.WantId)
		}
		if !strings.HasPrefix(message.Description, testCase.WantPrefix) {
			t.Errorf("kind %d description starts %q, want prefix %q", testCase.Kind, message.Description, testCase.WantPrefix)
		}
		if !strings.Contains(message.Description, testCase.WantMentions) {
			t.Errorf("kind %d description never says where a ref MAY be read", testCase.Kind)
		}
	}
}

// TestRefsConvergenceEqualityIgnoresRefIdentity pins the hand-written equality.
//
// This is the test that protects the single subtlety upstream bothers to document, and it is worth
// stating why an ordinary equality test would not: the hazard is not that the comparison is wrong
// about identity, it is that a comparison which is RIGHT about identity makes the fixpoint diverge.
// `refsJoin` mints a fresh ref id on every merge of two different refs, so an equality that looked
// at the id would report the environment as changed on every round forever.
//
// So `ref` must compare equal across different ids, `refValue` must compare equal when only its
// origin differs, and `guard` must NOT, because a guard is sound only for the ref it guards.
func TestRefsConvergenceEqualityIgnoresRefIdentity(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		Name  string
		A     *refsAccessType
		B     *refsAccessType
		Equal bool
		Why   string
	}{
		{
			Name:  "two refs with different identities are equal",
			A:     &refsAccessType{Kind: refsRef, RefId: 1, HasRefId: true},
			B:     &refsAccessType{Kind: refsRef, RefId: 999, HasRefId: true},
			Equal: true,
			Why:   "a join mints a fresh id, so comparing ids would prevent the fixpoint settling",
		},
		{
			Name:  "two ref values differing only in origin are equal",
			A:     &refsAccessType{Kind: refsRefValue, Span: 4, HasSpan: true, RefId: 1, HasRefId: true},
			B:     &refsAccessType{Kind: refsRefValue, Span: 4, HasSpan: true, RefId: 7, HasRefId: true},
			Equal: true,
			Why:   "refValue compares its access span and deliberately ignores ref origin",
		},
		{
			Name:  "two ref values with different access spans are not equal",
			A:     &refsAccessType{Kind: refsRefValue, Span: 4, HasSpan: true},
			B:     &refsAccessType{Kind: refsRefValue, Span: 5, HasSpan: true},
			Equal: false,
			Why:   "the access span is the one field refValue does compare",
		},
		{
			Name:  "two guards for different refs are not equal",
			A:     &refsAccessType{Kind: refsGuard, RefId: 1, HasRefId: true},
			B:     &refsAccessType{Kind: refsGuard, RefId: 2, HasRefId: true},
			Equal: false,
			Why:   "a guard authorizes a write to ONE ref; conflating two would authorize the wrong write",
		},
		{
			Name:  "different kinds are never equal",
			A:     &refsAccessType{Kind: refsRef},
			B:     &refsAccessType{Kind: refsRefValue},
			Equal: false,
			Why:   "the kinds are the lattice; a ref and a ref value are different findings",
		},
	} {
		if got := refsTypeEqual(testCase.A, testCase.B); got != testCase.Equal {
			t.Errorf("%s: equality is %v, want %v (%s)", testCase.Name, got, testCase.Equal, testCase.Why)
		}
	}
}

// TestRefsJoinMintsAFreshIdentityForDifferentRefs is the other half of the convergence story.
//
// It asserts the behaviour that MAKES the equality above necessary, so the two tests fail
// separately: if the join stopped minting, this fails while the equality test still passes, and a
// reader would otherwise be left wondering why the equality is written so strangely.
func TestRefsJoinMintsAFreshIdentityForDifferentRefs(t *testing.T) {
	t.Parallel()

	env := newRefsEnvironment()
	// The operands take identities the environment's counter has already issued, so a freshly
	// minted id cannot coincide with one of them by accident. Without this the test can pass or
	// fail on the counter's starting value rather than on the join's behaviour.
	first := &refsAccessType{Kind: refsRef, RefId: env.nextRefId(), HasRefId: true}
	second := &refsAccessType{Kind: refsRef, RefId: env.nextRefId(), HasRefId: true}

	joined := refsJoin(first, second, env.nextRefId)
	if joined.Kind != refsRef {
		t.Fatalf("joining two refs gave kind %d, want a ref", joined.Kind)
	}
	if joined.RefId == first.RefId || joined.RefId == second.RefId {
		t.Errorf("the join reused identity %d rather than minting a fresh one", joined.RefId)
	}
	if !refsTypeEqual(joined, first) {
		t.Error("the freshly minted ref does not compare equal to its operand, so the fixpoint cannot settle")
	}
}

// TestRefsSettlesOnEveryCorpusFixture asserts the sweep CONVERGES.
//
// The non-convergence finding is upstream's own invariant diagnostic and it exists so a divergent
// environment is loud rather than silent. Its presence anywhere in the corpus would mean the
// equality above is wrong, so this asserts it never appears; the bound is ten rounds and every
// fixture here settles well inside it.
func TestRefsSettlesOnEveryCorpusFixture(t *testing.T) {
	t.Parallel()

	for _, fixture := range refsCorpus {
		for _, diagnostic := range runRefsFixture(t, fixture).Diagnostics {
			if diagnostic.Message.Id == messageRefsDidNotConvergeId {
				t.Errorf("%s did not converge within the ten round bound", fixture.Name)
			}
		}
	}
}

// TestRefsRequiresTheTypedHarness guards the checker declaration.
//
// A typed rule handed a nil checker goes SILENT rather than crashing, because `GetSymbolAtLocation`
// on a nil checker returns nil rather than panicking. That is the dangerous failure mode: every
// silence test would pass vacuously and every fires test would look like a rule bug. This asserts
// the rule declares the checker, so a later revert of that field fails here loudly.
func TestRefsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !Refs.NeedsTypeChecker {
		t.Error("the rule stopped declaring NeedsTypeChecker; every typed fixture would pass vacuously")
	}
}

// TestRefsNameSignalWorksWithoutTheChecker is the other half of the two-signal claim.
//
// The rule comment asserts that neither signal alone covers the corpus. This pins the direction that
// is easy to lose: with NO React import at all, so `useRef` is an undeclared global and the file is
// `any`, a ref arriving by NAME must still report. If someone deleted the name signal believing the
// checker was sufficient, the imported corpus would still score 27 of 30 and only this would fail.
func TestRefsNameSignalWorksWithoutTheChecker(t *testing.T) {
	t.Parallel()

	source := "function Component(props) {\n  const value = props.ref.current;\n  return <div>{value}</div>;\n}\n"
	result := rule_testing.RunTypedFiles(t, Refs, map[string]string{"c.tsx": source}, "c.tsx")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("reported %d findings with no React types available, want 1 from the name signal", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Id != messageRefsValueAccessId {
		t.Errorf("reported %s, want %s", result.Diagnostics[0].Message.Id, messageRefsValueAccessId)
	}
}

// TestRefsJoinsTaintAtABranchMerge covers the phi join, which no imported fixture reaches.
//
// A mutation deleting the phi loop SURVIVED the entire 30 fixture corpus, and reading the corpus
// says why rather than guessing: not one non-Flow refs fixture merges a ref across two branches.
// The branch cases upstream ships for this rule are all Flow (`component C()` syntax), which our
// parser cannot accept, so the whole discrimination is invisible to the imported set.
//
// Both inputs below were measured against the running `eslint-plugin-react-hooks` 7.1.1 BEFORE
// being written here, because the brief's rule is that a fixture for a survivor is a hypothesis
// until upstream has confirmed the verdict. Both report exactly one finding there.
//
// The second case is the one that carries the real weight. A ref arriving on only ONE path still
// reports, which is what makes this a JOIN over the operands rather than a test that every operand
// is a ref: an implementation requiring agreement across predecessors would pass the first case and
// fail the second.
func TestRefsJoinsTaintAtABranchMerge(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		Name   string
		Source string
	}{
		{
			Name:   "a ref merged from both branches",
			Source: "function Component(props) {\n  const a = useRef(null);\n  const b = useRef(null);\n  let chosen;\n  if (props.cond) {\n    chosen = a;\n  } else {\n    chosen = b;\n  }\n  const value = chosen.current;\n  return <div>{value}</div>;\n}\n",
		},
		{
			Name:   "a ref arriving on only one branch",
			Source: "function Component(props) {\n  const a = useRef(null);\n  let chosen = props.other;\n  if (props.cond) {\n    chosen = a;\n  }\n  const value = chosen.current;\n  return <div>{value}</div>;\n}\n",
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			result := runRefsFixture(t, refsFixture{Name: testCase.Name, Source: testCase.Source})
			if len(result.Diagnostics) != 1 {
				ids := []string{}
				for _, diagnostic := range result.Diagnostics {
					ids = append(ids, diagnostic.Message.Id)
				}
				t.Fatalf("reported %v, want exactly one finding (measured on upstream)", ids)
			}
			if result.Diagnostics[0].Message.Id != messageRefsValueAccessId {
				t.Errorf("reported %s, want %s", result.Diagnostics[0].Message.Id, messageRefsValueAccessId)
			}
		})
	}
}

// TestRefsExemptsTheMergeRefsShape covers the `RefCallback` result exemption.
//
// A mutant breaking the `bivarianceHack` match SURVIVED the whole corpus, correctly: no upstream
// fixture merges refs. The case is real on production code, where it was 8 findings before the
// exemption existed, one in every form-field component that forwards a ref.
//
// The declaration below is copied from the shape `@types/react` actually uses rather than the
// obvious one. `RefCallback<T>` is an indexed access over `{ bivarianceHack(...): void }["bivarianceHack"]`
// (`@types/react/index.d.ts:176-185`), and resolving it ERASES the type alias, leaving the method's
// own symbol. A stub written as `type RefCallback<T> = (i: T) => void` keeps the alias, passes, and
// tests a checker behaviour that does not occur on real code — the exact failure where the rule and
// its fixtures share one wrong belief. This fixture is written the real way for that reason.
func TestRefsExemptsTheMergeRefsShape(t *testing.T) {
	t.Parallel()

	declarations := "declare module 'react' {\n" +
		"  export interface RefObject<T> { current: T }\n" +
		"  export type RefCallback<T> = { bivarianceHack(instance: T | null): void }['bivarianceHack'];\n" +
		"  export type Ref<T> = RefCallback<T> | RefObject<T> | null;\n" +
		"  export function useRef<T>(initial: T): RefObject<T>;\n" +
		"}"
	// The forwarded ref is DESTRUCTURED rather than read as `properties.ref`, and that difference is
	// upstream's rather than a convenience. Measured on the executable:
	// `mergeReferences([inner, properties.ref])` reports TWICE while
	// `function Component({ref}) { mergeReferences([inner, ref]) }` is clean. The member form is a
	// ref-value read in upstream's model regardless of what consumes it, and this port agrees on
	// both spellings. My first version of this fixture asserted silence for the member form and was
	// simply wrong about upstream; the probe corrected it rather than the rule.
	source := "import * as React from 'react';\n" +
		"declare function mergeReferences<T>(references: unknown[]): React.RefCallback<T>;\n" +
		"function Component({ref}: {ref: React.Ref<HTMLInputElement>}) {\n" +
		"  const inputReference = React.useRef<HTMLInputElement | null>(null);\n" +
		"  const setInputReference = mergeReferences<HTMLInputElement>([inputReference, ref]);\n" +
		"  return <input ref={setInputReference} />;\n" +
		"}\n"

	result := rule_testing.RunTypedFiles(t, Refs, map[string]string{"c.tsx": source, "react.d.ts": declarations}, "c.tsx")
	if len(result.Diagnostics) != 0 {
		ids := []string{}
		for _, diagnostic := range result.Diagnostics {
			ids = append(ids, diagnostic.Message.Id)
		}
		t.Errorf("reported %v on the mergeRefs shape, want silence (measured clean upstream)", ids)
	}
}

// TestRefsDoesNotTaintSiblingFieldsOfARefProp covers the aliasing restriction.
//
// A property load aliases its result to its OBJECT, which is how an alias chain keeps its identity.
// Applied to a props bag holding a ref beside ordinary fields it is a disaster: reading the ref
// field marks the whole object as a ref, and every OTHER field read off it then reports as a ref
// access. This was 6 findings in one file on Kirk's tree, all on a `currentWidth: number` and a
// `columnId: string`, and no upstream fixture can see it because upstream types each property
// independently rather than aliasing through the object.
//
// The assertion is that reading the ref field does not make the number field a finding. It also
// pins that the ref field itself is still tracked, so a fix that simply stopped seeing refs on
// props would fail the second half rather than passing this by going blind.
func TestRefsDoesNotTaintSiblingFieldsOfARefProp(t *testing.T) {
	t.Parallel()

	declarations := "declare module 'react' {\n  export interface RefObject<T> { current: T }\n  export function useRef<T>(initial: T): RefObject<T>;\n}"
	source := "import * as React from 'react';\n" +
		"interface Properties {\n  currentWidth: number;\n  guideElementReference: React.RefObject<HTMLElement | null>;\n}\n" +
		"function Component(properties: Properties) {\n" +
		"  const guide = properties.guideElementReference;\n" +
		"  const width = properties.currentWidth;\n" +
		"  return <div style={{width}} data-guide={guide} />;\n" +
		"}\n"

	result := rule_testing.RunTypedFiles(t, Refs, map[string]string{"c.tsx": source, "react.d.ts": declarations}, "c.tsx")
	if len(result.Diagnostics) != 0 {
		ids := []string{}
		for _, diagnostic := range result.Diagnostics {
			ids = append(ids, diagnostic.Message.Id)
		}
		t.Fatalf("reported %v, want silence: neither the ref OBJECT nor a sibling field is a ref VALUE read", ids)
	}

	// The other half: the ref field is still a ref, so reading its `current` here does report.
	// Without this a change that stopped recognizing refs on props would pass the assertion above.
	reading := "import * as React from 'react';\n" +
		"interface Properties {\n  currentWidth: number;\n  guideElementReference: React.RefObject<HTMLElement | null>;\n}\n" +
		"function Component(properties: Properties) {\n" +
		"  const value = properties.guideElementReference.current;\n" +
		"  return <div>{value}</div>;\n" +
		"}\n"
	readingResult := rule_testing.RunTypedFiles(t, Refs, map[string]string{"c.tsx": reading, "react.d.ts": declarations}, "c.tsx")
	if len(readingResult.Diagnostics) != 1 {
		t.Errorf("reading the ref field's current gave %d findings, want 1 — the rule has gone blind to refs on props", len(readingResult.Diagnostics))
	}
}

// A hook spelled through the React namespace admits its function to compilation.
//
// The gate asks whether a function creates JSX or calls a hook, and it read only `LoadGlobal`, so a
// bare `useRef` admitted and `React.useRef` did not. A file that imports React as a default and
// spells every hook that way was never compiled at all, and this rule reported nothing in it.
//
// Found on a real file rather than by a fixture. `useKingdomLive.tsx` reads a ref during render,
// ESLint reported it, and this rule was silent: every hook call in that file is namespaced. The
// sibling gate in `unsupported_syntax.go` already accepted both spellings, which is what made this
// a gap rather than a decision.
//
// A namespaced call lowers to a `MethodCall` whose property is a `Primitive` string rather than to a
// `PropertyLoad`, which is why matching on the property load did not work either and is worth
// pinning: the shape is not the one a reader predicts.
func TestRefsAdmitsNamespacedHookCalls(t *testing.T) {
	t.Parallel()

	namespaced := runRefsFixture(t, refsFixture{
		Source: "import * as ReactNamespace from 'react';\n" +
			"export function useOther() {\n" +
			"  const reference = ReactNamespace.useRef<number | null>(null);\n" +
			"  if(reference.current === null) { reference.current = 1; }\n" +
			"  return reference.current;\n" +
			"}\n",
	})
	rule_testing.ExpectFindings(t, namespaced, "refValueAccess")

	// The control that gives it meaning: a namespaced call to something that is not a hook must not
	// admit the function, or the gate stops gating.
	notAHook := runRefsFixture(t, refsFixture{
		Source: "import * as Other from 'react';\n" +
			"export function helper() {\n" +
			"  const value = Other.somethingElse<number | null>(null);\n" +
			"  return value;\n" +
			"}\n",
	})
	rule_testing.ExpectClean(t, notAHook)
}

// TestRefsCannotHaveRefTypeAgreesWithTheChecker is the proof behind refsCannotHaveRefType.
//
// That function answers "no ref" without asking the checker, so it is safe exactly when the checker
// would also say no. This walks every node of a source written to tempt it (refs inside object and
// array literals, behind `??` and `||`, returned from an arrow, passed through JSX, interpolated in a
// template) and asks the checker about each node the exclusion declines. The tempting shapes the
// exclusion keeps asking about must answer yes, which is what shows the checker half can see a ref
// here at all.
func TestRefsCannotHaveRefTypeAgreesWithTheChecker(t *testing.T) {
	t.Parallel()

	const source = `import {useRef, type RefObject} from "./react";

declare function render(value: unknown): void;
declare const maybe: RefObject<number> | undefined;

export function Component({label}: {label: string}) {
  const ref = useRef(0);
  const bag = {ref, nested: [ref, maybe]};
  const either = maybe ?? ref;
  const fallback = maybe || ref;
  const pick = (other: RefObject<number>) => other;
  const named = function namedFunction() { return ref; };
  const callback = function RefCallback() { return ref; };
  const chosen = label ? ref : pick(ref);
  const text = ` + "`${label}${ref.current}`" + `;
  const sum = ref.current + 1;
  const same = ref === maybe;
  const negated = !ref;
  render([bag, either, fallback, pick, named, callback, chosen, text, sum, same, negated, typeof ref, void ref]);
  return <div ref={undefined} title={label}>{label}<span>{text}</span></div>;
}
`

	declined := map[ast.Kind]int{}
	asked := map[string]bool{}
	probe := rule.Rule{
		Name:             "refs-cannot-have-ref-type-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(root *ast.Node) {
					var visit func(node *ast.Node) bool
					visit = func(node *ast.Node) bool {
						if refsCannotHaveRefType(node) {
							declined[node.Kind]++
							if refsNodeHasRefType(ctx, node) {
								t.Errorf("declined %s at %d, which the checker types as a ref", node.Kind, node.Pos())
							}
						}
						if node.Kind == ast.KindBinaryExpression || node.Kind == ast.KindConditionalExpression ||
							node.Kind == ast.KindFunctionExpression {
							text := ctx.SourceFile.Text()[node.Pos():node.End()]
							asked[strings.TrimSpace(text)] = !refsCannotHaveRefType(node) && refsNodeHasRefType(ctx, node)
						}
						return node.ForEachChild(visit)
					}
					visit(root)
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"react.d.ts": reactStub, "a.tsx": source}, "a.tsx")

	for _, kind := range []ast.Kind{ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindTemplateExpression, ast.KindJsxElement, ast.KindBinaryExpression,
		ast.KindPrefixUnaryExpression, ast.KindTypeOfExpression, ast.KindVoidExpression} {
		if declined[kind] == 0 {
			t.Errorf("no %s was declined, so this source does not test it", kind)
		}
	}
	for _, shape := range []string{"maybe ?? ref", "maybe || ref", "label ? ref : pick(ref)",
		"function RefCallback() { return ref; }"} {
		if !asked[shape] {
			t.Errorf("%q must still be asked and typed as a ref; it was not", shape)
		}
	}
}

// TestRefsReportsARefReadInsideUseMemoAtTheRead is TanStack Query's HydrationBoundary, reduced
// (#zx5xvtg item 8). A memo callback runs during render, so a ref read inside it is read during
// render, and React reports it at the read: upstream erases manual memoization before any validator
// runs, so the callback's body is render code. HydrationBoundary.tsx writes
// `// eslint-disable-next-line react-hooks/refs` above that read, and React honours it.
//
// cohere reported the same verdict at the `useMemo` call instead, one screen above, where the
// author's disable comment does not reach. The rule now lowers through the memo-erased graph, as
// upstream's validator reads it, so the finding is on the read and the suppression covers it.
func TestRefsReportsARefReadInsideUseMemoAtTheRead(t *testing.T) {
	t.Parallel()

	result := runRefsFixture(t, refsFixture{
		Name: "ref read inside a memo callback",
		Source: `function hydrate(options: unknown): void {}
function HydrationBoundary({state}: {state: number}) {
  const optionsRef = useRef(state);
  const queue = useMemo(() => {
    hydrate(optionsRef.current);
    return state;
  }, [state]);
  return <div>{queue}</div>;
}
`,
	})
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %v", result.MessageIds())
	}
	source := result.SourceFile.Text()
	reported := strings.Count(source[:result.Diagnostics[0].Range.Pos()], "\n")
	read := strings.Count(source[:strings.Index(source, "optionsRef.current")], "\n")
	if reported != read {
		t.Errorf("the finding is on line %d, the ref read on line %d: %q", reported+1, read+1,
			source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()])
	}
}
