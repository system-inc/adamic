package typescript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const explicitAnyFile = "/repository/source/Thing.ts"

// The corpus below is oxc's own, copied byte for byte out of its tester blocks by a script and
// verified back against the Rust source the same way, because a fixture retyped by hand encodes
// the belief the port already has and passes for exactly the reason the code is wrong.
//
// Two tester blocks exist. The first is a four-case smoke test that is not snapshotted; the second
// carries the real corpus and its snapshot. Both are imported here, since reading only the last
// block silently drops a corpus.

// TestNoExplicitAnyFires is every input upstream reports on, with the number of findings each one
// produces recovered from the snapshot rather than assumed to be one.
//
// Four inputs report twice, because they write two `any` tokens: a parameter and a return type, an
// element type and a return type, or a type-parameter default and a heritage type argument. A
// fixture asserting one finding per input would be wrong on all four, which is what the extractor
// warns about when the diagnostic count and the input count disagree.
func TestNoExplicitAnyFires(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"block one smoke case", "let x: any = 1", 1},
		{"upstream fail 1", "const number: any = 1", 1},
		{"upstream fail 2", "function generic(): any {}", 1},
		{"upstream fail 3", "function generic(): Array<any>", 1},
		{"upstream fail 4", "function generic(): any[] {}", 1},
		{"upstream fail 5", "function generic(param: Array<any>): number { return 1 }", 1},
		{"upstream fail 6", "function generic(param: any[]): number { return 1 }", 1},
		{"upstream fail 7", "function generic(param: Array<any>): Array<any>", 2},
		{"upstream fail 8", "function generic(): Array<Array<any>>", 1},
		{"upstream fail 9", "function generic(param: Array<any[]>): Array<any>", 2},
		{"upstream fail 10", "class Greeter { constructor(param: Array<any>) {} }", 1},
		{"upstream fail 11", "class Greeter { message: any }", 1},
		{"upstream fail 12", "class Greeter { message: Array<any> }", 1},
		{"upstream fail 13", "class Greeter { message: any[] }", 1},
		{"upstream fail 14", "class Greeter { message: Array<Array<any>> }", 1},
		// Upstream emits `Unexpected token` for this and no rule finding at all: a method
		// signature in an interface may not carry a body. typescript-go recovers from the same
		// syntax error and hands the rule an ordinary parameter list, so the `any` inside it
		// reports here. A parser-tolerance difference, not a rule difference; see the rule doc.
		{"upstream fail 15", "interface Greeter { constructor(param: Array<any>) {} }", 1},
		{"upstream fail 16", "interface Greeter { message: any }", 1},
		{"upstream fail 17", "interface Greeter { message: Array<any> }", 1},
		{"upstream fail 18", "interface Greeter { message: any[] }", 1},
		{"upstream fail 19", "interface Greeter { message: Array<Array<any>> }", 1},
		// Upstream emits `Unexpected token` for this and no rule finding at all: a method
		// signature in an interface may not carry a body. typescript-go recovers from the same
		// syntax error and hands the rule an ordinary parameter list, so the `any` inside it
		// reports here. A parser-tolerance difference, not a rule difference; see the rule doc.
		{"upstream fail 20", "type obj = { constructor(param: Array<any>) {} }", 1},
		{"upstream fail 21", "type obj = { message: any }", 1},
		{"upstream fail 22", "type obj = { message: Array<any> }", 1},
		{"upstream fail 23", "type obj = { message: any[] }", 1},
		{"upstream fail 24", "type obj = { message: Array<Array<any>> }", 1},
		{"upstream fail 25", "type obj = { message: string | any }", 1},
		{"upstream fail 26", "type obj = { message: string | Array<any> }", 1},
		{"upstream fail 27", "type obj = { message: string | Array<any[]> }", 1},
		{"upstream fail 28", "type obj = { message: string | Array<Array<any>> }", 1},
		{"upstream fail 29", "type obj = { message: string & any }", 1},
		{"upstream fail 30", "type obj = { message: string & any[] }", 1},
		{"upstream fail 31", "type obj = { message: string & Array<any> }", 1},
		{"upstream fail 32", "type obj = { message: string & Array<Array<any>> }", 1},
		{"upstream fail 33", "type obj = { message: string & Array<any[]> }", 1},
		{"upstream fail 34", "class Foo<T = any> extends Bar<any> {}", 2},
		{"upstream fail 35", "abstract class Foo<T = any> extends Bar<any> {}", 2},
		{"upstream fail 36", "function test<T extends Partial<any>>() {}", 1},
		{"upstream fail 37", "const test = <T extends Partial<any>>() => {};", 1},
		{"upstream fail 38", "function foo(a: number, ...rest: any[]): void { return; }", 1},
		{"upstream fail 39", "type Any = any;", 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExplicitAny, explicitAnyFile, testCase.sourceText)
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "unexpectedAny"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestNoExplicitAnyStaysSilent is upstream's default-option clean corpus.
//
// These are the cases that catch a port, because each was added when somebody hit that bug. They
// are almost all the same shapes as the failing inputs with a real type in place of `any`, which
// is the point: the rule must key on the keyword and not on the surrounding shape.
func TestNoExplicitAnyStaysSilent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
	}{
		{"block one smoke case", "let x: number = 1"},
		{"upstream clean 1", "const number: number = 1;"},
		{"upstream clean 2", "function greet(): string {}"},
		{"upstream clean 3", "function greet(): Array<string> {}"},
		{"upstream clean 4", "function greet(): string[] {}"},
		{"upstream clean 5", "function greet(): Array<Array<string>> {}"},
		{"upstream clean 6", "function greet(): Array<string[]> {}"},
		{"upstream clean 7", "function greet(param: Array<string>): Array<string> {}"},
		{"upstream clean 8", "\n                class Greeter {\n                  message: string;\n                }\n                    "},
		{"upstream clean 9", "\n                class Greeter {\n                  message: Array<string>;\n                }\n                    "},
		{"upstream clean 10", "\n                class Greeter {\n                  message: string[];\n                }\n                    "},
		{"upstream clean 11", "\n                class Greeter {\n                  message: Array<Array<string>>;\n                }\n                    "},
		{"upstream clean 12", "\n                class Greeter {\n                  message: Array<string[]>;\n                }\n                    "},
		{"upstream clean 13", "\n                interface Greeter {\n                  message: string;\n                }\n                    "},
		{"upstream clean 14", "\n                interface Greeter {\n                  message: Array<string>;\n                }\n                    "},
		{"upstream clean 15", "\n                interface Greeter {\n                  message: string[];\n                }\n                    "},
		{"upstream clean 16", "\n                interface Greeter {\n                  message: Array<Array<string>>;\n                }\n                    "},
		{"upstream clean 17", "\n                interface Greeter {\n                  message: Array<string[]>;\n                }\n                    "},
		{"upstream clean 18", "\n                type obj = {\n                  message: string;\n                };\n                    "},
		{"upstream clean 19", "\n                type obj = {\n                  message: Array<string>;\n                };\n                    "},
		{"upstream clean 20", "\n                type obj = {\n                  message: string[];\n                };\n                    "},
		{"upstream clean 21", "\n                type obj = {\n                  message: Array<Array<string>>;\n                };\n                    "},
		{"upstream clean 22", "\n                type obj = {\n                  message: Array<string[]>;\n                };\n                    "},
		{"upstream clean 23", "\n                type obj = {\n                  message: string | number;\n                };\n                    "},
		{"upstream clean 24", "\n                type obj = {\n                  message: string | Array<string>;\n                };\n                    "},
		{"upstream clean 25", "\n                type obj = {\n                  message: string | string[];\n                };\n                    "},
		{"upstream clean 26", "\n                type obj = {\n                  message: string | Array<Array<string>>;\n                };\n                    "},
		{"upstream clean 27", "\n                type obj = {\n                  message: string & number;\n                };\n                    "},
		{"upstream clean 28", "\n                type obj = {\n                  message: string & Array<string>;\n                };\n                    "},
		{"upstream clean 29", "\n                type obj = {\n                  message: string & string[];\n                };\n                    "},
		{"upstream clean 30", "\n                type obj = {\n                  message: string & Array<Array<string>>;\n                };\n                    "},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoExplicitAny, explicitAnyFile, testCase.sourceText))
		})
	}
}

