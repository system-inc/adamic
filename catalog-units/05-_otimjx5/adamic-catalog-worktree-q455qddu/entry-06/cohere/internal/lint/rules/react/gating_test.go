package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// A `.tsx` name, because every case here writes JSX in a component body exactly as React's own
// fixtures do. Parsing those as plain TypeScript reads `<div>` as a type assertion and changes what
// the tree contains, which would not alter this rule's verdict but would stop the fixtures from
// being the source they claim to be.
const gatingFile = "gating.tsx"

// The bodies React ships under `fixtures/compiler/gating/`, transcribed with the `// @dynamicGating`
// pragma line dropped. That pragma configures React's compiler rather than being part of the source
// under judgment, and this port's option defaults the other way, so keeping it would be a comment
// asserting a precondition the rule does not read.
//
// Every one of these was downloaded from `facebook/react` rather than typed, and all four were
// byte-compared against those downloaded files by a script before being pasted here. See the note
// on TestGatingTranscriptionsStillHoldTheirDirectives for why that comparison cannot live in this
// file and what is guarded here instead.
const (
	// error.dynamic-gating-invalid-identifier.js
	gatingUpstreamInvalidIdentifier = `function Foo() {
  'use memo if(true)';
  return <div>hello world</div>;
}
`
	// dynamic-gating-invalid-multiple.js
	gatingUpstreamMultiple = `function Foo() {
  'use memo if(getTrue)';
  'use memo if(getFalse)';
  return <div>hello world</div>;
}
`
	// dynamic-gating-enabled.js, and dynamic-gating-disabled.js is the same shape with `getFalse`.
	// Both are clean upstream: the directive is well formed and the names differ only in what the
	// flag returns at runtime, which no linter can see.
	gatingUpstreamEnabled = `function Foo() {
  'use memo if(getTrue)';
  return <div>hello world</div>;
}
`
	gatingUpstreamDisabled = `function Foo() {
  'use memo if(getFalse)';
  return <div>hello world</div>;
}
`
)

