package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// thisInSfcFile is where the fixtures pretend to live.
//
// A `.tsx` suffix rather than `.ts`, because the rule declines a file the parser does not read as
// JSX and every fixture below would pass vacuously under the other suffix. The neighbouring rules
// in this package spell it the same way for the same reason.
const thisInSfcFile = "/repository/source/Thing.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two tables below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_this_in_sfc.rs`: 24 pass and 14 fail, from the single
// tester block the extractor reports. The strings were not transcribed by hand. They were decoded
// out of the extractor's own dumper, written to Go literals by a script, and every one was then
// checked byte-for-byte against the Rust source, because a tool that writes a fixture can cook an
// escape and the result still compiles and still goes green.
//
// The snapshot records 15 diagnostics from those 14 fail inputs, so one finding per input is wrong
// here. The extra one belongs to the last case, which reads `this` twice: once in a `ref` callback
// body and once in a spread. Recovered by reading the snapshot's own source lines rather than by
// counting in order, and the two diagnostics print `this.itemRef = ref;` and
// `{...this.getBasicProps()}`, which appear in no other input.
func TestNoThisInSfcStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"props destructured from the parameter", "\n                    function Foo(props) {\n                      const { foo } = props;\n                      return <div bar={foo} />;\n                    }\n                  "},
		{"a destructured parameter", "\n                    function Foo({ foo }) {\n                      return <div bar={foo} />;\n                    }\n                  "},
		{"a class component method", "\n                    class Foo extends React.Component {\n                      render() {\n                        const { foo } = this.props;\n                        return <div bar={foo} />;\n                      }\n                    }\n                  "},
		{"an es5 component with an anonymous render", "\n                    const Foo = createReactClass({\n                      render: function() {\n                        return <div>{this.props.foo}</div>;\n                      }\n                    });\n                  "},
		{"a React.createClass with an anonymous render", "\n                    const Foo = React.createClass({\n                      render: function() {\n                        return <div>{this.props.foo}</div>;\n                      }\n                    });\n                  "},
		{"a lowercase function assigning to this", "\n                    function foo(bar) {\n                      this.bar = bar;\n                      this.props = 'baz';\n                      this.getFoo = function() {\n                        return this.bar + this.props;\n                      }\n                    }\n                  "},
		{"a conditional returning jsx", "\n                    function Foo(props) {\n                      return props.foo ? <span>{props.bar}</span> : null;\n                    }\n                  "},
		{"an early return of jsx", "\n                    function Foo(props) {\n                      if (props.foo) {\n                        return <div>{props.bar}</div>;\n                      }\n                      return null;\n                    }\n                  "},
		{"a conditional with no jsx", "\n                    function Foo(props) {\n                      if (props.foo) {\n                        something();\n                      }\n                      return null;\n                    }\n                  "},
		{"an arrow returning jsx", "const Foo = (props) => <span>{props.foo}</span>"},
		{"an arrow with a destructured parameter", "const Foo = ({ foo }) => <span>{foo}</span>"},
		{"an arrow with a conditional", "const Foo = (props) => props.foo ? <span>{props.bar}</span> : null;"},
		{"an arrow destructuring and conditional", "const Foo = ({ foo, bar }) => foo ? <span>{bar}</span> : null;"},
		{"an arrow inside a plain class method", "\n                    class Foo {\n                      bar() {\n                        () => {\n                          this.something();\n                          return null;\n                        };\n                      }\n                    }\n                  "},
		{"a class property arrow", "\n                    class Foo {\n                      bar = () => {\n                        this.something();\n                        return null;\n                      };\n                    }\n                  "},
		{"a method inside an object returned by a component", "\n                    export const Example = ({ prop }) => {\n                      return {\n                        handleClick: () => {},\n                        renderNode() {\n                          return <div onClick={this.handleClick} />;\n                        },\n                      };\n                    };\n                  "},
		{"a Meteor validated method", "\n                    export const prepareLogin = new ValidatedMethod({\n                      name: \"user.prepare\",\n                      validate: new SimpleSchema({\n                      }).validator(),\n                      run({ remember }) {\n                          if (Meteor.isServer) {\n                              const connectionId = this.connection.id; // react/no-this-in-sfc\n                              return Methods.prepareLogin(connectionId, remember);\n                          }\n                          return null;\n                      },\n                    });\n                  "},
		{"a function assigned to an object property", "\n                    obj.notAComponent = function () {\n                      return this.a || null;\n                    };\n                  "},
		{"a jQuery plugin function", "\n                    $.fn.getValueAsStringWeak = function (): string | null {\n                      const val = this.length === 1 ? this.val() : null;\n\n                      return typeof val === 'string' ? val : null;\n                    };\n                  "},
		{"this passed to bind inside a component", "\n                    const Foo = ({query}) => {\n                      return <div onClick={reopen.bind(this, query)}>click</div>\n                    }\n                  "},
		{"a nested function declaration", "\n                    function Foo() {\n                      function callback() {\n                        this.itemRef = null;\n                      }\n                      return <div ref={callback} />;\n                    }\n                  "},
		{"a static block inside a component", "\n                    function Foo() {\n                      class C {\n                        static {\n                          const callback = () => this.value;\n                        }\n                      }\n                      return <div />;\n                    }\n                  "},
		{"an accessor initializer inside a component", "\n                    function Foo() {\n                      class C {\n                        accessor handler = () => this.value;\n                      }\n                      return <div />;\n                    }\n                  "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText))
		})
	}
}

