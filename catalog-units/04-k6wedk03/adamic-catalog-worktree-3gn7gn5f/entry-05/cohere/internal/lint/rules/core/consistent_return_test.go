package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// consistentReturnFile is where the fixtures pretend to live.
const consistentReturnFile = "/repository/source/ConsistentReturn.ts"

// consistentReturnExpectation is one finding upstream asserts, with the name it renders and the
// position it points at.
//
// The name and the position are here because neither is visible to a message-id assertion, and both
// are load-bearing: this rule has three messages whose only difference in a fixture is the rendered
// name, and it computes a different span per node kind rather than reporting the function.
type consistentReturnExpectation struct {
	id     string
	name   string
	line   int
	column int
}

type consistentReturnCase struct {
	source      string
	optionsJson string
	findings    []consistentReturnExpectation
}

// decodeConsistentReturnOptionsForTest routes a case's options through the shipped decoder.
func decodeConsistentReturnOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeConsistentReturnOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding options %q: %v", optionsJson, err)
	}
	return decoded
}

var consistentReturnCleanCases = []consistentReturnCase{
	{source: "function foo() { return; }", optionsJson: ""},
	{source: "function foo() { if (true) return; }", optionsJson: ""},
	{source: "function foo() { if (true) return; else return; }", optionsJson: ""},
	{source: "function foo() { if (true) return true; else return false; }", optionsJson: ""},
	{source: "f(function() { return; })", optionsJson: ""},
	{source: "f(function() { if (true) return; })", optionsJson: ""},
	{source: "f(function() { if (true) return; else return; })", optionsJson: ""},
	{source: "f(function() { if (true) return true; else return false; })", optionsJson: ""},
	{source: "function foo() { function bar() { return true; } return; }", optionsJson: ""},
	{source: "function foo() { function bar() { return; } return false; }", optionsJson: ""},
	{source: "function Foo() { if (!(this instanceof Foo)) return new Foo(); }", optionsJson: ""},
	{source: "function foo() { if (true) return 5; else return undefined; }", optionsJson: ""},
	{source: "function foo() { if (true) return 5; else return void 0; }", optionsJson: ""},
	{source: "function foo() { if (true) return; else return undefined; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}"},
	{source: "function foo() { if (true) return; else return void 0; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}"},
	{source: "function foo() { if (true) return undefined; else return; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}"},
	{source: "function foo() { if (true) return undefined; else return void 0; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}"},
	{source: "function foo() { if (true) return void 0; else return; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}"},
	{source: "function foo() { if (true) return void 0; else return undefined; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}"},
	{source: "var x = () => {  return {}; };", optionsJson: ""},
	{source: "if (true) { return 1; } return 0;", optionsJson: ""},
	{source: "class Foo { constructor() { if (true) return foo; } }", optionsJson: ""},
	{source: "var Foo = class { constructor() { if (true) return foo; } }", optionsJson: ""},
}

