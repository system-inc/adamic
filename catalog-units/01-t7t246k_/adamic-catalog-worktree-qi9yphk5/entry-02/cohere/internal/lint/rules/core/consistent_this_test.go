package core

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// consistentThisFile is where the fixtures pretend to live.
const consistentThisFile = "/repository/source/ConsistentThis.ts"

// consistentThisOptions routes an alias list through the rule's own exported decoder.
//
// Building the settings struct directly would leave the decoder untested, and the decoder is
// where this rule's two most dangerous lines live: the default of ["that"] rather than the zero
// value, and reading upstream's variadic list, one alias per element. A struct built by hand passes
// every fixture while an inverted default ships.
func consistentThisOptions(t *testing.T, aliases ...string) any {
	t.Helper()
	raw, err := json.Marshal(aliases)
	if err != nil {
		t.Fatalf("could not marshal the alias list: %v", err)
	}
	decoded, err := DecodeConsistentThisOptions(raw)
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// The corpus is ESLint's own, imported verbatim from
// eslint/tests/lib/rules/consistent-this.js: 14 pass, 12 fail.
//
// Every case carries its own alias list, and the same source can be valid under one and invalid
// under another: "var foo = 42, self = this" is clean with the alias "self" and reports with the
// alias "that". Rendering the options beside each case is what makes that readable as a version
// of the same input rather than as a contradiction in the corpus.
//
// The cases were extracted by loading upstream's tester module with a stubbed RuleTester and
// rendering these literals from that JSON, so nothing was retyped and no escape sequence was
// hand-written. The generator refuses any case holding a byte outside printable ASCII.
func TestConsistentThisFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		aliases    []string
		sourceText string
		messages   []string
	}{
		{"a capture under an undesignated name", []string{"that"}, "var context = this", []string{"unexpectedAlias"}},
		{"the default alias when another is designated", []string{"self"}, "var that = this", []string{"unexpectedAlias"}},
		{"a second declarator capturing under the wrong name", []string{"that"}, "var foo = 42, self = this", []string{"unexpectedAlias"}},
		{"the alias holding a number", []string{"self"}, "var self = 42", []string{"aliasNotAssignedToThis"}},
		{"the alias declared and never assigned", []string{"self"}, "var self", []string{"aliasNotAssignedToThis"}},
		{"the alias declared then given a number, reporting twice", []string{"self"}, "var self; self = 42", []string{"aliasNotAssignedToThis", "aliasNotAssignedToThis"}},
		{"a bare assignment capture under the wrong name", []string{"that"}, "context = this", []string{"unexpectedAlias"}},
		{"the default alias when another is designated, assigned", []string{"self"}, "that = this", []string{"unexpectedAlias"}},
		{"the designated alias assigned outside its declaration", []string{"that"}, "self = this", []string{"unexpectedAlias"}},
		{"a compound assignment to the alias", []string{"self"}, "self += this", []string{"aliasNotAssignedToThis"}},
		{"the alias assigned only from an inner function", []string{"self"}, "var self; (function() { self = this; }())", []string{"aliasNotAssignedToThis"}},
		{"the alias assigned only from an inner function, again", []string{"self"}, "var self; (function() { self = this; }())", []string{"aliasNotAssignedToThis"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
				testCase.sourceText, consistentThisOptions(t, testCase.aliases...))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// The clean cases, and the destructuring ones are the whole reason four of them exist.
//
// "var {foo, bar} = this" captures properties of this rather than this itself, so no name in it
// is an alias for the context. A port matching on the initializer alone reports all four of the
// destructuring cases, since every one of them assigns a this expression.
func TestConsistentThisStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		aliases    []string
		sourceText string
	}{
		{"the default alias capturing this", []string{"that"}, "var foo = 42, that = this"},
		{"a designated alias capturing this", []string{"self"}, "var foo = 42, self = this"},
		{"an undesignated name holding a number", []string{"that"}, "var self = 42"},
		{"an undesignated name declared bare", []string{"that"}, "var self"},
		{"the alias assigned in the same scope", []string{"self"}, "var self; self = this"},
		{"the alias assigned after another declarator", []string{"self"}, "var foo, self; self = this"},
		{"the alias assigned after an unrelated write", []string{"self"}, "var foo, self; foo = 42; self = this"},
		{"an undesignated name assigned a number", []string{"that"}, "self = 42"},
		{"a member expression capturing this", []string{"self"}, "var foo = {}; foo.bar = this"},
		{"two designated aliases each capturing this", []string{"self", "vm"}, "var self = this; var vm = this;"},
		{"object destructuring of this", []string{"self"}, "var {foo, bar} = this"},
		{"object destructuring assignment of this", []string{"self"}, "({foo, bar} = this)"},
		{"array destructuring of this", []string{"self"}, "var [foo, bar] = this"},
		{"array destructuring assignment of this", []string{"self"}, "[foo, bar] = this"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ConsistentThis,
				consistentThisFile, testCase.sourceText, consistentThisOptions(t, testCase.aliases...)))
		})
	}
}