// The failing corpus, verbatim. All but the last report exactly once.
func TestNoThisInSfcFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"props destructured from this", "\n                    function Foo(props) {\n                      const { foo } = this.props;\n                      return <div>{foo}</div>;\n                    }\n                  "},
		{"a member read inside jsx", "\n                    function Foo(props) {\n                      return <div>{this.props.foo}</div>;\n                    }\n                  "},
		{"a state read inside jsx", "\n                    function Foo(props) {\n                      return <div>{this.state.foo}</div>;\n                    }\n                  "},
		{"state destructured from this", "\n                    function Foo(props) {\n                      const { foo } = this.state;\n                      return <div>{foo}</div>;\n                    }\n                  "},
		{"a member read in a conditional", "\n                    function Foo(props) {\n                      return props.foo ? <div>{this.props.bar}</div> : null;\n                    }\n                  "},
		{"a member read after an early return", "\n                    function Foo(props) {\n                      if (props.foo) {\n                        return <div>{this.props.bar}</div>;\n                      }\n                      return null;\n                    }\n                  "},
		{"a member read in the condition itself", "\n                    function Foo(props) {\n                      if (this.props.foo) {\n                        something();\n                      }\n                      return null;\n                    }\n                  "},
		{"an arrow reading this in jsx", "const Foo = (props) => <span>{this.props.foo}</span>"},
		{"an arrow reading this in a conditional", "const Foo = (props) => this.props.foo ? <span>{props.bar}</span> : null;"},
		{"a computed member read", "\n                    function Foo(props) {\n                      return <div>{this[\"props\"].foo}</div>;\n                    }\n                  "},
		{"a nested function declaration and a direct read", "\n                    function Foo(props) {\n                      function onClick(bar) {\n                        this.props.onClick();\n                      }\n                      return <div onClick={onClick}>{this.props.foo}</div>;\n                    }\n                  "},
		{"an arrow in an object literal", "\n                    function Foo() {\n                      const handlers = {\n                        click: () => this.props.click(),\n                      };\n                      return <button onClick={handlers.click} />;\n                    }\n                  "},
		{"an accessor computed key", "\n                    function Foo() {\n                      class C {\n                        accessor [this.props.name] = 0;\n                      }\n                      return <div />;\n                    }\n                  "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText), "noThisInSfc")
		})
	}
}

