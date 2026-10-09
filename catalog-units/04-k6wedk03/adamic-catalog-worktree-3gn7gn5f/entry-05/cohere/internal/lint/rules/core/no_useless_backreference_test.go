package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// backreferenceFile is where the fixtures pretend to live.
const backreferenceFile = "/repository/source/Backreference.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from
// `oxc/crates/oxc_linter/src/rules/eslint/no_useless_backreference.rs`: 99 pass, 88 fail. The
// snapshot records 93 diagnostics from those 88 fail inputs, so one finding per input is wrong
// here. Five inputs report twice and each carries two ids below; the counts were recovered by
// aligning every snapshot diagnostic on the source line it prints, which is sound for this rule
// because all 88 inputs are distinct as text and the alignment came out a bijection.
//
// The ids are per problem kind rather than one shared id, because upstream emits five different
// sentences and the kind is the diagnosis. A single id would let a port that calls a
// different-alternative case a forward reference pass every fixture here.
func TestNoUselessBackreferenceFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   []string
	}{
		{`/(b)(\2a)/`, []string{"nested"}},
		{`/\k<foo>(?<foo>bar)/`, []string{"forward"}},
		{`RegExp('(a|bc)|\\1')`, []string{"disjunctive"}},
		{`new RegExp('(?!(?<foo>\\n))\\1')`, []string{"intoNegativeLookaround"}},
		{`/(?<!(a)\1)b/`, []string{"backward"}},
		{`new RegExp('(\\1)')`, []string{"nested"}},
		{`/^(a\1)$/`, []string{"nested"}},
		{`/^((a)\1)$/`, []string{"nested"}},
		{`new RegExp('^(a\\1b)$')`, []string{"nested"}},
		{`RegExp('^((\\1))$')`, []string{"nested"}},
		{`/((\2))/`, []string{"nested"}},
		{`/a(?<foo>(.)b\1)/`, []string{"nested"}},
		{`/a(?<foo>\k<foo>)b/`, []string{"nested"}},
		{`/^(\1)*$/`, []string{"nested"}},
		{`/^(?:a)(?:((?:\1)))*$/`, []string{"nested"}},
		{`/(?!(\1))/`, []string{"nested"}},
		{`/a|(b\1c)/`, []string{"nested"}},
		{`/(a|(\1))/`, []string{"nested"}},
		{`/(a|(\2))/`, []string{"nested"}},
		{`/(?:a|(\1))/`, []string{"nested"}},
		{`/(a)?(b)*(\3)/`, []string{"nested"}},
		{`/(?<=(a\1))b/`, []string{"nested"}},
		{`/\1(a)/`, []string{"forward"}},
		{`/\1.(a)/`, []string{"forward"}},
		{`/(?:\1)(?:(a))/`, []string{"forward"}},
		{`/(?:\1)(?:((a)))/`, []string{"forward"}},
		{`/(?:\2)(?:((a)))/`, []string{"forward"}},
		{`/(?:\1)(?:((?:a)))/`, []string{"forward"}},
		{`/(\2)(a)/`, []string{"forward"}},
		{`RegExp('(a)\\2(b)')`, []string{"forward"}},
		{`/(?:a)(b)\2(c)/`, []string{"forward"}},
		{`/\k<foo>(?<foo>a)/`, []string{"forward"}},
		{`/(?:a(b)\2)(c)/`, []string{"forward"}},
		{`new RegExp('(a)(b)\\3(c)')`, []string{"forward"}},
		{`/\1(?<=(a))./`, []string{"forward"}},
		{`/\1(?<!(a))./`, []string{"forward"}},
		{`/(?<=\1)(?<=(a))/`, []string{"forward"}},
		{`/(?<!\1)(?<!(a))/`, []string{"forward"}},
		{`/(?=\1(a))./`, []string{"forward"}},
		{`/(?!\1(a))./`, []string{"forward"}},
		{`/(?<=(a)\1)b/`, []string{"backward"}},
		{`/(?<!.(a).\1.)b/`, []string{"backward"}},
		{`/(.)(?<!(b|c)\2)d/`, []string{"backward"}},
		{`/(?<=(?:(a)\1))b/`, []string{"backward"}},
		{`/(?<=(?:(a))\1)b/`, []string{"backward"}},
		{`/(?<=(a)(?:\1))b/`, []string{"backward"}},
		{`/(?<!(?:(a))(?:\1))b/`, []string{"backward"}},
		{`/(?<!(?:(a))(?:\1)|.)b/`, []string{"backward"}},
		{`/.(?!(?<!(a)\1))./`, []string{"backward"}},
		{`/.(?=(?<!(a)\1))./`, []string{"backward"}},
		{`/.(?!(?<=(a)\1))./`, []string{"backward"}},
		{`/.(?=(?<=(a)\1))./`, []string{"backward"}},
		{`/(a)|\1b/`, []string{"disjunctive"}},
		{`/^(?:(a)|\1b)$/`, []string{"disjunctive"}},
		{`/^(?:(a)|b(?:c|\1))$/`, []string{"disjunctive"}},
		{`/^(?:a|b(?:(c)|\1))$/`, []string{"disjunctive"}},
		{`/^(?:(a(?!b))|\1b)+$/`, []string{"disjunctive"}},
		{`/^(?:(?:(a)(?!b))|\1b)+$/`, []string{"disjunctive"}},
		{`/^(?:(a(?=a))|\1b)+$/`, []string{"disjunctive"}},
		{`/^(?:(a)(?=a)|\1b)+$/`, []string{"disjunctive"}},
		{`/.(?:a|(b)).|(?:(\1)|c)./`, []string{"disjunctive"}},
		{`/.(?!(a)|\1)./`, []string{"disjunctive"}},
		{`/.(?<=\1|(a))./`, []string{"disjunctive"}},
		{`/a(?!(b)).\1/`, []string{"intoNegativeLookaround"}},
		{`/(?<!(a))b\1/`, []string{"intoNegativeLookaround"}},
		{`/(?<!(a))(?:\1)/`, []string{"intoNegativeLookaround"}},
		{`/.(?<!a|(b)).\1/`, []string{"intoNegativeLookaround"}},
		{`/.(?!(a)).(?!\1)./`, []string{"intoNegativeLookaround"}},
		{`/.(?<!(a)).(?<!\1)./`, []string{"intoNegativeLookaround"}},
		{`/.(?=(?!(a))\1)./`, []string{"intoNegativeLookaround"}},
		{`/.(?<!\1(?!(a)))/`, []string{"intoNegativeLookaround"}},
		{`/\1(a)(b)\2/`, []string{"forward"}},
		{`/\1(a)\1/`, []string{"forward"}},
		{`/\1(a)\2(b)/`, []string{"forward", "forward"}},
		{`/\1.(?<=(a)\1)/`, []string{"forward", "backward"}},
		{`/(?!\1(a)).\1/`, []string{"forward", "intoNegativeLookaround"}},
		// The call is found from the file, before the walk reaches the literal, so its finding comes first.
		{`/(a)\2(b)/; RegExp('(\\1)');`, []string{"nested", "forward"}},
		{`RegExp('\\1(a){', flags);`, []string{"forward"}},
		{`new RegExp('\\1([[A--B]])', 'v')`, []string{"forward"}},
		{`/\k<foo>((?<foo>bar)|(?<foo>baz))/`, []string{"forward"}},
		// Among a name's groups, a problem in the reference's own branch is the one named (ESLint's corpus).
		{`/((?<foo>bar)|\k<foo>(?<foo>baz))/`, []string{"forward"}},
		{`/\k<foo>((?<foo>bar)|(?<foo>baz)|(?<foo>qux))/`, []string{"forward"}},
		{`/((?<foo>bar)|\k<foo>(?<foo>baz)|(?<foo>qux))/`, []string{"forward"}},
		{`/((?<foo>bar)|\k<foo>|(?<foo>baz))/`, []string{"disjunctive"}},
		{`/((?<foo>bar)|\k<foo>|(?<foo>baz)|(?<foo>qux))/`, []string{"disjunctive"}},
		{`/((?<foo>bar)|(?<foo>baz\k<foo>)|(?<foo>qux\k<foo>))/`, []string{"nested", "nested"}},
		{`/(?<=((?<foo>bar)|(?<foo>baz))\k<foo>)/`, []string{"backward"}},
		{`/((?!(?<foo>bar))|(?!(?<foo>baz)))\k<foo>/`, []string{"intoNegativeLookaround"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, testCase.sourceText),
				testCase.findings...)
		})
	}
}