// TestNoExplicitAnyIgnoreRestArgsStaysSilent is upstream's second option surface.
//
// Forty-five inputs, every one of them an `any` inside a rest parameter, across every construct
// that can hold one: function declarations, function expressions, arrow functions, call and
// construct signatures, method signatures, function and constructor types, and ambient
// declarations. All forty-five are exempt only because the option is on, which the companion test
// below pins by running the same inputs with the option off.
func TestNoExplicitAnyIgnoreRestArgsStaysSilent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
	}{
		{"upstream rest 1", "\n                        function foo(a: number, ...rest: any[]): void {\n                          return;\n                        }\n                      "},
		{"upstream rest 2", "function foo1(...args: any[]) {}"},
		{"upstream rest 3", "const bar1 = function (...args: any[]) {};"},
		{"upstream rest 4", "const baz1 = (...args: any[]) => {};"},
		{"upstream rest 5", "function foo2(...args: readonly any[]) {}"},
		{"upstream rest 6", "const bar2 = function (...args: readonly any[]) {};"},
		{"upstream rest 7", "const baz2 = (...args: readonly any[]) => {};"},
		{"upstream rest 8", "function foo3(...args: Array<any>) {}"},
		{"upstream rest 9", "const bar3 = function (...args: Array<any>) {};"},
		{"upstream rest 10", "const baz3 = (...args: Array<any>) => {};"},
		{"upstream rest 11", "function foo4(...args: ReadonlyArray<any>) {}"},
		{"upstream rest 12", "const bar4 = function (...args: ReadonlyArray<any>) {};"},
		{"upstream rest 13", "const baz4 = (...args: ReadonlyArray<any>) => {};"},
		{"upstream rest 14", "\n                interface Qux1 {\n                  (...args: any[]): void;\n                }\n                      "},
		{"upstream rest 15", "\n                interface Qux2 {\n                  (...args: readonly any[]): void;\n                }\n                      "},
		{"upstream rest 16", "\n                interface Qux3 {\n                  (...args: Array<any>): void;\n                }\n                      "},
		{"upstream rest 17", "\n                interface Qux4 {\n                  (...args: ReadonlyArray<any>): void;\n                }\n                      "},
		{"upstream rest 18", "function quux1(fn: (...args: any[]) => void): void {}"},
		{"upstream rest 19", "function quux2(fn: (...args: readonly any[]) => void): void {}"},
		{"upstream rest 20", "function quux3(fn: (...args: Array<any>) => void): void {}"},
		{"upstream rest 21", "function quux4(fn: (...args: ReadonlyArray<any>) => void): void {}"},
		{"upstream rest 22", "function quuz1(): (...args: any[]) => void {}"},
		{"upstream rest 23", "function quuz2(): (...args: readonly any[]) => void {}"},
		{"upstream rest 24", "function quuz3(): (...args: Array<any>) => void {}"},
		{"upstream rest 25", "function quuz4(): (...args: ReadonlyArray<any>) => void {}"},
		{"upstream rest 26", "type Fred1 = (...args: any[]) => void;"},
		{"upstream rest 27", "type Fred2 = (...args: readonly any[]) => void;"},
		{"upstream rest 28", "type Fred3 = (...args: Array<any>) => void;"},
		{"upstream rest 29", "type Fred4 = (...args: ReadonlyArray<any>) => void;"},
		{"upstream rest 30", "type Corge1 = new (...args: any[]) => void;"},
		{"upstream rest 31", "type Corge2 = new (...args: readonly any[]) => void;"},
		{"upstream rest 32", "type Corge3 = new (...args: Array<any>) => void;"},
		{"upstream rest 33", "type Corge4 = new (...args: ReadonlyArray<any>) => void;"},
		{"upstream rest 34", "\n                interface Grault1 {\n                  new (...args: any[]): void;\n                }\n                      "},
		{"upstream rest 35", "\n                interface Grault2 {\n                  new (...args: readonly any[]): void;\n                }\n                      "},
		{"upstream rest 36", "\n                interface Grault3 {\n                  new (...args: Array<any>): void;\n                }\n                      "},
		{"upstream rest 37", "\n                interface Grault4 {\n                  new (...args: ReadonlyArray<any>): void;\n                }\n                      "},
		{"upstream rest 38", "\n                interface Garply1 {\n                  f(...args: any[]): void;\n                }\n                      "},
		{"upstream rest 39", "\n                interface Garply2 {\n                  f(...args: readonly any[]): void;\n                }\n                      "},
		{"upstream rest 40", "\n                interface Garply3 {\n                  f(...args: Array<any>): void;\n                }\n                      "},
		{"upstream rest 41", "\n                interface Garply4 {\n                  f(...args: ReadonlyArray<any>): void;\n                }\n                      "},
		{"upstream rest 42", "declare function waldo1(...args: any[]): void;"},
		{"upstream rest 43", "declare function waldo2(...args: readonly any[]): void;"},
		{"upstream rest 44", "declare function waldo3(...args: Array<any>): void;"},
		{"upstream rest 45", "declare function waldo4(...args: ReadonlyArray<any>): void;"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoExplicitAny, explicitAnyFile,
				testCase.sourceText, NoExplicitAnyOptions{IgnoreRestArgs: true}))
		})
	}
}

