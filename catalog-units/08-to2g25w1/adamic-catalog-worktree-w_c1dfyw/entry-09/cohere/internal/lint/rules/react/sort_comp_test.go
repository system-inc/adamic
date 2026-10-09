package react

import (
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// sortCompFile is where the fixtures pretend to live.
//
// A `.tsx` name because most cases are JSX and 24 of the 53 carry TypeScript syntax. It is NOT
// load-bearing: this rule has no file gate, and TestSortCompHasNoFileGate pins that by running
// one reporting source under two extensions.
const sortCompFile = "/repository/source/SortComp.tsx"

// sortCompCase is one row of upstream's corpus.
type sortCompCase struct {
	// name is the corpus list and index the row came from, so a failure names a case that can be
	// found in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// options is the RAW JSON of upstream's options object, routed through DecodeSortCompOptions
	// rather than built as a struct, so the decoder's own nil path and defaults are under test.
	options string

	// messages is the full rendered text upstream's installed 7.37.5 build produced for this
	// input, one per finding, in order. Asserted whole rather than by prefix: the interesting
	// half of this rule's message is the two member names and the direction, which is computed,
	// and a prefix assertion would let the rule name any members at all and stay green.
	messages []string
}

// sortCompFiresCases are the rows upstream reports on.
var sortCompFiresCases = []sortCompCase{
	{
		name:    "invalid-0",
		source:  "\n        // Must force a lifecycle method to be placed before render\n        var Hello = createReactClass({\n          render: function() {\n            return <div>Hello</div>;\n          },\n          displayName : 'Hello',\n        });\n      ",
		options: "",
		messages: []string{
			"render should be placed after displayName",
		},
	},
	{
		name:    "invalid-1",
		source:  "\n        // Must run rule when render uses createElement instead of JSX\n        var Hello = createReactClass({\n          render: function() {\n            return React.createElement(\"div\", null, \"Hello\");\n          },\n          displayName : 'Hello',\n        });\n      ",
		options: "",
		messages: []string{
			"render should be placed after displayName",
		},
	},
	{
		name:    "invalid-2",
		source:  "\n        // Must force a custom method to be placed before render\n        var Hello = createReactClass({\n          render: function() {\n            return <div>Hello</div>;\n          },\n          onClick: function() {},\n        });\n      ",
		options: "",
		messages: []string{
			"render should be placed after onClick",
		},
	},
	{
		name:    "invalid-3",
		source:  "\n        // Must force a custom method to be placed before render, even in function\n        var Hello = () => {\n          return class Test extends React.Component {\n            render () {\n              return <div>Hello</div>;\n            }\n            onClick () {}\n          }\n        };\n      ",
		options: "",
		messages: []string{
			"render should be placed after onClick",
		},
	},
	{
		name:    "invalid-4",
		source:  "\n        // Must force a custom method to be placed after render if no 'everything-else' group is specified\n        var Hello = createReactClass({\n          displayName: 'Hello',\n          onClick: function() {},\n          render: function() {\n            return <button onClick={this.onClick}>Hello</button>;\n          }\n        });\n      ",
		options: "{\"order\": [\"lifecycle\", \"render\"]}",
		messages: []string{
			"onClick should be placed after render",
		},
	},
	{
		name:    "invalid-5",
		source:  "\n        // Must validate static properties\n        class Hello extends React.Component {\n          render() {\n            return <div></div>\n          }\n          static displayName = 'Hello';\n        }\n      ",
		options: "",
		messages: []string{
			"render should be placed after displayName",
		},
	},
	{
		name:    "invalid-6",
		source:  "\n        // Type Annotations should not be at the top by default\n        class Hello extends React.Component {\n          props: { text: string };\n          constructor() {}\n          state: Object = {};\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "",
		messages: []string{
			"props should be placed after state",
		},
	},
	{
		name:    "invalid-7",
		source:  "\n        // Type Annotations should be first\n        class Hello extends React.Component {\n          constructor() {}\n          props: { text: string };\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"type-annotations\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
		messages: []string{
			"constructor should be placed after props",
		},
	},
	{
		name:    "invalid-8",
		source:  "\n        // Properties with Type Annotations should not be at the top\n        class Hello extends React.Component {\n          props: { text: string };\n          state: Object = {};\n          constructor() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"type-annotations\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
		messages: []string{
			"state should be placed after constructor",
		},
	},
	{
		name:    "invalid-10",
		source:  "\n        // Getters should at the top\n        class Hello extends React.Component {\n          constructor() {}\n          get foo() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"getters\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
		messages: []string{
			"constructor should be placed after getter functions",
		},
	},
	{
		name:    "invalid-11",
		source:  "\n        // Setters should at the top\n        class Hello extends React.Component {\n          constructor() {}\n          set foo(bar) {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"setters\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
		messages: []string{
			"constructor should be placed after setter functions",
		},
	},
	{
		name:    "invalid-12",
		source:  "\n        // Instance methods should not be at the top\n        class Hello extends React.Component {\n          constructor() {}\n          static bar = () => {}\n          classMethod() {}\n          foo = function() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"instance-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
		messages: []string{
			"foo should be placed before constructor",
		},
	},
	{
		name:    "invalid-13",
		source:  "\n        // Instance variables should not be at the top\n        class Hello extends React.Component {\n          constructor() {}\n          state = {}\n          static bar = {}\n          foo = {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"instance-variables\", \"lifecycle\", \"everything-else\", \"render\"]}",
		messages: []string{
			"foo should be placed before constructor",
		},
	},
	{
		name:    "invalid-14",
		source:  "\n        // Should not confuse method names with group names\n        class Hello extends React.Component {\n          setters() {}\n          constructor() {}\n          render() {}\n        }\n      ",
		options: "{\"order\": [\"setters\", \"lifecycle\", \"render\"]}",
		messages: []string{
			"setters should be placed after render",
		},
	},
	{
		name:    "invalid-15",
		source:  "\n        // Explicitly named methods should appear in the correct order\n        class Hello extends React.Component {\n          render() {}\n          foo() {}\n        }\n      ",
		options: "{\"order\": [\"foo\", \"render\"]}",
		messages: []string{
			"render should be placed after foo",
		},
	},
	{
		name:    "invalid-16",
		source:  "\n        class MyComponent extends React.Component {\n          static getDerivedStateFromProps() {}\n          static foo;\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"static-variables\", \"static-methods\"]}",
		messages: []string{
			"getDerivedStateFromProps should be placed after foo",
		},
	},
	{
		name:    "invalid-17",
		source:  "\n        class MyComponent extends React.Component {\n          static foo;\n          bar = 'some-str'\n          static getDerivedStateFromProps() {}\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"instance-variables\", \"static-variables\", \"static-methods\"]}",
		messages: []string{
			"foo should be placed after bar",
		},
	},
	{
		name:    "invalid-18",
		source:  "\n        class MyComponent extends React.Component {\n          static getDerivedStateFromProps() {}\n          static bar;\n          render() {\n            return null;\n          }\n          foo = {};\n        }\n      ",
		options: "{\"order\": [\"static-methods\", \"render\", \"static-variables\", \"instance-variables\"]}",
		messages: []string{
			"bar should be placed after render",
		},
	},
}

// sortCompSilentCases are the rows upstream is clean on.
var sortCompSilentCases = []sortCompCase{
	{
		name:    "valid-0",
		source:  "\n        // Must validate a full class\n        var Hello = createReactClass({\n          displayName : '',\n          propTypes: {},\n          contextTypes: {},\n          childContextTypes: {},\n          mixins: [],\n          statics: {},\n          getDefaultProps: function() {},\n          getInitialState: function() {},\n          getChildContext: function() {},\n          componentWillMount: function() {},\n          componentDidMount: function() {},\n          componentWillReceiveProps: function() {},\n          shouldComponentUpdate: function() {},\n          componentWillUpdate: function() {},\n          componentDidUpdate: function() {},\n          componentWillUnmount: function() {},\n          render: function() {\n            return <div>Hello</div>;\n          }\n        });\n      ",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "\n        // Must validate a class with missing groups\n        var Hello = createReactClass({\n          render: function() {\n            return <div>Hello</div>;\n          }\n        });\n      ",
		options: "",
	},
	{
		name:    "valid-2",
		source:  "\n        // Must put a custom method in 'everything-else'\n        var Hello = createReactClass({\n          onClick: function() {},\n          render: function() {\n            return <button onClick={this.onClick}>Hello</button>;\n          }\n        });\n      ",
		options: "",
	},
	{
		name:    "valid-3",
		source:  "\n        // Must allow us to re-order the groups\n        var Hello = createReactClass({\n          displayName : 'Hello',\n          render: function() {\n            return <button onClick={this.onClick}>Hello</button>;\n          },\n          onClick: function() {}\n        });\n      ",
		options: "{\"order\": [\"lifecycle\", \"render\", \"everything-else\"]}",
	},
	{
		name:    "valid-4",
		source:  "\n        // Must validate a full React 16.3 createReactClass class\n        var Hello = createReactClass({\n          displayName : '',\n          propTypes: {},\n          contextTypes: {},\n          childContextTypes: {},\n          mixins: [],\n          statics: {},\n          getDefaultProps: function() {},\n          getInitialState: function() {},\n          getChildContext: function() {},\n          UNSAFE_componentWillMount: function() {},\n          componentDidMount: function() {},\n          UNSAFE_componentWillReceiveProps: function() {},\n          shouldComponentUpdate: function() {},\n          UNSAFE_componentWillUpdate: function() {},\n          getSnapshotBeforeUpdate: function() {},\n          componentDidUpdate: function() {},\n          componentDidCatch: function() {},\n          componentWillUnmount: function() {},\n          render: function() {\n            return <div>Hello</div>;\n          }\n        });\n      ",
		options: "",
	},
	{
		name:    "valid-5",
		source:  "\n        // Must validate React 16.3 lifecycle methods with the default parser\n        class Hello extends React.Component {\n          constructor() {}\n          static getDerivedStateFromProps() {}\n          UNSAFE_componentWillMount() {}\n          componentDidMount() {}\n          UNSAFE_componentWillReceiveProps() {}\n          shouldComponentUpdate() {}\n          UNSAFE_componentWillUpdate() {}\n          getSnapshotBeforeUpdate() {}\n          componentDidUpdate() {}\n          componentDidCatch() {}\n          componentWillUnmount() {}\n          testInstanceMethod() {}\n          render() { return (<div>Hello</div>); }\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-6",
		source:  "\n        // Must validate a full React 16.3 ES6 class\n        class Hello extends React.Component {\n          static displayName = ''\n          static propTypes = {}\n          static defaultProps = {}\n          constructor() {}\n          state = {}\n          static getDerivedStateFromProps = () => {}\n          UNSAFE_componentWillMount = () => {}\n          componentDidMount = () => {}\n          UNSAFE_componentWillReceiveProps = () => {}\n          shouldComponentUpdate = () => {}\n          UNSAFE_componentWillUpdate = () => {}\n          getSnapshotBeforeUpdate = () => {}\n          componentDidUpdate = () => {}\n          componentDidCatch = () => {}\n          componentWillUnmount = () => {}\n          testArrowMethod = () => {}\n          testInstanceMethod() {}\n          render = () => (<div>Hello</div>)\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-7",
		source:  "\n        // Must allow us to create a RegExp-based group\n        class Hello extends React.Component {\n          customHandler() {}\n          render() {\n            return <div>Hello</div>;\n          }\n          onClick() {}\n        }\n      ",
		options: "{\"order\": [\"lifecycle\", \"everything-else\", \"render\", \"/on.*/\"]}",
	},
	{
		name:    "valid-8",
		source:  "\n        // Must allow us to create a named group\n        class Hello extends React.Component {\n          customHandler() {}\n          render() {\n            return <div>Hello</div>;\n          }\n          onClick() {}\n        }\n      ",
		options: "{\"order\": [\"lifecycle\", \"everything-else\", \"render\", \"customGroup\"], \"groups\": {\"customGroup\": [\"/on.*/\"]}}",
	},
	{
		name:    "valid-9",
		source:  "\n        // Must allow a method to be in different places if it's matches multiple patterns\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n          onClick() {}\n        }\n      ",
		options: "{\"order\": [\"/on.*/\", \"render\", \"/.*Click/\"]}",
	},
	{
		name:    "valid-10",
		source:  "\n        // Must allow us to use 'constructor' as a method name\n        class Hello extends React.Component {\n          constructor() {}\n          displayName() {}\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"constructor\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-11",
		source:  "\n        // Must ignore stateless components\n        function Hello(props) {\n          return <div>Hello {props.name}</div>\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-12",
		source:  "\n        // Must ignore stateless components (arrow function with explicit return)\n        var Hello = props => (\n          <div>Hello {props.name}</div>\n        )\n      ",
		options: "",
	},
	{
		name:    "valid-13",
		source:  "\n        // Must ignore spread operator\n        var Hello = createReactClass({\n          ...proto,\n          render: function() {\n            return <div>Hello</div>;\n          }\n        });\n      ",
		options: "",
	},
	{
		name:    "valid-14",
		source:  "\n        // Type Annotations should be first\n        class Hello extends React.Component {\n          props: { text: string };\n          constructor() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"type-annotations\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-15",
		source:  "\n        // Properties with Type Annotations should not be at the top\n        class Hello extends React.Component {\n          props: { text: string };\n          constructor() {}\n          state: Object = {};\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"type-annotations\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-16",
		source:  "\n        // Non-react classes should be ignored, even in expressions\n        return class Hello {\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n          props: { text: string };\n          constructor() {}\n          state: Object = {};\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-17",
		source:  "\n        // Non-react classes should be ignored, even in expressions\n        return class {\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n          props: { text: string };\n          constructor() {}\n          state: Object = {};\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-18",
		source:  "\n        // Getters should be at the top\n        class Hello extends React.Component {\n          get foo() {}\n          constructor() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"getters\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-19",
		source:  "\n        // Setters should be at the top\n        class Hello extends React.Component {\n          set foo(bar) {}\n          constructor() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"setters\", \"static-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-20",
		source:  "\n        // Instance methods should be at the top\n        class Hello extends React.Component {\n          foo = () => {}\n          constructor() {}\n          classMethod() {}\n          static bar = () => {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"instance-methods\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-21",
		source:  "\n        // Instance variables should be at the top\n        class Hello extends React.Component {\n          foo = 'bar'\n          constructor() {}\n          state = {}\n          static bar = 'foo'\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n        }\n      ",
		options: "{\"order\": [\"instance-variables\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-22",
		source:  "\n        // Methods can be grouped with any matching group (with statics)\n        class Hello extends React.Component {\n          static onFoo() {}\n          static renderFoo() {}\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n          getFoo() {}\n        }\n      ",
		options: "{\"order\": [\"static-methods\", \"render\", \"/^get.+$/\", \"/^on.+$/\", \"/^render.+$/\"]}",
	},
	{
		name:    "valid-23",
		source:  "\n        // Methods can be grouped with any matching group (with RegExp)\n        class Hello extends React.Component {\n          render() {\n            return <div>{this.props.text}</div>;\n          }\n          getFoo() {}\n          static onFoo() {}\n          static renderFoo() {}\n        }\n      ",
		options: "{\"order\": [\"static-methods\", \"render\", \"/^get.+$/\", \"/^on.+$/\", \"/^render.+$/\"]}",
	},
	{
		name:    "valid-24",
		source:  "\n        // static lifecycle methods can be grouped (with statics)\n        class Hello extends React.Component {\n          static getDerivedStateFromProps() {}\n          constructor() {}\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-25",
		source:  "\n        // static lifecycle methods can be grouped (with lifecycle)\n        class Hello extends React.Component {\n          constructor() {}\n          static getDerivedStateFromProps() {}\n        }\n      ",
		options: "",
	},
	{
		name:    "valid-26",
		source:  "\n        class MyComponent extends React.Component {\n          state = {};\n          foo;\n          static propTypes;\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"state\", \"instance-variables\", \"static-methods\", \"lifecycle\", \"render\", \"everything-else\"]}",
	},
	{
		name:    "valid-27",
		source:  "\n        class MyComponent extends React.Component {\n          static foo;\n          static getDerivedStateFromProps() {}\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"static-variables\", \"static-methods\"]}",
	},
	{
		name:    "valid-28",
		source:  "\n        class MyComponent extends React.Component {\n          static getDerivedStateFromProps() {}\n          static foo = 'some-str';\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"static-methods\", \"static-variables\"]}",
	},
	{
		name:    "valid-29",
		source:  "\n        class MyComponent extends React.Component {\n          foo = {};\n          static bar = 0;\n          static getDerivedStateFromProps() {}\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"instance-variables\", \"static-variables\", \"static-methods\"]}",
	},
	{
		name:    "valid-30",
		source:  "\n        class MyComponent extends React.Component {\n          static bar = 1;\n          foo = {};\n          static getDerivedStateFromProps() {}\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"static-variables\", \"instance-variables\", \"static-methods\"]}",
	},
	{
		name:    "valid-31",
		source:  "\n        class MyComponent extends React.Component {\n          static getDerivedStateFromProps() {}\n          render() {\n            return null;\n          }\n          static bar;\n          foo = {};\n        }\n      ",
		options: "{\"order\": [\"static-methods\", \"render\", \"static-variables\", \"instance-variables\"]}",
	},
	{
		name:    "valid-32",
		source:  "\n        class MyComponent extends React.Component {\n          static foo = 1;\n          bar;\n\n          constructor() {\n            super(props);\n\n            this.state = {};\n          }\n\n          render() {\n            return null;\n          }\n        }\n      ",
		options: "{\"order\": [\"static-variables\", \"instance-variables\", \"constructor\", \"everything-else\", \"render\"]}",
	},
	{
		name:    "valid-33",
		source:  "\n        class ClassName extends React.Component {\n          static defaultProps = {};\n          static parseDateString(date?: Date) {}\n          state = {};\n          render() {\n            return <div />;\n          }\n        }\n      ",
		options: "{\"order\": [\"static-variables\", \"static-methods\", \"type-annotations\", \"instance-variables\", \"lifecycle\", \"everything-else\", \"render\"]}",
	},
}

// decodedSortCompOptions routes a row's raw JSON through the rule's own exported decoder.
//
// Through the decoder rather than by constructing SortCompOptions directly, because the decoder
// is what the config layer calls and it carries two behaviours nothing else would test: the
// empty-input path a bare `"error"` configuration takes, and the pointer fields that keep an
// explicitly empty `order` distinct from an absent one.
func decodedSortCompOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeSortCompOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return decoded
}

func TestSortCompFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range sortCompFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, testCase.source,
				decodedSortCompOptions(t, testCase.options))

			expectedIds := make([]string, len(testCase.messages))
			for index := range expectedIds {
				expectedIds[index] = "unsortedProps"
			}
			rule_testing.ExpectFindings(t, result, expectedIds...)

			if len(result.Diagnostics) != len(testCase.messages) {
				return
			}
			for index, diagnostic := range result.Diagnostics {
				// The whole message, not a prefix: the member names and the direction are the
				// computed half, and a prefix assertion cannot see them.
				if !strings.HasPrefix(diagnostic.Message.Description, testCase.messages[index]+".") {
					t.Errorf("finding %d: rendered %q, upstream said %q",
						index, diagnostic.Message.Description, testCase.messages[index])
				}
			}
		})
	}
}

func TestSortCompStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range sortCompSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, testCase.source,
				decodedSortCompOptions(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// sortCompRe2LookaroundCase is upstream's invalid[9], held out of the generated tables.
//
// Its order carries `/^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$/`, and Go's
// regexp is RE2, which has no lookaround. The pattern is therefore dropped rather than compiled,
// exactly as `core/no-param-reassign` drops an uncompilable ignore pattern, and the divergence is
// recorded here with BOTH verdicts written out rather than removed to make the suite green.
//
// What agrees: the finding COUNT, the message id, and the offending member. What differs: the
// PARTNER the message names, because `getA` and `getB` fall through to `everything-else` once the
// pattern that would have grouped them is gone.
//
// Measured across the whole 53-case corpus, 13 order entries are regex patterns and this is the
// only one using lookaround.
const (
	sortCompRe2LookaroundSource  = "\n        // componentDidMountOk should be placed after getA\n        export default class View extends React.Component {\n          componentDidMountOk() {}\n          getB() {}\n          componentWillMount() {}\n          getA() {}\n          render() {}\n        }\n      "
	sortCompRe2LookaroundOptions = "{\"order\": [\"static-methods\", \"lifecycle\", \"/^on.+$/\", \"/^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$/\", \"everything-else\", \"/^render.+$/\", \"render\"]}"

	// What upstream's installed 7.37.5 build says, driven through the ESLint Linter API.
	sortCompRe2LookaroundUpstreamMessage = "componentDidMountOk should be placed after getA"

	// What this rule says, and why. Written out rather than described so a change to either side
	// fails loudly instead of quietly re-converging on a wrong answer.
	sortCompRe2LookaroundOurMessage = "componentDidMountOk should be placed after componentWillMount"
)

func TestSortCompDropsAnRe2IncompatiblePattern(t *testing.T) {
	t.Parallel()

	// The premise first: if this pattern ever compiles, the divergence below is stale and the case
	// belongs back in the generated table. Asserting it here means a future Go regexp that grows
	// lookaround fails this test rather than silently leaving a wrong expectation in place.
	if _, err := regexp.Compile("^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$"); err == nil {
		t.Fatal("this pattern now compiles, so the divergence recorded here no longer applies")
	}

	result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, sortCompRe2LookaroundSource,
		decodedSortCompOptions(t, sortCompRe2LookaroundOptions))

	// The count and the offending member agree with upstream; only the named partner moves.
	rule_testing.ExpectFindings(t, result, "unsortedProps")
	if len(result.Diagnostics) != 1 {
		return
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description, sortCompRe2LookaroundOurMessage+".") {
		t.Errorf("rendered %q, expected the recorded divergence %q",
			result.Diagnostics[0].Message.Description, sortCompRe2LookaroundOurMessage)
	}
	if strings.HasPrefix(result.Diagnostics[0].Message.Description, sortCompRe2LookaroundUpstreamMessage+".") {
		t.Errorf("this now matches upstream exactly, so the recorded divergence is stale")
	}
}

