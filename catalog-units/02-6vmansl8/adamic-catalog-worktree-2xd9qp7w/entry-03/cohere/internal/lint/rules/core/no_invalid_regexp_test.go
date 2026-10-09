package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noInvalidRegexpFile = "/repository/source/Patterns.ts"

func TestNoInvalidRegexpFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A pattern that cannot parse. Written as a literal the compiler would refuse it; written
		// as a string it reaches runtime and throws there.
		{"an unclosed character class", "export const a = new RegExp('[');\n"},
		{"an unclosed group", "export const a = new RegExp('(');\n"},
		{"a quantifier with nothing to repeat", "export const a = new RegExp('*');\n"},
		{"a backwards range", "export const a = new RegExp('[z-a]');\n"},

		// Called rather than constructed. RegExp is one of the few constructors that works both
		// ways, and a rule listening only for `new` is silent on half the calls anyone writes.
		{"a plain call rather than new", "export const a = RegExp('[');\n"},

		// Flags. An unknown flag is a SyntaxError rather than something the engine ignores.
		{"an unknown flag", "export const a = new RegExp('.', 'z');\n"},
		{"a duplicated flag", "export const a = new RegExp('.', 'ii');\n"},

		// u and v are mutually exclusive, and this is the case the ordering inside the flag check
		// exists for: both are individually valid, so a checker that struck valid flags first and
		// asked no further question would accept the pair.
		{"u and v together", "export const a = new RegExp('.', 'uv');\n"},

		// Valid without u, invalid with it. Pins that the flags are applied to the pattern rather
		// than checked beside it: `\p` is a property escape only under u, and a lone `{` after it
		// is a syntax error there and a literal brace otherwise.
		{"a pattern the u flag invalidates", "export const a = new RegExp('\\\\p{', 'u');\n"},

		// The default of `allowConstructorFlags`. Under `["a"]` this exact line is clean, and with
		// nothing configured the allow-list is empty, so it must report. Every other case here
		// exercises flags no allow-list would name; this one pins the unconfigured branch.
		{"a flag only an allow-list could excuse", "export const a = new RegExp('.', 'a');\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoInvalidRegexp, noInvalidRegexpFile, testCase.sourceText),
				"invalidRegexp")
		})
	}
}

func TestNoInvalidRegexpStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shapes the rule is asking for.
		{"a valid pattern", "export const a = new RegExp('[a-z]+');\n"},
		{"a valid pattern with flags", "export const a = new RegExp('[a-z]+', 'gi');\n"},
		{"every valid flag at once", "export const a = new RegExp('.', 'dgimsuy');\n"},

		// No arguments produces the empty pattern and cannot fail.
		{"no arguments at all", "export const a = new RegExp();\n"},

		// Not a literal, so what it holds at runtime is unknowable. Reporting it would flag correct
		// code, which is worse than missing a defect.
		{"a pattern from a variable", "export const a = new RegExp(source);\n"},
		{"flags from a variable", "export const a = new RegExp('[a-z]', flags);\n"},

		// Unknown flags with a pattern that some reachable flag combination accepts. `\p{L}` is
		// valid under u, so the variable may well supply it, and this must stay silent. It is the
		// counterpart to the u-invalidates case above and the reason that branch is not simply
		// "check without flags".
		{"a u-only pattern with unknown flags", "export const a = new RegExp('\\\\p{L}', flags);\n"},

		// The case that separates "invalid under every reachable flag set" from "invalid under
		// any". `\p{` is a malformed property escape under u and v, and a literal backslash-p
		// followed by a brace without them, so a checker asking "any" would report a pattern that
		// is fine the moment the flags variable holds something other than u or v.
		//
		// Added after mutation testing: weakening the guard from && to || broke nothing, because
		// every silent case here was valid under all three flag sets and never reached the branch.
		// A clean case positioned where the branch cannot be reached passes while measuring nothing.
		{"a pattern only some flag sets reject", "export const a = new RegExp('\\\\p{', flags);\n"},

		// A different identifier that happens to take similar arguments.
		{"a constructor that is not RegExp", "export const a = new Other('[');\n"},
		{"a call that is not RegExp", "export const a = other('[');\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoInvalidRegexp, noInvalidRegexpFile, testCase.sourceText))
		})
	}
}