// TestNoExplicitAnyIgnoreRestArgsIsWhatSilencesThem runs the same forty-five inputs with the option
// off and requires every one of them to report.
//
// Without this the clean test above passes vacuously under any rule that reports nothing, and more
// subtly it passes under a rule whose rest detection is broken in the always-exempt direction. A
// clean case is only evidence about the thing you think it is when the edit you believe makes it
// clean is shown to move the verdict.
func TestNoExplicitAnyIgnoreRestArgsIsWhatSilencesThem(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"\n                        function foo(a: number, ...rest: any[]): void {\n                          return;\n                        }\n                      ",
		"function foo1(...args: any[]) {}",
		"const bar1 = function (...args: any[]) {};",
		"const baz1 = (...args: any[]) => {};",
		"function foo2(...args: readonly any[]) {}",
		"const bar2 = function (...args: readonly any[]) {};",
		"const baz2 = (...args: readonly any[]) => {};",
		"function foo3(...args: Array<any>) {}",
		"const bar3 = function (...args: Array<any>) {};",
		"const baz3 = (...args: Array<any>) => {};",
		"function foo4(...args: ReadonlyArray<any>) {}",
		"const bar4 = function (...args: ReadonlyArray<any>) {};",
		"const baz4 = (...args: ReadonlyArray<any>) => {};",
		"\n                interface Qux1 {\n                  (...args: any[]): void;\n                }\n                      ",
		"\n                interface Qux2 {\n                  (...args: readonly any[]): void;\n                }\n                      ",
		"\n                interface Qux3 {\n                  (...args: Array<any>): void;\n                }\n                      ",
		"\n                interface Qux4 {\n                  (...args: ReadonlyArray<any>): void;\n                }\n                      ",
		"function quux1(fn: (...args: any[]) => void): void {}",
		"function quux2(fn: (...args: readonly any[]) => void): void {}",
		"function quux3(fn: (...args: Array<any>) => void): void {}",
		"function quux4(fn: (...args: ReadonlyArray<any>) => void): void {}",
		"function quuz1(): (...args: any[]) => void {}",
		"function quuz2(): (...args: readonly any[]) => void {}",
		"function quuz3(): (...args: Array<any>) => void {}",
		"function quuz4(): (...args: ReadonlyArray<any>) => void {}",
		"type Fred1 = (...args: any[]) => void;",
		"type Fred2 = (...args: readonly any[]) => void;",
		"type Fred3 = (...args: Array<any>) => void;",
		"type Fred4 = (...args: ReadonlyArray<any>) => void;",
		"type Corge1 = new (...args: any[]) => void;",
		"type Corge2 = new (...args: readonly any[]) => void;",
		"type Corge3 = new (...args: Array<any>) => void;",
		"type Corge4 = new (...args: ReadonlyArray<any>) => void;",
		"\n                interface Grault1 {\n                  new (...args: any[]): void;\n                }\n                      ",
		"\n                interface Grault2 {\n                  new (...args: readonly any[]): void;\n                }\n                      ",
		"\n                interface Grault3 {\n                  new (...args: Array<any>): void;\n                }\n                      ",
		"\n                interface Grault4 {\n                  new (...args: ReadonlyArray<any>): void;\n                }\n                      ",
		"\n                interface Garply1 {\n                  f(...args: any[]): void;\n                }\n                      ",
		"\n                interface Garply2 {\n                  f(...args: readonly any[]): void;\n                }\n                      ",
		"\n                interface Garply3 {\n                  f(...args: Array<any>): void;\n                }\n                      ",
		"\n                interface Garply4 {\n                  f(...args: ReadonlyArray<any>): void;\n                }\n                      ",
		"declare function waldo1(...args: any[]): void;",
		"declare function waldo2(...args: readonly any[]): void;",
		"declare function waldo3(...args: Array<any>): void;",
		"declare function waldo4(...args: ReadonlyArray<any>): void;",
	}

	for _, sourceText := range testCases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExplicitAny, explicitAnyFile, sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected the rest-parameter `any` to report with ignoreRestArgs off, got silence for %q", sourceText)
			}
		})
	}
}

