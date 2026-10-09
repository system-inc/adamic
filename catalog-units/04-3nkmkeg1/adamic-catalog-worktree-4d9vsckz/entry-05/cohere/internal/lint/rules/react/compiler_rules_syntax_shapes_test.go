package react

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// compilerRulesLoweringEverySyntax are the rules that lower a component or hook through the shared
// representation, so one lowering defect takes all of them down together.
//
// That is how it surfaced: `async *[Symbol.asyncIterator]() {}` in an object literal, in TanStack
// Query's own tests, panicked all six with "Unhandled case in Node.Text: *ast.ComputedPropertyName",
// because the object-member arm read a method's name as text and a computed name has none
// (#zx5xvtg). Nothing in this tree had written that shape, so no fixture could have seen it.
var compilerRulesLoweringEverySyntax = []rule.Rule{
	Immutability, Purity, Refs, SetStateInEffect, SetStateInRender, StaticComponents,
}

// compilerRuleSyntaxShapes is the sweep: shapes a component or hook body can hold that the
// conformance corpus writes rarely or never, each one a place the lowering reads a node's name,
// key, label or expression. The first is the TanStack repro as it was found.
var compilerRuleSyntaxShapes = map[string]string{
	"async generator method with a computed name":  "const stream = { async *[Symbol.asyncIterator]() { yield 1; } };",
	"method with a computed name":                  "const handlers = { [props.kind]() { return 1; } };",
	"getter and setter with computed names":        "const box = { get [props.kind]() { return 1; }, set [props.kind](value) {} };",
	"getter and setter with plain names":           "const box = { get size() { return 1; }, set size(value) {} };",
	"method with a string and a numeric name":      "const table = { 'quoted name'() {}, 42() {} };",
	"generator method":                             "const sequence = { *values() { yield 1; } };",
	"computed property assignment":                 "const keyed = { [props.kind]: 1, [`${props.kind}-x`]: 2 };",
	"computed destructuring":                       "const { [props.kind]: picked = 0, ...rest } = props;",
	"computed assignment pattern":                  "let picked; ({ [props.kind]: picked } = props);",
	"class with computed and private members":      "class Local { #hidden = 1; [Symbol.iterator]() {} static [props.kind] = 2; get #secret() { return 1; } }",
	"labeled loop with break and continue":         "outer: for (const item of props.items) { for (const inner of item) { if (inner) continue outer; break outer; } }",
	"meta properties":                              "const target = typeof import.meta; function Inner() { return new.target; }",
	"tagged template and optional call":            "const text = String.raw`a${props.kind}b`; props.callback?.(1); props.object?.[props.kind]?.();",
	"jsx namespaced attribute and member tag":      "const element = <svg xlink:href={props.kind}><props.Tag a:b='c' /></svg>;",
	"async arrow, await and for await":             "const load = async () => { for await (const chunk of props.stream) { await chunk; } };",
	"enum, namespace and type-only declarations":   "type Local = { [key: string]: number }; interface Shape { [Symbol.iterator](): void }",
	"exotic numeric and bigint keys":               "const numbers = { 0x10: 1, 1n: 2, 1e3: 3 };",
	"spread, rest parameters and default patterns": "const call = (...values) => values; const [first = 1, , third] = props.items ?? [];",
}

// TestCompilerRulesLowerEverySyntaxShape runs every compiler rule over every shape, inside a component
// and inside a hook, and requires that none panics. Findings are not asserted: some shapes are
// reportable, and what each rule says about them is its own fixtures' business. The question here is
// only whether the lowering can read the code at all.
func TestCompilerRulesLowerEverySyntaxShape(t *testing.T) {
	t.Parallel()

	for shapeName, shape := range compilerRuleSyntaxShapes {
		sources := map[string]string{
			"component": "export function Component(props) {\n  " + shape + "\n  return <div>{props.kind}</div>;\n}\n",
			"hook":      "export function useShape(props) {\n  " + shape + "\n  return useMemo(() => props.kind, [props.kind]);\n}\n",
		}
		for host, source := range sources {
			for _, compilerRule := range compilerRulesLoweringEverySyntax {
				t.Run(fmt.Sprintf("%s in a %s, %s", shapeName, host, compilerRule.Name), func(t *testing.T) {
					t.Parallel()
					defer func() {
						if recovered := recover(); recovered != nil {
							t.Errorf("%s panicked on %s:\n%s\n%v", compilerRule.Name, shapeName, source, recovered)
						}
					}()
					rule_testing.RunTyped(t, compilerRule, "shape.tsx", source)
				})
			}
		}
	}
}
