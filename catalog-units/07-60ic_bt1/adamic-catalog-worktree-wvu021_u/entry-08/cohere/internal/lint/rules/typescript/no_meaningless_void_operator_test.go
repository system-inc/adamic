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

const noMeaninglessVoidOperatorFile = "/repository/source/Discarding.ts"

func noMeaninglessVoidOperatorCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// noMeaninglessVoidOperatorFinding is one expected finding with every layer it can be wrong at.
//
// wantRepair carries the branch, because this rule offers the SAME edit three ways: as a fix
// (removing the operator cannot change behavior), as a suggestion (checkNever, where the call does
// not return), and not at all (a non-call void that is not a whole statement, where removing it
// changes the expression's value). A row asserts which one arrived as much as what it writes.
type noMeaninglessVoidOperatorFinding struct {
	wantId                    string
	wantSpan                  string
	wantMessage               string
	wantRepair                string
	wantSuggestion            string
	wantSuggestionDescription string
}

// noMeaninglessVoidOperatorRow is one corpus row, replayed through the installed rule.
type noMeaninglessVoidOperatorRow struct {
	// origin is "upstream-valid" or "upstream-invalid" for a row from typescript-eslint 8.71.0's own
	// test file, and "extra" for one added here and measured the same way.
	origin          string
	sourceText      string
	optionsJson     string
	wantComposedFix string
	wantFindings    []noMeaninglessVoidOperatorFinding
}

// noMeaninglessVoidOperatorOptionsFor routes a case's options through the rule's own decoder.
func noMeaninglessVoidOperatorOptionsFor(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return DefaultNoMeaninglessVoidOperatorSettings()
	}
	decoded, err := DecodeNoMeaninglessVoidOperatorOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// applyNoMeaninglessVoidOperatorSuggestion rewrites source with one suggestion's fixes.
//
// The harness applies fixes and has no suggestion support, so the checkNever branch's repair would
// otherwise go entirely unasserted. That branch is exactly the one where a mistake matters most,
// since it is the half upstream refuses to apply unattended.
func applyNoMeaninglessVoidOperatorSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(fixes, func(first, second int) bool { return fixes[first].Range.Pos() > fixes[second].Range.Pos() })
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes", start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

// The corpus is typescript-eslint 8.71.0's own, extracted mechanically, plus rows added here.
//
// Upstream's test file at the v8.71.0 tag was loaded with its RuleTester stubbed so every row came out
// as data, then every row was replayed against the INSTALLED 8.71.0 over a real program (strict, lib
// ES2022) to record what it reports: the message id, the span, the rendered message, whether the
// repair is a fix, a suggestion or nothing, and the file the fixes compose to. The replay reproduced
// all 43 of upstream's own assertions before anything was taken from it, which is what says the
// harness reads 8.71 rather than whatever happened to be installed.
//
// The "extra" rows are the edges upstream's file does not pin and this port decides: how the literal
// `0` is recognized (`0x0` and `0.0` are it, `0n` and `-0` are not), compound and logical assignment,
// an instantiation expression (not unwrapped, so a non-call), `new` and a tagged template (non-calls),
// a parenthesized statement (still a statement, so the fix applies), a nested comma sequence, a
// thenable union, an optional call, a conditional of calls, and the type branch's rendered `{{type}}`
// on undefined and on a union.
//
// The harness writes each fixture as `strings.TrimSpace(source)+"\n"`, so the spans and the composed
// rewrites are against that trimmed text.
func noMeaninglessVoidOperatorSilentRows() []noMeaninglessVoidOperatorRow {
	return []noMeaninglessVoidOperatorRow{
		{origin: "upstream-valid", sourceText: "\n(() => {})();\n\nfunction foo() {}\nfoo(); // nothing to discard\n\nfunction bar(): number {\n  return 2;\n}\nvoid bar();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare function fail(): never;\nvoid fail();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare function getValue(): string;\nvoid getValue();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const box: { getValue(): string };\nvoid box.getValue();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const box: { value: string };\nvoid box.value.toUpperCase();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const box: { method(): string } | undefined;\nvoid box?.method();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare function getValue(): string | void;\nvoid getValue();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare function getValue(): string | undefined;\nvoid getValue();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const box: { method?: () => string };\nvoid box.method?.();\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare function getValue(): string;\nvoid (getValue() as string);\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "void 0;", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const promise: Promise<number>;\nvoid promise;\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const thenable: { then(onFulfilled: () => void): void };\nvoid thenable;\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\nvoid new Promise<void>(resolve => {\n  resolve();\n});\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare function fn(): void;\ndeclare function getValue(): string;\nvoid (fn(), getValue());\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare let x: number;\nvoid (x = 1);\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare let x: number;\n() => void (x = 1);\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare let x: number;\nvoid (x += 1);\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare let x: number;\ndeclare let y: number;\nvoid (x = y = 1);\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare let x: string;\ndeclare function getValue(): string;\nvoid (x = getValue());\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare const obj: { prop: number };\nvoid (obj.prop = 1);\n    ", optionsJson: ""},
		{origin: "upstream-valid", sourceText: "\ndeclare let x: number;\ndeclare let y: number;\nvoid ((x = 1), (y = 2));\n    ", optionsJson: ""},
		{origin: "extra", sourceText: "void 0x0;", optionsJson: ""},
		{origin: "extra", sourceText: "void 0.0;", optionsJson: ""},
		{origin: "extra", sourceText: "void (0);", optionsJson: ""},
		{origin: "extra", sourceText: "declare let x: number | undefined;\nvoid (x ??= 1);", optionsJson: ""},
		{origin: "extra", sourceText: "declare let a: number;\ndeclare const b: number[];\nvoid ([a] = b);", optionsJson: ""},
		{origin: "extra", sourceText: "void (async () => {})();", optionsJson: ""},
		{origin: "extra", sourceText: "declare const p: Promise<void> | undefined;\nvoid p;", optionsJson: ""},
		{origin: "extra", sourceText: "declare function fail(): never;\nvoid fail();", optionsJson: ""},
	}
}

