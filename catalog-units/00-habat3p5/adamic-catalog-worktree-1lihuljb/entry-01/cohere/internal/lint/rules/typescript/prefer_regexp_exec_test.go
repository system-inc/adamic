package typescript

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// preferRegexpExecFile is where the fixtures pretend to live.
const preferRegexpExecFile = "/repository/source/PreferRegexpExec.ts"

// preferRegexpExecPointer spells an expected rewrite in a fixture row.
func preferRegexpExecPointer(value string) *string { return &value }

// preferRegexpExecCase is one imported corpus row.
type preferRegexpExecCase struct {
	sourceText string
	wantIds    []string
	wantOutput *string
}

// The corpus is typescript-eslint's own, extracted mechanically rather than retyped.
//
// `tests/rules/prefer-regexp-exec.test.ts` (clone, 8.69.0) was loaded with its RuleTester stubbed,
// then every case was replayed against the INSTALLED 8.67.0 rule through the ESLint Linter API,
// under upstream's own fixture tsconfig so the type graph is the one upstream's cases were written
// against. The expectations are what the installed rule answered, and they agree with upstream's own
// annotations exactly: 25 valid cases report nothing, 12 invalid cases report, zero disagree.
//
// This rule needs types, so the fixtures use the typed harness. The corpus is proof the checker is
// live rather than answering `any`: cases like `s: 'a' | 'b'` and `string & { __HTML_ESCAPED__: void }`
// only report when the type actually resolves to a string, and a run without a program makes the
// installed rule refuse to load at all rather than go quiet.
func preferRegexpExecCases() []preferRegexpExecCase {
	return []preferRegexpExecCase{
		{"'something'.match();", []string{}, nil},
		{"'something'.match(/thing/g);", []string{}, nil},
		{"\nconst text = 'something';\nconst search = /thing/g;\ntext.match(search);\n    ", []string{}, nil},
		{"\nconst match = (s: RegExp) => 'something';\nmatch(/thing/);\n    ", []string{}, nil},
		{"\nconst a = { match: (s: RegExp) => 'something' };\na.match(/thing/);\n    ", []string{}, nil},
		{"\nfunction f(s: string | string[]) {\n  s.match(/e/);\n}\n    ", []string{}, nil},
		{"(Math.random() > 0.5 ? 'abc' : 123).match(2);", []string{}, nil},
		{"'212'.match(2);", []string{}, nil},
		{"'212'.match(+2);", []string{}, nil},
		{"'oNaNo'.match(NaN);", []string{}, nil},
		{"'Infinity contains -Infinity and +Infinity in JavaScript.'.match(Infinity);", []string{}, nil},
		{"'Infinity contains -Infinity and +Infinity in JavaScript.'.match(+Infinity);", []string{}, nil},
		{"'Infinity contains -Infinity and +Infinity in JavaScript.'.match(-Infinity);", []string{}, nil},
		{"'void and null'.match(null);", []string{}, nil},
		{"\nconst matchers = ['package-lock.json', /regexp/];\nconst file = '';\nmatchers.some(matcher => !!file.match(matcher));\n    ", []string{}, nil},
		{"\nconst matchers = [/regexp/, 'package-lock.json'];\nconst file = '';\nmatchers.some(matcher => !!file.match(matcher));\n    ", []string{}, nil},
		{"\nconst matchers = [{ match: (s: RegExp) => false }];\nconst file = '';\nmatchers.some(matcher => !!file.match(matcher));\n    ", []string{}, nil},
		{"\nfunction test(pattern: string) {\n  'hello hello'.match(RegExp(pattern, 'g'))?.reduce(() => []);\n}\n    ", []string{}, nil},
		{"\nfunction test(pattern: string) {\n  'hello hello'.match(new RegExp(pattern, 'gi'))?.reduce(() => []);\n}\n    ", []string{}, nil},
		{"\nfunction findMatches(text: string, pattern: string, flags: string) {\n  return text.match(new RegExp(pattern, flags));\n}\n    ", []string{}, nil},
		{"\nconst matchCount = (str: string, re: RegExp) => {\n  return (str.match(re) || []).length;\n};\n    ", []string{}, nil},
		{"\nfunction test(str: string) {\n  str.match('[a-z');\n}\n    ", []string{}, nil},
		{"\nconst text = 'something';\ndeclare const search: RegExp;\ntext.match(search);\n      ", []string{}, nil},
		{"\nconst text = 'something';\ndeclare const obj: { search: RegExp };\ntext.match(obj.search);\n      ", []string{}, nil},
		{"\nconst text = 'something';\ndeclare function returnsRegexp(): RegExp;\ntext.match(returnsRegexp());\n      ", []string{}, nil},
		{"'something'.match(/thing/);", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("/thing/.exec('something');")},
		{"'something'.match('^[a-z]+thing/?$');", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("/^[a-z]+thing\\/?$/.exec('something');")},
		{"\nconst text = 'something';\nconst search = /thing/;\ntext.match(search);\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nconst text = 'something';\nconst search = /thing/;\nsearch.exec(text);\n      ")},
		{"\nconst text = 'something';\nconst search = 'thing';\ntext.match(search);\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nconst text = 'something';\nconst search = 'thing';\nRegExp(search).exec(text);\n      ")},
		{"\nfunction f(s: 'a' | 'b') {\n  s.match('a');\n}\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nfunction f(s: 'a' | 'b') {\n  /a/.exec(s);\n}\n      ")},
		{"\ntype SafeString = string & { __HTML_ESCAPED__: void };\nfunction f(s: SafeString) {\n  s.match(/thing/);\n}\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\ntype SafeString = string & { __HTML_ESCAPED__: void };\nfunction f(s: SafeString) {\n  /thing/.exec(s);\n}\n      ")},
		{"\nfunction f<T extends 'a' | 'b'>(s: T) {\n  s.match(/thing/);\n}\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nfunction f<T extends 'a' | 'b'>(s: T) {\n  /thing/.exec(s);\n}\n      ")},
		{"\nconst text = 'something';\nconst search = new RegExp('test', '');\ntext.match(search);\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nconst text = 'something';\nconst search = new RegExp('test', '');\nsearch.exec(text);\n      ")},
		{"\nfunction test(pattern: string) {\n  'check'.match(new RegExp(pattern, undefined));\n}\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nfunction test(pattern: string) {\n  new RegExp(pattern, undefined).exec('check');\n}\n      ")},
		{"\nfunction test(pattern: string) {\n  'check'.match(new RegExp(pattern));\n}\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nfunction test(pattern: string) {\n  new RegExp(pattern).exec('check');\n}\n      ")},
		{"\nfunction test(pattern: string) {\n  'check'.match(new RegExp(pattern, 'i'));\n}\n      ", []string{"regExpExecOverStringMatch"}, preferRegexpExecPointer("\nfunction test(pattern: string) {\n  new RegExp(pattern, 'i').exec('check');\n}\n      ")},
		{"\nfunction temp(text: string): void {\n  text.match(new RegExp(`${'hello'}`));\n  text.match(new RegExp(`${'hello'.toString()}`));\n}\n      ", []string{"regExpExecOverStringMatch", "regExpExecOverStringMatch"}, preferRegexpExecPointer("\nfunction temp(text: string): void {\n  new RegExp(`${'hello'}`).exec(text);\n  new RegExp(`${'hello'.toString()}`).exec(text);\n}\n      ")},
	}
}

