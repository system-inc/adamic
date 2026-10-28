package flow

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// This census uses statement() receivers, never expression Type() receivers.
func TestEveryStatementIsClassified(t *testing.T) {
	t.Parallel()
	classified := map[string]string{}
	for _, name := range []string{"SourceLine", "Debugger", "Break", "Continue", "AllocateEnvironment"} {
		classified[name] = "reads and writes nothing"
	}
	for _, name := range []string{"WriteLine", "Declare", "Assign", "Evaluate", "Panic", "SetIndex", "SetProperty", "Return", "Throw"} {
		classified[name] = "operands or stores"
	}
	for _, name := range []string{"If", "Loop", "Block", "ForOf", "Switch", "Try", "Labeled"} {
		classified[name] = "nested control flow"
	}
	found := map[string]bool{}
	files, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			method, ok := decl.(*ast.FuncDecl)
			if !ok || method.Recv == nil || method.Name.Name != "statement" {
				continue
			}
			name := method.Recv.List[0].Type.(*ast.Ident).Name
			found[name] = true
			if classified[name] == "" {
				t.Errorf("statement %s is unclassified", name)
			}
		}
	}
	for name := range classified {
		if !found[name] {
			t.Errorf("classified nonstatement %s", name)
		}
	}
	// Exhaustive statement dispatchers must explicitly recognize metadata. Reflection
	// walkers see only File and Line, neither an operand nor a nested statement.
	dispatchers := 0
	err = filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || d.Name() == "performance" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			sw, ok := node.(*ast.TypeSwitchStmt)
			if !ok {
				return true
			}
			debugger, source, exhaustive := false, false, false
			for _, clause := range sw.Body.List {
				c := clause.(*ast.CaseClause)
				for _, choice := range c.List {
					if s, ok := choice.(*ast.SelectorExpr); ok {
						debugger = debugger || s.Sel.Name == "Debugger"
						source = source || s.Sel.Name == "SourceLine"
					}
				}
				if c.List == nil {
					ast.Inspect(c, func(n ast.Node) bool {
						if call, ok := n.(*ast.CallExpr); ok {
							if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "panic" {
								exhaustive = true
							}
						}
						return true
					})
				}
			}
			if debugger {
				dispatchers++
				if !source {
					t.Errorf("%s: statement dispatcher loses SourceLine (exhaustive=%t)", path, exhaustive)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if dispatchers < 4 {
		t.Fatalf("census found only %d dispatchers", dispatchers)
	}
	typ := reflect.TypeOf(ir.SourceLine{})
	if typ.NumField() != 2 || typ.Field(0).Type.Kind() != reflect.String || typ.Field(1).Type.Kind() != reflect.Int {
		t.Fatal("SourceLine acquired executable fields")
	}
	t.Logf("%d statement kinds; %d explicit dispatchers; SourceLine reads and writes nothing", len(found), dispatchers)
}

func TestSourceLineHasNoFlowInstruction(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Main: []ir.Statement{ir.SourceLine{File: "fixture.a", Line: 3}, ir.Evaluate{Value: ir.NumberConstant{Value: 42}}}}
	graph := Build(program, -1)
	count := 0
	for _, block := range graph.Blocks {
		count += len(block.Instructions)
	}
	if count != 1 {
		t.Fatalf("SourceLine changed flow: %d instructions", count)
	}
}

// CheckDebugProgram is the external lowerer's test bridge into the ordinary
// four-check corpus. A unique fixture path keeps its marked IR out of products.
func CheckDebugProgram(t *testing.T, path string, program *ir.Program) {
	t.Helper()
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	entry := &loweredProgram{}
	entry.once.Do(func() { entry.program = program })
	loweredPrograms.Store(absolute, entry)
	t.Cleanup(func() { loweredPrograms.Delete(absolute) })
	checkAllFlowProgram(t, path)
}