// The upstream corpus, split by verdict.
//
// React ships one error-named gating fixture, `error.dynamic-gating-invalid-identifier`, whose
// `## Error` block states the message, the location, and the caret span. The duplicate case is not
// error-named because upstream runs it under `@panicThreshold:"none"`, which downgrades the same
// diagnostic into the `## Logs` block rather than suppressing it. Its logged payload carries
// `"category":"Gating"` and the exact message, so it is a reporting case here: the panic threshold
// decides whether compilation aborts, not whether the finding exists.
//
// That is the same shape as the port brief's warning about a case decided above the rule, arriving
// in the opposite direction. `rule_testing.Run` has no panic-threshold layer, so the honest recording
// is the finding, with the reason at the line.
func TestGatingFiresOnUpstreamCorpus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			// `## Error`: "Dynamic gating directive is not a valid JavaScript identifier",
			// at 4:2 in upstream's own numbering, which is line 2 here once the pragma is dropped.
			name:       "invalid identifier, a boolean literal",
			sourceText: gatingUpstreamInvalidIdentifier,
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// `## Logs`: "Multiple dynamic gating directives found", one finding for the body.
			name:       "two well formed directives in one body",
			sourceText: gatingUpstreamMultiple,
			wantIds:    []string{"multipleGatingDirectives"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, Gating, gatingFile, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// Upstream's clean gating fixtures. Both write a well formed directive, which is the case the rule
// must not report, and they are the only inputs in the corpus that prove the grammar is being
// matched rather than every `use memo` prefix being flagged.
func TestGatingStaysSilentOnUpstreamCorpus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
	}{
		{"a well formed directive", gatingUpstreamEnabled},
		{"a well formed directive naming a different flag", gatingUpstreamDisabled},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, Gating, gatingFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not ship, each written from reading the grammar rather than by analogy with a
// neighbouring rule. The imported corpus writes exactly four inputs and covers none of the
// boundaries below, so without these the rule's discriminations are untested.
func TestGatingFiresOnCasesUpstreamDoesNotShip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			// `[^)]*` permits an empty condition, so the string matches the grammar and is then
			// judged. The empty string is not an identifier, so it reports. This is the case the
			// regex's `*` makes reachable and the one a hand-written parser most easily drops.
			name:       "an empty condition still matches the grammar",
			sourceText: "function Foo() {\n  'use memo if()';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// A reserved word is an identifier name and is still rejected, which is the whole
			// reason the reserved set is written out rather than delegated to `IsIdentifierName`.
			name:       "a reserved word is not a valid condition",
			sourceText: "function Foo() {\n  'use memo if(class)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// `await` and `enum` come from Babel's `isReservedWord` with `inModule` true rather
			// than from its keyword list, so they are the two members most easily lost.
			name:       "await is reserved through the module path",
			sourceText: "function Foo() {\n  'use memo if(await)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			name:       "enum is reserved through the module path",
			sourceText: "function Foo() {\n  'use memo if(enum)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// A condition that is an expression rather than a name. `isIdentifierName` rejects the
			// dot, which is the ordinary malformed case a user would actually write.
			name:       "a member expression is not a name",
			sourceText: "function Foo() {\n  'use memo if(flags.on)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// A digit cannot open an identifier, which is the start-versus-part distinction. Only
			// a condition whose first rune is illegal and whose later runes are legal can see it.
			name:       "a leading digit fails the start rune",
			sourceText: "function Foo() {\n  'use memo if(1flag)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// Every invalid directive reports, because upstream collects across the body before
			// returning. Two findings rather than one, and neither is the duplicate message.
			name:       "two malformed directives both report, and neither is the duplicate",
			sourceText: "function Foo() {\n  'use memo if(true)';\n  'use memo if(false)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective", "invalidGatingDirective"},
		},
		{
			// Invalid outranks duplicate. Upstream returns its invalid errors before the count
			// check runs, so a body holding one good directive and one bad one reports only the
			// bad one even though two directives are present.
			name:       "one valid and one invalid reports only the invalid",
			sourceText: "function Foo() {\n  'use memo if(getTrue)';\n  'use memo if(true)';\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// Three well formed directives are still one finding, not two. The duplicate message
			// is per body rather than per extra directive.
			name:       "three directives report once",
			sourceText: "function Foo() {\n  'use memo if(a)';\n  'use memo if(b)';\n  'use memo if(c)';\n  return 1;\n}\n",
			wantIds:    []string{"multipleGatingDirectives"},
		},
		{
			// Double quotes are a directive too. Probed on our parser: both quote styles produce a
			// KindStringLiteral expression statement, and upstream reads Babel's cooked value,
			// which is quote-blind.
			name:       "double quotes are a directive",
			sourceText: "function Foo() {\n  \"use memo if(true)\";\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// A source file has a directive prologue of its own, and both authorities read
			// directives off whatever body they are handed rather than off functions only.
			name:       "a top level directive is judged",
			sourceText: "'use memo if(true)';\nexport const x = 1;\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// An arrow function with a block body carries a prologue. Upstream's own
			// `arrow-function-expr-gating-test` fixture writes this shape, though not with a
			// malformed condition.
			name:       "an arrow function block body is judged",
			sourceText: "const Foo = () => {\n  'use memo if(true)';\n  return 1;\n};\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// A class method body is a block with a prologue like any other.
			name:       "a class method body is judged",
			sourceText: "class C {\n  m() {\n    'use memo if(true)';\n    return 1;\n  }\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
		{
			// The one deliberate widening in this rule, recorded as reporting rather than hidden.
			// A bare nested block is not a function body, so Babel would never collect a directive
			// from it and upstream is silent here. This port walks every block, so it reports.
			//
			// Kept rather than guarded because the input is not one anybody writes: a string
			// statement opening a bare block, spelled as a gating directive with a malformed
			// condition. Narrowing to function bodies would mean re-deriving which of our block
			// parents count as one, and getting that wrong loses real findings, while this
			// direction costs a finding on source that does not occur.
			name:       "a bare nested block is walked too, which upstream would not do",
			sourceText: "function Foo() {\n  {\n    'use memo if(true)';\n  }\n  return 1;\n}\n",
			wantIds:    []string{"invalidGatingDirective"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, Gating, gatingFile, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// The silent side of the boundaries. Every case here is a string that either fails the grammar or
// is not a directive at all, and each one would report under a plausible wrong implementation.
func TestGatingStaysSilentOnCasesUpstreamDoesNotShip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
	}{
		{
			// The plain opt-in directive, which is a different feature entirely. A port matching
			// on a `use memo` prefix rather than the full `use memo if(` opening reports it.
			name:       "the plain opt in directive is not a gating directive",
			sourceText: "function Foo() {\n  'use memo';\n  return 1;\n}\n",
		},
		{
			name:       "the opt out directive is not a gating directive",
			sourceText: "function Foo() {\n  'use no memo';\n  return 1;\n}\n",
		},
		{
			// A space before the paren fails the literal prefix. Upstream's production has no
			// whitespace tolerance anywhere, and a port normalizing spaces would accept this.
			name:       "a space before the paren fails the prefix",
			sourceText: "function Foo() {\n  'use memo if (true)';\n  return 1;\n}\n",
		},
		{
			// Trailing text fails the `$` anchor, so the string is not a gating directive and its
			// condition is never judged, even though the condition itself is malformed.
			name:       "trailing text fails the end anchor",
			sourceText: "function Foo() {\n  'use memo if(true) ';\n  return 1;\n}\n",
		},
		{
			// A close paren inside the condition fails `[^)]*`. This is the case that separates a
			// correct suffix-strip from a naive one: stripping the last paren leaves `a)b`, which
			// a port without the containment check would judge as a malformed condition and
			// report. Upstream does not match the string at all, so it is silent.
			name:       "a close paren inside the condition fails the character class",
			sourceText: "function Foo() {\n  'use memo if(a)b)';\n  return 1;\n}\n",
		},
		{
			// No close paren at all.
			name:       "an unterminated condition is not a directive",
			sourceText: "function Foo() {\n  'use memo if(true';\n  return 1;\n}\n",
		},
		{
			// The prologue boundary, and the reason `ast.IsPrologueDirective` could not be used.
			// That predicate answers true for this statement, measured directly, so a port built
			// on it reports a mid-body string neither authority calls a directive.
			name:       "a string after a statement is not in the prologue",
			sourceText: "function Foo() {\n  let z = 1;\n  'use memo if(true)';\n  return 1;\n}\n",
		},
		{
			// Same boundary with a valid directive ahead of the break, proving the run ends rather
			// than the whole body being scanned. Two directives are present textually and only one
			// is in the prologue, so the duplicate check must not fire either.
			name:       "the prologue ends at the first non string statement",
			sourceText: "function Foo() {\n  'use memo if(getTrue)';\n  let z = 1;\n  'use memo if(getFalse)';\n  return 1;\n}\n",
		},
		{
			// A template literal is not a directive in any implementation. Probed on our parser:
			// it arrives as KindNoSubstitutionTemplateLiteral, a different kind entirely.
			name:       "a template literal is not a directive",
			sourceText: "function Foo() {\n  `use memo if(true)`;\n  return 1;\n}\n",
		},
		{
			// A parenthesized string is a KindParenthesizedExpression statement, which is not a
			// directive to Babel either. Probed rather than assumed, since the rest of this
			// package skips parentheses in several places and the instinct is to do it here too.
			name:       "a parenthesized string is not a directive",
			sourceText: "function Foo() {\n  ('use memo if(true)');\n  return 1;\n}\n",
		},
		{
			// A single well formed directive, which is the ordinary correct usage and the case a
			// rule reporting on the mere presence of a directive would break.
			name:       "one well formed directive is fine",
			sourceText: "function Foo() {\n  'use memo if(myFlag)';\n  return 1;\n}\n",
		},
		{
			// Identifier punctuation, both leading characters and both in the tail.
			name:       "dollar and underscore are identifier runes",
			sourceText: "function Foo() {\n  'use memo if($_flag1)';\n  return 1;\n}\n",
		},
		{
			// Non-ASCII identifiers are valid to both authorities, which use full Unicode tables
			// rather than an ASCII range. A port approximating with `a-zA-Z` reports this.
			name:       "a non ascii identifier is valid",
			sourceText: "function Foo() {\n  'use memo if(café)';\n  return 1;\n}\n",
		},
		{
			// The thirty-five words the shim reserves and Babel does not. If this rule had been
			// built on `scanner.StringToToken` these would all report, and a flag getter named
			// `type` or `async` is entirely ordinary.
			name:       "typescript only keywords are valid conditions",
			sourceText: "function Foo() {\n  'use memo if(type)';\n  return 1;\n}\n",
		},
		{
			name:       "async is a valid condition",
			sourceText: "function Foo() {\n  'use memo if(async)';\n  return 1;\n}\n",
		},
		{
			// Babel keeps `eval` and `arguments` in a `strictBind` list that `isValidIdentifier`
			// never consults, so both are valid conditions. A port reaching for a broader
			// "reserved in strict mode" notion rejects them.
			name:       "eval is strictBind only and stays valid",
			sourceText: "function Foo() {\n  'use memo if(eval)';\n  return 1;\n}\n",
		},
		{
			name:       "arguments is strictBind only and stays valid",
			sourceText: "function Foo() {\n  'use memo if(arguments)';\n  return 1;\n}\n",
		},
		{
			// `undefined` and `NaN` are global objects rather than reserved words. oxc has an
			// `is_global_object` predicate that would reject them and `is_valid_identifier` does
			// not call it, so both authorities accept them.
			name:       "undefined is a global rather than a reserved word",
			sourceText: "function Foo() {\n  'use memo if(undefined)';\n  return 1;\n}\n",
		},
		{
			// Two directives in two different bodies are not duplicates of each other. A rule
			// counting per file rather than per body reports this.
			name:       "one directive each in two functions is not a duplicate",
			sourceText: "function A() {\n  'use memo if(a)';\n  return 1;\n}\nfunction B() {\n  'use memo if(b)';\n  return 1;\n}\n",
		},
		{
			// A nested function's directive does not join its parent's count either.
			name:       "a nested function has its own prologue",
			sourceText: "function A() {\n  'use memo if(a)';\n  function B() {\n    'use memo if(b)';\n    return 1;\n  }\n  return B;\n}\n",
		},
		{
			// An entirely unrelated directive, and an unrelated string statement, are both silent.
			name:       "an unrelated directive is not judged",
			sourceText: "function Foo() {\n  'use strict';\n  return 1;\n}\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, Gating, gatingFile, testCase.sourceText))
		})
	}
}

