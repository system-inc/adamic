package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// paramReassignFile is where the fixtures pretend to live.
const paramReassignFile = "/repository/source/ParamReassign.ts"

// decodedParamOptions routes a fixture's options through the rule's own decoder.
//
// Building the struct directly would leave the decoder and its two regex lists untested, and the
// regex list is the one field whose wire shape has no upstream counterpart: upstream compiles a
// fresh pattern at every call and this compiles once and caches.
func decodedParamOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoParamReassignOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
	return decoded
}

// The corpus is ESLint's own, extracted from the tester rather than retyped.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-param-reassign.js`: 40 valid and 38
// invalid. Extracted by evaluating the tester with a stubbed RuleTester, so no source string here
// was typed by hand.
//
// Each invalid row carries the message id, the RENDERED message text, and the byte offsets of the
// span, all measured by running the installed rule. The rendered text matters because this rule
// interpolates the parameter name into both messages, and an id assertion cannot see a message
// naming the wrong parameter -- which is exactly what a rule reporting at the anchor rather than at
// the occurrence would produce on a two-parameter function.
//
// The offsets are stated against the source AS THE HARNESS WRITES IT. `RunTyped` writes each fixture
// as the trimmed source plus a newline, and every case here is a single line with no leading
// whitespace, so the offsets are unshifted. That is checked rather than assumed: the assertion
// slices the same string it passed in and compares the reported text.
func TestNoParamReassignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		options     any
		messageId   string
		messageText string
		spanStart   int
		spanEnd     int
	}{
		{"function foo(bar) { bar = 13; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 20, 23},
		{"function foo(bar) { bar += 13; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 20, 23},
		{"function foo(bar) { (function() { bar = 13; })(); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 34, 37},
		{"function foo(bar) { ++bar; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 22, 25},
		{"function foo(bar) { bar++; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 20, 23},
		{"function foo(bar) { --bar; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 22, 25},
		{"function foo(bar) { bar--; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 20, 23},
		{"function foo({bar}) { bar = 13; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 22, 25},
		{"function foo([, {bar}]) { bar = 13; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 26, 29},
		{"function foo(bar) { ({bar} = {}); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 22, 25},
		{"function foo(bar) { ({x: [, bar = 0]} = {}); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 28, 31},
		{"function foo(bar) { for (bar in baz); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 25, 28},
		{"function foo(bar) { for (bar of baz); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'bar'.", 25, 28},
		{"function foo(bar) { bar.a = 0; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 20, 23},
		{"function foo(bar) { bar.get(0).a = 0; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 20, 23},
		{"function foo(bar) { delete bar.a; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 27, 30},
		{"function foo(bar) { ++bar.a; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 22, 25},
		{"function foo(bar) { for (bar.a in {}); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 25, 28},
		{"function foo(bar) { for (bar.a of []); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 25, 28},
		{"function foo(bar) { (bar ? bar : [])[0] = 1; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 27, 30},
		{"function foo(bar) { [bar.a] = []; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 21, 24},
		{"function foo(bar) { [bar.a] = []; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 21, 24},
		{"function foo(bar) { [bar.a] = []; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^a.*$"]}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 21, 24},
		{"function foo(bar) { [bar.a] = []; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^B.*$"]}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 21, 24},
		{"function foo(bar) { ({foo: bar.a} = {}); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'bar'.", 27, 30},
		{"function foo(a) { ({a} = obj); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParam", "Assignment to function parameter 'a'.", 20, 21},
		{"function foo(a) { ([...a] = obj); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'a'.", 23, 24},
		{"function foo(a) { ({...a} = obj); }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'a'.", 23, 24},
		{"function foo(a) { ([...a.b] = obj); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 23, 24},
		{"function foo(a) { ({...a.b} = obj); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 23, 24},
		{"function foo(a) { for ({bar: a.b} in {}); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 29, 30},
		{"function foo(a) { for ([a.b] of []); }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 24, 25},
		{"function foo(a) { a &&= b; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'a'.", 18, 19},
		{"function foo(a) { a ||= b; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'a'.", 18, 19},
		{"function foo(a) { a ??= b; }", nil, "assignmentToFunctionParam", "Assignment to function parameter 'a'.", 18, 19},
		{"function foo(a) { a.b &&= c; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 18, 19},
		{"function foo(a) { a.b.c ||= d; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 18, 19},
		{"function foo(a) { a[b] ??= c; }", decodedParamOptions(t, `{"props": true}`), "assignmentToFunctionParamProp", "Assignment to property of function parameter 'a'.", 18, 19},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText+" "+testCase.messageId, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
				testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.messageId)

			finding := result.Diagnostics[0]

			// Equality on the rendered text, not `strings.Contains`. A predicate weaker than the
			// property it guards is not a guard, and the two messages here differ only by a phrase
			// in the middle, so a containment check on the parameter name passes for either one.
			if finding.Message.Description != testCase.messageText {
				t.Errorf("message %q, want %q", finding.Message.Description, testCase.messageText)
			}

			if finding.Range.Pos() != testCase.spanStart || finding.Range.End() != testCase.spanEnd {
				t.Errorf("reported [%d,%d), want [%d,%d)",
					finding.Range.Pos(), finding.Range.End(), testCase.spanStart, testCase.spanEnd)
			}

			// The span must be the occurrence's own identifier, which is the only assertion that
			// can see a finding of the right width in the wrong place.
			written := strings.TrimSpace(testCase.sourceText) + "\n"
			reported := written[finding.Range.Pos():finding.Range.End()]
			if !strings.HasSuffix(testCase.messageText, "'"+reported+"'.") {
				t.Errorf("span covers %q, which is not the name the message reports", reported)
			}
		})
	}
}

// The clean cases are the whole discrimination, and four of them are textually indistinguishable
// from failing ones: what separates them is which binding the name resolves to.
func TestNoParamReassignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
	}{
		{"function foo(a) { var b = a; }", nil},
		{"function foo(a) { for (b in a); }", nil},
		{"function foo(a) { for (b of a); }", nil},
		{"function foo(a) { a.prop = 'value'; }", nil},
		{"function foo(a) { for (a.prop in obj); }", nil},
		{"function foo(a) { for (a.prop of arr); }", nil},
		{"function foo(a) { (function() { var a = 12; a++; })(); }", nil},
		{"function foo() { someGlobal = 13; }", nil},
		{"function foo() { someGlobal = 13; }", nil},
		{"function foo(a) { a.b = 0; }", nil},
		{"function foo(a) { delete a.b; }", nil},
		{"function foo(a) { ++a.b; }", nil},
		{"function foo(a) { [a.b] = []; }", nil},
		{"function foo(a) { bar(a.b).c = 0; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { data[a.b] = 0; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { +a.b; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { (a ? [] : [])[0] = 1; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { (a.b ? [] : [])[0] = 1; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { a.b = 0; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`)},
		{"function foo(a) { ++a.b; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`)},
		{"function foo(a) { delete a.b; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`)},
		{"function foo(a) { for (a.b in obj); }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`)},
		{"function foo(a) { for (a.b of arr); }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`)},
		{"function foo(a, z) { a.b = 0; x.y = 0; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a", "x"]}`)},
		{"function foo(a) { a.b.c = 0;}", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsFor": ["a"]}`)},
		{"function foo(aFoo) { aFoo.b = 0; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^a.*$"]}`)},
		{"function foo(aFoo) { ++aFoo.b; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^a.*$"]}`)},
		{"function foo(aFoo) { delete aFoo.b; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^a.*$"]}`)},
		{"function foo(a, z) { aFoo.b = 0; x.y = 0; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^a.*$", "^x.*$"]}`)},
		{"function foo(aFoo) { aFoo.b.c = 0;}", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^a.*$"]}`)},
		{"function foo(a) { ({ [a]: variable } = value) }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { ([...a.b] = obj); }", decodedParamOptions(t, `{"props": false}`)},
		{"function foo(a) { ({...a.b} = obj); }", decodedParamOptions(t, `{"props": false}`)},
		{"function foo(a) { for (obj[a.b] in obj); }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { for (obj[a.b] of arr); }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { for (bar in a.b); }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { for (bar of a.b); }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { for (bar in baz) a.b; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { for (bar of baz) a.b; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(bar, baz) { bar.a = true; baz.b = false; }", decodedParamOptions(t, `{"props": true, "ignorePropertyModificationsForRegex": ["^(foo|bar)$"], "ignorePropertyModificationsFor": ["baz"]}`)},

		// A catch binding spelled like the parameter, which is not it.
		{"function foo(a) { try {} catch (a) { a = 1; } }", nil},

		// A property NAMED like the parameter, on some other object. `WritesToBinding` declines the
		// name side of a property access, and this is the shape that reaches that guard.
		{"function foo(a) { other.a = 1; }", nil},

		// A label and an object literal key spelled like the parameter.
		{"function foo(a) { a: for(;;) break a; }", nil},
		{"function foo(a) { var o = { a: 1 }; }", nil},

		// A parameter property on the RIGHT of an assignment, which is a read of it rather than a
		// write through it. The corpus writes every clean shape where the parameter is read to
		// decide WHERE a write lands and never the plainer one where it is simply the value being
		// read, so a climb that forgot which side of the assignment it came up passes all 78
		// imported cases and reports all three of these. Found by a mutant; measured silent.
		{"function foo(a) { x = a.b; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { x.y = a.b; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { x += a.b; }", decodedParamOptions(t, `{"props": true}`)},
		{"function foo(a) { var x = a.b; }", decodedParamOptions(t, `{"props": true}`)},

		// `props` off is the default, so a property write is clean unless the option turns it on.
		// The corpus writes this shape under an explicit `false` and never under an absent option,
		// which is the shape the live config actually produces.
		{"function foo(a) { a.b = 0; }", nil},
		{"function foo(a) { ++a.b; }", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
				paramReassignFile, testCase.sourceText, testCase.options))
		})
	}
}

// A CONCISE arrow body, which is the one place the climb reaches a function with no statement first.
//
// Every other shape puts a statement between the expression and its function, and the statement stops
// the climb, so the function-scope half of the stop set is never asked. A concise arrow has no
// statement at all: `(a) => a.b = 1` climbs from the identifier through the property access and the
// assignment straight to the arrow. Found by a mutant that removed the whole function-scope arm and
// survived all 78 corpus cases, then reachability confirmed by walking the parse rather than argued.
//
// The silent half is the load-bearing one. `(a) => a.b` reaches the same arrow with nothing to
// report, so a stop set that never stopped would answer the same way here; what separates them is
// that without the arm the climb runs off the top of the tree, and the arm is what makes the walk
// terminate at the right boundary rather than by exhausting parents.
func TestNoParamReassignStopsAtAConciseArrowBody(t *testing.T) {
	t.Parallel()

	reporting := []struct {
		sourceText string
		messageId  string
		options    any
	}{
		{"const f = (a) => a.b = 1;", "assignmentToFunctionParamProp",
			decodedParamOptions(t, `{"props": true}`)},
		{"const f = (a) => a.b++;", "assignmentToFunctionParamProp",
			decodedParamOptions(t, `{"props": true}`)},
		{"const f = (a) => delete a.b;", "assignmentToFunctionParamProp",
			decodedParamOptions(t, `{"props": true}`)},
		{"const f = (a) => (a.b, a.c = 1);", "assignmentToFunctionParamProp",
			decodedParamOptions(t, `{"props": true}`)},
		{"const f = (a) => a = 1;", "assignmentToFunctionParam", nil},
	}
	for _, testCase := range reporting {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
				paramReassignFile, testCase.sourceText, testCase.options), testCase.messageId)
		})
	}

	t.Run("a concise arrow that merely reads", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
			"const f = (a) => a.b;", decodedParamOptions(t, `{"props": true}`)))
	})

	// The case where the function-scope arm changes the VERDICT rather than only where the walk
	// stops. Everything above reaches a statement or a declaration eventually, so removing the arm
	// moves where the climb ends without moving what it answers -- both routes reach the same
	// silence, which is why the first fixtures written for this mutant did not kill it.
	//
	// Here the parameter is read INSIDE a nested arrow whose whole value is then written through.
	// Without the arm the climb escapes the arrow, reaches the assignment outside it, and reports a
	// write through a parameter that is only being closed over. Measured silent upstream.
	silentAcrossAFunctionBoundary := []string{
		"function foo(a) { ((b) => a).c = 1; }",
		"function foo(a) { (function() { return a; }).c = 1; }",
		"function foo(a) { [(b) => a][0] = 1; }",
		"function foo(a) { x = ((a) => a).b; }",
	}
	for _, sourceText := range silentAcrossAFunctionBoundary {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
				paramReassignFile, sourceText, decodedParamOptions(t, `{"props": true}`)))
		})
	}
}