// TestNoExplicitAnyPointsAtTheKeywordItself asserts where each finding lands, which no message-id
// assertion can see.
//
// The span is the `any` token alone and nothing around it: not the annotation, not the declaration,
// not the type argument that holds it. That matters twice over here, because the same range is what
// the `fixToUnknown` repair overwrites, so a span one byte wide in either direction would write
// `unknown` over a colon or eat the following bracket. Columns cross-checked against oxc's own
// snapshot, which is one-based where these are zero-based.
func TestNoExplicitAnyPointsAtTheKeywordItself(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		wantSpans  []string
		wantStarts []int
	}{
		{"variable annotation", "const number: any = 1", []string{"any"}, []int{14}},
		{"return type", "function generic(): any {}", []string{"any"}, []int{20}},
		{"type argument", "function generic(): Array<any>", []string{"any"}, []int{26}},
		{"array element type", "function generic(): any[] {}", []string{"any"}, []int{20}},
		{"two findings in source order", "function generic(param: Array<any>): Array<any>",
			[]string{"any", "any"}, []int{30, 43}},
		{"type parameter default and heritage argument", "class Foo<T = any> extends Bar<any> {}",
			[]string{"any", "any"}, []int{14, 31}},
		{"bare type alias", "type Any = any;", []string{"any"}, []int{11}},
		{"as assertion", "const x = y as any;", []string{"any"}, []int{15}},
		{"index signature value", "interface I { [k: string]: any }", []string{"any"}, []int{27}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExplicitAny, explicitAnyFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantSpans[index] {
					t.Errorf("finding %d covered %q, wanted %q", index, reported, testCase.wantSpans[index])
				}
				if diagnostic.Range.Pos() != testCase.wantStarts[index] {
					t.Errorf("finding %d started at %d, wanted %d", index, diagnostic.Range.Pos(), testCase.wantStarts[index])
				}
			}
		})
	}
}

