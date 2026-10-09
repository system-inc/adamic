package typescript

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noArrayDeleteFile is the fixture name every case in this file runs under.
//
// A TypeScript extension is required rather than incidental: the corpus writes `declare const`
// and type annotations in nearly every case, so a `.js` name would make most of these parse
// errors rather than rule inputs.
const noArrayDeleteFile = "noArrayDelete.ts"

// TestNoArrayDeleteStaysSilent carries all nine of tsgolint's valid cases verbatim.
//
// Copied byte for byte from tsgolint's own test file and verified against it mechanically rather
// than by reading, because a fixture I invent encodes the same belief as the port, so it passes
// for exactly the reason the code would be wrong.
//
// The set is more pointed than its size suggests. Three cases separate a PROPERTY access from an
// ELEMENT access on a receiver that really is an array, which is the distinction a port written
// from the message text alone would miss. Three more pin that `any`, `unknown` and `never` are
// silent, and they are silent because none of them is an array rather than through any explicit
// skip. The last one, `delete console.log()`, is not an access at all.
func TestNoArrayDeleteStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a property access on a plain object", "\n      declare const obj: { a: 1; b: 2 };\n      delete obj.a;\n    "},
		{"a computed string key on a plain object", "\n      declare const obj: { a: 1; b: 2 };\n      delete obj['a'];\n    "},
		{"a property access under four array levels", "\n      declare const arr: { a: 1; b: 2 }[][][][];\n      delete arr[0][0][0][0].a;\n    "},
		{"an element access on any", "\n      declare const maybeArray: any;\n      delete maybeArray[0];\n    "},
		{"an element access on unknown", "\n      declare const maybeArray: unknown;\n      delete maybeArray[0];\n    "},
		{"a property access on a constrained generic object", "\n      declare function getObject<T extends { a: 1; b: 2 }>(): T;\n      delete getObject().a;\n    "},
		{"a property access on an object with a generic member", "\n      declare function getObject<T extends number>(): { a: T; b: 2 };\n      delete getObject().a;\n    "},
		{"an element access on never", "\n      declare const test: never;\n      delete test[0];\n    "},
		{"a call expression rather than an access", "\n      delete console.log();\n    "},

		// Beyond upstream's corpus, and each one measured rather than reasoned about.
		//
		// A UNION MIXING an array with a non-array is the case tsgolint's corpus never writes, and
		// its absence is what let a mutant rewriting the union arm from every to some survive the
		// whole imported set. Upstream's only union case is `number[] | string[] | boolean[]`,
		// where every member is an array and both spellings therefore report, so nothing in the
		// imported floor can tell the two apart. Measured through the @typescript-eslint Linter
		// API on a real program: silent. Our rule agrees.
		{"a union mixing an array with a string", "declare const arr: number[] | string;\ndelete arr[0];\n"},
		{"a union mixing an array with an object", "declare const arr: number[] | { a: 1 };\ndelete arr[0];\n"},
		{"a union with no array member at all", "declare const arr: { a: 1 } | { b: 2 };\ndelete arr[\"a\"];\n"},

		// The intersection arm's opposite direction, for symmetry. An intersection reports when ANY
		// member is an array, so one with none must be silent, and the imported corpus only ever
		// writes intersections that DO contain an array.
		{"an intersection with no array member", "declare const arr: { a: 1 } & { b: 2 };\ndelete arr[\"a\"];\n"},

		// A DOTTED property access on a genuine array, which is where the two upstreams actually
		// disagree and where the disagreement is invisible from either rule file alone.
		//
		// tsgolint guards with `ast.IsElementAccessExpression`, and in the TypeScript AST a dotted
		// access is a PropertyAccessExpression rather than an element access, so this declines.
		// @typescript-eslint guards with `argument.type !== MemberExpression`, and in the ESTree
		// AST a dotted access IS a MemberExpression, so it passes the guard, reaches the type test,
		// finds a real array, and reports.
		//
		// Measured rather than read, because the two guards look equivalent until you know which
		// AST each names. On `declare const arr: number[]; delete arr.length;` @typescript-eslint
		// reports and offers `arr.splice(length, 1)`, which references an undeclared identifier and
		// would not compile. tsgolint is silent, our rule is silent, and tsgolint is what oxlint
		// runs, so silence is both the faithful answer and the defensible one.
		//
		// Upstream's own valid corpus already contains dotted accesses, but only on receivers that
		// are NOT arrays, so every one of them is silent for two independent reasons at once and
		// none can isolate this guard.
		{"a dotted property access on a genuine array", "declare const arr: number[];\ndelete arr.length;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoArrayDelete, noArrayDeleteFile, testCase.sourceText))
		})
	}
}

