package typescript

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// consistentTypeAssertionsFile names the fixture file.
//
// A .ts file, because upstream writes angle-bracket assertions throughout and those are not
// parseable in .tsx at all. The three cases that genuinely need JSX carry their own file name.
const consistentTypeAssertionsFile = "/repository/source/Assertions.ts"

// consistentTypeAssertionsJsxFile is for the three cases whose subject is a JSX attribute.
const consistentTypeAssertionsJsxFile = "/repository/source/Assertions.tsx"

// consistentTypeAssertionsCaseName numbers a row so a failure names which one.
func consistentTypeAssertionsCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// asTheAssertionsHarnessWroteIt transforms a fixture the way the harness transforms its input.
//
// rule_testing writes each typed fixture as strings.TrimSpace(contents)+"\n", so both a span
// sliced from the Go literal and an expected fix output are offset from what the rule saw.
func asTheAssertionsHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// applyConsistentTypeAssertionsSuggestion rewrites source with one suggestion's fixes.
//
// The harness has no suggestion support: ExpectFixedSource covers fixes, and a rule offering
// suggestions has nothing that applies them. Thirty-nine of upstream's findings carry two
// suggestions each and the corpus records the exact output of both, so asserting only message ids
// would leave the entire suggestion surface unchecked.
//
// Back to front so an earlier edit cannot move a later one's offsets, which is what the fix engine
// does. These suggestions carry two fixes each, so the ordering is load-bearing here rather than
// incidental.
func applyConsistentTypeAssertionsSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(fixes, func(first, second int) bool {
		return fixes[first].Range.Pos() > fixes[second].Range.Pos()
	})
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes",
				start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

// decodeConsistentTypeAssertionsOptionsForTest routes a fixture through the rule's own decoder.
//
// The option surface is a discriminated union upstream: `{assertionStyle: "never"}` admits no other
// key, while the two style values that are not never carry two independent literal-assertion
// settings. Nothing but the decoder can tell a port that reads that shape from one that guesses.
func decodeConsistentTypeAssertionsOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeConsistentTypeAssertionsOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// runConsistentTypeAssertions runs one case, with or without options.
func runConsistentTypeAssertions(t *testing.T, fileName string, sourceText string, options string) rule_testing.Result {
	t.Helper()
	if options == "" {
		return rule_testing.RunTyped(t, ConsistentTypeAssertions, fileName, sourceText)
	}
	return rule_testing.RunTypedWithOptions(t, ConsistentTypeAssertions, fileName, sourceText,
		decodeConsistentTypeAssertionsOptionsForTest(t, options))
}

// TestConsistentTypeAssertionsStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All eighty-five of upstream's passing inputs, extracted from the clone's test file by parsing it
// with the TypeScript compiler rather than by reading it, then byte verified. Every one was replayed
// through the installed 8.x build, one program per case, and all eighty-five reported nothing.
//
// The options travel with each row rather than being set once, because this rule's corpus is mostly
// an option matrix: the same source is valid under one assertion style and invalid under another,
// and the two literal-assertion settings cross that independently.
func TestConsistentTypeAssertionsStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		fileName   string
	}{
		{sourceText: "const x = new Generic<int>() as Foo;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = b as A;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = [1] as readonly number[];", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = 'string' as a | b;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = !'string' as A;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = (a as A) + b;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = new Generic<string>() as Foo;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = new (Generic<string> as Foo)();", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = new (Generic<string> as Foo)('string');", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = () => ({ bar: 5 }) as Foo;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = () => bar as Foo;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = bar<string>`${'baz'}` as Foo;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = { key: 'value' } as const;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <Foo>new Generic<int>();", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <A>b;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <readonly number[]>[1];", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <a | b>'string';", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <A>!'string';", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <A>a + b;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <Foo>new Generic<string>();", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = new (<Foo>Generic<string>)();", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = new (<Foo>Generic<string>)('string');", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = () => <Foo>{ bar: 5 };", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = () => <Foo>bar;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <Foo>bar<string>`${'baz'}`;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <const>{ key: 'value' };", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = {} as Foo<int>;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = {} as a | b;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = ({} as A) + b;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "new print({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nfunction foo() {\n  throw { bar: 5 } as Foo;\n}\n      ", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "function b(x = {} as Foo.Bar) {}", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "function c(x = {} as Foo) {}", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.call({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print`${{ bar: 5 } as Foo}`;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <Foo<int>>{};", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <a | b>{};", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <A>{} + b;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "new print(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nfunction foo() {\n  throw <Foo>{ bar: 5 };\n}\n      ", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.call(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print`${<Foo>{ bar: 5 }}`;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "new print({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nfunction foo() {\n  throw { bar: 5 } as Foo;\n}\n      ", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "function b(x = {} as Foo.Bar) {}", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "function c(x = {} as Foo) {}", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.call({ bar: 5 } as Foo);", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print`${{ bar: 5 } as Foo}`;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "new print(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nfunction foo() {\n  throw <Foo>{ bar: 5 };\n}\n      ", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.call(<Foo>{ bar: 5 });", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print`${<Foo>{ bar: 5 }}`;", options: "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = [] as string[];", options: "{\"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = ['a'] as Array<string>;", options: "{\"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <string[]>[];", options: "{\"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <Array<string>>[];", options: "{\"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print([5] as Foo);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nfunction foo() {\n  throw [5] as Foo;\n}\n      ", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "function b(x = [5] as Foo.Bar) {}", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.([5] as Foo);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.call([5] as Foo);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print`${[5] as Foo}`;", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "new Print([5] as Foo);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const bar = <Foo style={[5] as Bar} />;", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}", fileName: consistentTypeAssertionsJsxFile},
		{sourceText: "print(<Foo>[5]);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nfunction foo() {\n  throw <Foo>[5];\n}\n      ", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "function b(x = <Foo.Bar>[5]) {}", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.(<Foo>[5]);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print?.call(<Foo>[5]);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "print`${<Foo>[5]}`;", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "new Print(<Foo>[5]);", options: "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = <const>[1];", options: "{\"assertionStyle\": \"never\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const x = [1] as const;", options: "{\"assertionStyle\": \"never\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "const bar = <Foo style={{ bar: 5 } as Bar} />;", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}", fileName: consistentTypeAssertionsJsxFile},
		{sourceText: "123;", options: "", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nconst x = { key: 'value' } as any;\n      ", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}", fileName: consistentTypeAssertionsFile},
		{sourceText: "\nconst x = { key: 'value' } as unknown;\n      ", options: "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}", fileName: consistentTypeAssertionsFile},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeAssertionsCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runConsistentTypeAssertions(t, testCase.fileName,
				testCase.sourceText, testCase.options))
		})
	}
}

// TestConsistentTypeAssertionsFiresOnUpstreamFailCases is the imported failing corpus, verbatim,
// with the span, the rendered message, the applied fix and every suggestion asserted.
//
// This rule ships both a fix and suggestions, and the corpus specifies all of it: twenty-one cases
// record the exact text the fixer must write, and thirty-nine findings record two suggestions each
// with their own output. A message id assertion sees none of that, and a wrong repair applied
// unattended is the failure this project has paid for twice tonight.
//
// The fix is the angle-bracket to as conversion and it reasons about operator PRECEDENCE to decide
// parenthesization, so a case like `<string>(foo + bar)` is where a naive text swap breaks. Those
// rows are in here rather than left to inference.
func TestConsistentTypeAssertionsFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText      string
		options         string
		fileName        string
		wantIds         []string
		wantSpans       []string
		wantMessages    []string
		wantFixed       string
		wantSuggestions [][]string
	}{
		{
			sourceText:   "const x = new Generic<int>() as Foo;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"new Generic<int>() as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = b as A;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"b as A"},
			wantMessages: []string{"Use '<A>' instead of 'as A'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = [1] as readonly number[];",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"[1] as readonly number[]"},
			wantMessages: []string{"Use '<readonly number[]>' instead of 'as readonly number[]'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = 'string' as a | b;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"'string' as a | b"},
			wantMessages: []string{"Use '<a | b>' instead of 'as a | b'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = !'string' as A;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"!'string' as A"},
			wantMessages: []string{"Use '<A>' instead of 'as A'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = (a as A) + b;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"a as A"},
			wantMessages: []string{"Use '<A>' instead of 'as A'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new Generic<string>() as Foo;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"new Generic<string>() as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new (Generic<string> as Foo)();",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"Generic<string> as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new (Generic<string> as Foo)('string');",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"Generic<string> as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = () => ({ bar: 5 }) as Foo;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"({ bar: 5 }) as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = () => bar as Foo;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"bar as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = bar<string>`${'baz'}` as Foo;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"bar<string>`${'baz'}` as Foo"},
			wantMessages: []string{"Use '<Foo>' instead of 'as Foo'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = { key: 'value' } as const;",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"{ key: 'value' } as const"},
			wantMessages: []string{"Use '<const>' instead of 'as const'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <Foo>new Generic<int>();",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>new Generic<int>()"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = new Generic<int>() as Foo;",
		},
		{
			sourceText:   "const x = <A>b;",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<A>b"},
			wantMessages: []string{"Use 'as A' instead of '<A>'."},
			wantFixed:    "const x = b as A;",
		},
		{
			sourceText:   "const x = <readonly number[]>[1];",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<readonly number[]>[1]"},
			wantMessages: []string{"Use 'as readonly number[]' instead of '<readonly number[]>'."},
			wantFixed:    "const x = [1] as readonly number[];",
		},
		{
			sourceText:   "const x = <a | b>'string';",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<a | b>'string'"},
			wantMessages: []string{"Use 'as a | b' instead of '<a | b>'."},
			wantFixed:    "const x = 'string' as a | b;",
		},
		{
			sourceText:   "const x = <A>!'string';",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<A>!'string'"},
			wantMessages: []string{"Use 'as A' instead of '<A>'."},
			wantFixed:    "const x = !'string' as A;",
		},
		{
			sourceText:   "const x = <A>a + b;",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<A>a"},
			wantMessages: []string{"Use 'as A' instead of '<A>'."},
			wantFixed:    "const x = (a as A) + b;",
		},
		{
			sourceText:   "const x = <Foo>new Generic<string>();",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>new Generic<string>()"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = new Generic<string>() as Foo;",
		},
		{
			sourceText:   "const x = new (<Foo>Generic<string>)();",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>Generic<string>"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = new ((Generic<string>) as Foo)();",
		},
		{
			sourceText:   "const x = new (<Foo>Generic<string>)('string');",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>Generic<string>"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = new ((Generic<string>) as Foo)('string');",
		},
		{
			sourceText:   "const x = () => <Foo>{ bar: 5 };",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>{ bar: 5 }"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = () => ({ bar: 5 } as Foo);",
		},
		{
			sourceText:   "const x = () => <Foo>bar;",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>bar"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = () => (bar as Foo);",
		},
		{
			sourceText:   "const x = <Foo>bar<string>`${'baz'}`;",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<Foo>bar<string>`${'baz'}`"},
			wantMessages: []string{"Use 'as Foo' instead of '<Foo>'."},
			wantFixed:    "const x = bar<string>`${'baz'}` as Foo;",
		},
		{
			sourceText:   "const x = <const>{ key: 'value' };",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<const>{ key: 'value' }"},
			wantMessages: []string{"Use 'as const' instead of '<const>'."},
			wantFixed:    "const x = { key: 'value' } as const;",
		},
		{
			sourceText:   "const x = new Generic<int>() as Foo;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"new Generic<int>() as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = b as A;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"b as A"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = [1] as readonly number[];",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"[1] as readonly number[]"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = 'string' as a | b;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"'string' as a | b"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = !'string' as A;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"!'string' as A"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = (a as A) + b;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"a as A"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new Generic<string>() as Foo;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"new Generic<string>() as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new (Generic<string> as Foo)();",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"Generic<string> as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new (Generic<string> as Foo)('string');",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"Generic<string> as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = () => ({ bar: 5 }) as Foo;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"({ bar: 5 }) as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = () => bar as Foo;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"bar as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = bar<string>`${'baz'}` as Foo;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"bar<string>`${'baz'}` as Foo"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <Foo>new Generic<int>();",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>new Generic<int>()"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <A>b;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<A>b"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <readonly number[]>[1];",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<readonly number[]>[1]"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <a | b>'string';",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<a | b>'string'"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <A>!'string';",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<A>!'string'"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <A>a + b;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<A>a"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <Foo>new Generic<string>();",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>new Generic<string>()"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new (<Foo>Generic<string>)();",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>Generic<string>"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = new (<Foo>Generic<string>)('string');",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>Generic<string>"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = () => <Foo>{ bar: 5 };",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>{ bar: 5 }"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = () => <Foo>bar;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>bar"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <Foo>bar<string>`${'baz'}`;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<Foo>bar<string>`${'baz'}`"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:      "const x = {} as Foo<int>;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as Foo<int>"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: Foo<int> = {};", "const x = {} satisfies Foo<int>;"}},
		},
		{
			sourceText:      "const x = {} as a | b;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as a | b"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: a | b = {};", "const x = {} satisfies a | b;"}},
		},
		{
			sourceText:      "const x = ({} as A) + b;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as A"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x = ({} satisfies A) + b;"}},
		},
		{
			sourceText:      "const x = <Foo<int>>{};",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo<int>>{}"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: Foo<int> = {};", "const x = {} satisfies Foo<int>;"}},
		},
		{
			sourceText:      "const x = <a | b>{};",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<a | b>{}"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: a | b = {};", "const x = {} satisfies a | b;"}},
		},
		{
			sourceText:      "const x = <A>{} + b;",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<A>{}"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x = {} satisfies A + b;"}},
		},
		{
			sourceText:      "const x = {} as Foo<int>;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as Foo<int>"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: Foo<int> = {};", "const x = {} satisfies Foo<int>;"}},
		},
		{
			sourceText:      "const x = {} as a | b;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as a | b"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: a | b = {};", "const x = {} satisfies a | b;"}},
		},
		{
			sourceText:      "const x = ({} as A) + b;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as A"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x = ({} satisfies A) + b;"}},
		},
		{
			sourceText:      "print({ bar: 5 } as Foo);",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{ bar: 5 } as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "new print({ bar: 5 } as Foo);",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{ bar: 5 } as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"new print({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "\nfunction foo() {\n  throw { bar: 5 } as Foo;\n}\n      ",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{ bar: 5 } as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"\nfunction foo() {\n  throw { bar: 5 } satisfies Foo;\n}\n      "}},
		},
		{
			sourceText:      "function b(x = {} as Foo.Bar) {}",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as Foo.Bar"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"function b(x = {} satisfies Foo.Bar) {}"}},
		},
		{
			sourceText:      "function c(x = {} as Foo) {}",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{} as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"function c(x = {} satisfies Foo) {}"}},
		},
		{
			sourceText:      "print?.({ bar: 5 } as Foo);",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{ bar: 5 } as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print?.({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "print?.call({ bar: 5 } as Foo);",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{ bar: 5 } as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print?.call({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "print`${{ bar: 5 } as Foo}`;",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"{ bar: 5 } as Foo"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print`${{ bar: 5 } satisfies Foo}`;"}},
		},
		{
			sourceText:      "const x = <Foo<int>>{};",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo<int>>{}"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: Foo<int> = {};", "const x = {} satisfies Foo<int>;"}},
		},
		{
			sourceText:      "const x = <a | b>{};",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<a | b>{}"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: a | b = {};", "const x = {} satisfies a | b;"}},
		},
		{
			sourceText:      "const x = <A>{} + b;",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<A>{}"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x = {} satisfies A + b;"}},
		},
		{
			sourceText:      "print(<Foo>{ bar: 5 });",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo>{ bar: 5 }"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "new print(<Foo>{ bar: 5 });",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo>{ bar: 5 }"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"new print({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "\nfunction foo() {\n  throw <Foo>{ bar: 5 };\n}\n      ",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo>{ bar: 5 }"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"\nfunction foo() {\n  throw { bar: 5 } satisfies Foo;\n}\n      "}},
		},
		{
			sourceText:      "print?.(<Foo>{ bar: 5 });",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo>{ bar: 5 }"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print?.({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "print?.call(<Foo>{ bar: 5 });",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo>{ bar: 5 }"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print?.call({ bar: 5 } satisfies Foo);"}},
		},
		{
			sourceText:      "print`${<Foo>{ bar: 5 }}`;",
			options:         "{\"assertionStyle\": \"angle-bracket\", \"objectLiteralTypeAssertions\": \"never\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSpans:       []string{"<Foo>{ bar: 5 }"},
			wantMessages:    []string{"Always prefer const x: T = { ... }."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print`${{ bar: 5 } satisfies Foo}`;"}},
		},
		{
			sourceText:   "const foo = <Foo style={{ bar: 5 } as Bar} />;",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsJsxFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"{ bar: 5 } as Bar"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const a = <any>(b, c);",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>(b, c)"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "const a = (b, c) as any;",
		},
		{
			sourceText:   "const f = <any>(() => {});",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>(() => {})"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "const f = (() => {}) as any;",
		},
		{
			sourceText:   "const f = <any>function () {};",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>function () {}"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "const f = function () {} as any;",
		},
		{
			sourceText:   "const f = <any>(async () => {});",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>(async () => {})"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "const f = (async () => {}) as any;",
		},
		{
			sourceText:   "\nfunction* g() {\n  const y = <any>(yield a);\n}\n      ",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>(yield a)"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "\nfunction* g() {\n  const y = (yield a) as any;\n}\n      ",
		},
		{
			sourceText:   "\ndeclare let x: number, y: number;\nconst bs = <any>(x <<= y);\n      ",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>(x <<= y)"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "\ndeclare let x: number, y: number;\nconst bs = (x <<= y) as any;\n      ",
		},
		{
			sourceText:   "const ternary = <any>(true ? x : y);",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<any>(true ? x : y)"},
			wantMessages: []string{"Use 'as any' instead of '<any>'."},
			wantFixed:    "const ternary = (true ? x : y) as any;",
		},
		{
			sourceText:   "const x = [] as string[];",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"[] as string[]"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <string[]>[];",
			options:      "{\"assertionStyle\": \"never\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"never"},
			wantSpans:    []string{"<string[]>[]"},
			wantMessages: []string{"Do not use any type assertions."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = [] as string[];",
			options:      "{\"assertionStyle\": \"angle-bracket\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"angle-bracket"},
			wantSpans:    []string{"[] as string[]"},
			wantMessages: []string{"Use '<string[]>' instead of 'as string[]'."},
			wantFixed:    "",
		},
		{
			sourceText:   "const x = <string[]>[];",
			options:      "{\"assertionStyle\": \"as\"}",
			fileName:     consistentTypeAssertionsFile,
			wantIds:      []string{"as"},
			wantSpans:    []string{"<string[]>[]"},
			wantMessages: []string{"Use 'as string[]' instead of '<string[]>'."},
			wantFixed:    "const x = [] as string[];",
		},
		{
			sourceText:      "const x = [] as string[];",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[] as string[]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: string[] = [];", "const x = [] satisfies string[];"}},
		},
		{
			sourceText:      "const x = <string[]>[];",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"angle-bracket\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"<string[]>[]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const x: string[] = [];", "const x = [] satisfies string[];"}},
		},
		{
			sourceText:      "print([5] as Foo);",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[5] as Foo"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print([5] satisfies Foo);"}},
		},
		{
			sourceText:      "new print([5] as Foo);",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[5] as Foo"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"new print([5] satisfies Foo);"}},
		},
		{
			sourceText:      "function b(x = [5] as Foo.Bar) {}",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[5] as Foo.Bar"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"function b(x = [5] satisfies Foo.Bar) {}"}},
		},
		{
			sourceText:      "\nfunction foo() {\n  throw [5] as Foo;\n}\n      ",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[5] as Foo"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"\nfunction foo() {\n  throw [5] satisfies Foo;\n}\n      "}},
		},
		{
			sourceText:      "print`${[5] as Foo}`;",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[5] as Foo"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print`${[5] satisfies Foo}`;"}},
		},
		{
			sourceText:      "const foo = () => [5] as Foo;",
			options:         "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"as\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"[5] as Foo"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const foo = () => [5] satisfies Foo;"}},
		},
		{
			sourceText:      "new print(<Foo>[5]);",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"angle-bracket\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"<Foo>[5]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"new print([5] satisfies Foo);"}},
		},
		{
			sourceText:      "function b(x = <Foo.Bar>[5]) {}",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"angle-bracket\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"<Foo.Bar>[5]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"function b(x = [5] satisfies Foo.Bar) {}"}},
		},
		{
			sourceText:      "\nfunction foo() {\n  throw <Foo>[5];\n}\n      ",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"angle-bracket\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"<Foo>[5]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"\nfunction foo() {\n  throw [5] satisfies Foo;\n}\n      "}},
		},
		{
			sourceText:      "print`${<Foo>[5]}`;",
			options:         "{\"arrayLiteralTypeAssertions\": \"never\", \"assertionStyle\": \"angle-bracket\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"<Foo>[5]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"print`${[5] satisfies Foo}`;"}},
		},
		{
			sourceText:      "const foo = <Foo>[5];",
			options:         "{\"arrayLiteralTypeAssertions\": \"allow-as-parameter\", \"assertionStyle\": \"angle-bracket\"}",
			fileName:        consistentTypeAssertionsFile,
			wantIds:         []string{"unexpectedArrayTypeAssertion"},
			wantSpans:       []string{"<Foo>[5]"},
			wantMessages:    []string{"Always prefer const x: T[] = [ ... ]."},
			wantFixed:       "",
			wantSuggestions: [][]string{{"const foo: Foo = [5];", "const foo = [5] satisfies Foo;"}},
		},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeAssertionsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := runConsistentTypeAssertions(t, testCase.fileName,
				testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			written := asTheAssertionsHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				if gotSpan := written[reported.Range.Pos():reported.Range.End()]; gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
				if reported.Message.Description != testCase.wantMessages[findingIndex] {
					t.Errorf("finding %d reads %q, wanted %q", findingIndex,
						reported.Message.Description, testCase.wantMessages[findingIndex])
				}
			}
			if testCase.wantFixed != "" {
				rule_testing.ExpectFixedSource(t, result,
					asTheAssertionsHarnessWroteIt(testCase.wantFixed))
			}
			for findingIndex, wantOutputs := range testCase.wantSuggestions {
				if len(wantOutputs) == 0 {
					continue
				}
				reported := result.Diagnostics[findingIndex]
				if len(reported.Suggestions) != len(wantOutputs) {
					t.Fatalf("finding %d offers %d suggestions, wanted %d",
						findingIndex, len(reported.Suggestions), len(wantOutputs))
				}
				for suggestionIndex, wantOutput := range wantOutputs {
					applied := applyConsistentTypeAssertionsSuggestion(t, written,
						reported.Suggestions[suggestionIndex])
					if applied != asTheAssertionsHarnessWroteIt(wantOutput) {
						t.Errorf("finding %d suggestion %d writes %q, wanted %q",
							findingIndex, suggestionIndex, applied,
							asTheAssertionsHarnessWroteIt(wantOutput))
					}
				}
			}
		})
	}
}