// A local named RegExp is whatever the caller passed, and a pattern under `v` has a grammar the engine
// does not parse, so both stay silent, as ESLint's corpus pins the first and its validator accepts the
// second (#jjfa7qb). The global with a broken pattern still reports, so the silence is not the rule
// going quiet.
func TestNoInvalidRegexpAsksForTheGlobalAndLeavesVPatternsAlone(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoInvalidRegexp, noInvalidRegexpFile,
		"function foo(RegExp: (pattern: string) => void) { RegExp('['); }"))
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoInvalidRegexp, noInvalidRegexpFile,
		"new RegExp('[A--[0-9]]', 'v');"))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoInvalidRegexp, noInvalidRegexpFile,
		"new RegExp('[');"), "invalidRegexp")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoInvalidRegexp, noInvalidRegexpFile,
		"new RegExp('.', 'vz');"), "invalidRegexp")
}

// allowConstructorFlags read the way upstream reads it: extra flags struck once each like the
// language's, case-sensitive, joined across strings, and a language flag in the list changing nothing.
// The rows are ESLint's own, so each one is a verdict upstream has already given.
func TestNoInvalidRegexpHonorsAllowConstructorFlags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		options    string
		sourceText string
		reports    bool
	}{
		{"an allowed flag alone", `{"allowConstructorFlags": ["a"]}`, "new RegExp('.', 'a');\n", false},
		{"an allowed flag beside a valid one", `{"allowConstructorFlags": ["a"]}`, "new RegExp('.', 'ga');\n", false},
		{"two allowed flags in any order", `{"allowConstructorFlags": ["a", "z"]}`, "new RegExp('.', 'zga');\n", false},
		{"one string allowing two flags", `{"allowConstructorFlags": ["az"]}`, "new RegExp('.', 'za');\n", false},
		{"an allowed flag with an unknown pattern", `{"allowConstructorFlags": ["a"]}`, "new RegExp(pattern, 'ga');\n", false},
		{"an empty allow-list", `{"allowConstructorFlags": []}`, "new RegExp('.', 'a');\n", true},
		{"an empty options object", `{}`, "new RegExp('.', 'a');\n", true},
		{"a flag the list does not name", `{"allowConstructorFlags": ["a"]}`, "new RegExp('.', 'z');\n", true},
		{"the other case of an allowed flag", `{"allowConstructorFlags": ["a"]}`, "RegExp('.', 'A');\n", true},
		{"an allowed flag twice", `{"allowConstructorFlags": ["a"]}`, "new RegExp('.', 'aga');\n", true},
		{"an allowed flag twice with a non-literal pattern", `{"allowConstructorFlags": ["a"]}`, "new RegExp(pattern, 'aa');\n", true},

		// A language flag in the list is dropped rather than added a second time, so it is still
		// struck once and a second copy is still a duplicate.
		{"a language flag listed and doubled", `{"allowConstructorFlags": ["u"]}`, "new RegExp('.', 'uu');\n", true},

		// An allowed flag carries no meaning into the pattern, so `\u{0}` is still read without `u`.
		{"an allowed flag over a pattern only u accepts", `{"allowConstructorFlags": ["a"]}`, "new RegExp('\\\\u{0}*', 'a');\n", true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options, err := DecodeNoInvalidRegexpOptions([]byte(testCase.options))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.options, err)
			}
			result := rule_testing.RunTypedWithOptions(t, NoInvalidRegexp, noInvalidRegexpFile, testCase.sourceText, options)
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "invalidRegexp")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// The decoder holds upstream's schema: an object with one array of unique strings and nothing else.
func TestNoInvalidRegexpDecodesOnlyUpstreamsSchema(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		refused bool
	}{
		{"an empty object", `{}`, false},
		{"an empty list", `{"allowConstructorFlags": []}`, false},
		{"a list of flags", `{"allowConstructorFlags": ["a", "z"]}`, false},
		{"an unknown key", `{"allowConstructorFlag": ["a"]}`, true},
		{"a key in the wrong case", `{"AllowConstructorFlags": ["a"]}`, true},
		{"an item that is not a string", `{"allowConstructorFlags": [1]}`, true},
		{"a repeated item", `{"allowConstructorFlags": ["a", "a"]}`, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeNoInvalidRegexpOptions([]byte(testCase.raw))
			if testCase.refused && err == nil {
				t.Errorf("%s decoded, and upstream's schema refuses it", testCase.raw)
			}
			if !testCase.refused && err != nil {
				t.Errorf("%s was refused, and upstream's schema accepts it: %v", testCase.raw, err)
			}
		})
	}
}
