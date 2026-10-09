package rule_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// declarationsProbe builds a program where one symbol is declared in two files: `Shared`, an interface
// two scripts each declare, which merges into one global symbol. It returns that symbol and both files,
// and the symbol of `parseInt`, whose declarations are all in the lib.
func declarationsProbe(t *testing.T) (shared *ast.Symbol, first *ast.SourceFile, second *ast.SourceFile, global *ast.Symbol) {
	t.Helper()
	directory := t.TempDir()
	files := map[string]string{
		"tsconfig.json": `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], "types": []}, "include": ["*.ts"]}`,
		"first.ts":      "interface Shared { a: string }\ndeclare const fromFirst: Shared;\nparseInt('1');\n",
		"second.ts":     "interface Shared { b: string }\ndeclare const fromSecond: Shared;\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(directory, "tsconfig.json"), CurrentDirectory: directory, SingleThreaded: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, sourceFile := range graph.ProjectFiles() {
		switch filepath.Base(sourceFile.FileName().AsString()) {
		case "first.ts":
			first = sourceFile
		case "second.ts":
			second = sourceFile
		}
	}
	if first == nil || second == nil {
		t.Fatal("the probe program does not hold both files")
	}
	checker, release := graph.CheckerForFile(context.Background(), first)
	defer release()
	statements := first.Statements.Nodes
	shared = checker.GetSymbolAtLocation(statements[0].Name())
	callee := statements[2].AsExpressionStatement().Expression.AsCallExpression().Expression
	global = checker.GetSymbolAtLocation(callee)
	if shared == nil || len(shared.Declarations) != 2 || global == nil {
		t.Fatalf("the probe did not produce a symbol merged across two files and a lib global (%v, %v)", shared, global)
	}
	return shared, first, second, global
}

// DeclarationsIn keeps exactly the declarations in the file it is handed, from either side of a merge,
// so a rule reading through it cannot reach another file's declaration.
func TestDeclarationsInKeepsOnlyTheFilesOwn(t *testing.T) {
	t.Parallel()
	shared, first, second, global := declarationsProbe(t)
	for _, sourceFile := range []*ast.SourceFile{first, second} {
		own := rule.DeclarationsIn(sourceFile, shared)
		if len(own) != 1 || ast.GetSourceFileOfNode(own[0]) != sourceFile {
			t.Errorf("%s: kept %d declarations, want its own one", filepath.Base(sourceFile.FileName().AsString()), len(own))
		}
	}
	if own := rule.DeclarationsIn(first, global); len(own) != 0 {
		t.Errorf("a lib global kept %d declarations in a source file, want none", len(own))
	}
	if rule.DeclarationsIn(first, nil) != nil || rule.DeclarationsIn(nil, shared) != nil {
		t.Error("a nil symbol or file gave declarations")
	}
}

// IsDeclaredOnlyInDeclarationFiles tells a lib global from a binding a source file declares.
func TestIsDeclaredOnlyInDeclarationFilesTellsALibGlobalFromASourceBinding(t *testing.T) {
	t.Parallel()
	shared, _, _, global := declarationsProbe(t)
	if !rule.IsDeclaredOnlyInDeclarationFiles(global) {
		t.Error("parseInt, declared only in the lib, was not recognised as a global")
	}
	if rule.IsDeclaredOnlyInDeclarationFiles(shared) {
		t.Error("an interface two source files declare was taken for a global")
	}
	if rule.IsDeclaredOnlyInDeclarationFiles(nil) {
		t.Error("a nil symbol was taken for a global")
	}
	if !rule.IsDeclaredInASourceFile(shared) {
		t.Error("an interface two source files declare was not recognised as declared in a source file")
	}
	if rule.IsDeclaredInASourceFile(global) || rule.IsDeclaredInASourceFile(nil) {
		t.Error("parseInt, or a nil symbol, was taken for something a source file declares")
	}
}
