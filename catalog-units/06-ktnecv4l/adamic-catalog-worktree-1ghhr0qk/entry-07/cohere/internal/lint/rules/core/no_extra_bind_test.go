package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// extraBindFile is where the fixtures pretend to live.
const extraBindFile = "/repository/source/ExtraBind.ts"

// The corpus is ESLint's own, extracted from the tester rather than retyped.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-extra-bind.js`: 11 valid and 32
// invalid. Extracted by evaluating the tester with a stubbed RuleTester and serialising what it was
// handed, so no source string here was typed by hand and none of the escape sequences could be
// cooked on the way in.
//
// Each invalid row carries three things upstream states and a message-id fixture cannot see: the
// byte offsets of the reported span, and the exact text the fixer must write. Twelve rows carry a
// nil rewrite, which is upstream's `output: null` -- a case it reports and deliberately declines to
// repair, because the repair would delete a comment or a side effect. Reproducing the decline is as
// much the port as reproducing the repair.
//
// The offsets were measured by running the installed rule over each case and converting its
// one-based line and column to a byte offset, which two of these cases need because their span sits
// on the second line.
func TestNoExtraBindFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		spanStart   int
		spanEnd     int
		fixedSource any
	}{
		{"var a = function() { return 1; }.bind(b)", 33, 37, "var a = function() { return 1; }"},
		{"var a = function() { return 1; }['bind'](b)", 33, 39, "var a = function() { return 1; }"},
		{"var a = function() { return 1; }[`bind`](b)", 33, 39, "var a = function() { return 1; }"},
		{"var a = (() => { return 1; }).bind(b)", 30, 34, "var a = (() => { return 1; })"},
		{"var a = (() => { return this; }).bind(b)", 33, 37, "var a = (() => { return this; })"},
		{"var a = function() { (function(){ this.c }) }.bind(b)", 46, 50, "var a = function() { (function(){ this.c }) }"},
		{"var a = function() { function c(){ this.d } }.bind(b)", 46, 50, "var a = function() { function c(){ this.d } }"},
		{"var a = function() { return 1; }.bind(this)", 33, 37, "var a = function() { return 1; }"},
		{"var a = function() { (function(){ (function(){ this.d }.bind(c)) }) }.bind(b)", 70, 74, "var a = function() { (function(){ (function(){ this.d }.bind(c)) }) }"},
		{"var a = (function() { return 1; }).bind(this)", 35, 39, "var a = (function() { return 1; })"},
		{"var a = (function() { return 1; }.bind)(this)", 34, 38, "var a = (function() { return 1; })"},
		{"var a = function() {}.bind(b++)", 22, 26, nil},
		{"var a = function() {}.bind(b())", 22, 26, nil},
		{"var a = function() {}.bind(b.c)", 22, 26, nil},
		{"var a = function() {}/**/.bind(b)", 26, 30, "var a = function() {}/**/"},
		{"var a = function() {}/**/['bind'](b)", 26, 32, "var a = function() {}/**/"},
		{"var a = function() {}//comment\n.bind(b)", 32, 36, "var a = function() {}//comment\n"},
		{"var a = function() {}./**/bind(b)", 26, 30, nil},
		{"var a = function() {}[/**/'bind'](b)", 26, 32, nil},
		{"var a = function() {}.//\nbind(b)", 25, 29, nil},
		{"var a = function() {}.bind/**/(b)", 22, 26, nil},
		{"var a = function() {}.bind(\n/**/b)", 22, 26, nil},
		{"var a = function() {}.bind(b/**/)", 22, 26, nil},
		{"var a = function() {}.bind(b//\n)", 22, 26, nil},
		{"var a = function() {}.bind(b\n/**/)", 22, 26, nil},
		{"var a = function() {}.bind(b)/**/", 22, 26, "var a = function() {}/**/"},
		{"var a = function() { return 1; }.bind?.(b)", 33, 37, "var a = function() { return 1; }"},
		{"var a = function() { return 1; }?.bind(b)", 34, 38, "var a = function() { return 1; }"},
		{"var a = (function() { return 1; }?.bind)(b)", 35, 39, "var a = (function() { return 1; })"},
		{"var a = function() { return 1; }['bind']?.(b)", 33, 39, "var a = function() { return 1; }"},
		{"var a = function() { return 1; }?.['bind'](b)", 35, 41, "var a = function() { return 1; }"},
		{"var a = (function() { return 1; }?.['bind'])(b)", 36, 42, "var a = (function() { return 1; })"},

		// Beyond the corpus. A method shorthand and a class method rebind `this`, so a `this` inside
		// one says nothing about the function being judged -- the same distinction upstream's corpus
		// makes with a nested function declaration, on shapes it never writes.
		{"var a = function() { var o = { m() { return this; } }; }.bind(b)", 57, 61, "var a = function() { var o = { m() { return this; } }; }"},
		{"var a = function() { class C { m() { return this; } } }.bind(b)", 56, 60, "var a = function() { class C { m() { return this; } } }"},

		// A getter and a setter, which are the two accessor shapes and neither is a corpus case.
		{"var a = function() { var o = { get x() { return this; } }; }.bind(b)", 61, 65, "var a = function() { var o = { get x() { return this; } }; }"},

		// A `this` inside a nested arrow does NOT save the outer function, so this reports while the
		// corpus's `function() { return () => this; }` is clean. The difference is which function the
		// arrow is written inside: here it is a nested function, so its `this` is that function's.
		{"var a = function() { function c() { return () => this; } }.bind(b)", 59, 63, "var a = function() { function c() { return () => this; } }"},

		// An argument that is a literal rather than an identifier, which upstream's allowed set
		// covers and its corpus never writes.
		{"var a = function() {}.bind(1)", 22, 26, "var a = function() {}"},
		{"var a = function() {}.bind(null)", 22, 26, "var a = function() {}"},
		{"var a = function() {}.bind(function(){})", 22, 26, "var a = function() {}"},

		{"var a = function() {}.bind(1n)", 22, 26, "var a = function() {}"},
		{"var a = function() {}.bind(/re/)", 22, 26, "var a = function() {}"},
		{"var a = function() {}.bind(true)", 22, 26, "var a = function() {}"},
		{"var a = function() {}.bind('s')", 22, 26, "var a = function() {}"},

		// Three shapes that report and are NOT repaired, because upstream's allowed set is estree's
		// `Literal` and these are outside it. Each is inside the intuitive reading of the set and
		// outside the real one, which is why all three are measured against the installed rule
		// rather than reasoned about, and why the corpus -- which writes no literal argument at all
		// -- cannot separate the two readings.
		//
		// A no-substitution template is its own node kind rather than a literal. This one was in the
		// allowed set first, on the intuitive reading, and the differential caught it.
		{"var a = function() {}.bind(`t`)", 22, 26, nil},

		// A negative number is a unary expression over a literal, not a literal.
		{"var a = function() {}.bind(-1)", 22, 26, nil},

		// An arrow argument is outside the set even though evaluating one is inert. Upstream's own
		// comment says the set is stricter than it needs to be, and this is where that shows.
		{"var a = function() {}.bind(() => {})", 22, 26, nil},

		// An object and an array literal, which are not `Literal` in an estree tree either.
		{"var a = function() {}.bind({})", 22, 26, nil},
		{"var a = function() {}.bind([])", 22, 26, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExtraBind, extraBindFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpected")

			reported := result.Diagnostics[0].Range
			if reported.Pos() != testCase.spanStart || reported.End() != testCase.spanEnd {
				t.Errorf("reported [%d,%d), want [%d,%d) -- text %q",
					reported.Pos(), reported.End(), testCase.spanStart, testCase.spanEnd,
					testCase.sourceText[reported.Pos():reported.End()])
			}

			switch want := testCase.fixedSource.(type) {
			case nil:
				// `output: null` upstream. The finding stands and the repair is declined, so
				// asserting no fixes is asserting the decline rather than asserting nothing.
				if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
					t.Errorf("proposed %d fixes on a case upstream declines to repair", len(fixes))
				}
			case string:
				rule_testing.ExpectFixedSource(t, result, want)
			}
		})
	}
}

