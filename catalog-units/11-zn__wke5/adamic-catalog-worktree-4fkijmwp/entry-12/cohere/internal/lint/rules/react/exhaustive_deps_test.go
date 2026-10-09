package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// exhaustiveDepsFile is where the fixtures pretend to live.
//
// A real path with a real extension matters more here than for a syntactic rule: this rule reads
// the checker, so the harness builds an actual program, and the corpus mixes TypeScript and JSX
// into the same cases.
const exhaustiveDepsFile = "/repository/source/ExhaustiveDeps.tsx"

// The corpus is oxc's 325 cases, with every verdict re-derived from REACT rather than from oxc.
//
// oxc's own pass/fail split is deliberately NOT the expectation. React is this rule's authority
// and the two disagree on seven inputs, so using the split would have pinned the reimplementation
// wherever it is wrong. Each case below was run through eslint-plugin-react-hooks 7.1.1 via the
// ESLint Linter API with `@typescript-eslint/parser`, and each message it produced was mapped to
// this rule's message id. Every string is byte-identical to oxc's file; only the verdicts come
// from elsewhere.
//
// One note about the parser, because it produced a confident wrong answer first. Run with the
// default JavaScript parser, this same comparison reported ELEVEN disagreements. Seven were the
// probe failing to parse TypeScript, which surfaces as a parse error rather than as a finding, so
// those cases read as React declining inputs it actually reports on. A probe that cannot read its
// input agrees with everything.
//
// Findings are expected in SOURCE POSITION order, which is this rule's emission order: entries in
// the dependency array are reported as the array is read, and the finding about the array as a
// whole comes last. ESLint reports the summary first, so an expectation copied from its output
// order would be wrong about eleven cases while being right about every id.
func TestExhaustiveDepsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
	}{
		{"function Foo2() { useEffect(() => { foo() }, []); const foo = () => { bar() }; function bar () { foo() } }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const value = props.flag ? {} : {};\n          useCallback(() => {\n            return value;\n          }, [value]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          const value = props.cached || {};\n          useCallback(() => {\n            return value;\n          }, [value]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.toString());\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.bar.baz);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.bar?.baz);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.bar.toString());\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = someFunc();\n          useEffect(() => {\n            console.log(local);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Counter(unstableProp) {\n          let [count, setCount] = useState(0);\n          setCount = unstableProp\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(c => c + 1);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          let local = 42;\n          useEffect(() => {\n            console.log(local);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = /foo/;\n          useEffect(() => {\n            console.log(local);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const value = useMemo(() => { return 2*2; });\n          const fn = useCallback(() => { alert('foo'); });\n        }", []string{"exhaustiveDepsUselessWithoutDependencies", "exhaustiveDepsUselessWithoutDependencies"}},
		{"function MyComponent({ fn1, fn2 }) {\n          const value = useMemo(fn1);\n          const fn = useCallback(fn2);\n        }", []string{"exhaustiveDepsUselessWithoutDependencies", "exhaustiveDepsUselessWithoutDependencies"}},
		{"function MyComponent() {\n          useEffect()\n          useLayoutEffect()\n          useCallback()\n          useMemo()\n        }", []string{"exhaustiveDepsMissingCallback", "exhaustiveDepsMissingCallback", "exhaustiveDepsMissingCallback", "exhaustiveDepsMissingCallback"}},
		{"function MyComponent() {\n          const local = someFunc();\n          useEffect(() => {\n            if (true) {\n              console.log(local);\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            try {\n              console.log(local);\n            } finally {}\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            function inner() {\n              console.log(local);\n            }\n            inner();\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local1 = someFunc();\n          {\n            const local2 = someFunc();\n            useEffect(() => {\n              console.log(local1);\n              console.log(local2);\n            }, []);\n          }\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local1 = {};\n          const local2 = {};\n          useEffect(() => {\n            console.log(local1);\n            console.log(local2);\n          }, [local1]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local1 = {};\n          const local2 = {};\n          useMemo(() => {\n            console.log(local1);\n          }, [local1, local2]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent() {\n          const local1 = someFunc();\n          function MyNestedComponent() {\n            const local2 = {};\n            useCallback(() => {\n              console.log(local1);\n              console.log(local2);\n            }, [local1]);\n          }\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n            console.log(local);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n            console.log(local);\n          }, [local, local]);\n        }", []string{"exhaustiveDepsDuplicate"}},
		{"function MyComponent() {\n          useCallback(() => {}, [window]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          let local = props.foo;\n          useCallback(() => {}, [local]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent({ history }) {\n          useEffect(() => {\n            return history.listen();\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent({ history }) {\n          useEffect(() => {\n            return [\n              history.foo.bar[2].dobedo.listen(),\n              history.foo.bar().dobedo.listen[2]\n            ];\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent({ history }) {\n          useEffect(() => {\n            return [\n              history?.foo\n            ];\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          useEffect(() => {}, ['foo']);\n        }", []string{"exhaustiveDepsLiteral"}},
		{"function MyComponent({ foo, bar, baz }) {\n          useEffect(() => {\n            console.log(foo, bar, baz);\n          }, ['foo', 'bar']);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsLiteral", "exhaustiveDepsLiteral"}},
		{"function MyComponent({ foo, bar, baz }) {\n          useEffect(() => {\n            console.log(foo, bar, baz);\n          }, [42, false, null]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsLiteral", "exhaustiveDepsLiteral", "exhaustiveDepsLiteral"}},
		{"function MyComponent() {\n          const dependencies = [];\n          useEffect(() => {}, dependencies);\n        }", []string{"exhaustiveDepsNotArrayLiteral"}},
		{"function MyComponent() {\n          const local = {};\n          const dependencies = [local];\n          useEffect(() => {\n            console.log(local);\n          }, dependencies);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsNotArrayLiteral"}},
		{"function MyComponent() {\n          const local = {};\n          const dependencies = [local];\n          useEffect(() => {\n            console.log(local);\n          }, [...dependencies]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsSpread"}},
		{"function MyComponent() {\n          const local = someFunc();\n          useEffect(() => {\n            console.log(local);\n          }, [local, ...dependencies]);\n        }", []string{"exhaustiveDepsSpread"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n          }, [computeCacheKey(local)]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsComplexExpression"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.items[0]);\n          }, [props.items[0]]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsComplexExpression"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.items[0]);\n          }, [props.items, props.items[0]]);\n        }", []string{"exhaustiveDepsComplexExpression"}},
		{"function MyComponent({ items }) {\n          useEffect(() => {\n            console.log(items[0]);\n          }, [items[0]]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsComplexExpression"}},
		{"function MyComponent({ items }) {\n          useEffect(() => {\n            console.log(items[0]);\n          }, [items, items[0]]);\n        }", []string{"exhaustiveDepsComplexExpression"}},
		{"function MyComponent(props) {\n          const local = {};\n          useCallback(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, [props, props.foo]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          const local = {};\n          useCallback(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, [props.foo, props]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          const local = {};\n          useCallback(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useCallback(() => {\n            const { foo } = props;\n            console.log(foo);\n          }, [props.bar]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useCallback(() => {\n            const { foo: { bar } } = props;\n            console.log(bar);\n          }, [props.foo.bar]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const foo = props.foo;\n          useEffect(() => {\n            const { bar } = foo();\n            console.log(bar);\n          }, [props.foo.bar]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {id: 42};\n          useEffect(() => {\n            console.log(local);\n          }, [local.id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {id: 42};\n          const fn = useCallback(() => {\n            console.log(local);\n          }, [local.id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {id: 42};\n          const fn = useCallback(() => {\n            console.log(local);\n          }, [local.id, local]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          const fn = useCallback(() => {\n            console.log(props.foo.bar.baz);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let color = {}\n          const fn = useCallback(() => {\n            console.log(props.foo.bar.baz);\n            console.log(color);\n          }, [props.foo, props.foo.bar.baz]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const fn = useCallback(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo.bar.baz, props.foo]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          const fn = useCallback(() => {\n            console.log(props.foo.bar.baz);\n            console.log(props.foo.fizz.bizz);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const fn = useCallback(() => {\n            console.log(props.foo.bar);\n          }, [props.foo.bar.baz]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const fn = useCallback(() => {\n            console.log(props);\n            console.log(props.hello);\n          }, [props.foo.bar.baz]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n          }, [local, local]);\n        }", []string{"exhaustiveDepsDuplicate"}},
		{"function MyComponent() {\n          const local1 = {};\n          useCallback(() => {\n            const local1 = {};\n            console.log(local1);\n          }, [local1]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent() {\n          const local1 = {};\n          useCallback(() => {}, [local1]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let a, b, c, d, e, f, g;\n          useEffect(() => {\n            console.log(b, e, d, c, a, g, f);\n          }, [c, a, g]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let a, b, c, d, e, f, g;\n          useEffect(() => {\n            console.log(b, e, d, c, a, g, f);\n          }, [a, c, g]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let a, b, c, d, e, f, g;\n          useEffect(() => {\n            console.log(b, e, d, c, a, g, f);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const local = {};\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n            console.log(local);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const local = {};\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n            console.log(local);\n          }, [props]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n          }, []);\n          useCallback(() => {\n            console.log(props.foo);\n          }, []);\n          useMemo(() => {\n            console.log(props.foo);\n          }, []);\n          React.useEffect(() => {\n            console.log(props.foo);\n          }, []);\n          React.useCallback(() => {\n            console.log(props.foo);\n          }, []);\n          React.useMemo(() => {\n            console.log(props.foo);\n          }, []);\n          React.notReactiveHook(() => {\n            console.log(props.foo);\n          }, []);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useCustomEffect(() => {\n            console.log(props.foo);\n          }, []);\n          useEffect(() => {\n            console.log(props.foo);\n          }, []);\n          React.useEffect(() => {\n            console.log(props.foo);\n          }, []);\n          React.useCustomEffect(() => {\n            console.log(props.foo);\n          }, []);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n          }, [a ? local : b]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsComplexExpression"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n          }, [a && local]);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsComplexExpression"}},
		{"function MyComponent(props) {\n          useEffect(() => {}, [props?.attribute.method()]);\n        }", []string{"exhaustiveDepsComplexExpression"}},
		{"function MyComponent(props) {\n          useEffect(() => {}, [props.method()]);\n        }", []string{"exhaustiveDepsComplexExpression"}},
		{"function MyComponent() {\n          const ref = useRef();\n          const [state, setState] = useState();\n          useEffect(() => {\n            ref.current = {};\n            setState(state + 1);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const ref = useRef();\n          const [state, setState] = useState();\n          useEffect(() => {\n            ref.current = {};\n            setState(state + 1);\n          }, [ref]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const ref1 = useRef();\n          const ref2 = useRef();\n          useEffect(() => {\n            ref1.current.focus();\n            console.log(ref2.current.textContent);\n            alert(props.someOtherRefs.current.innerHTML);\n            fetch(props.color);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const ref1 = useRef();\n          const ref2 = useRef();\n          useEffect(() => {\n            ref1.current.focus();\n            console.log(ref2.current.textContent);\n            alert(props.someOtherRefs.current.innerHTML);\n            fetch(props.color);\n          }, [ref1.current, ref2.current, props.someOtherRefs, props.color]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          const ref1 = useRef();\n          const ref2 = useRef();\n          useEffect(() => {\n            ref1?.current?.focus();\n            console.log(ref2?.current?.textContent);\n            alert(props.someOtherRefs.current.innerHTML);\n            fetch(props.color);\n          }, [ref1?.current, ref2?.current, props.someOtherRefs, props.color]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent() {\n          const ref = useRef();\n          useEffect(() => {\n            console.log(ref.current);\n          }, [ref.current]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent({ activeTab }) {\n          const ref1 = useRef();\n          const ref2 = useRef();\n          useEffect(() => {\n            ref1.current.scrollTop = 0;\n            ref2.current.scrollTop = 0;\n          }, [ref1.current, ref2.current, activeTab]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent({ activeTab, initY }) {\n          const ref1 = useRef();\n          const ref2 = useRef();\n          const fn = useCallback(() => {\n            ref1.current.scrollTop = initY;\n            ref2.current.scrollTop = initY;\n          }, [ref1.current, ref2.current, activeTab, initY]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent() {\n          const ref = useRef();\n          useEffect(() => {\n            console.log(ref.current);\n          }, [ref.current, ref]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"const MyComponent = forwardRef((props, ref) => {\n          useImperativeHandle(ref, () => ({\n            focus() {\n              alert(props.hello);\n            }\n          }), [])\n        });", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            if (props.onChange) {\n              props.onChange();\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            if (props?.onChange) {\n              props?.onChange();\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            function play() {\n              props.onPlay();\n            }\n            function pause() {\n              props.onPause();\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            if (props.foo.onChange) {\n              props.foo.onChange();\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            props.onChange();\n            if (props.foo.onChange) {\n              props.foo.onChange();\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const [skillsCount] = useState();\n          useEffect(() => {\n            if (skillsCount === 0 && !props.isEditMode) {\n              props.toggleEditMode();\n            }\n          }, [skillsCount, props.isEditMode, props.toggleEditMode]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          const [skillsCount] = useState();\n          useEffect(() => {\n            if (skillsCount === 0 && !props.isEditMode) {\n              props.toggleEditMode();\n            }\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            externalCall(props);\n            props.onChange();\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            props.onChange();\n            externalCall(props);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local1 = 42;\n          const local2 = '42';\n          const local3 = null;\n          const local4 = {};\n          useEffect(() => {\n            console.log(local1);\n            console.log(local2);\n            console.log(local3);\n            console.log(local4);\n          }, [local1, local3]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          useEffect(() => {\n            window.scrollTo(0, 0);\n          }, [window]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"import MutableStore from 'store';\n        function MyComponent() {\n          useEffect(() => {\n            console.log(MutableStore.hello);\n          }, [MutableStore.hello]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"import MutableStore from 'store';\n        let z = {};\n\n        function MyComponent(props) {\n          let x = props.foo;\n          {\n            let y = props.bar;\n            useEffect(() => {\n              console.log(MutableStore.hello.world, props.foo, x, y, z, global.stuff);\n            }, [MutableStore.hello.world, props.foo, x, y, z, global.stuff]);\n          }\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"import MutableStore from 'store';\n        let z = {};\n\n        function MyComponent(props) {\n          let x = props.foo;\n          {\n            let y = props.bar;\n            useEffect(() => {\n              // nothing\n            }, [MutableStore.hello.world, props.foo, x, y, z, global.stuff]);\n          }\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"import MutableStore from 'store';\n        let z = {};\n\n        function MyComponent(props) {\n          let x = props.foo;\n          {\n            let y = props.bar;\n            const fn = useCallback(() => {\n              // nothing\n            }, [MutableStore.hello.world, props.foo, x, y, z, global.stuff]);\n          }\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"import MutableStore from 'store';\n        let z = {};\n\n        function MyComponent(props) {\n          let x = props.foo;\n          {\n            let y = props.bar;\n            const fn = useCallback(() => {\n              // nothing\n            }, [MutableStore?.hello?.world, props.foo, x, y, z, global?.stuff]);\n          }\n        }", []string{"exhaustiveDepsUnnecessary"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n          let [, dispatch] = React.useReducer();\n          let taint = props.foo;\n\n          function handleNext1(value) {\n            let value2 = value * taint;\n            setState(value2);\n            console.log('hello');\n          }\n          const handleNext2 = (value) => {\n            setState(taint(value));\n            console.log('hello');\n          };\n          let handleNext3 = function(value) {\n            setTimeout(() => console.log(taint));\n            dispatch({ type: 'x', value });\n          };\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n          }, []);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext2);\n          }, []);\n          useMemo(() => {\n            return Store.subscribe(handleNext3);\n          }, []);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n          let [, dispatch] = React.useReducer();\n          let taint = props.foo;\n\n          // Shouldn't affect anything\n          function handleChange() {}\n\n          function handleNext1(value) {\n            let value2 = value * taint;\n            setState(value2);\n            console.log('hello');\n          }\n          const handleNext2 = (value) => {\n            setState(taint(value));\n            console.log('hello');\n          };\n          let handleNext3 = function(value) {\n            console.log(taint);\n            dispatch({ type: 'x', value });\n          };\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n          }, []);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext2);\n          }, []);\n          useMemo(() => {\n            return Store.subscribe(handleNext3);\n          }, []);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n          let [, dispatch] = React.useReducer();\n          let taint = props.foo;\n\n          // Shouldn't affect anything\n          const handleChange = () => {};\n\n          function handleNext1(value) {\n            let value2 = value * taint;\n            setState(value2);\n            console.log('hello');\n          }\n          const handleNext2 = (value) => {\n            setState(taint(value));\n            console.log('hello');\n          };\n          let handleNext3 = function(value) {\n            console.log(taint);\n            dispatch({ type: 'x', value });\n          };\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n          }, []);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext2);\n          }, []);\n          useMemo(() => {\n            return Store.subscribe(handleNext3);\n          }, []);\n        }", []string{"exhaustiveDepsMissing", "exhaustiveDepsMissing", "exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n\n          function handleNext(value) {\n            setState(value);\n          }\n\n          useEffect(() => {\n            return Store.subscribe(handleNext);\n          }, [handleNext]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n\n          const handleNext = (value) => {\n            setState(value);\n          };\n\n          useEffect(() => {\n            return Store.subscribe(handleNext);\n          }, [handleNext]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n\n          const handleNext = (value) => {\n            setState(value);\n          };\n\n          useEffect(() => {\n            return Store.subscribe(handleNext);\n          }, [handleNext]);\n\n          return <div onClick={handleNext} />;\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          function handleNext1() {\n            console.log('hello');\n          }\n          const handleNext2 = () => {\n            console.log('hello');\n          };\n          let handleNext3 = function() {\n            console.log('hello');\n          };\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n          }, [handleNext1]);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext2);\n          }, [handleNext2]);\n          useMemo(() => {\n            return Store.subscribe(handleNext3);\n          }, [handleNext3]);\n        }", []string{"exhaustiveDepsConstruction", "exhaustiveDepsConstruction", "exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          function handleNext1() {\n            console.log('hello');\n          }\n          const handleNext2 = () => {\n            console.log('hello');\n          };\n          let handleNext3 = function() {\n            console.log('hello');\n          };\n          useEffect(() => {\n            handleNext1();\n            return Store.subscribe(() => handleNext1());\n          }, [handleNext1]);\n          useLayoutEffect(() => {\n            handleNext2();\n            return Store.subscribe(() => handleNext2());\n          }, [handleNext2]);\n          useMemo(() => {\n            handleNext3();\n            return Store.subscribe(() => handleNext3());\n          }, [handleNext3]);\n        }", []string{"exhaustiveDepsConstruction", "exhaustiveDepsConstruction", "exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          function handleNext1() {\n            console.log('hello');\n          }\n          const handleNext2 = () => {\n            console.log('hello');\n          };\n          let handleNext3 = function() {\n            console.log('hello');\n          };\n          useEffect(() => {\n            handleNext1();\n            return Store.subscribe(() => handleNext1());\n          }, [handleNext1]);\n          useLayoutEffect(() => {\n            handleNext2();\n            return Store.subscribe(() => handleNext2());\n          }, [handleNext2]);\n          useMemo(() => {\n            handleNext3();\n            return Store.subscribe(() => handleNext3());\n          }, [handleNext3]);\n          return (\n            <div\n              onClick={() => {\n                handleNext1();\n                setTimeout(handleNext2);\n                setTimeout(() => {\n                  handleNext3();\n                });\n              }}\n            />\n          );\n        }", []string{"exhaustiveDepsConstruction", "exhaustiveDepsConstruction", "exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          const handleNext1 = () => {\n            console.log('hello');\n          };\n          function handleNext2() {\n            console.log('hello');\n          }\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n            return Store.subscribe(handleNext2);\n          }, [handleNext1, handleNext2]);\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n            return Store.subscribe(handleNext2);\n          }, [handleNext1, handleNext2]);\n        }", []string{"exhaustiveDepsConstruction", "exhaustiveDepsConstruction", "exhaustiveDepsConstruction", "exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          let handleNext = () => {\n            console.log('hello');\n          };\n          if (props.foo) {\n            handleNext = () => {\n              console.log('hello');\n            };\n          }\n          useEffect(() => {\n            return Store.subscribe(handleNext);\n          }, [handleNext]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function MyComponent(props) {\n          let [, setState] = useState();\n          let taint = props.foo;\n\n          function handleNext(value) {\n            let value2 = value * taint;\n            setState(value2);\n            console.log('hello');\n          }\n\n          useEffect(() => {\n            return Store.subscribe(handleNext);\n          }, [handleNext]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Counter() {\n          let [count, setCount] = useState(0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count + 1);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Counter() {\n          let [count, setCount] = useState(0);\n          let [increment, setIncrement] = useState(0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count + increment);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Counter() {\n          let [count, setCount] = useState(0);\n          let [increment, setIncrement] = useState(0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => count + increment);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n           return <h1>{count}</h1>;\n         }", []string{"exhaustiveDepsMissing"}},
		{"function Counter() {\n          let [count, setCount] = useState(0);\n          let increment = useCustomHook();\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => count + increment);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Counter({ step }) {\n          let [count, setCount] = useState(0);\n\n          function increment(x) {\n            return x + step;\n          }\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => increment(count));\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Counter({ step }) {\n          let [count, setCount] = useState(0);\n\n          function increment(x) {\n            return x + step;\n          }\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => increment(count));\n            }, 1000);\n            return () => clearInterval(id);\n          }, [increment]);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Counter({ increment }) {\n          let [count, setCount] = useState(0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => count + increment);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Counter() {\n          const [count, setCount] = useState(0);\n\n          function tick() {\n            setCount(count + 1);\n          }\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              tick();\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Podcasts() {\n          useEffect(() => {\n            alert(podcasts);\n          }, []);\n          let [podcasts, setPodcasts] = useState(null);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Podcasts({ fetchPodcasts, id }) {\n          let [podcasts, setPodcasts] = useState(null);\n          useEffect(() => {\n            fetchPodcasts(id).then(setPodcasts);\n          }, [id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Podcasts({ api: { fetchPodcasts }, id }) {\n          let [podcasts, setPodcasts] = useState(null);\n          useEffect(() => {\n            fetchPodcasts(id).then(setPodcasts);\n          }, [id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Podcasts({ fetchPodcasts, fetchPodcasts2, id }) {\n          let [podcasts, setPodcasts] = useState(null);\n          useEffect(() => {\n            setTimeout(() => {\n              console.log(id);\n              fetchPodcasts(id).then(setPodcasts);\n              fetchPodcasts2(id).then(setPodcasts);\n            });\n          }, [id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Podcasts({ fetchPodcasts, id }) {\n          let [podcasts, setPodcasts] = useState(null);\n          useEffect(() => {\n            console.log(fetchPodcasts);\n            fetchPodcasts(id).then(setPodcasts);\n          }, [id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Podcasts({ fetchPodcasts, id }) {\n          let [podcasts, setPodcasts] = useState(null);\n          useEffect(() => {\n            console.log(fetchPodcasts);\n            fetchPodcasts?.(id).then(setPodcasts);\n          }, [id]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Hello() {\n          const [state, setState] = useState(0);\n          useEffect(() => {\n            setState({});\n          });\n        }", []string{"exhaustiveDepsSetStateNoDependencies"}},
		{"function Hello() {\n          const [data, setData] = useState(0);\n          useEffect(() => {\n            fetchData.then(setData);\n          });\n        }", []string{"exhaustiveDepsSetStateNoDependencies"}},
		{"function Hello({ country }) {\n          const [data, setData] = useState(0);\n          useEffect(() => {\n            fetchData(country).then(setData);\n          });\n        }", []string{"exhaustiveDepsSetStateNoDependencies"}},
		{"function Hello({ prop1, prop2 }) {\n          const [state, setState] = useState(0);\n          useEffect(() => {\n            if (prop1) {\n              setState(prop2);\n            }\n          });\n        }", []string{"exhaustiveDepsSetStateNoDependencies"}},
		{"function Thing() {\n          useEffect(async () => {}, []);\n        }", []string{"exhaustiveDepsAsyncEffect"}},
		{"function Thing() {\n          useEffect(async () => {});\n        }", []string{"exhaustiveDepsAsyncEffect"}},
		{"function Example({ prop }) {\n          const foo = useCallback(() => {\n            prop.hello(foo);\n          }, [foo]);\n          const bar = useCallback(() => {\n            foo();\n          }, [foo]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          function myEffect() {\n            console.log(local);\n          }\n          useEffect(myEffect, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          const myEffect = () => {\n            console.log(local);\n          };\n          useEffect(myEffect, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          const myEffect = function() {\n            console.log(local);\n          };\n          useEffect(myEffect, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          const myEffect = () => {\n            otherThing();\n          };\n          const otherThing = () => {\n            console.log(local);\n          };\n          useEffect(myEffect, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          const myEffect = debounce(() => {\n            console.log(local);\n          }, delay);\n\n          useEffect(myEffect, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const local = {};\n          const myEffect = debounce(() => {\n            console.log(local);\n          }, delay);\n          useEffect(myEffect, [local]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent({myEffect}) {\n          useEffect(myEffect, []);\n        }", []string{"exhaustiveDepsUnknownDependencies"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(debounce(() => {\n            console.log(local);\n          }, delay), []);\n        }", []string{"exhaustiveDepsUnknownDependencies"}},
		{"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          let foo = {}\n          useEffect(() => {\n            foo.bar.baz = 43;\n            props.foo.bar.baz = 1;\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function Component() {\n          const foo = {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = [];\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = () => {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = function bar(){};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = class {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = true ? {} : 'fine';\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = bar || {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = bar ?? {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = bar && {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = bar ? baz ? {} : null : null;\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          let foo = {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          var foo = {};\n          useMemo(() => foo, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = {};\n          useCallback(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = {};\n          useEffect(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = {};\n          useLayoutEffect(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Component() {\n          const foo = {};\n          useImperativeHandle(\n            ref,\n            () => {\n               console.log(foo);\n            },\n            [foo]\n          );\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo(section) {\n          const foo = section.section_components?.edges ?? [];\n          useEffect(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo(section) {\n          const foo = {};\n          console.log(foo);\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = <>Hi!</>;\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = <div>Hi!</div>;\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = bar = {};\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = new String('foo'); // Note 'foo' will be boxed, and thus an object and thus compared by reference.\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = new Map([]);\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = /reg/;\n          useMemo(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          class Bar {};\n          useMemo(() => {\n            console.log(new Bar());\n          }, [Bar]);\n        }", []string{"exhaustiveDepsConstruction"}},
		{"function Foo() {\n          const foo = {};\n          useLayoutEffect(() => {\n            console.log(foo);\n          }, [foo]);\n          useEffect(() => {\n            console.log(foo);\n          }, [foo]);\n        }", []string{"exhaustiveDepsConstruction", "exhaustiveDepsConstruction"}},
		{"import { useEffect } from 'react'\n\n        export const Test = () => {\n          const handleFrame = () => {\n            setTimeout(handleFrame)\n          }\n\n          useEffect(() => {\n            setTimeout(handleFrame)\n          }, [])\n\n          return (\n            <></>\n          )\n        }", []string{"exhaustiveDepsMissing"}},
		{"import { useCallback, useEffect } from \"react\";\n\n        function Component({ foo }) {\n          const log = useCallback(() => {\n          console.log(foo);\n        }, [foo]);\n        useEffect(() => {\n          log();\n        }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent({ theme }) {\n          const onStuff = useEffectEvent(() => {\n            showNotification(theme);\n          });\n          useEffect(() => {\n            onStuff();\n          }, [onStuff]);\n          React.useEffect(() => {\n            onStuff();\n          }, [onStuff]);\n        }", []string{"exhaustiveDepsEffectEvent", "exhaustiveDepsEffectEvent"}},
		{"function MyComponent(props) {\n          if (props.ok) {\n            const callback = () => props.value;\n            useMemo(callback, []);\n          }\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent({ myRef }) {\n    useCallback(() => { console.log(myRef.current); }, [myRef.current]);\n    // React Hook useCallback has a missing dependency: 'myRef'. Either include it or remove the dependency array. Mutable values like 'myRef.current' aren't valid dependencies because mutating them doesn't re-render the component.\n}", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent2({ myRef }) {\n    useCallback(() => { console.log(myRef.current); }, []);\n    // React Hook useCallback has a missing dependency: 'myRef'. Either include it or remove the dependency array.\n}", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent({ obj, flag }) {\n          return useMemo(() => {\n            return (() => {\n              return flag ? obj.a : obj.b;\n            })();\n          }, []);\n        }", []string{"exhaustiveDepsMissing"}},
		{"import React from 'react';\n        export function useThing(context: { paramsLoader: () => void }) {\n          return React.useCallback(() => {\n            const { paramsLoader } = context;\n            paramsLoader();\n          }, [context.paramsLoader]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            const { foo, bar } = props;\n            console.log(foo);\n            console.log(bar);\n          }, [props.foo, props.bar]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            const { bar } = props.foo;\n            console.log(bar);\n          }, [props.foo.bar]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent(props) {\n          useEffect(() => {\n            const { foo: { bar } } = props;\n            console.log(bar);\n          }, [props.foo]);\n        }", []string{"exhaustiveDepsMissing"}},
		{"function MyComponent() {\n          const [, setState] = useState(0);\n          useCallback(() => setState(1), [setState.foo]);\n        }", []string{"exhaustiveDepsUnnecessary"}},
	}

	for index, testCase := range cases {
		t.Run(exhaustiveDepsCaseName(index, testCase.sourceText), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases are where the discrimination lives, which is the usual shape and unusually true
// here: this rule's whole difficulty is staying quiet when the array is already right, and most of
// these were added upstream because somebody hit a false positive.
//
// Three of them are NOT clean upstream and are marked at the line. They are the deliberate
// divergence recorded in the rule's doc comment: a TypeScript non-null assertion inside the
// dependency array, which React turns into two findings because its chain walk throws on a node
// kind it does not name, and which this rule unwraps because `!` cannot change what is read.
func TestExhaustiveDepsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function MyComponent() {\n          const local = {};\n          useEffect(() => {\n            console.log(local);\n          });\n        }",
		"function MyComponent() {\n          useEffect(() => {\n            const local = {};\n            console.log(local);\n          }, []);\n        }",
		"function MyComponent() {\n          const local = someFunc();\n          useEffect(() => {\n            console.log(local);\n          }, [local]);\n        }",
		"function MyComponent() {\n          useEffect(() => {\n            console.log(props.foo);\n          }, []);\n        }",
		"function MyComponent() {\n          const local1 = {};\n          {\n            const local2 = {};\n            useEffect(() => {\n              console.log(local1);\n              console.log(local2);\n            });\n          }\n        }",
		"function MyComponent() {\n          const local1 = someFunc();\n          {\n            const local2 = someFunc();\n            useCallback(() => {\n              console.log(local1);\n              console.log(local2);\n            }, [local1, local2]);\n          }\n        }",
		"function MyComponent() {\n          const local1 = someFunc();\n          function MyNestedComponent() {\n            const local2 = someFunc();\n            useCallback(() => {\n              console.log(local1);\n              console.log(local2);\n            }, [local2]);\n          }\n        }",
		"function MyComponent() {\n          const local = someFunc();\n          useEffect(() => {\n            console.log(local);\n            console.log(local);\n          }, [local]);\n        }",
		"function MyComponent() {\n          useEffect(() => {\n            console.log(unresolved);\n          }, []);\n        }",
		"function MyComponent() {\n          const local = someFunc();\n          useEffect(() => {\n            console.log(local);\n          }, [,,,local,,,]);\n        }",
		"function MyComponent({ foo }) {\n          useEffect(() => {\n            console.log(foo.length);\n          }, [foo]);\n        }",
		"function MyComponent({ foo }) {\n          useEffect(() => {\n            console.log(foo.length);\n            console.log(foo.slice(0));\n          }, [foo]);\n        }",
		"function MyComponent({ history }) {\n          useEffect(() => {\n            return history.listen();\n          }, [history]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {});\n          useLayoutEffect(() => {});\n          useImperativeHandle(props.innerRef, () => {});\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, [props.bar, props.foo]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, [props.foo, props.bar]);\n        }",
		"function MyComponent(props) {\n          const local = someFunc();\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n            console.log(local);\n          }, [props.foo, props.bar, local]);\n        }",
		"function MyComponent(props) {\n          const local = {};\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.bar);\n          }, [props, props.foo]);\n\n          let color = someFunc();\n          useEffect(() => {\n            console.log(props.foo.bar.baz);\n            console.log(color);\n          }, [props.foo, props.foo.bar.baz, color]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            const { foo, bar } = props;\n            console.log(foo);\n            console.log(bar);\n          }, [props]);\n        }",
		"function useThing(context) {\n          return useCallback(() => {\n            const { paramsLoader } = context;\n            paramsLoader();\n          }, [context]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            const [bar] = props.foo;\n            console.log(bar);\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          const foo = useRef()\n          useEffect(() => {\n            const { current: bar } = foo;\n            console.log(bar);\n          }, []);\n        }",
		"function MyComponent(props) {\n          const bar = props.bar;\n          useEffect(() => {\n            const { bar } = foo();\n            console.log(bar);\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo?.bar?.baz ?? null);\n          }, [props.foo]);\n        }",
		// Reports twice upstream, a complex expression and a missing dependency, both
		// caused by `!` being unreadable to its chain walk rather than by a judgment.
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n          }, [props.foo!]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo?.bar);\n          }, [props.foo?.bar]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo?.bar);\n          }, [props.foo.bar]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo.bar);\n          }, [props.foo?.bar]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo.bar);\n            console.log(props.foo?.bar);\n          }, [props.foo?.bar]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo.bar);\n            console.log(props.foo?.bar);\n          }, [props.foo.bar]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo);\n            console.log(props.foo?.bar);\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo?.toString());\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useMemo(() => {\n            console.log(props.foo?.toString());\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.toString());\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo.bar?.toString());\n          }, [props.foo.bar]);\n        }",
		"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.bar?.toString());\n          }, [props.foo.bar]);\n        }",
		"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo.bar.toString());\n          }, [props?.foo?.bar]);\n        }",
		"function MyComponent(props) {\n          useCallback(() => {\n            console.log(props.foo?.bar?.baz);\n          }, [props?.foo.bar?.baz]);\n        }",
		"function MyComponent() {\n          const myEffect = () => {\n            // Doesn't use anything\n          };\n          useEffect(myEffect, []);\n        }",
		"const local = {};\n        function MyComponent() {\n          const myEffect = () => {\n            console.log(local);\n          };\n          useEffect(myEffect, []);\n        }",
		"const local = {};\n        function MyComponent() {\n          function myEffect() {\n            console.log(local);\n          }\n          useEffect(myEffect, []);\n        }",
		"function MyComponent() {\n          const local = someFunc();\n          function myEffect() {\n            console.log(local);\n          }\n          useEffect(myEffect, [local]);\n        }",
		"function MyComponent() {\n          function myEffect() {\n            console.log(global);\n          }\n          useEffect(myEffect, []);\n        }",
		"const local = {};\n        function MyComponent() {\n          const myEffect = () => {\n            otherThing()\n          }\n          const otherThing = () => {\n            console.log(local);\n          }\n          useEffect(myEffect, []);\n        }",
		"function MyComponent({delay}) {\n          const local = {};\n          const myEffect = debounce(() => {\n            console.log(local);\n          }, delay);\n          useEffect(myEffect, [myEffect]);\n        }",
		"function MyComponent({myEffect}) {\n          useEffect(myEffect, [,myEffect]);\n        }",
		"function MyComponent({myEffect}) {\n          useEffect(myEffect, [,myEffect,,]);\n        }",
		"let local = {};\n        function myEffect() {\n          console.log(local);\n        }\n        function MyComponent() {\n          useEffect(myEffect, []);\n        }",
		"function MyComponent({myEffect}) {\n          useEffect(myEffect, [myEffect]);\n        }",
		"function MyComponent({myEffect}) {\n          useEffect(myEffect);\n        }",
		"function MyComponent(props) {\n          useCustomEffect(() => {\n            console.log(props.foo);\n          });\n        }",
		"function MyComponent(props) {\n          useCustomEffect(() => {\n            console.log(props.foo);\n          }, [props.foo]);\n        }",
		"function MyComponent(props) {\n          useCustomEffect(() => {\n            console.log(props.foo);\n          }, []);\n        }",
		"function MyComponent(props) {\n          useWithoutEffectSuffix(() => {\n            console.log(props.foo);\n          }, []);\n        }",
		"function MyComponent(props) {\n          return renderHelperConfusedWithEffect(() => {\n            console.log(props.foo);\n          }, []);\n        }",
		"const local = {};\n        useEffect(() => {\n          console.log(local);\n        }, []);",
		"const local1 = {};\n        {\n          const local2 = {};\n          useEffect(() => {\n            console.log(local1);\n            console.log(local2);\n          }, []);\n        }",
		"function MyComponent() {\n          const ref = useRef();\n          useEffect(() => {\n            console.log(ref.current);\n          }, [ref]);\n        }",
		"function MyComponent() {\n          const ref = useRef();\n          useEffect(() => {\n            console.log(ref.current);\n          }, []);\n        }",
		"function MyComponent({ maybeRef2, foo }) {\n          const definitelyRef1 = useRef();\n          const definitelyRef2 = useRef();\n          const maybeRef1 = useSomeOtherRefyThing();\n          const [state1, setState1] = useState();\n          const [state2, setState2] = React.useState();\n          const [state3, dispatch1] = useReducer();\n          const [state4, dispatch2] = React.useReducer();\n          const [state5, maybeSetState] = useFunnyState();\n          const [state6, maybeDispatch] = useFunnyReducer();\n          const [isPending1] = useTransition();\n          const [isPending2, startTransition2] = useTransition();\n          const [isPending3] = React.useTransition();\n          const [isPending4, startTransition4] = React.useTransition();\n          const mySetState = useCallback(() => {}, []);\n          let myDispatch = useCallback(() => {}, []);\n\n          useEffect(() => {\n            // Known to be static\n            console.log(definitelyRef1.current);\n            console.log(definitelyRef2.current);\n            console.log(maybeRef1.current);\n            console.log(maybeRef2.current);\n            setState1();\n            setState2();\n            dispatch1();\n            dispatch2();\n            startTransition1();\n            startTransition2();\n            startTransition3();\n            startTransition4();\n\n            // Dynamic\n            console.log(state1);\n            console.log(state2);\n            console.log(state3);\n            console.log(state4);\n            console.log(state5);\n            console.log(state6);\n            console.log(isPending2);\n            console.log(isPending4);\n            mySetState();\n            myDispatch();\n\n            // Not sure; assume dynamic\n            maybeSetState();\n            maybeDispatch();\n          }, [\n            // Dynamic\n            state1, state2, state3, state4, state5, state6,\n            maybeRef1, maybeRef2,\n            isPending2, isPending4,\n\n            // Not sure; assume dynamic\n            mySetState, myDispatch,\n            maybeSetState, maybeDispatch\n\n            // In this test, we don't specify static deps.\n            // That should be okay.\n          ]);\n        }",
		"function MyComponent({ maybeRef2 }) {\n          const definitelyRef1 = useRef();\n          const definitelyRef2 = useRef();\n          const maybeRef1 = useSomeOtherRefyThing();\n\n          const [state1, setState1] = useState();\n          const [state2, setState2] = React.useState();\n          const [state3, dispatch1] = useReducer();\n          const [state4, dispatch2] = React.useReducer();\n\n          const [state5, maybeSetState] = useFunnyState();\n          const [state6, maybeDispatch] = useFunnyReducer();\n\n          const mySetState = useCallback(() => {}, []);\n          let myDispatch = useCallback(() => {}, []);\n\n          useEffect(() => {\n            // Known to be static\n            console.log(definitelyRef1.current);\n            console.log(definitelyRef2.current);\n            console.log(maybeRef1.current);\n            console.log(maybeRef2.current);\n            setState1();\n            setState2();\n            dispatch1();\n            dispatch2();\n\n            // Dynamic\n            console.log(state1);\n            console.log(state2);\n            console.log(state3);\n            console.log(state4);\n            console.log(state5);\n            console.log(state6);\n            mySetState();\n            myDispatch();\n\n            // Not sure; assume dynamic\n            maybeSetState();\n            maybeDispatch();\n          }, [\n            // Dynamic\n            state1, state2, state3, state4, state5, state6,\n            maybeRef1, maybeRef2,\n\n            // Not sure; assume dynamic\n            mySetState, myDispatch,\n            maybeSetState, maybeDispatch,\n\n            // In this test, we specify static deps.\n            // That should be okay too!\n            definitelyRef1, definitelyRef2, setState1, setState2, dispatch1, dispatch2\n          ]);\n        }",
		"const MyComponent = forwardRef((props, ref) => {\n          useImperativeHandle(ref, () => ({\n            focus() {\n              alert(props.hello);\n            }\n          }))\n        });",
		"const MyComponent = forwardRef((props, ref) => {\n          useImperativeHandle(ref, () => ({\n            focus() {\n              alert(props.hello);\n            }\n          }), [props.hello])\n        });",
		"function MyComponent(props) {\n          let obj = someFunc();\n          useEffect(() => {\n            obj.foo = true;\n          }, [obj]);\n        }",
		"function MyComponent(props) {\n          let foo = {}\n          useEffect(() => {\n            foo.bar.baz = 43;\n          }, [foo.bar]);\n        }",
		"function MyComponent() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef.current = {};\n            return () => {\n              console.log(myRef.current.toString())\n            };\n          }, []);\n          return <div />;\n        }",
		"function MyComponent() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef.current = {};\n            return () => {\n              console.log(myRef?.current?.toString())\n            };\n          }, []);\n          return <div />;\n        }",
		"function useMyThing(myRef) {\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef.current = {};\n            return () => {\n              console.log(myRef.current.toString())\n            };\n          }, [myRef]);\n        }",
		"function MyComponent() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {};\n            const node = myRef.current;\n            node.addEventListener('mousemove', handleMove);\n            return () => node.removeEventListener('mousemove', handleMove);\n          }, []);\n          return <div ref={myRef} />;\n        }",
		"function useMyThing(myRef) {\n          useEffect(() => {\n            const handleMove = () => {};\n            const node = myRef.current;\n            node.addEventListener('mousemove', handleMove);\n            return () => node.removeEventListener('mousemove', handleMove);\n          }, [myRef]);\n          return <div ref={myRef} />;\n        }",
		"function useMyThing(myRef) {\n          useCallback(() => {\n            const handleMouse = () => {};\n            myRef.current.addEventListener('mousemove', handleMouse);\n            myRef.current.addEventListener('mousein', handleMouse);\n            return function() {\n              setTimeout(() => {\n                myRef.current.removeEventListener('mousemove', handleMouse);\n                myRef.current.removeEventListener('mousein', handleMouse);\n              });\n            }\n          }, [myRef]);\n        }",
		"function useMyThing() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {\n              console.log(myRef.current)\n            };\n            window.addEventListener('mousemove', handleMove);\n            return () => window.removeEventListener('mousemove', handleMove);\n          }, []);\n          return <div ref={myRef} />;\n        }",
		"function useMyThing() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {\n              return () => window.removeEventListener('mousemove', handleMove);\n            };\n            window.addEventListener('mousemove', handleMove);\n            return () => {};\n          }, []);\n          return <div ref={myRef} />;\n        }",
		"function MyComponent() {\n          const local1 = 42;\n          const local2 = '42';\n          const local3 = null;\n          useEffect(() => {\n            console.log(local1);\n            console.log(local2);\n            console.log(local3);\n          }, []);\n        }",
		"function MyComponent() {\n          const local1 = 42;\n          const local2 = '42';\n          const local3 = null;\n          useEffect(() => {\n            console.log(local1);\n            console.log(local2);\n            console.log(local3);\n          }, [local1, local2, local3]);\n        }",
		"function MyComponent(props) {\n          const local = props.local;\n          useEffect(() => {}, [local]);\n        }",
		"function Foo({ activeTab }) {\n          useEffect(() => {\n            window.scrollTo(0, 0);\n          }, [activeTab]);\n        }",
		"function MyComponent(props) {\n          useEffect(() => {\n            console.log(props.foo.bar.baz);\n          }, [props]);\n          useEffect(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo]);\n          useEffect(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo.bar]);\n          useEffect(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo.bar.baz]);\n        }",
		"function MyComponent(props) {\n          const fn = useCallback(() => {\n            console.log(props.foo.bar.baz);\n          }, [props]);\n          const fn2 = useCallback(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo]);\n          const fn3 = useMemo(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo.bar]);\n          const fn4 = useMemo(() => {\n            console.log(props.foo.bar.baz);\n          }, [props.foo.bar.baz]);\n        }",
		"function MyComponent(props) {\n          function handleNext1() {\n            console.log('hello');\n          }\n          const handleNext2 = () => {\n            console.log('hello');\n          };\n          let handleNext3 = function() {\n            console.log('hello');\n          };\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n          }, []);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext2);\n          }, []);\n          useMemo(() => {\n            return Store.subscribe(handleNext3);\n          }, []);\n        }",
		"function MyComponent(props) {\n          function handleNext() {\n            console.log('hello');\n          }\n          useEffect(() => {\n            return Store.subscribe(handleNext);\n          }, []);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext);\n          }, []);\n          useMemo(() => {\n            return Store.subscribe(handleNext);\n          }, []);\n        }",
		"function MyComponent(props) {\n          let [, setState] = useState();\n          let [, dispatch] = React.useReducer();\n\n          function handleNext1(value) {\n            let value2 = value * 100;\n            setState(value2);\n            console.log('hello');\n          }\n          const handleNext2 = (value) => {\n            setState(foo(value));\n            console.log('hello');\n          };\n          let handleNext3 = function(value) {\n            console.log(value);\n            dispatch({ type: 'x', value });\n          };\n          useEffect(() => {\n            return Store.subscribe(handleNext1);\n          }, []);\n          useLayoutEffect(() => {\n            return Store.subscribe(handleNext2);\n          }, []);\n          useMemo(() => {\n            return Store.subscribe(handleNext3);\n          }, []);\n        }",
		"function useInterval(callback, delay) {\n          const savedCallback = useRef();\n          useEffect(() => {\n            savedCallback.current = callback;\n          });\n          useEffect(() => {\n            function tick() {\n              savedCallback.current();\n            }\n            if (delay !== null) {\n              let id = setInterval(tick, delay);\n              return () => clearInterval(id);\n            }\n          }, [delay]);\n        }",
		"function Counter() {\n          const [count, setCount] = useState(0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(c => c + 1);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"function Counter(unstableProp) {\n          let [count, setCount] = useState(0);\n          setCount = unstableProp\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(c => c + 1);\n            }, 1000);\n            return () => clearInterval(id);\n          }, [setCount]);\n\n          return <h1>{count}</h1>;\n        }",
		"function Counter() {\n          const [count, setCount] = useState(0);\n\n          function tick() {\n            setCount(c => c + 1);\n          }\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              tick();\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"function Counter() {\n          const [count, dispatch] = useReducer((state, action) => {\n            if (action === 'inc') {\n              return state + 1;\n            }\n          }, 0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              dispatch('inc');\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"function Counter() {\n          const [count, dispatch] = useReducer((state, action) => {\n            if (action === 'inc') {\n              return state + 1;\n            }\n          }, 0);\n\n          const tick = () => {\n            dispatch('inc');\n          };\n\n          useEffect(() => {\n            let id = setInterval(tick, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"function Podcasts() {\n          useEffect(() => {\n            setPodcasts([]);\n          }, []);\n          let [podcasts, setPodcasts] = useState(null);\n        }",
		"function withFetch(fetchPodcasts) {\n          return function Podcasts({ id }) {\n            let [podcasts, setPodcasts] = useState(null);\n            useEffect(() => {\n              fetchPodcasts(id).then(setPodcasts);\n            }, [id]);\n          }\n        }",
		"function Podcasts({ id }) {\n          let [podcasts, setPodcasts] = useState(null);\n          useEffect(() => {\n            function doFetch({ fetchPodcasts }) {\n              fetchPodcasts(id).then(setPodcasts);\n            }\n            doFetch({ fetchPodcasts: API.fetchPodcasts });\n          }, [id]);\n        }",
		"function Counter() {\n          let [count, setCount] = useState(0);\n\n          function increment(x) {\n            return x + 1;\n          }\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(increment);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"function Counter() {\n          let [count, setCount] = useState(0);\n\n          function increment(x) {\n            return x + 1;\n          }\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => increment(count));\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"import increment from './increment';\n        function Counter() {\n          let [count, setCount] = useState(0);\n\n          useEffect(() => {\n            let id = setInterval(() => {\n              setCount(count => count + increment);\n            }, 1000);\n            return () => clearInterval(id);\n          }, []);\n\n          return <h1>{count}</h1>;\n        }",
		"function withStuff(increment) {\n          return function Counter() {\n            let [count, setCount] = useState(0);\n\n            useEffect(() => {\n              let id = setInterval(() => {\n                setCount(count => count + increment);\n              }, 1000);\n              return () => clearInterval(id);\n            }, []);\n\n            return <h1>{count}</h1>;\n          }\n        }",
		"function App() {\n          const [query, setQuery] = useState('react');\n          const [state, setState] = useState(null);\n          useEffect(() => {\n            let ignore = false;\n            fetchSomething();\n            async function fetchSomething() {\n              const result = await (await fetch('http://hn.algolia.com/api/v1/search?query=' + query)).json();\n              if (!ignore) setState(result);\n            }\n            return () => { ignore = true; };\n          }, [query]);\n          return (\n            <>\n              <input value={query} onChange={e => setQuery(e.target.value)} />\n              {JSON.stringify(state)}\n            </>\n          );\n        }",
		"function Hello() {\n          const [state, setState] = useState(0);\n          useEffect(() => {\n            const handleResize = () => setState(window.innerWidth);\n            window.addEventListener('resize', handleResize);\n            return () => window.removeEventListener('resize', handleResize);\n          });\n        }",
		"function Example() {\n          useEffect(() => {\n            arguments\n          }, [])\n        }",
		"function Example() {\n          useEffect(() => {\n            const bar = () => {\n              arguments;\n            };\n            bar();\n          }, [])\n        }",
		"function Example(props) {\n          useEffect(() => {\n            let topHeight = 0;\n            topHeight = props.upperViewHeight;\n          }, [props.upperViewHeight]);\n        }",
		"function Example(props) {\n          useEffect(() => {\n            let topHeight = 0;\n            topHeight = props?.upperViewHeight;\n          }, [props?.upperViewHeight]);\n        }",
		"function Example(props) {\n          useEffect(() => {\n            let topHeight = 0;\n            topHeight = props?.upperViewHeight;\n          }, [props]);\n        }",
		"function useFoo(foo){\n          return useMemo(() => foo, [foo]);\n        }",
		"function useFoo(){\n          const foo = 'hi!';\n          return useMemo(() => foo, [foo]);\n        }",
		"function useFoo(){\n          let {foo} = {foo: 1};\n          return useMemo(() => foo, [foo]);\n        }",
		"function useFoo(){\n          let [foo] = [1];\n          return useMemo(() => foo, [foo]);\n        }",
		"function useFoo() {\n          const foo = 'fine';\n          if (true) {\n            // Shadowed variable with constant construction in a nested scope is fine.\n            const foo = {};\n          }\n          return useMemo(() => foo, [foo]);\n        }",
		"function MyComponent({foo}) {\n          return useMemo(() => foo, [foo])\n        }",
		"function MyComponent() {\n          const foo = true ? 'fine' : 'also fine';\n          return useMemo(() => foo, [foo]);\n        }",
		"function MyComponent() {\n          useEffect(() => {\n            console.log('banana banana banana');\n          }, undefined);\n        }",
		"const useVirtuosoItems = <T extends ListItem>() => {\n  const groups = useAtomValue();\n  const groupCollapsedState = useAtomValue();\n\n  return useMemo(() => {\n    groupCollapsedState;\n    groups;\n    const items: VirtuosoItem<T>[] = [];\n    return items;\n  }, [groupCollapsedState, groups]);\n};",
		"function MyComponent() {\n    const options = useOptions();\n    useEffect(() => {\n        if (!dropTargetRef.current) {\n            return;\n        }\n        return dropTargetForElements({\n            onDropTargetChange: (args) => {\n                if (options && dropTargetRef.current) {\n                  delete dropTargetRef.current.dataset['draggedOver'];\n                }\n            }\n        });\n    }, [options]);\n}",
		"export function useCanvasZoomOrScroll() {\n           useEffect(() => {\n               let wheelStopTimeoutId: { current: number | undefined } = { current: undefined };\n\n               wheelStopTimeoutId = requestAnimationFrameTimeout(() => {\n                   setLastInteraction?.(null);\n               }, 300);\n\n               return () => {\n                   if (wheelStopTimeoutId.current !== undefined) {\n                       console.log('h1');\n                   }\n               };\n           }, []);\n        }",
		"function X() {\n  const defaultParam1 = \"\";\n  const myFunction = useCallback(\n    (param1 = defaultParam1, param2) => {\n    },\n    [defaultParam1]\n  );\n\n  return null;\n}\n",
		"function MyComponent() { const recursive = useCallback((n: number): number => (n <= 0 ? 0 : n + recursive(n - 1)), []); return recursive }",
		// Reports twice upstream, a complex expression and a missing dependency, both
		// caused by `!` being unreadable to its chain walk rather than by a judgment.
		"function MyComponent(props) { useEffect(() => { console.log(props.foo!.bar) }, [props.foo!.bar]) }",
		// Reports twice upstream, a complex expression and a missing dependency, both
		// caused by `!` being unreadable to its chain walk rather than by a judgment.
		"function MyComponent(props) { useEffect(() => { console.log((props.foo).bar) }, [props.foo!.bar]) }",
		"function MyComponent(props) { const external = {}; const y = useMemo(() => { const z = foo<typeof external>(); return z; }, []) }",
		"function Test() { const [state, setState] = useState(); useEffect(() => { console.log(\"state\", state); }); }",
		"function MyComponent({ theme }) {\n          const onStuff = useEffectEvent(() => {\n            showNotification(theme);\n          });\n          useEffect(() => {\n            onStuff();\n          }, []);\n          React.useEffect(() => {\n            onStuff();\n          }, []);\n        }",
		"export const FileSize = ({ file, showSize = true }) => {\n          const fileSizeInMB = useMemo(\n            () => (showSize ? (file.size / (1024 * 1024)).toFixed(2) : undefined),\n            [showSize, file.size],\n          );\n          return fileSizeInMB;\n        }",
		"function MyComponent({ obj }) {\n          useMemo(() => {\n            return (obj.value * 2).toFixed(2);\n          }, [obj.value]);\n        }",
		"function MyComponent({ data }) {\n          useCallback(() => {\n            console.log((data.count + 1).toString());\n          }, [data.count]);\n        }",
		"\n          function getColumns() { return []; }\n          function MyComponent() {\n            const columns = useMemo(getColumns, []);\n            return columns;\n          }\n        ",
		"\n          const getFields = () => { return []; };\n          function MyComponent() {\n            const fields = useMemo(getFields, []);\n            return fields;\n          }\n        ",
		"\n          function checkIfIsSafari() { return false; }\n          function MyComponent() {\n            const isSafari = useMemo(checkIfIsSafari, []);\n            return isSafari;\n          }\n        ",
		"\n          function setup() { console.log('setup'); }\n          function MyComponent() {\n            useEffect(setup, []);\n          }\n        ",
		"function MyComponent4({ myRef }) { useCallback(() => { console.log(myRef.current); }, [myRef]); }",
		"function MyComponent({ obj, flag }) {\n          return useMemo(() => {\n            return (() => {\n              return flag ? obj.a : obj.b;\n            })();\n          }, [obj.a, obj.b, flag]);\n        }",
		"function MyComponent({ obj }) {\n          return useMemo(() => {\n            return (function() {\n              return obj.a;\n            })();\n          }, [obj.a]);\n        }",
		"function MyComponent({ obj }) {\n          return useMemo(() => {\n            return (() => {\n              obj.method();\n            })();\n          }, [obj]);\n        }",
		"import { useEffect, useState } from \"react\";\n\nexport const useTest = () => {\n    const [state] = useState<Record<\"a\", string>>({\n        a: \"a\",\n    });\n\n    useEffect(() => {\n        const a = \"a\" as keyof typeof state;\n        console.log(a);\n    }, []);\n\n    console.log(state);\n}",
		"const Component = () => {\n  const DATA = 'test' as const;\n  const data = useMemo(() => DATA, []);\n  return <div>{data}</div>;\n};",
		"const ReactActual = jest.requireActual('react');\nconst Component = ({ filter }) => {\n    const [data, setData] = ReactActual.useState(filter);\n    ReactActual.useEffect(() => {\n        setData(filter);\n    }, [filter]);\n\n    return <div>test</div>;\n};",
		"const Component = () => {\n          const setRef = React.useRef<(value: string) => void | null>(\n            null,\n          ) as React.MutableRefObject<(value: string) => void | null>;\n\n          React.useEffect(() => {\n            if (setRef.current) {\n              console.log(setRef.current);\n            }\n          }, []);\n\n          return <div>test</div>;\n        };",
		"function MyComponent() {\n          const condition = new Date('2026-01-01') > new Date();\n          useCallback(() => {\n            return condition;\n          }, [condition]);\n        }",
		"function MyComponent() {\n          const label = 'total: ' + [1, 2, 3].length;\n          useCallback(() => {\n            return label;\n          }, [label]);\n        }",
		"function MyComponent(props) {\n          const isError = props.value instanceof Error;\n          useMemo(() => isError, [isError]);\n        }",
	}

	for index, sourceText := range cases {
		t.Run(exhaustiveDepsCaseName(index, sourceText), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, sourceText))
		})
	}
}

