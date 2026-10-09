package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// isMountedFile is where the fixtures pretend to live.
//
// A `.tsx` extension because upstream gates the whole rule on `source_type().is_jsx()` and two of
// the six imported cases contain JSX in their render method, which would not parse otherwise.
const isMountedFile = "/repository/source/IsMounted.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two upstream tests below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_is_mounted.rs`: 3 pass, 3 fail, and the snapshot records
// 3 diagnostics from those 3 inputs, so one finding per input is measured rather than assumed. The
// extractor reported no discrepancy, one tester block, and no case carrying options.
//
// Three fail cases is the smallest corpus in this namespace, and all three are the same shape:
// `if (!this.isMounted()) { return; }` inside a method. So upstream protects this port barely at
// all, and every discrimination the rule actually makes is pinned by an invented case below
// instead. Each of those says why it exists and what measurement produced it.
func TestNoIsMountedFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"createReactClass componentDidUpdate", `
            var Hello = createReactClass({
                componentDidUpdate: function() {
                  if (!this.isMounted()) {
                    return;
                  }
                },
                render: function() {
                  return <div>Hello</div>;
                }
            });
            `},
		{"createReactClass someMethod", `
            var Hello = createReactClass({
                someMethod: function() {
                  if (!this.isMounted()) {
                    return;
                  }
                },
                render: function() {
                  return <div onClick={this.someMethod.bind(this)}>Hello</div>;
                }
            });
            `},
		{"a class extending React.Component", `
            class Hello extends React.Component {
                someMethod() {
                  if (!this.isMounted()) {
                    return;
                  }
                }
                render() {
                  return <div onClick={this.someMethod.bind(this)}>Hello</div>;
                }
            };
            `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoIsMounted, isMountedFile, testCase.sourceText), "noIsMounted")
		})
	}
}

