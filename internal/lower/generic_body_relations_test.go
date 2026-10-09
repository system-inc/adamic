package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func genericBodyUnsound(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("testdata/generic_body_relations/refused.a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestGenericBodyRelationsRefuse(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, slot string }{
		{"initializer", genericBodyUnsound(t), `T["value"]`},
		{"assignment", `function test<T extends {readonly value:number}>(source:T):void { let value:T['value']=source.value; value=2; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"return", `function test<T extends {readonly value:number}>(source:T):T['value'] { return 2; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"field initializer", `class Box<T extends {readonly value:number}> { readonly value:T['value']=2; } const box=new Box<{readonly value:1}>(); console.log(String(box.value));`, `T["value"]`},
		{"different binder", `function test<T extends {readonly value:number}, U extends {readonly value:number}>(source:U):void { const value:T['value']=source.value; } test<{readonly value:1}, {readonly value:2}>({value:2});`, `T["value"]`},
		{"nested", `function test<T extends {readonly inner:{readonly value:number}}>(source:T):void { const value:T['inner']['value']=2; } test<{readonly inner:{readonly value:1}}>({inner:{value:1}});`, `T["inner"]["value"]`},
		{"key parameter", `function test<T extends {readonly value:number}, K extends 'value'>(source:T, key:K):void { const value:T[K]=2; } test<{readonly value:1}, 'value'>({value:1},'value');`, `T[K]`},
		{"array initializer", `function test<T extends {readonly value:number}>(source:T):void { const values:T['value'][]=[2]; } test<{readonly value:1}>({value:1});`, `T["value"][]`},
		{"destructuring initializer", `function test<T extends {readonly value:number}>(source:T):void { const [value]:[T['value']]=[2]; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"destructuring assignment", `function test<T extends {readonly value:number}>(source:T):void { let value:T['value']=source.value; [value]=[2]; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"object initializer", `function test<T extends {readonly value:number}>(source:T):void { const value:{item:T['value']}={item:2}; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"index signature", `function test<T extends {readonly value:number}>(source:T):void { const values:{[key:string]:T['value']}={first:2}; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"mapped keys", `type Keys<T> = { [K in keyof T]:number }; function test<T extends {readonly value:number}>(source:T):void { const values:Keys<T>=null as any; } test<{readonly value:1}>({value:1});`, `Keys<T>`},
		{"element assignment", `function test<T extends {readonly value:number}>(values:T['value'][]):void { values[0]=2; } test<{readonly value:1}>([]);`, `T["value"]`},
		{"increment", `function test<T extends {readonly value:number}>(source:T):void { let value:T['value']=source.value; value++; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"compound assignment", `function test<T extends {readonly value:number}>(source:T):void { let value:T['value']=source.value; value+=2; } test<{readonly value:1}>({value:1});`, `T["value"]`},
		{"alias", `type Value<T extends {readonly value:number}>=T['value']; function test<T extends {readonly value:number}>(source:T):void { const value:Value<T>=2; } test<{readonly value:1}>({value:1});`, `Value<T>`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if !errors.As(err, &refusal) {
				t.Fatalf("want generic body refusal, got %v", err)
			}
			if probe.name == "initializer" && !strings.Contains(refusal.What, "from source 2") {
				t.Fatalf("missing own literal type: %v", err)
			}
			for _, part := range []string{probe.slot, "from source", "checker constraint: T extends", "adamic/generic-body-relations"} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("missing %q: %v", part, err)
				}
			}
		})
	}
}

// Bypass only the up-front walk. The real instantiator and witness must reject
// the very same checker-accepted program at its resolved initializer relation.
func TestGenericBodyRelationsWitness(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "witness.a")
	if err := os.WriteFile(path, []byte(genericBodyUnsound(t)), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	var declaration, call *ast.Node
	for _, node := range program.Files()[0].Statements.Nodes {
		if node.Kind == ast.KindFunctionDeclaration {
			declaration = node
		}
		if node.Kind == ast.KindExpressionStatement {
			call = node.AsExpressionStatement().Expression
		}
	}
	resolved := checked.GetResolvedSignature(call)
	l := &lowering{program: program, checker: checked}
	l.typeMapper = genericSignatureMapper(resolved)
	if l.typeMapper == nil {
		t.Fatal("resolved signature has no checker mapper")
	}
	err = l.genericBodyWitness(declaration)
	if err == nil || !strings.Contains(err.Error(), "internal compiler error: generic body initializer witness") || !strings.Contains(err.Error(), "slot 1") {
		t.Fatalf("want resolved initializer witness 2 into 1, got %v", err)
	}
	t.Log(err)
}

// The ruled body relation requires the source binder itself even for a readonly
// bound. This moved from the invariance-only positive table.
func TestGenericBodyRelationsReadonlyBound(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, pets+`function mix<Pack extends readonly Animal[], Narrow extends Pack>(pack: Pack, narrow: Narrow): number {
 const slot: Pack = narrow;
 return pack.length + slot.length;
 }`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "slot Pack from source Narrow") {
		t.Fatalf("want dependent-binder refusal, got %v", err)
	}
}

func TestGenericBodyRelationsIndexedReturnMapper(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function test<T extends {readonly value:number}>(source:T['value']):T['value'] { const value:T['value']=source; return value; } console.log(String(test<{readonly value:1}>(1)));`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenericBodyRelationsThisRead(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Box<T> { constructor(readonly value:T) {} same():this { return this; } } const box=new Box<number>(1); console.log(String(box.same().value));`)
	var notYet *NotYet
	if err != nil && (!errors.As(err, &notYet) || !strings.Contains(err.Error(), "function returning this")) {
		t.Fatal(err)
	}
}