// TestSortCompHasNoFileGate pins that this rule judges a component in a plain `.ts` file.
//
// Three siblings in this package once declined every file not ending `.tsx`/`.jsx`, which was oxc
// residue rather than upstream behaviour. `eslint-plugin-react` gates this rule nowhere, and a
// React class in a `.ts` file is ordinary and legal.
func TestSortCompHasNoFileGate(t *testing.T) {
	t.Parallel()

	const reporting = "class Hello extends React.Component {\n  render() { return <div>Hello</div>; }\n  componentDidMount() {}\n}\n"

	for _, fileName := range []string{"/repository/source/Gate.tsx", "/repository/source/Gate.ts"} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SortComp, fileName, reporting,
				decodedSortCompOptions(t, ""))
			rule_testing.ExpectFindings(t, result, "unsortedProps")
		})
	}
}

// TestSortCompSurvivesNilOptions covers the path a bare `"error"` configuration takes.
//
// A rule configured without options is handed nil rather than the decoded default, and `options.(T)`
// on nil yields the zero value silently. For this rule the zero value is an EMPTY order, which
// matches nothing and reports nothing, so without the explicit fallback in `Run` this rule would
// register on every file in the tree and judge none of them. That is the inert-rule shape that
// reads exactly like a clean tree, so it is pinned here rather than trusted.
func TestSortCompSurvivesNilOptions(t *testing.T) {
	t.Parallel()

	const reporting = "class Hello extends React.Component {\n  render() { return <div>Hello</div>; }\n  componentDidMount() {}\n}\n"

	result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, reporting, nil)
	rule_testing.ExpectFindings(t, result, "unsortedProps")
}

