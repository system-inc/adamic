package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDeclarationContractFacts(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["prelude.d.ts"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("const f:(x:number)=>number=(x=42)=>x;export {};\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "prelude.d.ts"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	var parameter *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindParameter && node.AsParameterDeclaration().Initializer != nil {
			parameter = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if parameter == nil {
		t.Fatal("no parameter")
	}
	wire, err := p.Inspect(file, uint64(parameter.Pos()), uint64(parameter.End()), "Parameter", "declaration-contract")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[3] != "1" || fields[4] != "0" || fields[5] != "1" || fields[7] != "0" {
		t.Fatalf("contract metadata: %q", fields)
	}
	id, err := strconv.Atoi(fields[11])
	if err != nil {
		t.Fatal(err)
	}
	signature := c.GetSignaturesOfType(checker.Checker_getContextualType(c, parameter.Parent, checker.ContextFlagsNone), checker.SignatureKindCall)[0]
	expected := checker.Checker_getTypeOfSymbol(c, checker.Signature_parameters(signature)[0])
	if p.typesByID[id-1] != expected {
		t.Fatal("contextual type differs from direct checker")
	}
	if _, err := p.Inspect(file, uint64(source.Pos()), uint64(source.End()), "SourceFile", "declaration-contract"); err == nil {
		t.Fatal("accepted a non-parameter")
	}
	if _, err := p.Inspect(file, uint64(parameter.Pos()), uint64(parameter.End()), "Parameter", "declaration-contract\nextra"); err == nil {
		t.Fatal("accepted question suffix")
	}
	if !strings.Contains(wire, "declaration-contract") {
		t.Fatal("missing schema name")
	}
}