func noMeaninglessVoidOperatorFiringRows() []noMeaninglessVoidOperatorRow {
	return []noMeaninglessVoidOperatorRow{
		{
			origin:          "upstream-invalid",
			sourceText:      "void (() => {})();",
			optionsJson:     "",
			wantComposedFix: "(() => {})();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void (() => {})()", wantMessage: "void operator shouldn't be used on void; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\nfunction foo() {}\nvoid foo();\n      ",
			optionsJson:     "",
			wantComposedFix: "function foo() {}\nfoo();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void foo()", wantMessage: "void operator shouldn't be used on void; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string };\nvoid box;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\nbox;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void box", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string };\nvoid box.value;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\nbox.value;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void box.value", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string } | undefined;\nvoid box?.value;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string } | undefined;\nbox?.value;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void box?.value", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string };\nvoid (<string>box.value);\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\n(<string>box.value);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (<string>box.value)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string };\nvoid (box.value as string);\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\n(box.value as string);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (box.value as string)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string };\nvoid (box.value satisfies string);\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\n(box.value satisfies string);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (box.value satisfies string)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: string };\nvoid box.value!;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\nbox.value!;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void box.value!", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const wrapper: { box?: { value: string } };\nvoid wrapper?.box?.value!;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const wrapper: { box?: { value: string } };\nwrapper?.box?.value!;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void wrapper?.box?.value!", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare function fn(): void;\ndeclare const box: { value: string };\nvoid (fn(), box.value);\n      ",
			optionsJson:     "",
			wantComposedFix: "declare function fn(): void;\ndeclare const box: { value: string };\n(fn(), box.value);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (fn(), box.value)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare function fn(): void;\ndeclare const box: { value: string };\nvoid (fn(), box.value)!;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare function fn(): void;\ndeclare const box: { value: string };\n(fn(), box.value)!;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (fn(), box.value)!", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\nfunction bar(x: never) {\n  void x;\n}\n      ",
			optionsJson:     "",
			wantComposedFix: "function bar(x: never) {\n  x;\n}\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare const box: { value: never };\nvoid box.value;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: never };\nbox.value;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void box.value", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:      "upstream-invalid",
			sourceText:  "\ndeclare function fail(): never;\nvoid fail();\n      ",
			optionsJson: "{\"checkNever\": true}",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void fail()", wantMessage: "void operator shouldn't be used on never; it should convey that a return value is being ignored", wantRepair: "Suggestion", wantSuggestion: "declare function fail(): never;\nfail();\n", wantSuggestionDescription: "Remove 'void'"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "void 1;",
			optionsJson:     "",
			wantComposedFix: "1;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void 1", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "void '0';",
			optionsJson:     "",
			wantComposedFix: "'0';\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void '0'", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:      "upstream-invalid",
			sourceText:  "\ndeclare const value: string;\nconst result = void value;\n      ",
			optionsJson: "",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void value", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "None"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare let x: number;\nvoid x;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\nx;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare let x: number;\nvoid ++x;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\n++x;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void ++x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "upstream-invalid",
			sourceText:      "\ndeclare let x: number;\nvoid x++;\n      ",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\nx++;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x++", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "void 0n;",
			optionsJson:     "",
			wantComposedFix: "0n;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void 0n", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "void -0;",
			optionsJson:     "",
			wantComposedFix: "-0;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void -0", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "void null;",
			optionsJson:     "",
			wantComposedFix: "null;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void null", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "void undefined;",
			optionsJson:     "",
			wantComposedFix: "undefined;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void undefined", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare function make<T>(): T;\nvoid make<string>;",
			optionsJson:     "",
			wantComposedFix: "declare function make<T>(): T;\nmake<string>;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void make<string>", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "class Foo {}\nvoid new Foo();",
			optionsJson:     "",
			wantComposedFix: "class Foo {}\nnew Foo();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void new Foo()", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare function tag(strings: TemplateStringsArray): string;\nvoid tag`x`;",
			optionsJson:     "",
			wantComposedFix: "declare function tag(strings: TemplateStringsArray): string;\ntag`x`;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void tag`x`", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare let x: number;\n(void x);",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\n(x);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare function fn(): void;\ndeclare function getValue(): string;\ndeclare const box: { value: string };\nvoid (fn(), (getValue(), box.value));",
			optionsJson:     "",
			wantComposedFix: "declare function fn(): void;\ndeclare function getValue(): string;\ndeclare const box: { value: string };\n(fn(), (getValue(), box.value));\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (fn(), (getValue(), box.value))", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:      "extra",
			sourceText:  "declare let x: number;\nconst f = () => void x;",
			optionsJson: "",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "None"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "function foo() {}\nvoid void foo();",
			optionsJson:     "",
			wantComposedFix: "function foo() {}\nfoo();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void void foo()", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
				{wantId: "meaninglessVoidOperator", wantSpan: "void foo()", wantMessage: "void operator shouldn't be used on void; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "async function f(p: Promise<number>) {\n  void await p;\n}",
			optionsJson:     "",
			wantComposedFix: "async function f(p: Promise<number>) {\n  await p;\n}\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void await p", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare let x: number;\nvoid (x as any);",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\n(x as any);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (x as any)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare const x: void;\nvoid x;",
			optionsJson:     "",
			wantComposedFix: "declare const x: void;\nx;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare const x: void | undefined;\nvoid x;",
			optionsJson:     "",
			wantComposedFix: "declare const x: void | undefined;\nx;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare const cond: boolean;\ndeclare function a(): void;\ndeclare function b(): void;\nvoid (cond ? a() : b());",
			optionsJson:     "",
			wantComposedFix: "declare const cond: boolean;\ndeclare function a(): void;\ndeclare function b(): void;\n(cond ? a() : b());\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (cond ? a() : b())", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "function foo(): undefined { return undefined; }\nvoid foo();",
			optionsJson:     "",
			wantComposedFix: "function foo(): undefined { return undefined; }\nfoo();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void foo()", wantMessage: "void operator shouldn't be used on undefined; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "function foo() {}\nvoid /* c */ foo();",
			optionsJson:     "",
			wantComposedFix: "function foo() {}\nfoo();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void /* c */ foo()", wantMessage: "void operator shouldn't be used on void; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "function foo() {}\nvoid(foo());",
			optionsJson:     "",
			wantComposedFix: "function foo() {}\n(foo());\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void(foo())", wantMessage: "void operator shouldn't be used on void; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare const x: never;\nvoid x;",
			optionsJson:     "{\"checkNever\": true}",
			wantComposedFix: "declare const x: never;\nx;\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void x", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare function fail(): void | never;\nvoid fail();",
			optionsJson:     "{\"checkNever\": true}",
			wantComposedFix: "declare function fail(): void | never;\nfail();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void fail()", wantMessage: "void operator shouldn't be used on void; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare const box: { method(): void } | undefined;\nvoid box?.method();",
			optionsJson:     "",
			wantComposedFix: "declare const box: { method(): void } | undefined;\nbox?.method();\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOperator", wantSpan: "void box?.method()", wantMessage: "void operator shouldn't be used on void | undefined; it should convey that a return value is being ignored", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare let x: number;\nvoid (x satisfies number);",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\n(x satisfies number);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (x satisfies number)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare const box: { value: string };\nvoid box['value'];",
			optionsJson:     "",
			wantComposedFix: "declare const box: { value: string };\nbox['value'];\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void box['value']", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "void 'use strict';",
			optionsJson:     "",
			wantComposedFix: "'use strict';\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void 'use strict'", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
		{
			origin:          "extra",
			sourceText:      "declare let x: number;\nvoid (x = 1, x);",
			optionsJson:     "",
			wantComposedFix: "declare let x: number;\n(x = 1, x);\n",
			wantFindings: []noMeaninglessVoidOperatorFinding{
				{wantId: "meaninglessVoidOnNonCall", wantSpan: "void (x = 1, x)", wantMessage: "void operator is useless here; it should only discard a call's return value", wantRepair: "Fix"},
			},
		},
	}
}