// The eight corpus cases this port does not carry, named rather than deleted.
//
// Step 5 of the porting standard says to copy every case verbatim, and a case that is dropped
// silently looks identical to a case nobody noticed. These are dropped for SCOPE rather than for
// disagreement: each asserts only the two diagnostic kinds this port leaves out, or mixes one of
// those with a kind it does carry so no expectation over the whole input would be honest.
//
// The test asserts they still parse and that the rule survives them, which is the part that can
// regress: an unported kind's input must not crash the walk or produce a finding of a DIFFERENT
// kind, and only running them says so.
func TestExhaustiveDepsScopeIsStated(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		reason     string
	}{
		{"function MyComponent(props) {\n          let value;\n          let value2;\n          let value3;\n          let value4;\n          let asyncValue;\n          useEffect(() => {\n            if (value4) {\n              value = {};\n            }\n            value2 = 100;\n            value = 43;\n            value4 = true;\n            console.log(value2);\n            console.log(value3);\n            setTimeout(() => {\n              asyncValue = 100;\n            });\n          }, []);\n        }",
			"asserts only unported kinds"},
		{"function MyComponent(props) {\n          let value;\n          let value2;\n          let value3;\n          let asyncValue;\n          useEffect(() => {\n            value = {};\n            value2 = 100;\n            value = 43;\n            console.log(value2);\n            console.log(value3);\n            setTimeout(() => {\n              asyncValue = 100;\n            });\n          }, [value, value2, value3]);\n        }",
			"asserts only unported kinds"},
		{"function MyComponent() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef.current.addEventListener('mousemove', handleMove);\n            return () => myRef.current.removeEventListener('mousemove', handleMove);\n          }, []);\n          return <div ref={myRef} />;\n        }",
			"asserts only unported kinds"},
		{"function MyComponent() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef?.current?.addEventListener('mousemove', handleMove);\n            return () => myRef?.current?.removeEventListener('mousemove', handleMove);\n          }, []);\n          return <div ref={myRef} />;\n        }",
			"asserts only unported kinds"},
		{"function MyComponent() {\n          const myRef = useRef();\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef.current.addEventListener('mousemove', handleMove);\n            return () => myRef.current.removeEventListener('mousemove', handleMove);\n          });\n          return <div ref={myRef} />;\n        }",
			"asserts only unported kinds"},
		{"function useMyThing(myRef) {\n          useEffect(() => {\n            const handleMove = () => {};\n            myRef.current.addEventListener('mousemove', handleMove);\n            return () => myRef.current.removeEventListener('mousemove', handleMove);\n          }, [myRef]);\n        }",
			"asserts only unported kinds"},
		{"function useMyThing(myRef) {\n          useEffect(() => {\n            const handleMouse = () => {};\n            myRef.current.addEventListener('mousemove', handleMouse);\n            myRef.current.addEventListener('mousein', handleMouse);\n            return function() {\n              setTimeout(() => {\n                myRef.current.removeEventListener('mousemove', handleMouse);\n                myRef.current.removeEventListener('mousein', handleMouse);\n              });\n            }\n          }, [myRef]);\n        }",
			"asserts only unported kinds"},
		{"function useMyThing(myRef, active) {\n          useEffect(() => {\n            const handleMove = () => {};\n            if (active) {\n              myRef.current.addEventListener('mousemove', handleMove);\n              return function() {\n                setTimeout(() => {\n                  myRef.current.removeEventListener('mousemove', handleMove);\n                });\n              }\n            }\n          }, [myRef, active]);\n        }",
			"asserts only unported kinds"},
	}

	for index, testCase := range cases {
		t.Run(exhaustiveDepsCaseName(index, testCase.sourceText), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, testCase.sourceText)
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Message.Id == "exhaustiveDepsRefCleanup" ||
					diagnostic.Message.Id == "exhaustiveDepsStaleAssignment" {
					t.Errorf("this port does not carry %s, so nothing should report it (%s)",
						diagnostic.Message.Id, testCase.reason)
				}
			}
		})
	}
}