// The scope model, measured against the installed rule rather than derived from its source.
//
// Upstream's `reference.from === scope` decides whether a later assignment rescues a declaration,
// and what that expression MEANS was the hard part of this port. It is the innermost eslint scope,
// and eslint opens one for a block, a switch, a loop body, a catch and a `with`, but not for an
// unbraced `if` consequent, a labeled statement, or any expression nesting.
//
// Upstream's corpus writes none of these shapes, so every row here is a case it cannot express, and
// each was measured by driving the installed rule. The reporting rows are upstream reporting CORRECT
// code: in every one of them the write reaches the same var declaration and the alias does end up
// holding `this`. Reproduced rather than corrected, and written down here so the next reader does
// not helpfully fix it back to the intuitive answer.
//
// The last row caught a real defect in this port. A function's own body block is that function's
// scope rather than a nested one, and without that exemption the inner write never rescued the inner
// declaration, so this port reported twice where upstream reports once at column 5. Nothing in the
// imported corpus could see it: it writes no nested function declaring its own alias.
func TestConsistentThisScopeModel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"a bare block", "var self; { self = this; }", []string{"aliasNotAssignedToThis"}},
		{"a loop body", "var self; for(;;) { self = this; }", []string{"aliasNotAssignedToThis"}},
		{"a while body", "var self; while(a) { self = this; }", []string{"aliasNotAssignedToThis"}},
		{"a switch", "var self; switch(a){case 1: self = this;}", []string{"aliasNotAssignedToThis"}},
		{"a try block", "var self; try { self = this; } catch(e) {}", []string{"aliasNotAssignedToThis"}},
		{"a with statement", "var self; with(o) self = this;", []string{"aliasNotAssignedToThis"}},
		{"an unbraced if consequent", "var self; if(a) self = this;", nil},
		{"a labeled statement", "var self; label: self = this;", nil},
		{"a loop with no block body", "var self; for(var i=0;;) self = this;", nil},
		{"parenthesized, which is invisible", "var self; (self = this);", nil},
		{"a sequence expression", "var self; a, self = this;", nil},
		{"a conditional expression", "var self; a ? self = this : 0;", nil},
		{"a nested function declaring its own alias", "var self; function f() { var self; self = this; }", []string{"aliasNotAssignedToThis"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
				testCase.sourceText, consistentThisOptions(t, "self"))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// Shapes around the two judgments that upstream's corpus does not write.
//
// The `let` and `const` rows matter because upstream's corpus is entirely `var`, and a port keying
// on the declaration keyword rather than on the binding would go silent on both. The `||=` row is
// the compound-operator discrimination reaching a logical assignment, which reports twice: once
// because the alias never received a plain assignment and once because the operator is not `=`.
func TestConsistentThisOtherDeclarationForms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"a let declaration rescued in the same scope", "let self; self = this;", nil},
		{"a const capture of this", "const self = this;", nil},
		{"a redeclaration carrying the capture", "var self; var self = this;", nil},
		{"an assignment with no declaration at all", "self = this;", nil},
		{"a second declarator capturing under another name", "var self = this, other = this;", []string{"unexpectedAlias"}},
		{"a later write of something else", "var self; self = this; self = 42;", []string{"aliasNotAssignedToThis"}},
		{"a logical assignment of this", "var self; self ||= this;", []string{"aliasNotAssignedToThis", "aliasNotAssignedToThis"}},
		{"a method body declaring an unassigned alias", "class C { m() { var self; } }", []string{"aliasNotAssignedToThis"}},
		{"a function body declaring an unassigned alias", "function f() { var self; }", []string{"aliasNotAssignedToThis"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
				testCase.sourceText, consistentThisOptions(t, "self"))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// The decoder's own defaults and shapes, which no source fixture can reach.
//
// The default is the dangerous line: upstream defaults to ["that"], and a generic decoder handing
// back a zero-value struct would give an EMPTY alias list, which does not disable the rule. It makes
// aliasNotAssignedToThis unreachable and makes unexpectedAlias report every capture of `this` under
// any name at all. So the nil case is asserted by behaviour rather than only by field.
func TestDecodeConsistentThisOptions(t *testing.T) {
	t.Parallel()

	t.Run("an absent option is the alias that", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeConsistentThisOptions(nil)
		if err != nil {
			t.Fatalf("the decoder refused an absent option: %v", err)
		}
		settings, ok := decoded.(ConsistentThisSettings)
		if !ok {
			t.Fatalf("expected ConsistentThisSettings, got %T", decoded)
		}
		if len(settings.Aliases) != 1 || settings.Aliases[0] != "that" {
			t.Errorf("expected the default alias list [that], got %v", settings.Aliases)
		}
	})

	t.Run("a bare severity leaves the rule on its default", func(t *testing.T) {
		t.Parallel()
		// A rule configured as "error" is handed nil options, which arrives at Run as an untyped
		// nil rather than as settings. Asserting through the rule rather than the decoder is what
		// covers the fallback inside Run, which no decoder test can reach.
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, ConsistentThis,
			consistentThisFile, "var context = this", nil), "unexpectedAlias")
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ConsistentThis,
			consistentThisFile, "var that = this", nil))
	})

	t.Run("one element is one alias", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeConsistentThisOptions(json.RawMessage(`["self"]`))
		if err != nil {
			t.Fatalf("the decoder refused one alias: %v", err)
		}
		settings := decoded.(ConsistentThisSettings)
		if len(settings.Aliases) != 1 || settings.Aliases[0] != "self" {
			t.Errorf("expected [self], got %v", settings.Aliases)
		}
	})

	t.Run("every element is an alias, upstream's variadic spelling", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeConsistentThisOptions(json.RawMessage(`["self","vm"]`))
		if err != nil {
			t.Fatalf("the decoder refused two aliases: %v", err)
		}
		settings := decoded.(ConsistentThisSettings)
		if len(settings.Aliases) != 2 || settings.Aliases[0] != "self" || settings.Aliases[1] != "vm" {
			t.Errorf("expected [self vm], got %v", settings.Aliases)
		}
	})

	t.Run("the nested-array workaround is refused rather than read", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeConsistentThisOptions(json.RawMessage(`[["self","vm"]]`)); err == nil {
			t.Error("an element that is not a string decoded; upstream's schema refuses it")
		}
	})

	t.Run("an empty name is refused", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeConsistentThisOptions(json.RawMessage(`[""]`)); err == nil {
			t.Error("expected the decoder to refuse an empty alias")
		}
		if _, err := DecodeConsistentThisOptions(json.RawMessage(`["self",""]`)); err == nil {
			t.Error("expected the decoder to refuse an empty alias in a list")
		}
	})

	t.Run("an empty list falls back to the default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeConsistentThisOptions(json.RawMessage(`[]`))
		if err != nil {
			t.Fatalf("the decoder refused an empty list: %v", err)
		}
		settings := decoded.(ConsistentThisSettings)
		if len(settings.Aliases) != 1 || settings.Aliases[0] != "that" {
			t.Errorf("expected the default alias list [that], got %v", settings.Aliases)
		}
	})
}

