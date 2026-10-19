package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// A panic in the compiler exits 4 with "adamic: compiler bug:", its value and its stack, never Go's 2, which is
// usage's. Mutants: reportCompilerBug without its recover (the panic escapes the test), or setting 2.
func TestCompilerBugExitsFour(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	code := func() (code int) {
		defer reportCompilerBug(&code, &output)
		panic("probe")
	}()
	if code != 4 || !strings.HasPrefix(output.String(), "adamic: compiler bug: probe\n") || !strings.Contains(output.String(), "goroutine ") {
		t.Fatalf("a compiler panic: exit %d, stderr %q; want exit 4, the panic's value and its stack", code, output.String())
	}
	output.Reset()
	code = func() (code int) {
		defer reportCompilerBug(&code, &output)
		return 1
	}()
	if code != 1 || output.Len() != 0 {
		t.Fatalf("no panic: exit %d, stderr %q; want the command's own exit 1 and nothing written", code, output.String())
	}
}

// run's first statement defers reportCompilerBug on its own exit code, so every command is covered. Mutant: delete
// the defer.
func TestRunRecoversCompilerBugs(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "run" {
			continue
		}
		if len(function.Body.List) > 0 {
			if deferred, ok := function.Body.List[0].(*ast.DeferStmt); ok {
				if called, ok := deferred.Call.Fun.(*ast.Ident); ok && called.Name == "reportCompilerBug" && len(deferred.Call.Args) == 2 {
					if code, ok := deferred.Call.Args[0].(*ast.UnaryExpr); ok && code.Op == token.AND {
						if name, ok := code.X.(*ast.Ident); ok && name.Name == "code" {
							return
						}
					}
				}
			}
		}
		t.Fatal("run's first statement is not defer reportCompilerBug(&code, ...)")
	}
	t.Fatal("main.go has no run")
}