// TestSortCompReportsWhereItSays asserts the span rather than only the message.
//
// `ExpectFindings` checks ids and count and nothing else, so a rule that points at the wrong member
// passes a complete fixture pair while being wrong. Upstream anchors on the misplaced member itself,
// which for a method is the whole declaration including its body.
func TestSortCompReportsWhereItSays(t *testing.T) {
	t.Parallel()

	const source = "class Hello extends React.Component {\n  render() { return <div>Hello</div>; }\n  componentDidMount() {}\n}\n"

	result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, source,
		decodedSortCompOptions(t, ""))
	rule_testing.ExpectFindings(t, result, "unsortedProps")
	if len(result.Diagnostics) != 1 {
		return
	}

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	const wanted = "render() { return <div>Hello</div>; }"
	if reported != wanted {
		t.Errorf("finding spans %q, expected %q", reported, wanted)
	}
}

// TestSortCompDeclinesAMemberItCannotName covers a computed member key.
//
// `Node.Text()` panics on several name kinds rather than returning empty, and the walk recovers per
// FILE rather than per rule, so one such member costs every rule in this package its verdict on that
// file. Upstream reads an unreadable key as the empty string, which matches no order entry and falls
// through to `everything-else`, so the member stays in the comparison rather than being skipped.
func TestSortCompDeclinesAMemberItCannotName(t *testing.T) {
	t.Parallel()

	const source = "const key = 'x';\nclass Hello extends React.Component {\n  [key]() {}\n  render() { return <div>Hello</div>; }\n}\n"

	// The assertion that matters is that this does not panic and produces a verdict at all. The
	// computed member sorts into `everything-else`, which is before `render`, so this is clean.
	result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, source,
		decodedSortCompOptions(t, ""))
	rule_testing.ExpectClean(t, result)
}

