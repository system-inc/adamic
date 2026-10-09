package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const unsafeFunctionFile = "/repository/source/Thing.ts"

func TestNoUnsafeFunctionTypeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a variable annotation", "export let handler: Function;\n"},
		{"a parameter annotation", "export function run(callback: Function) {\n    return callback;\n}\n"},
		{"a return annotation", "export function make(): Function {\n    return () => {};\n}\n"},
		{"a property annotation", "export interface Thing {\n    handler: Function;\n}\n"},
		// A heritage clause is a different node kind from a type reference, so a rule listening only
		// for annotations would pass every case above and miss both of these.
		{"an interface extending it", "export interface Thing extends Function {}\n"},
		{"a class implementing it", "export class Thing implements Function {}\n"},
		// Nested inside another type, which a listener keyed to declarations rather than to type
		// references would not reach.
		{"inside an array type", "export let handlers: Function[];\n"},
		{"inside a union", "export let handler: Function | undefined;\n"},
		{"as a type argument", "export let handlers: Array<Function>;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUnsafeFunctionType, unsafeFunctionFile, testCase.sourceText),
				"bannedFunctionType")
		})
	}
}

func TestNoUnsafeFunctionTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The decision is "is this the global Function type", and the way to get it wrong is to match
		// the identifier anywhere it appears. Each of these contains the word in a position that is
		// not the banned type.
		{"an explicit signature", "export let handler: () => void;\n"},
		{"a parameterised signature", "export let handler: (input: string) => number;\n"},
		{"a name that merely contains it", "export let handler: FunctionLike;\n"},
		{"a property named function-ish", "export const Settings = { Function: true };\n"},
		{"a string containing the word", "export const Message = 'Function';\n"},
		// `Function` as a value rather than a type. The rule is about type positions, and a listener
		// keyed to identifiers rather than to type nodes would fire on all of these.
		{"a call to the constructor", "export const Made = Function('return 1');\n"},
		{"a static method call", "export const Bound = Function.prototype.bind;\n"},
		{"an instanceof check", "export const IsFunction = (value: unknown) => value instanceof Function;\n"},
		// A qualified name. `Foo.Function` is a member of some namespace and has nothing to do with
		// the global type, and the TypeName of that reference is not an identifier. Without the
		// identifier-kind check the rule reads the trailing name and fires, which is a false positive
		// on a type the author owns. This fixture is the only thing pinning that check: mutation
		// testing found it surviving with the rest of the pair in place.
		{"a qualified name ending in Function", "export let handler: Foo.Function;\n"},
		{"a namespaced import member", "import * as ns from './ns.ts';\nexport let handler: ns.Function;\n"},
		// The shadowing guard. A file that declares its own Function type is skipped entirely, so the
		// rule cannot flag a type the author defined.
		{"a locally declared interface", "interface Function {\n    tag: string;\n}\nexport let handler: Function;\n"},
		{"a locally declared type alias", "type Function = (input: string) => void;\nexport let handler: Function;\n"},
		{"a locally declared class", "class Function {\n    tag = 'x';\n}\nexport let handler: Function;\n"},
		// An enum shadows too, measured with the checker rather than assumed: an enum declares a
		// value and a TYPE of the same name, so the annotation resolves to the enum rather than to
		// the global. The shadow walk counted only interface, type alias and class until it was
		// lifted onto the shelf, so this was a false positive on a name the author owns.
		{"a locally declared enum", "enum Function {\n    A,\n}\nexport let handler: Function;\n"},
		// oxc's own passing test case, and it was a false positive here until oxc's fixtures were
		// read. A top-level scan misses a declaration nested inside a block, so the shadow check has
		// to walk the whole file.
		{"a block-scoped type alias", "{\n    type Function = () => void;\n    let value: Function;\n}\n"},
		{"a generic signature", "export let value: <T>(input: T) => T;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeFunctionType, unsafeFunctionFile, testCase.sourceText))
		})
	}
}
