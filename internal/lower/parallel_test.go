package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Inspect the checker's representation directly. Syntax-only readonly checks
// miss mapped properties and as const, so these are part of the safety boundary.
func TestParallelReadonlyCheckerRepresentation(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "types.a")
	source := `interface Explicit { readonly value: number; }
const explicit: Explicit = { value: 1 };
const mapped: Readonly<{ value: number }> = { value: 1 };
const asserted = { value: 1 } as const;
const mutable = { value: 1 };
const array: readonly number[] = [1];
const namedArray: ReadonlyArray<number> = [1];
const tuple = [1, 'a'] as const;
const primitive: number | boolean | string | undefined | null = null;
const nested: Readonly<{ field: Readonly<{ values: ReadonlyArray<string> }> }> = { field: { values: ['x'] } };
const bad: Readonly<{ field: Readonly<{ values: string[] }> }> = { field: { values: ['x'] } };
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	c, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := &lowering{program: program, checker: c}
	variables := map[string]*ast.Node{}
	program.Files()[0].AsNode().ForEachChild(func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableStatement {
			for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				variables[declaration.Name().Text()] = declaration.Name()
			}
		}
		return false
	})
	for _, name := range []string{"explicit", "mapped", "asserted", "mutable"} {
		properties := c.GetPropertiesOfType(c.GetTypeAtLocation(variables[name]))
		if len(properties) != 1 {
			t.Fatalf("%s: properties %v", name, properties)
		}
		if readonly := c.IsReadonlySymbol(properties[0]); readonly != (name != "mutable") {
			t.Errorf("%s readonly=%v", name, readonly)
		}
	}
	for _, name := range []string{"array", "namedArray"} {
		if !l.isLibraryType(c.GetTypeAtLocation(variables[name]), "ReadonlyArray") {
			t.Errorf("%s is not represented by ReadonlyArray", name)
		}
	}
	tuple := c.GetTypeAtLocation(variables["tuple"])
	if !checker.IsTupleType(tuple) || !tuple.TargetTupleType().IsReadonly() {
		t.Fatal("as const tuple lost readonly target")
	}
	for _, name := range []string{"explicit", "mapped", "asserted", "array", "namedArray", "tuple", "primitive", "nested"} {
		if why := l.shareable(c.GetTypeOfSymbol(l.symbol(variables[name])), name, variables[name]); why != "" {
			t.Errorf("%s: %s", name, why)
		}
	}
	if why := l.shareable(c.GetTypeAtLocation(variables["bad"]), "bad", variables["bad"]); why != "bad.field.values is a mutable string[]" {
		t.Errorf("nested path: %q", why)
	}
}

func TestParallelGlobalMarkingRoots(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `import { parallelMap } from 'adamic';
const lookup: ReadonlyMap<string, number> = new Map([['a', 1]]);
const pick = (item: string): number => lookup.get(item) ?? 0;
function summarize(item: string): number { return pick(item); }
const items: readonly string[] = ['a'];
parallelMap(items, summarize);
`)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	walk(program.Main, func(node any) bool {
		if mapped, ok := node.(ir.ParallelMap); ok {
			found = true
			if mapped.Result != ir.Number || len(mapped.Shared) != 2 {
				t.Errorf("result=%v roots=%v", mapped.Result, mapped.Shared)
			}
			for _, root := range mapped.Shared {
				if read, ok := root.(ir.Read); !ok || read.Checked {
					t.Errorf("marking must not introduce a TDZ observation: %v", root)
				}
			}
		}
		return true
	})
	if !found {
		t.Fatal("no ParallelMap")
	}
}
