package lint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// serialTests are the package's top-level tests that run alone, each for a reason that still holds. Go runs
// every serial top-level test to completion before it releases any parallel one, so a serial test's whole
// time is added to the package's wall, and the first one to need a shared build pays for it alone. Eleven
// tests added without t.Parallel took the package from 2,218s to 3,422s on the seat, 95% of go test's hour,
// with one of them at 1,102s (#axg2xys). A new test is parallel unless it earns a line here.
var serialTests = map[string]string{
	"TestThroughput":                                   "samples time, so nothing may compete with it",
	"TestJsxLintReleaseAndThroughput":                  "samples time, so nothing may compete with it",
	"TestProfileArtifacts":                             "saves the snapshots TestProfileSnapshotsAgree reads, so it finishes first",
	"TestExecuteFailsOnStderrOtherThanModuleDownloads": "its probe sets PATH, so it is parallel only after the probe's branch returns",
	"TestCompilerGuardBackend":                         "a subprocess entry point that returns at once unless its variable is set",
	"TestChildCPUHangGuard":                            "measures a child's CPU against a one-second budget",
	"TestChildCPUWaitGuard":                            "measures a child's CPU and wall against a one-second budget",
	"TestChildWallBackstop":                            "measures a child's wall against a 200ms backstop",
}

// TestTopLevelTestsAreParallel requires every top-level test in this package to call t.Parallel as its first
// statement, unless serialTests names it. A name listed there that no longer exists fails too, so the list
// can't keep a reason for a test that is gone.
func TestTopLevelTestsAreParallel(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !strings.HasPrefix(function.Name.Name, "Test") || function.Name.Name == "TestMain" {
				continue
			}
			seen[function.Name.Name] = true
			if _, serial := serialTests[function.Name.Name]; serial {
				continue
			}
			if !startsWithParallel(function) {
				t.Errorf("%s: %s does not call t.Parallel first; make it parallel, or name it in serialTests with its reason", path, function.Name.Name)
			}
		}
	}
	for name := range serialTests {
		if !seen[name] {
			t.Errorf("serialTests names %s, which no file in this package declares", name)
		}
	}
}

// startsWithParallel is whether function's first statement is t.Parallel(), on its first parameter.
func startsWithParallel(function *ast.FuncDecl) bool {
	if function.Body == nil || len(function.Body.List) == 0 || len(function.Type.Params.List) == 0 || len(function.Type.Params.List[0].Names) == 0 {
		return false
	}
	statement, ok := function.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := statement.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Parallel" {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	return ok && receiver.Name == function.Type.Params.List[0].Names[0].Name
}
