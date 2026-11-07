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
func TestParallelReadonlyCheckerRepresentation(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "types.a")
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
	if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
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
			test.Fatalf("%s: properties %v", name, properties)
		}
		if readonly := c.IsReadonlySymbol(properties[0]); readonly != (name != "mutable") {
			test.Errorf("%s readonly=%v", name, readonly)
		}
	}
	for _, name := range []string{"array", "namedArray"} {
		if !l.isLibraryType(c.GetTypeAtLocation(variables[name]), "ReadonlyArray") {
			test.Errorf("%s is not represented by ReadonlyArray", name)
		}
	}
	tuple := c.GetTypeAtLocation(variables["tuple"])
	if !checker.IsTupleType(tuple) || !tuple.TargetTupleType().IsReadonly() {
		test.Fatal("as const tuple lost readonly target")
	}
	for _, name := range []string{
		"explicit",
		"mapped",
		"asserted",
		"array",
		"namedArray",
		"tuple",
		"primitive",
		"nested",
	} {
		if why := l.shareable(c.GetTypeOfSymbol(l.symbol(variables[name])), name, variables[name]); why != "" {
			test.Errorf("%s: %s", name, why)
		}
	}
	if why := l.shareable(
		c.GetTypeAtLocation(variables["bad"]),
		"bad",
		variables["bad"],
	); why != "bad.field.values is a mutable string[]" {
		test.Errorf("nested path: %q", why)
	}
}

func TestParallelGlobalMarkingRoots(test *testing.T) {
	test.Parallel()
	program, lowerSourceError := lowerSource(test, `import { parallelMap } from 'adamic';
const lookup: ReadonlyMap<string, number> = new Map([['a', 1]]);
const pick = (item: string): number => lookup.get(item) ?? 0;
function summarize(item: string): number { return pick(item); }
const items: readonly string[] = ['a'];
parallelMap(items, summarize);
`)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	found := false
	walk(program.Main, func(node any) bool {
		if mapped, ok := node.(ir.ParallelMap); ok {
			found = true
			if mapped.Result != ir.Number || len(mapped.Shared) != 2 {
				test.Errorf("result=%v roots=%v", mapped.Result, mapped.Shared)
			}
			for _, root := range mapped.Shared {
				if read, ok := root.(ir.Read); !ok || read.Checked {
					test.Errorf("marking must not introduce a TDZ observation: %v", root)
				}
			}
		}
		return true
	})
	if !found {
		test.Fatal("no ParallelMap")
	}
}