// The clean cases are the whole discrimination and they are what catches a port.
//
// Several groups are worth naming because each encodes a distinction that is easy to miss:
// a callee that is not the global RegExp (`regExp`, `Regexp`, `RegExp.foo`, `foo.RegExp`); a
// pattern the rule cannot read (an identifier, a concatenation, a template with substitutions);
// an escape that is not a backreference at all (`\0`, `\11` with one group, a `\1` inside a
// character class, a doubled backslash); a backreference that is genuinely reachable, including
// the many lookaround shapes where the group does run before the reference; and patterns whose
// syntax is broken, which upstream skips because its parser refuses them.
func TestNoUselessBackreferenceStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{`'\1(a)'`},
		{`regExp('\\1(a)')`},
		{`new Regexp('\\1(a)', 'u')`},
		{`RegExp.foo('\\1(a)', 'u')`},
		{`new foo.RegExp('\\1(a)')`},
		{`RegExp(p)`},
		{`new RegExp(p, 'u')`},
		{`RegExp('\\1(a)' + suffix)`},
		{"new RegExp(`${prefix}\\\\1(a)`)"},
		{`/(?:)/`},
		{`/(?:a)/`},
		{`new RegExp('')`},
		{`RegExp('(?:a)|(?:b)*')`},
		{`/^ab|[cd].\n$/`},
		{`/(a)/`},
		{`RegExp('(a)|(b)')`},
		{`new RegExp('\\n\\d(a)')`},
		{`/\0(a)/`},
		{`/\0(a)/u`},
		{`/(?<=(a))(b)(?=(c))/`},
		{`/(?<!(a))(b)(?!(c))/`},
		{`/(?<foo>a)/`},
		{`RegExp('\1(a)')`},
		{`RegExp('\\\\1(a)')`},
		{`/\\1(a)/`},
		{`/\1/`},
		{`/^\1$/`},
		{`/\2(a)/`},
		{`/\1(?:a)/`},
		{`/\1(?=a)/`},
		{`/\1(?!a)/`},
		{`/^[\1](a)$/`},
		{`new RegExp('[\\1](a)')`},
		{`/\11(a)/`},
		{`/\k<foo>(a)/`},
		{`/^(a)\1\\2$/`},
		{`/(a)\1/`},
		{`/(a).\1/`},
		{`RegExp('(a)\\1(b)')`},
		{`/(a)(b)\2(c)/`},
		{`/(?<foo>a)\k<foo>/`},
		{`new RegExp('(.)\\1')`},
		{`RegExp('(a)\\1(?:b)')`},
		{`/(a)b\1/`},
		{`/((a)\2)/`},
		{`/((a)b\2c)/`},
		{`/^(?:(a)\1)$/`},
		{`/^((a)\2)$/`},
		{`/^(((a)\3))|b$/`},
		{`/a(?<foo>(.)b\2)/`},
		{`/(a)?(b)*(\1)(c)/`},
		{`/(a)?(b)*(\2)(c)/`},
		{`/(?<=(a))b\1/`},
		{`/(?<=(?=(a)\1))b/`},
		{`/(?<!\1(a))b/`},
		{`/(?<=\1(a))b/`},
		{`/(?<!\1.(a))b/`},
		{`/(?<=\1.(a))b/`},
		{`/(?<=(?:\1.(a)))b/`},
		{`/(?<!(?:\1)((a)))b/`},
		{`/(?<!(?:\2)((a)))b/`},
		{`/(?=(?<=\1(a)))b/`},
		{`/(?=(?<!\1(a)))b/`},
		{`/(.)(?<=\2(a))b/`},
		{`/^(a)\1|b/`},
		{`/^a|(b)\1/`},
		{`/^a|(b|c)\1/`},
		{`/^(a)|(b)\2/`},
		{`/^(?:(a)|(b)\2)$/`},
		{`/^a|(?:.|(b)\1)/`},
		{`/^a|(?:.|(b).(\1))/`},
		{`/^a|(?:.|(?:(b)).(\1))/`},
		{`/^a|(?:.|(?:(b)|c).(\1))/`},
		{`/^a|(?:.|(?:(b)).(\1|c))/`},
		{`/^a|(?:.|(?:(b)|c).(\1|d))/`},
		{`/.(?=(b))\1/`},
		{`/.(?<=(b))\1/`},
		{`/a(?!(b)\1)./`},
		{`/a(?<!\1(b))./`},
		{`/a(?!(b)(\1))./`},
		{`/a(?!(?:(b)\1))./`},
		{`/a(?!(?:(b))\1)./`},
		{`/a(?<!(?:\1)(b))./`},
		{`/a(?<!(?:(?:\1)(b)))./`},
		{`/(?<!(a))(b)(?!(c))\2/`},
		{`/a(?!(b|c)\1)./`},
		{`RegExp('\\1(a)[')`},
		{`new RegExp('\\1(a){', 'u')`},
		{`new RegExp('\\1(a)\\2', 'ug')`},
		{`RegExp('\\1(a)\\k<foo>', 'u')`},
		{`new RegExp('\\k<foo>(?<foo>a)\\k<bar>')`},
		{`new RegExp('([[A--B]])\\1', 'v')`},
		{`new RegExp('[[]\\1](a)', 'v')`},
		{`/((?<foo>bar)\k<foo>|(?<foo>baz))/`},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, testCase.sourceText))
		})
	}
}