// A POSTFIX update through a parameter, which the corpus writes only in its prefix form.
//
// Upstream's tree gives both fixities one node type and its rule answers true for that type without
// looking at the operator; ours gives them two kinds, so the postfix one is a separate arm that no
// imported case reaches. Found by a mutant that neutralised it and survived all 78 corpus cases.
func TestNoParamReassignSeesPostfixUpdates(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function foo(a) { a.b++; }",
		"function foo(a) { a.b--; }",

		// The same update used as a value, which is the shape where the climb has to answer before
		// it reaches the enclosing assignment.
		"function foo(a) { x = a.b++; }",
	}
	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
				paramReassignFile, sourceText, decodedParamOptions(t, `{"props": true}`)),
				"assignmentToFunctionParamProp")
		})
	}
}

// A write inside a parameter's DEFAULT VALUE reports, and the parameter's own binding does not.
//
// The corpus writes neither shape, and the obvious reading of upstream -- that it walks the body --
// loses all of the reporting half. It excludes only the reference that IS the declaration, through
// the initialization flag on it, so everything else in the parameter list is an ordinary reference.
// Found by a mutant that widened the search from the body to the whole function and survived every
// fixture; measured against the installed rule on the five shapes below.
func TestNoParamReassignSeesIntoParameterDefaults(t *testing.T) {
	t.Parallel()

	reporting := []struct {
		sourceText string
		messageId  string
		options    any
	}{
		{"function foo(a, b = (a = 1)) { }", "assignmentToFunctionParam", nil},
		{"function foo(a = (a = 1)) { }", "assignmentToFunctionParam", nil},
		{"function foo(a, b = ++a) { }", "assignmentToFunctionParam", nil},
		{"function foo(a, b = (a.c = 1)) { }", "assignmentToFunctionParamProp",
			decodedParamOptions(t, `{"props": true}`)},
	}
	for _, testCase := range reporting {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
				paramReassignFile, testCase.sourceText, testCase.options), testCase.messageId)
		})
	}

	// The silent half, and it is what keeps the widening from being a false-positive class. A
	// parameter's own binding is the declaration, and a default that merely READS another parameter
	// is a read.
	silent := []struct {
		sourceText string
		options    any
	}{
		{"function foo(a = 1) { }", nil},
		{"function foo(a, b = a) { }", nil},
		{"function foo({a = 1}) { }", nil},
		{"function foo({a, b = a.c}) { }", decodedParamOptions(t, `{"props": true}`)},
	}
	for _, testCase := range silent {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
				paramReassignFile, testCase.sourceText, testCase.options))
		})
	}
}

