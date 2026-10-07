package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"testing"
)

func TestSyntaxReferenceFacts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	text := "let r = RegExp; r = other; const object = {r}; export {r}; function f() { return new.target; } const {RegExp: R} = globalThis;\n"
	for path, contents := range map[string]string{file: text, config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"]},"files":["input.a"]}`} {
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	nodes, ids := syntaxNodes(program.Compiler.GetSourceFile(file))
	seenDeclaration, seenWrite, seenShorthand, seenExport, seenMeta, seenBinding := false, false, false, false, false, false
	identity := ""
	for _, node := range nodes {
		wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), nodeKind(node), "syntax-reference-facts")
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		if len(fields) < 10 || fields[0] != "1" || fields[1] != "syntax-reference-facts" {
			t.Fatal("malformed reference fact", fields)
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "r" {
			if node.Parent.Kind == ast.KindVariableDeclaration {
				if fields[2] != "1" || fields[6] == "0" {
					t.Fatal("lost declaration", fields)
				}
				identity = fields[6]
				seenDeclaration = true
			}
			if node.Parent.Kind == ast.KindBinaryExpression {
				if fields[3] != "1" {
					t.Fatal("lost compiler write classification", fields)
				}
				seenWrite = true
			}
			if node.Parent.Kind == ast.KindShorthandPropertyAssignment {
				if fields[6] != identity {
					t.Fatal("shorthand did not read value binding", fields)
				}
				seenShorthand = true
			}
			if node.Parent.Kind == ast.KindExportSpecifier {
				if fields[6] != identity {
					t.Fatal("export did not read local binding", fields)
				}
				seenExport = true
			}
		}
		if node.Kind == ast.KindMetaProperty {
			if fields[4] != "NewKeyword" {
				t.Fatal("lost meta keyword", fields)
			}
			seenMeta = true
		}
		if node.Kind == ast.KindBindingElement {
			if fields[5] == "0" || ids[node.AsBindingElement().PropertyName] == 0 {
				t.Fatal("lost binding property", fields)
			}
			seenBinding = true
		}
	}
	if !seenDeclaration || !seenWrite || !seenShorthand || !seenExport || !seenMeta || !seenBinding {
		t.Fatal("missing reference controls")
	}
	root := program.Compiler.GetSourceFile(file).AsNode()
	if _, err := program.Inspect(file, uint64(root.Pos()), uint64(root.End()), nodeKind(root), "syntax-reference-facts\nextra"); err == nil {
		t.Fatal("accepted reference question arguments")
	}
}