// TestNoUselessBackreferenceStillReportsBesideALegacyOctal covers a case upstream's corpus does
// not reach.
//
// Written for a surviving mutant. The check that a numeric reference past the last group only
// invalidates the pattern under `u` or `v` could be dropped entirely and every one of the 187
// upstream cases still passed, because none of them puts a useless backreference and an
// out-of-range escape in the same flagless pattern. Without the flag `\2` here is a legacy octal
// and the pattern is valid (confirmed against node), so the useless `\1` must still report; with
// the flag the same text is a syntax error and the whole pattern is upstream's business, not this
// rule's.
func TestNoUselessBackreferenceStillReportsBesideALegacyOctal(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, `/\1(a)\2/`)
	rule_testing.ExpectFindings(t, result, "forward")

	// The same text under `u` is a syntax error, so the rule steps aside entirely.
	rule_testing.ExpectClean(t,
		rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, `/\1(a)\2/u`))
}

// TestNoUselessBackreferenceHandlesConstructorArgumentShapes covers arguments the corpus leaves untested.
//
// A regex literal handed to RegExp is a literal like any other and is checked under its own flags, and
// the call is checked too, its pattern being the literal's text as String() makes it, `/\1(a)/`. So the
// pattern reports twice, at the literal and at the call, which is what ESLint does with these exact
// inputs.
//
// Flags that cannot be read check the pattern under none, as ESLint's `flags || ""` does, whether the
// argument is an identifier nothing gives a constant value or a template with a substitution. Flags held
// in a constant binding are read: `const flags = 'gus'` puts the pattern under `u`, where its brace is a
// syntax error and the rule steps aside.
func TestNoUselessBackreferenceHandlesConstructorArgumentShapes(t *testing.T) {
	t.Parallel()

	t.Run("a regex literal argument reports at the literal and at the call", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, `RegExp(/\1(a)/, '');`)
		rule_testing.ExpectFindings(t, result, "forward", "forward")
	})

	for _, sourceText := range []string{
		`RegExp('\\1(a)', flags);`,
		"RegExp('\\\\1(a)', `${flags}`);",
	} {
		t.Run("unreadable flags check under none: "+sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, sourceText), "forward")
		})
	}

	t.Run("flags in a constant binding are read", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile,
			`const flags = 'gus'; RegExp('\\1(a){', flags);`))
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile,
			`let flags = 'u'; flags = ''; RegExp('\\1(a){', flags);`), "forward")
	})
}

