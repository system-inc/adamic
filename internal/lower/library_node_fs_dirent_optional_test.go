package lower

import (
	"context"
	"errors"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const direntOptionalReceiverSource = "import {readdirSync, type Dirent} from 'node:fs';\nfunction inspect(dirent: Dirent | undefined) { console.log(`${dirent?.isFile()}`); }\ninspect(undefined);\ninspect(readdirSync('.', {withFileTypes: true})[0]);\n"
const direntOptionalReceiverCatch = "unguarded nodeFSDirectoryCall must refuse node:fs.Dirent.isFile through an optional call"

func TestNodeFSDirentOptionalReceiverRequiresGuard(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", direntOptionalReceiverSource)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	func() {
		checked, release := program.Checker(context.Background(), file)
		defer release()
		l := &lowering{program: program, checker: checked, result: &ir.Program{Locals: []ir.Local{{Name: "dirent", Type: ir.Object, Global: true, Function: -1}}}, this: -1, functionIndex: -1}
		calls := 0
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().QuestionDotToken != nil {
					calls++
					l.locals = map[*ast.Symbol]int{l.symbol(callee.AsPropertyAccessExpression().Expression): 0}
					value, handled, err := l.nodeFSDirectoryCall(node)
					var refused *NotYet
					if !handled || !errors.As(err, &refused) || refused.What != "node:fs.Dirent.isFile through an optional call" {
						t.Fatalf("%s: handled=%t value=%T err=%v", direntOptionalReceiverCatch, handled, value, err)
					}
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if calls != 1 {
			t.Fatalf("want one optional Dirent call, got %d", calls)
		}
	}()
	if _, err := Lower(context.Background(), program); err != nil {
		t.Fatalf("guarded full Lower: %v", err)
	}
}

func TestNodeFSDirentOptionalReceiverFullLower(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, direntOptionalReceiverSource, ".ts"); err != nil {
		t.Fatal(err)
	}
}

func TestNodeFSDirentOptionalFunctionsAndContinuationsStop(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{
		"function inspect(value: Dirent | undefined) { return value?.isFile?.(); }",
		"function inspect(value: {entry: Dirent} | undefined) { return value?.entry.isFile(); }",
	} {
		_, err := lowerSourceAs(t, "import {type Dirent} from 'node:fs';\n"+expression, ".ts")
		var refused *NotYet
		if !errors.As(err, &refused) || refused.What != "a call through ?. (an optional call)" {
			t.Fatalf("optional function or continuation must retain refusal, got %v", err)
		}
	}
}

func direntOptionalReceiverMarkMutant(t *testing.T) string {
	t.Helper()
	mutations := []buildcache.Mutation{{File: "internal/lower/library_node_fs_directory.go", Before: "(!receiverOptional || l.optionalReceiverCall != node)", After: "!receiverOptional"}}
	return buildcache.GoTestMutated(t, "", "dirent-optional-mark-lower.test", "./internal/lower", nil, mutations)
}

func TestProduct_DirentOptionalReceiverMarkMutant(t *testing.T) {
	t.Parallel()
	direntOptionalReceiverMarkMutant(t)
}

func TestNodeFSDirentOptionalReceiverMarkMutantIsCaught(t *testing.T) {
	t.Parallel()
	lower := direntOptionalReceiverMarkMutant(t)
	checkEntryReceiverMarkMutant(t, lower, "TestNodeFSDirentOptionalReceiverRequiresGuard", direntOptionalReceiverCatch)
}