// exhaustiveDepsCaseName gives a subtest a stable, readable name without putting a whole
// multi-line component into it.
func exhaustiveDepsCaseName(index int, sourceText string) string {
	first := sourceText
	if cut := strings.IndexByte(first, '\n'); cut >= 0 {
		first = first[:cut]
	}
	if len(first) > 60 {
		first = first[:60]
	}
	return exhaustiveDepsIndexLabel(index) + " " + first
}

func exhaustiveDepsIndexLabel(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// A rule declaring the checker must be handed one, and the plain harness hands it nil.
//
// The failure this guards is silence rather than a crash: `GetSymbolAtLocation` on a nil checker
// returns nil rather than panicking, so every reference would look unresolved, every dependency
// would look external, and the rule would go completely quiet while every StaysSilent case above
// passed vacuously. A later revert to `rule_testing.Run` would look like a green suite.
func TestExhaustiveDepsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !ExhaustiveDeps.NeedsTypeChecker {
		t.Fatal("this rule resolves every identifier through the checker and must declare it")
	}
	source := "function C(props) { useEffect(() => { console.log(props.foo); }, []); }"
	if findings := rule_testing.Run(t, ExhaustiveDeps, exhaustiveDepsFile, source); len(findings.Diagnostics) != 0 {
		t.Fatalf("the untyped harness produced %d findings, so this test is not measuring what it claims",
			len(findings.Diagnostics))
	}
	if findings := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, source); len(findings.Diagnostics) != 1 {
		t.Fatalf("the typed harness produced %d findings, want 1", len(findings.Diagnostics))
	}
}