// TestConsistentTypeAssertionsOnPrecedenceShapesUpstreamsCorpusDoesNotWrite covers the fixer
// decisions the imported corpus cannot separate.
//
// This rule's repair is applied unattended and it decides parenthesization by comparing operator
// precedences, so a wrong comparison writes code that means something else. Upstream's 186 cases
// exercise the fixer well but never nest two assertions, so the strict greater-than can be relaxed
// to greater-or-equal and every one of them still passes.
//
// Every verdict and every repair below was measured against the installed 8.x build on that exact
// source, one program per case.
//
// The nested-assertion case is asserted separately, in the test below this one, because two
// assertions on one expression produce two OVERLAPPING fixes and the harness refuses to apply
// overlapping edits rather than guessing which wins. That refusal is right: the fix engine has an
// overlap policy and a fixture reimplementing it would assert a result the real pipeline never
// produces. So that case is pinned on the fix text it proposes instead.
func TestConsistentTypeAssertionsOnPrecedenceShapesUpstreamsCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantFixed  string
	}{
		{
			name:       "angle-bracket-inside-an-as",
			why:        "an angle-bracket assertion whose PARENT is an `as` assertion, so the fix's own precedence equals its parent's. Upstream wraps, which is what makes the comparison a strict greater-than rather than greater-or-equal: at equal precedence associativity decides and the fixer cannot know it is safe. Nothing in upstream's 186 cases puts one assertion directly inside another, so relaxing the comparison survives all of them",
			sourceText: "declare const foo: any;\ndeclare const bar: any;\nconst x = <Foo>foo as Bar;\n",
			wantIds:    []string{"as"},
			wantFixed:  "declare const foo: any;\ndeclare const bar: any;\nconst x = (foo as Foo) as Bar;\n",
		},
		{
			name:       "satisfies-operand",
			why:        "a satisfies expression as the operand, which sits at the same precedence and wraps for the same reason",
			sourceText: "declare const foo: any;\ndeclare const bar: any;\ndeclare const Generic: any;\nconst x = <Foo>(foo satisfies Bar);\n",
			wantIds:    []string{"as"},
			wantFixed:  "declare const foo: any;\ndeclare const bar: any;\ndeclare const Generic: any;\nconst x = (foo satisfies Bar) as Foo;\n",
		},
		{
			name:       "new-without-arguments",
			why:        "a `new` parent written WITHOUT arguments, which binds differently from one with them and is the reason the fixer passes a flag rather than defaulting it",
			sourceText: "declare const foo: any;\ndeclare const bar: any;\ndeclare const Generic: any;\nconst x = new (<Foo>Generic)();\n",
			wantIds:    []string{"as"},
			wantFixed:  "declare const foo: any;\ndeclare const bar: any;\ndeclare const Generic: any;\nconst x = new (Generic as Foo)();\n",
		},
		{
			name:       "new-with-arguments",
			why:        "the same parent WITH an argument, kept beside the row above so the flag is measured in both directions",
			sourceText: "declare const foo: any;\ndeclare const bar: any;\ndeclare const Generic: any;\nconst x = new (<Foo>Generic)('s');\n",
			wantIds:    []string{"as"},
			wantFixed:  "declare const foo: any;\ndeclare const bar: any;\ndeclare const Generic: any;\nconst x = new (Generic as Foo)('s');\n",
		},
	}
	const options = `{"assertionStyle": "as"}`
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConsistentTypeAssertions(t, consistentTypeAssertionsFile,
				testCase.sourceText, options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result,
				asTheAssertionsHarnessWroteIt(testCase.wantFixed))
		})
	}
}