// TestNoUselessBackreferenceIgnoresALocalRegExp pins that a local named RegExp is not the global.
//
// These five are ESLint's own pass cases. The port used to report them, since it read the callee's
// spelling and had no checker to ask what the name was bound to; the shelf's ReferenceTracker asks, and
// follows the global through an alias or `globalThis` as well, which the table covers.
func TestNoUselessBackreferenceIgnoresALocalRegExp(t *testing.T) {
	t.Parallel()

	cases := []string{
		`let RegExp; new RegExp('\\1(a)');`,
		`function foo() { var RegExp; RegExp('\\1(a)', 'u'); }`,
		"function foo() { var RegExp; RegExp('\\\\1(a)', `u`); }",
		`function foo(RegExp) { new RegExp('\\1(a)'); }`,
		`if (foo) { const RegExp = bar; RegExp('\\1(a)'); }`,
	}
	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, sourceText))
		})
	}
}

// TestNoUselessBackreferenceDeclinesAPatternABuiltinBuilds pins the one place this port departs from
// ESLint. ESLint evaluates `String.raw`...“ by calling String.raw, and reports the call below;
// reference.ConstantStringIn evaluates no built-in call, so the pattern is not a constant here and the
// call is not checked, the silent direction.
func TestNoUselessBackreferenceDeclinesAPatternABuiltinBuilds(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile,
		"RegExp(String.raw`\\1(a)`);"))
}

