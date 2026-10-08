package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobalScopeQuestions(t *testing.T) {
	for _, row := range []struct{ name, source, metadata string }{
		{"module.a", "export {};Object=1;({String}=o);", "1|0|2|0"},
		{"script.a", "Object=1;({String}=o);", "0|0|2|0"},
		{"script.js", "Object=1;({String}=o);", "0|1|2|0"},
	} {
		t.Run(row.name, func(t *testing.T) {
			directory := t.TempDir()
			file := filepath.Join(directory, row.name)
			config := filepath.Join(directory, "tsconfig.json")
			if err := os.WriteFile(file, []byte(row.source), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"moduleDetection":"auto","allowJs":true,"target":"ES2022","lib":["ES2022"]},"files":["`+row.name+`"]}`), 0600); err != nil {
				t.Fatal(err)
			}
			p, err := Open(config, []string{file})
			if err != nil {
				t.Fatal(err)
			}
			sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
			wire, err := p.Inspect(file, 0, uint64(len(row.source)), "SourceFile", "global-source-facts")
			if err != nil {
				t.Fatal(err)
			}
			f := decodedFields(t, wire)
			if strings.Join(f[2:], "|") != row.metadata {
				t.Fatalf("metadata: %q", f)
			}
			var object *ast.Node
			var walk func(*ast.Node)
			walk = func(n *ast.Node) {
				if n.Kind == ast.KindIdentifier && n.Text() == "Object" {
					object = n
				}
				n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
			}
			walk(sf.AsNode())
			if object == nil {
				t.Fatal("Object absent")
			}
			wire, err = p.Inspect(file, uint64(object.Pos()), uint64(object.End()), "Identifier", "global-binding-facts")
			if err != nil {
				t.Fatal(err)
			}
			f = decodedFields(t, wire)
			if len(f) < 8 || f[2] != "1" || f[5] != "Object" {
				t.Fatalf("ordinary global symbol absent: %q", f)
			}
			for _, question := range []string{"global-source-facts", "global-binding-facts\nextra", "unknown-question"} {
				if _, err := p.Inspect(file, uint64(object.Pos()), uint64(object.End()), "Identifier", question); err == nil {
					t.Fatalf("accepted malformed request %q", question)
				}
			}
			if _, err := p.Inspect(file, 0, uint64(len(row.source)), "SourceFile", "global-binding-facts"); err == nil {
				t.Fatal("accepted binding on source")
			}
		})
	}
}
