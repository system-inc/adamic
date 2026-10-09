package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// findDOMNodeFile is where the fixtures pretend to live.
//
// A `.tsx` extension because most of the imported cases render JSX from a component method, which
// would not parse under a plain `.ts`. The rule itself does not gate on the extension, which was
// measured rather than assumed: oxc's `run` never asks `source_type()`, and a bare `findDOMNode(x)`
// in a `.js` file reports on oxlint the same as in a `.tsx`.
const findDOMNodeFile = "/repository/source/FindDomNode.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two upstream tests below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_find_dom_node.rs`: 5 pass, 6 fail. The extractor
// reported one tester block, no discrepancy, and 6 diagnostics from those 6 fail inputs, so one
// finding per input is measured rather than assumed. The 11 "with options" cases the extractor
// counts are the tester's trailing `None, None, None` tuple slots, not real options; the rule reads
// no configuration and has no `Decode`. One fail case carries `Some(PathBuf::from("demo.ts"))` as
// its third slot, which is a filename rather than an option, and the rule's answer does not depend
// on it.
//
// The strings below were written to disk by a script reading the extractor's own dump rather than
// transcribed, then compared byte against byte. None of the eleven contains a backslash or a
// backtick, so a Go raw string carries each one unchanged; that was checked rather than assumed,
// because the tool writing a fixture is exactly what has cooked escapes for three previous porters.
func TestNoFindDOMNodeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"React.findDOMNode in createReactClass componentDidMount", `
            var Hello = createReactClass({
              componentDidMount: function() {
                React.findDOMNode(this).scrollIntoView();
              },
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
		{"ReactDOM.findDOMNode in createReactClass componentDidMount", `
            var Hello = createReactClass({
              componentDidMount: function() {
                ReactDOM.findDOMNode(this).scrollIntoView();
              },
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
		{"ReactDom.findDOMNode in createReactClass componentDidMount", `
            var Hello = createReactClass({
              componentDidMount: function() {
                ReactDom.findDOMNode(this).scrollIntoView();
              },
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
		{"a bare findDOMNode call in a class method", `
            class Hello extends Component {
              componentDidMount() {
                findDOMNode(this).scrollIntoView();
              }
              render() {
                return <div>Hello</div>;
              }
            }
            `},
		{"a bare findDOMNode call assigned to a field", `
            class Hello extends Component {
              componentDidMount() {
                this.node = findDOMNode(this);
              }
              render() {
                return <div>Hello</div>;
              }
            }
            `},
		{"ReactDOM.findDOMNode after a default import", `
            import ReactDOM from 'react-dom';
            class Demo extends React.Component {
              foo() {
                ReactDOM.findDOMNode(this);
              }
            }
            `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, testCase.sourceText), "noFindDOMNode")
		})
	}
}

// The clean cases upstream ships, each declining for a different reason.
//
// The third is the sharp one: `this.someFunc = React.findDOMNode` *reads* the member without
// calling it, so a rule anchored on the member access rather than on the call reports it and passes
// nothing else here. The fifth is the other discrimination that matters, `SomeModule.findDOMNode`,
// which pins that the receiver is checked against a fixed set of three names rather than accepted
// as any namespace.
func TestNoFindDOMNodeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plain function with no React in it", `var Hello = function() {};`},
		{"createReactClass with only a render method", `
            var Hello = createReactClass({
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
		{"React.findDOMNode read without being called", `
            var Hello = createReactClass({
              componentDidMount: function() {
                someNonMemberFunction(arg);
                this.someFunc = React.findDOMNode;
              },
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
		{"a different React member being called", `
            var Hello = createReactClass({
              componentDidMount: function() {
                React.someFunc(this);
              },
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
		{"findDOMNode called on some other module", `
            var Hello = createReactClass({
              componentDidMount: function() {
                SomeModule.findDOMNode(this).scrollIntoView();
              },
              render: function() {
                return <div>Hello</div>;
              }
            });
            `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not ship, each measured on the release oxlint binary rather than reasoned.
//
// The imported corpus writes exactly two shapes, a bare call and a dotted call on one of three
// receivers, both unparenthesized and both in a component method. So it protects almost none of the
// discriminations the rule makes, and everything below exists because a mutation or a probe showed
// the imports could not see it.
func TestNoFindDOMNodeFiresOnCasesUpstreamDoesNotShip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// Parentheses are skipped on the callee and on the receiver, which is the opposite of the
		// sibling `no-is-mounted` and is upstream's behavior because its accessors unwrap. All
		// three of these report on oxlint.
		{"a parenthesized callee", `(findDOMNode)(this);`},
		{"a parenthesized receiver", `(ReactDOM).findDOMNode(this);`},
		{"a doubly parenthesized receiver", `((ReactDOM)).findDOMNode(this);`},

		// Optional chaining in either position. Neither changes the callee's kind in our tree, and
		// both report upstream.
		{"an optional member call", `ReactDOM?.findDOMNode(this);`},
		{"an optional bare call", `findDOMNode?.(this);`},

		// A subscript is a static property name to oxc, for a string and for a template alike.
		// Nothing in the corpus writes one.
		{"a string subscript", `ReactDOM["findDOMNode"](this);`},
		{"a template subscript", "ReactDOM[`findDOMNode`](this);"},

		// The rule resolves nothing, so a local binding that has no relation to React reports on
		// its spelling alone. This is the case that pins the rule as syntactic: a port asking the
		// checker where the callee came from would decline it.
		{"a shadowing parameter", `function f(findDOMNode) { findDOMNode(this); }`},

		// No component and no method around it. Unlike `no-is-mounted` in this package, this rule
		// has no containment test at all, and every imported fail case sits inside a component,
		// so nothing there can catch a port that added one.
		{"a call at the top level of a file", `findDOMNode(document.body);`},

		// The one imported case carrying a real import is testing the call, not the import. With
		// the import line removed it reports identically.
		{"the imported demo case without its import", `
            class Demo extends React.Component {
              foo() {
                ReactDOM.findDOMNode(this);
              }
            }
            `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, testCase.sourceText), "noFindDOMNode")
		})
	}
}

// The silent cases upstream does not ship. Each was run on oxlint and produced zero findings.
func TestNoFindDOMNodeStaysSilentOnCasesUpstreamDoesNotShip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The receiver set is closed and exactly cased. `reactdom` is the mutation that a corpus
		// with three accepted receivers and no rejected one cannot catch, and `ReactNative` is
		// there so widening the set to any identifier starting with `React` also fails.
		{"a lowercased receiver", `reactdom.findDOMNode(this);`},
		{"a receiver outside the set", `ReactNative.findDOMNode(this);`},
		{"an unrelated module", `SomeModule.findDOMNode(this);`},

		// The receiver must be the identifier itself, not something ending in it. Upstream's
		// `is_specific_id` asks about the node.
		{"a nested receiver", `a.ReactDOM.findDOMNode(this);`},
		{"a receiver returned by a call", `getReactDOM().findDOMNode(this);`},

		// A subscript holding a variable names some other property.
		{"a variable subscript", `ReactDOM[k](this);`},

		// Read without being called, in the bare form. The corpus has the member form of this and
		// not this one, and a rule anchored on the identifier rather than on the call reports it.
		{"a bare reference never called", `var f = findDOMNode;`},

		// `new` is a different node kind upstream, so this is silent. Upstream being narrow, and
		// reproduced deliberately.
		{"a construction rather than a call", `new findDOMNode(this);`},

		// A near-miss name, so a port comparing a prefix rather than the whole text fails here.
		{"a similarly named function", `findDOMNodes(this); findDOMNode2(this); myFindDOMNode(this);`},

		// A private name carries its hash in `Text()`, so it cannot equal the bare name.
		{"a private member", `class C { #findDOMNode() {} m() { this.#findDOMNode(); } }`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, testCase.sourceText))
		})
	}
}

