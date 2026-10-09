package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's whole use-isnan corpus, replayed through the installed rule under the TypeScript
 * parser (#6esg2nx), now agreeing on every row and every span (#9g0v4j6).
 *
 * Extracted from upstream's tests/lib/rules/use-isnan.js at v10.8.1 by stubbing RuleTester, each row
 * run through ESLint's Linter with @typescript-eslint/parser, and the answers written here. Upstream's
 * `caseNaN` and `switchNaN` are this rule's `caseWithNaN` and `switchOnNaN`. The rows run through the
 * typed harness, because 18 of them shadow NaN or Number with a local and only the checker can tell.
 *
 * Every row that sets `enforceForIndexOf: true` is also run with it false, and must then report no
 * indexOfNaN: the option was decoded and never read until this task, so off and on looked the same.
 */

func TestUseIsNaNAgreesWithESLintsCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		options string
		code    string
		want    []string
	}{
		{"{}", "var x = NaN;", nil},
		{"{}", "isNaN(NaN) === true;", nil},
		{"{}", "isNaN(123) !== true;", nil},
		{"{}", "Number.isNaN(NaN) === true;", nil},
		{"{}", "Number.isNaN(123) !== true;", nil},
		{"{}", "foo(NaN + 1);", nil},
		{"{}", "foo(1 + NaN);", nil},
		{"{}", "foo(NaN - 1)", nil},
		{"{}", "foo(1 - NaN)", nil},
		{"{}", "foo(NaN * 2)", nil},
		{"{}", "foo(2 * NaN)", nil},
		{"{}", "foo(NaN / 2)", nil},
		{"{}", "foo(2 / NaN)", nil},
		{"{}", "var x; if (x = NaN) { }", nil},
		{"{}", "var x = Number.NaN;", nil},
		{"{}", "isNaN(Number.NaN) === true;", nil},
		{"{}", "Number.isNaN(Number.NaN) === true;", nil},
		{"{}", "foo(Number.NaN + 1);", nil},
		{"{}", "foo(1 + Number.NaN);", nil},
		{"{}", "foo(Number.NaN - 1)", nil},
		{"{}", "foo(1 - Number.NaN)", nil},
		{"{}", "foo(Number.NaN * 2)", nil},
		{"{}", "foo(2 * Number.NaN)", nil},
		{"{}", "foo(Number.NaN / 2)", nil},
		{"{}", "foo(2 / Number.NaN)", nil},
		{"{}", "var x; if (x = Number.NaN) { }", nil},
		{"{}", "x === Number[NaN];", nil},
		{"{}", "x === (NaN, 1)", nil},
		{"{}", "x === (doStuff(), NaN, 1)", nil},
		{"{}", "x === (doStuff(), Number.NaN, 1)", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(NaN) { case foo: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(foo) { case NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(NaN) { case NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(foo) { case bar: break; case NaN: break; default: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: NaN; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { default: NaN; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(Nan) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch('NaN') { default: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo(NaN)) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo.NaN) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case Nan: break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case 'NaN': break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case foo(NaN): break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case foo.NaN: break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: break; case 1: break; default: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(Number.NaN) { case foo: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(foo) { case Number.NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(NaN) { case Number.NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":false}", "switch(foo) { case bar: break; case Number.NaN: break; default: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: Number.NaN; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { default: Number.NaN; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(Number.Nan) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch('Number.NaN') { default: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo(Number.NaN)) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo.Number.NaN) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case Number.Nan: break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case 'Number.NaN': break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case foo(Number.NaN): break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case foo.Number.NaN: break }", nil},
		{"{\"enforceForSwitchCase\":true}", "switch((NaN, doStuff(), 1)) {}", nil},
		{"{\"enforceForSwitchCase\":true}", "switch((Number.NaN, doStuff(), 1)) {}", nil},
		{"{}", "foo.indexOf(NaN)", nil},
		{"{}", "foo.lastIndexOf(NaN)", nil},
		{"{}", "foo.indexOf(Number.NaN)", nil},
		{"{}", "foo.lastIndexOf(Number.NaN)", nil},
		{"{}", "foo.indexOf(NaN)", nil},
		{"{}", "foo.lastIndexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":false}", "foo.indexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":false}", "foo.lastIndexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "indexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "lastIndexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "new foo.indexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.bar(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.IndexOf(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo[indexOf](NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo[lastIndexOf](NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "indexOf.foo(NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf()", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf()", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(a)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Nan)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(a, NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN, b, c)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(a, b)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN, NaN, b)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(...NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN())", nil},
		{"{}", "foo.indexOf(Number.NaN)", nil},
		{"{}", "foo.lastIndexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":false}", "foo.indexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":false}", "foo.lastIndexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "indexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "lastIndexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "new foo.indexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.bar(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.IndexOf(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo[indexOf](Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo[lastIndexOf](Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "indexOf.foo(Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Number.Nan)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(a, Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Number.NaN, b, c)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Number.NaN, NaN, b)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(...Number.NaN)", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Number.NaN())", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf((NaN, 1))", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf((NaN, 1))", nil},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf((Number.NaN, 1))", nil},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf((Number.NaN, 1))", nil},
		{"{}", "let NaN; if (x === NaN) {}", nil},
		{"{}", "let Number; if (x === Number.NaN) {}", nil},
		{"{}", "function f(NaN) { return x === NaN; }", nil},
		{"{}", "function f(Number) { return x === Number.NaN; }", nil},
		{"{}", "if (x === NaN) {}", []string{"comparisonWithNaN x === NaN"}},
		{"{}", "if (x === Number.NaN) {}", []string{"comparisonWithNaN x === Number.NaN"}},
		{"{\"enforceForSwitchCase\":true}", "let NaN; switch (foo) { case NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "let Number; switch (foo) { case Number.NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "let NaN; switch (NaN) { case a: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "let Number; switch (Number.NaN) { case a: break; }", nil},
		{"{\"enforceForIndexOf\":true}", "function foo(NaN) { return arr.indexOf(NaN); }", nil},
		{"{\"enforceForIndexOf\":true}", "function foo(NaN) { return arr.lastIndexOf(NaN); }", nil},
		{"{\"enforceForIndexOf\":true}", "function foo(Number) { return arr.indexOf(Number.NaN); }", nil},
		{"{\"enforceForIndexOf\":true}", "function foo(Number) { return arr.lastIndexOf(Number.NaN); }", nil},
		{"{}", "let Number; if (x === Number?.NaN) {}", nil},
		{"{}", "function f(Number) { return x === Number?.NaN; }", nil},
		{"{\"enforceForSwitchCase\":true}", "let Number; switch (foo) { case Number?.NaN: break; }", nil},
		{"{\"enforceForSwitchCase\":true}", "let Number; switch (Number?.NaN) { case a: break; }", nil},
		{"{\"enforceForIndexOf\":true}", "function foo(Number) { return arr.indexOf(Number?.NaN); }", nil},
		{"{\"enforceForIndexOf\":true}", "function foo(Number) { return arr.lastIndexOf(Number?.NaN); }", nil},
		{"{}", "123 == NaN;", []string{"comparisonWithNaN 123 == NaN"}},
		{"{}", "123 === NaN;", []string{"comparisonWithNaN 123 === NaN"}},
		{"{}", "NaN === \"abc\";", []string{"comparisonWithNaN NaN === \"abc\""}},
		{"{}", "NaN == \"abc\";", []string{"comparisonWithNaN NaN == \"abc\""}},
		{"{}", "123 != NaN;", []string{"comparisonWithNaN 123 != NaN"}},
		{"{}", "123 !== NaN;", []string{"comparisonWithNaN 123 !== NaN"}},
		{"{}", "NaN !== \"abc\";", []string{"comparisonWithNaN NaN !== \"abc\""}},
		{"{}", "NaN != \"abc\";", []string{"comparisonWithNaN NaN != \"abc\""}},
		{"{}", "NaN < \"abc\";", []string{"comparisonWithNaN NaN < \"abc\""}},
		{"{}", "\"abc\" < NaN;", []string{"comparisonWithNaN \"abc\" < NaN"}},
		{"{}", "NaN > \"abc\";", []string{"comparisonWithNaN NaN > \"abc\""}},
		{"{}", "\"abc\" > NaN;", []string{"comparisonWithNaN \"abc\" > NaN"}},
		{"{}", "NaN <= \"abc\";", []string{"comparisonWithNaN NaN <= \"abc\""}},
		{"{}", "\"abc\" <= NaN;", []string{"comparisonWithNaN \"abc\" <= NaN"}},
		{"{}", "NaN >= \"abc\";", []string{"comparisonWithNaN NaN >= \"abc\""}},
		{"{}", "\"abc\" >= NaN;", []string{"comparisonWithNaN \"abc\" >= NaN"}},
		{"{}", "123 == Number.NaN;", []string{"comparisonWithNaN 123 == Number.NaN"}},
		{"{}", "123 === Number.NaN;", []string{"comparisonWithNaN 123 === Number.NaN"}},
		{"{}", "Number.NaN === \"abc\";", []string{"comparisonWithNaN Number.NaN === \"abc\""}},
		{"{}", "Number.NaN == \"abc\";", []string{"comparisonWithNaN Number.NaN == \"abc\""}},
		{"{}", "123 != Number.NaN;", []string{"comparisonWithNaN 123 != Number.NaN"}},
		{"{}", "123 !== Number.NaN;", []string{"comparisonWithNaN 123 !== Number.NaN"}},
		{"{}", "Number.NaN !== \"abc\";", []string{"comparisonWithNaN Number.NaN !== \"abc\""}},
		{"{}", "Number.NaN != \"abc\";", []string{"comparisonWithNaN Number.NaN != \"abc\""}},
		{"{}", "Number.NaN < \"abc\";", []string{"comparisonWithNaN Number.NaN < \"abc\""}},
		{"{}", "\"abc\" < Number.NaN;", []string{"comparisonWithNaN \"abc\" < Number.NaN"}},
		{"{}", "Number.NaN > \"abc\";", []string{"comparisonWithNaN Number.NaN > \"abc\""}},
		{"{}", "\"abc\" > Number.NaN;", []string{"comparisonWithNaN \"abc\" > Number.NaN"}},
		{"{}", "Number.NaN <= \"abc\";", []string{"comparisonWithNaN Number.NaN <= \"abc\""}},
		{"{}", "\"abc\" <= Number.NaN;", []string{"comparisonWithNaN \"abc\" <= Number.NaN"}},
		{"{}", "Number.NaN >= \"abc\";", []string{"comparisonWithNaN Number.NaN >= \"abc\""}},
		{"{}", "\"abc\" >= Number.NaN;", []string{"comparisonWithNaN \"abc\" >= Number.NaN"}},
		{"{}", "x === Number?.NaN;", []string{"comparisonWithNaN x === Number?.NaN"}},
		{"{}", "x !== Number?.NaN;", []string{"comparisonWithNaN x !== Number?.NaN"}},
		{"{}", "x === Number['NaN'];", []string{"comparisonWithNaN x === Number['NaN']"}},
		{"{}", "/* just\n                adding */ x /* some */ === /* comments */ NaN; // here", []string{"comparisonWithNaN x /* some */ === /* comments */ NaN"}},
		{"{}", "(1, 2) === NaN;", []string{"comparisonWithNaN (1, 2) === NaN"}},
		{"{}", "x === (doStuff(), NaN);", []string{"comparisonWithNaN x === (doStuff(), NaN)"}},
		{"{}", "x === (doStuff(), Number.NaN);", []string{"comparisonWithNaN x === (doStuff(), Number.NaN)"}},
		{"{}", "x == (doStuff(), NaN);", []string{"comparisonWithNaN x == (doStuff(), NaN)"}},
		{"{}", "x == (doStuff(), Number.NaN);", []string{"comparisonWithNaN x == (doStuff(), Number.NaN)"}},
		{"{}", "switch(NaN) { case foo: break; }", []string{"switchOnNaN switch(NaN) { case foo: break; }"}},
		{"{}", "switch(foo) { case NaN: break; }", []string{"caseWithNaN case NaN: break;"}},
		{"{}", "switch(NaN) { case foo: break; }", []string{"switchOnNaN switch(NaN) { case foo: break; }"}},
		{"{}", "switch(foo) { case NaN: break; }", []string{"caseWithNaN case NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(NaN) {}", []string{"switchOnNaN switch(NaN) {}"}},
		{"{\"enforceForSwitchCase\":true}", "switch(NaN) { case foo: break; }", []string{"switchOnNaN switch(NaN) { case foo: break; }"}},
		{"{\"enforceForSwitchCase\":true}", "switch(NaN) { default: break; }", []string{"switchOnNaN switch(NaN) { default: break; }"}},
		{"{\"enforceForSwitchCase\":true}", "switch(NaN) { case foo: break; default: break; }", []string{"switchOnNaN switch(NaN) { case foo: break; default: break; }"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case NaN: }", []string{"caseWithNaN case NaN:"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case NaN: break; }", []string{"caseWithNaN case NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case (NaN): break; }", []string{"caseWithNaN case (NaN): break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: break; case NaN: break; default: break; }", []string{"caseWithNaN case NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: case NaN: default: break; }", []string{"caseWithNaN case NaN:"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: break; case NaN: break; case baz: break; case NaN: break; }", []string{"caseWithNaN case NaN: break;", "caseWithNaN case NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(NaN) { case NaN: break; }", []string{"switchOnNaN switch(NaN) { case NaN: break; }", "caseWithNaN case NaN: break;"}},
		{"{}", "switch(Number.NaN) { case foo: break; }", []string{"switchOnNaN switch(Number.NaN) { case foo: break; }"}},
		{"{}", "switch(foo) { case Number.NaN: break; }", []string{"caseWithNaN case Number.NaN: break;"}},
		{"{}", "switch(Number.NaN) { case foo: break; }", []string{"switchOnNaN switch(Number.NaN) { case foo: break; }"}},
		{"{}", "switch(foo) { case Number.NaN: break; }", []string{"caseWithNaN case Number.NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(Number.NaN) {}", []string{"switchOnNaN switch(Number.NaN) {}"}},
		{"{\"enforceForSwitchCase\":true}", "switch(Number.NaN) { case foo: break; }", []string{"switchOnNaN switch(Number.NaN) { case foo: break; }"}},
		{"{\"enforceForSwitchCase\":true}", "switch(Number.NaN) { default: break; }", []string{"switchOnNaN switch(Number.NaN) { default: break; }"}},
		{"{\"enforceForSwitchCase\":true}", "switch(Number.NaN) { case foo: break; default: break; }", []string{"switchOnNaN switch(Number.NaN) { case foo: break; default: break; }"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case Number.NaN: }", []string{"caseWithNaN case Number.NaN:"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case Number.NaN: break; }", []string{"caseWithNaN case Number.NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case (Number.NaN): break; }", []string{"caseWithNaN case (Number.NaN): break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: break; case Number.NaN: break; default: break; }", []string{"caseWithNaN case Number.NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: case Number.NaN: default: break; }", []string{"caseWithNaN case Number.NaN:"}},
		{"{\"enforceForSwitchCase\":true}", "switch(foo) { case bar: break; case NaN: break; case baz: break; case Number.NaN: break; }", []string{"caseWithNaN case NaN: break;", "caseWithNaN case Number.NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch(Number.NaN) { case Number.NaN: break; }", []string{"switchOnNaN switch(Number.NaN) { case Number.NaN: break; }", "caseWithNaN case Number.NaN: break;"}},
		{"{\"enforceForSwitchCase\":true}", "switch((doStuff(), NaN)) {}", []string{"switchOnNaN switch((doStuff(), NaN)) {}"}},
		{"{\"enforceForSwitchCase\":true}", "switch((doStuff(), Number.NaN)) {}", []string{"switchOnNaN switch((doStuff(), Number.NaN)) {}"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(NaN)", []string{"indexOfNaN foo.indexOf(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN)", []string{"indexOfNaN foo.lastIndexOf(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo['indexOf'](NaN)", []string{"indexOfNaN foo['indexOf'](NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo[`indexOf`](NaN)", []string{"indexOfNaN foo[`indexOf`](NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo['lastIndexOf'](NaN)", []string{"indexOfNaN foo['lastIndexOf'](NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo().indexOf(NaN)", []string{"indexOfNaN foo().indexOf(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.bar.lastIndexOf(NaN)", []string{"indexOfNaN foo.bar.lastIndexOf(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf?.(NaN)", []string{"indexOfNaN foo.indexOf?.(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo?.indexOf(NaN)", []string{"indexOfNaN foo?.indexOf(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "(foo?.indexOf)(NaN)", []string{"indexOfNaN (foo?.indexOf)(NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(Number.NaN)", []string{"indexOfNaN foo.indexOf(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Number.NaN)", []string{"indexOfNaN foo.lastIndexOf(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo['indexOf'](Number.NaN)", []string{"indexOfNaN foo['indexOf'](Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo['lastIndexOf'](Number.NaN)", []string{"indexOfNaN foo['lastIndexOf'](Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo().indexOf(Number.NaN)", []string{"indexOfNaN foo().indexOf(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.bar.lastIndexOf(Number.NaN)", []string{"indexOfNaN foo.bar.lastIndexOf(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf?.(Number.NaN)", []string{"indexOfNaN foo.indexOf?.(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo?.indexOf(Number.NaN)", []string{"indexOfNaN foo?.indexOf(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "(foo?.indexOf)(Number.NaN)", []string{"indexOfNaN (foo?.indexOf)(Number.NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf((1, NaN))", []string{"indexOfNaN foo.indexOf((1, NaN))"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf((1, Number.NaN))", []string{"indexOfNaN foo.indexOf((1, Number.NaN))"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf((1, NaN))", []string{"indexOfNaN foo.lastIndexOf((1, NaN))"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf((1, Number.NaN))", []string{"indexOfNaN foo.lastIndexOf((1, Number.NaN))"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(NaN, 1)", []string{"indexOfNaN foo.indexOf(NaN, 1)"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN, 1)", []string{"indexOfNaN foo.lastIndexOf(NaN, 1)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(NaN, b)", []string{"indexOfNaN foo.indexOf(NaN, b)"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN, b)", []string{"indexOfNaN foo.lastIndexOf(NaN, b)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf(Number.NaN, b)", []string{"indexOfNaN foo.indexOf(Number.NaN, b)"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(Number.NaN, b)", []string{"indexOfNaN foo.lastIndexOf(Number.NaN, b)"}},
		{"{\"enforceForIndexOf\":true}", "foo.lastIndexOf(NaN, NaN)", []string{"indexOfNaN foo.lastIndexOf(NaN, NaN)"}},
		{"{\"enforceForIndexOf\":true}", "foo.indexOf((1, NaN), 1)", []string{"indexOfNaN foo.indexOf((1, NaN), 1)"}},
	}
	if len(cases) != 234 {
		t.Fatalf("%d rows, and 234 were replayed", len(cases))
	}

	for _, testCase := range cases {
		// The harness writes the file trimmed, so positions are in the trimmed text.
		source := strings.TrimSpace(testCase.code) + "\n"
		run := func(options string) []string {
			var settings UseIsNaNOptions
			if err := rule.UnmarshalOptions([]byte(options), &settings); err != nil {
				t.Fatalf("decoding %s: %v", options, err)
			}
			result := rule_testing.RunTypedWithOptions(t, UseIsNaN, isNaNFile, source, settings)
			var got []string
			for _, diagnostic := range result.Diagnostics {
				got = append(got, diagnostic.Message.Id+" "+source[diagnostic.Range.Pos():diagnostic.Range.End()])
			}
			return got
		}
		got := strings.Join(run(testCase.options), " | ")
		want := strings.Join(testCase.want, " | ")
		if got != want {
			t.Errorf("%s %q: reported [%s], ESLint reports [%s]", testCase.options, testCase.code, got, want)
		}
		if strings.Contains(testCase.options, `"enforceForIndexOf":true`) {
			off := strings.Replace(testCase.options, `"enforceForIndexOf":true`, `"enforceForIndexOf":false`, 1)
			if findings := run(off); len(findings) != 0 {
				t.Errorf("%s %q: reported %v with the option off", off, testCase.code, findings)
			}
		}
	}
}
