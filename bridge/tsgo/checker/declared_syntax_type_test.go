package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDeclaredSyntaxType(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{file: "type Props = { enabled: boolean }; function C(p: Props) { return p.enabled; } const v: number = 1;\n", config: `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["input.a"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	nodes, ids := syntaxNodes(program.Compiler.GetSourceFile(file))
	count := 0
	for _, node := range nodes {
		var expected *ast.Node
		switch node.Kind {
		case ast.KindParameter:
			expected = node.AsParameterDeclaration().Type
		case ast.KindTypeAliasDeclaration:
			expected = node.AsTypeAliasDeclaration().Type
		case ast.KindPropertySignature:
			expected = node.AsPropertySignatureDeclaration().Type
		case ast.KindVariableDeclaration:
			expected = node.AsVariableDeclaration().Type
		default:
			continue
		}
		wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), nodeKind(node), "declared-syntax-type")
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		if len(fields) != 3 || fields[0] != "1" || fields[1] != "declared-syntax-type" || fields[2] != strconv.FormatUint(ids[expected], 10) {
			t.Fatal("lost declared syntax type", fields)
		}
		if expected != nil {
			count++
		}
	}
	if count < 4 {
		t.Fatal("missing type controls", count)
	}
	node := program.Compiler.GetSourceFile(file).AsNode()
	if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), nodeKind(node), "declared-syntax-type\nextra"); err == nil {
		t.Fatal("accepted syntax type arguments")
	}
}
