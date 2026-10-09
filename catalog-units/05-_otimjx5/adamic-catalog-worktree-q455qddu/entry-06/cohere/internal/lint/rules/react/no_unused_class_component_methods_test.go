package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noUnusedClassComponentMethodsFile is where the fixtures pretend to live.
//
// A `.tsx` name because most cases carry JSX. There is no file gate and
// `TestNoUnusedClassComponentMethodsHasNoFileGate` reports on a `.ts` file too.
const noUnusedClassComponentMethodsFile = "/repository/source/NoUnusedClassComponentMethods.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-unused-class-component-methods.js` holds
// 36 valid and 22 invalid cases. Every string below was pulled out of that file by loading it with a
// stubbed `RuleTester` and serializing the captured object to JSON, then emitted into these tables
// by a generator. No case was typed by hand.
//
// All 58 were run against the installed build, 7.37.5, through the ESLint Linter API with the
// TypeScript parser, and the two authorities agreed on every one. The comparison was on the RENDERED
// message text rather than on message ids, because this corpus states raw text in its `errors`
// entries and both of this rule's ids interpolate; comparing ids would have compared nothing.
//
// # The invalid table asserts the rendered text as well as the id
//
// Both ids interpolate a name and one of them also interpolates the class, so an id assertion cannot
// see a message naming the wrong thing. The fourth column is the exact string the installed build
// rendered, captured in the same run that verified the ids, asserted as a PREFIX because this port
// appends a sentence of reasoning after upstream's wording.