// The `additionalHooks` option, and the effect-name boundary it exposes.
//
// The inventory says this rule takes no options. It is wrong: oxc declares `additionalHooks` and
// React's `meta.schema` declares that plus three more. Every case here goes through the rule's own
// exported decoder rather than through a hand-built struct, so the JSON tag and the empty-string
// case are under test rather than assumed.
//
// The last two rows are the reason this test exists rather than only the first three. A mutation
// replacing the effect-name pattern `Effect($|[^a-z])` with a plain `Effect` substring test survived
// the entire 325-case corpus, because neither upstream corpus writes a hook name where the two
// disagree. The axis is observable and was measured against React before either row was written:
// an EFFECT keeps a declared dependency it never reads, and a non-effect calls that dependency
// unnecessary. So `useEffective` reports and `useMyEffect2` is silent, and a substring test gets
// both backwards.
func TestExhaustiveDepsAdditionalHooks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		optionsJson string
		sourceText  string
		wantIds     []string
	}{
		{"an unnamed hook is not checked at all", `{}`,
			"function C() { const a = someFunc(); useCustomEffect(() => { console.log(a); }, []); }",
			nil},
		{"a named hook is checked", `{"additionalHooks": "useCustomEffect"}`,
			"function C() { const a = someFunc(); useCustomEffect(() => { console.log(a); }, []); }",
			[]string{"exhaustiveDepsMissing"}},
		{"an empty pattern names nothing", `{"additionalHooks": ""}`,
			"function C() { const a = someFunc(); useCustomEffect(() => { console.log(a); }, []); }",
			nil},
		{"a malformed pattern disables the extension rather than failing the run",
			`{"additionalHooks": "["}`,
			"function C() { const a = someFunc(); useCustomEffect(() => { console.log(a); }, []); }",
			nil},
		// `Effect` followed by a lowercase letter is NOT an effect, so an unread declared
		// dependency is unnecessary rather than kept.
		{"Effect followed by a lowercase letter is not an effect",
			`{"additionalHooks": "useEffective"}`,
			"function C() { const a = someFunc(); useEffective(() => {}, [a]); }",
			[]string{"exhaustiveDepsUnnecessary"}},
		// `Effect` followed by a digit IS an effect, so the same dependency is kept.
		{"Effect followed by a digit is an effect",
			`{"additionalHooks": "useMyEffect2"}`,
			"function C() { const a = someFunc(); useMyEffect2(() => {}, [a]); }",
			nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := rule.DecodeOptionsInto[ExhaustiveDepsOptions]()(
				json.RawMessage(testCase.optionsJson))
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.optionsJson, err)
			}
			result := rule_testing.RunTypedWithOptions(t, ExhaustiveDeps, exhaustiveDepsFile,
				testCase.sourceText, decoded)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The suggested dependency array, asserted as text.