// TestNoArrayDeleteFires carries all twenty two of tsgolint's invalid cases verbatim.
//
// Each reports exactly once, so the count is twenty two findings over twenty two inputs and there
// is no per-input recovery to do.
//
// The composite type cases are the ones worth naming, because they pin an asymmetry the rule file
// makes easy to read as one predicate. A UNION reports only when every member is an array, and
// `number[] | string[] | boolean[]` is here to hold that direction down. An INTERSECTION reports
// when any member is, which is why `number[] & unknown` and `string & Array<number>` both report
// even though neither is wholly an array.
func TestNoArrayDeleteFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a numeric literal index on a number array", "\n        declare const arr: number[];\n        delete arr[0];\n      "},
		{"a variable index", "\n        declare const arr: number[];\n        declare const key: number;\n        delete arr[key];\n      "},
		{"an enum member index", "\n        declare const arr: number[];\n\n        enum Keys {\n          A,\n          B,\n        }\n\n        delete arr[Keys.A];\n      "},
		{"a sequence expression index", "\n        declare const arr: number[];\n        declare function doWork(): void;\n        delete arr[(doWork(), 1)];\n      "},
		{"the Array generic spelling", "\n        declare const arr: Array<number>;\n        delete arr[0];\n      "},
		{"an array literal receiver", "delete [1, 2, 3][0];"},
		{"a conditional expression index on an unknown array", "\n        declare const arr: unknown[];\n        delete arr[Math.random() ? 0 : 1];\n      "},
		{"a union whose every member is an array", "\n        declare const arr: number[] | string[] | boolean[];\n        delete arr[0];\n      "},
		{"an intersection with unknown", "\n        declare const arr: number[] & unknown;\n        delete arr[0];\n      "},
		{"an array of a union element type", "\n        declare const arr: (number | string)[];\n        delete arr[0];\n      "},
		{"a nested property path ending in an array", "\n        declare const obj: { a: { b: { c: number[] } } };\n        delete obj.a.b.c[0];\n      "},
		{"a generic constrained to an array", "\n        declare function getArray<T extends number[]>(): T;\n        delete getArray()[0];\n      "},
		{"an array of a constrained generic", "\n        declare function getArray<T extends number>(): T[];\n        delete getArray()[0];\n      "},
		{"a parameter typed as an array", "\n        function deleteFromArray(a: number[]) {\n          delete a[0];\n        }\n      "},
		{"a parameter typed as an array of a constrained generic", "\n        function deleteFromArray<T extends number>(a: T[]) {\n          delete a[0];\n        }\n      "},
		{"a parameter whose generic is constrained to an array", "\n        function deleteFromArray<T extends number[]>(a: T) {\n          delete a[0];\n        }\n      "},
		{"a tuple receiver", "\n        declare const tuple: [number, string];\n        delete tuple[0];\n      "},
		{"a spread array literal receiver", "\n        declare const a: number[];\n        declare const b: number;\n\n        delete [...a, ...a][b];\n      "},
		{"comments inside and around the expression", "\n        declare const a: number[];\n        declare const b: number;\n\n        // before expression\n        delete /** multi\n        line */ a[((\n        // single-line\n        b /* inline */ /* another-inline */ )\n        ) /* another-one */ ] /* before semicolon */; /* after semicolon */\n        // after expression\n      "},
		{"doubled parentheses around the whole access", "\n        declare const a: number[];\n        declare const b: number;\n\n        delete ((a[((b))]));\n      "},
		{"an arithmetic index with its own parentheses", "\n        declare const a: number[];\n        declare const b: number;\n\n        delete a[(b + 1) * (b + 2)];\n      "},
		{"an intersection of a string and an array", "\n        declare const arr: string & Array<number>;\n        delete arr[0];\n      "},

		// Beyond upstream's corpus, and each one measured through the @typescript-eslint Linter API
		// on a real program before being written here rather than reasoned about from the source.
		//
		// A READONLY array and a READONLY tuple both report. That is worth pinning because the
		// intuitive reading is the opposite: `delete` on a readonly receiver is already a type
		// error, so a reader may expect the rule to defer to the compiler and stay quiet. It does
		// not, in either implementation, and a port that "helpfully" skipped them would be silently
		// narrower than the gate.
		{"a readonly array", "declare const arr: readonly number[];\ndelete arr[0];\n"},
		{"a readonly tuple", "declare const arr: readonly [number, string];\ndelete arr[0];\n"},

		// A union of an array and a TUPLE, which is the every-arm's positive direction: both members
		// satisfy the predicate through its two different halves rather than through the same one.
		{"a union of an array and a tuple", "declare const arr: number[] | [string, number];\ndelete arr[0];\n"},

		// An intersection carrying an array and an object, matching the some-arm against a member
		// that is a plain object rather than the `unknown` and `string` the corpus writes.
		{"an intersection of an array and an object", "declare const arr: number[] & { a: 1 };\ndelete arr[0];\n"},

		// Index EXPRESSION shapes, which the rule never inspects at all. Nothing reads the argument
		// beyond taking its range, so a negative literal and a computed string key both report on an
		// array receiver exactly as a plain numeric literal does. Written because the intuitive
		// reading is that a rule about array indices would look at the index, and it does not.
		{"a negative index on an array", "declare const arr: number[];\ndelete arr[-1];\n"},
		{"a computed string key on an array", "declare const arr: number[];\ndelete arr[\"length\"];\n"},

		// Parentheses around the RECEIVER, distinct from upstream's case parenthesizing the whole
		// access. The rule skips parentheses on the delete operand only, so this reaches the element
		// access through a different route than `delete ((a[b]))` does.
		{"parentheses around the receiver", "declare const arr: number[];\ndelete (arr)[0];\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoArrayDelete, noArrayDeleteFile, testCase.sourceText), "noArrayDelete")
		})
	}
}

