package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// directMutationStateFile is where the fixtures pretend to live.
//
// A `.tsx` suffix because upstream gates the whole rule on `source_type().is_jsx()` and most of the
// imported cases carry JSX in a render method, which would not parse otherwise. The suffix is also
// what the rule itself reads to reproduce that gate, so a fixture in a `.ts` file would go silent
// for the right reason and assert nothing.
const directMutationStateFile = "/repository/source/DirectMutationState.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two upstream tests below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_direct_mutation_state.rs`, checked byte against byte by
// a script that asserted each extracted string appears in the Rust source rather than by reading
// them. The extractor reported one tester block, 8 pass and 15 fail, no case carrying options, and
// a DISCREPANCY: 22 diagnostics against 15 fail inputs.
//
// That discrepancy is real rather than a second-snapshot artifact. There is one snapshot file and
// its 22 location headers partition over the 15 inputs uniquely, because diagnostics appear in
// input order and every fail input reports at least once: the first case writes to state three
// times, the fifth twice, the last five times, and the remaining twelve once each. The partition
// was checked against the source line each diagnostic prints rather than by counting alone.
//
// So `wantSpans` below carries one entry per expected finding, in order, and its length is the
// finding count. That is deliberate: `ExpectFindings` asserts ids and count and nothing else, and
// this rule's whole subject is which sub-expression a write is rooted at, which only a span can see.

// TestNoDirectMutationStateFires is upstream's fail corpus.
//
// `wantSpans` is the text the upstream snapshot underlines for each diagnostic, in order, extracted
// from the snapshot mechanically rather than transcribed. Two different spans live in here and the
// difference is the rule's own: an assignment reports its left-hand side alone, while an update
// expression reports the whole expression including the operator. `this.state.foo` against
// `this.state.foo++` is that difference, and no message-id assertion can see it.
func TestNoDirectMutationStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{"two lifecycle writes in one ES5 component plus one in a second", `
                  var Hello = createReactClass({

                    componentWillMount() {
                      this.state.foo = "Chicken, you're so beautiful"
                    },

                          render: function() {
                            this.state.foo = "Chicken, you're so beautiful"
                            return <div>Hello{this.props.name} <Hello2/></div>;
                          }
                        });

                  var Hello2 = createReactClass({
                          render: () => {
                             this.state.foo = "Chicken, you're so beautiful"
                            return <div>Hello {this.props.name}</div>;
                          }
                        });
          `, []string{"this.state.foo", "this.state.foo", "this.state.foo"}},
		{"an increment in render", `
                 var Hello = createReactClass({
                   render: function() {
                     this.state.foo++;
                     return <div>Hello {this.props.name}</div>;
                   }
                 });
               `, []string{"this.state.foo++"}},
		{"a two-level nested property write", `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name= "bar"
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, []string{"this.state.person.name"}},
		{"a three-level nested property write", `
          var Hello = createReactClass({
            render: function() {
              this.state.person.name.first = "bar"
              return <div>Hello</div>;
            }
          });
        `, []string{"this.state.person.name.first"}},
		{"two three-level writes in one render", `
          var Hello = createReactClass({
            render: function() {
              this.state.person.name.first = "bar"
              this.state.person.name.last = "baz"
              return <div>Hello</div>;
            }
          });
        `, []string{"this.state.person.name.first", "this.state.person.name.last"}},
		{"a write in a method the constructor calls", `
          class Hello extends React.Component {
            constructor() {
              someFn()
            }
            someFn() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"a write inside a constructor callback", `
          class Hello extends React.Component {
            constructor(props) {
              super(props)
              doSomethingAsync(() => {
                this.state = "bad";
              });
            }
          }
        `, []string{"this.state"}},
		{"componentWillMount", `
          class Hello extends React.Component {
            componentWillMount() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"componentDidMount", `
          class Hello extends React.Component {
            componentDidMount() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"componentWillReceiveProps", `
          class Hello extends React.Component {
            componentWillReceiveProps() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"shouldComponentUpdate", `
          class Hello extends React.Component {
            shouldComponentUpdate() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"componentWillUpdate", `
          class Hello extends React.Component {
            componentWillUpdate() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"componentDidUpdate", `
          class Hello extends React.Component {
            componentDidUpdate() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"componentWillUnmount", `
          class Hello extends React.Component {
            componentWillUnmount() {
              this.state.foo = "bar"
            }
          }
        `, []string{"this.state.foo"}},
		{"five computed-subscript writes in one method", `
          class Hello extends React.Component {
            update(id) {
              this.state.x[id] = 1;
              this.state.x[id].y = 1;
              this.state.items[id].prop = {};
              this.state.x[id]++;
              this.state.items[id].prop++;
            }
          }
        `, []string{"this.state.x[id]", "this.state.x[id].y", "this.state.items[id].prop", "this.state.x[id]++", "this.state.items[id].prop++"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDirectMutationState, directMutationStateFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantSpans))
			for index := range wantIds {
				wantIds[index] = "noDirectMutationState"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				if index >= len(result.Diagnostics) {
					break
				}
				diagnostic := result.Diagnostics[index]
				if got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]; got != wantSpan {
					t.Errorf("finding %d reported %q, want %q", index, got, wantSpan)
				}
			}
		})
	}
}