var consistentReturnFiringCases = []consistentReturnCase{
	{source: "function foo() { if (true) return true; else return; }", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturnValue", name: "Function 'foo'", line: 1, column: 46}}},
	{source: "var foo = () => { if (true) return true; else return; }", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturnValue", name: "Arrow function", line: 1, column: 47}}},
	{source: "function foo() { if (true) return; else return false; }", optionsJson: "", findings: []consistentReturnExpectation{{id: "unexpectedReturnValue", name: "Function 'foo'", line: 1, column: 41}}},
	{source: "f(function() { if (true) return true; else return; })", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturnValue", name: "Function", line: 1, column: 44}}},
	{source: "f(function() { if (true) return; else return false; })", optionsJson: "", findings: []consistentReturnExpectation{{id: "unexpectedReturnValue", name: "Function", line: 1, column: 39}}},
	{source: "f(a => { if (true) return; else return false; })", optionsJson: "", findings: []consistentReturnExpectation{{id: "unexpectedReturnValue", name: "Arrow function", line: 1, column: 33}}},
	{source: "function foo() { if (true) return true; return undefined; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}", findings: []consistentReturnExpectation{{id: "missingReturnValue", name: "Function 'foo'", line: 1, column: 41}}},
	{source: "function foo() { if (true) return true; return void 0; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}", findings: []consistentReturnExpectation{{id: "missingReturnValue", name: "Function 'foo'", line: 1, column: 41}}},
	{source: "function foo() { if (true) return undefined; return true; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}", findings: []consistentReturnExpectation{{id: "unexpectedReturnValue", name: "Function 'foo'", line: 1, column: 46}}},
	{source: "function foo() { if (true) return void 0; return true; }", optionsJson: "{\"treatUndefinedAsUnspecified\":true}", findings: []consistentReturnExpectation{{id: "unexpectedReturnValue", name: "Function 'foo'", line: 1, column: 43}}},
	{source: "if (true) { return 1; } return;", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturnValue", name: "Program", line: 1, column: 25}}},
	{source: "function foo() { if (a) return true; }", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "function 'foo'", line: 1, column: 10}}},
	{source: "function _foo() { if (a) return true; }", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "function '_foo'", line: 1, column: 10}}},
	{source: "f(function foo() { if (a) return true; });", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "function 'foo'", line: 1, column: 12}}},
	{source: "f(function() { if (a) return true; });", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "function", line: 1, column: 3}}},
	{source: "f(() => { if (a) return true; });", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "arrow function", line: 1, column: 6}}},
	{source: "var obj = {foo() { if (a) return true; }};", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "method 'foo'", line: 1, column: 12}}},
	{source: "class A {foo() { if (a) return true; }};", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "method 'foo'", line: 1, column: 10}}},
	{source: "if (a) return true;", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "program", line: 1, column: 1}}},
	{source: "class A { CapitalizedFunction() { if (a) return true; } }", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "method 'CapitalizedFunction'", line: 1, column: 11}}},
	{source: "({ constructor() { if (a) return true; } });", optionsJson: "", findings: []consistentReturnExpectation{{id: "missingReturn", name: "method 'constructor'", line: 1, column: 4}}},
}

// TestConsistentReturnFires runs upstream's 21 invalid cases.
//
// Each asserts the message id, the rendered name, and the span. The name is asserted because the
// three messages differ by casing of the same word -- upstream applies upperCaseFirst to two of
// them and not to the third -- and a port collapsing the two would pass every id assertion. The
// span is asserted because upstream computes four different locations by node kind, and pointing
// at the whole function instead would also pass every id assertion.
func TestConsistentReturnFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range consistentReturnFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentReturn, consistentReturnFile,
				testCase.source, decodeConsistentReturnOptionsForTest(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.findings))
			for i, finding := range testCase.findings {
				wantIds[i] = finding.id
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for i, finding := range testCase.findings {
				got := result.Diagnostics[i].Message.Description
				// Upstream's three message templates, rendered with the name it asserts.
				var want string
				switch finding.id {
				case "missingReturn":
					want = "Expected to return a value at the end of " + finding.name + "."
				case "missingReturnValue":
					want = finding.name + " expected a return value."
				case "unexpectedReturnValue":
					want = finding.name + " expected no return value."
				default:
					t.Fatalf("unknown message id %q", finding.id)
				}
				if !strings.HasPrefix(got, want) {
					t.Errorf("finding %d: message should begin %q, got %q", i, want, got)
				}

				// The span. Upstream asserts a line and a column; `Run` does not trim its input, so
				// the literal above is byte for byte what is on disk and the offset can be turned
				// into a one-based line and column directly.
				line, column := consistentReturnLineAndColumn(testCase.source,
					result.Diagnostics[i].Range.Pos())
				if line != finding.line || column != finding.column {
					t.Errorf("finding %d should point at line %d column %d, got line %d column %d "+
						"(reported text %q)", i, finding.line, finding.column, line, column,
						testCase.source[result.Diagnostics[i].Range.Pos():result.Diagnostics[i].Range.End()])
				}
			}
		})
	}
}