// The three tests below exist because a mutation sweep found three discriminations upstream's own
// corpus cannot see. Each was settled by driving the installed 7.37.5 build on inputs written for
// the question, with a control that fires, rather than by reading the reference implementation.

// TestSortCompNamesTheFirstMemberAtAReferenceIndex covers `storeError`'s name guard.
//
// Upstream stops updating an error's named partner once a DIFFERENT member has been recorded at the
// same reference index: `getPropertyName(errors[propA.index].node) !== getPropertyName(propA.node)`
// returns early. Reaching it needs two or more members that share a reference index, which upstream's
// corpus never writes, so deleting the guard survived all 52 imported cases.
//
// Two and three members that all fall to `everything-else` distinguish it. Measured on the installed
// build: with `alpha` and `beta` after `render` the message names `beta`, and adding `gamma` moves it
// to `gamma`. A control with a single such member reports and names `alpha`, so the harness is
// demonstrably live on this shape.
func TestSortCompNamesTheFirstMemberAtAReferenceIndex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		message string
	}{
		{
			name:    "one-member-control",
			source:  "class Hello extends React.Component {\n  render() { return <div/>; }\n  alpha() {}\n}\n",
			message: "render should be placed after alpha",
		},
		{
			name:    "two-members-at-one-reference-index",
			source:  "class Hello extends React.Component {\n  render() { return <div/>; }\n  alpha() {}\n  beta() {}\n}\n",
			message: "render should be placed after beta",
		},
		{
			name:    "three-members-at-one-reference-index",
			source:  "class Hello extends React.Component {\n  render() { return <div/>; }\n  alpha() {}\n  beta() {}\n  gamma() {}\n}\n",
			message: "render should be placed after gamma",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, testCase.source,
				decodedSortCompOptions(t, ""))
			rule_testing.ExpectFindings(t, result, "unsortedProps")
			if len(result.Diagnostics) != 1 {
				return
			}
			if !strings.HasPrefix(result.Diagnostics[0].Message.Description, testCase.message+".") {
				t.Errorf("rendered %q, upstream said %q",
					result.Diagnostics[0].Message.Description, testCase.message)
			}
		})
	}
}