//
// This is the assertion that separates this rule from every other one in the catalog, and it is not
// optional here. `ExpectFindings` compares message ids and a count and nothing else, so a rule that
// names the RIGHT problem and then advises the WRONG array passes a complete fixture pair. That is
// not a hypothetical: a mutation stopping optional chaining from being recorded survived all 325
// corpus cases, because every one of them asserts which finding fired and none asserts what it
// offered. The array is the whole product of this rule, so it gets its own test.
//
// Every expected string below was taken from React 7.1.1's own suggestion on the same input, and
// the same comparison was run over seventeen real components from this repository, where all
// seventeen suggested arrays matched character for character.
func TestExhaustiveDepsSuggestsTheCorrectedArray(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantAdvice string
	}{
		{"a plain missing dependency",
			"function C(props) { useEffect(() => { console.log(props.foo); }, []); }",
			"Update the dependency array to [props.foo]."},
		{"a property path is kept whole",
			"function C(props) { useEffect(() => { console.log(props.foo.bar.baz); }, []); }",
			"Update the dependency array to [props.foo.bar.baz]."},
		// The mutant M10 exists for: an optional step has to be written back as `?.`, because
		// `props.foo.bar` throws where `props.foo?.bar` did not. No id assertion can see this.
		{"an optional step is written back as an optional step",
			"function C(props) { useEffect(() => { console.log(props.foo?.bar); }, []); }",
			"Update the dependency array to [props.foo?.bar]."},
		{"a deeper optional chain keeps the marker where it was written",
			"function C(props) { useEffect(() => { console.log(props.foo?.bar.baz); }, []); }",
			"Update the dependency array to [props.foo?.bar.baz]."},
		// An author who already sorted gets a sorted answer; one who did not keeps their order.
		// The two rows have identical ids and identical counts and differ only here.
		{"an already sorted array stays sorted",
			"function C(props) { useEffect(() => { console.log(props.a, props.b, props.c); }, [props.a, props.b]); }",
			"Update the dependency array to [props.a, props.b, props.c]."},
		{"a deliberately unsorted array keeps its order",
			"function C(props) { useEffect(() => { console.log(props.a, props.b, props.c); }, [props.b, props.a]); }",
			"Update the dependency array to [props.b, props.a, props.c]."},
		// A non-effect recomputes the whole array, so fixing the missing one drops the unnecessary
		// one. An effect would keep it. This row is the only thing that pins that asymmetry.
		{"a non-effect drops an unnecessary dependency while fixing a missing one",
			"function C(props) { const unused = someFunc(); useMemo(() => { console.log(props.a); }, [unused]); }",
			"Update the dependency array to [props.a]."},
		{"a stable value is never suggested",
			"function C(props) { const [s, setS] = useState(0); useEffect(() => { setS(props.a); }, []); }",
			"Update the dependency array to [props.a]."},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding to read the advice from, got %d",
					len(result.Diagnostics))
			}
			if len(result.Diagnostics[0].Suggestions) != 1 {
				t.Fatalf("want exactly one suggestion, got %d",
					len(result.Diagnostics[0].Suggestions))
			}
			// Equality on the rendered text, not `strings.Contains`. A predicate weaker than the
			// property it guards is not a guard, and `[props.foo]` is a substring of
			// `[props.foo, wrong]`.
			if advice := result.Diagnostics[0].Suggestions[0].Message.Description; advice != testCase.wantAdvice {
				t.Errorf("advice was %q, want %q", advice, testCase.wantAdvice)
			}
		})
	}
}