// consistentReturnLineAndColumn converts a byte offset into upstream's one-based line and column.
//
// Upstream reports a zero-based column internally and the tester asserts it one-based, which is why
// the Program case asserts column 1 for offset 0 while the rule's own `loc` says column 0.
func consistentReturnLineAndColumn(source string, offset int) (int, int) {
	line, column := 1, 1
	for i := 0; i < offset && i < len(source); i++ {
		if source[i] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

// TestConsistentReturnStaysSilent runs upstream's 23 valid cases.
func TestConsistentReturnStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range consistentReturnCleanCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ConsistentReturn,
				consistentReturnFile, testCase.source,
				decodeConsistentReturnOptionsForTest(t, testCase.optionsJson)))
		})
	}
}

// consistentReturnAddedCases are shapes upstream's corpus does not write.
//
// Every verdict and every rendered name below was measured by driving the installed rule with the
// typescript-eslint parser, not derived from reading the helper. The name renderings are the bulk
// of them, because `getFunctionNameWithKind` has nine arms the corpus never exercises and a port
// can get all nine wrong while passing every imported case.
var consistentReturnAddedCases = []struct {
	name        string
	source      string
	optionsJson string
	// message is the whole expected message prefix, or "" when the case is clean.
	message string
	reason  string
}{
	{
		name:    "a static method renders static before method",
		source:  "class A { static foo() { if (a) return true; } }",
		message: "Expected to return a value at the end of static method 'foo'.",
		reason:  "Upstream pushes `static` first, per the class-features proposal ordering.",
	},
	{
		name:    "a private method renders an unquoted hash name",
		source:  "class A { #foo() { if (a) return true; } }",
		message: "Expected to return a value at the end of private method #foo.",
		reason: "A private name is rendered unquoted where an ordinary one is quoted, and " +
			"Text() already carries the hash, so adding one would double it.",
	},
	{
		name:    "static and private compose in that order",
		source:  "class A { static #foo() { if (a) return true; } }",
		message: "Expected to return a value at the end of static private method #foo.",
		reason:  "The one case that pins the order of the two modifier words against each other.",
	},
	{
		name:    "an async function renders async",
		source:  "async function foo() { if (a) return true; }",
		message: "Expected to return a value at the end of async function 'foo'.",
		reason:  "Async is read from the modifier list rather than from a node field here.",
	},
	{
		name:    "a generator renders generator",
		source:  "function* foo() { if (a) return true; }",
		message: "Expected to return a value at the end of generator function 'foo'.",
		reason: "The asterisk is a token on the declaration rather than a modifier, so it needs " +
			"its own accessor per function kind.",
	},
	{
		name:    "async and generator compose in that order",
		source:  "async function* foo() { if (a) return true; }",
		message: "Expected to return a value at the end of async generator function 'foo'.",
		reason:  "Pins async ahead of generator.",
	},
	{
		name:    "a getter renders getter and not method",
		source:  "class A { get foo() { if (a) return true; } }",
		message: "Expected to return a value at the end of getter 'foo'.",
		reason:  "Upstream reads the accessor kind, which our parser gives its own node kind.",
	},
	{
		name:    "a setter renders setter",
		source:  "class A { set foo(v) { if (a) return true; } }",
		message: "Expected to return a value at the end of setter 'foo'.",
		reason:  "The other half of the accessor pair.",
	},
	{
		name:    "an object literal getter renders the same as a class one",
		source:  "var o = { get foo() { if (a) return true; } };",
		message: "Expected to return a value at the end of getter 'foo'.",
		reason:  "Upstream reads the parent's kind field, which is the same for both.",
	},
	{
		name:   "a class constructor is exempt from the end judgment",
		source: "class A { constructor() { if (a) return true; } }",
		reason: "A return in a constructor is a language feature rather than a value channel.",
	},
	{
		name:    "a class constructor is NOT exempt from the per-return judgment",
		source:  "class A { constructor() { if (a) return true; else return; } }",
		message: "Constructor expected a return value.",
		reason: "Only the end-of-function judgment is exempt, which is where upstream's " +
			"isClassConstructor test actually sits; its ReturnStatement listener runs inside a " +
			"constructor like anywhere else. Measured against the installed rule, which reports " +
			"here and renders the name as the bare capitalised word with no modifiers. This is " +
			"also the only case that reaches the constructor arm of the name renderer at all: a " +
			"mutation rewriting that arm survived every other fixture.",
	},
	{
		name:    "a class expression constructor behaves the same way",
		source:  "var A = class { constructor() { if (a) return true; else return; } };",
		message: "Constructor expected a return value.",
		reason:  "The exemption keys on the node kind, which a class expression shares.",
	},
	{
		name:    "the constructor name renders no modifiers",
		source:  "class A { constructor() { if (a) return; else return true; } }",
		message: "Constructor expected no return value.",
		reason: "Upstream returns the bare word and discards whatever modifier tokens it had " +
			"pushed, so a port appending them would render `method constructor` here.",
	},
	{
		name:    "an object literal method named constructor is NOT exempt",
		source:  "({ constructor() { if (a) return true; } });",
		message: "Expected to return a value at the end of method 'constructor'.",
		reason: "This is the pair that shows the exemption keys on the node kind rather than on " +
			"the name. Both halves are measured against the installed rule.",
	},
	{
		name:    "a function expression on a property renders as a method",
		source:  "var o = { foo: function() { if (a) return true; } };",
		message: "Expected to return a value at the end of method 'foo'.",
		reason: "Upstream reads the PARENT's node type, so an anonymous function in a property " +
			"is a method with the property's name. Reading the function alone renders `function` " +
			"with no name, which is wrong twice over.",
	},
	{
		name:    "a class field holding a function renders as a method",
		source:  "class A { foo = function() { if (a) return true; }; }",
		message: "Expected to return a value at the end of method 'foo'.",
		reason:  "The PropertyDefinition arm of the same parent rule.",
	},
	{
		name:    "a quoted key renders its cooked name",
		source:  "var o = { 'a-b': function() { if (a) return true; } };",
		message: "Expected to return a value at the end of method 'a-b'.",
		reason:  "getStaticPropertyName cooks the string, so the quotes in the source do not double.",
	},
	{
		name:    "a computed key whose value is static renders it",
		source:  "class A { ['computed']() { if (a) return true; } }",
		message: "Expected to return a value at the end of method 'computed'.",
		reason:  "property.Name reads through the brackets for a literal.",
	},
	{
		name:    "a computed key whose value is not static renders no name",
		source:  "class A { [x]() { if (a) return true; } }",
		message: "Expected to return a value at the end of method.",
		reason: "The message ends at the kind with no name. property.Name refuses a bare " +
			"identifier inside brackets, which is exactly the answer needed here, and rendering " +
			"the variable's spelling would invent a name the code does not have.",
	},
	{
		name:    "an async arrow renders async arrow function",
		source:  "const f = async () => { if (a) return true; };",
		message: "Expected to return a value at the end of async arrow function.",
		reason:  "An arrow pushes `arrow` before `function`, after the async token.",
	},
	{
		name:   "a trailing throw makes the end unreachable",
		source: "function foo() { if (a) { return true; } throw new Error(); }",
		reason: "This is the control-flow half of the rule. Syntactically the function has one " +
			"return and no other, so a port without a graph reports it. Measured silent upstream.",
	},
	{
		name:   "an else-throw makes the end unreachable",
		source: "function foo() { if (a) return true; else throw 1; }",
		reason: "Same judgment reached through a different shape.",
	},
	{
		name:   "an infinite loop makes the end unreachable",
		source: "function foo() { while(true) { return 1; } }",
		reason: "The graph has to know `while (true)` never exits normally. Measured agreeing " +
			"with upstream, which is the strongest single check that EndReachable is the right " +
			"question rather than a near-miss.",
	},
	{
		name:    "a call that never returns does not make the end unreachable",
		source:  "function foo() { if (a) return 1; process.exit(); }",
		message: "Expected to return a value at the end of function 'foo'.",
		reason: "The complement of the three above: neither upstream nor the graph models a " +
			"never-returning call, so this reports. Included so the three clean cases cannot be " +
			"passed by a rule that simply never reports the end judgment.",
	},
	{
		name:    "a union return type does not exempt the function",
		source:  "function foo(): number | undefined { if (a) return true; }",
		message: "Expected to return a value at the end of function 'foo'.",
		reason: "Upstream never consults a type annotation. Filtering on the declared return type " +
			"is what @typescript-eslint/consistent-return adds on top of this rule, so narrowing " +
			"here would change what that wrapper wraps. Measured reporting upstream.",
	},
	{
		name:   "an ES5 constructor is exempt by its capitalised name",
		source: "function Foo() { if (a) return new Foo(); }",
		reason: "isES5Constructor keys on the function's own name starting upper case.",
	},
	{
		name:    "a leading underscore is not a capital",
		source:  "function _foo() { if (a) return true; }",
		message: "Expected to return a value at the end of function '_foo'.",
		reason: "Upstream compares the first character against its own lower case rather than " +
			"testing an ASCII range, and `_` equals its own lower case, so it is not a " +
			"constructor. This case is in the corpus and it is the one that separates the two " +
			"readings, so it is repeated here with the reasoning attached.",
	},
	{
		name:    "a capitalised METHOD is not an ES5 constructor",
		source:  "class A { CapitalizedFunction() { if (a) return true; } }",
		message: "Expected to return a value at the end of method 'CapitalizedFunction'.",
		reason: "The exemption is about a function declaration or expression, not about " +
			"capitalisation anywhere. Also in the corpus, repeated here for the pairing.",
	},
}

