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

func TestWave12NextRawQuestions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "import type {T} from './other';import {f,a} from './other';f();a;\n"
	for name, text := range map[string]string{"input.a": source, "other.ts": "export type T=number;export function f():never{throw 0};export const { a } = {a:1};\n", "tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"]},"files":["input.a","other.ts"]}`} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var identifier, call, compound *ast.Node
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && n.Text() == "a" {
			compound = n
		}
		if n.Kind == ast.KindCallExpression {
			call = n
			identifier = n.AsCallExpression().Expression
		}
		n.ForEachChild(visit)
		return false
	}
	sf.AsNode().ForEachChild(visit)
	if call == nil || identifier == nil || compound == nil {
		t.Fatal("missing controls")
	}
	ask := func(n *ast.Node, q string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), q)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	ancestry := ask(identifier, "symbol-ancestry")
	symbol := c.GetAliasedSymbol(c.GetSymbolAtLocation(identifier))
	declaration := symbol.Declarations[0]
	if len(ancestry) < 17 || ancestry[3] != "1" || ancestry[4] != filepath.Join(directory, "other.ts") || ancestry[5] != "0" || ancestry[7] != "1" || ancestry[9] != "FunctionDeclaration" || ancestry[10] != strconv.Itoa(declaration.Pos()) || ancestry[14] != "f" {
		t.Fatalf("alias ancestry: %q", ancestry)
	}
	compoundAncestry := ask(compound, "symbol-ancestry")
	foundCompound := false
	for _, field := range compoundAncestry {
		if field == "{ a }" {
			foundCompound = true
		}
	}
	if !foundCompound {
		t.Fatalf("compound declaration name missing: %q", compoundAncestry)
	}
	location := ask(call, "call-declaration")
	signature := c.GetResolvedSignature(call)
	if len(location) != 12 || location[2] != "1" || location[3] != filepath.Join(directory, "other.ts") || location[4] != "FunctionDeclaration" || location[7] != "1" || location[8] != "1" || location[9] != "0" || location[10] != "0" || location[11] != strconv.FormatUint(uint64(checker.TypeFlagsNever), 10) || c.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsNever == 0 {
		t.Fatalf("call declaration: %q", location)
	}
	imports := ask(sf.AsNode(), "program-imports")
	if imports[2] != strconv.Itoa(len(p.Compiler.SourceFiles())) {
		t.Fatalf("program count: %q", imports)
	}
	at := 3
	found := false
	for range p.Compiler.SourceFiles() {
		name := imports[at]
		count, err := strconv.Atoi(imports[at+2])
		if err != nil {
			t.Fatal(err)
		}
		at += 3
		if name == file {
			if count != 2 || imports[at] != "ImportDeclaration" || imports[at+1] != "1" || imports[at+2] != "1" || imports[at+3] != filepath.Join(directory, "other.ts") || imports[at+5] != "0" || imports[at+7] != filepath.Join(directory, "other.ts") {
				t.Fatalf("resolved imports: %q", imports[at:at+count*4])
			}
			found = true
		}
		at += count * 4
	}
	if !found || at != len(imports) {
		t.Fatal("imports missing or trailing")
	}
	for _, q := range []string{"symbol-ancestry\nextra", "call-declaration\nextra", "program-imports\nextra", "program-imports"} {
		if _, err := p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), "Identifier", q); err == nil {
			t.Fatal("invalid request accepted", q)
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "symbol-ancestry"); err == nil {
		t.Fatal("ancestry accepted SourceFile")
	}
	if _, err := p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), "Identifier", "call-declaration"); err == nil {
		t.Fatal("signature accepted Identifier")
	}
}