// Class bodies, which the corpus never writes and which cost every finding inside one.
//
// Upstream registers three listeners and covers all of these without naming them, because an estree
// tree makes a method's value a plain function expression. Our parser gives each its own node kind
// with no function expression underneath, so registering only upstream's three goes silent on every
// parameter in every class. Found by a differential run, where all six of its mismatches were this.
func TestNoParamReassignReachesIntoClassBodies(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		messageId  string
		options    any
	}{
		{"class C { m(a) { a = 1; } }", "assignmentToFunctionParam", nil},
		{"class C { constructor(a) { a = 1; } }", "assignmentToFunctionParam", nil},
		{"class C { set x(a) { a = 1; } }", "assignmentToFunctionParam", nil},
		{"class C { static m(a) { a = 1; } }", "assignmentToFunctionParam", nil},
		{"class C { m(a) { a.b = 1; } }", "assignmentToFunctionParamProp",
			decodedParamOptions(t, `{"props": true}`)},

		// A getter takes no parameters, so it is here for the listener rather than for a finding;
		// the property write is on the setter beside it.
		{"class C { get x() { return 1; } set x(a) { a = 1; } }", "assignmentToFunctionParam", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
				testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.messageId)
			if got := result.Diagnostics[0].Message.Description; got != "Assignment to function parameter 'a'." &&
				got != "Assignment to property of function parameter 'a'." {
				t.Errorf("message %q names the wrong parameter", got)
			}
		})
	}
}

