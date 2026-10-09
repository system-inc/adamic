package typescript

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const mixedEnumsFile = "/repository/source/Enums.ts"

// The tables below are typescript-eslint's own corpus for this rule, extracted from
// tests/rules/no-mixed-enums.test.ts by parsing it with the TypeScript compiler API and
// emitting each case through JSON serialization. Nothing was retyped, and every string was
// byte-verified against the literals in that file before this table was written.
//
// Three upstream cases import an ambient declaration from a sibling module. They need a second
// file in the program, so they run through RunTypedFiles in their own tests below rather than
// being weakened into single-file approximations.

func TestNoMixedEnumsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nenum Fruit {}\n    ",
		"\nenum Fruit {\n  Apple,\n}\n    ",
		"\nenum Fruit {\n  Apple = false,\n}\n    ",
		"\nenum Fruit {\n  Apple,\n  Banana,\n}\n    ",
		"\nenum Fruit {\n  Apple = 0,\n  Banana,\n}\n    ",
		"\nenum Fruit {\n  Apple,\n  Banana = 1,\n}\n    ",
		"\nenum Fruit {\n  Apple = 0,\n  Banana = 1,\n}\n    ",
		"\nenum Fruit {\n  Apple,\n  Banana = false,\n}\n    ",
		"\nconst getValue = () => 0;\nenum Fruit {\n  Apple,\n  Banana = getValue(),\n}\n    ",
		"\nconst getValue = () => 0;\nenum Fruit {\n  Apple = getValue(),\n  Banana = getValue(),\n}\n    ",
		"\nconst getValue = () => '';\nenum Fruit {\n  Apple = '',\n  Banana = getValue(),\n}\n    ",
		"\nconst getValue = () => '';\nenum Fruit {\n  Apple = getValue(),\n  Banana = '',\n}\n    ",
		"\nconst getValue = () => '';\nenum Fruit {\n  Apple = getValue(),\n  Banana = getValue(),\n}\n    ",
		"\nenum First {\n  A = 1,\n}\n\nenum Second {\n  A = First.A,\n  B = 2,\n}\n    ",
		"\nenum First {\n  A = '',\n}\n\nenum Second {\n  A = First.A,\n  B = 'b',\n}\n    ",
		"\nenum Foo {\n  A,\n}\nenum Foo {\n  B,\n}\n    ",
		"\nenum Foo {\n  A = 0,\n}\nenum Foo {\n  B,\n}\n    ",
		"\nenum Foo {\n  A,\n}\nenum Foo {\n  B = 1,\n}\n    ",
		"\nenum Foo {\n  A = 0,\n}\nenum Foo {\n  B = 1,\n}\n    ",
		"\nenum Foo {\n  A = 'a',\n}\nenum Foo {\n  B = 'b',\n}\n    ",
		"\ndeclare const Foo: any;\nenum Foo {\n  A,\n}\n    ",
		"\nenum Foo {\n  A = 1,\n}\nenum Foo {\n  B = 2,\n}\n    ",
		"\nenum Foo {\n  A = `A`,\n}\nenum Foo {\n  B = `B`,\n}\n    ",
		"\nenum Foo {\n  A = false, // (TS error)\n}\nenum Foo {\n  B = `B`,\n}\n    ",
		"\nenum Foo {\n  A = 'A',\n}\nenum Foo {\n  B = false, // (TS error)\n}\n    ",
		"\nimport { Enum } from \"module-that-does't-exist\";\n\ndeclare module \"module-that-doesn't-exist\" {\n  enum Enum {\n    StringLike = 'StringLike',\n  }\n}\n    ",
		"\nnamespace Test {\n  export enum Bar {\n    A = 1,\n  }\n}\nnamespace Test {\n  export enum Bar {\n    B = 2,\n  }\n}\n    ",
		"\nnamespace Outer {\n  namespace Test {\n    export enum Bar {\n      A = 1,\n    }\n  }\n}\nnamespace Outer {\n  namespace Test {\n    export enum Bar {\n      B = 'B',\n    }\n  }\n}\n    ",
		"\nnamespace Outer {\n  namespace Test {\n    export enum Bar {\n      A = 1,\n    }\n  }\n}\nnamespace Different {\n  namespace Test {\n    export enum Bar {\n      B = 'B',\n    }\n  }\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(fmt.Sprintf("upstream valid %d", index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, sourceText))
		})
	}
}

func TestNoMixedEnumsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   int
	}{
		{"\nenum Fruit {\n  Apple,\n  Banana = 'banana',\n}\n      ", 1},
		{"\nenum Fruit {\n  Apple,\n  Banana = 'banana',\n  Cherry = 'cherry',\n}\n      ", 1},
		{"\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry = 'cherry',\n}\n      ", 1},
		{"\nenum Fruit {\n  Apple = 0,\n  Banana = 'banana',\n}\n      ", 1},
		{"\nconst getValue = () => 0;\nenum Fruit {\n  Apple = getValue(),\n  Banana = 'banana',\n}\n      ", 1},
		{"\nconst getValue = () => '';\nenum Fruit {\n  Apple,\n  Banana = getValue(),\n}\n      ", 1},
		{"\nconst getValue = () => '';\nenum Fruit {\n  Apple = getValue(),\n  Banana = 0,\n}\n      ", 1},
		{"\nenum First {\n  A = 1,\n}\n\nenum Second {\n  A = First.A,\n  B = 'b',\n}\n      ", 1},
		{"\nenum First {\n  A = 'a',\n}\n\nenum Second {\n  A = First.A,\n  B = 1,\n}\n      ", 1},
		{"\nenum Foo {\n  A,\n}\nenum Foo {\n  B = 'b',\n}\n      ", 1},
		{"\nenum Foo {\n  A = 1,\n}\nenum Foo {\n  B = 'b',\n}\n      ", 1},
		{"\nenum Foo {\n  A = 'a',\n}\nenum Foo {\n  B,\n}\n      ", 1},
		{"\nenum Foo {\n  A = 'a',\n}\nenum Foo {\n  B = 0,\n}\n      ", 1},
		{"\nenum Foo {\n  A,\n}\nenum Foo {\n  B = 'b',\n}\nenum Foo {\n  C = 'c',\n}\n      ", 2},
		{"\nenum Foo {\n  A,\n}\nenum Foo {\n  B = 'b',\n}\nenum Foo {\n  C,\n}\n      ", 1},
		{"\nenum Foo {\n  A,\n}\nenum Foo {\n  B,\n}\nenum Foo {\n  C = 'c',\n}\n      ", 1},
		{"\nenum Foo {\n  A = 1,\n}\nenum Foo {\n  B = 'B',\n}\n      ", 1},
		{"\nnamespace Test {\n  export enum Bar {\n    A = 1,\n  }\n}\nnamespace Test {\n  export enum Bar {\n    B = 'B',\n  }\n}\n      ", 1},
		{"\nnamespace Test {\n  export enum Bar {\n    A,\n  }\n}\nnamespace Test {\n  export enum Bar {\n    B = 'B',\n  }\n}\n      ", 1},
		{"\nnamespace Outer {\n  export namespace Test {\n    export enum Bar {\n      A = 1,\n    }\n  }\n}\nnamespace Outer {\n  export namespace Test {\n    export enum Bar {\n      B = 'B',\n    }\n  }\n}\n      ", 1},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("upstream invalid %d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// mixedEnumsDeclarationSource is upstream's own tests/fixtures/mixed-enums-decl.ts, copied
// verbatim. The three cases below resolve an import against it, and without it in the program
// they would silently become tests of an unresolved import instead of tests of enum merging.
const mixedEnumsDeclarationSource = "export enum Enum {\n  A = 'A',\n  B = 'B',\n}\n"

func TestNoMixedEnumsAcrossFiles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"upstream valid multi 0", "\nimport { Enum } from './mixed-enums-decl';\n\ndeclare module './mixed-enums-decl' {\n  enum Enum {\n    StringLike = 'StringLike',\n  }\n}\n    ", 0},
		{"upstream invalid multi 0", "\nimport { Enum } from './mixed-enums-decl';\n\ndeclare module './mixed-enums-decl' {\n  enum Enum {\n    Numeric = 0,\n  }\n}\n      ", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"/repository/source/Enums.ts":            testCase.sourceText,
				"/repository/source/mixed-enums-decl.ts": mixedEnumsDeclarationSource,
			}
			result := rule_testing.RunTypedFiles(t, NoMixedEnums, files, "/repository/source/Enums.ts")
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsSpans pins WHERE the finding points, which ExpectFindings cannot see.
//
// The expected text is derived from upstream's own line and column numbers, and the source is
// trimmed the same way rule_testing.RunTyped trims it before writing the fixture to disk. Without
// that trim the leading newline every corpus case carries would shift every column by one and
// the failure would read exactly like an off-by-one in the rule.
func TestNoMixedEnumsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantText   string
	}{
		{"\nenum Fruit {\n  Apple,\n  Banana = 'banana',\n}\n      ", "'banana'"},
		{"\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry = 'cherry',\n}\n      ", "'cherry'"},
		{"\nconst getValue = () => '';\nenum Fruit {\n  Apple,\n  Banana = getValue(),\n}\n      ", "getValue()"},
		{"\nconst getValue = () => '';\nenum Fruit {\n  Apple = getValue(),\n  Banana = 0,\n}\n      ", "0"},
		{"\nenum First {\n  A = 1,\n}\n\nenum Second {\n  A = First.A,\n  B = 'b',\n}\n      ", "'b'"},
		{"\nenum First {\n  A = 'a',\n}\n\nenum Second {\n  A = First.A,\n  B = 1,\n}\n      ", "1"},
		{"\nenum Foo {\n  A = 'a',\n}\nenum Foo {\n  B,\n}\n      ", "B"},
		{"\nenum Foo {\n  A,\n}\nenum Foo {\n  B,\n}\nenum Foo {\n  C = 'c',\n}\n      ", "'c'"},
		{"\nenum Foo {\n  A = 1,\n}\nenum Foo {\n  B = 'B',\n}\n      ", "'B'"},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("span %d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			// Slice the source the HARNESS wrote, not the Go literal.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			finding := result.Diagnostics[0]
			got := onDisk[finding.Range.Pos():finding.Range.End()]
			if got != testCase.wantText {
				t.Fatalf("finding points at %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// TestNoMixedEnumsMessage asserts the reported id and description against literals typed here
// rather than against the rule's own constant, which would move with any mutation to it.
func TestNoMixedEnumsMessage(t *testing.T) {
	t.Parallel()

	message := buildNoMixedEnumsMessage()
	if message.Id != "mixed" {
		t.Fatalf("message id is %q, want %q", message.Id, "mixed")
	}
	if !strings.Contains(message.Description, "both number members and string members") {
		t.Fatalf("description does not say what is wrong: %q", message.Description)
	}
}

// TestNoMixedEnumsRequiresTheTypedHarness pins that this rule declares the checker and guards
// on it. A typed rule run through the plain harness gets a nil checker; without the guard the
// shim returns nil rather than panicking, so the failure would be SILENCE, and every
// StaysSilent case above would keep passing vacuously if the declaration were ever reverted.
func TestNoMixedEnumsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoMixedEnums.NeedsTypeChecker {
		t.Fatal("rule must declare NeedsTypeChecker")
	}
	mixed := "enum E {\n  A = 1,\n  B = 'b',\n}\n"
	typed := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, mixed)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("typed harness: want 1 finding, got %d", len(typed.Diagnostics))
	}
	untyped := rule_testing.Run(t, NoMixedEnums, mixedEnumsFile, mixed)
	if len(untyped.Diagnostics) != 0 {
		t.Fatalf("untyped harness: want 0 findings from the nil-checker guard, got %d", len(untyped.Diagnostics))
	}
}

// TestNoMixedEnumsParentheses covers a case upstream's corpus does not write at all, and which
// only exists because our parser and ESTree disagree about whether a parenthesis is a node.
//
// ESTree strips parentheses, so upstream's initializer switch is handed the inner literal. Ours
// produces a KindParenthesizedExpression, so without the skip in noMixedEnumsClassify every
// parenthesized initializer would fall through to the checker arm, where a boolean classifies as
// Number instead of Unknown and the enum would report instead of being abandoned.
//
// Each row was measured against the installed @typescript-eslint rule, and the bare forms are
// carried alongside as controls so a row going silent for an unrelated reason is visible.
func TestNoMixedEnumsParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		// The parenthesized number still reports, and its SPAN excludes the parentheses. Verified
		// against the parser directly: ESTree reports the range of the inner literal.
		{"parenthesized number after string", "enum E {\n  A = 'a',\n  B = (1),\n}\n", 1},
		{"bare number after string (control)", "enum E {\n  A = 'a',\n  B = 1,\n}\n", 1},

		// The row that settles the whole question. If parentheses routed to the checker arm, a
		// parenthesized boolean would classify as Number and this would REPORT. Upstream is
		// silent, which is only possible if it sees the boolean literal through the wrapper.
		{"parenthesized boolean after string", "enum E {\n  A = 'a',\n  B = (false),\n}\n", 0},
		{"bare boolean after string (control)", "enum E {\n  A = 'a',\n  B = false,\n}\n", 0},

		// And the control proving the pair without the boolean does disagree, so the two silent
		// rows above are silent because of the Unknown abort rather than because nothing is wrong.
		{"the same pair without the boolean (control)", "enum E {\n  A = 'a',\n  B = 1,\n}\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsParenthesizedSpanExcludesTheWrapper pins the span rather than the count, because
// a finding anchored on the KindParenthesizedExpression would satisfy the count assertion above
// while pointing at text upstream never highlights.
func TestNoMixedEnumsParenthesizedSpanExcludesTheWrapper(t *testing.T) {
	t.Parallel()

	sourceText := "enum E {\n  A = 'a',\n  B = (1),\n}\n"
	result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	onDisk := strings.TrimSpace(sourceText) + "\n"
	finding := result.Diagnostics[0]
	got := onDisk[finding.Range.Pos():finding.Range.End()]
	if got != "1" {
		t.Fatalf("finding points at %q, want %q with the parentheses excluded", got, "1")
	}
}

// TestNoMixedEnumsUnknownAborts pins that an Unknown member abandons the WHOLE declaration rather
// than being skipped over. The middle row is the one that costs a finding, and it is the reason
// this is a separate test: reading the upstream source shows a `return` inside a loop and not what
// that return is worth.
func TestNoMixedEnumsUnknownAborts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"disagreeing pair reports", "enum E {\n  A = 1,\n  C = 'c',\n}\n", 1},
		{"a boolean between them silences it", "enum E {\n  A = 1,\n  B = false,\n  C = 'c',\n}\n", 0},
		{"a boolean AFTER the disagreement is too late", "enum E {\n  A = 1,\n  C = 'c',\n  B = false,\n}\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsAdoptsNumberIntoAnUnsetDesiredType covers upstream's `desiredType ??= currentType`
// line, which is only reachable when the FIRST member takes the checker arm and no merge case set a
// desired type. Without the adoption the scan reaches the string with nothing to compare against and
// goes silent. Measured reporting upstream.
func TestNoMixedEnumsAdoptsNumberIntoAnUnsetDesiredType(t *testing.T) {
	t.Parallel()

	sourceText := "declare const f: () => any;\nenum E {\n  A = f(),\n  B = 1,\n  C = 'c',\n}\n"
	result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, sourceText)
	rule_testing.ExpectFindings(t, result, "mixed")
}

// TestNoMixedEnumsMemberWithoutInitializerIsTheAnchor pins where the finding lands when the
// offending member has no initializer to point at. Upstream reports on the member itself, spanning
// only the name, and no imported fixture covers it because every invalid case in the corpus has an
// initializer on the reported member.
func TestNoMixedEnumsMemberWithoutInitializerIsTheAnchor(t *testing.T) {
	t.Parallel()

	sourceText := "enum E {\n  A = 'a',\n  B,\n}\n"
	result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	onDisk := strings.TrimSpace(sourceText) + "\n"
	finding := result.Diagnostics[0]
	got := onDisk[finding.Range.Pos():finding.Range.End()]
	if got != "B" {
		t.Fatalf("finding points at %q, want the member name %q", got, "B")
	}
}

// TestNoMixedEnumsMergeOrdering pins that a third declaration is judged against the FIRST rather
// than against the one immediately before it. That is the observable consequence of upstream
// breaking out of its scope scan on the earliest match, and it is not something a reader would
// predict from the rule's description.
func TestNoMixedEnumsMergeOrdering(t *testing.T) {
	t.Parallel()

	sourceText := "enum E {\n  A = 1,\n}\nenum E {\n  B = 'b',\n}\nenum E {\n  C = 2,\n}\n"
	result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, sourceText)
	// Only the middle declaration reports. The third agrees with the FIRST, so if it were judged
	// against the second there would be two findings here.
	rule_testing.ExpectFindings(t, result, "mixed")
}