// TestNoExplicitAnyReportsOneMessage asserts the identifier and the description against literals
// typed here rather than against the rule's own constant.
//
// Comparing a finding to the constant it was reported with is equality that looks correct and
// measures nothing, because both sides move together under any mutation of the constant. These are
// literals, so a renamed identifier or a rewritten description fails here.
func TestNoExplicitAnyReportsOneMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoExplicitAny, explicitAnyFile, "const number: any = 1")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Id != "unexpectedAny" {
		t.Errorf("message id was %q, wanted %q", result.Diagnostics[0].Message.Id, "unexpectedAny")
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description,
		"This annotation is `any`, which switches the type system off") {
		t.Errorf("description began %q", result.Diagnostics[0].Message.Description)
	}
	if !strings.Contains(result.Diagnostics[0].Message.Description, "`unknown`") {
		t.Errorf("description never mentions the replacement it recommends: %q",
			result.Diagnostics[0].Message.Description)
	}
}

// TestNoExplicitAnyOffersNoRepairByDefault pins that the rewrite is opt-in.
//
// Measured on the release binary: `oxlint --fix` with no options left the file byte-identical, and
// the same run with `fixToUnknown` set rewrote it. A port offering the repair unconditionally would
// pass every message-id fixture in this file while silently rewriting Kirk's tree, which is the
// exact failure mode an unattended fix has and a suggestion does not.
func TestNoExplicitAnyOffersNoRepairByDefault(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoExplicitAny, explicitAnyFile, "let x: any = 1")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Errorf("wanted no fix without fixToUnknown, got %d", len(result.Diagnostics[0].Fixes))
	}
	if len(result.Diagnostics[0].Suggestions) != 0 {
		t.Errorf("wanted no suggestion either, got %d", len(result.Diagnostics[0].Suggestions))
	}
}