// TestPreferRegexpExecMatchesUpstream replays the whole corpus.
func TestPreferRegexpExecMatchesUpstream(t *testing.T) {
	t.Parallel()

	cases := preferRegexpExecCases()

	reporting := 0
	for _, testCase := range cases {
		if len(testCase.wantIds) > 0 {
			reporting++
		}
	}
	if len(cases) != 37 {
		t.Fatalf("expected 37 corpus cases, have %d", len(cases))
	}
	if reporting != 12 {
		t.Fatalf("expected 12 reporting cases, have %d", reporting)
	}

	for _, testCase := range cases {
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile,
			testCase.sourceText)
		rule_testing.ExpectFindings(t, result, testCase.wantIds...)
	}
}

// TestPreferRegexpExecRewritesWhatUpstreamRewrites is the fixer's only real coverage.
func TestPreferRegexpExecRewritesWhatUpstreamRewrites(t *testing.T) {
	t.Parallel()

	checked := 0
	for _, testCase := range preferRegexpExecCases() {
		if testCase.wantOutput == nil {
			continue
		}
		checked++
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile,
			testCase.sourceText)
		// The typed harness writes each fixture to disk as `TrimSpace(source) + "\n"`, because a
		// real program is built from real files. So the rewritten text differs from upstream's
		// recorded `output` in leading and trailing whitespace and nowhere else, and the
		// expectation is normalised the same way rather than the corpus being edited.
		//
		// Only the OUTER whitespace is touched. Every newline and indent inside the fixture is
		// compared exactly, which is where a fixer's real mistakes live.
		rule_testing.ExpectFixedSource(t, result,
			strings.TrimSpace(*testCase.wantOutput)+"\n")
	}
	if checked != 12 {
		t.Fatalf("expected to check 12 rewrites, checked %d", checked)
	}
}

