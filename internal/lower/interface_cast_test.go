package lower

import (
	"strings"
	"testing"
)

const interfaceCastSource = `interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
function identifier(node: Node): Identifier { return node as Identifier; }
`

// Malformed constructions are admitted; their fields are checked when read. Each row but "in .a" lowers
// as .ts, where the cast stays a checked view; "in .a" pins that .a refuses the downcast, whose checks
// no proof erases (#m8dvvmf).
func TestDefaultTaggedInterfaceAdmission(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct{ name, source, reason string }{
		{
			"complete",
			`const node: Node = { kind: 'identifier', pos: 0, name: 'ok' } as Identifier; console.log(identifier(node).name);`,
			"",
		},
		{
			"missing payload",
			`const node: Node = { kind: 'identifier', pos: 0 }; console.log(identifier(node).name ?? 'missing');`,
			"",
		},
		{
			"wrong payload",
			`const raw = { kind: 'identifier' as const, pos: 0, name: 42 }; const node: Node = raw; console.log(identifier(node).name);`,
			"",
		},
		{
			"broad tag",
			`function make(kind: Node['kind']): Node { return { kind, pos: 0 }; } console.log(identifier(make('identifier')).name);`,
			"",
		},
		{
			"alias write",
			`const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const alias = raw; alias.name = 'changed'; console.log(identifier(raw).name);`,
			"",
		},
		{
			"unused bad factory",
			`function bad(): Node { return { kind: 'identifier', pos: 0 }; } console.log(identifier({ kind: 'identifier', pos: 0, name: 'ok' } as Identifier).name);`,
			"",
		},
		{"different tag", `const node: Node = { kind: 'number', pos: 0 }; console.log(identifier(node).name);`, ""},
		{
			"dynamic complete",
			`function make(kind: Node['kind']): Node { const raw = { kind, pos: 0, name: 'ok' }; return raw; } console.log(identifier(make('identifier')).name);`,
			"",
		},
		{
			"optional target",
			`interface Optional extends Node { readonly kind: 'identifier'; readonly name?: string; } function opt(node: Node): Optional { return node as Optional; } console.log(opt({ kind: 'identifier', pos: 0 }).kind);`,
			"",
		},
		{
			"optional target read",
			`interface Optional extends Node { readonly kind: 'identifier'; readonly name?: string; } function opt(node: Node): Optional { return node as Optional; } console.log(opt({kind: 'identifier', pos: 0}).name ?? 'missing');`,
			"",
		},
		{
			"spread",
			`const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const node: Node = { ...raw }; console.log(identifier(node).name);`,
			"",
		},
		{
			"reflection alias",
			`const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const O = Object; O.assign(raw, { name: 'changed' }); console.log(identifier(raw).name);`,
			"Object as a value",
		},
		{
			"generic",
			`function make<T extends string>(name: T): Node { const raw = { kind: 'identifier' as const, pos: 0, name }; return raw; } console.log(identifier(make('ok')).name);`,
			"",
		},
		{
			"staged class",
			`class Incomplete { readonly kind = 'identifier'; readonly pos = 0; } console.log(identifier(new Incomplete()).name);`,
			"",
		},
		{
			"forged payload",
			`const name = 42 as unknown as string; console.log(identifier({ kind: 'identifier', pos: 0, name } as Identifier).name);`,
			"cast the runtime can't check",
		},
		{
			"in .a",
			`console.log(identifier({ kind: 'identifier', pos: 0, name: 'ok' } as Identifier).name);`,
			"a downcast from Node to Identifier in .a whose checks no proof erases",
		},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			switch testCase.name {
			case "complete", "alias write", "unused bad factory", "dynamic complete", "spread", "generic", "optional target", "optional target read":
				lowersAndAgreesWithNodeAs(test, interfaceCastSource+testCase.source, ".ts")
				return
			}
			extension := ".ts"
			if testCase.name == "in .a" {
				extension = ".a"
			}
			_, lowerSourceAsError := lowerSourceAs(test, interfaceCastSource+testCase.source, extension)
			if testCase.reason == "" {
				if lowerSourceAsError != nil {
					test.Fatal(lowerSourceAsError)
				}
			} else if lowerSourceAsError == nil || !strings.Contains(lowerSourceAsError.Error(), testCase.reason) {
				test.Fatalf("want refusal containing %q, got %v", testCase.reason, lowerSourceAsError)
			}
		})
	}
}