// TestNoArrayDeleteSuggestions applies each suggested repair and compares the resulting source
// against the output tsgolint records for it.
//
// This is the assertion the harness does not supply and the one that can catch what no message id
// can. ExpectFixedSource applies FIXES, and the repair here is a SUGGESTION, which the engine
// never applies unattended. So the applier below is hand-rolled. It is worth the twenty lines
// because a finding carrying a repair that deletes the wrong range satisfies its message id
// perfectly, and every id fixture above stays green over it.
//
// A suggestion rather than a fix is the correct classification and both upstreams agree on it:
// `delete arr[0]` leaves the length alone and puts a hole at index zero, while `arr.splice(0, 1)`
// shortens the array and reindexes everything after it. The two are not the same program, so a
// human has to choose. Applying it unattended would silently change behavior.
//
// All twenty two of upstream's exact outputs are here rather than a sample, because the three
// edits this repair makes are each independently wrong-able. THE LEADING SPACE is the cheapest
// signal in the whole file: every expected output begins with one, because only the `delete`
// keyword token is removed and the space after it is not, so a repair that widened its removal by
// a single byte would fail all twenty two while satisfying every id assertion.
func TestNoArrayDeleteSuggestions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantOutput string
	}{
		{"a numeric literal index on a number array", "\n        declare const arr: number[];\n        delete arr[0];\n      ", "\n        declare const arr: number[];\n         arr.splice(0, 1);\n      "},
		{"a variable index", "\n        declare const arr: number[];\n        declare const key: number;\n        delete arr[key];\n      ", "\n        declare const arr: number[];\n        declare const key: number;\n         arr.splice(key, 1);\n      "},
		{"an enum member index", "\n        declare const arr: number[];\n\n        enum Keys {\n          A,\n          B,\n        }\n\n        delete arr[Keys.A];\n      ", "\n        declare const arr: number[];\n\n        enum Keys {\n          A,\n          B,\n        }\n\n         arr.splice(Keys.A, 1);\n      "},
		{"a sequence expression index keeps its own parentheses", "\n        declare const arr: number[];\n        declare function doWork(): void;\n        delete arr[(doWork(), 1)];\n      ", "\n        declare const arr: number[];\n        declare function doWork(): void;\n         arr.splice((doWork(), 1), 1);\n      "},
		{"the Array generic spelling", "\n        declare const arr: Array<number>;\n        delete arr[0];\n      ", "\n        declare const arr: Array<number>;\n         arr.splice(0, 1);\n      "},
		{"an array literal receiver", "delete [1, 2, 3][0];", " [1, 2, 3].splice(0, 1);"},
		{"a conditional expression index on an unknown array", "\n        declare const arr: unknown[];\n        delete arr[Math.random() ? 0 : 1];\n      ", "\n        declare const arr: unknown[];\n         arr.splice(Math.random() ? 0 : 1, 1);\n      "},
		{"a union whose every member is an array", "\n        declare const arr: number[] | string[] | boolean[];\n        delete arr[0];\n      ", "\n        declare const arr: number[] | string[] | boolean[];\n         arr.splice(0, 1);\n      "},
		{"an intersection with unknown", "\n        declare const arr: number[] & unknown;\n        delete arr[0];\n      ", "\n        declare const arr: number[] & unknown;\n         arr.splice(0, 1);\n      "},
		{"an array of a union element type", "\n        declare const arr: (number | string)[];\n        delete arr[0];\n      ", "\n        declare const arr: (number | string)[];\n         arr.splice(0, 1);\n      "},
		{"a nested property path ending in an array", "\n        declare const obj: { a: { b: { c: number[] } } };\n        delete obj.a.b.c[0];\n      ", "\n        declare const obj: { a: { b: { c: number[] } } };\n         obj.a.b.c.splice(0, 1);\n      "},
		{"a generic constrained to an array", "\n        declare function getArray<T extends number[]>(): T;\n        delete getArray()[0];\n      ", "\n        declare function getArray<T extends number[]>(): T;\n         getArray().splice(0, 1);\n      "},
		{"an array of a constrained generic", "\n        declare function getArray<T extends number>(): T[];\n        delete getArray()[0];\n      ", "\n        declare function getArray<T extends number>(): T[];\n         getArray().splice(0, 1);\n      "},
		{"a parameter typed as an array", "\n        function deleteFromArray(a: number[]) {\n          delete a[0];\n        }\n      ", "\n        function deleteFromArray(a: number[]) {\n           a.splice(0, 1);\n        }\n      "},
		{"a parameter typed as an array of a constrained generic", "\n        function deleteFromArray<T extends number>(a: T[]) {\n          delete a[0];\n        }\n      ", "\n        function deleteFromArray<T extends number>(a: T[]) {\n           a.splice(0, 1);\n        }\n      "},
		{"a parameter whose generic is constrained to an array", "\n        function deleteFromArray<T extends number[]>(a: T) {\n          delete a[0];\n        }\n      ", "\n        function deleteFromArray<T extends number[]>(a: T) {\n           a.splice(0, 1);\n        }\n      "},
		{"a tuple receiver", "\n        declare const tuple: [number, string];\n        delete tuple[0];\n      ", "\n        declare const tuple: [number, string];\n         tuple.splice(0, 1);\n      "},
		{"a spread array literal receiver", "\n        declare const a: number[];\n        declare const b: number;\n\n        delete [...a, ...a][b];\n      ", "\n        declare const a: number[];\n        declare const b: number;\n\n         [...a, ...a].splice(b, 1);\n      "},
		{"comments stay exactly where they were", "\n        declare const a: number[];\n        declare const b: number;\n\n        // before expression\n        delete /** multi\n        line */ a[((\n        // single-line\n        b /* inline */ /* another-inline */ )\n        ) /* another-one */ ] /* before semicolon */; /* after semicolon */\n        // after expression\n      ", "\n        declare const a: number[];\n        declare const b: number;\n\n        // before expression\n         /** multi\n        line */ a.splice(((\n        // single-line\n        b /* inline */ /* another-inline */ )\n        ) /* another-one */ , 1) /* before semicolon */; /* after semicolon */\n        // after expression\n      "},
		{"doubled parentheses around the whole access survive", "\n        declare const a: number[];\n        declare const b: number;\n\n        delete ((a[((b))]));\n      ", "\n        declare const a: number[];\n        declare const b: number;\n\n         ((a.splice(((b)), 1)));\n      "},
		{"an arithmetic index with its own parentheses", "\n        declare const a: number[];\n        declare const b: number;\n\n        delete a[(b + 1) * (b + 2)];\n      ", "\n        declare const a: number[];\n        declare const b: number;\n\n         a.splice((b + 1) * (b + 2), 1);\n      "},
		{"an intersection of a string and an array", "\n        declare const arr: string & Array<number>;\n        delete arr[0];\n      ", "\n        declare const arr: string & Array<number>;\n         arr.splice(0, 1);\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayDelete, noArrayDeleteFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 {
				t.Fatalf("want one suggestion, got %d", len(suggestions))
			}
			if suggestions[0].Message.Id != "useSplice" {
				t.Errorf("suggestion id is %q, want %q", suggestions[0].Message.Id, "useSplice")
			}
			// Three fixes per suggestion, one per edited token. Pinned because the adapter carries
			// a fix slice of any length into one diagnostic, and the edit engine judges each
			// proposal independently, so a repair that lost one of the three would still apply the
			// other two and leave source that parses.
			if len(suggestions[0].Fixes) != 3 {
				t.Errorf("the suggestion carries %d fixes, want three: the delete token, the open bracket, and the close bracket", len(suggestions[0].Fixes))
			}
			// The harness trims the fixture before building the program, so offsets are against
			// the trimmed text rather than against the literal written above, and the expected
			// output has to be trimmed the SAME way rather than the obvious way.
			//
			// TrimSpace on both sides is wrong here and it took a red fixture to see it, then a
			// second red one to see it properly. Upstream writes every case indented inside a raw
			// string, so trimming the input at both ends is right and matches the harness. But the
			// repair's whole signature is the LEADING SPACE it leaves where the `delete` keyword
			// was, and `delete [1, 2, 3][0];` is the one case whose keyword sits at offset zero,
			// so trimming its expected output destroys the single byte the assertion exists to
			// check and the test then demands the opposite of upstream while looking reasonable.
			//
			// Trimming only newlines off the expected output is equally wrong in the other
			// direction: every other case carries eight spaces of indentation that the input's
			// TrimSpace removed, so the two sides disagree on all of them.
			//
			// What is correct is to remove from the expected output exactly the leading whitespace
			// the harness removed from the input, byte for byte, and nothing more. That keeps the
			// one meaningful leading space and drops the shared indentation, because the space the
			// repair leaves behind is not part of the input's prefix at all.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			got := applyNoArrayDeleteSuggestion(t, source, suggestions[0])
			leadingTrivia := testCase.sourceText[:len(testCase.sourceText)-len(strings.TrimLeft(testCase.sourceText, " \t\n"))]
			want := strings.TrimRight(strings.TrimPrefix(testCase.wantOutput, leadingTrivia), " \t\n") + "\n"
			if got != want {
				t.Errorf("applying the suggestion produced\n%q\nwant\n%q", got, want)
			}
		})
	}
}