// TestNoUselessBackreferencePointsAtThePattern asserts where each finding lands, which ExpectFindings
// cannot see: the whole regex literal, or the whole RegExp call, as ESLint reports it. The message names
// the backreference, so two useless references in one pattern are two findings at one place.
func TestNoUselessBackreferencePointsAtThePattern(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		reported   []string
	}{
		{`/\1(a)/`, []string{`/\1(a)/`}},
		{`x = /(a|\1b)/g`, []string{`/(a|\1b)/g`}},
		{`/\1(a)\2(b)/`, []string{`/\1(a)\2(b)/`, `/\1(a)\2(b)/`}},
		{`RegExp('(a)\\2(b)')`, []string{`RegExp('(a)\\2(b)')`}},
		{`new RegExp('\\1([[A--B]])', 'v')`, []string{`new RegExp('\\1([[A--B]])', 'v')`}},
		{"foo;\nnew RegExp(\n  '\\\\1(a)'\n)", []string{"new RegExp(\n  '\\\\1(a)'\n)"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("expected %d findings, got %d", len(testCase.reported), len(result.Diagnostics))
			}
			source := result.SourceFile.Text()
			for index, want := range testCase.reported {
				diagnostic := result.Diagnostics[index]
				if got := source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != want {
					t.Errorf("finding %d points at %q, expected %q", index, got, want)
				}
			}
		})
	}
}

