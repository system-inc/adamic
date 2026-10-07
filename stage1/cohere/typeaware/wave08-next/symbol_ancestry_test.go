package wave08next

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func TestRawSymbolAncestry(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "tsconfig.json")
	source := filepath.Join(root, "main.ts")
	ambient := filepath.Join(root, "node.d.ts")
	for path, text := range map[string]string{
		config:  `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"types":[],"moduleDetection":"force"},"include":["*.ts"]}`,
		ambient: `export {}; declare global { namespace NodeJS { interface Process { exit(code?:number):never; } } var process:NodeJS.Process; var console:{log(value:unknown):void}; }`,
		source:  `const {x}= {x:1}; console.log(x); process.exit(0);`,
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
	checker, release := program.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	seen := map[string]bool{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && (node.Text() == "exit" || node.Text() == "console" || node.Text() == "x") {
			fields := SymbolAncestryFields(checker, node)
			if len(fields) < 10 || fields[0] != "1" || fields[1] != "wave08-symbol-ancestry" || fields[3] != "1" {
				t.Fatalf("raw symbol shape differs: %q", fields)
			}
			if node.Text() == "exit" {
				if len(fields) < 24 || fields[4] != ambient || fields[5] != "1" || fields[7] != "MethodSignature" || fields[11] != "InterfaceDeclaration" || fields[12] != "Process" || fields[19] != "ModuleDeclaration" || fields[20] != "NodeJS" {
					t.Fatalf("process ancestor differs: %q", fields)
				}
				seen["exit"] = true
			}
			if node.Text() == "console" {
				if len(fields) < 27 || fields[7] != "VariableDeclaration" || fields[11] != "VariableDeclarationList" || fields[15] != "VariableStatement" || fields[19] != "ModuleBlock" || fields[23] != "ModuleDeclaration" || fields[26] != "1" {
					t.Fatalf("global ancestor differs: %q", fields)
				}
				seen["console"] = true
			}
			if node.Text() == "x" {
				seen["pattern"] = true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if !seen["exit"] || !seen["console"] || !seen["pattern"] {
		t.Fatalf("missing ancestry witnesses: %v", seen)
	}
	t.Log("raw NodeJS.Process, declare-global console, and destructuring ancestry match the pinned checker")
}