func TestNoMeaninglessVoidOperatorStaysSilent(t *testing.T) {
	t.Parallel()

	for index, testCase := range noMeaninglessVoidOperatorSilentRows() {
		t.Run(noMeaninglessVoidOperatorCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoMeaninglessVoidOperator,
				noMeaninglessVoidOperatorFile, testCase.sourceText,
				noMeaninglessVoidOperatorOptionsFor(t, testCase.optionsJson)))
		})
	}
}

// Every row asserts the span, the whole rendered message, and the repair. The message matters more
// here than on most rules: the type branch interpolates `typeToString`, so one rule renders "on
// void", "on undefined", "on never" and "on void | undefined" for different inputs.
func TestNoMeaninglessVoidOperatorFires(t *testing.T) {
	t.Parallel()

	for index, testCase := range noMeaninglessVoidOperatorFiringRows() {
		t.Run(noMeaninglessVoidOperatorCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMeaninglessVoidOperator,
				noMeaninglessVoidOperatorFile, testCase.sourceText,
				noMeaninglessVoidOperatorOptionsFor(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage, diagnostic.Message.Description)
				}

				switch want.wantRepair {
				case "Fix":
					if len(diagnostic.Suggestions) != 0 {
						t.Fatalf("finding %d: the fix branch must offer no suggestion, got %d",
							position, len(diagnostic.Suggestions))
					}
					if len(diagnostic.Fixes) != 1 {
						t.Fatalf("finding %d: expected one fix, got %d", position, len(diagnostic.Fixes))
					}
				case "None":
					if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
						t.Fatalf("finding %d: a void that is not a whole statement must offer no repair, "+
							"got %d fixes and %d suggestions", position, len(diagnostic.Fixes),
							len(diagnostic.Suggestions))
					}
				case "Suggestion":
					if len(diagnostic.Fixes) != 0 {
						t.Fatalf("finding %d: the suggestion branch must apply nothing unattended, got %d fixes",
							position, len(diagnostic.Fixes))
					}
					if len(diagnostic.Suggestions) != 1 {
						t.Fatalf("finding %d: expected one suggestion, got %d", position, len(diagnostic.Suggestions))
					}
					suggestion := diagnostic.Suggestions[0]
					if suggestion.Message.Description != want.wantSuggestionDescription {
						t.Fatalf("finding %d suggestion text: expected %q, got %q",
							position, want.wantSuggestionDescription, suggestion.Message.Description)
					}
					if suggestion.Message.Id != "removeVoid" {
						t.Fatalf("finding %d suggestion id: got %q", position, suggestion.Message.Id)
					}
					applied := applyNoMeaninglessVoidOperatorSuggestion(t, onDisk, suggestion)
					if applied != want.wantSuggestion {
						t.Fatalf("finding %d suggestion applied: expected %q, got %q",
							position, want.wantSuggestion, applied)
					}
				default:
					t.Fatalf("finding %d: unknown repair %q", position, want.wantRepair)
				}
			}

			// The whole-file rewrite, asserted against what the installed build composed rather than
			// against anything this rule produced. For the double void it is neither finding's own
			// rewrite, so a row compares the composition. A row with no fix composes to nothing.
			if testCase.wantComposedFix != "" {
				rule_testing.ExpectFixedSource(t, result, testCase.wantComposedFix)
			}
		})
	}
}

