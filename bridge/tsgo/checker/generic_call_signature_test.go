package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWave24FifthRawQuestions(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	for path, text := range map[string]string{filepath.Join(directory, "globals.d.ts"): "declare const positiveControl:number;", config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`, file: "function f<T>(...callbacks:(()=>T)[]):void{}f(async()=>1);interface I{run():Promise<number>;}class C implements I{async run(){return 1;}}export {};"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var call, class *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		if node.Kind == ast.KindClassDeclaration {
			class = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(program.Compiler.GetSourceFile(tspath.RootedFilePath(file)).AsNode())
	if call == nil || class == nil {
		t.Fatal("missing positive controls")
	}
	inspect := func(node *ast.Node, kind, question string) []string {
		wire, err := program.InspectWave24(file, uint64(node.Pos()), uint64(node.End()), kind, question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	fields := inspect(call, "CallExpression", "generic-call-signature")
	if fields[2] != "1" || fields[3] != "1" || fields[5] != "1" || fields[6] != "1" {
		t.Fatal("generic rest signature missing", fields)
	}
	identity := fields[7]
	indexFields := inspect(call, "CallExpression", "type-index\n"+identity+"\nnumber")
	if indexFields[3] != "1" || indexFields[4] != "1" {
		t.Fatal("missing numeric rest element", indexFields)
	}
	heritage := inspect(class, "ClassDeclaration", "heritage-types")
	if heritage[4] != "1" {
		t.Fatal("missing ordered heritage root", heritage)
	}
	for _, question := range []string{"generic-call-signature\nwrong", "type-index\n0\nnumber", "type-index\n" + identity + "\nwrong", "type-index\n0" + identity + "\nnumber", "type-index\n" + strconv.FormatUint(^uint64(0), 10) + "\nnumber", "heritage-types\nwrong"} {
		if _, err := program.InspectWave24(file, uint64(call.Pos()), uint64(call.End()), "CallExpression", question); err == nil {
			t.Fatal("accepted malformed question", question)
		}
	}
	if _, err := program.InspectWave24(file, uint64(class.Pos()), uint64(class.End()), "ClassDeclaration", "generic-call-signature"); err == nil {
		t.Fatal("accepted non-call generic question")
	}
}
