package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// didMountFile is where the fixtures pretend to live.
//
// A `.tsx` extension because upstream gates the whole rule on `source_type().is_jsx()` and three of
// the twenty-five imported cases contain JSX in a render method, which would not parse otherwise.
// The gate itself is pinned by a case below that writes the same reporting source to a `.ts` name.
const didMountFile = "/repository/source/DidMount.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two tables below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_did_mount_set_state.rs`, checked byte against byte with
// a script rather than by reading: all twenty-five decoded strings were confirmed present as
// substrings of the Rust source, against a control string confirmed absent.
//
// The extractor reported two tester blocks, 12 pass and 7 fail in the first and 2 pass and 4 fail
// in the second, and printed a discrepancy warning: 7 diagnostics against 11 snapshotted fail
// inputs. That warning is an artifact of the extractor reading one snapshot file. This rule
// snapshots twice, `react_no_did_mount_set_state.snap` carrying 7 entries and
// `react_no_did_mount_set_state@disallow_in_func.snap` carrying 4, which is 11 diagnostics from 11
// fail inputs and exactly one finding each. Recovered by counting the location headers in both
// files rather than by assuming the totals lined up.
//
// The second block is the same rule under `["disallow-in-func"]`, which decodes to
// `DisallowInFunc: true`. Three of its four failing cases are byte-identical to cases that PASS in
// the first block, which is the single most useful thing the corpus says about this rule: the
// option is not an edge refinement, it is the switch that moves half the passing set into the
// failing one.

// TestNoDidMountSetStateFires runs the eleven failing cases from both upstream blocks.
//
// One finding each, which the two snapshots measure rather than the fixture assuming. The
// interesting member is "direct call beside a nested one": it holds two `setState` calls, one
// written directly in the lifecycle body and one inside an event callback, and reports exactly
// once by default. A port that ignored the function count would report twice here and still look
// plausible, because the case is in the failing list either way.
func TestNoDidMountSetStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		sourceText     string
		disallowInFunc bool
	}{
		{"createReactClass componentDidMount", `
        var Hello = createReactClass({
          componentDidMount: function() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          },
          render: function() {
            return <div>Hello {this.state.name}</div>;
          }
        });
        `, false},
		{"createReactClass named function expression", `
        var Hello = createReactClass({
          componentDidMount: function componentDidMount() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
        });
        `, false},
		{"class method componentDidMount", `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
        }
        `, false},
		{"class field arrow componentDidMount", `
        class Hello extends React.Component {
          componentDidMount = () => {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
        }
        `, false},
		{"direct call beside a nested one", `
        var Hello = createReactClass({
          componentDidMount: function() {
            this.setState({ data: 1 });
            someClass.onSomeEvent(function(data) {
              this.setState({ data: 2 });
            })
          }
        });
        `, false},
		{"inside an if block", `
        class Hello extends React.Component {
          componentDidMount() {
            if (true) {
              this.setState({ data: 123 });
            }
          }
        }
        `, false},
		{"inside a conditional expression", `
        class Hello extends React.Component {
          componentDidMount() {
            const x = true ? this.setState({ data: 123 }) : null;
          }
        }
        `, false},
		{"direct call under the option", `
            var Hello = createReactClass({
              componentDidMount: function() {
                this.setState({
                  name: this.props.name.toUpperCase()
                });
              }
            });
            `, true},
		{"event callback under the option", `
            var Hello = createReactClass({
              componentDidMount: function() {
                someClass.onSomeEvent(function(data) {
                  this.setState({
                    data: data
                  });
                })
              }
            });
            `, true},
		{"setTimeout arrow under the option", `
            var Hello = createReactClass({
              componentDidMount: function() {
                setTimeout(() => {
                  this.setState({ data: 123 });
                }, 100);
              }
            });
            `, true},
		{"promise then arrow under the option", `
            class Hello extends React.Component {
              componentDidMount() {
                Promise.resolve().then(() => {
                  this.setState({ data: 123 });
                });
              }
            }
            `, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidMountSetState, didMountFile, testCase.sourceText,
				NoMethodSetStateOptions{DisallowInFunc: testCase.disallowInFunc})
			rule_testing.ExpectFindings(t, result, "noDidMountSetState")
		})
	}
}