// Both of upstream's directions are carried whole, so a bad filter cannot quietly empty either.
func TestNoMeaninglessVoidOperatorCorpusIsWhole(t *testing.T) {
	t.Parallel()

	counts := map[string]int{}
	for _, row := range append(noMeaninglessVoidOperatorSilentRows(), noMeaninglessVoidOperatorFiringRows()...) {
		counts[row.origin]++
	}
	if counts["upstream-valid"] != 22 || counts["upstream-invalid"] != 21 {
		t.Fatalf("upstream's 8.71.0 file has 22 valid and 21 invalid rows, the corpus carries %d and %d",
			counts["upstream-valid"], counts["upstream-invalid"])
	}
}

// TestDecodeNoMeaninglessVoidOperatorOptions pins the decoder against its default.
//
// This option defaults to FALSE, so an absent key and a zero value happen to agree and the generic
// decoder would work by luck. The decoder is hand-rolled anyway and tested anyway, because the
// coincidence is not a property of the option surface: adding one default-true key would make the
// generic path silently wrong, and this test is what would notice.
func TestDecodeNoMeaninglessVoidOperatorOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "emptyObject", raw: `{}`, want: false},
		{name: "explicitTrue", raw: `{"checkNever": true}`, want: true},
		{name: "explicitFalse", raw: `{"checkNever": false}`, want: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoMeaninglessVoidOperatorOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(NoMeaninglessVoidOperatorOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than the options struct", decoded)
			}
			if options.CheckNever != testCase.want {
				t.Fatalf("checkNever: expected %v, got %v", testCase.want, options.CheckNever)
			}
		})
	}
}

