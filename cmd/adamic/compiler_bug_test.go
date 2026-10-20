package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
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

// A panic deep in a real build, through the real command: the driver below supplies native's runtime archive with a
// function that panics, so `adamic build` panics inside native.Build on run's goroutine, as a compiler bug there
// would. Real programs that panic the compiler come and go as they're fixed (#my60v84), which is why this one is made.
// Mutant: delete run's defer (exit 2).
func TestCompilerPanicExitsFourThroughTheCommand(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCompilerBugDriver$", "--", "build", "testdata/build_target.a", "-o", filepath.Join(t.TempDir(), "program"))
	command.Env = append(os.Environ(), "ADAMIC_COMPILER_BUG_DRIVER=1")
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	code := 0
	if err := command.Run(); err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	if code != 4 || output.Len() != 0 || !strings.HasPrefix(diagnostic.String(), "adamic: compiler bug: "+supplierPanic+"\n") || !strings.Contains(diagnostic.String(), "goroutine ") {
		t.Fatalf("adamic build whose runtime supplier panics: exit %d, stdout %q, stderr %q; want exit 4 and the compiler bug with its stack", code, output.String(), diagnostic.String())
	}
}

const supplierPanic = "probe: the runtime supplier panicked"

// Not parallel: it replaces native's process-wide runtime supplier, in the driver's own process only.
func TestCompilerBugDriver(t *testing.T) {
	if os.Getenv("ADAMIC_COMPILER_BUG_DRIVER") != "1" {
		return
	}
	native.SupplyRuntimes(func(native.RuntimeRequest) (string, error) { panic(supplierPanic) })
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	t.Fatal("missing driver arguments")
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