// TestNoDidMountSetStateStaysSilent runs the fourteen passing cases from both upstream blocks.
//
// Five of the twelve default-mode cases are a `setState` that really is lexically inside
// `componentDidMount`, reached through a callback, a nested function declaration, an arrow argument,
// a `setTimeout` or a promise continuation. Those five are the corpus doing the work: they are what
// a port asking only "is this inside componentDidMount" gets wrong, and three of them reappear as
// failures in the option table above.
func TestNoDidMountSetStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		sourceText     string
		disallowInFunc bool
	}{
		{"render only, no lifecycle", `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
        `, false},
		{"empty componentDidMount", `
        var Hello = createReactClass({
          componentDidMount: function() {}
        });
        `, false},
		{"setState referenced but not called", `
        var Hello = createReactClass({
          componentDidMount: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          }
        });
        `, false},
		{"setState inside an event callback", `
        var Hello = createReactClass({
          componentDidMount: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        });
        `, false},
		{"setState inside a nested function declaration", `
        var Hello = createReactClass({
          componentDidMount: function() {
            function handleEvent(data) {
              this.setState({
                data: data
              });
            }
            someClass.onSomeEvent(handleEvent)
          }
        });
        `, false},
		{"setState inside an arrow passed to a method", `
        class Hello extends React.Component {
          componentDidMount() {
            this.handleEvent(() => {
              this.setState({ data: 123 });
            });
          }
        }
        `, false},
		{"componentDidUpdate on a class", `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({ data: 123 });
          }
        }
        `, false},
		{"componentWillMount on a class", `
        class Hello extends React.Component {
          componentWillMount() {
            this.setState({ data: 123 });
          }
        }
        `, false},
		{"componentDidUpdate on createReactClass", `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            this.setState({ data: 123 });
          }
        });
        `, false},
		{"plain function, no enclosing component", `
        function Hello() {
          this.setState({ data: 123 });
        }
        `, false},
		{"setState inside a setTimeout arrow", `
        var Hello = createReactClass({
          componentDidMount: function() {
            setTimeout(() => {
              this.setState({ data: 123 });
            }, 100);
          }
        });
        `, false},
		{"setState inside a promise then arrow", `
        class Hello extends React.Component {
          componentDidMount() {
            Promise.resolve().then(() => {
              this.setState({ data: 123 });
            });
          }
        }
        `, false},
		{"empty componentDidMount under the option", `
            var Hello = createReactClass({
              componentDidMount: function() {}
            });
            `, true},
		{"render only under the option", `
            var Hello = createReactClass({
              render: function() {
                return <div>Hello {this.props.name}</div>;
              }
            });
            `, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidMountSetState, didMountFile, testCase.sourceText,
				NoMethodSetStateOptions{DisallowInFunc: testCase.disallowInFunc})
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Cases upstream does not ship, each covering a discrimination its corpus leaves untested.
//
// Upstream's twenty-five cases exercise the function count thoroughly and almost nothing else. They
// never write a `setState` on anything but `this`, never write the lifecycle name any way but as a
// plain identifier, never put a `componentDidMount` on an object no React factory constructs, and
// never check that the finding points anywhere in particular. Every case below exists because a
// mutation to the rule survived the imported corpus, or because reading the reference implementation
// showed a branch nothing upstream reaches.
func TestNoDidMountSetStateOwnCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		sourceText     string
		fileName       string
		disallowInFunc bool
		findings       []string
	}{
		// The object gate. oxc requires the member expression's object to be the `this` keyword
		// itself, and nothing in the corpus writes anything else, so every one of these is a branch
		// no upstream case reaches.
		{"setState on another object", `
class Hello extends React.Component {
  componentDidMount() {
    that.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},
		{"setState one level down from this", `
class Hello extends React.Component {
  componentDidMount() {
    this.child.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},
		// Upstream matches on the property name exactly, so a different method on `this` is clean.
		// The corpus never writes one, which means a rule comparing nothing at all would pass it.
		{"a different method on this", `
class Hello extends React.Component {
  componentDidMount() {
    this.forceUpdate({ data: 1 });
  }
}
`, didMountFile, false, nil},

		// The computed member spelling. oxc's `static_property_name()` resolves a string subscript,
		// so this reports upstream and reports here. Nothing in the corpus writes one, and a port
		// handling only the dotted spelling passes all twenty-five imported cases.
		{"computed string subscript", `
class Hello extends React.Component {
  componentDidMount() {
    this["setState"]({ data: 1 });
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},
		// A computed subscript naming a different method. The reporting computed case above is
		// satisfied by a rule that resolves the subscript and then compares nothing, so this is
		// the input that separates resolving from comparing. Added for a surviving mutant that
		// replaced the name comparison in the subscript path with a constant.
		{"computed string subscript naming another method", `
class Hello extends React.Component {
  componentDidMount() {
    this["forceUpdate"]({ data: 1 });
  }
}
`, didMountFile, false, nil},
		// A no-substitution template subscript, which oxc's static_property_name also resolves.
		{"template subscript", `
class Hello extends React.Component {
  componentDidMount() {
    this[` + "`setState`" + `]({ data: 1 });
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},
		// A variable subscript names some other property and cannot be resolved statically, so it
		// is silent at both tools.
		{"computed identifier subscript", `
class Hello extends React.Component {
  componentDidMount() {
    this[setState]({ data: 1 });
  }
}
`, didMountFile, false, nil},

		// The component gate, which upstream pins only with a bare `function Hello()`. This is the
		// sharper version: a `componentDidMount` written correctly on an ordinary object literal
		// that no React factory ever constructs. The walk climbs past the lifecycle property, finds
		// no component, and answers nothing. A port that reported as soon as it found the lifecycle
		// name passes upstream's version of this and fails here.
		{"componentDidMount on a plain object", `
var Hello = {
  componentDidMount: function() {
    this.setState({ data: 1 });
  }
};
`, didMountFile, false, nil},
		// And the class half of the same gate: a class with no heritage clause is not a component.
		{"componentDidMount on a class extending nothing", `
class Hello {
  componentDidMount() {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},
		// A class extending something that is not a React base. `isComponentBase` requires the
		// pragma on the namespaced spelling, and nothing upstream writes a non-React superclass.
		{"class extending an unrelated base", `
class Hello extends Widget.Component {
  componentDidMount() {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},

		// The latch. Upstream's components are all written at module scope, so the count never has
		// a chance to pick up an enclosing function after the lifecycle property is found. Here the
		// whole component is returned from a factory function, which puts one more function scope
		// between the lifecycle method and the top of the file. Without the `inLifecycle` latch the
		// count reaches two and this direct call falls out of the report as though it were nested.
		// Nothing upstream can see this and it is the defect most likely to ship.
		{"component defined inside a factory function", `
function makeHello() {
  return class Hello extends React.Component {
    componentDidMount() {
      this.setState({ data: 1 });
    }
  };
}
`, didMountFile, false, []string{"noDidMountSetState"}},

		// Non-function scopes are transparent to the count. Upstream covers an `if` block and a
		// conditional expression; a loop body and a `try` block are the same judgment and are here
		// because a port could plausibly treat any nested block as a boundary.
		{"inside a loop body", `
class Hello extends React.Component {
  componentDidMount() {
    for (const x of xs) {
      this.setState({ data: x });
    }
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},
		{"inside a try block", `
class Hello extends React.Component {
  componentDidMount() {
    try {
      this.setState({ data: 1 });
    } catch (error) {}
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},

		// Two nested function scopes rather than one. Upstream's option cases all cross exactly one,
		// so a rule comparing the count against the wrong side of the boundary is not distinguished
		// by any of them.
		{"two levels of nesting under the option", `
var Hello = createReactClass({
  componentDidMount: function() {
    outer(function() {
      inner(function() {
        this.setState({ data: 1 });
      });
    });
  }
});
`, didMountFile, true, []string{"noDidMountSetState"}},
		{"two levels of nesting by default", `
var Hello = createReactClass({
  componentDidMount: function() {
    outer(function() {
      inner(function() {
        this.setState({ data: 1 });
      });
    });
  }
});
`, didMountFile, false, nil},

		// A class field holding a plain function expression rather than an arrow. Upstream ships the
		// arrow spelling as a failure and never writes this one, and both are `PropertyDefinition`
		// to oxc, so both report.
		{"class field function expression", `
class Hello extends React.Component {
  componentDidMount = function() {
    this.setState({ data: 1 });
  };
}
`, didMountFile, false, []string{"noDidMountSetState"}},

		// The computed lifecycle key REPORTS. An earlier version of this file asserted silence
		// here and defended it in a doc comment as a deliberate divergence; it was a defect. oxc's
		// `key.static_name()` resolves a computed string key, and the release oxlint binary reports
		// on this exact source. Measured, not reasoned.
		{"computed lifecycle key", `
var Hello = createReactClass({
  ["componentDidMount"]: function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, []string{"noDidMountSetState"}},
		// The plain string key, the same resolution by a different spelling, also measured as
		// reporting. Nothing in the corpus writes either spelling.
		{"string literal lifecycle key", `
var Hello = createReactClass({
  "componentDidMount": function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, []string{"noDidMountSetState"}},
		// A computed key holding a variable resolves to nothing at both tools, which is the line
		// that separates "resolve the computed spelling" from "ignore the brackets".
		{"computed lifecycle key holding a variable", `
var Hello = createReactClass({
  [componentDidMount]: function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, nil},
		// A getter REPORTS. oxc matches `MethodDefinition` without asking the method kind, and a
		// getter is one in its tree. An earlier version excluded accessors on a guess and its doc
		// comment said upstream had no such node, which was wrong. Confirmed on the release binary.
		{"getter named for the lifecycle method", `
class Hello extends React.Component {
  get componentDidMount() {
    this.setState({ data: 1 });
    return 1;
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},
		// And the setter spelling, measured the same way.
		{"setter named for the lifecycle method", `
class Hello extends React.Component {
  set componentDidMount(value) {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},
		// A method the walk merely PASSES THROUGH is a function scope too. This shipped as a false
		// positive: the latch counted the lifecycle method's own scope by hand, and `isFunctionScope`
		// omitted method kinds entirely, so `setState` inside an object method declared in the
		// lifecycle body came out one scope shallow and reported. Silent at the release binary,
		// confirmed with a control that fires on the direct form in the same run. Found by the
		// porter of `no-will-update-set-state`, whose sweep exposed it in their copy of this walk.
		{"a method declared inside the lifecycle body", `
class Hello extends React.Component {
  componentDidMount() {
    const helper = { run() { this.setState({ data: 1 }); } };
    helper.run();
  }
}
`, didMountFile, false, nil},
		// An accessor gets the same count correction a method gets, so a callback inside a getter
		// crosses two scopes and is silent by default. This is the case that fails if accessors are
		// added to the lifecycle kinds without also being added to the count correction.
		{"callback inside a getter, default", `
class Hello extends React.Component {
  get componentDidMount() {
    setTimeout(() => {
      this.setState({ data: 1 });
    }, 100);
    return 1;
  }
}
`, didMountFile, false, nil},
		{"callback inside a getter, under the option", `
class Hello extends React.Component {
  get componentDidMount() {
    setTimeout(() => {
      this.setState({ data: 1 });
    }, 100);
    return 1;
  }
}
`, didMountFile, true, []string{"noDidMountSetState"}},
		// A computed class method name, the class-side twin of the computed object key.
		{"computed class method name", `
class Hello extends React.Component {
  ["componentDidMount"]() {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},

		// The ES5 factory name set, which is NARROWER than the shelf helper this rule calls.
		// `internal/utilities/react.IsEs5ComponentCall` accepts `createClass` as well, and oxc's
		// `CREATE_CLASS` at `oxc_linter/src/utils/react.rs:554` is the single constant
		// `createReactClass`. Both of these are clean on the release binary and a port following
		// the shelf reports on both. Nothing in the corpus writes `createClass`.
		{"createClass is not the factory oxc accepts", `
var Hello = createClass({
  componentDidMount: function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, nil},
		{"React.createClass is not the factory oxc accepts", `
var Hello = React.createClass({
  componentDidMount: function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, nil},
		// The namespaced spelling that IS accepted, so the narrowing above is a name check rather
		// than the member arm being dropped wholesale.
		{"React.createReactClass is accepted", `
var Hello = React.createReactClass({
  componentDidMount: function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, []string{"noDidMountSetState"}},
		// A parenthesized factory reports, because oxc's identifier arm goes through
		// `get_identifier_reference()` which skips parentheses, and our shelf helper skips them
		// too. Measured on the binary; guessing this direction has already gone wrong twice.
		{"parenthesized factory", `
var Hello = (createReactClass)({
  componentDidMount: function() {
    this.setState({ data: 1 });
  }
});
`, didMountFile, false, []string{"noDidMountSetState"}},

		// A parenthesized `this` is SILENT, the opposite answer to the parenthesized factory above,
		// because oxc destructures `Expression::ThisExpression` and a parenthesis is a different
		// variant. Both directions were measured rather than reasoned.
		{"parenthesized this", `
class Hello extends React.Component {
  componentDidMount() {
    (this).setState({ data: 1 });
  }
}
`, didMountFile, false, nil},
		{"parenthesized callee", `
class Hello extends React.Component {
  componentDidMount() {
    (this.setState)({ data: 1 });
  }
}
`, didMountFile, false, nil},

		// The component gate is looser than it looks: nothing requires the lifecycle property to be
		// a member OF the component, only that a component is somewhere further out. A plain object
		// literal declared inside a real component's render method reports on the release binary.
		// Reproduced rather than improved on.
		{"plain object literal nested inside a component", `
class Hello extends React.Component {
  render() {
    var o = {
      componentDidMount: function() {
        this.setState({ data: 1 });
      }
    };
    return o;
  }
}
`, didMountFile, false, []string{"noDidMountSetState"}},

		// No file gate, asserted rather than assumed. Byte-identical to a reporting case above but
		// written to a `.ts` name. This case previously expected silence, locking in a gate carried
		// over from the oxc port; eslint-plugin-react, the authority this rule is ported against,
		// does not gate on the file name at all. A React class in a `.ts` file is ordinary and
		// legal, and the gate hid it across more `.ts` files than the `.tsx` files it could see.
		//
		// Upstream's corpus is entirely `.tsx`, so nothing imported from it can tell a working gate
		// from an absent one. That is exactly why this case is written by hand.
		{"reporting source in a non-JSX file", `
class Hello extends React.Component {
  componentDidMount() {
    this.setState({ data: 1 });
  }
}
`, "/repository/source/DidMount.ts", false, []string{"noDidMountSetState"}},

		// The lifecycle name is compared exactly. `componentDidMountSomething` is a different
		// method and the corpus only ever writes neighbouring lifecycle names, never a prefix of
		// this one, so a port using a prefix comparison passes every imported case.
		{"a method whose name starts with the lifecycle name", `
class Hello extends React.Component {
  componentDidMountLater() {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},
		// And a name of exactly the same length, which is a different mutant from the one above.
		// `componentDidMountLater` is longer, so a rule comparing name lengths rather than names
		// declines it and passes that case; `componentDidPaint` has the same seventeen characters
		// as the real name and is the only input that separates the two. Found by a surviving
		// mutant that replaced the string comparison with a length comparison, not by reading.
		{"a method whose name is the same length as the lifecycle name", `
class Hello extends React.Component {
  componentDidPaint() {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},

		// The sibling rule's subject, kept here deliberately. These two rules differ only in the
		// lifecycle name they watch, so this asserts that this rule is silent on the other one's
		// failing shape rather than trusting that the string constant is right.
		{"componentDidUpdate class field", `
class Hello extends React.Component {
  componentDidUpdate = () => {
    this.setState({ data: 1 });
  }
}
`, didMountFile, false, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoDidMountSetState, testCase.fileName, testCase.sourceText,
				NoMethodSetStateOptions{DisallowInFunc: testCase.disallowInFunc})
			if len(testCase.findings) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestNoDidMountSetStateReportsTheCallee pins where the finding points and what it says.
//
// `ExpectFindings` asserts message ids and a count and nothing else, so every fixture above passes
// unchanged if the rule reports the whole call, the enclosing statement, or the lifecycle method.
// oxc reports `call_expr.callee.span()`, thirteen characters for `this.setState`, which the upstream
// snapshot underlines. The neighbouring `no-is-mounted` in this same package reports its whole call
// including the parentheses, so the two genuinely differ and the difference is invisible to an id
// assertion.
//
// The message text is asserted with an exact prefix rather than with `strings.Contains` on an
// interpolated value, because a containment check on a rendered string passes for a string that has
// the needle plus a defect.
func TestNoDidMountSetStateReportsTheCallee(t *testing.T) {
	t.Parallel()

	sourceText := `
class Hello extends React.Component {
  componentDidMount() {
    this.setState({ data: 123 });
  }
}
`
	result := rule_testing.RunWithOptions(t, NoDidMountSetState, didMountFile, sourceText,
		NoMethodSetStateOptions{})
	rule_testing.ExpectFindings(t, result, "noDidMountSetState")

	diagnostic := result.Diagnostics[0]
	reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
	if reported != "this.setState" {
		t.Errorf("finding points at %q, want %q", reported, "this.setState")
	}

	if !strings.HasPrefix(diagnostic.Message.Description, "`setState` is called in `componentDidMount`.") {
		t.Errorf("message description = %q, want it to open with the finding", diagnostic.Message.Description)
	}
}

// TestNoDidMountSetStateReportsTheCalleeOnASubscript pins the span of the computed spelling too.
//
// A separate case because the callee is a different node kind there, `ElementAccessExpression`
// rather than `PropertyAccessExpression`, and a rule reporting the access's own name node rather
// than the callee would point at four characters in one spelling and sixteen in the other while
// passing an id assertion in both.
func TestNoDidMountSetStateReportsTheCalleeOnASubscript(t *testing.T) {
	t.Parallel()

	sourceText := `
class Hello extends React.Component {
  componentDidMount() {
    this["setState"]({ data: 123 });
  }
}
`
	result := rule_testing.RunWithOptions(t, NoDidMountSetState, didMountFile, sourceText,
		NoMethodSetStateOptions{})
	rule_testing.ExpectFindings(t, result, "noDidMountSetState")

	diagnostic := result.Diagnostics[0]
	reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
	if reported != `this["setState"]` {
		t.Errorf("finding points at %q, want %q", reported, `this["setState"]`)
	}
}