// The span, asserted against React's own byte offsets rather than against this rule's behavior.
//
// `dynamic-gating-invalid-multiple.expect.md` logs its finding at `index: 105` through `index: 128`
// over upstream's file, and slicing upstream's bytes there gives `'use memo if(getTrue)';`. That is
// the whole directive statement including the semicolon, which is Babel's Directive node and is
// twenty three bytes wide.
//
// This matters because it is the one place the two implementations disagree: oxc reports
// `directive.expression.span`, the string literal alone, which is twenty two bytes and excludes the
// semicolon. Both message ids are identical under either choice, so nothing above this test can see
// which one shipped.
func TestGatingReportsTheWholeDirectiveIncludingTheSemicolon(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			name:       "the invalid identifier case",
			sourceText: gatingUpstreamInvalidIdentifier,
			want:       "'use memo if(true)';",
		},
		{
			name:       "the duplicate case points at the first directive",
			sourceText: gatingUpstreamMultiple,
			want:       "'use memo if(getTrue)';",
		},
		{
			name:       "a double quoted directive spans its own quotes",
			sourceText: "function Foo() {\n  \"use memo if(true)\";\n  return 1;\n}\n",
			want:       "\"use memo if(true)\";",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, Gating, gatingFile, testCase.sourceText)
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