// TestConsistentReturnAddedCases covers what upstream's corpus does not write.
func TestConsistentReturnAddedCases(t *testing.T) {
	t.Parallel()

	for _, testCase := range consistentReturnAddedCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentReturn, consistentReturnFile,
				testCase.source, decodeConsistentReturnOptionsForTest(t, testCase.optionsJson))

			if testCase.message == "" {
				if len(result.Diagnostics) != 0 {
					t.Fatalf("expected no findings, got %d %v (%s)",
						len(result.Diagnostics), result.MessageIds(), testCase.reason)
				}
				return
			}
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d %v (%s)",
					len(result.Diagnostics), result.MessageIds(), testCase.reason)
			}
			if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, testCase.message) {
				t.Errorf("message should begin %q, got %q (%s)", testCase.message, got, testCase.reason)
			}
		})
	}
}

// TestConsistentReturnDecoderRoundTrip pins the option surface through the shipped decoder.
func TestConsistentReturnDecoderRoundTrip(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		optionsJson string
		want        bool
	}{
		{"an empty object", "{}", false},
		{"an explicit false", `{"treatUndefinedAsUnspecified":false}`, false},
		{"an explicit true", `{"treatUndefinedAsUnspecified":true}`, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeConsistentReturnOptions(json.RawMessage(testCase.optionsJson))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.optionsJson, err)
			}
			if got := decoded.(ConsistentReturnSettings).TreatUndefinedAsUnspecified; got != testCase.want {
				t.Errorf("treatUndefinedAsUnspecified should be %v, got %v", testCase.want, got)
			}
		})
	}

	// A rule configured as a bare severity is handed nil options rather than a struct. Asserted by
	// bypassing the decoder entirely, which is the path no other fixture here takes.
	result := rule_testing.RunWithOptions(t, ConsistentReturn, consistentReturnFile,
		"function foo() { if (true) return true; else return; }", nil)
	rule_testing.ExpectFindings(t, result, "missingReturnValue")
}