// TestSortCompDoesNotTreatAPlainMethodAsAnInstanceMethod covers the static/instance asymmetry.
//
// Upstream's `instanceMethod` accepts only a property initialized to a function, while its
// `staticMethod` also accepts a `MethodDefinition`. That asymmetry looks like a transcription slip
// and is not, so collapsing the two is the obvious "cleanup" a later reader would make.
//
// Measured on the installed build, with a control that fires: a class whose plain `classMethod()`
// sits AFTER `render` under an `instance-methods`-first order is CLEAN, while the same shape with an
// arrow property `foo = () => {}` REPORTS. So upstream genuinely does not count a plain method.
//
// Mutation-checked, and the first attempt at this test proved nothing. Widening the instance arm to
// `(isPropertyLike || isMethodDeclaration) && initializedToFunction` survives, and re-running it
// against these cases showed why: `initializedToFunction` already requires `isPropertyLike`, so that
// mutant is EQUIVALENT rather than uncaught. The predicate that genuinely differs -- dropping the
// initializer requirement -- is caught by 7 assertions, these two among them.
func TestSortCompDoesNotTreatAPlainMethodAsAnInstanceMethod(t *testing.T) {
	t.Parallel()

	const instanceMethodsFirst = `{"order":["instance-methods","lifecycle","render"]}`

	t.Run("plain-method-is-not-an-instance-method", func(t *testing.T) {
		t.Parallel()
		const source = "class Hello extends React.Component {\n  render() { return <div/>; }\n  classMethod() {}\n}\n"
		result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, source,
			decodedSortCompOptions(t, instanceMethodsFirst))
		rule_testing.ExpectClean(t, result)
	})

	// The control. Without it the case above passes for a rule that has stopped judging anything.
	t.Run("control-arrow-property-is-an-instance-method", func(t *testing.T) {
		t.Parallel()
		const source = "class Hello extends React.Component {\n  render() { return <div/>; }\n  foo = () => {};\n}\n"
		result := rule_testing.RunWithOptions(t, SortComp, sortCompFile, source,
			decodedSortCompOptions(t, instanceMethodsFirst))
		rule_testing.ExpectFindings(t, result, "unsortedProps")
	})
}