// The one input in the corpus that reports twice, which is the whole reason the extractor warns
// about a discrepancy between 15 diagnostics and 14 inputs.
//
// The arrow is component-named and lives inside a plain class's constructor, so it is a component
// and the class is not one. Two reads inside it are both attributed to it: one in the body of a
// `ref` callback arrow, and one in a spread. Neither is exempted, because an arrow does not rebind
// `this` and a spread is not a nested context at all. Asserting two ids here rather than one is
// what separates this from a port that reports once and looks correct.
// The one input in the corpus that reports twice, which is the whole reason the extractor warns
// about a discrepancy between 15 diagnostics and 14 inputs.
//
// The arrow is component-named and lives inside a plain class's constructor, so the arrow is a
// component and the enclosing class is not one. Two reads inside it are both attributed to it: one
// in the body of a `ref` callback arrow, and one in a spread. Neither is exempted, because an arrow
// does not rebind `this` and a spread is not a nested context at all. Asserting two ids rather than
// one is what separates this from a port that reports once and looks correct.
func TestNoThisInSfcReportsTwiceInOneComponent(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile,
		"\n                    class ItemAdapter {\n                      constructor() {\n                        const ElementWrapper = (props) => (\n                          <div\n                            ref={ref => {\n                              this.itemRef = ref;\n                            }}\n                            {...this.getBasicProps()}\n                          >\n                            {props.label}\n                          </div>\n                        );\n                        this.el = ElementWrapper;\n                      }\n                    }\n                  ")
	rule_testing.ExpectFindings(t, result, "noThisInSfc", "noThisInSfc")
}

// Where the finding points, which no message-id assertion can see.
//
// oxc reports `this_expr.span`, the four characters of the keyword alone, and never the member
// expression around it. ESLint's rule reports the whole `MemberExpression`, so a port following
// ESLint would be green on every id assertion above while pointing at `this.props.foo`. The
// upstream snapshot underlines exactly `this` on all fifteen findings, which is the authority here.
func TestNoThisInSfcPointsAtTheKeywordAlone(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a dotted read", "function Foo(props) {\n  return <div>{this.props.foo}</div>;\n}\n"},
		{"a computed read", "function Foo(props) {\n  return <div>{this[\"props\"].foo}</div>;\n}\n"},
		{"an arrow component", "const Foo = (props) => <span>{this.props.foo}</span>;\n"},
		{"a read in a condition", "function Foo(props) {\n  if (this.props.foo) {\n    something();\n  }\n  return null;\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != "this" {
				t.Errorf("want the finding to point at `this`, got %q", reported)
			}
		})
	}
}

// The rendered message text, asserted exactly rather than by a substring.
//
// A predicate weaker than the property it guards is not a guard: `strings.Contains` on an
// interpolated value passes for a message that is wrong in a way the needle does not reach. This
// message interpolates nothing, so equality is the honest assertion and it also pins that nothing
// was added to it.
func TestNoThisInSfcMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile,
		"function Foo(props) {\n  return <div>{this.props.foo}</div>;\n}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noThisInSfc" {
		t.Errorf("want the id noThisInSfc, got %q", got)
	}
	if got := result.Diagnostics[0].Message.Description; got != messageNoThisInSfc.Description {
		t.Errorf("the rendered description drifted from the declared one:\n got %q", got)
	}
	// The description says why the code is wrong rather than restating the rule name, so it must
	// not be the rule's own title.
	if strings.Contains(result.Diagnostics[0].Message.Description, "no-this-in-sfc") {
		t.Error("the description restates the rule name")
	}
}

// The one upstream pass case this rule does not decide, moved out of the clean table deliberately.
//
// Upstream ships it as passing, and it passes there because of an `eslint-disable-next-line
// react/no-this-in-sfc` comment sitting above the read, not because of anything the rule judges.
// Measured on the release binary: with the comment the file is silent, and deleting that one line
// makes the same file report once. So the case is a test of directive handling wearing a rule
// fixture's clothes.
//
// In this tree directives live in `internal/suppression`, a layer above the rule, and
// `rule_testing.Run` walks the rule alone and never consults it. Leaving the case in the clean table
// would therefore have asserted that this rule declines an input it must report, and the only way
// to make it green would have been to break the port. It is pinned here as REPORTING instead, which
// is what the rule genuinely decides, and the suppression layer is what upstream's version of this
// assertion belongs to.
func TestNoThisInSfcReportsUnderADisableDirective(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile,
		"\n                    class ItemAdapter {\n                      constructor() {\n                        const ElementWrapper = () => (\n                          <div ref={ref => {\n                            // eslint-disable-next-line react/no-this-in-sfc\n                            this.itemRef = ref;\n                          }} />\n                        );\n                        this.el = ElementWrapper;\n                      }\n                    }\n                  ")
	rule_testing.ExpectFindings(t, result, "noThisInSfc")
}