// TestConsistentReturnSpanPointsAtTheKeyNotTheFirstToken pins the method arm of the report range.
//
// A mutation replacing the key span with the whole method node survived every fixture here,
// including upstream's four method cases at columns 12, 10, 11 and 4. The reason is that
// `TokenRange` trims leading trivia, so for `class A { foo() {} }` the method node's first token IS
// the key and the two spans start at the same byte. The corpus writes no method carrying a
// modifier, so nothing in it can separate them.
//
// A modifier does separate them. Measured against the installed rule: `class A { static foo() }`
// reports at column 18, which is `foo`, while the method node begins at column 11 on `static`.
// Same for `get`, `async` and the generator asterisk.
func TestConsistentReturnSpanPointsAtTheKeyNotTheFirstToken(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		// wantText is the source the finding's own range covers.
		wantText string
	}{
		{"class A { static foo() { if (a) return true; } }", "foo"},
		{"class A { get foo() { if (a) return true; } }", "foo"},
		{"class A { async foo() { if (a) return true; } }", "foo"},
		{"class A { *foo() { if (a) return true; } }", "foo"},
		{"var o = { async foo() { if (a) return true; } };", "foo"},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistentReturn, consistentReturnFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "missingReturn")
			reported := testCase.source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Errorf("finding should point at %q, got %q", testCase.wantText, reported)
			}
		})
	}
}