// TestNoExplicitAnyFixToUnknownRewritesTheKeyword applies the repair and compares the resulting
// source, which is the only assertion that can see a fix writing the right text over the wrong span.
//
// All eight of upstream's fix vectors are here: two from the first tester block, six from the
// second. The extractor reported only two, because it reads the first block's `fix` binding and
// stops, so the six in the second block were found by grepping the rule file. A port trusting the
// tool's count would have asserted a quarter of the repairs upstream pins.
func TestNoExplicitAnyFixToUnknownRewritesTheKeyword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		sourceText string
		wantSource string
	}{
		{"let x: any = 1", "let x: unknown = 1"},
		{"function foo(): any", "function foo(): unknown"},
		{"function foo(args: any): void {}", "function foo(args: unknown): void {}"},
		{"function foo(args: any[]): void {}", "function foo(args: unknown[]): void {}"},
		{"function foo(...args: any[]): void {}", "function foo(...args: unknown[]): void {}"},
		{"function foo(args: Array<any>): void {}", "function foo(args: Array<unknown>): void {}"},
		// Beyond upstream's vectors: two findings in one file, to pin that the repairs do not shift
		// each other's spans when both are applied.
		{"function generic(param: Array<any>): Array<any>", "function generic(param: Array<unknown>): Array<unknown>"},
		{"class Foo<T = any> extends Bar<any> {}", "class Foo<T = unknown> extends Bar<unknown> {}"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoExplicitAny, explicitAnyFile, testCase.sourceText,
				NoExplicitAnyOptions{FixToUnknown: true})
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource)
		})
	}
}

// TestNoExplicitAnyFixToUnknownDoesNotDefeatIgnoreRestArgs runs both flags together.
//
// Upstream notes at the bottom of its fix vectors that it has no way to test this combination,
// because its harness panics when an expected fix does not occur. Ours does not have that
// limitation, so the case upstream could not write is written here: an exempt `any` must produce no
// finding and therefore no repair, while a reportable one in the same file is still rewritten.
func TestNoExplicitAnyFixToUnknownDoesNotDefeatIgnoreRestArgs(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoExplicitAny, explicitAnyFile,
		"function foo(a: any, ...rest: any[]): void {}",
		NoExplicitAnyOptions{FixToUnknown: true, IgnoreRestArgs: true})
	rule_testing.ExpectFindings(t, result, "unexpectedAny")
	rule_testing.ExpectFixedSource(t, result, "function foo(a: unknown, ...rest: any[]): void {}")
}

// TestNoExplicitAnyDeclinesJavaScriptFiles is the file gate, and it is the fixture this port would
// most regret not having.
//
// typescript-go resolves JSDoc type comments into real type nodes, so a `.js` file genuinely
// produces a `KindAnyKeyword` node here that oxc's parser never produces. Both JSDoc inputs below
// were run against the release binary and both are silent there, with a `.ts` control firing in the
// same batch. Without the gate this rule would report every JSDoc `any` in every JavaScript file in
// the tree, and no imported fixture could see it because upstream cannot express the input.
func TestNoExplicitAnyDeclinesJavaScriptFiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		fileName   string
		sourceText string
	}{
		{"/repository/source/Thing.js", "/** @type {any} */ let x = 1;"},
		{"/repository/source/Thing.jsx", "/** @param {any} p */ function f(p) {}"},
		{"/repository/source/Thing.mjs", "/** @type {any} */ let x = 1;"},
		{"/repository/source/Thing.cjs", "/** @type {any} */ let x = 1;"},
		// An `any` written as real syntax rather than in a comment. typescript-go parses it in a
		// `.js` file too, so this is a second route to the same false positive.
		{"/repository/source/Thing.js", "let x: any = 1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.fileName+" "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoExplicitAny, testCase.fileName, testCase.sourceText))
		})
	}
}