// The finding points at the declaration or the assignment, and the two messages are distinct.
//
// Every ExpectFindings assertion above is satisfied by a rule reporting the whole statement, so the
// span is the only thing separating that from upstream's node. Upstream reports the VariableDeclarator
// and the AssignmentExpression, neither of which includes the `var` keyword or the semicolon.
func TestConsistentThisPointsAtTheAssignment(t *testing.T) {
	t.Parallel()

	declaration := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
		"var context = this", consistentThisOptions(t, "self"))
	rule_testing.ExpectFindings(t, declaration, "unexpectedAlias")
	source := declaration.SourceFile.Text()
	reported := source[declaration.Diagnostics[0].Range.Pos():declaration.Diagnostics[0].Range.End()]
	if reported != "context = this" {
		t.Errorf("expected the finding on the declarator, pointed at %q", reported)
	}

	assignment := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
		"self = 42", consistentThisOptions(t, "self"))
	rule_testing.ExpectFindings(t, assignment, "aliasNotAssignedToThis")
	assignmentSource := assignment.SourceFile.Text()
	assignmentReported := assignmentSource[assignment.Diagnostics[0].Range.Pos():assignment.Diagnostics[0].Range.End()]
	if assignmentReported != "self = 42" {
		t.Errorf("expected the finding on the assignment, pointed at %q", assignmentReported)
	}
}

