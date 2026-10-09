package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const selfAssignFile = "/repository/source/Assign.ts"

const selfAssignDeclarations = "declare let a: any, b: any, c: any, i: number;\n" +
	"declare const o: any;\ndeclare function f(): any;\n"

func TestNoSelfAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{"an identifier to itself", "export function run() { a = a; }\n", 1},
		{"a logical-and assignment", "export function run() { a &&= a; }\n", 1},
		{"a logical-or assignment", "export function run() { a ||= a; }\n", 1},
		{"a nullish assignment", "export function run() { a ??= a; }\n", 1},

		// The props half. These are the shapes the option exists for and the reason token equality
		// is not enough: the tokens differ in two of the three.
		{"a member to itself", "export function run() { o.b = o.b; }\n", 1},
		{"a deep member chain", "export function run() { o.b.c = o.b.c; }\n", 1},
		{"optional chaining on one side", "export function run() { o.b = o?.b; }\n", 1},
		{"a static computed key against a dotted one", "export function run() { o.b = o['b']; }\n", 1},
		{"a dotted key against a static computed one", "export function run() { o['b'] = o.b; }\n", 1},
		{"a numeric computed key", "export function run() { o[0] = o[0]; }\n", 1},
		{"a regular expression key against its quoted spelling", "export function run() { o['/x/g'] = o[/x/g]; }\n", 1},
		{"a null key against its dotted spelling", "export function run() { o[null] = o.null; }\n", 1},

		// An identifier subscript read twice with nothing between is the same reference, which is
		// upstream's answer and the one its docs give for `a[b] = a[b]`. Silent here once, on the
		// theory that i might change between the reads; nothing in the statement can change it.
		{"an identifier subscript", "export function run() { o[i] = o[i]; }\n", 1},
		{"an identifier subscript inside a chain", "export function run() { o[i].b = o[i].b; }\n", 1},
		{"a private name", "export class C { #p = 1; run() { this.#p = this.#p; } }\n", 1},

		// Destructuring, compared positionally.
		{"an array pattern", "export function run() { [a, b] = [a, b]; }\n", 2},
		{"one matching element of an array pattern", "export function run() { [a, b] = [a, c]; }\n", 1},
		{"a nested array pattern", "export function run() { [a, [b]] = [a, [b]]; }\n", 2},
		{"a matching rest and spread", "export function run() { [a, ...b] = [a, ...b]; }\n", 2},
		// The rest guard truncates at the rest, it does not discard what came before it. `a = a` at
		// index 0 is a real self-assignment whatever the tail looks like, so this reports once and
		// stops. Belongs here rather than in the silent table, which is where I first wrote it.
		{"a leading match before a longer spread", "export function run() { [a, ...b] = [a, ...b, 1]; }\n", 1},

		// Object destructuring, compared by name rather than position.
		{"an object pattern", "export function run() { ({ a, b } = { a, b }); }\n", 2},
		{"an object pattern with the names reordered", "export function run() { ({ a, b } = { b, a }); }\n", 2},
		{"an explicit property", "export function run() { ({ a: a } = { a: a }); }\n", 1},
		{"a string key matching a shorthand", "export function run() { ({ a } = { 'a': a }); }\n", 1},
		{"a computed template key", "export function run() { ({ 'a': b } = { [`a`]: b }); }\n", 1},

		// Only properties after the last spread can be claimed, and these are after it.
		{"a property after a spread", "export function run() { ({ a } = { ...b, a }); }\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Spelled as an explicit list rather than a count, so this reads as ExpectFindings
			// like every other rule's pair. The count form was equivalent and TestEveryRuleShips-
			// AFixturePair could not see it: the guard looks for a call to ExpectFindings, and a
			// hand-rolled length check proves the same thing while being invisible to the thing
			// that checks rules can be proven to fire.
			expected := make([]string, testCase.wantCount)
			for index := range expected {
				expected[index] = "selfAssignment"
			}
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoSelfAssign, selfAssignFile,
				selfAssignDeclarations+testCase.sourceText), expected...)
		})
	}
}

func TestNoSelfAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"different identifiers", "export function run() { a = b; }\n"},
		{"different members", "export function run() { o.b = o.c; }\n"},
		{"the same property on different objects", "export function run() { a.b = b.b; }\n"},

		// The swap, which is why array comparison is positional rather than by membership. This is
		// the single most important silent case in the rule.
		{"a swap", "export function run() { [a, b] = [b, a]; }\n"},
		{"a three-way rotation", "export function run() { [a, b, c] = [b, c, a]; }\n"},

		// A rest target absorbs everything remaining, so it does not match a spread followed by
		// more elements: the target ends up longer than it was. ESLint's own comment names this
		// exact shape.
		//
		// The rest must be the first element for this to be a clean case. My first version put `a`
		// before it, and that fixture failed correctly: at index 0 there is no rest yet, `a = a` is
		// a genuine self-assignment, and only the tail differs. The rule was right and the fixture
		// was asserting something that is not true.
		{"a rest against a spread with a trailing element", "export function run() { [...b] = [...b, 1]; }\n"},

		// Everything up to the last spread can be overwritten by it, so nothing before it can be
		// claimed as a self-assignment.
		{"a property before a spread", "export function run() { ({ a } = { a, ...b }); }\n"},

		// After a spread on the value side every later index is unknown, so no position past it can
		// be claimed. The spread must come first for this to measure the truncation: with a match
		// ahead of it the loop reports that match and returns before the truncation matters.
		{"an element after an array spread", "export function run() { [b, a] = [...c, a]; }\n"},

		// A method or getter is a definition, not a value read back, so it is never the same
		// reference as the target. Needs a matching name on both sides to reach the check at all.
		{"a method with the same name", "export function run() { ({ a } = { a() {} }); }\n"},
		{"a getter with the same name", "export function run() { ({ a } = { get a() { return 1; } }); }\n"},

		// A subscript that is an expression rather than a name is never compared, so it is silent
		// even when written identically on both sides.
		{"an expression subscript", "export function run() { o[i + 1] = o[i + 1]; }\n"},
		{"an identifier subscript against a dotted key", "export function run() { o[b] = o.b; }\n"},

		// A private name is not a string key, so it never meets the quoted spelling of its text.
		{"a private name against its quoted text", "export class C { #p = 1; run() { this['#p'] = this.#p; } }\n"},
		{"a quoted text against a private name", "export class C { #p = 1; run() { this.#p = this['#p']; } }\n"},

		// A destructuring default is a fallback, taken only when the element is missing.
		{"a default in an object pattern", "export function run() { ({ a = 1 } = { a }); }\n"},
		{"a default in an array pattern", "export function run() { [a = 1] = [a]; }\n"},
		{"a default that names its own target", "export function run() { [a = a] = [b]; }\n"},
		{"a default under a property", "export function run() { ({ b: a = a } = o); }\n"},

		// A call is evaluated twice, so the two sides are not the same reference even though they
		// are written the same way.
		{"a call on both sides", "export function run() { f().b = f().b; }\n"},

		// Compound arithmetic writes a different value, so it is not this rule's business.
		{"a plus-equals", "export function run() { a += a; }\n"},
		{"a minus-equals", "export function run() { a -= a; }\n"},

		{"an empty object pattern", "export function run() { ({} = {}); }\n"},
		{"a declaration rather than an assignment", "export function run() { let z = z; }\n"},
		{"no assignment at all", "export const value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoSelfAssign, selfAssignFile,
				selfAssignDeclarations+testCase.sourceText))
		})
	}
}

// decodedNoSelfAssign runs the registered decoder over raw option JSON, the way the config layer does
func decodedNoSelfAssign(t *testing.T, raw string) any {
	t.Helper()
	var input []byte
	if raw != "" {
		input = []byte(raw)
	}
	decoded, err := DecodeNoSelfAssignOptions(input)
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestNoSelfAssignPropsOption holds the one option to upstream's default and its off switch.
//
// The default is true, so the interesting inputs are the ones that must keep it true: no options,
// a bare severity, and an empty object, which a zero-value decode would quietly turn off.
func TestNoSelfAssignPropsOption(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		options    string
		sourceText string
		wantCount  int
	}{
		{"a member under a bare severity", "", "export function run() { o.b = o.b; }\n", 1},
		{"a member under an empty object", "{}", "export function run() { o.b = o.b; }\n", 1},
		{"a member with props on", `{"props": true}`, "export function run() { o.b = o.b; }\n", 1},
		{"a member with props off", `{"props": false}`, "export function run() { o.b = o.b; }\n", 0},
		{"a subscript with props off", `{"props": false}`, "export function run() { o[i] = o[i]; }\n", 0},
		{"a member inside a pattern with props off", `{"props": false}`, "export function run() { [o.b] = [o.b]; }\n", 0},

		// Props only gates the member half: a name assigned to itself reports either way.
		{"a name with props off", `{"props": false}`, "export function run() { a = a; }\n", 1},
		{"a pattern with props off", `{"props": false}`, "export function run() { [a, o.b] = [a, o.b]; }\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			expected := make([]string, testCase.wantCount)
			for index := range expected {
				expected[index] = "selfAssignment"
			}
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoSelfAssign, selfAssignFile,
				selfAssignDeclarations+testCase.sourceText, decodedNoSelfAssign(t, testCase.options)), expected...)
		})
	}
}

// TestNoSelfAssignRefusesUnknownOptions holds the decoder to upstream's `additionalProperties: false`.
func TestNoSelfAssignRefusesUnknownOptions(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{`{"prop": false}`, `{"Props": false}`, `{"props": "no"}`, `false`} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeNoSelfAssignOptions([]byte(raw)); err == nil {
				t.Errorf("decoded %s, expected it refused", raw)
			}
		})
	}
}
