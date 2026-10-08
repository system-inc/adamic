package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave24NextRawQuestions(t *testing.T) {
	if checker.TypeFlagsNever != 262144 || ast.FunctionFlagsGenerator != 1 || ast.FunctionFlagsAsync != 2 || ast.NodeFlagsConst != 2 || ast.NodeFlagsAmbient != 8388608 || ast.SymbolFlagsFunction != 16 {
		t.Fatal("pinned wave-24 flags changed")
	}
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	source := "declare const target:string;function helper():never {throw 1;} helper();import(target);export {};\n"
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"]}}`, file: source, filepath.Join(directory, "globals.d.ts"): "declare var process:{exit(code?:number):never};\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := program.Compiler.GetSourceFile(tspath.RootedFilePath(file))
	var name, call *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.Text() == "helper" {
			name = node
		}
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindIdentifier {
			call = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := program.InspectWave24(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	if name == nil || call == nil {
		t.Fatal("missing positive controls")
	}
	origin := ask(name, "symbol-provenance")
	if len(origin) < 12 || origin[2] == "0" || origin[3] != "16" || origin[4] != "1" || origin[5] != file || origin[10] != "FunctionDeclaration" || origin[11] != "helper" {
		t.Fatal("raw declaration mismatch", origin)
	}
	signature := ask(call, "resolved-call-origin")
	if signature[2] != "262144" || signature[3] != "1" || signature[4] != file || signature[5] != "FunctionDeclaration" || signature[len(signature)-1] != "1" {
		t.Fatal("raw signature mismatch", signature)
	}
	edges := ask(sf.AsNode(), "program-module-edges")
	found := false
	for at := 3; at < len(edges); {
		path := edges[at]
		computed := edges[at+2]
		count := 0
		for _, digit := range edges[at+3] {
			count = count*10 + int(digit-'0')
		}
		if path == file && computed == "1" {
			found = true
		}
		at += 4 + count
	}
	if !found {
		t.Fatal("computed import missing", edges)
	}
	for _, question := range []string{"symbol-provenance\nwrong", "resolved-call-origin", "program-module-edges"} {
		if _, err := program.InspectWave24(file, uint64(name.Pos()), uint64(name.End()), "Identifier", question); err == nil {
			t.Fatal("accepted mismatched question", question)
		}
	}
}