// The clean cases are what separate a pointless bind from a useful one, and every one of them is a
// false positive this rule would otherwise ship.
func TestNoExtraBindStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ sourceText string }{
		{"var a = function(b) { return b }.bind(c, d)"},
		{"var a = function(b) { return b }.bind(...c)"},
		{"var a = function() { this.b }()"},
		{"var a = function() { this.b }.foo()"},
		{"var a = f.bind(a)"},
		{"var a = function() { return this.b }.bind(c)"},
		{"var a = (() => { return b }).bind(c, d)"},
		{"(function() { (function() { this.b }.bind(this)) }.bind(c))"},
		{"var a = function() { return 1; }[bind](b)"},
		{"var a = function() { return 1; }[`bi${n}d`](b)"},
		{"var a = function() { return () => this; }.bind(b)"},

		// Beyond the corpus. `bind` reached through a private name, which `Text()` reports as `bind`
		// with the hash stripped, so a rule comparing the text alone would report it.
		{"class C { #bind() {}; m() { return function(){}.#bind(b); } }"},

		// A template with a substitution names something unknown at parse time. The corpus writes
		// this shape with a variable in it; this is the byte-identical prefix case, where the
		// substitution is the empty string and the text still is not statically `bind`.
		{"var a = function() { return 1; }[`bind${x}`](b)"},

		// A subscript that IS a static string and is not `bind`. The corpus separates a static
		// subscript from a dynamic one and never separates a static subscript from a static one
		// naming something else, so a rule reading only the node kind passes every imported case.
		// Found by a mutant that let the string arm accept any text; measured silent upstream.
		{"var a = function() { return 1; }[\"call\"](b)"},
		{"var a = function() { return 1; }[`call`](b)"},
		{"var a = function() { return 1; }[\"\"](b)"},

		// Zero arguments and three arguments. The corpus covers two arguments and a spread; these are
		// the other two counts, and a port reading `len(arguments) > 0` would report the second.
		{"var a = function() { return 1; }.bind()"},
		{"var a = function() { return 1; }.bind(b, c, d)"},

		// The receiver is not the function. `bind` called ON something else, with a function as its
		// argument, is a different expression entirely.
		{"var a = b.bind(function() { return 1; })"},

		// The `*.bind` access is an ARGUMENT of a call rather than its callee, so nothing is being
		// bound. One argument, so the argument-count check cannot cover for the callee check that
		// actually declines it; found by a mutant that dropped the callee identity comparison and
		// survived every other fixture. Measured silent upstream.
		{"foo(function(){}.bind)"},
		{"foo((function(){}.bind))"},

		// A `new` rather than a call. Upstream keys on a call expression specifically, and our tree
		// gives `new` its own node kind, so this declines at the same place.
		{"new (function(){}.bind)(b)"},

		// A nested arrow reading `this` inside the bound function itself, which is the corpus's
		// passing case written as a statement rather than a return.
		{"var a = function() { (() => this.b)(); }.bind(c)"},

		// `this` in the parameter defaults of the bound function, which is still its own `this`,
		// because a default is evaluated with the function's own binding. Upstream gets this for
		// free -- its `this` listener fires anywhere between the function's entry and its exit, and
		// the parameter list is inside that window -- and a port scanning only the body reports it.
		// Written after measuring the installed rule silent; the corpus writes no `this` in a
		// parameter position, so nothing imported could have caught it.
		{"var a = function(x = this) { return x; }.bind(b)"},
		{"var a = function(x = () => this) { return x; }.bind(b)"},

		// A method named `bind` on an object, which is a property access whose receiver is not a
		// function expression at all.
		{"var a = ({ bind() {} }).bind(b)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoExtraBind, extraBindFile, testCase.sourceText))
		})
	}
}

