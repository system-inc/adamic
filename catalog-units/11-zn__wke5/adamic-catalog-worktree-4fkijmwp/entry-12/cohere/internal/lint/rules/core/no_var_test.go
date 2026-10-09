package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const varFile = "/repository/source/Thing.ts"

func TestNoVarFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a statement", "var value = 1;\nexport const Use = value;\n"},
		{"inside a function", "export function run() {\n    var value = 1;\n    return value;\n}\n"},
		{"inside a block", "export function run(flag: boolean) {\n    if(flag) {\n        var value = 1;\n        return value;\n    }\n    return 0;\n}\n"},
		// The three loop forms. Each hangs its declaration list off the loop rather than off a
		// statement, so a listener on variable statements alone catches none of them, and a var in
		// a loop is the shape with the sharpest consequence: one binding shared by every iteration.
		{"a for initializer", "export function run() {\n    for(var index = 0; index < 3; index++) {}\n}\n"},
		{"a for-of binding", "export function run(list: number[]) {\n    for(var item of list) { void item; }\n}\n"},
		{"a for-in binding", "export function run(record: object) {\n    for(var key in record) { void key; }\n}\n"},
		// The nearest shapes to the declare global exemption, each of which ESLint 10.8.1 still reports
		// (probed with lintText): only a var directly in the global block is exempt (#hzhs9f8).
		{"a var in a declare module block", "declare module 'thing' {\n    var value: number;\n}\n"},
		{"a var in a declare namespace block", "declare namespace Space {\n    var value: number;\n}\nexport {};\n"},
		{"a var in a namespace nested inside declare global", "declare global {\n    namespace Space {\n        var value: number;\n    }\n}\nexport {};\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoVar, varFile, testCase.sourceText), "unexpectedVar")
		})
	}
}

func TestNoVarStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"const", "const value = 1;\nexport const Use = value;\n"},
		{"let", "let value = 1;\nexport const Use = value;\n"},
		{"a let for initializer", "export function run() {\n    for(let index = 0; index < 3; index++) {}\n}\n"},
		{"a const for-of binding", "export function run(list: number[]) {\n    for(const item of list) { void item; }\n}\n"},
		// A loop that assigns to an existing binding declares nothing, so there is no declaration
		// list to judge. This is the boundary between "the initializer declares" and "the
		// initializer assigns", and a check reading the initializer without testing its kind would
		// crash or misfire here.
		{"a for initializer that assigns rather than declares", "export function run() {\n    let index = 0;\n    for(index = 0; index < 3; index++) {}\n}\n"},
		{"a for with no initializer", "export function run(flag: boolean) {\n    for(; flag; ) { break; }\n}\n"},
		// An ambient declaration describes a runtime this file does not control, so the keyword is
		// not a scoping choice the author made.
		{"declare var", "declare var globalThing: number;\nexport const Use = globalThing;\n"},
		// Base's GraphQlMetadataStorage.ts: only a var in declare global types a property of
		// globalThis, and ESLint skips exactly this shape (#hzhs9f8).
		{"a var directly in declare global", "declare global {\n    var BaseGraphQlMetadataStorage: unknown;\n}\nexport {};\n"},
		// `using` carries its own flag and is block-scoped, so reading var as "no let and no const"
		// must not sweep it up.
		{"a using declaration", "export function run(resource: { [Symbol.dispose](): void }) {\n    using held = resource;\n    void held;\n}\n"},
		{"the word var in a string", "export const Message = 'var value = 1';\nexport const Use = Message;\n"},
		{"no declarations", "export function run() {\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoVar, varFile, testCase.sourceText))
		})
	}
}