// applyNoArrayDeleteSuggestion rewrites source with one suggestion's fixes, latest range first.
//
// Back to front so an earlier edit cannot move the offsets of a later one. Unlike the neighbouring
// rule this one genuinely needs the ordering: every suggestion carries three fixes over three
// distinct ranges in the same expression, so applying them front to back would shift the two that
// follow and produce garbage that still compiles as a string.
func applyNoArrayDeleteSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(fixes, func(i, j int) bool { return fixes[i].Range.Pos() > fixes[j].Range.Pos() })
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes", start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

// TestNoArrayDeleteSpans asserts where each finding points, by slicing the source with the
// finding's own range and comparing the text.
//
// ExpectFindings asserts message ids and a count and NOTHING else, so a rule pointing at the wrong
// node passes a complete fixture pair while being wrong. That is not hypothetical for an adapted
// rule: the adapter's node-report forms once passed raw node.Loc, which includes leading trivia, so
// an indented statement reported text beginning at the end of the previous line. It reaches a user
// as a caret on the wrong line and no message id fixture can see it.
//
// The eighteen cases here are every upstream invalid case that records an explicit Line, Column and
// EndColumn. Those three numbers are converted to the text they select, so what is asserted below
// is tsgolint's own recorded span rather than a span I chose. All eighteen select the WHOLE delete
// expression, `delete` keyword included, which is what makes the four cases at column eleven worth
// keeping: they are indented inside a function body, so a trivia-including range would begin on the
// line above and they are the ones that would catch a regression.
//
// The four upstream cases that record no position are covered by the suggestion outputs instead,
// which pin three exact byte ranges each and are strictly stronger than a span assertion.
func TestNoArrayDeleteSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"a numeric literal index", "\n        declare const arr: number[];\n        delete arr[0];\n      ", "delete arr[0]"},
		{"a variable index", "\n        declare const arr: number[];\n        declare const key: number;\n        delete arr[key];\n      ", "delete arr[key]"},
		{"an enum member index", "\n        declare const arr: number[];\n\n        enum Keys {\n          A,\n          B,\n        }\n\n        delete arr[Keys.A];\n      ", "delete arr[Keys.A]"},
		{"a sequence expression index", "\n        declare const arr: number[];\n        declare function doWork(): void;\n        delete arr[(doWork(), 1)];\n      ", "delete arr[(doWork(), 1)]"},
		{"the Array generic spelling", "\n        declare const arr: Array<number>;\n        delete arr[0];\n      ", "delete arr[0]"},
		{"an array literal receiver", "delete [1, 2, 3][0];", "delete [1, 2, 3][0]"},
		{"a conditional expression index", "\n        declare const arr: unknown[];\n        delete arr[Math.random() ? 0 : 1];\n      ", "delete arr[Math.random() ? 0 : 1]"},
		{"a union of arrays", "\n        declare const arr: number[] | string[] | boolean[];\n        delete arr[0];\n      ", "delete arr[0]"},
		{"an intersection with unknown", "\n        declare const arr: number[] & unknown;\n        delete arr[0];\n      ", "delete arr[0]"},
		{"an array of a union element type", "\n        declare const arr: (number | string)[];\n        delete arr[0];\n      ", "delete arr[0]"},
		{"a nested property path", "\n        declare const obj: { a: { b: { c: number[] } } };\n        delete obj.a.b.c[0];\n      ", "delete obj.a.b.c[0]"},
		{"a generic constrained to an array", "\n        declare function getArray<T extends number[]>(): T;\n        delete getArray()[0];\n      ", "delete getArray()[0]"},
		{"an array of a constrained generic", "\n        declare function getArray<T extends number>(): T[];\n        delete getArray()[0];\n      ", "delete getArray()[0]"},
		{"a parameter typed as an array", "\n        function deleteFromArray(a: number[]) {\n          delete a[0];\n        }\n      ", "delete a[0]"},
		{"a parameter of an array of a constrained generic", "\n        function deleteFromArray<T extends number>(a: T[]) {\n          delete a[0];\n        }\n      ", "delete a[0]"},
		{"a parameter whose generic is constrained to an array", "\n        function deleteFromArray<T extends number[]>(a: T) {\n          delete a[0];\n        }\n      ", "delete a[0]"},
		{"a tuple receiver", "\n        declare const tuple: [number, string];\n        delete tuple[0];\n      ", "delete tuple[0]"},
		{"an intersection of a string and an array", "\n        declare const arr: string & Array<number>;\n        delete arr[0];\n      ", "delete arr[0]"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayDelete, noArrayDeleteFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			finding := result.Diagnostics[0]
			if finding.Range.End() > len(source) {
				t.Fatalf("the finding ends at %d over %d bytes", finding.Range.End(), len(source))
			}
			if got := source[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("the finding points at %q, want %q", got, testCase.wantText)
			}
			// Asserted against literals typed here rather than against the rule's own message
			// constants, because comparing a finding to the constant it was built from moves both
			// sides together under mutation and reads as a passing test over a changed message.
			if finding.Message.Id != "noArrayDelete" {
				t.Errorf("the message id is %q, want %q", finding.Message.Id, "noArrayDelete")
			}
			if finding.Message.Description != "Using the `delete` operator with an array expression is unsafe." {
				t.Errorf("the message reads %q", finding.Message.Description)
			}
			if finding.RuleName != "@typescript-eslint/no-array-delete" {
				t.Errorf("the finding reports under %q, and a name that is not the registered one is unsuppressable by the comment its author wrote", finding.RuleName)
			}
			if len(finding.Suggestions) != 1 {
				t.Fatalf("want one suggestion, got %d", len(finding.Suggestions))
			}
			if finding.Suggestions[0].Message.Description != "Use `array.splice()` instead." {
				t.Errorf("the suggestion reads %q", finding.Suggestions[0].Message.Description)
			}
			// A suggestion rather than a fix, asserted because the distinction is permission. The
			// engine applies a fix unattended, and splice reindexes the array where delete does not,
			// so this repair changing arms from suggestion to fix would silently rewrite behavior.
			if len(finding.Fixes) != 0 {
				t.Errorf("the finding carries %d fixes, and this repair must be a suggestion because splice reindexes the array", len(finding.Fixes))
			}
		})
	}
}