// The finding points at the name and never at the call, which `ExpectFindings` cannot see.
//
// The subscript rows are the reason this test is not a formality. `static_property_info` hands back
// the *literal node's* span, so `ReactDOM["findDOMNode"](this)` underlines thirteen bytes including
// the quotes rather than the eleven of the cooked name. A port reporting the name inside the quotes
// would be off by one on each side, and every assertion in the two tests above would still pass.
func TestNoFindDOMNodeReportsTheNameNode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a bare call", `findDOMNode(this).scrollIntoView();`, `findDOMNode`},
		{"a React member", `React.findDOMNode(this);`, `findDOMNode`},
		{"a ReactDOM member", `ReactDOM.findDOMNode(this);`, `findDOMNode`},
		{"a ReactDom member", `ReactDom.findDOMNode(this);`, `findDOMNode`},
		{"a parenthesized callee", `(findDOMNode)(this);`, `findDOMNode`},
		{"a parenthesized receiver", `(ReactDOM).findDOMNode(this);`, `findDOMNode`},
		{"an optional member call", `ReactDOM?.findDOMNode(this);`, `findDOMNode`},
		{"an optional bare call", `findDOMNode?.(this);`, `findDOMNode`},
		{"a string subscript keeps its quotes", `ReactDOM["findDOMNode"](this);`, `"findDOMNode"`},
		{"a template subscript keeps its backticks", "ReactDOM[`findDOMNode`](this);", "`findDOMNode`"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noFindDOMNode")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Fatalf("reported span = %q, want %q", reported, testCase.want)
			}
		})
	}
}