// TestNoMixedEnumsNamespaceMerging covers the exported/non-exported split. An exported enum inside a
// namespace merges across blocks and is judged against its partner; a non-exported one does not
// merge and is judged alone. The second row is the control: without it the first row's finding
// could equally be explained by the two enums simply being in the same file.
func TestNoMixedEnumsNamespaceMerging(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{
			"exported enums merge across namespace blocks",
			"namespace N {\n  export enum E {\n    A = 1,\n  }\n}\nnamespace N {\n  export enum E {\n    B = 'b',\n  }\n}\n",
			1,
		},
		{
			"non-exported enums do not merge",
			"namespace N {\n  enum E {\n    A = 1,\n  }\n}\nnamespace N {\n  enum E {\n    B = 'b',\n  }\n}\n",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsEmptyDeclarationIsNotAReference pins that an empty `enum E {}` does not become the
// declaration a later one is judged against. Without the skip the later pair would be compared to a
// declaration with no members, which has no kind at all.
func TestNoMixedEnumsEmptyDeclarationIsNotAReference(t *testing.T) {
	t.Parallel()

	sourceText := "enum E {}\nenum E {\n  A = 1,\n  B = 'b',\n}\n"
	result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, sourceText)
	rule_testing.ExpectFindings(t, result, "mixed")
}

// TestNoMixedEnumsConstAndAmbientEnums pins that the rule reaches const and declared enums, which the
// corpus never writes. Both were measured reporting upstream.
func TestNoMixedEnumsConstAndAmbientEnums(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"const enum", "const enum E {\n  A = 1,\n  B = 'b',\n}\n"},
		{"declare enum", "declare enum E {\n  A = 1,\n  B = 'b',\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "mixed")
		})
	}
}

