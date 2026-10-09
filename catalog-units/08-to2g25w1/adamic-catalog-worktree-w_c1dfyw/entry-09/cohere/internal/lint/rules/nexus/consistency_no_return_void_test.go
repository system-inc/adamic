package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistencyNoReturnVoidFile = "/repository/source/ConsistencyNoReturnVoid.ts"

func consistencyNoReturnVoidSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// The real sites are the ten in `modules/phi/PhiSocialMediaCommandLineInterface.ts` at HEAD on
// 2026-10-01, in its two shapes: the braceless `if` (eight of them) and the braced one (two). Each case
// asserts the finding and the whole rewritten source.
func TestConsistencyNoReturnVoidFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		fixed  string
	}{
		{
			"PhiSocialMediaCommandLineInterface.ts:106, the braceless if gains braces",
			consistencyNoReturnVoidSource(
				"export async function run(subcommand: string, rest: string[]) {",
				"    if(subcommand === 'show') {",
				"        if(!rest[0]) return void console.log('Usage: ahra phi social show <postId>');",
				"        await show(rest[0]);",
				"    }",
				"}",
			),
			consistencyNoReturnVoidSource(
				"export async function run(subcommand: string, rest: string[]) {",
				"    if(subcommand === 'show') {",
				"        if(!rest[0]) { console.log('Usage: ahra phi social show <postId>'); return; }",
				"        await show(rest[0]);",
				"    }",
				"}",
			),
		},
		{
			"PhiSocialMediaCommandLineInterface.ts:148, inside a block it splits at the same indentation",
			consistencyNoReturnVoidSource(
				"export async function run(postId: string, imagePath: string) {",
				"    if(!postId || !imagePath) {",
				"        return void console.log('Usage: ahra phi social upload <postId> --image <path>');",
				"    }",
				"}",
			),
			consistencyNoReturnVoidSource(
				"export async function run(postId: string, imagePath: string) {",
				"    if(!postId || !imagePath) {",
				"        console.log('Usage: ahra phi social upload <postId> --image <path>');",
				"        return;",
				"    }",
				"}",
			),
		},
		{
			"an else branch and a loop body are braceless positions too",
			consistencyNoReturnVoidSource(
				"function run(ready: boolean, items: string[]) {",
				"    if(ready) start(); else return void stop();",
				"    for(const item of items) return void use(item);",
				"}",
			),
			consistencyNoReturnVoidSource(
				"function run(ready: boolean, items: string[]) {",
				"    if(ready) start(); else { stop(); return; }",
				"    for(const item of items) { use(item); return; }",
				"}",
			),
		},
		{
			"a case clause on its own line keeps its indentation",
			consistencyNoReturnVoidSource(
				"function run(kind: string) {",
				"    switch(kind) {",
				"        case 'A':",
				"            return void start();",
				"    }",
				"}",
			),
			consistencyNoReturnVoidSource(
				"function run(kind: string) {",
				"    switch(kind) {",
				"        case 'A':",
				"            start();",
				"            return;",
				"    }",
				"}",
			),
		},
		{
			"a case clause sharing the line keeps both statements beside it",
			consistencyNoReturnVoidSource(
				"function run(kind: string) {",
				"    switch(kind) {",
				"        case 'A': return void start();",
				"    }",
				"}",
			),
			consistencyNoReturnVoidSource(
				"function run(kind: string) {",
				"    switch(kind) {",
				"        case 'A': start(); return;",
				"    }",
				"}",
			),
		},
		{
			"a default clause is a statement list too",
			consistencyNoReturnVoidSource(
				"function run(kind: string) {",
				"    switch(kind) {",
				"        default:",
				"            return void start();",
				"    }",
				"}",
			),
			consistencyNoReturnVoidSource(
				"function run(kind: string) {",
				"    switch(kind) {",
				"        default:",
				"            start();",
				"            return;",
				"    }",
				"}",
			),
		},
		{
			"void 0 is return undefined in disguise",
			consistencyNoReturnVoidSource(
				"function run() {",
				"    return void 0;",
				"}",
			),
			"",
		},
		{
			"await and this survive as the statement's first token",
			consistencyNoReturnVoidSource(
				"async function run() {",
				"    return void await this.save();",
				"}",
			),
			consistencyNoReturnVoidSource(
				"async function run() {",
				"    await this.save();",
				"    return;",
				"}",
			),
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoReturnVoid, consistencyNoReturnVoidFile, testCase.source)
			want := strings.Count(testCase.source, "return void")
			wantIds := make([]string, 0, want)
			for range want {
				wantIds = append(wantIds, consistencyNoReturnVoidId)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
			if testCase.fixed == "" {
				for _, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Fatalf("expected no fix, got %q", diagnostic.Fixes[0].Text)
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.fixed)
		})
	}
}

// The statements this rule reports and must not rewrite, because the rewritten text would not parse as
// what it says or would eat something the author wrote.
func TestConsistencyNoReturnVoidReportsWithoutAFix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		// `return 0` is a number; the operand here is a digit, which cannot start an identifier.
		{"a numeric operand", []string{"function run() {", "    return void 0;", "}"}},
		// As a statement, `(a, b)` would join the previous line when it has no semicolon.
		{"a parenthesized operand", []string{"function run() {", "    start()", "    return void (first(), second());", "}"}},
		{"an array operand", []string{"function run() {", "    return void [first(), second()];", "}"}},
		{"a template operand", []string{"function run() {", "    return void `${first()}`;", "}"}},
		{"a unary operand", []string{"function run(count: number) {", "    return void -count;", "}"}},
		// As a statement, `function () {}()` is a declaration followed by a stray call.
		{"a function expression operand", []string{"function run() {", "    return void function() { start(); }();", "}"}},
		{"an async arrow operand", []string{"function run() {", "    return void async function() { start(); }();", "}"}},
		{"parentheses around the void", []string{"function run() {", "    return (void start());", "}"}},
		{"a comment between the tokens", []string{"function run() {", "    return void /* log it */ start();", "}"}},
		{"no closing semicolon", []string{"function run() {", "    return void start()", "}"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoReturnVoid, consistencyNoReturnVoidFile, consistencyNoReturnVoidSource(testCase.lines...))
			rule_testing.ExpectFindings(t, result, consistencyNoReturnVoidId)
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected no fix, got %q", result.Diagnostics[0].Fixes[0].Text)
			}
		})
	}
}

func TestConsistencyNoReturnVoidStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		// SeeClusterCard.tsx:254 at HEAD. The callback idiom: an arrow body, not a return statement.
		{"the arrow-shorthand callback idiom", []string{
			"const element = <Button onClick={() => void handleEnroll()} disabled={isEnrolling}>Enroll</Button>;",
		}},
		{"the rewritten form", []string{
			"export async function run(rest: string[]) {",
			"    if(!rest[0]) {",
			"        console.log('Usage: ahra phi social show <postId>');",
			"        return;",
			"    }",
			"}",
		}},
		{"a void expression statement marking a dropped promise", []string{
			"function run() {",
			"    void startServer();",
			"    return;",
			"}",
		}},
		{"a void expression somewhere inside the returned value", []string{
			"function run() {",
			"    return [void start(), 1];",
			"}",
		}},
		{"a bare return and a return of a value", []string{
			"function run(ready: boolean) {",
			"    if(!ready) return;",
			"    return ready;",
			"}",
		}},
		{"the word void in a type and a string", []string{
			"function run(): void {",
			"    const text = 'return void start();';",
			"    use(text);",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoReturnVoid, "/repository/source/ConsistencyNoReturnVoid.tsx", consistencyNoReturnVoidSource(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}
