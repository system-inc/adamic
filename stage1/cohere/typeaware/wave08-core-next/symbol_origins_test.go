package wave08core_test

import (
	"context"
	wave08core "github.com/system-inc/adamic/stage1/cohere/typeaware/wave08-core-next"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func TestSymbolOriginsPreservesAliasesAndDeclarationFiles(t *testing.T) {
	root := t.TempDir()
	config, source, imported := filepath.Join(root, "tsconfig.json"), filepath.Join(root, "main.ts"), filepath.Join(root, "alias.d.ts")
	for path, text := range map[string]string{
		config:   `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"types":[],"module":"NodeNext"},"include":["*.ts"]}`,
		imported: `export { Symbol };`,
		source:   `import { Symbol as Imported } from "./alias"; Symbol(); Imported(); { declare function Local(): symbol; Local(); }`,
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := bridge.Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Compiler.GetSourceFile(source)
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	seen := map[string]bool{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindIdentifier {
			callee := node.AsCallExpression().Expression
			fields := wave08core.SymbolOriginsFields(c, callee)
			if len(fields) < 4 {
				t.Fatalf("missing raw declarations for %s: %v", callee.Text(), fields)
			}
			want := "0"
			if callee.Text() == "Symbol" {
				want = "1"
			}
			if fields[3] != want {
				t.Fatalf("%s first declaration file flag: %v, want %s", callee.Text(), fields, want)
			}
			seen[callee.Text()] = true
		}
		node.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
	for _, name := range []string{"Symbol", "Imported", "Local"} {
		if !seen[name] {
			t.Fatalf("missing witness %s", name)
		}
	}
	if ast.KindCallExpression != 214 || ast.KindIdentifier != 79 || ast.KindParenthesizedExpression != 218 {
		t.Fatal("pinned parser numeric kinds changed")
	}
	t.Log("global library, unfollowed import alias and ambient source declaration flags match; numeric kinds 214,79,218 pinned")
}