// TestNoMixedEnumsTemplateLiteralsAreStrings pins the template arm, which every imported case leaves
// invisible: the corpus writes templates only inside enums whose members already agree, so a port
// classifying a template as Number passes all 51 upstream cases. Found by a surviving mutant, and
// each row below was measured against the installed rule before it was written.
//
// The substitution row matters most. Upstream returns String from the node kind alone, so a template
// interpolating a NUMBER is still a string member. Asking the checker would be more careful and
// would disagree with the rule being ported.
func TestNoMixedEnumsTemplateLiteralsAreStrings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"number then template", "enum E {\n  A = 1,\n  B = `b`,\n}\n", 1},
		{"number then plain string (control)", "enum E {\n  A = 1,\n  B = 'b',\n}\n", 1},
		{"template then number", "enum E {\n  A = `a`,\n  B = 1,\n}\n", 1},
		{"template with a numeric substitution is still a string", "declare const n: number;\nenum E {\n  A = 1,\n  B = `${n}`,\n}\n", 1},
		{"two templates agree", "enum E {\n  A = `a`,\n  B = `b`,\n}\n", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsCheckerArm pins the fallback that classifies anything the literal switch does not
// name. The two clean rows are controls: without them a rule that classified every call as String
// would still pass the reporting row.
func TestNoMixedEnumsCheckerArm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"call returning string after number", "declare const f: () => string;\nenum E {\n  A = 1,\n  B = f(),\n}\n", 1},
		{"call returning number after number (control)", "declare const g: () => number;\nenum E {\n  A = 1,\n  B = g(),\n}\n", 0},
		{"call returning string after string (control)", "declare const f: () => string;\nenum E {\n  A = 'a',\n  B = f(),\n}\n", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsNullAborts pins that a `null` initializer is the same third verdict a boolean is,
// abandoning the declaration rather than classifying as either kind. Upstream reaches it through
// `typeof null === 'object'` falling into its Literal switch's default arm. The second row is the
// control proving the pair does disagree once the null is removed.
func TestNoMixedEnumsNullAborts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"null between disagreeing members silences it", "enum E {\n  A = 1,\n  B = null,\n  C = 'c',\n}\n", 0},
		{"the same pair without the null (control)", "enum E {\n  A = 1,\n  C = 'c',\n}\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoMixedEnumsUnknownFirstMemberSilencesEverything pins the inputs that reach for an UNSET
// desired type. Every one of them is clean, and that is the measurement behind the equivalence
// verdict recorded on the adoption line in the rule: there is no input where the scan continues
// with no desired type, because every route to that state runs through an Unknown first member,
// which aborts on the first iteration.
//
// U2 is the control. Without it the four clean rows read as "this shape is just never reported",
// which is the wrong conclusion: the same pair without the leading boolean does report.
func TestNoMixedEnumsUnknownFirstMemberSilencesEverything(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"boolean first, then number, then string", "enum E {\n  A = false,\n  B = 1,\n  C = 'c',\n}\n", 0},
		{"the same without the leading boolean (control)", "enum E {\n  B = 1,\n  C = 'c',\n}\n", 1},
		{"boolean first, then string, then number", "enum E {\n  A = false,\n  B = 'b',\n  C = 1,\n}\n", 0},
		{"null first, then number, then string", "enum E {\n  A = null,\n  B = 1,\n  C = 'c',\n}\n", 0},
		{"merged sibling whose first member is a boolean", "enum E {\n  Z = false,\n}\nenum E {\n  A = 1,\n  B = 'b',\n}\n", 0},
		{"merged sibling with a boolean, then a string-first pair", "enum E {\n  Z = false,\n}\nenum E {\n  A = 'a',\n  B = 1,\n}\n", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMixedEnums, mixedEnumsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "mixed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}