// TestConsistentTypeAssertionsDecoderReadsUpstreamsUnion pins the three defaults that are not zero
// values.
//
// Every one of this rule's defaults is a non-zero value, so a decoder handing back a zero struct
// gives an assertion style matching no branch and the rule reports nothing at all. That is the
// failure mode this whole file would otherwise pass through silently, since every fixture above
// names its options explicitly.
func TestConsistentTypeAssertionsDecoderReadsUpstreamsUnion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want ConsistentTypeAssertionsOptions
	}{
		{raw: `{}`, want: DefaultConsistentTypeAssertionsSettings()},
		{
			raw: `{"assertionStyle": "never"}`,
			want: ConsistentTypeAssertionsOptions{
				AssertionStyle:             ConsistentTypeAssertionsNever,
				ObjectLiteralTypeAssertion: ConsistentTypeAssertionsAllow,
				ArrayLiteralTypeAssertion:  ConsistentTypeAssertionsAllow,
			},
		},
		{
			raw: `{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}`,
			want: ConsistentTypeAssertionsOptions{
				AssertionStyle:             ConsistentTypeAssertionsAngleBracket,
				ObjectLiteralTypeAssertion: ConsistentTypeAssertionsNeverLiteral,
				ArrayLiteralTypeAssertion:  ConsistentTypeAssertionsAllow,
			},
		},
		{
			// A value outside upstream's enum is a configuration error ESLint refuses before the
			// rule runs, so it falls back to the default rather than to a zero value, which would
			// silently turn the rule off.
			raw:  `{"assertionStyle": "bogus"}`,
			want: DefaultConsistentTypeAssertionsSettings(),
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded := decodeConsistentTypeAssertionsOptionsForTest(t, testCase.raw)
			settings, isSettings := decoded.(ConsistentTypeAssertionsOptions)
			if !isSettings {
				t.Fatalf("the decoder returned %T rather than the rule's own options type", decoded)
			}
			if settings != testCase.want {
				t.Errorf("decoded to %+v, wanted %+v", settings, testCase.want)
			}
		})
	}
}