// TestNoArrayDeleteRequiresTheTypedHarness asserts the rule declares the checker, so a later revert
// to rule_testing.Run fails loudly instead of going green.
//
// This rule is the silent kind rather than the panicking kind, which is the more dangerous of the
// two. The listener dereferences ctx.TypeChecker, and under rule_testing.Run that field is nil, so a
// fixture set moved to the untyped harness would see every Fires case fail and every StaysSilent
// case pass VACUOUSLY, having proven nothing at all.
//
// The nil guard the standing advice asks for now lives at the top of the listener, because
// absorbing the rule off the adapter made that listener ours to edit. It is unreachable through
// registration, since NeedsTypeChecker is declared; it covers the harness path, where a Context is
// built by hand. This test pins the declaration AND the guard, so losing either one fails loudly
// rather than going vacuously green.
func TestNoArrayDeleteRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoArrayDelete.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}
	if NoArrayDelete.Name != "@typescript-eslint/no-array-delete" {
		t.Errorf("the registered name is %q, and the inventory writes typescript/no-array-delete", NoArrayDelete.Name)
	}

	// A finding the typed harness produces and the untyped one cannot.
	source := "declare const arr: number[];\ndelete arr[0];\n"
	typed := rule_testing.RunTyped(t, NoArrayDelete, noArrayDeleteFile, source)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness found %d findings, want one", len(typed.Diagnostics))
	}

	// The guard the absorption made possible. Driving the listener with a checker-less Context must
	// return rather than dereference nil, and this is the only path that reaches that branch, since
	// registration always supplies a checker.
	listeners := NoArrayDelete.Run(rule.Context{SourceFile: typed.SourceFile}, nil)
	listener, hasListener := listeners[ast.KindDeleteExpression]
	if !hasListener {
		t.Fatal("the rule stopped listening on delete expressions")
	}
	for _, statement := range typed.SourceFile.Statements.Nodes {
		if statement.Kind == ast.KindExpressionStatement {
			listener(statement.AsExpressionStatement().Expression)
		}
	}
}