// TestPreferRegexpExecStaysSilentWhenTheGlobalFlagIsUnproven pins the direction of the guard.
//
// The rule reports only when it can PROVE the argument carries no `g`. That asymmetry is the whole
// safety property: reporting an unproven case rewrites a global match, which returns every match,
// into an `exec`, which returns one -- a behaviour change rather than a style change.
//
// Every expectation below was measured against the installed 8.67.0 build. The `let` case is the one
// a plausible implementation gets wrong: resolving the binding finds `/x/g` either way, and only
// refusing to trust a reassignable binding keeps the verdicts apart.
func TestPreferRegexpExecStaysSilentWhenTheGlobalFlagIsUnproven(t *testing.T) {
	t.Parallel()

	silent := []string{
		"declare const a: string;\na.match(/x/g);",
		"declare const a: string;\nconst r = /x/g;\na.match(r);",
		// A `let` bound to a GLOBAL regex, which is clean for the flag rather than for the keyword.
		// The non-global `let` reports; see the reassignment test below.
		"declare const a: string;\nlet r = /x/g;\na.match(r);",
		"declare function f(p: string): void;\ndeclare const a: string;\ndeclare const flags: string;\na.match(new RegExp('x', flags));",
		"declare const a: string;\ndeclare const r: RegExp;\na.match(r);",
	}
	for _, source := range silent {
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile, source)
		rule_testing.ExpectClean(t, result)
	}

	// The control: the same shapes with the flag provably absent all report, so the silences above
	// are the guard rather than a rule that never fires on a variable argument.
	reporting := []string{
		"declare const a: string;\na.match(/x/);",
		"declare const a: string;\nconst r = /x/;\na.match(r);",
		"declare const a: string;\na.match(new RegExp('x', 'i'));",
	}
	for _, source := range reporting {
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile, source)
		rule_testing.ExpectFindings(t, result, "regExpExecOverStringMatch")
	}
}

// TestPreferRegexpExecEvaluatesAReferenceRatherThanRecursingIntoIt pins the semantics that a
// differential caught and the corpus could not.
//
// Upstream settles a bound identifier by EVALUATING it with `getStaticValue` and requiring a real
// value back. The intuitive alternative -- resolve the binding, then ask the same "could this carry
// a g" question of whatever it was initialised with -- agrees on every one of upstream's 37 corpus
// rows and disagrees on real code: it reported 8 sites in this repository that upstream leaves
// alone, all of them `const r = new RegExp(templateWithSubstitution)`.
//
// The line is whether the value can be COMPUTED, not whether a `g` can be ruled out. Every
// expectation below was measured against the installed 8.67.0 build.
func TestPreferRegexpExecEvaluatesAReferenceRatherThanRecursingIntoIt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		report bool
	}{
		// Computable, so reportable.
		{"a bound regex literal", "declare const a: string;\nconst r = /x/;\na.match(r);", true},
		{"a bound string literal", "declare const a: string;\nconst s = 'thing';\na.match(s);", true},
		{"a bound concatenation", "declare const a: string;\nconst s = 'a' + 'b';\na.match(s);", true},
		{"a bound construction from literals", "declare const a: string;\nconst r = new RegExp('test', '');\na.match(r);", true},

		// Not computable, so not reportable -- the class that over-reported.
		{"a bound construction from a template", "declare const a: string;\ndeclare const t: string;\nconst r = new RegExp(`^${t}$`);\na.match(r);", false},
		{"a bound construction from a variable", "declare const a: string;\ndeclare const p: string;\nconst r = new RegExp(p);\na.match(r);", false},
		{"a declared const with no value", "declare const a: string;\ndeclare const s: string;\na.match(s);", false},

		// Computable and global, so not reportable for the original reason.
		{"a bound global regex", "declare const a: string;\nconst r = /x/g;\na.match(r);", false},
		{"a bound global construction", "declare const a: string;\nconst r = new RegExp('x', 'g');\na.match(r);", false},

		// An INLINE construction is judged syntactically instead, so the template that is clean when
		// bound reports when written in place. This asymmetry is upstream's and is the sharpest
		// evidence that a reference and a construction take different paths.
		{"an inline construction from a template", "declare const a: string;\ndeclare const t: string;\na.match(new RegExp(`^${t}$`));", true},
		{"an inline construction from a variable", "declare const a: string;\ndeclare const p: string;\na.match(new RegExp(p));", true},
	}
	for _, testCase := range cases {
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile, testCase.source)
		if testCase.report {
			rule_testing.ExpectFindings(t, result, "regExpExecOverStringMatch")
		} else {
			rule_testing.ExpectClean(t, result)
		}
	}
}