// A nested function whose own parameter shadows an outer one reports EXACTLY ONCE.
//
// This is the shape the corpus never writes and the one the declaration-anchor family gets wrong.
// The corpus shadows a parameter with a `var` and never with another parameter, so a port comparing
// declaration KIND rather than node identity passes all 78 imported cases: two parameters are the
// same kind, so a kind test calls the inner one a match for the outer, and the write is then
// attributed to both functions and reported twice.
//
// It is a count assertion rather than a silence one, because upstream DOES report each of these --
// the inner function's own parameter is being reassigned. Written first as a clean case on the
// reasoning that a shadow is exempt, which is true of the binding it shadows and not of the shadow
// itself. Measured against the installed rule, which reports one finding on each.
func TestNoParamReassignAttributesAShadowedWriteToOneFunction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantName   string
	}{
		// Two parameters spelled the same, one nested inside the other's function. One finding.
		{"function foo(a) { function inner(a) { a = 1; } }", "a"},
		{"function foo(a) { const inner = (a) => { a = 1; }; }", "a"},

		// The same nesting with the write reaching the OUTER parameter instead, which is the case
		// that separates "attributed to the right function" from "attributed to one function".
		{"function foo(a) { function inner(b) { a = 1; } }", "a"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
				testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "assignmentToFunctionParam")
			want := "Assignment to function parameter '" + testCase.wantName + "'."
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Errorf("message %q, want %q", got, want)
			}
		})
	}
}