// TestConsistentTypeAssertionsFallsBackWhenHandedNilOptions covers a bare severity configuration.
//
// A rule named as "error" with no object is handed nil, and the fallback is the only thing between
// that and an empty assertion style that matches no branch.
func TestConsistentTypeAssertionsFallsBackWhenHandedNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, ConsistentTypeAssertions,
		consistentTypeAssertionsFile, "declare const foo: any;\nconst x = <Foo>foo;\n", nil)
	rule_testing.ExpectFindings(t, result, "as")
}

// TestConsistentTypeAssertionsWrapsAnOperandOfEqualPrecedence pins the one comparison the imported
// corpus cannot separate, at the layer that decides it.
//
// An assertion whose operand is ANOTHER assertion puts the two at equal precedence, and upstream
// wraps: measured, `<Foo><Bar>foo` fixes to `(foo as Bar) as Foo`. That is what makes the comparison
// a strict greater-than rather than greater-or-equal, and relaxing it survives all 186 imported
// cases because none of them nests two assertions.
//
// It cannot go through ExpectFixedSource: the two findings propose OVERLAPPING edits, and the
// harness refuses those rather than guessing which wins. That refusal is right, and upstream reaches
// its final text by running the fixer repeatedly rather than by resolving the overlap in one pass.
//
// So the assertion is on what each finding PROPOSES, which is the layer this rule decides. The inner
// assertion is wrapped because its parent is the outer one, at equal precedence: that is the
// comparison under test. The outer one is not wrapped, because its own parent is a declarator.
func TestConsistentTypeAssertionsWrapsAnOperandOfEqualPrecedence(t *testing.T) {
	t.Parallel()

	const sourceText = "declare const foo: any;\nconst x = <Foo><Bar>foo;\n"
	result := runConsistentTypeAssertions(t, consistentTypeAssertionsFile, sourceText,
		`{"assertionStyle": "as"}`)
	rule_testing.ExpectFindings(t, result, "as", "as")

	written := asTheAssertionsHarnessWroteIt(sourceText)
	wanted := map[string]string{
		// The outer assertion, whose operand text is the inner assertion as still written.
		"<Foo><Bar>foo": "<Bar>foo as Foo",
		// The inner one, wrapped because its parent is an assertion of equal precedence. Relaxing
		// the comparison to greater-or-equal drops these parentheses, and nothing else notices.
		"<Bar>foo": "(foo as Bar)",
	}
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted two findings, got %d", len(result.Diagnostics))
	}
	for _, diagnostic := range result.Diagnostics {
		span := written[diagnostic.Range.Pos():diagnostic.Range.End()]
		wantText, known := wanted[span]
		if !known {
			t.Fatalf("a finding points at %q, which is neither assertion", span)
		}
		if len(diagnostic.Fixes) != 1 {
			t.Fatalf("the finding at %q proposes %d fixes, wanted one", span, len(diagnostic.Fixes))
		}
		if diagnostic.Fixes[0].Text != wantText {
			t.Errorf("the fix for %q writes %q, wanted %q", span, diagnostic.Fixes[0].Text, wantText)
		}
	}
}