// The corrected array is offered as a SUGGESTION and never as a FIX.
//
// Both upstreams already say this rewrite must not be applied unattended: React promotes it to a fix
// only under an option named `enableDangerousAutofixThisMayCauseInfiniteLoops`, and oxc classifies
// its own version as a dangerous suggestion in five of its six dependency-array fix vectors. The
// reason is the asymmetry in the rule's doc comment — an array naming a value rebuilt every render
// re-runs the effect every render, and an effect that sets state then never stops.
//
// `internal/fix.ProposalsFrom` collects `Fixes` and ignores `Suggestions`, so this is the mechanical
// difference between advice a person reads and an edit that lands in their file unattended. A later
// change moving the rewrite into `Fixes` would be invisible to every other test here.
func TestExhaustiveDepsProposesNoAutomaticFix(t *testing.T) {
	t.Parallel()

	source := "function C(props) { useEffect(() => { console.log(props.foo); }, []); }"
	result := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if fixes := len(result.Diagnostics[0].Fixes); fixes != 0 {
		t.Errorf("this rule proposes %d automatic fixes and must propose none: a wrong dependency array applied unattended is an infinite render loop", fixes)
	}
	if suggestions := len(result.Diagnostics[0].Suggestions); suggestions != 1 {
		t.Errorf("the corrected array must still be offered as a suggestion, got %d", suggestions)
	}
}