// The typed harness is required, and this pins that a later revert to the plain one fails loudly.
//
// The failure without a checker is SILENCE rather than a crash, because GetSymbolAtLocation tolerates
// a nil receiver and answers nil. That is the more dangerous of the two failure modes: every
// StaysSilent case passes vacuously and only the Fires cases notice.
func TestNoParamReassignNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoParamReassign.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring the checker, which makes every silent fixture vacuous")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoParamReassign, paramReassignFile,
		"function foo(bar) { bar = 13; }"))
}

// The decoder and the two allowance lists, which are the lines with no upstream counterpart.
func TestNoParamReassignOptions(t *testing.T) {
	t.Parallel()

	propertyWrite := "function foo(a) { a.b = 0; }"

	t.Run("nil options, the shape the live config produces", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
			paramReassignFile, propertyWrite, nil))
	})

	t.Run("a zero-value struct declines property writes", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoParamReassign,
			paramReassignFile, propertyWrite, NoParamReassignOptions{}))
	})

	t.Run("empty input decodes without erroring", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoParamReassignOptions(nil)
		if err != nil {
			t.Fatalf("decoding empty input: %v", err)
		}
		options, ok := decoded.(NoParamReassignOptions)
		if !ok {
			t.Fatalf("decoded to %T", decoded)
		}
		if options.Props {
			t.Error("empty input decoded to props on, which widens the rule by default")
		}
	})

	t.Run("malformed input is an error", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoParamReassignOptions([]byte("[")); err == nil {
			t.Error("malformed options decoded without an error")
		}
	})

	t.Run("an uncompilable pattern is dropped rather than fatal", func(t *testing.T) {
		t.Parallel()
		// A divergence, stated at the rule and pinned here. Upstream builds a RegExp at every call
		// and a bad pattern throws, taking the run down; this drops it, so the name stops being
		// excused and the rule reports MORE. That is the visible direction rather than the one that
		// hides findings.
		options := decodedParamOptions(t,
			`{"props": true, "ignorePropertyModificationsForRegex": ["^(unclosed"]}`)
		result := rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
			propertyWrite, options)
		rule_testing.ExpectFindings(t, result, "assignmentToFunctionParamProp")
	})

	t.Run("a pattern Go accepts and JavaScript does not is still a pattern", func(t *testing.T) {
		t.Parallel()
		// The two dialects are not the same. Upstream compiles with the Unicode flag and this
		// compiles with RE2, which has no backreferences and no lookaround, so a pattern using
		// either is dropped here and honoured there. No corpus case uses one, and the divergence is
		// recorded rather than papered over.
		options := decodedParamOptions(t,
			`{"props": true, "ignorePropertyModificationsForRegex": ["^(?!x)a$"]}`)
		result := rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
			propertyWrite, options)
		rule_testing.ExpectFindings(t, result, "assignmentToFunctionParamProp")
	})

	t.Run("both lists apply together", func(t *testing.T) {
		t.Parallel()
		// Upstream's last valid case pairs a literal name with a pattern; this is the same pairing
		// asserted from the other side, with each list excusing a different parameter and a third
		// parameter matching neither and reporting.
		options := decodedParamOptions(t,
			`{"props": true, "ignorePropertyModificationsFor": ["one"], "ignorePropertyModificationsForRegex": ["^tw"]}`)
		result := rule_testing.RunTypedWithOptions(t, NoParamReassign, paramReassignFile,
			"function foo(one, two, three) { one.a = 0; two.a = 0; three.a = 0; }", options)
		rule_testing.ExpectFindings(t, result, "assignmentToFunctionParamProp")
		if got := result.Diagnostics[0].Message.Description; got != "Assignment to property of function parameter 'three'." {
			t.Errorf("reported %q, want the third parameter", got)
		}
	})
}

// Both messages, asserted against literals typed here rather than against the rule's own builders,
// which would move with them under mutation.
func TestNoParamReassignMessages(t *testing.T) {
	t.Parallel()

	binding := messageAssignmentToParameter("bar")
	if binding.Id != "assignmentToFunctionParam" {
		t.Errorf("binding message id is %q", binding.Id)
	}
	if binding.Description != "Assignment to function parameter 'bar'." {
		t.Errorf("binding message renders as %q", binding.Description)
	}

	property := messageAssignmentToParameterProperty("bar")
	if property.Id != "assignmentToFunctionParamProp" {
		t.Errorf("property message id is %q", property.Id)
	}
	if property.Description != "Assignment to property of function parameter 'bar'." {
		t.Errorf("property message renders as %q", property.Description)
	}
}