// TestPreferRegexpExecWrapsAWeakPrecedenceParent pins the outer half of upstream's wrapping fixer.
//
// When the call sits under an operator that binds more loosely, the whole replacement is
// parenthesised, because the new expression is not the one the surrounding code was written around.
// Upstream's corpus has no case in this shape -- all twelve of its reporting cases are statements
// with a bare receiver -- so these expectations come from driving the installed build.
func TestPreferRegexpExecWrapsAWeakPrecedenceParent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		want   string
	}{
		{"declare const a: string;\ndeclare const b: string;\nconst z = a.match(/x/) || b;",
			"declare const a: string;\ndeclare const b: string;\nconst z = (/x/.exec(a)) || b;"},
		{"declare const a: string;\nconst f = () => a.match(/x/);",
			"declare const a: string;\nconst f = () => /x/.exec(a);"},
		{"declare const a: string;\na.match(/x/);",
			"declare const a: string;\n/x/.exec(a);"},
	}
	for _, testCase := range cases {
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile, testCase.source)
		rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.want)+"\n")
	}
}

// TestPreferRegexpExecFoldsAWrittenOnceBindingRegardlessOfKeyword pins the rule that decides whether
// a bound regex can be trusted, and it is not the declaration keyword.
//
// `getStaticValue` folds a variable when its scope shows exactly ONE write, so `let` and `var` fold
// as readily as `const`, and any reassignment anywhere disqualifies the fold. A first draft here
// tested for `const` instead -- the intuitive conservative choice -- and was silent on
// `let r = /x/; a.match(r)`, which upstream reports. A mutation deleting the guard survived every
// corpus row, because upstream's 37 cases never bind a regex to a `let` at all.
//
// The last row is the one that rules out an ordering shortcut: a write AFTER the call site
// disqualifies the fold just as much as one before it, so this cannot be answered by looking only at
// what precedes the use.
func TestPreferRegexpExecFoldsAWrittenOnceBindingRegardlessOfKeyword(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		report bool
	}{
		{"const, written once", "declare const a: string;\nconst r = /x/;\na.match(r);", true},
		{"let, written once", "declare const a: string;\nlet r = /x/;\na.match(r);", true},
		{"var, written once", "declare const a: string;\nvar r = /x/;\na.match(r);", true},
		{"let, reassigned before use", "declare const a: string;\nlet r = /x/;\nr = /y/g;\na.match(r);", false},
		{"let, reassigned after use", "declare const a: string;\nlet r = /x/;\na.match(r);\nr = /y/g;", false},
		{"let, reassigned inside a function", "declare const a: string;\nlet r = /x/;\nfunction f() {\n  r = /y/g;\n}\na.match(r);", false},
	}
	for _, testCase := range cases {
		result := rule_testing.RunTyped(t, PreferRegexpExec, preferRegexpExecFile, testCase.source)
		if testCase.report {
			rule_testing.ExpectFindings(t, result, "regExpExecOverStringMatch")
		} else {
			rule_testing.ExpectClean(t, result)
		}
	}
}
