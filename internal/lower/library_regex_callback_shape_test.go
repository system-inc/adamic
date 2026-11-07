package lower

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryRegexOffsetRequiresNoCaptures(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"console.log('a'.replace(/(a)/, (match: string, offset: number) => `${match}:${offset}`));",
		"console.log('a'.replace(/(?=(a))a/, (match: string, offset: number) => `${match}:${offset}`));",
		"function replace(pattern: RegExp): string { return 'a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`); }",
		"let pattern = /a/; console.log('a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`));",
	} {
		path := filepath.Join(test.TempDir(), "main.a")
		if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
			test.Fatal(writeError)
		}
		program, loadError := load.Load([]string{path})
		if loadError != nil {
			test.Fatal(loadError)
		}
		checker, release := program.Checker(context.Background(), program.Files()[0])
		lowering := &lowering{program: program, checker: checker}
		found := false
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression &&
				node.AsCallExpression().Arguments != nil &&
				len(node.AsCallExpression().Arguments.Nodes) == 2 {
				found = true
				if lowering.regexCallbackNoCaptures(node.AsCallExpression().Arguments.Nodes[0], 0) {
					test.Fatal("capture-free producer proof admitted an unproven pattern")
				}
			}
			return node.ForEachChild(visit)
		}
		program.Files()[0].ForEachChild(visit)
		release()
		if !found {
			test.Fatal("no producer examined")
		}
	}
}