// TestNoMeaninglessVoidOperatorFallsBackToTheDefaultOnNilOptions pins the nil-options path.
//
// A rule configured as a bare "error" is handed nil options, and every fixture above reaches the
// rule through the decoder, so nothing there can see the fallback. The separating input is a call
// returning `never`: silent under the default, reporting when checkNever is on. Since 8.71 it has to
// be a call, because `void x` on a `never` name is a non-call finding under either setting.
func TestNoMeaninglessVoidOperatorFallsBackToTheDefaultOnNilOptions(t *testing.T) {
	t.Parallel()

	const neverArgument = "declare function fail(): never;\nvoid fail();"

	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoMeaninglessVoidOperator,
		noMeaninglessVoidOperatorFile, neverArgument, nil))

	// The control: the same input with checkNever on must report, so the silence above is the
	// default being applied rather than the rule being inert.
	reported := rule_testing.RunTypedWithOptions(t, NoMeaninglessVoidOperator,
		noMeaninglessVoidOperatorFile, neverArgument,
		NoMeaninglessVoidOperatorOptions{CheckNever: true})
	rule_testing.ExpectFindings(t, reported, "meaninglessVoidOperator")
}

// TestNoMeaninglessVoidOperatorNeedsTheTypedHarness pins the checker declaration.
//
// Under the plain harness the checker is nil, this rule returns immediately, and every clean
// fixture would pass having proven nothing. Asserting the declaration means a later revert fails
// loudly rather than going vacuously green.
func TestNoMeaninglessVoidOperatorNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoMeaninglessVoidOperator.NeedsTypeChecker {
		t.Fatal("the rule resolves the argument's type, so it must declare NeedsTypeChecker")
	}

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoMeaninglessVoidOperator,
		noMeaninglessVoidOperatorFile, "function foo() {}\nvoid foo();"))
}