// Parenthesis behavior, which no imported fixture in either corpus writes.
//
// Guessing here is invisible at fixture time and ships a divergence in either direction, so all six
// forms were run against `~/Projects/system/oxc/target/release/oxlint` and the results are pinned
// rather than reasoned about. The anchor requires the `this`'s immediate parent to be a member
// expression and skips no parentheses, so where the parentheses land decides the verdict:
// `(this).props` puts a parenthesized expression between them and is silent, while `(this.props).a`
// leaves them outside the member expression and reports.
//
// Adding the `SkipParentheses` that usually reads as a free correctness improvement would report
// the two silent rows.
func TestNoThisInSfcParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"the unparenthesized control", "function Foo(props) {\n  return <div>{this.props.foo}</div>;\n}\n", true},
		{"parentheses around the member expression", "function Foo(props) {\n  return <div>{(this.props).foo}</div>;\n}\n", true},
		{"parentheses around the keyword", "function Foo(props) {\n  return <div>{(this).props.foo}</div>;\n}\n", false},
		{"doubled parentheses around the keyword", "function Foo(props) {\n  return <div>{((this)).props.foo}</div>;\n}\n", false},
		{"a bare keyword with no member access", "function Foo(props) {\n  const x = this;\n  return <div>{x}</div>;\n}\n", false},
		{"a call rather than a member access", "function Foo(props) {\n  return <div>{this()}</div>;\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "noThisInSfc")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// What counts as a component, which is the whole rule and where a port of it goes wrong.
//
// The corpus exercises a capitalized declaration and a capitalized arrow and almost nothing else,
// so the boundary is invented here from probing the release binary. Every row below was measured,
// and several contradict what the rule's name suggests: returning JSX is not part of the test at
// all, and no wrapper is recognized.
func TestNoThisInSfcComponentBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		// The name is read off the function's own identifier, never off the variable it is
		// assigned to. These three rows are the discriminating ones and they disagree.
		{"a named function expression keeps its own name", "const Foo = function Bar(props) {\n  return this.props.a;\n};\n", true},
		{"an anonymous function expression has no name to read", "const Foo = function (props) {\n  return this.props.a;\n};\n", false},
		{"a lowercase function id beats an uppercase variable", "const Foo = function bar(props) {\n  return this.props.a;\n};\n", false},

		// Returning JSX is irrelevant. Upstream's own corpus has a failing case with no JSX in it.
		{"a component that returns no jsx at all", "function Foo(props) {\n  return this.props.a;\n}\n", true},

		// No wrapper is recognized, because the arrow's parent is a call argument rather than a
		// variable declarator, so no name can be read.
		{"an arrow handed to memo", "const Foo = memo((props) => this.props.a);\n", false},
		{"an arrow handed to React.memo", "const Foo = React.memo((props) => this.props.a);\n", false},
		{"an arrow handed to forwardRef", "const Foo = React.forwardRef((props, ref) => this.props.a);\n", false},
		{"an arrow in an object property", "const o = { Foo: (props) => this.props.a };\n", false},
		{"a destructured binding", "const { Foo } = { Foo: (props) => this.props.a };\n", false},
		{"a default export of an anonymous arrow", "export default (props) => this.props.a;\n", false},

		// The first character decides, and only an ASCII uppercase letter counts.
		{"a lowercase declaration", "function foo(props) {\n  return this.props.a;\n}\n", false},
		{"a leading underscore", "function _Foo(props) {\n  return this.props.a;\n}\n", false},
		{"a leading dollar sign", "const $Foo = (props) => this.props.a;\n", false},
		{"a hook name", "function useFoo() {\n  return this.props.a;\n}\n", false},
		{"a let binding still counts", "let Foo = (props) => this.props.a;\n", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "noThisInSfc")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// The name test is ASCII-uppercase, and the shelf helper that looks right here is not.
//
// `internal/utilities/react.IsLikelyComponentName` is `unicode.IsUpper(runes[0])` and answers true for
// all three of these, so a port reaching for it by name would report all three. oxc's
// `is_react_component_name` is `c.is_ascii_uppercase()` and all three are silent on the release
// binary. This is the fixture that fails if somebody later swaps the local predicate for the shelf.
func TestNoThisInSfcNameTestIsAscii(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"function Фoo(props) {\n  return this.props.a;\n}\n",
		"function Λoo(props) {\n  return this.props.a;\n}\n",
		"function Éoo(props) {\n  return this.props.a;\n}\n",
	} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoThisInSfc, thisInSfcFile, sourceText))
	}

	// The control, so a zero above cannot come from the source failing to parse.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoThisInSfc, thisInSfcFile,
		"function Foo(props) {\n  return this.props.a;\n}\n"), "noThisInSfc")
}