// TestNoDirectMutationStateStaysSilent is upstream's pass corpus.
//
// Each of the eight is clean for a different reason and together they are the only imported check
// on this port's premises. Reading them in order: a component that only reads state; a local object
// that happens to have a `state` key, which is the ownership check; a file with no component; a
// plain class, which is the component gate; two constructor writes, which are the constructor
// exemption; a nested class inside a constructor, which is the `Class` break in the ancestor walk;
// and a component constructor inside a `describe` callback, which is the exemption surviving two
// intervening call expressions that are outside the class rather than inside the constructor.
func TestNoDirectMutationStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"ES5 component reading state in render", `var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });`},
		{"a plain object literal that happens to carry a state key", `
          var Hello = createReactClass({
            render: function() {
              var obj = {state: {}};
              obj.state.name = 'foo';
              return <div>Hello {obj.state.name}</div>;
            }
          });
        `},
		{"no component at all", `
           var Hello = 'foo';
           module.exports = {};
         `},
		{"a plain class that is not a component", `
           class Hello {
             getFoo() {
               this.state.foo = 'bar'
               return this.state.foo;
             }
           }
         `},
		{"a constructor writing a nested state property", `
           class Hello extends React.Component {
             constructor() {
               this.state.foo = 'bar'
             }
           }
         `},
		{"the same constructor write with a numeric right side", `
        class Hello extends React.Component {
          constructor() {
            this.state.foo = 1;
          }
        }
      `},
		{"a nested class declaration inside a component constructor", `
       class OneComponent extends Component {
         constructor() {
           super();
           class AnotherComponent extends Component {
             constructor() {
               super();
             }
           }
           this.state = {};
         }
       }
     `},
		{"a component constructor inside a describe block", `
     describe('Component spec', () => {
        it('should apply default props on rerender', () => {
          class Outer extends Component {
            constructor() {
              super();
              this.state = { i: 1 };
            }
          }
        });
     });
`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoDirectMutationState, directMutationStateFile, testCase.sourceText))
		})
	}
}

