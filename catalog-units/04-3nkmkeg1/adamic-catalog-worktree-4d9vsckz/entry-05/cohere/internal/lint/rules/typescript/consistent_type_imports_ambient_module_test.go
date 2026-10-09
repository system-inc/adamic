package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// An import may sit inside an ambient module body as well as at the top of a file, and upstream
// listens to ImportDeclaration wherever it appears. Every expectation below is what the installed
// ESLint 10.8.1 with typescript-eslint 8.71.0 reports and writes for the same source: the
// findings' spans and the fixed text both. Upstream's own corpus has no declare-module rows to
// replay, so the release binary is the authority here.
func TestConsistentTypeImportsReachesAmbientModuleBodies(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name     string
		fileName string
		source   string
		options  string
		spans    []string
		ids      []string
		fixed    string
	}{
		{
			// TanStack examples/vue/*/src/shims-vue.d.ts, verbatim.
			name:     "theVueShim",
			fileName: "shims-vue.d.ts",
			source: "declare module '*.vue' {\n  import { DefineComponent } from 'vue'\n" +
				"  const component: DefineComponent<{}, {}, any>\n  export default component\n}\n",
			spans: []string{"import { DefineComponent } from 'vue'"},
			ids:   []string{"typeOverValue"},
			fixed: "declare module '*.vue' {\n  import type { DefineComponent } from 'vue'\n" +
				"  const component: DefineComponent<{}, {}, any>\n  export default component\n}\n",
		},
		{
			// A top-level import and one inside a module body, each judged on its own.
			name:     "topLevelAndAmbientImportsTogether",
			fileName: "consistent_type_imports.ts",
			source: "import { Top } from 'top'\nexport const value: Top = null!\ndeclare module 'inner' {\n" +
				"  import { Inner } from 'y'\n  const component: Inner\n  export default component\n}\n",
			spans: []string{"import { Top } from 'top'", "import { Inner } from 'y'"},
			ids:   []string{"typeOverValue", "typeOverValue"},
			fixed: "import type { Top } from 'top'\nexport const value: Top = null!\ndeclare module 'inner' {\n" +
				"  import type { Inner } from 'y'\n  const component: Inner\n  export default component\n}\n",
		},
		{
			// The inverted setting reaches the module body too.
			name:     "noTypeImportsInsideTheModuleBody",
			fileName: "shims-vue.d.ts",
			source: "declare module '*.vue' {\n  import type { DefineComponent } from 'vue'\n" +
				"  const component: DefineComponent\n  export default component\n}\n",
			options: "{\"prefer\":\"no-type-imports\"}",
			spans:   []string{"import type { DefineComponent } from 'vue'"},
			ids:     []string{"avoidImportType"},
			fixed: "declare module '*.vue' {\n  import { DefineComponent } from 'vue'\n" +
				"  const component: DefineComponent\n  export default component\n}\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports, testCase.fileName, testCase.source,
				decodeConsistentTypeImportsFixtureOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
			for position, diagnostic := range result.Diagnostics {
				if position >= len(testCase.spans) {
					break
				}
				covered := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
				if covered != testCase.spans[position] {
					t.Errorf("finding %d covers %q, want %q", position, covered, testCase.spans[position])
				}
			}
			rule_testing.ExpectFixedSource(t, result, testCase.fixed)
		})
	}
}

// A binding the module body uses as a value stays a value import, inside the body as at the top.
func TestConsistentTypeImportsLeavesAmbientValueImports(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports, "value-use.d.ts",
		"declare module 'made' {\n  import { thing } from 'x'\n  export default thing\n}\n",
		DefaultConsistentTypeImportsOptions())
	rule_testing.ExpectClean(t, result)
}
