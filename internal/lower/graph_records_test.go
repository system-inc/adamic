package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A record on a graph cycle is stopped by one of two rules, since neither form joins a graph region:
// graphTypes stops a selected component that holds a record type, which catches ir.RecordLiteral (how
// a record literal lowers today), and graphAdoptionBoundary refuses ir.ObjectLiteral{Record: true} (the
// older form). #a9e21hc was a record form the boundary's rule didn't see. So every IR expression whose
// Type method answers ir.Record outright, a record maker, is named here with the rule that covers it,
// and a new one is red until someone decides which rule holds it.
func TestGraphRecordFormsMeetARule(t *testing.T) {
	t.Parallel()
	covered := map[string]string{
		"RecordLiteral": "graphTypes' record-in-component stop (#a9e21hc)",
	}
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var makers []string
	expressions := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Name.Name != "Type" || function.Body == nil {
				continue
			}
			receiver, ok := function.Recv.List[0].Type.(*ast.Ident)
			if !ok {
				continue
			}
			expressions++
			// Any return of the bare identifier Record makes it a record maker.
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if returned, ok := node.(*ast.ReturnStmt); ok && len(returned.Results) == 1 {
					if identifier, ok := returned.Results[0].(*ast.Ident); ok && identifier.Name == "Record" && !slices.Contains(makers, receiver.Name) {
						makers = append(makers, receiver.Name)
					}
				}
				return true
			})
		}
	}
	if expressions < 100 {
		t.Fatalf("found %d Type methods in ../ir: the enumeration no longer finds the IR", expressions)
	}
	for _, maker := range makers {
		if covered[maker] == "" {
			t.Errorf("ir.%s answers Record: decide which rule stops it on a graph cycle (graphTypes' record-in-component stop, or graphAdoptionBoundary's refusal), then name it here", maker)
		}
	}
	for name := range covered {
		if !slices.Contains(makers, name) {
			t.Errorf("ir.%s is named here but no longer answers Record", name)
		}
	}
	// The older form is a flag on ObjectLiteral, not a type of its own: its arm in graphAdoptionBoundary
	// stays, read from the source so the arm can't go quietly.
	source, err := parser.ParseFile(files, "graph_types.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	flagged := false
	ast.Inspect(source, func(node ast.Node) bool {
		if function, ok := node.(*ast.FuncDecl); ok && function.Name.Name == "graphAdoptionBoundary" {
			ast.Inspect(function.Body, func(inner ast.Node) bool {
				if selector, ok := inner.(*ast.SelectorExpr); ok && selector.Sel.Name == "Record" {
					if identifier, ok := selector.X.(*ast.Ident); ok && identifier.Name == "value" {
						flagged = true
					}
				}
				return true
			})
		}
		return true
	})
	if !flagged {
		t.Error("graphAdoptionBoundary no longer reads ObjectLiteral.Record: the older record form has no rule")
	}
}