// TestConsistentTypeAssertionsOnSuggestionAndTemplateShapes covers two guards the imported corpus
// tests only on one side.
//
// The annotation suggestion is offered only when the declarator has no annotation of its own, and
// every one of upstream's suggestion-bearing cases has an unannotated declarator, so dropping that
// guard survives all 186 and then offers a repair that writes two types onto one binding.
//
// The parameter-position exemption treats a TAGGED template as a parameter and an untagged one as
// an ordinary expression, and upstream writes only the tagged form.
//
// Measured against the installed 8.x build, one program per case, suggestions included.
func TestConsistentTypeAssertionsOnSuggestionAndTemplateShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		why             string
		sourceText      string
		options         string
		wantIds         []string
		wantSuggestions []string
	}{
		{
			name:            "declarator-with-annotation",
			why:             "a declarator that ALREADY carries an annotation, which is offered only the satisfies suggestion. Adding the annotation one would produce two types on the same binding, and upstream guards on exactly that; every case in its corpus with suggestions has an unannotated declarator, so dropping the guard survives all 186",
			sourceText:      "declare type Foo = {};\nconst x: Foo = {} as Foo;\n",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSuggestions: []string{"replaceObjectTypeAssertionWithSatisfies"},
		},
		{
			name:            "declarator-without-annotation",
			why:             "the unannotated half, offered both suggestions, kept beside the row above so the guard is measured in both directions",
			sourceText:      "declare type Foo = {};\nconst x = {} as Foo;\n",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"never\"}",
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSuggestions: []string{"replaceObjectTypeAssertionWithAnnotation", "replaceObjectTypeAssertionWithSatisfies"},
		},
		{
			name:            "untagged-template",
			why:             "an assertion inside an UNTAGGED template, which reports under allow-as-parameter: an untagged template coerces its interpolation, so an annotation would still be checkable there and the position is not a parameter",
			sourceText:      "declare type Foo = {};\nconst x = `${{} as Foo}`;\n",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			wantIds:         []string{"unexpectedObjectTypeAssertion"},
			wantSuggestions: []string{"replaceObjectTypeAssertionWithSatisfies"},
		},
		{
			name:            "tagged-template",
			why:             "the tagged form, which is exempt. The pair is what pins the tagged requirement, and upstream's corpus writes only one of the two",
			sourceText:      "declare type Foo = {};\ndeclare function tag(s: any, ...v: any[]): string;\nconst x = tag`${{} as Foo}`;\n",
			options:         "{\"assertionStyle\": \"as\", \"objectLiteralTypeAssertions\": \"allow-as-parameter\"}",
			wantIds:         []string{},
			wantSuggestions: []string{},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConsistentTypeAssertions(t, consistentTypeAssertionsFile,
				testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			var offered []string
			for _, diagnostic := range result.Diagnostics {
				for _, suggestion := range diagnostic.Suggestions {
					offered = append(offered, suggestion.Message.Id)
				}
			}
			if len(offered) != len(testCase.wantSuggestions) {
				t.Fatalf("offered %v suggestions, wanted %v (%s)",
					offered, testCase.wantSuggestions, testCase.why)
			}
			for index, wanted := range testCase.wantSuggestions {
				if offered[index] != wanted {
					t.Errorf("suggestion %d is %q, wanted %q (%s)",
						index, offered[index], wanted, testCase.why)
				}
			}
		})
	}
}