// TestConsistentReturnOverReportsOnATryWithADeclarationAndAFinally pins a known divergence.
//
// This case reports here and is silent in the installed rule, and the cause is in the shared
// control-flow shelf rather than in this rule: `EndReachable` answers true for a `try` holding a
// variable declaration when the statement also has a `finally`, even though every path through it
// returns or throws. Removing either the declaration or the `finally` gives the right answer.
//
// Pinned as REPORTING rather than as clean, because that is what the rule does today. It is written
// down so the divergence is visible in the suite rather than only in a dry run, and so that fixing
// the shelf turns this test red and brings whoever fixes it straight here.
//
// Measured cost on the ahra tree: 141 findings against the installed rule's 130, with all 130
// agreeing exactly and nothing missing in the other direction.
func TestConsistentReturnHandlesATryWithADeclarationAndAFinally(t *testing.T) {
	t.Parallel()

	// Was an over-report and is now correct. The function returns on every path: the `try` returns
	// a value and the `catch` rethrows, so nothing falls off the end.
	//
	// The cause was in `control_flow_graph` rather than here. An implicit exception raised inside a
	// `catch` body was routed into the `try` statement's ordinary continuation, which made the
	// `finally`'s normal entry reachable and therefore the function's end reachable. ESLint's
	// `makeFirstThrowablePathInTryOrCatchBlock` adds that path to `thrownForkContext` in both the
	// try and catch positions; ours had a separate branch for the catch case. Removing it brought
	// the tree from 141 findings to eslint's 131, exactly, with no other rule moving.
	reporting := "async function f() { try { const r = g(); return r; } catch (e) { throw e; } finally { k(); } }"
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, ConsistentReturn, consistentReturnFile, reporting))

	// The two controls that localised the cause, kept because they are what proves the repair did
	// not simply silence the rule on every try/finally shape. Each differs from the case above by
	// exactly one element and each was already correct.
	withoutDeclaration := "async function f() { try { return 1; } catch (e) { throw e; } finally { k(); } }"
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, ConsistentReturn, consistentReturnFile, withoutDeclaration))

	withoutFinally := "async function f() { try { const r = g(); return r; } catch (e) { throw e; } }"
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, ConsistentReturn, consistentReturnFile, withoutFinally))
}