// TestNoUnusedClassComponentMethodsFires runs upstream's reporting cases.
func TestNoUnusedClassComponentMethodsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
		wantTexts  []string
	}{
		{"upstream invalid-0", "\n        class Foo extends React.Component {\n          getDerivedStateFromProps() {}\n          render() {\n            return <div>Example</div>;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"getDerivedStateFromProps\" of class \"Foo\""}},
		{"upstream invalid-1", "\n        class Foo extends React.Component {\n          property = {}\n          render() {\n            return <div>Example</div>;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"property\" of class \"Foo\""}},
		{"upstream invalid-2", "\n        class Foo extends React.Component {\n          handleClick() {}\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"handleClick\" of class \"Foo\""}},
		{"upstream invalid-3", "\n        var Foo = createReactClass({\n          handleClick() {},\n          render() {\n            return null;\n          },\n        })\n      ", []string{"unused"}, []string{"Unused method or property \"handleClick\""}},
		{"upstream invalid-4", "\n        var Foo = createReactClass({\n          a: 3,\n          render() {\n            return null;\n          },\n        })\n      ", []string{"unused"}, []string{"Unused method or property \"a\""}},
		{"upstream invalid-5", "\n        class Foo extends React.Component {\n          handleScroll() {}\n          handleClick() {}\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass", "unusedWithClass"}, []string{"Unused method or property \"handleScroll\" of class \"Foo\"", "Unused method or property \"handleClick\" of class \"Foo\""}},
		{"upstream invalid-6", "\n        class Foo extends React.Component {\n          handleClick = () => {}\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"handleClick\" of class \"Foo\""}},
		{"upstream invalid-7", "\n        class Foo extends React.Component {\n          action = async () => {}\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"action\" of class \"Foo\""}},
		{"upstream invalid-8", "\n        class Foo extends React.Component {\n          async action() {\n            console.log('error');\n          }\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"action\" of class \"Foo\""}},
		{"upstream invalid-9", "\n        class Foo extends React.Component {\n          * action() {\n            console.log('error');\n          }\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"action\" of class \"Foo\""}},
		{"upstream invalid-10", "\n        class Foo extends React.Component {\n          async * action() {\n            console.log('error');\n          }\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"action\" of class \"Foo\""}},
		{"upstream invalid-11", "\n        class Foo extends React.Component {\n          getInitialState() {}\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"getInitialState\" of class \"Foo\""}},
		{"upstream invalid-12", "\n        class Foo extends React.Component {\n          action = function() {\n            console.log('error');\n          }\n          render() {\n            return null;\n          }\n        }\n      ", []string{"unusedWithClass"}, []string{"Unused method or property \"action\" of class \"Foo\""}},
		{"upstream invalid-13", "\n         class ClassAssignPropertyInMethodTest extends React.Component {\n           constructor() {\n             this.foo = 3;\n           }\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"ClassAssignPropertyInMethodTest\""}},
		{"upstream invalid-14", "\n         class Foo extends React.Component {\n           foo;\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-15", "\n         class Foo extends React.Component {\n           foo = a;\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-16", "\n         class Foo extends React.Component {\n           ['foo'];\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-17", "\n         class Foo extends React.Component {\n           ['foo'] = a;\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-18", "\n         class Foo extends React.Component {\n           foo = a;\n           render() {\n             return <SomeComponent foo={this[foo]} />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-19", "\n         class Foo extends React.Component {\n           private foo;\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-20", "\n         class Foo extends React.Component {\n           private foo() {}\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
		{"upstream invalid-21", "\n         class Foo extends React.Component {\n           private foo = 3;\n           render() {\n             return <SomeComponent />;\n           }\n         }\n       ", []string{"unusedWithClass"}, []string{"Unused method or property \"foo\" of class \"Foo\""}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoUnusedClassComponentMethods(t, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			for index, diagnostic := range result.Diagnostics {
				if !strings.HasPrefix(diagnostic.Message.Description, testCase.wantTexts[index]) {
					t.Errorf("finding %d reads %q, want it to begin %q",
						index, diagnostic.Message.Description, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestNoUnusedClassComponentMethodsStaysSilent runs upstream's clean cases.
func TestNoUnusedClassComponentMethodsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream valid-0", "\n        class SmockTestForTypeOfNullError extends React.Component {\n          handleClick() {}\n          foo;\n          render() {\n            let a;\n            return <button disabled onClick={this.handleClick} foo={this.foo}>Text</button>;\n          }\n        }\n      "},
		{"upstream valid-1", "\n        class Foo extends React.Component {\n          handleClick() {}\n          render() {\n            return <button onClick={this.handleClick}>Text</button>;\n          }\n        }\n      "},
		{"upstream valid-2", "\n        var Foo = createReactClass({\n          handleClick() {},\n          render() {\n            return <button onClick={this.handleClick}>Text</button>;\n          },\n        })\n      "},
		{"upstream valid-3", "\n        class Foo extends React.Component {\n          action() {}\n          componentDidMount() {\n            this.action();\n          }\n          render() {\n            return null;\n          }\n        }\n      "},
		{"upstream valid-4", "\n        var Foo = createReactClass({\n          action() {},\n          componentDidMount() {\n            this.action();\n          },\n          render() {\n            return null;\n          },\n        })\n      "},
		{"upstream valid-5", "\n        class Foo extends React.Component {\n          action() {}\n          componentDidMount() {\n            const action = this.action;\n            action();\n          }\n          render() {\n            return null;\n          }\n        }\n      "},
		{"upstream valid-6", "\n        class Foo extends React.Component {\n          getValue() {}\n          componentDidMount() {\n            const action = this.getValue();\n          }\n          render() {\n            return null;\n          }\n        }\n      "},
		{"upstream valid-7", "\n        class Foo extends React.Component {\n          handleClick = () => {}\n          render() {\n            return <button onClick={this.handleClick}>Button</button>;\n          }\n        }\n      "},
		{"upstream valid-8", "\n        class Foo extends React.Component {\n          renderContent() {}\n          render() {\n            return <div>{this.renderContent()}</div>;\n          }\n        }\n      "},
		{"upstream valid-9", "\n        class Foo extends React.Component {\n          renderContent() {}\n          render() {\n            return (\n              <div>\n                <div>{this.renderContent()}</div>;\n              </div>\n            );\n          }\n        }\n      "},
		{"upstream valid-10", "\n        class Foo extends React.Component {\n          property = {}\n          render() {\n            return <div property={this.property}>Example</div>;\n          }\n        }\n      "},
		{"upstream valid-11", "\n        class Foo extends React.Component {\n          action = () => {}\n          anotherAction = () => {\n            this.action();\n          }\n          render() {\n            return <button onClick={this.anotherAction}>Example</button>;\n          }\n        }\n      "},
		{"upstream valid-12", "\n        class Foo extends React.Component {\n          action = () => {}\n          anotherAction = () => this.action()\n          render() {\n            return <button onClick={this.anotherAction}>Example</button>;\n          }\n        }\n      "},
		{"upstream valid-13", "\n        class Foo extends React.Component {\n          getValue = () => {}\n          value = this.getValue()\n          render() {\n            return this.value;\n          }\n        }\n      "},
		{"upstream valid-14", "\n        class Foo {\n          action = () => {}\n          anotherAction = () => this.action()\n        }\n      "},
		{"upstream valid-15", "\n        class Foo extends React.Component {\n          action = async () => {}\n          render() {\n            return <button onClick={this.action}>Click</button>;\n          }\n        }\n      "},
		{"upstream valid-16", "\n        class Foo extends React.Component {\n          async action() {\n            console.log('error');\n          }\n          render() {\n            return <button onClick={() => this.action()}>Click</button>;\n          }\n        }\n      "},
		{"upstream valid-17", "\n        class Foo extends React.Component {\n          * action() {\n            console.log('error');\n          }\n          render() {\n            return <button onClick={() => this.action()}>Click</button>;\n          }\n        }\n      "},
		{"upstream valid-18", "\n        class Foo extends React.Component {\n          async * action() {\n            console.log('error');\n          }\n          render() {\n            return <button onClick={() => this.action()}>Click</button>;\n          }\n        }\n      "},
		{"upstream valid-19", "\n        class Foo extends React.Component {\n          action = function() {\n            console.log('error');\n          }\n          render() {\n            return <button onClick={() => this.action()}>Click</button>;\n          }\n        }\n      "},
		{"upstream valid-20", "\n        class ClassAssignPropertyInMethodTest extends React.Component {\n          constructor() {\n            this.foo = 3;;\n          }\n          render() {\n            return <SomeComponent foo={this.foo} />;\n          }\n        }\n      "},
		{"upstream valid-21", "\n        class ClassPropertyTest extends React.Component {\n          foo;\n          render() {\n            return <SomeComponent foo={this.foo} />;\n          }\n        }\n      "},
		{"upstream valid-22", "\n        class ClassPropertyTest extends React.Component {\n          foo = a;\n          render() {\n            return <SomeComponent foo={this.foo} />;\n          }\n        }\n      "},
		{"upstream valid-23", "\n        class Foo extends React.Component {\n          ['foo'] = a;\n          render() {\n            return <SomeComponent foo={this['foo']} />;\n          }\n        }\n      "},
		{"upstream valid-24", "\n        class Foo extends React.Component {\n          ['foo'];\n          render() {\n            return <SomeComponent foo={this['foo']} />;\n          }\n        }\n      "},
		{"upstream valid-25", "\n        class ClassComputedTemplatePropertyTest extends React.Component {\n          [`foo`] = a;\n          render() {\n            return <SomeComponent foo={this[`foo`]} />;\n          }\n        }\n      "},
		{"upstream valid-26", "\n        class ClassComputedTemplatePropertyTest extends React.Component {\n          state = {}\n          render() {\n            return <div />;\n          }\n        }\n      "},
		{"upstream valid-27", "\n        class ClassLiteralComputedMemberTest extends React.Component {\n          ['foo']() {}\n          render() {\n            return <SomeComponent foo={this.foo} />;\n          }\n        }\n      "},
		{"upstream valid-28", "\n        class ClassComputedTemplateMemberTest extends React.Component {\n          [`foo`]() {}\n          render() {\n            return <SomeComponent foo={this.foo} />;\n          }\n        }\n      "},
		{"upstream valid-29", "\n        class ClassUseAssignTest extends React.Component {\n          foo() {}\n          render() {\n            this.foo;\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid-30", "\n        class ClassUseAssignTest extends React.Component {\n          foo() {}\n          render() {\n            const { foo } = this;\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid-31", "\n        class ClassUseDestructuringTest extends React.Component {\n          foo() {}\n          render() {\n            const { foo } = this;\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid-32", "\n        class ClassUseDestructuringTest extends React.Component {\n          ['foo']() {}\n          render() {\n            const { 'foo': bar } = this;\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid-33", "\n        class ClassComputedMemberTest extends React.Component {\n          [foo]() {}\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid-34", "\n        class ClassWithLifecyleTest extends React.Component {\n          constructor(props) {\n            super(props);\n          }\n          static getDerivedStateFromProps() {}\n          componentWillMount() {}\n          UNSAFE_componentWillMount() {}\n          componentDidMount() {}\n          componentWillReceiveProps() {}\n          UNSAFE_componentWillReceiveProps() {}\n          shouldComponentUpdate() {}\n          componentWillUpdate() {}\n          UNSAFE_componentWillUpdate() {}\n          static getSnapshotBeforeUpdate() {}\n          componentDidUpdate() {}\n          componentDidCatch() {}\n          componentWillUnmount() {}\n          getChildContext() {}\n          render() {\n            return <SomeComponent />;\n          }\n        }\n      "},
		{"upstream valid-35", "\n        var ClassWithLifecyleTest = createReactClass({\n          mixins: [],\n          constructor(props) {\n          },\n          getDefaultProps() {\n            return {}\n          },\n          getInitialState: function() {\n            return {x: 0};\n          },\n          componentWillMount() {},\n          UNSAFE_componentWillMount() {},\n          componentDidMount() {},\n          componentWillReceiveProps() {},\n          UNSAFE_componentWillReceiveProps() {},\n          shouldComponentUpdate() {},\n          componentWillUpdate() {},\n          UNSAFE_componentWillUpdate() {},\n          componentDidUpdate() {},\n          componentDidCatch() {},\n          componentWillUnmount() {},\n          getChildContext() {},\n          render() {\n            return <SomeComponent />;\n          },\n        })\n      "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoUnusedClassComponentMethods(t, testCase.sourceText))
		})
	}
}

// runNoUnusedClassComponentMethods drives the rule on an untyped program.
//
// `Run` rather than `RunTyped`, because every question this rule asks is on the AST: what a class
// extends, what members it declares, and which `this` accesses appear inside them.
// `TestNoUnusedClassComponentMethodsNeedsNoChecker` pins that the two harnesses agree.
func runNoUnusedClassComponentMethods(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, NoUnusedClassComponentMethods, noUnusedClassComponentMethodsFile, sourceText)
}

// TestNoUnusedClassComponentMethodsDefinitionVersusUse pins the narrowest distinction here.
//
// Upstream treats a member access as a DEFINITION when its parent is an assignment and it is the
// left side, and as a USE otherwise. That puts `+=` on the definition side and `++` on the use side,
// which reads backwards until you see that an update expression is not an assignment in the AST.
// The corpus writes neither, so a port that had grouped all writes together would pass every
// imported case. Every verdict measured on the installed build.
func TestNoUnusedClassComponentMethodsDefinitionVersusUse(t *testing.T) {
	t.Parallel()

	const head = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } "
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a lone assignment is a definition and reports", head + "componentDidMount() { this.foo = 1; } }\n", []string{"unusedWithClass"}},
		{"a compound assignment is also a definition", head + "componentDidMount() { this.foo += 1; } }\n", []string{"unusedWithClass"}},
		{"an increment is a USE, so nothing reports", head + "componentDidMount() { this.foo++; } }\n", nil},
		{"a delete is a use", head + "componentDidMount() { delete this.foo; } }\n", nil},
		{"assigning through a member is a use of the outer name", head + "componentDidMount() { this.foo.bar = 1; } }\n", nil},
		{"defining then reading is clean", head + "componentDidMount() { this.foo = 1; return this.foo; } }\n", nil},
		{"a method and an assignment of the same name report twice", head + "foo() {} componentDidMount() { this.foo = 1; } }\n", []string{"unusedWithClass", "unusedWithClass"}},
		{"an assignment in the constructor still reports", head + "constructor() { super(); this.foo = 1; } }\n", []string{"unusedWithClass"}},
		{"a nested assignment still reports", head + "componentDidMount() { if (true) { this.foo = 1; } } }\n", []string{"unusedWithClass"}},
		{"a use inside a nested arrow counts", head + "handleClick() {} componentDidMount() { const f = () => this.handleClick(); } }\n", nil},
		{"an optional-chained use counts", head + "handleClick() {} componentDidMount() { this?.handleClick(); } }\n", nil},
		// The six below pin the two halves of the assignment test, each of which survived the rest
		// of this suite as its own mutant. The left-side test is what keeps a member on the RIGHT of
		// an assignment a use, and the operator test is what keeps a member inside an ordinary
		// binary expression a use. The corpus writes neither shape. Every verdict measured on the
		// installed build, and the last row is the sharpest: one statement where the left side is a
		// definition and the right side is a use, so only `bar` reports.
		{"a member on the RIGHT of an assignment is a use", head + "handleClick() {} componentDidMount() { const x = this.handleClick; } }\n", nil},
		{"a member in a comparison is a use", head + "handleClick() {} componentDidMount() { if (this.handleClick === 1) {} } }\n", nil},
		{"a member in a logical is a use", head + "handleClick() {} componentDidMount() { const x = this.handleClick || 1; } }\n", nil},
		{"a member in an addition is a use", head + "handleClick() {} componentDidMount() { const x = this.handleClick + 1; } }\n", nil},
		{"a member as a comma operand is a use", head + "handleClick() {} componentDidMount() { (this.handleClick, 1); } }\n", nil},
		{"one statement, a definition on the left and a use on the right", head + "foo() {} componentDidMount() { this.bar = this.foo; } }\n", []string{"unusedWithClass"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnusedClassComponentMethods(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoUnusedClassComponentMethodsStaticIsInvisible pins the static flag.
//
// Entering a static member makes every `this` access inside it invisible upstream, on the reasoning
// that `this` in a static method is the constructor rather than the instance. So a static member is
// neither reported nor able to mark anything used, and a method called only from a static method
// still reports. Both measured.
func TestNoUnusedClassComponentMethodsStaticIsInvisible(t *testing.T) {
	t.Parallel()

	const head = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } "
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a static method is never reported", head + "static helper() {} }\n", nil},
		{"a static property is never reported", head + "static defaultProps = {}; }\n", nil},
		{"a use inside a static method does NOT count", head + "handleClick() {} static x() { this.handleClick(); } }\n", []string{"unusedWithClass"}},
		{"an assignment inside a static method is not a definition either", head + "static x() { this.foo = 1; } }\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnusedClassComponentMethods(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoUnusedClassComponentMethodsKeyShapes pins which keys the rule can read.
//
// Upstream's `isKeyLiteralLike` accepts a literal, a template with no substitutions, and an
// uncomputed identifier. A computed key still counts when what is inside is a literal, because that
// arm fires regardless of `computed`, and a private name has its own node type upstream never sees.
// Every verdict measured on the installed build.
func TestNoUnusedClassComponentMethodsKeyShapes(t *testing.T) {
	t.Parallel()

	const head = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } "
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain identifier key reports", head + "handleClick() {} }\n", []string{"unusedWithClass"}},
		{"a string-literal key reports", head + "'handleClick'() {} }\n", []string{"unusedWithClass"}},
		{"a computed string key reports", head + "['handleClick']() {} }\n", []string{"unusedWithClass"}},
		{"a computed template key reports", head + "[`handleClick`]() {} }\n", []string{"unusedWithClass"}},
		{"a numeric key reports", head + "42() {} }\n", []string{"unusedWithClass"}},
		{"a private name is invisible", head + "#secret() {} }\n", nil},
		{"a getter reports", head + "get value() { return 1; } }\n", []string{"unusedWithClass"}},
		{"a setter reports", head + "set value(v: number) {} }\n", []string{"unusedWithClass"}},
		{"a string key used through a dot is clean", head + "'handleClick'() {} componentDidMount() { this.handleClick(); } }\n", nil},
		{"a string key used through a subscript is clean", head + "'handleClick'() {} componentDidMount() { this['handleClick'](); } }\n", nil},
		{"a computed key used through a dot is clean", head + "['handleClick']() {} componentDidMount() { this.handleClick(); } }\n", nil},
		{"a variable subscript names nothing and does not count as a use", head + "handleClick() {} componentDidMount() { const k = 'handleClick'; this[k](); } }\n", []string{"unusedWithClass"}},
		// A computed IDENTIFIER key names something known only at run time and contributes nothing.
		// Upstream's `isKeyLiteralLike` requires `computed === false` for its identifier arm, so
		// this and the computed-literal row above fall on opposite sides of the same test. This port
		// accepted both at first and upstream's own clean case
		// `class ClassComputedMemberTest ... { [foo]() {} }` is what caught it, which is the whole
		// argument for importing the clean cases rather than only the reporting ones.
		{"a computed identifier key is invisible", "declare const React: any;\ndeclare const foo: string;\n" + head[len("declare const React: any;\n"):] + "[foo]() {} }\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnusedClassComponentMethods(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoUnusedClassComponentMethodsBothIds pins which message each shape produces.
//
// `unusedWithClass` names the class and `unused` does not, and upstream picks on whether the class
// node has a name. A `createReactClass` object never has one, and neither does an anonymous default
// export. The two ids carry different interpolations, so this asserts the rendered text as well.
func TestNoUnusedClassComponentMethodsBothIds(t *testing.T) {
	t.Parallel()

	named := runNoUnusedClassComponentMethods(t,
		"declare const React: any;\nclass Foo extends React.Component { render() { return null; } handleClick() {} }\n")
	rule_testing.ExpectFindings(t, named, "unusedWithClass")
	if !strings.HasPrefix(named.Diagnostics[0].Message.Description,
		`Unused method or property "handleClick" of class "Foo"`) {
		t.Errorf("named reads %q", named.Diagnostics[0].Message.Description)
	}

	anonymous := runNoUnusedClassComponentMethods(t,
		"declare const React: any;\nexport default class extends React.Component { render() { return null; } handleClick() {} }\n")
	rule_testing.ExpectFindings(t, anonymous, "unused")
	if !strings.HasPrefix(anonymous.Diagnostics[0].Message.Description,
		`Unused method or property "handleClick"`) {
		t.Errorf("anonymous reads %q", anonymous.Diagnostics[0].Message.Description)
	}
	if strings.Contains(anonymous.Diagnostics[0].Message.Description, "of class") {
		t.Errorf("anonymous message names a class: %q", anonymous.Diagnostics[0].Message.Description)
	}

	object := runNoUnusedClassComponentMethods(t,
		"declare function createReactClass(spec: any): any;\ncreateReactClass({ render() { return null; }, handleClick() {} });\n")
	rule_testing.ExpectFindings(t, object, "unused")
	if !strings.HasPrefix(object.Diagnostics[0].Message.Description,
		`Unused method or property "handleClick"`) {
		t.Errorf("object reads %q", object.Diagnostics[0].Message.Description)
	}
}

// TestNoUnusedClassComponentMethodsLifecycleSets pins the two era-specific exemptions.
//
// `state` is exempt in a class and NOT in a `createReactClass` object; `getInitialState`,
// `getDefaultProps` and `mixins` are exempt in the object and NOT in a class. The two sets are
// mirrors, and a port that merged them would pass every imported case, because the corpus writes
// each name only on the side where it is exempt. Every verdict measured.
func TestNoUnusedClassComponentMethodsLifecycleSets(t *testing.T) {
	t.Parallel()

	const classHead = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } "
	const objectHead = "declare function createReactClass(spec: any): any;\ncreateReactClass({ render() { return null; }, "
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"state is exempt in a class", classHead + "state = {}; }\n", nil},
		{"getInitialState REPORTS in a class", classHead + "getInitialState() {} }\n", []string{"unusedWithClass"}},
		{"getDefaultProps REPORTS in a class", classHead + "getDefaultProps() {} }\n", []string{"unusedWithClass"}},
		{"mixins REPORTS in a class", classHead + "mixins = []; }\n", []string{"unusedWithClass"}},
		{"getInitialState is exempt in an object", objectHead + "getInitialState() { return {}; } });\n", nil},
		{"getDefaultProps is exempt in an object", objectHead + "getDefaultProps() { return {}; } });\n", nil},
		{"mixins is exempt in an object", objectHead + "mixins: [] });\n", nil},
		{"componentDidMount is exempt in both", classHead + "componentDidMount() {} }\n", nil},
		{"an UNSAFE lifecycle name is exempt", classHead + "UNSAFE_componentWillMount() {} }\n", nil},
		{"getDerivedStateFromProps is NOT a listed lifecycle name", classHead + "getDerivedStateFromProps() {} }\n", []string{"unusedWithClass"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnusedClassComponentMethods(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoUnusedClassComponentMethodsDestructuringCountsAsUse pins the `{ foo } = this` path.
//
// Upstream has a whole `VariableDeclarator` listener for it, and the renamed form
// `const { foo: bar } = this` counts the SOURCE name rather than the local one, because it reads
// `prop.key`. Both measured.
func TestNoUnusedClassComponentMethodsDestructuringCountsAsUse(t *testing.T) {
	t.Parallel()

	const head = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } handleClick() {} "
	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain destructure counts", head + "componentDidMount() { const { handleClick } = this; } }\n", nil},
		{"a renamed destructure counts the source name", head + "componentDidMount() { const { handleClick: h } = this; } }\n", nil},
		{"destructuring a different name does not count", head + "componentDidMount() { const { somethingElse } = this; } }\n", []string{"unusedWithClass"}},
		{"destructuring from something other than this does not count", head + "componentDidMount() { const other = { handleClick: 1 }; const { handleClick } = other; } }\n", []string{"unusedWithClass"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnusedClassComponentMethods(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoUnusedClassComponentMethodsSpans asserts WHERE each finding points.
//
// Upstream reports on the KEY node rather than on the member, so a finding on an unused method
// points at the name and not at the body. That is invisible to every id assertion above, and it is
// the difference between a reader seeing one word underlined and seeing twenty lines.
func TestNoUnusedClassComponentMethodsSpans(t *testing.T) {
	t.Parallel()

	const head = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } "
	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{"a method points at its name", head + "handleClick() { return 1; } }", []string{"handleClick"}},
		{"a property points at its name", head + "handleClick = () => {}; }", []string{"handleClick"}},
		{"an assignment points at the property name", head + "componentDidMount() { this.foo = 1; } }", []string{"foo"}},
		{"a string key points at the literal", head + "'handleClick'() {} }", []string{"'handleClick'"}},
		{"a computed key points INSIDE the brackets", head + "['handleClick']() {} }", []string{"'handleClick'"}},
		{"a getter points at its name", head + "get value() { return 1; } }", []string{"value"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoUnusedClassComponentMethods(t, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			source := result.SourceFile.Text()
			for index, diagnostic := range result.Diagnostics {
				got := source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.wantTexts[index] {
					t.Errorf("finding %d points at %q, want %q", index, got, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestNoUnusedClassComponentMethodsComponentGate pins what counts as a component.
//
// A plain class is not one, a class EXPRESSION is not listened for, and only `createReactClass`
// counts on the object side: the shelf's `react.IsEs5ComponentCall` also accepts `createClass` and
// the namespaced spelling, and upstream accepts neither. Measured.
func TestNoUnusedClassComponentMethodsComponentGate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a class with no heritage is not a component", "class Foo { render() { return null; } handleClick() {} }\n", nil},
		{"a foreign base is not a component", "declare const Foo: any;\nclass Bar extends Foo.Component { render() { return null; } handleClick() {} }\n", nil},
		{"a class EXPRESSION is not listened for", "declare const React: any;\nconst Foo = class extends React.Component { render() { return null; } handleClick() {} };\n", nil},
		{"createReactClass counts", "declare function createReactClass(spec: any): any;\ncreateReactClass({ render() { return null; }, handleClick() {} });\n", []string{"unused"}},
		{"a bare createClass does not", "declare function createClass(spec: any): any;\ncreateClass({ render() { return null; }, handleClick() {} });\n", nil},
		{"React.createClass does not", "declare const React: any;\nReact.createClass({ render() { return null; }, handleClick() {} });\n", nil},
		{"an unrelated call does not", "declare function other(spec: any): any;\nother({ render() { return null; }, handleClick() {} });\n", nil},
		{"a bare object literal is not a component", "const x = { render() { return null; }, handleClick() {} };\n", nil},
		// Upstream checks the parent CALLEE and nothing else, so an object in any argument slot
		// counts. This port checked the first slot at first and a mutant dropping that check
		// survived every fixture; probing the installed build showed the mutant was right and the
		// rule was wrong. A nested object is still not a component, because its parent is a property
		// assignment rather than the call.
		{"an object in the SECOND argument slot still counts", "declare function createReactClass(a: any, b: any): any;\ndeclare const other: any;\ncreateReactClass(other, { render() { return null; }, handleClick() {} });\n", []string{"unused"}},
		{"a nested object inside the specification is not a component", "declare function createReactClass(spec: any): any;\ncreateReactClass({ render() { return null; }, config: { handleClick() {} } });\n", []string{"unused"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnusedClassComponentMethods(t, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestNoUnusedClassComponentMethodsNestedClass records a stated divergence.
//
// Upstream keeps ONE mutable `classInfo` slot. Entering an inner class declaration overwrites it and
// the inner class's exit sets it to null, so the outer class stops collecting from that point and
// never reports. Measured on the installed build: the source below reports only the INNER class's
// unused member, and the outer `outerUnused` is silently dropped.
//
// This port walks each component separately and reports BOTH, which is wider than upstream. The
// alternative is reproducing a single-slot bug that silently loses findings, and the shape that
// triggers it, a component declared inside another component's method, is rare. Stated here rather
// than hidden, so the next reader meets a decision rather than a surprise.
func TestNoUnusedClassComponentMethodsNestedClass(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\n" +
		"class Outer extends React.Component {\n" +
		"  render() { return null; }\n" +
		"  componentDidMount() {\n" +
		"    class Inner extends React.Component { render() { return null; } innerUnused() {} }\n" +
		"  }\n" +
		"  outerUnused() {}\n" +
		"}\n"

	result := runNoUnusedClassComponentMethods(t, source)
	rule_testing.ExpectFindings(t, result, "unusedWithClass", "unusedWithClass")

	source_ := result.SourceFile.Text()
	names := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		names = append(names, source_[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	// Upstream reports only `innerUnused` here. This port reports both, and the order follows the
	// walk: the outer class is entered first, so its member is judged first.
	if names[0] != "outerUnused" || names[1] != "innerUnused" {
		t.Errorf("reported %v, want [outerUnused innerUnused]", names)
	}
}

// TestNoUnusedClassComponentMethodsNeedsNoChecker pins that this rule is syntactic.
func TestNoUnusedClassComponentMethodsNeedsNoChecker(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } handleClick() {} }\n"

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUnusedClassComponentMethods,
		noUnusedClassComponentMethodsFile, source), "unusedWithClass")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnusedClassComponentMethods,
		noUnusedClassComponentMethodsFile, source), "unusedWithClass")
}

// TestNoUnusedClassComponentMethodsHasNoFileGate reports in a `.ts` file, not only in `.tsx`.
//
// Three shipped rules in this package carried a `.tsx`-only gate that was oxc residue rather than
// upstream behavior, each with a test asserting the gate, so the suite locked the bug in. This pins
// the opposite. The source carries no JSX, because JSX in a `.ts` file is a syntax error and a case
// that failed to parse would be silent for a reason unrelated to any gate.
func TestNoUnusedClassComponentMethodsHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nclass Foo extends React.Component { render() { return null; } handleClick() {} }\n"
	for _, fileName := range []string{
		"/repository/source/NoUnusedClassComponentMethods.tsx",
		"/repository/source/NoUnusedClassComponentMethods.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUnusedClassComponentMethods, fileName, source),
				"unusedWithClass")
		})
	}
}

// TestNoUnusedClassComponentMethodsSurvivesShapesThatWouldPanic drives shapes with pieces missing.
//
// The walk recovers per FILE rather than per rule, so one nil dereference here takes the file away
// from every rule in the tree, and no `ExpectFindings` fixture can see a panic.
func TestNoUnusedClassComponentMethodsSurvivesShapesThatWouldPanic(t *testing.T) {
	t.Parallel()

	sources := []string{
		"declare function createReactClass(spec: any): any;\ncreateReactClass();\n",
		"declare function createReactClass(spec: any): any;\ncreateReactClass(1);\n",
		"declare const React: any;\nclass Foo extends React.Component {}\n",
		"declare const React: any;\ndeclare const spread: any;\nclass Foo extends React.Component { render() { return null; } componentDidMount() { const { ...rest } = this; } }\n",
		"declare const React: any;\nclass Foo extends React.Component { render() { return null; } componentDidMount() { this[0] = 1; } }\n",
		"declare const React: any;\nclass Foo extends React.Component { render() { return null; } componentDidMount() { const [a] = this as any; } }\n",
		"(()=>{})();\n",
		"class Base {}\nclass Derived extends Base { constructor() { super(); } }\n",
	}

	for index, sourceText := range sources {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			t.Parallel()
			runNoUnusedClassComponentMethods(t, sourceText)
		})
	}
}