// Which construct rebinds `this`, and the asymmetry between the two function kinds.
//
// A `Function` ancestor ends the outward walk whether or not it is a component, so a nested
// function declaration or an anonymous function expression exempts its whole body. An arrow that is
// not component-named is stepped over and the walk continues, which is what lets a callback inside
// a component still belong to it. The corpus covers the nested declaration and the inline arrow;
// the rest is measured.
func TestNoThisInSfcNesting(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"a nested arrow is not a boundary", "function Foo(p) {\n  const cb = () => this.props.a;\n  return <div />;\n}\n", true},
		{"a nested function declaration is", "function Foo(p) {\n  function cb() {\n    return this.props.a;\n  }\n  return <div />;\n}\n", false},
		{"a nested anonymous function expression is", "function Foo(p) {\n  const cb = function () {\n    return this.props.a;\n  };\n  return <div />;\n}\n", false},
		{"a nested capitalized arrow becomes the component itself", "function Foo(p) {\n  const Cb = () => this.props.a;\n  return <div />;\n}\n", true},
		{"arrows nest arbitrarily deep", "function Foo(p) {\n  const a = () => () => () => this.props.x;\n  return <div />;\n}\n", true},
		{"an object method is a function", "function Foo(p) {\n  return { m() { return this.props.a; } };\n}\n", false},
		{"an object getter is too", "function Foo(p) {\n  return { get m() { return this.props.a; } };\n}\n", false},
		{"a class method inside a component", "function Foo(p) {\n  class C { m() { return this.props.a; } }\n  return <div />;\n}\n", false},
		{"a plain class property initializer", "function Foo(p) {\n  class C { h = () => this.props.a; }\n  return <div />;\n}\n", false},
		{"a plain class computed key", "function Foo(p) {\n  class C { [this.props.n] = 0; }\n  return <div />;\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "noThisInSfc")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// The class-component ancestor exemption, across every heritage spelling.
//
// After a component is found, upstream walks that component's own ancestors and exempts if any is
// a React component. The corpus contains no case of a component nested inside a class component at
// all, so every row here is measured on the release binary. The four accepted spellings and the
// three rejected ones are what `IsEs6ComponentClass` answers, which is why that shelf helper is
// used rather than replaced.
func TestNoThisInSfcClassComponentAncestor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"React.Component exempts", "class A extends React.Component {\n  m() { const Foo = (p) => this.props.a; }\n}\n", false},
		{"React.PureComponent exempts", "class A extends React.PureComponent {\n  m() { const Foo = (p) => this.props.a; }\n}\n", false},
		{"a bare Component exempts", "class A extends Component {\n  m() { const Foo = (p) => this.props.a; }\n}\n", false},
		{"a bare PureComponent exempts", "class A extends PureComponent {\n  m() { const Foo = (p) => this.props.a; }\n}\n", false},
		{"a class expression exempts", "const A = class extends React.Component {\n  m() { const Foo = (p) => this.props.a; }\n};\n", false},
		{"another namespace does not", "class A extends Foo.Component {\n  m() { const Bar = (p) => this.props.a; }\n}\n", true},
		{"an unrelated base does not", "class A extends Base {\n  m() { const Foo = (p) => this.props.a; }\n}\n", true},
		{"no heritage clause does not", "class A {\n  m() { const Foo = (p) => this.props.a; }\n}\n", true},
		{"a plain class nested inside a component still exempts", "class Outer extends React.Component {\n  m() { class Inner { n() { const Foo = (p) => this.props.a; } } }\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "noThisInSfc")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// The ES5 factory name, where the shelf helper is wrong in the other direction.
//
// `internal/utilities/react.IsEs5ComponentCall` accepts `createClass` and `React.createClass` as well
// as `createReactClass`, while oxc keys `is_es5_component` on `createReactClass` alone. Measured on
// the release binary: only the strict spellings exempt, and the two loose ones report. No imported
// fixture can see this, because the corpus never nests a component inside an ES5 factory.
//
// Upstream's `React.createClass` PASS case looks like evidence for the opposite and is not: it
// passes because its `render: function () {...}` is anonymous, so no name can be read and the walk
// ends at the function. Removing the factory leaves it passing, which was measured. That is why
// this test writes a component-named arrow inside each factory rather than reusing that shape.
func TestNoThisInSfcEs5FactoryNameIsStrict(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"createReactClass exempts", "const X = createReactClass({\n  m() { const Foo = (p) => this.props.a; },\n});\n", false},
		{"React.createReactClass exempts", "const X = React.createReactClass({\n  m() { const Foo = (p) => this.props.a; },\n});\n", false},
		{"a bare createClass does not", "const X = createClass({\n  m() { const Foo = (p) => this.props.a; },\n});\n", true},
		{"React.createClass does not", "const X = React.createClass({\n  m() { const Foo = (p) => this.props.a; },\n});\n", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "noThisInSfc")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// The accessor split, which our AST spells with one node kind where oxc has two.
//
// oxc's `PropertyDefinition` exempts unconditionally and its `AccessorProperty` exempts only when
// the value contains the `this`, so a computed KEY on an accessor is not exempt while everything
// about a plain property is. Both accessor rows are in the corpus, one passing and one failing, so
// this is the one part of the exemption the imported fixtures can see; the two plain rows are
// measured and pin that the unconditional arm is really unconditional.
func TestNoThisInSfcAccessorSplit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"an accessor value exempts", "function Foo() {\n  class C { accessor h = () => this.value; }\n  return <div />;\n}\n", false},
		{"an accessor computed key does not", "function Foo() {\n  class C { accessor [this.props.n] = 0; }\n  return <div />;\n}\n", true},
		{"a plain property value exempts", "function Foo() {\n  class C { h = () => this.value; }\n  return <div />;\n}\n", false},
		{"a plain property computed key also exempts", "function Foo() {\n  class C { [this.props.n] = 0; }\n  return <div />;\n}\n", false},
		{"a static block exempts", "function Foo() {\n  class C { static { const cb = () => this.value; } }\n  return <div />;\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisInSfc, thisInSfcFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "noThisInSfc")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// No file-suffix gate, asserted rather than assumed.
//
// oxc gated the whole rule on `source_type().is_jsx()` and this package spelled that as the file
// suffix, so a `.ts` file registered no listener at all. eslint-plugin-react, the authority this
// rule is ported against, has no such gate, and a function component in a `.ts` file is ordinary.
// Both suffixes now report the same source.
func TestNoThisInSfcReportsUnderAnySuffix(t *testing.T) {
	t.Parallel()

	const sourceText = "function Foo(props) {\n  return this.props.a;\n}\n"
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoThisInSfc, thisInSfcFile, sourceText), "noThisInSfc")
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoThisInSfc, "/repository/source/Thing.ts", sourceText), "noThisInSfc")
}

// A `this` in the ARGUMENT of an element access, which the parent-kind check accepts.
//
// Upstream tests `is_member_expression_kind` on the parent alone and never asks whether the `this`
// is the OBJECT of that member expression. So `a[this]` inside a component satisfies the anchor
// even though nothing is being read off `this`. Measured on the release binary rather than assumed,
// because it looks like a case a port would narrow away as obviously unintended.
func TestNoThisInSfcAcceptsAnElementAccessArgument(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoThisInSfc, thisInSfcFile,
		"function Foo(props) {\n  return <div>{props.a[this]}</div>;\n}\n"), "noThisInSfc")
}