// TestNoDirectMutationStateMeasuredAgainstTheReleaseBinary covers the discriminations upstream's
// corpus does not reach.
//
// Every case here was run through `~/Projects/system/oxc/target/release/oxlint` with
// `{"plugins":["react"],"rules":{"react/no-direct-mutation-state":"error"}}` before it was written,
// so the expectation is a measurement rather than a reading of the Rust. That mattered: three of
// these fall the opposite way from what the rule's name predicts, and two of them are parenthesis
// forms no fixture in either corpus writes.
func TestNoDirectMutationStateMeasuredAgainstTheReleaseBinary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantSpans []string
	}{
		// --- the shape of a write ---
		{"a compound assignment operator", `
class A extends React.Component {
  m() { this.state.x += 1; }
}`, []string{"this.state.x"}},
		{"a logical assignment operator", `
class A extends React.Component {
  m() { this.state.x ??= 1; }
}`, []string{"this.state.x"}},
		{"a prefix decrement reports the whole expression", `
class A extends React.Component {
  m() { --this.state.x; }
}`, []string{"--this.state.x"}},
		// A comparison shares our KindBinaryExpression with an assignment and is the only thing
		// separating them, so this is the fixture that makes the operator check load-bearing.
		{"a comparison against state is not a write", `
class A extends React.Component {
  m() { if (this.state.x === 1) { return; } }
}`, nil},
		// `delete` is a genuine mutation of state that upstream does not see, because it is neither
		// an assignment nor an update expression. Reproduced rather than improved on.
		// Our KindPrefixUnaryExpression carries `!`, `-`, `+` and `~` as well as `++` and `--`,
		// while oxc's UpdateExpression is only the last two. So the prefix arm needs an operator
		// check that the postfix arm does not, and without it every read of state under a negation
		// reports. Found by a mutation that widened the arm to every operator and survived: the
		// `delete` case below cannot see it, because `delete` is its own node kind rather than a
		// prefix unary. Ground truth checked on the release binary, which is silent on all three.
		{"a prefix unary that is not an update is not a write", `
class A extends React.Component {
  m() {
    if (!this.state.x) { return; }
    var y = -this.state.n;
    var z = ~this.state.q;
  }
}`, nil},
		{"delete is silent, which is upstream being narrow", `
class A extends React.Component {
  m() { delete this.state.x; }
}`, nil},
		{"a destructuring target is silent", `
class A extends React.Component {
  m() { [this.state.x] = [1]; }
}`, nil},
		{"an object destructuring target is silent", `
class A extends React.Component {
  m() { ({ y: this.state.x } = {}); }
}`, nil},
		{"a mutating method call is silent", `
class A extends React.Component {
  m() { this.state.list.push(1); }
}`, nil},
		{"a local alias is not tracked", `
class A extends React.Component {
  m() { const s = this.state; s.x = 1; }
}`, nil},

		// --- where the chain is rooted ---
		{"a computed hop above a dotted root reports", `
class A extends React.Component {
  m() { this.state["x"] = 1; }
}`, []string{`this.state["x"]`}},
		// The mirror of the case above and the one a uniform treatment of computed access gets
		// wrong. The root is an element access, so the static-member requirement rejects it before
		// the subscript is compared, even though that subscript is literally "state".
		{"a computed root is silent even when it names state", `
class A extends React.Component {
  m() { this["state"].x = 1; }
}`, nil},
		{"a receiver that is not this", `
class A extends React.Component {
  m() { that.state.x = 1; }
}`, nil},
		{"state reached through another property", `
class A extends React.Component {
  m() { this.props.state.x = 1; }
}`, nil},
		{"a property named state on this that is not the root", `
class A extends React.Component {
  m() { this.a.state = 1; }
}`, nil},

		// --- parentheses, inconsistent within this one rule ---
		// The outer parens are dropped from oxc's tree before the rule runs, so the span excludes
		// them. Ours keeps them, so the rule skips them explicitly at the target and the span
		// matches only because of that.
		{"parens around the whole assignment target report the inner span", `
class A extends React.Component {
  m() { (this.state.x) = 1; }
}`, []string{"this.state.x"}},
		// And the update arm keeps them, because the reported node is the expression rather than
		// the operand. The two lines above and below are the same parenthesization reported two
		// different widths, which no message-id assertion can see.
		{"parens around an update operand report the whole expression", `
class A extends React.Component {
  m() { (this.state.x)++; }
}`, []string{"(this.state.x)++"}},
		{"parens around this are silent", `
class A extends React.Component {
  m() { (this).state.x = 1; }
}`, nil},
		{"parens around this.state are silent", `
class A extends React.Component {
  m() { (this.state).x = 1; }
}`, nil},
		{"parens nested inside the chain are silent", `
class A extends React.Component {
  m() { ((this).state).x = 1; }
}`, nil},

		// --- the constructor exemption and its escape hatch ---
		// Upstream's failing case is a callback and reads as "it runs later". This one is an
		// argument evaluated synchronously during construction and reports anyway, because the
		// latch is set by any call expression ancestor. Upstream being crude, reproduced.
		{"a call in argument position defeats the constructor exemption", `
class A extends React.Component {
  constructor() { foo(this.state.x = 1); }
}`, []string{"this.state.x"}},
		{"a constructor write with no call is exempt", `
class A extends React.Component {
  constructor() { this.state.x = 1; }
}`, nil},
		{"the exemption does not extend to other methods", `
class A extends React.Component {
  m() { foo(this.state.x = 1); }
}`, []string{"this.state.x"}},

		// --- the component gate ---
		// The shelf's IsEs5ComponentCall accepts both of these and oxc accepts neither. Following
		// the helper's name rather than its body would report here, and no imported fixture could
		// see it.
		{"createClass is not a component to this rule", `
var H = createClass({ m: function() { this.state.x = 1; } });`, nil},
		{"React.createClass is not a component to this rule", `
var H = React.createClass({ m: function() { this.state.x = 1; } });`, nil},
		{"React.createReactClass is", `
var H = React.createReactClass({ m: function() { this.state.x = 1; } });`, []string{"this.state.x"}},
		{"a class extending the bare Component", `
class A extends Component {
  m() { this.state.x = 1; }
}`, []string{"this.state.x"}},
		{"a class extending PureComponent", `
class A extends React.PureComponent {
  m() { this.state.x = 1; }
}`, []string{"this.state.x"}},
		// A component assigned to a variable is a class expression rather than a declaration, and
		// both kinds have to break the walk and both have to be tested for componenthood.
		{"a component written as a class expression", `
var A = class extends React.Component {
  m() { this.state.x = 1; }
};`, []string{"this.state.x"}},
		// The walk breaks on a class, so an inner plain class hides its own writes even though a
		// real component encloses it. Upstream ships the constructor version of this as a passing
		// case; this is the method version.
		{"a plain class nested in a component method", `
class Outer extends React.Component {
  m() {
    class Inner { q() { this.state.x = 1; } }
  }
}`, nil},
		// But the break is on a class only, so a call expression does not stop the walk and an ES5
		// component created inside a real component still reports.
		{"an ES5 component nested inside a component method reports", `
class Outer extends React.Component {
  m() {
    var o = createReactClass({ r: function() { this.state.x = 1; } });
  }
}`, []string{"this.state.x"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDirectMutationState, directMutationStateFile, testCase.source)

			if len(testCase.wantSpans) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			wantIds := make([]string, len(testCase.wantSpans))
			for index := range wantIds {
				wantIds[index] = "noDirectMutationState"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				if index >= len(result.Diagnostics) {
					break
				}
				diagnostic := result.Diagnostics[index]
				if got := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != wantSpan {
					t.Errorf("finding %d reported %q, want %q", index, got, wantSpan)
				}
			}
		})
	}
}

