package lower

import (
	"strings"
	"testing"
)

const interfaceCastSource = `interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
function identifier(node: Node): Identifier { return node as Identifier; }
`

// Not parallel: the experimental flag is process-wide. Setenv restores it before parallel tests.
func TestInterfaceConstructionProof(t *testing.T) {
	t.Setenv("ADAMIC_INTERFACE_DOWNCASTS", "1")
	for _, test := range []struct{ name, source, reason string }{
		{"complete", `const node: Node = { kind: 'identifier', pos: 0, name: 'ok' } as Identifier; console.log(identifier(node).name);`, ""},
		{"missing payload", `const node: Node = { kind: 'identifier', pos: 0 }; console.log(identifier(node).name ?? 'missing');`, "lacks required field name"},
		{"wrong payload", `const raw = { kind: 'identifier' as const, pos: 0, name: 42 }; const node: Node = raw; console.log(identifier(node).name);`, "incompatible field name"},
		{"broad tag", `function make(kind: Node['kind']): Node { return { kind, pos: 0 }; } console.log(identifier(make('identifier')).name);`, "lacks required field name"},
		{"alias write", `const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const alias = raw; alias.name = 'changed'; console.log(identifier(raw).name);`, "write may invalidate"},
		{"unused bad factory", `function bad(): Node { return { kind: 'identifier', pos: 0 }; } console.log(identifier({ kind: 'identifier', pos: 0, name: 'ok' } as Identifier).name);`, "lacks required field name"},
		{"different tag", `const node: Node = { kind: 'number', pos: 0 }; console.log(identifier(node).name);`, ""},
		{"dynamic complete", `function make(kind: Node['kind']): Node { const raw = { kind, pos: 0, name: 'ok' }; return raw; } console.log(identifier(make('identifier')).name);`, ""},
		{"optional target", `interface Optional extends Node { readonly kind: 'identifier'; readonly name?: string; } function opt(node: Node): Optional { return node as Optional; }`, "assertion could forge"},
		{"spread", `const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const node: Node = { ...raw }; console.log(identifier(node).name);`, "unproven field names"},
		{"reflection alias", `const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' }; const O = Object; O.assign(raw, { name: 'changed' }); console.log(identifier(raw).name);`, "reflection or host"},
		{"generic", `function make<T extends string>(name: T): Node { const raw = { kind: 'identifier' as const, pos: 0, name }; return raw; } console.log(identifier(make('ok')).name);`, "generic or bodyless"},
		{"staged class", `class Incomplete { readonly kind = 'identifier'; readonly pos = 0; } console.log(identifier(new Incomplete()).name);`, "opaque, staged"},
		{"forged payload", `const name = 42 as unknown as string; console.log(identifier({ kind: 'identifier', pos: 0, name } as Identifier).name);`, "assertion could forge"},
	} {
		t.Run(test.name, func(t *testing.T) {
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

// Not parallel: this proves the environment flag is default-off without sharing it.
func TestInterfaceConstructionFlagOff(t *testing.T) {
	t.Setenv("ADAMIC_INTERFACE_DOWNCASTS", "")
	_, err := lowerSource(t, interfaceCastSource+`console.log(identifier({ kind: 'identifier', pos: 0, name: 'ok' } as Identifier).name);`)
	if err == nil || !strings.Contains(err.Error(), "cast the runtime can't check") {
		t.Fatalf("flag off must refuse, got %v", err)
	}
}
