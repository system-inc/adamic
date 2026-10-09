package lower

import (
	"strings"
	"testing"
)

const interfaceCastSource = `interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
function identifier(node: Node): Identifier { return node as Identifier; }
`

// Malformed constructions are admitted; their fields are checked when read.
func TestDefaultTaggedInterfaceAdmission(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, reason string }{
		{"complete", `const node: Node = { kind: 'identifier', pos: 0, name: 'ok' } as Identifier; console.log(identifier(node).name);`, ""},
		{"missing payload", `const node: Node = { kind: 'identifier', pos: 0 }; console.log(identifier(node).name ?? 'missing');`, ""},
		{"wrong payload", `const raw = { kind: 'identifier' as const, pos: 0, name: 42 }; const node: Node = raw; console.log(identifier(node).name);`, ""},
		{"broad tag", `function make(kind: Node['kind']): Node { return { kind, pos: 0 }; } console.log(identifier(make('identifier')).name);`, ""},
		{"alias write", `const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const alias = raw; alias.name = 'changed'; console.log(identifier(raw).name);`, ""},
		{"unused bad factory", `function bad(): Node { return { kind: 'identifier', pos: 0 }; } console.log(identifier({ kind: 'identifier', pos: 0, name: 'ok' } as Identifier).name);`, ""},
		{"different tag", `const node: Node = { kind: 'number', pos: 0 }; console.log(identifier(node).name);`, ""},
		{"dynamic complete", `function make(kind: Node['kind']): Node { const raw = { kind, pos: 0, name: 'ok' }; return raw; } console.log(identifier(make('identifier')).name);`, ""},
		{"optional target", `interface Optional extends Node { readonly kind: 'identifier'; readonly name?: string; } function opt(node: Node): Optional { return node as Optional; }`, ""},
		{"optional target read", `interface Optional extends Node { readonly kind: 'identifier'; readonly name?: string; } function opt(node: Node): Optional { return node as Optional; } console.log(opt({kind: 'identifier', pos: 0}).name ?? 'missing');`, ""},
		{"optional nullable target read", `interface Optional extends Node { readonly kind: 'identifier'; readonly name?: string | null; } function opt(node: Node): Optional { return node as Optional; } const raw = {kind: 'identifier' as const, pos: 0, name: null}; const source: Node = raw; console.log(String(opt(source).name));`, "a checked field alias requiring an optional, accessor, or representation conversion"},
		{"spread", `const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const node: Node = { ...raw }; console.log(identifier(node).name);`, ""},
		{"reflection alias", `const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const O = Object; O.assign(raw, { name: 'changed' }); console.log(identifier(raw).name);`, "Object as a value"},
		{"generic", `function make<T extends string>(name: T): Node { const raw = { kind: 'identifier' as const, pos: 0, name }; return raw; } console.log(identifier(make('ok')).name);`, ""},
		{"staged class", `class Incomplete { readonly kind = 'identifier'; readonly pos = 0; } console.log(identifier(new Incomplete()).name);`, ""},
		{"forged payload", `const name = 42 as unknown as string; console.log(identifier({ kind: 'identifier', pos: 0, name } as Identifier).name);`, "cast the runtime can't check"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, interfaceCastSource+test.source)
			if test.reason == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want refusal containing %q, got %v", test.reason, err)
			}
		})
	}
}

// Default admission has no environment switch.
func TestDefaultTaggedInterfaceNeedsNoFlag(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, interfaceCastSource+`console.log(identifier({ kind: 'identifier', pos: 0, name: 'ok' } as Identifier).name);`)
	if err != nil {
		t.Fatalf("default admission: %v", err)
	}
}
