package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestGlobalLoadsRetainBindingProvenance(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name       string
		prefix     string
		parameters string
		kind       GlobalBindingKind
		module     string
		imported   string
		loads      int
	}{
		{"named", `import {value} from './library';`, "", GlobalBindingKindImportSpecifier, "./library", "value", 1},
		{"renamed", `import {exported as value} from './library';`, "", GlobalBindingKindImportSpecifier, "./library", "exported", 1},
		{"default", `import value from './library';`, "", GlobalBindingKindImportDefault, "./library", "", 1},
		{"namespace", `import * as value from './library';`, "", GlobalBindingKindImportNamespace, "./library", "", 1},
		{"module local", `const value = 1;`, "", GlobalBindingKindModuleLocal, "", "", 1},
		{"global", "", "", GlobalBindingKindGlobal, "", "", 1},
		{"shadowed import", `import {value} from './library';`, "value: number", GlobalBindingKindGlobal, "", "", 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			visited, loads := 0, 0
			probe := rule.Rule{Name: "global-load-provenance", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(node *ast.Node) {
						if node.Name() == nil || node.Name().Text() != "Component" {
							return
						}
						function := Lower(node, ctx.TypeChecker)
						if function == nil {
							t.Fatal("function did not lower")
						}
						visited++
						for _, instruction := range function.Instructions {
							load, ok := instruction.Value.(*LoadGlobal)
							if !ok || load.Name != "value" {
								continue
							}
							loads++
							if load.BindingKind != testCase.kind || load.Source != testCase.module || load.Imported != testCase.imported {
								t.Errorf("load=%+v, want kind=%v source=%q imported=%q", load, testCase.kind, testCase.module, testCase.imported)
							}
						}
					})
				}}
			}}
			rule_testing.RunTypedFiles(t, probe, map[string]string{
				"/library.ts": `export const value = 1; export const exported = 2; export default value;`,
				"/fixture.ts": testCase.prefix + " function Component(" + testCase.parameters + ") {return value;}",
			}, "/fixture.ts")
			if visited != 1 || loads != testCase.loads {
				t.Fatalf("visited=%d loads=%d, want 1 and %d", visited, loads, testCase.loads)
			}
		})
	}
}
