package lower

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryRegexOffsetRequiresNoCaptures(t *testing.T) {
	for _, source := range []string{
		"console.log('a'.replace(/(a)/, (match: string, offset: number) => `${match}:${offset}`));",
		"console.log('a'.replace(/(?=(a))a/, (match: string, offset: number) => `${match}:${offset}`));",
		"function replace(pattern: RegExp): string { return 'a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`); }",
		"let pattern = /a/; console.log('a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`));",
	} {
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		checker, release := program.Checker(context.Background(), program.Files()[0])
		l := &lowering{program: program, checker: checker}
		found := false
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression && node.AsCallExpression().Arguments != nil && len(node.AsCallExpression().Arguments.Nodes) == 2 {
				found = true
				if l.regexCallbackNoCaptures(node.AsCallExpression().Arguments.Nodes[0], 0) {
					t.Fatal("capture-free producer proof admitted an unproven pattern")
				}
			}
			return node.ForEachChild(visit)
		}
		program.Files()[0].ForEachChild(visit)
		release()
		if !found {
			t.Fatal("no producer examined")
		}
	}
}