// The every-render construction check, on shapes the imported corpus does not separate.
//
// `getConstructionExpressionType` recurses through a ternary and through the logical operators, so
// a value built in EITHER branch makes the whole expression a construction. A mutation looking only
// at a ternary's first branch survived all 325 corpus cases: every conditional it writes happens to
// construct on the left. The rows below were measured against React 7.1.1 before they were written,
// which is what separates a fixture from a guess about a survivor.
//
// The third row is the one that makes the other two mean something. Without a ternary that
// constructs in neither branch, a rule answering "any ternary is a construction" would pass the
// first two and be wrong about ordinary code.
func TestExhaustiveDepsConstructionRecursesThroughBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reports    bool
	}{
		{"a ternary constructing in its first branch",
			"function C() { const o = cond ? {} : other; useEffect(() => { console.log(o); }, [o]); }",
			true},
		{"a ternary constructing in its second branch",
			"function C() { const o = cond ? other : {}; useEffect(() => { console.log(o); }, [o]); }",
			true},
		{"a ternary constructing in neither branch",
			"function C() { const o = cond ? a : b; useEffect(() => { console.log(o); }, [o]); }",
			false},
		{"a logical operator constructing on its right",
			"function C() { const o = a || {}; useEffect(() => { console.log(o); }, [o]); }",
			true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, testCase.sourceText)
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "exhaustiveDepsConstruction")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Where the finding points, which no message-id assertion can see.
//
// Three of this rule's findings point at three different things and the choice is upstream's rather
// than convenient: the array as a whole for a list problem, the offending entry for an entry
// problem, and the DECLARATION for a construction, because the fix is at the declaration and not at
// the Hook. A rule pointing all three at the Hook call would pass every other test in this file.
func TestExhaustiveDepsPointsAtTheRightNode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"a list problem points at the whole array",
			"function C(props) { useEffect(() => { console.log(props.foo); }, []); }",
			"[]"},
		{"an entry problem points at the entry",
			"function C(props) { useEffect(() => { console.log(props.foo); }, [props.foo, 'x']); }",
			"'x'"},
		{"a construction points at the declaration, where the repair goes",
			"function C() { const o = {}; useEffect(() => { console.log(o); }, [o]); }",
			"o = {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatal("no finding to read a span from")
			}
			last := result.Diagnostics[len(result.Diagnostics)-1]
			reported := result.SourceFile.Text()[last.Range.Pos():last.Range.End()]
			if reported != testCase.wantText {
				t.Errorf("the finding covers %q, want %q", reported, testCase.wantText)
			}
		})
	}
}

