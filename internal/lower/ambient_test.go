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

func TestDefinedModuleAndTypeDeclarationsCompile(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`interface Value { readonly value: number; }
        type Amount = number;
        export function next(value: Value): Amount { return value.value + 1; }
        console.log("defined");`,
		`declare namespace Types { interface Value { readonly value: number; } type Amount = number; }
        const value: Types.Value = { value: 1 };
        console.log(value.value.toFixed(0));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

// The checker rejects the older identifier-module spelling with TS1540 before
// Lower runs. Its AST still has the same refusal if a checked producer supplies it.
func TestAmbientIdentifierModuleIsRefused(t *testing.T) {
	t.Parallel()
	graph, statements := namespaceGraphForTest(t, `declare namespace Missing { const value: number; }`)
	declaration := statements[0]
	declaration.AsModuleDeclaration().Keyword = ast.KindModuleKeyword
	err := graph.lowering.ambientRefusal(declaration)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "an ambient namespace or module holding values" || !strings.HasSuffix(refused.Where, "graph.a:1:1") {
		t.Fatalf("want outermost ambient module refusal, got %v", err)
	}
}

func TestDefinedImportedModuleCompiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for name, source := range map[string]string{
		"value.a": `export interface Value { readonly value: number; }
            export type Amount = number;
            export function next(value: Value): Amount { return value.value + 1; }`,
		"main.a": `import { next, type Value, type Amount } from './value.a';
            const value: Value = { value: 1 };
            const amount: Amount = next(value);
            console.log(amount.toFixed(0));`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Lower(context.Background(), program); err != nil {
		t.Fatal(err)
	}
}