// Each invalid directive points at itself rather than all findings landing on the first one, which
// a loop reporting the wrong variable would produce with the right count and the right ids.
func TestGatingPointsEachInvalidFindingAtItsOwnDirective(t *testing.T) {
	t.Parallel()

	const sourceText = "function Foo() {\n  'use memo if(true)';\n  'use memo if(false)';\n  return 1;\n}\n"

	result := rule_testing.Run(t, Gating, gatingFile, sourceText)
	rule_testing.ExpectFindings(t, result, "invalidGatingDirective", "invalidGatingDirective")

	want := []string{"'use memo if(true)';", "'use memo if(false)';"}
	for index, diagnostic := range result.Diagnostics {
		reported := sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != want[index] {
			t.Fatalf("finding %d reported %q, want %q", index, reported, want[index])
		}
	}
}

// The message text, asserted by equality against literals typed here rather than against the rule's
// own message constants. Comparing to the constants is equality, it looks correct, and both sides
// move together under mutation, so it guards nothing.
func TestGatingMessageText(t *testing.T) {
	t.Parallel()

	invalid := rule_testing.Run(t, Gating, gatingFile, gatingUpstreamInvalidIdentifier)
	if len(invalid.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(invalid.Diagnostics))
	}
	if invalid.Diagnostics[0].Message.Id != "invalidGatingDirective" {
		t.Fatalf("message id = %q", invalid.Diagnostics[0].Message.Id)
	}
	const wantInvalid = "The condition inside `use memo if(...)` is not a JavaScript identifier. " +
		"React Compiler imports that exact name from the gating module and calls it to decide " +
		"whether the memoized version of this function runs, so it has to be something " +
		"importable. A literal like `true`, an expression, or a reserved word cannot be " +
		"imported, and the compiler rejects the whole file rather than guessing. Name the flag " +
		"getter and write that identifier instead."
	if invalid.Diagnostics[0].Message.Description != wantInvalid {
		t.Fatalf("invalid description = %q", invalid.Diagnostics[0].Message.Description)
	}

	multiple := rule_testing.Run(t, Gating, gatingFile, gatingUpstreamMultiple)
	if len(multiple.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(multiple.Diagnostics))
	}
	if multiple.Diagnostics[0].Message.Id != "multipleGatingDirectives" {
		t.Fatalf("message id = %q", multiple.Diagnostics[0].Message.Id)
	}
	const wantMultiple = "This function body carries more than one `use memo if(...)` directive. " +
		"Gating resolves to a single imported condition per function, so a second directive is " +
		"not an additional condition, it is an ambiguity the compiler cannot resolve. Keep the " +
		"one directive that describes when this function should be memoized and delete the rest."
	if multiple.Diagnostics[0].Message.Description != wantMultiple {
		t.Fatalf("multiple description = %q", multiple.Diagnostics[0].Message.Description)
	}
}

