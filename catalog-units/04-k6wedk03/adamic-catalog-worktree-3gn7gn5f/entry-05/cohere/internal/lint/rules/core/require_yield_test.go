package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const requireYieldFile = "/repository/source/Thing.ts"

func TestRequireYieldFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The four shapes a generator takes, enumerated before the listener was written rather
		// than after. An arrow function cannot be a generator, so there is no fifth.
		{"a function declaration", "export function* generate() {\n    return 1;\n}\n"},
		{"a function expression", "export const generate = function* () {\n    return 1;\n};\n"},
		{"a class method", "export class Thing {\n    *generate() {\n        return 1;\n    }\n}\n"},
		{"an object method shorthand", "export const holder = {\n    *generate() {\n        return 1;\n    },\n};\n"},
		{"a static method", "export class Thing {\n    static *generate() {\n        return 1;\n    }\n}\n"},
		// The subtlety. A yield inside a nested function belongs to that function, so it must not
		// satisfy the outer generator. A walk that descended would call this one satisfied.
		{
			"a yield belonging to a nested generator",
			"export function* generate() {\n    const inner = function* () {\n        yield 1;\n    };\n    return inner;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, RequireYield, requireYieldFile, testCase.sourceText),
				"missingYield")
		})
	}
}

func TestRequireYieldStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a yield", "export function* generate() {\n    yield 1;\n}\n"},
		{"a delegating yield", "declare function other(): Generator<number>;\nexport function* generate() {\n    yield* other();\n}\n"},
		{"a yield with no argument", "export function* generate() {\n    yield;\n}\n"},
		// The yield is this function's even though it sits several statements deep, so the walk
		// has to descend through blocks while stopping at function boundaries.
		{"a yield inside a nested block", "export function* generate(flag: boolean) {\n    if(flag) {\n        for(let index = 0; index < 3; index++) {\n            yield index;\n        }\n    }\n}\n"},
		{"a yield in a class method", "export class Thing {\n    *generate() {\n        yield 1;\n    }\n}\n"},
		// A plain function is not a generator, so its lack of yield means nothing. This is the
		// boundary: the asterisk is the whole trigger.
		{"a plain function declaration", "export function generate() {\n    return 1;\n}\n"},
		{"a plain function expression", "export const generate = function () {\n    return 1;\n};\n"},
		{"a plain method", "export class Thing {\n    generate() {\n        return 1;\n    }\n}\n"},
		{"an arrow function", "export const generate = () => 1;\n"},
		// An empty body is exempt, matching ESLint's own `node.body.body.length > 0` guard. I had
		// this in the firing set until the real tree produced the counterexample: the tree holds
		// `function*() {}.constructor`, which reaches the GeneratorFunction constructor and is
		// never called, so an empty body is exactly right there.
		{"an empty generator body", "export function* generate() {}\n"},
		{"an empty generator method", "export class Thing {\n    *generate() {}\n}\n"},
		// An overload signature and an ambient declaration have no body to search, so there is
		// nothing to judge and reading a nil body must not report.
		{"an ambient generator declaration", "declare function generate(): Generator<number>;\nexport const Use = generate;\n"},
		{"no functions at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, RequireYield, requireYieldFile, testCase.sourceText))
		})
	}
}