// TestNoDirectMutationStateReportsInNonJsxFiles pins the ABSENCE of a source-type gate.
//
// oxc's `should_run` returns false for a file the parser does not read as JSX, and that gate came
// along with the port from oxc. eslint-plugin-react, the authority this rule is ported against,
// has no such gate, and a React class in a `.ts` file is ordinary and legal. This case previously
// expected silence, which locked the gate in rather than catching it.
//
// The `.tsx` control below is kept: it is what makes this a measurement of the suffix rather than
// of the rule going silent for some other reason.
func TestNoDirectMutationStateReportsInNonJsxFiles(t *testing.T) {
	t.Parallel()

	const source = `
class A extends React.Component {
  m() { this.state.x = 1; }
}`
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDirectMutationState, "/repository/source/Plain.ts", source),
		"noDirectMutationState")

	// The control: the identical source in a JSX file also reports, so both suffixes agree.
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDirectMutationState, directMutationStateFile, source), "noDirectMutationState")
}

// TestNoDirectMutationStateMessage asserts the rendered message exactly.
//
// Equality rather than `strings.Contains`, because a predicate weaker than the property it guards
// is not a guard: a description carrying a doubled word or a stray interpolation would satisfy a
// substring check while being wrong on screen.
func TestNoDirectMutationStateMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDirectMutationState, directMutationStateFile, `
class A extends React.Component {
  m() { this.state.x = 1; }
}`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noDirectMutationState" {
		t.Fatalf("message id = %q, want %q", got, "noDirectMutationState")
	}
	if got := result.Diagnostics[0].Message.Description; got != messageNoDirectMutationState.Description {
		t.Fatalf("message description = %q, want %q", got, messageNoDirectMutationState.Description)
	}
}
