package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestReadSymbolQuestion(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{filepath.Join(directory, "prelude.d.ts"): "declare const fixture: unknown;", config: `{"compilerOptions":{"lib":["ES2022"],"types":[]},"files":["prelude.d.ts"]}`, file: `const local = Math; const object = {local}; export {local as exported};`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	count := 0
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier {
			expected := c.GetSymbolAtLocation(n)
			if n.Parent.Kind == ast.KindShorthandPropertyAssignment && n.Parent.Name() == n {
				expected = c.GetShorthandAssignmentValueSymbol(n.Parent)
				count++
			}
			if n.Parent.Kind == ast.KindExportSpecifier {
				expected = c.GetExportSpecifierLocalTargetSymbol(n.Parent)
				count++
			}
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "read-symbol")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			if expected == nil {
				if fields[2] != "0" {
					t.Fatal("unresolved symbol became present")
				}
			} else if fields[3] != strconv.FormatUint(p.symbolID(expected), 10) {
				t.Fatal("read-symbol returned property or export instead of read binding")
			}
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "read-symbol\nsuffix"); err == nil {
				t.Fatal("accepted suffix")
			}
		}
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(source.AsNode())
	if count < 3 {
		t.Fatalf("special sites missing: %d", count)
	}
	if _, err := p.Inspect(file, 0, uint64(source.End()), strings.TrimPrefix(source.Kind.String(), "Kind"), "read-symbol"); err == nil {
		t.Fatal("accepted wrong kind")
	}
}
