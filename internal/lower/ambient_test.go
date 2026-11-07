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

func TestDefinedModuleAndTypeDeclarationsCompile(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`interface Value { readonly value: number; }
        type Amount = number;
        export function next(value: Value): Amount { return value.value + 1; }
        console.log("defined");`,
		`declare namespace Types { interface Value { readonly value: number; } type Amount = number; }
        const value: Types.Value = { value: 1 };
        console.log(value.value.toFixed(0));`,
	} {
		if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
	}
}

// The checker rejects the older identifier-module spelling with TS1540 before
// Lower runs. Its AST still has the same refusal if a checked producer supplies it.
func TestAmbientIdentifierModuleIsRefused(test *testing.T) {
	test.Parallel()
	graph, statements := namespaceGraphForTest(test, `declare namespace Missing { const value: number; }`)
	declaration := statements[0]
	declaration.AsModuleDeclaration().Keyword = ast.KindModuleKeyword
	ambientRefusalError := graph.lowering.ambientRefusal(declaration)
	var refused *Refused
	if !errors.As(ambientRefusalError, &refused) ||
		refused.What != "an ambient namespace or module holding values" ||
		!strings.HasSuffix(refused.Where, "graph.a:1:1") {
		test.Fatalf("want outermost ambient module refusal, got %v", ambientRefusalError)
	}
}

func TestDefinedImportedModuleCompiles(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	for name, source := range map[string]string{
		"value.a": `export interface Value { readonly value: number; }
            export type Amount = number;
            export function next(value: Value): Amount { return value.value + 1; }`,
		"main.a": `import { next, type Value, type Amount } from './value.a';
            const value: Value = { value: 1 };
            const amount: Amount = next(value);
            console.log(amount.toFixed(0));`,
	} {
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, loadError := load.Load([]string{filepath.Join(directory, "main.a")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	if _, lowerError := Lower(context.Background(), program); lowerError != nil {
		test.Fatal(lowerError)
	}
}