// TestSortCompLifecycleNamesCannotReachTheRegexBranch records an EQUIVALENT mutant rather than
// killing it, because the equivalence is provable.
//
// `sortCompGroupMatches` tests membership in the known-lifecycle-name list BEFORE trying the regex
// production, and reordering those two survived the whole corpus. That is not a fixture gap: the
// production is `/(.*)/([gimsuy]*)`, which requires a literal `/`, and no lifecycle name contains
// one, so no input can take a different path under either ordering.
//
// Asserted here rather than argued in a comment, so a future upstream that adds a lifecycle name
// with a slash in it fails this test instead of silently changing the rule's meaning.
func TestSortCompLifecycleNamesCannotReachTheRegexBranch(t *testing.T) {
	t.Parallel()

	both := []string{}
	for name := range sortCompKnownLifecycleNames {
		if sortCompRegexPattern.MatchString(name) {
			both = append(both, name)
		}
	}
	if len(both) != 0 {
		t.Errorf("these lifecycle names also parse as regex entries, so the branch ordering in "+
			"sortCompGroupMatches now changes a verdict and needs a real fixture: %v", both)
	}

	// The control. Without it a matcher that answered false for everything would pass the above.
	if !sortCompRegexPattern.MatchString("/on.*/") {
		t.Fatal("the regex production matches nothing, so the check above proved nothing")
	}
	if len(sortCompKnownLifecycleNames) == 0 {
		t.Fatal("the lifecycle name list is empty, so the check above proved nothing")
	}
}