// An arrow is neither a scope this rule judges nor transparent, and both intuitive readings are
// wrong in the same direction.
//
// Upstream hooks Program, FunctionExpression and FunctionDeclaration exits and nothing else, so an
// arrow body's scope is never visited and the declared-and-never-assigned judgment does not run
// inside one. The assignment judgment still does, because it hangs off the declarator and assignment
// listeners rather than off a scope. So `var f = () => { var self; }` is clean while
// `var f = () => { var self = 42; }` reports, and a port that treats an arrow as a scope of its own
// reports the first, as does a port that treats it as transparent.
//
// This port reported all three arrow rows before these fixtures existed, and nothing in upstream's
// corpus could see it: the corpus writes no arrow at all. Every expectation here was measured
// against the installed rule.
//
// The last row is the one that pins the descent resuming: a real function nested inside an arrow is
// a scope again, so its unassigned alias reports.
func TestConsistentThisTreatsAnArrowAsNeitherScopeNorTransparent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"an unassigned alias in an arrow is exempt", "var f = () => { var self; };", nil},
		{"even when a later write would have rescued it", "var f = () => { var self; self = this; };", nil},
		{"a top level arrow, same exemption", "() => { var self; };", nil},
		{"but a bad initializer still reports", "var f = () => { var self = 42; };", []string{"aliasNotAssignedToThis"}},
		{"and a capture under the wrong name still reports", "var f = () => { var context = this; };", []string{"unexpectedAlias"}},
		{"a correct capture in an arrow is clean", "var f = () => { var self = this; };", nil},
		{"an arrow writing an outer alias does not rescue it", "var self; var f = () => { self = this; };", []string{"aliasNotAssignedToThis"}},
		{"a function nested in an arrow is a scope again", "var f = () => { function g() { var self; } };", []string{"aliasNotAssignedToThis"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
				testCase.sourceText, consistentThisOptions(t, "self"))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// Function expressions are a scope, which the imported corpus cannot show.
//
// Its two nested-function cases write `(function() { self = this; }())` around a write to an OUTER
// alias, so they exercise the outer scope declining a foreign write rather than a function
// expression judging its own declarations. Dropping function expressions from the scope list
// survives the whole imported corpus and costs the first two rows here.
func TestConsistentThisJudgesAFunctionExpressionAsItsOwnScope(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"a function expression rescues its own alias", "var f = function() { var self; self = this; };", nil},
		{"and reports one it never assigns", "var f = function() { var self; };", []string{"aliasNotAssignedToThis"}},
		{"a bad initializer inside one", "var f = function() { var self = 42; };", []string{"aliasNotAssignedToThis"}},
		{"an object method is a function expression too", "var o = { m: function() { var self; } };", []string{"aliasNotAssignedToThis"}},
		{"an inner write does not rescue an outer alias", "var self; var f = function() { self = this; };", []string{"aliasNotAssignedToThis"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistentThis, consistentThisFile,
				testCase.sourceText, consistentThisOptions(t, "self"))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}