// TestNoUselessBackreferenceEveryPathStartsAtTheRoot pins the invariant that lets the
// lowest-common-ancestor walk start at index 1.
//
// Written for a surviving mutant. Starting that walk at 0 instead of 1 survived the whole corpus,
// and reading the code says why: the root node is pushed before the scan loop and never popped
// inside it, so every recorded path begins with index 0 and the comparison at 0 is always 0 == 0.
// The mutant is therefore equivalent, and the only input that could distinguish the two is a path
// not beginning at the root, which the scanner cannot produce. This asserts that rather than
// leaving it as an argument, so a later change that pushes something before the root fails here
// instead of silently changing what the walk means.
// Every path recorded by the scanner must begin with the root node, index 0. That invariant is
// what makes starting the lowest-common-ancestor walk at 1 rather than 0 the same computation:
// the comparison at 0 is 0 == 0 for every pair of paths, so it can only ever advance.
func TestNoUselessBackreferenceEveryPathStartsAtTheRoot(t *testing.T) {
	t.Parallel()

	patterns := []string{
		`\1(a)`, `(b)(\2a)`, `(a|\1b)`, `\1(?!(a))`, `(?<=(a)\1)b`, `\k<foo>(?<foo>bar)`,
		`((?<foo>bar)|\k<foo>(?<foo>baz)|(?<foo>qux))`, `(?<=((?<foo>bar)|(?<foo>baz))\k<foo>)`,
		`.(?=(?!(a))\1).`, `^(?:(a(?!b))|\1b)+$`, `\1([[A--B]])`, `[\1](a)`,
	}
	for _, pattern := range patterns {
		for _, flagText := range []string{"", "u", "v"} {
			structure, ok := scanRegexStructure(pattern, regexsyntax.ParseRegexFlags(flagText))
			if !ok {
				continue
			}
			for _, group := range structure.groups {
				if len(group.path) == 0 || group.path[0] != 0 {
					t.Errorf("group path %v in %q flags %q does not start at the root", group.path, pattern, flagText)
				}
			}
			for _, reference := range structure.backreferences {
				if len(reference.path) == 0 || reference.path[0] != 0 {
					t.Errorf("reference path %v in %q flags %q does not start at the root", reference.path, pattern, flagText)
				}
			}
		}
	}
}

// TestNoUselessBackreferenceStaysSilentOnAPatternTheEngineWouldRefuse covers the scanner's error
// paths, which the corpus reaches only glancingly.
//
// Written for three surviving mutants: dropping the balanced-parenthesis check, accepting a `{`
// with no digits as a quantifier, and ignoring a failed escape skip. All three are reachable and
// none is defensive, because a pattern arrives here as a *string* and a string can hold anything.
// `RegExp('\\1(a')` is a runtime error the engine raises, not a lint finding, and a scanner that
// walks it anyway builds a group list whose spans are nonsense and then reports against them.
//
// Upstream reaches two of these (`RegExp('\\1(a)[')` is in its pass vector) and stops there, so
// the unbalanced and trailing-backslash shapes are ours.
func TestNoUselessBackreferenceStaysSilentOnAPatternTheEngineWouldRefuse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an unclosed group", `RegExp('\\1(a')`},
		{"an unopened group", `RegExp('\\1(a))')`},
		{"a trailing backslash", `RegExp('(a)\\1\\')`},
		{"an unterminated class", `RegExp('(a)\\1[')`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, testCase.sourceText))
		})
	}

	// A brace that opens no quantifier is a literal brace without `u`, so the pattern is valid and
	// the useless backreference in it still reports. This is the contrast that makes the digit
	// check in opensQuantifier a real decision rather than an unreachable guard: read it as a
	// quantifier and the pattern is judged well-formed under `u` when it is not.
	t.Run("a literal brace does not suppress the finding", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t,
			rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, `RegExp('\\1(a){,}')`),
			"forward")
	})
	t.Run("the same brace under u is a syntax error", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t,
			rule_testing.RunTyped(t, NoUselessBackreference, backreferenceFile, `RegExp('\\1(a){,}', 'u')`))
	})
}

// The registry entry, so a rule that passes every fixture and lints zero files fails here.
//
// The NeedsTypeChecker assertion is the negative on purpose. This rule is purely syntactic: it
// reads pattern text and the shape of the call around it, and never asks what a name binds to.
// Declaring the checker would take an exclusive per-file lock for nothing. The one question that
// would need it, whether a `RegExp` callee is locally shadowed, is the stated divergence above.
func TestNoUselessBackreferenceIsRegistered(t *testing.T) {
	t.Parallel()

	var found *rule.Registration
	for index, registration := range rule.Registered() {
		if registration.Rule.Name == "no-useless-backreference" {
			found = &rule.Registered()[index]
			break
		}
	}
	if found == nil {
		t.Fatal("no-useless-backreference is absent from the registry")
	}
	// The global RegExp is told from a local one, and a constant argument followed to its binding,
	// through the checker, so a registered copy without it would find no call at all.
	if !found.Rule.NeedsTypeChecker {
		t.Error("the registered copy does not declare NeedsTypeChecker, so it would check no RegExp call")
	}
}
