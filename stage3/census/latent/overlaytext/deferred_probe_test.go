package overlaytext

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// Read the active overlay source, rather than a pinned copy of the fixed body.
// The companion compiled guard exercises main's real predicate registry; this
// ordinary-package guard keeps its isolation boundary and overlay wiring live.
func TestLatentDeferredProbeSnapshotBoundaryHcxsh1k(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../lower.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "lower.go.txt", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	probes, snapshots, resets := 0, 0, 0
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "latentDeferredStop" {
			continue
		}
		probes++
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
				return true
			}
			id, ok := assignment.Lhs[0].(*ast.Ident)
			if ok && id.Name == "probe" {
				star, ok := assignment.Rhs[0].(*ast.StarExpr)
				if !ok {
					return true
				}
				call, ok := star.X.(*ast.CallExpr)
				if !ok || len(call.Args) != 0 {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if ok && selector.Sel.Name == "latentSnapshot" {
					if receiver, ok := selector.X.(*ast.Ident); ok && receiver.Name == "l" {
						snapshots++
					}
				}
			}
			if selector, ok := assignment.Lhs[0].(*ast.SelectorExpr); ok && selector.Sel.Name == "refusal" {
				if receiver, ok := selector.X.(*ast.Ident); ok && receiver.Name == "probe" {
					resets++
				}
			}
			return true
		})
	}
	if probes != 1 || snapshots != 1 || resets != 1 {
		t.Fatalf("deferred probe must isolate state then reset its refusal: methods=%d snapshots=%d resets=%d", probes, snapshots, resets)
	}
	script, err := os.ReadFile("../make_overlay.py")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(script), "(territory / 'deferred_probe_test.go.txt').read_text()") != 1 {
		t.Fatal("compiled predicate-registration guard is absent from the overlay")
	}
}
