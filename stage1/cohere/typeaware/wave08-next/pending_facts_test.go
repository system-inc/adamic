package wave08next

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPendingFactsAndRegistrationRefusals(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "tsconfig.json")
	source := filepath.Join(root, "main.ts")
	other := filepath.Join(root, "other.ts")
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","noEmit":true},"include":["*.ts"]}`, other: `export const value=1;`, source: `import {value} from './other.js'; function fn(a:number):number{return a+value;} fn(1); import('./other.js');`} {
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
	var call *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindIdentifier && node.AsCallExpression().Expression.Text() == "fn" {
			call = node
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if call == nil {
		t.Fatal("call absent")
	}
	fields, err := ResolvedCalleeFields(c, call)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 14 || fields[2] != "1" || fields[3] != source || fields[4] != "FunctionDeclaration" || fields[10] != "1" || fields[13] != "64" {
		t.Fatalf("resolved callee differs: %q", fields)
	}
	fields, err = ProgramLoadsFields(program.Compiler, file.AsNode())
	if err != nil {
		t.Fatal(err)
	}
	resolved := 0
	for _, field := range fields {
		if field == other {
			resolved++
		}
	}
	if resolved < 3 {
		t.Fatalf("static and dynamic module resolutions absent: %d", resolved)
	}
	for _, request := range []struct {
		node     *ast.Node
		question string
	}{{call, "wave08-resolved-callee"}, {file.AsNode(), "wave08-program-loads"}} {
		_, err = program.Inspect(source, uint64(request.node.Pos()), uint64(request.node.End()), strings.TrimPrefix(request.node.Kind.String(), "Kind"), request.question)
		if err == nil || !strings.Contains(err.Error(), "unsupported checker question") {
			t.Fatalf("unregistered request did not refuse: %v", err)
		}
	}
	t.Log("raw resolved callee and static/dynamic module facts pass; both native requests explicitly refuse because registration is absent")
}

func TestNativeFactFlagMasks(t *testing.T) {
	if checker.TypeFlagsNever != 262144 || checker.TypeFlagsNumber != 64 || ast.FunctionFlagsGenerator != 1 || ast.FunctionFlagsAsync != 2 {
		t.Fatal("update native masks for the pinned checker")
	}
}
