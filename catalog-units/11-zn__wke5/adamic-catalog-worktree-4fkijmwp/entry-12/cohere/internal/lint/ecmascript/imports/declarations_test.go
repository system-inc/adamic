package imports

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func parseDeclarationFile(t *testing.T, sourceText string) *ast.SourceFile {
	t.Helper()
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/repository/source/shims.d.ts",
		PathKey:  "/repository/source/shims.d.ts",
	}, sourceText, core.ScriptKindTS)
	if file == nil {
		t.Fatalf("no file parsed from %q", sourceText)
	}
	return file
}

// The module each listed declaration imports from, in the order Declarations returns them.
func declarationSources(file *ast.SourceFile) []string {
	sources := []string{}
	for _, declaration := range Declarations(file) {
		sources = append(sources, declaration.AsImportDeclaration().ModuleSpecifier.Text())
	}
	return sources
}

func TestDeclarationsReachesAmbientModuleBodies(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		sourceText  string
		wantSources []string
	}{
		{"top-level imports only", "import a from 'one';\nimport { b } from 'two';\n", []string{"one", "two"}},
		// TanStack examples/vue/*/src/shims-vue.d.ts, verbatim.
		{
			"an import inside a declare module body",
			"declare module '*.vue' {\n  import { DefineComponent } from 'vue'\n" +
				"  const component: DefineComponent<{}, {}, any>\n  export default component\n}\n",
			[]string{"vue"},
		},
		{
			"top level and module body together, in source order",
			"import a from 'first';\ndeclare module 'm' {\n  import b from 'second';\n}\nimport c from 'third';\n",
			[]string{"first", "second", "third"},
		},
		{
			"two module bodies",
			"declare module 'm' { import a from 'one'; }\ndeclare module 'n' { import b from 'two'; }\n",
			[]string{"one", "two"},
		},
		// A dotted name nests one declaration's body inside another's.
		{"a dotted namespace", "declare module A.B {\n  import c from 'deep';\n}\n", []string{"deep"}},
		{"a file with no imports", "export const value = 1;\n", []string{}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := declarationSources(parseDeclarationFile(t, testCase.sourceText))
			if !slices.Equal(got, testCase.wantSources) {
				t.Errorf("Declarations found imports from %q, want %q", got, testCase.wantSources)
			}
		})
	}
}

func TestDeclarationsOfNoFile(t *testing.T) {
	t.Parallel()
	if declarations := Declarations(nil); declarations != nil {
		t.Errorf("Declarations(nil) = %v, want nil", declarations)
	}
}