// TestNoExplicitAnyCoversEveryTypeScriptExtension is the control for the test above.
//
// A gate that declined everything would pass the JavaScript test vacuously. Each extension here was
// run against the release binary and each one reports there, `.d.ts` included: a rule about type
// annotations does not exempt the files that are nothing but type annotations, which is the
// plausible wrong guess.
func TestNoExplicitAnyCoversEveryTypeScriptExtension(t *testing.T) {
	t.Parallel()

	for _, fileName := range []string{
		"/repository/source/Thing.ts",
		"/repository/source/Thing.tsx",
		"/repository/source/Thing.mts",
		"/repository/source/Thing.cts",
		"/repository/source/Thing.d.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoExplicitAny, fileName, "declare const x: any;"),
				"unexpectedAny")
		})
	}
}

// TestNoExplicitAnyIgnoreRestArgsExemptsByContainmentNotByShape is the measurement upstream's corpus
// cannot make.
//
// All forty-five of upstream's option-carrying pass cases write the rest type as `any[]`,
// `readonly any[]`, `Array<any>` or `ReadonlyArray<any>`, so a port that exempted only those four
// shapes would pass every one of them. These five inputs separate that reading from the real one,
// and every verdict below was taken from the release binary rather than from reading the Rust.
func TestNoExplicitAnyIgnoreRestArgsExemptsByContainmentNotByShape(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// Not an array shape at all, and still exempt.
		{"nested in a generic", "function r04(...args: Map<any, any>) {}", nil},
		// Two levels deep inside the rest type.
		{"nested array", "function r06(...args: any[][]) {}", nil},
		// The case upstream commented out under a `// todo`, which is stale: this is silent on the
		// release binary, with a control firing on the next line of the same file.
		{"bare rest any", "function r01(...args: any) {}", nil},
		// The input that pins the walk as scoped rather than whole-file: the `any` inside the inner
		// rest parameter is exempt, the return type beside it is not.
		{"return type beside an exempt rest", "function r05(cb: (...a: any[]) => any) {}",
			[]string{"unexpectedAny"}},
		// A non-rest parameter sharing a signature with a rest one still reports.
		{"non-rest parameter alongside", "function r03(a: any, ...b: any[]) {}", []string{"unexpectedAny"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoExplicitAny, explicitAnyFile, testCase.sourceText,
				NoExplicitAnyOptions{IgnoreRestArgs: true})
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoExplicitAnyReportsPositionsOurTreeActuallyWrites covers `any` homes upstream's corpus does
// not, taken from reading our own code rather than invented.
//
// Our tree is full of these shapes and none of them appears in the imported corpus, so without them
// the port's coverage of real inputs would rest entirely on the assumption that one node kind
// behaves the same everywhere. Each was confirmed reporting on the release binary.
func TestNoExplicitAnyReportsPositionsOurTreeActuallyWrites(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"as assertion", "const x = y as any;", 1},
		{"angle bracket assertion", "const x = <any>y;", 1},
		{"index signature value", "interface I { [k: string]: any }", 1},
		{"generic default on a function", "function f<T = any>() {}", 1},
		{"record value type", "type R = Record<string, any>;", 1},
		{"promise type argument", "async function f(): Promise<any> {}", 1},
		{"optional property", "interface I { p?: any }", 1},
		{"tuple member", "type T = [any, string];", 1},
		{"catch clause annotation", "try {} catch (e: any) {}", 1},
		{"satisfies operand", "const x = {} satisfies any;", 1},
		{"mapped type value", "type M = { [K in string]: any };", 1},
		{"conditional branch", "type C<T> = T extends string ? any : never;", 1},
		{"declared class property", "class C { p: any; }", 1},
		{"arrow return type", "const f = (): any => 1;", 1},
		{"two in one union", "type U = any | any;", 2},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExplicitAny, explicitAnyFile, testCase.sourceText)
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "unexpectedAny"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestNoExplicitAnyIgnoresIdentifiersNamedAny pins that the rule keys on the keyword node and not on
// the text `any` anywhere it appears.
//
// A port built on text matching rather than on the node kind would pass every fixture above and
// report all of these. `type Any = any;` is upstream's own case and reports exactly once, which is
// the same distinction stated from the other side.
func TestNoExplicitAnyIgnoresIdentifiersNamedAny(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"const any = 1;",
		"function any() {}",
		"const o = { any: 1 };",
		"const s = \"any\";",
		"// any",
		"type any2 = string;",
		"import { any } from \"./m\";",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoExplicitAny, explicitAnyFile, sourceText))
		})
	}
}