// Several violations in one file report once each, at their own name.
//
// Every imported fail case carries exactly one finding, so nothing there can catch a rule that
// reports the first violation and stops, or one that reports per file.
func TestNoFindDOMNodeReportsEachCallSeparately(t *testing.T) {
	t.Parallel()

	const sourceText = `class C {
                a() { findDOMNode(this); }
                b() { ReactDOM.findDOMNode(this); }
                c() { React.findDOMNode(this); }
            }`

	result := rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noFindDOMNode", "noFindDOMNode", "noFindDOMNode")
	for index, diagnostic := range result.Diagnostics {
		reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != "findDOMNode" {
			t.Fatalf("finding %d reported %q, want %q", index, reported, "findDOMNode")
		}
	}
}

// A nested call reports at both levels, so the outer match does not consume the inner one.
func TestNoFindDOMNodeReportsANestedCall(t *testing.T) {
	t.Parallel()

	const sourceText = `findDOMNode(ReactDOM.findDOMNode(this));`

	result := rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noFindDOMNode", "noFindDOMNode")
}

// The rendered message is asserted exactly rather than by substring, because a fixture whose
// predicate is weaker than the property it guards is not a guard.
func TestNoFindDOMNodeMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoFindDOMNode, findDOMNodeFile, `findDOMNode(this);`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noFindDOMNode" {
		t.Fatalf("message id = %q, want %q", got, "noFindDOMNode")
	}
	if got := result.Diagnostics[0].Message.Description; got != messageNoFindDOMNode.Description {
		t.Fatalf("message description = %q, want %q", got, messageNoFindDOMNode.Description)
	}
}

// The rule declines the type checker, and this pins that the plain harness is the right one.
//
// A later revert adding `NeedsTypeChecker` would make every silent case above pass vacuously under
// `rule_testing.Run`, since the plain harness hands a checker-declaring rule a nil checker and the rule
// goes completely quiet. Asserting the flag directly fails loudly instead. The rule genuinely does
// not need it: oxc's `run` reads only the callee's syntax, and the measured behavior on a shadowing
// parameter and on an aliased import proves no resolution happens.
func TestNoFindDOMNodeDoesNotNeedTheTypeChecker(t *testing.T) {
	t.Parallel()

	if NoFindDOMNode.NeedsTypeChecker {
		t.Fatal("no-find-dom-node declares the type checker; it decides on syntax alone")
	}
	if NoFindDOMNode.Name != "react/no-find-dom-node" {
		t.Fatalf("rule name = %q, want %q", NoFindDOMNode.Name, "no-find-dom-node")
	}
}