// The non-effect recompute, and the bug a surviving mutant found.
//
// `useMemo` and `useCallback` rebuild the whole suggested array from the reads rather than adding
// the missing entries to what was declared, so fixing a missing dependency also drops the entries
// that were never read. An effect keeps them. The corpus never separates the two, because a case
// needs a satisfying entry AND a missing one AND an author order that is not the sorted order, and
// no upstream case has all three.
//
// This test exists because a mutation disabling the recompute SURVIVED all 325 corpus cases, and
// then a direct comparison of the two code paths showed the mutant matching React on three shapes
// while the original did not. The survivor was right about the code: the missing paths were being
// sorted at gather time to make a Go map deterministic, which quietly sorted every suggestion. That
// is the sixth survivor category and this brief's list does not name it — a mutant that survives
// because the ORIGINAL is wrong.
//
// Every expectation below was read off React 7.1.1 before it was written.
func TestExhaustiveDepsNonEffectRebuildsTheArray(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantAdvice string
	}{
		{"a non-effect keeps the author's order for what it re-adds",
			"function C(props) { useMemo(() => { console.log(props.c, props.b, props.a); }, [props.c, props.b]); }",
			"Update the dependency array to [props.c, props.b, props.a]."},
		{"a non-effect drops an outer-scope entry it never reads",
			"const o = 1; function C(props) { useMemo(() => { console.log(props.b, props.a); }, [props.b, o]); }",
			"Update the dependency array to [props.b, props.a]."},
		// THE pair that separates the two code paths, and it took three wrong guesses to find.
		//
		// It needs all four of: a missing dependency, at least two SATISFYING entries, a declared
		// order that differs from the READ order, and a declared array that is not already sorted so
		// no re-sort masks the difference. The callback reads b then c; the array declares c then b.
		// A non-effect rebuilds from the reads and answers in READ order; an effect adds to what was
		// declared and answers in DECLARED order. Both were read off React 7.1.1.
		//
		// A mutation disabling the recompute survived the 325-case corpus and then survived three
		// fixtures written for it, each from a hypothesis that was wrong about which input could
		// see it. The rule this brief already states is what got here: when a fixture written for a
		// survivor does not kill it, the hypothesis is what needs replacing, not the fixture count.
		{"a non-effect answers in read order where an effect answers in declared order",
			"function C(props) { useMemo(() => { console.log(props.b, props.c, props.a); }, [props.c, props.b]); }",
			"Update the dependency array to [props.b, props.c, props.a]."},
		{"the same input as an effect keeps the declared order",
			"function C(props) { useEffect(() => { console.log(props.b, props.c, props.a); }, [props.c, props.b]); }",
			"Update the dependency array to [props.c, props.b, props.a]."},
		{"a non-effect drops an unread local while keeping read order",
			"function C(props) { const u = someFunc(); useCallback(() => { console.log(props.b, props.a); }, [u, props.b]); }",
			"Update the dependency array to [props.b, props.a]."},
		// The effect side, so the asymmetry is pinned from both directions rather than only from
		// the branch that changes. The distinguishing entry has to be a LOCAL that is never read:
		// an outer-scope entry is dropped by BOTH, which is a wrong guess this test made first and
		// React corrected. The two rows below are the same input under two hook names.
		{"an effect keeps an unread local",
			"function C(props) { const u = someFunc(); useEffect(() => { console.log(props.b, props.a); }, [props.b, u]); }",
			"Update the dependency array to [props.a, props.b, u]."},
		{"a non-effect drops the same unread local",
			"function C(props) { const u = someFunc(); useMemo(() => { console.log(props.b, props.a); }, [props.b, u]); }",
			"Update the dependency array to [props.a, props.b]."},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ExhaustiveDeps, exhaustiveDepsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].Suggestions) != 1 {
				t.Fatalf("want one finding carrying one suggestion, got %d findings",
					len(result.Diagnostics))
			}
			if advice := result.Diagnostics[0].Suggestions[0].Message.Description; advice != testCase.wantAdvice {
				t.Errorf("advice was %q, want %q", advice, testCase.wantAdvice)
			}
		})
	}
}

// exhaustiveDepsWithOptions runs a fixture through the registered decoder with the given options.
func exhaustiveDepsWithOptions(t *testing.T, optionsJson string, sourceText string) rule_testing.Result {
	t.Helper()
	decoded, err := rule.DecodeOptionsInto[ExhaustiveDepsOptions]()(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", optionsJson, err)
	}
	return rule_testing.RunTypedWithOptions(t, ExhaustiveDeps, exhaustiveDepsFile, sourceText, decoded)
}

// requireExplicitEffectDeps, from React's row and its controls: an effect with no second argument
// reports, an explicit `undefined` is an answer, and the option is what turns it on.
func TestExhaustiveDepsRequireExplicitEffectDeps(t *testing.T) {
	t.Parallel()

	const required = `{"requireExplicitEffectDeps": true}`
	cases := []struct {
		name        string
		optionsJson string
		sourceText  string
		wantIds     []string
	}{
		{"React's row: an effect with no array", required,
			"function MyComponent(props) { useEffect(() => { console.log(props.foo); }); }",
			[]string{"exhaustiveDepsRequireExplicitEffectDeps"}},
		{"an explicit undefined answers it", required,
			"function MyComponent(props) { useEffect(() => { console.log(props.foo); }, undefined); }",
			nil},
		{"without the option the same effect is clean", `{}`,
			"function MyComponent(props) { useEffect(() => { console.log(props.foo); }); }",
			nil},
		{"a Hook that is not an effect is not asked", required,
			"function MyComponent(props) { useMemo(() => props.foo); }",
			[]string{"exhaustiveDepsUselessWithoutDependencies"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := exhaustiveDepsWithOptions(t, testCase.optionsJson, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// experimental_autoDependenciesHooks, from React's three rows: a literal `null` array is no array to
// check, the absent array is not the setState loop, and a real array is still checked.
func TestExhaustiveDepsAutoDependenciesHooks(t *testing.T) {
	t.Parallel()

	const auto = `{"additionalHooks": "useSpecialEffect", "experimental_autoDependenciesHooks": ["useSpecialEffect"]}`
	const additionalOnly = `{"additionalHooks": "useSpecialEffect"}`
	const setStateWithoutArray = "function MyComponent() { const [state, setState] = React.useState<number>(0); " +
		"useSpecialEffect(() => { const someNumber: typeof state = 2; setState(prevState => prevState + someNumber); }) }"
	cases := []struct {
		name        string
		optionsJson string
		sourceText  string
		wantIds     []string
	}{
		{"a null array behaves like no array", auto,
			"function MyComponent(props) { useSpecialEffect(() => { console.log(props.foo); }, null); }",
			nil},
		{"the absent array is not the setState loop", auto, setStateWithoutArray, nil},
		{"the same Hook not named auto is the setState loop", additionalOnly, setStateWithoutArray,
			[]string{"exhaustiveDepsSetStateNoDependencies"}},
		{"a real array is still checked", auto,
			"function MyComponent() { const [state, setState] = React.useState<number>(0); " +
				"useSpecialEffect(() => { const someNumber: typeof state = 2; setState(prevState => prevState + someNumber + state); }, []) }",
			[]string{"exhaustiveDepsMissing"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := exhaustiveDepsWithOptions(t, testCase.optionsJson, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// enableDangerousAutofixThisMayCauseInfiniteLoops, at each of the five places React offers a repair,
// with the source React's own fixer produces, measured through ESLint's verifyAndFix on the installed
// build. Without the option none of them is applied, which the last row pins.
func TestExhaustiveDepsDangerousAutofix(t *testing.T) {
	t.Parallel()

	const dangerous = `{"enableDangerousAutofixThisMayCauseInfiniteLoops": true}`
	cases := []struct {
		name       string
		sourceText string
		wantId     string
		wantSource string
	}{
		{"React's row: the corrected array",
			"function MyComponent() { const local = {}; useEffect(() => { console.log(local); }, []); }",
			"exhaustiveDepsMissing",
			"function MyComponent() { const local = {}; useEffect(() => { console.log(local); }, [local]); }"},
		{"a callback that is not a function written here",
			"function MyComponent() { const myEffect = debounce(() => {}, 1); useEffect(myEffect, []); }",
			"exhaustiveDepsMissing",
			"function MyComponent() { const myEffect = debounce(() => {}, 1); useEffect(myEffect, [myEffect]); }"},
		{"an array added to an effect that sets state",
			"function MyComponent() { const [data, setData] = useState(0); useEffect(() => { setData(1); }); }",
			"exhaustiveDepsSetStateNoDependencies",
			"function MyComponent() { const [data, setData] = useState(0); useEffect(() => { setData(1); }, []); }"},
		{"a useEffectEvent function removed from the array",
			"function MyComponent() { const onEvent = useEffectEvent(() => {}); useEffect(() => { onEvent(); }, [onEvent]); }",
			"exhaustiveDepsEffectEvent",
			"function MyComponent() { const onEvent = useEffectEvent(() => {}); useEffect(() => { onEvent(); }, []); }"},
		{"a function used outside the Hook wrapped in useCallback",
			"function MyComponent() { const handle = () => {}; useEffect(() => { handle(); }, [handle]); return handle; }",
			"exhaustiveDepsConstruction",
			"function MyComponent() { const handle = useCallback(() => {}); useEffect(() => { handle(); }, [handle]); return handle; }"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := exhaustiveDepsWithOptions(t, dangerous, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource+"\n")

			// The control: without the option the finding stands and carries no fix.
			unconfigured := exhaustiveDepsWithOptions(t, `{}`, testCase.sourceText)
			rule_testing.ExpectFindings(t, unconfigured, testCase.wantId)
			for _, diagnostic := range unconfigured.Diagnostics {
				if len(diagnostic.Fixes) > 0 {
					t.Errorf("without the option %s carries a fix", diagnostic.Message.Id)
				}
			}
		})
	}
}
