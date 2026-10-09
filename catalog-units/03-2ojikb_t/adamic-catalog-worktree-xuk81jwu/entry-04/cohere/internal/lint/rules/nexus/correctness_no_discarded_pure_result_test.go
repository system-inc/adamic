package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoDiscardedPureResultFile = "/repository/source/CorrectnessNoDiscardedPureResult.ts"

var correctnessNoDiscardedPureResultPrelude = strings.Join([]string{
	"declare let name: string;",
	"declare const maybeName: string | undefined;",
	"declare const path: string;",
	"declare let list: string[];",
	"declare const more: string[];",
	"declare const items: readonly string[];",
	"declare const either: string | string[];",
	"declare const pattern: RegExp;",
	"declare const matches: string[];",
	"declare function use(value: unknown): void;",
	"declare function replacer(match: string): string;",
	"",
}, "\n")

func correctnessNoDiscardedPureResultSource(lines ...string) string {
	return correctnessNoDiscardedPureResultPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoDiscardedPureResultRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoDiscardedPureResult, map[string]string{
		correctnessNoDiscardedPureResultFile: sourceText,
	}, correctnessNoDiscardedPureResultFile)
}

// correctnessNoDiscardedPureResultExpect asserts the reported spans in source order, and that no
// finding carries a fix.
func correctnessNoDiscardedPureResultExpect(t *testing.T, result rule_testing.Result, want ...string) {
	t.Helper()
	// The shared assertion by message id, which proves the rule can fire; the checks below hold
	// each finding's exact text.
	if len(want) > 0 {
		wantIds := make([]string, len(want))
		for index := range wantIds {
			wantIds[index] = correctnessNoDiscardedPureResultId
		}
		rule_testing.ExpectFindings(t, result, wantIds...)
	}
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	text := result.SourceFile.Text()
	var got []string
	for _, diagnostic := range diagnostics {
		if diagnostic.Message.Id != correctnessNoDiscardedPureResultId {
			t.Fatalf("unexpected message id %q", diagnostic.Message.Id)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
		got = append(got, text[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d findings %q, got %d %q", len(want), want, len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("finding %d is %q, want %q", index, got[index], want[index])
		}
	}
}

// The slip itself, as the research names it: a trim, a replace and a concat whose results are lost.
// Each is fixed by keeping the result, which is silent.
func TestCorrectnessNoDiscardedPureResultFiresOnTheSlipAndNotOnItsFix(t *testing.T) {
	t.Parallel()

	slip := correctnessNoDiscardedPureResultRun(t, correctnessNoDiscardedPureResultSource(
		"export function normalize() {",
		"    name.trim();",
		"    path.replace('/old/', '/new/');",
		"    list.concat(more);",
		"}",
	))
	correctnessNoDiscardedPureResultExpect(t, slip, "name.trim()", "path.replace('/old/', '/new/')", "list.concat(more)")

	fixed := correctnessNoDiscardedPureResultRun(t, correctnessNoDiscardedPureResultSource(
		"export function normalize() {",
		"    name = name.trim();",
		"    const moved = path.replace('/old/', '/new/');",
		"    list = list.concat(more);",
		"    return moved;",
		"}",
	))
	rule_testing.ExpectClean(t, fixed)
}

func TestCorrectnessNoDiscardedPureResultFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"a slice", []string{"list.slice(1);"}, []string{"list.slice(1)"}},
		{"an optional call", []string{"maybeName?.toLowerCase();"}, []string{"maybeName?.toLowerCase()"}},
		{"parenthesized", []string{"(name.toUpperCase());"}, []string{"name.toUpperCase()"}},
		{"a readonly array", []string{"items.indexOf('a');"}, []string{"items.indexOf('a')"}},
		{"a union of string and array", []string{"either.includes('a');"}, []string{"either.includes('a')"}},
		{"a replace with a regex and a string", []string{"name.replace(pattern, '');"}, []string{"name.replace(pattern, '')"}},
		{"a join and a split", []string{"list.join(',');", "path.split('/');"}, []string{"list.join(',')", "path.split('/')"}},
		{"a chain ending in a pure call", []string{"name.trim().padStart(4);"}, []string{"name.trim().padStart(4)"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoDiscardedPureResultRun(t, correctnessNoDiscardedPureResultSource(testCase.lines...))
			correctnessNoDiscardedPureResultExpect(t, result, testCase.want...)
		})
	}
}

func TestCorrectnessNoDiscardedPureResultStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"discarded on purpose with void", []string{"void name.trim();"}},
		{"a replace walking matches with a callback", []string{
			"name.replace(pattern, function(match) { matches.push(match); return match; });",
		}},
		{"a replace with a named replacer function", []string{"name.replace(pattern, replacer);"}},
		{"a replace with a replacer of type any", []string{"declare const loose: any;", "name.replace(pattern, loose);"}},
		{"a mutating array method", []string{"list.push('a');", "list.sort();", "list.splice(0, 1);"}},
		{"a callback array method, a style question", []string{"list.map((each) => use(each));", "list.forEach(use);"}},
		{"a method that can throw on a bad argument", []string{"name.normalize('NFC');", "name.repeat(2);", "name.localeCompare('a');"}},
		{"a class of the project's own with a trim method", []string{
			"class Text { trim(): Text { use(this); return this; } }",
			"new Text().trim();",
		}},
		{"an interface of the project's own named String", []string{
			"interface Label { slice(start: number): Label }",
			"declare const label: Label;",
			"label.slice(1);",
		}},
		{"a local interface named Array, which is not the library's", []string{
			"interface Array<T> { slice(start: number): void; first: T }",
			"declare const local: Array<string>;",
			"local.slice(1);",
		}},
		{"an any receiver", []string{"declare const loose: any;", "loose.trim();"}},
		{"a result kept", []string{"use(name.trim());", "const pieces = path.split('/');", "use(pieces);"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoDiscardedPureResultRun(t, correctnessNoDiscardedPureResultSource(testCase.lines...)))
		})
	}
}