// The clean cases upstream ships, each failing for a different reason.
//
// The first has no call and no object at all. The second is the sharp one: it *reads*
// `this.isMounted` without calling it, so a rule matching the member access rather than the call
// reports it. The third calls a differently-named method, catching a rule that matched on a
// substring or on the object alone.
func TestNoIsMountedStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a function assigned to a variable", `
            var Hello = function() {
            };
        `},
		{"reading isMounted without calling it", `
            var Hello = createReactClass({
                componentDidUpdate: function() {
                    someNonMemberFunction(arg);
                    this.someFunc = this.isMounted;
                },
                render: function() {
                    return <div>Hello</div>;
                }
            });
            `},
		{"a differently named method", `
            class Hello extends React.Component {
                notIsMounted() {}
                render() {
                    this.notIsMounted();
                    return <div>Hello</div>;
                }
            };
            `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoIsMounted, isMountedFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not ship, each pinning a judgment its three fail inputs cannot see.
//
// Every one of these was run against the release oxlint binary at
// `~/Projects/system/oxc/target/release/oxlint` before being written down, so each asserts a
// measured upstream behavior rather than a belief about one. The imported corpus writes all three
// of its findings inside a real React component, which means the single most likely way to get this
// rule wrong, gating it on component membership, passes the entire import.
func TestNoIsMountedFiresOnCasesUpstreamDoesNotShip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// This is the rule's central trap and the reason it is not a component rule. The class
		// extends nothing and has no relation to React, and `EnclosingComponent` answers nil for
		// it, yet oxlint reports here. A port gated on component membership passes all six
		// imported cases and fails this one.
		{"a class extending nothing", "class Whatever { m() { this.isMounted(); } }"},

		// The same trap on the ES5 side: a bare object literal never passed to `createReactClass`.
		// `IsEs5ComponentCall` answers false, oxlint reports.
		{"a plain object literal", "var o = { m: function() { this.isMounted(); } };"},

		// Our AST spells an object shorthand method and a class method with one kind,
		// `KindMethodDeclaration`, while oxc spells them `ObjectProperty` and `MethodDefinition`.
		// Neither shorthand form appears upstream, so this pins that the shared arm is reached.
		{"an object shorthand method", "var o = { m() { this.isMounted(); } };"},

		// Accessors are `ObjectProperty` and `MethodDefinition` upstream and three separate kinds
		// here. Measured: oxlint reports on all three of these sources.
		{"an object getter", "var o = { get m() { this.isMounted(); return 1; } };"},
		{"an object setter", "var p = { set s(v) { this.isMounted(); } };"},
		{"a class getter", "class C { get q() { this.isMounted(); return 1; } }"},

		// A static method is still a MethodDefinition upstream, and the rule never reads the static
		// flag. Measured reporting.
		{"a static class method", "class W { static m() { this.isMounted(); } }"},

		// The ancestor walk does not stop at the first function boundary, so an arrow nested inside
		// a property still finds the property above it. A port checking only the immediate parent,
		// or bailing at a function, is silent here and green on the whole import.
		{"an arrow inside an object property", "var o = { m: () => { this.isMounted(); } };"},

		// Where oxc and ESLint genuinely disagree, and oxc is followed. ESLint's
		// `'name' in callee.property` declines both of these; `static_property_name()` accepts
		// them. Measured on both tools: ESLint 0 findings, oxlint 1 each.
		{"a computed string subscript", `var o = { m() { this["isMounted"](); } };`},
		{"a computed template subscript", "var o = { m() { this[`isMounted`](); } };"},

		// Optional chaining on either side still reads `isMounted` off `this`. Measured reporting,
		// and this is the one shape ESLint and oxc agree on among the unusual spellings.
		{"an optional member access", "var o = { m() { this?.isMounted(); } };"},
		{"an optional call", "var o = { m() { this.isMounted?.(); } };"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoIsMounted, isMountedFile, testCase.sourceText), "noIsMounted")
		})
	}
}

// Clean cases upstream does not ship, each pinning a boundary the import leaves open.
func TestNoIsMountedStaysSilentOnCasesUpstreamDoesNotShip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The containment gate is real and this is what proves it. Without an object property or
		// method ancestor the rule is silent, so the rule is not simply "any `this.isMounted()`".
		// A port that dropped the ancestor walk entirely passes all six imported cases and fails
		// these two.
		{"a bare function declaration", "function f() { this.isMounted(); }"},
		{"a top level call", "this.isMounted();"},

		// The near miss that looks like it belongs with the methods. A class field holding an arrow
		// is a `PropertyDefinition` upstream, which is absent from oxc's match, and oxlint is
		// silent on exactly this source. Nothing in the import writes a class field, so accepting
		// `KindPropertyDeclaration` would over-report invisibly.
		{"a class field holding an arrow", "class W { m = () => { this.isMounted(); } }"},

		// A variable subscript names whatever that variable holds, not this method. This is the
		// other side of the computed-key partition and keeps the string and template arms from
		// being written as "any element access".
		{"a variable subscript", "var o = { m() { this[isMounted](); } };"},

		// Reached any way other than `this` is out of scope, and both are silent upstream. These
		// pin the documented gap so a later attempt to follow a binding is a deliberate change
		// rather than an accident.
		{"an aliased method", "var o = { m() { var f = this.isMounted; f(); } };"},
		{"a destructured method", "var o = { m() { const { isMounted } = this; isMounted(); } };"},

		// A different object, and a `this` one level removed. The second catches a rule that
		// searched for `this` anywhere in the callee rather than requiring it to be the object.
		{"a different object", "var o = { m() { that.isMounted(); } };"},
		{"this one level removed", "var o = { m() { this.x.isMounted(); } };"},

		// A private field carries its hash in the node's own text, so it can never equal the
		// method name being matched.
		{"a private name", "class W { #isMounted() {} m() { this.#isMounted(); } }"},

		// A computed subscript on some *other* object. This exists because a sweep dropping the
		// `this` guard from the element-access arm survived the whole suite: every subscript case
		// above writes `this`, so nothing could see the guard disappear. Measured silent on oxlint.
		{"a computed subscript on another object", `var o = { m() { that["isMounted"](); } };`},
		{"a computed subscript one level removed", `var p = { m() { foo.bar["isMounted"](); } };`},

		// Parentheses are the sharp pair, and the rule was written the wrong way round first.
		// `(this).isMounted()` and `(this.isMounted)()` are the same call as the reporting one, and
		// oxlint is silent on both, because oxc matches `ThisExpression` against the object node
		// directly and a parenthesized object is a different node. Measured against a control on
		// the same shape that does report, so these are real absences rather than a bad probe. The
		// first version of this rule skipped parentheses and reported both, which read better and
		// diverged from the gate being replaced.
		{"a parenthesized this", "var o = { m() { (this).isMounted(); } };"},
		{"a parenthesized callee", "var o = { m() { (this.isMounted)(); } };"},

		// Name discrimination in both directions: a prefix and a suffix of the real name.
		{"a name that merely starts the same", "var o = { m() { this.isMount(); } };"},
		{"a name that merely ends the same", "var o = { m() { this.reallyIsMounted(); } };"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoIsMounted, isMountedFile, testCase.sourceText))
		})
	}
}

// The span is asserted separately, because ExpectFindings sees message ids and count and nothing
// else, and this rule's two upstreams point at different places.
//
// ESLint reports `callee`, spanning `this.isMounted`. oxc reports `call_expr.span`, spanning
// `this.isMounted()` with the parentheses. Measured on oxlint for the first source below: offset
// 23, length 16. Following ESLint here would pass every assertion above while underlining the
// wrong text on every finding in the tree.
func TestNoIsMountedReportsTheWholeCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a class method", "class W { m() { this.isMounted(); } }", "this.isMounted()"},
		{"an argument bearing call", "var o = { m() { this.isMounted(1, 2); } };", "this.isMounted(1, 2)"},
		{"a computed subscript", `var o = { m() { this["isMounted"](); } };`, `this["isMounted"]()`},
		{"an optional call", "var o = { m() { this.isMounted?.(); } };", "this.isMounted?.()"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoIsMounted, isMountedFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noIsMounted")
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

// A single input carrying several violations reports once each, and each finding points at its own
// call. The import has exactly one finding per file, so nothing there can catch a rule that reports
// the first violation and stops, or one that reports the enclosing method rather than the call.
func TestNoIsMountedReportsEachCallSeparately(t *testing.T) {
	t.Parallel()

	const sourceText = `class W {
                a() { this.isMounted(); }
                b() { this.isMounted(); }
            }`

	result := rule_testing.Run(t, NoIsMounted, isMountedFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noIsMounted", "noIsMounted")
	for index, diagnostic := range result.Diagnostics {
		reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != "this.isMounted()" {
			t.Fatalf("finding %d reported %q, want %q", index, reported, "this.isMounted()")
		}
	}
}

// The rendered message is asserted exactly rather than by substring, because a fixture whose
// predicate is weaker than the property it guards is not a guard.
func TestNoIsMountedMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoIsMounted, isMountedFile, "class W { m() { this.isMounted(); } }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noIsMounted" {
		t.Fatalf("message id = %q, want %q", got, "noIsMounted")
	}
	if got := result.Diagnostics[0].Message.Description; got != messageNoIsMounted.Description {
		t.Fatalf("message description = %q, want %q", got, messageNoIsMounted.Description)
	}
}