// The option surface, routed through the rule's own registered decoder rather than by building the
// struct here. That is what puts the default under test: this rule inverts upstream's precondition,
// so the default is the single line most likely to have no upstream counterpart, and handing
// `RunWithOptions` a struct directly would leave the decode path unexercised.
func TestGatingOptionDecodesAndDefaultsToChecking(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[GatingOptions]()

	// A bare severity hands the rule nil options, which is how it is configured in practice. The
	// zero value must be the checking one, or the rule is inert everywhere it is actually used.
	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, Gating, gatingFile, gatingUpstreamInvalidIdentifier, nil),
		"invalidGatingDirective")

	// An empty object is the same default arriving through the decoder rather than past it.
	empty, err := decode(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("decoding an empty object: %v", err)
	}
	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, Gating, gatingFile, gatingUpstreamInvalidIdentifier, empty),
		"invalidGatingDirective")

	// Set explicitly false, which must not be read as unset.
	off, err := decode(json.RawMessage(`{"requireDynamicGatingOption": false}`))
	if err != nil {
		t.Fatalf("decoding an explicit false: %v", err)
	}
	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, Gating, gatingFile, gatingUpstreamInvalidIdentifier, off),
		"invalidGatingDirective")

	// Set true, restoring upstream's silence. This is the only configuration under which React and
	// oxc's own behavior is reproduced exactly, and it must reach the rule through the decoder.
	on, err := decode(json.RawMessage(`{"requireDynamicGatingOption": true}`))
	if err != nil {
		t.Fatalf("decoding an explicit true: %v", err)
	}
	rule_testing.ExpectClean(t,
		rule_testing.RunWithOptions(t, Gating, gatingFile, gatingUpstreamInvalidIdentifier, on))
	rule_testing.ExpectClean(t,
		rule_testing.RunWithOptions(t, Gating, gatingFile, gatingUpstreamMultiple, on))
}

// The transcribed fixture bodies are guarded against escape cooking, which is the failure this
// port could not otherwise see.
//
// A gating directive is a string literal nested inside a Go string literal, so any tool between
// upstream and this file can turn a written escape into the character it denotes while leaving
// source that still compiles and still goes green. The bodies above were byte-compared against the
// upstream files themselves, downloaded from `facebook/react`, by a script rather than by eye, and
// all four matched with only the compiler pragma line and the FIXTURE_ENTRYPOINT export removed.
//
// That comparison cannot live here, because a test in this file could only compare a literal
// against another literal typed in the same file, and both would cook together. What is checkable
// in-process is the property the cooking would destroy: each body must still contain the exact
// directive text, spelled with the quote style upstream used, and the two multi-directive bodies
// must hold two distinct conditions rather than one repeated.
func TestGatingTranscriptionsStillHoldTheirDirectives(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		body     string
		contains []string
	}{
		{
			name:     "error.dynamic-gating-invalid-identifier",
			body:     gatingUpstreamInvalidIdentifier,
			contains: []string{"  'use memo if(true)';\n", "return <div>hello world</div>;"},
		},
		{
			name: "dynamic-gating-invalid-multiple",
			body: gatingUpstreamMultiple,
			contains: []string{
				"  'use memo if(getTrue)';\n",
				"  'use memo if(getFalse)';\n",
			},
		},
		{
			name:     "dynamic-gating-enabled",
			body:     gatingUpstreamEnabled,
			contains: []string{"  'use memo if(getTrue)';\n"},
		},
		{
			name:     "dynamic-gating-disabled",
			body:     gatingUpstreamDisabled,
			contains: []string{"  'use memo if(getFalse)';\n"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			for _, want := range testCase.contains {
				if !strings.Contains(testCase.body, want) {
					t.Fatalf("body no longer contains %q:\n%s", want, testCase.body)
				}
			}
			// Single quotes, which is what upstream writes. A tool normalizing quote style would
			// leave every verdict unchanged and the transcription no longer upstream's.
			if strings.Contains(testCase.body, "\"use memo") {
				t.Fatalf("body was renormalized to double quotes:\n%s", testCase.body)
			}
		})
	}
}