// The repair is two disjoint spans, and this pins the gap between them rather than only the result.
//
// A single span from the receiver to the end of the call would produce the same output on every
// corpus case, because none of them writes anything between the member access and the argument list
// that has to survive. Upstream's comment names the shape that does, and it is the reason the fixer
// is built as a pair, so it is written here.
func TestNoExtraBindRepairKeepsWhatSitsBetweenTheTwoSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		fixedSource string
	}{
		// Upstream's own example, from the comment above its `tokenPairs`. The parenthesis between
		// `.bind` and the argument list is outside both removal spans and survives.
		{"var a = (function(){}.bind ) (obj)", "var a = (function(){} ) "},
		{"var a = (function(){}?.['bind'] ) ?.(obj)", "var a = (function(){} ) "},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExtraBind, extraBindFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpected")
			rule_testing.ExpectFixedSource(t, result, testCase.fixedSource)
		})
	}
}

// The message is a value with no interpolation, so the assertion is on the constant's own fields,
// compared against literals typed here rather than against the rule's own constant.
func TestNoExtraBindMessage(t *testing.T) {
	t.Parallel()

	if messageUnnecessaryBind.Id != "unexpected" {
		t.Errorf("message id is %q", messageUnnecessaryBind.Id)
	}
	if messageUnnecessaryBind.Description == "" {
		t.Error("message carries no description")
	}
}
