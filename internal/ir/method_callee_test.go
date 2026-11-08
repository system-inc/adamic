package ir_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMethodCalleeSeesTheNarrowedField(t *testing.T) {
	t.Parallel()
	method := ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "run", Of: ir.Closure, Method: true}
	field := method
	field.Method = false
	narrowed := ir.Defined{Value: method, Message: "undefined where the checker narrowed it away: a call since the narrowing put it back"}
	for _, row := range []struct {
		name    string
		closure ir.Expression
		method  bool
		defined bool
	}{
		{"a call through an interface", method, true, false},
		{"the same call narrowed", narrowed, true, true},
		{"a field read as a value", field, false, false},
		{"a narrowed field read as a value", ir.Defined{Value: field, Message: narrowed.Message}, false, false},
		{"a narrowed variable", ir.Defined{Value: ir.Read{Local: 0, Of: ir.Closure}, Message: narrowed.Message}, false, false},
		{"a variable", ir.Read{Local: 0, Of: ir.Closure}, false, false},
	} {
		property, defined, ok := ir.MethodCallee(row.closure)
		if ok != row.method || (defined != nil) != row.defined {
			t.Errorf("%s: MethodCallee = %t with a Defined %t, want %t and %t", row.name, ok, defined != nil, row.method, row.defined)
			continue
		}
		if ok && property.Name != "run" {
			t.Errorf("%s: MethodCallee read the field %q, want run", row.name, property.Name)
		}
		if defined != nil && defined.Message != narrowed.Message {
			t.Errorf("%s: MethodCallee's Defined says %q, want the narrowing's own words", row.name, defined.Message)
		}
	}
}

// A narrowed call through an interface is the same call as an unnarrowed one to every analysis that asks
// about the method arm: the class methods it can reach, and what a void call releases of their result.
func TestNarrowedMethodCallKeepsTheMethodArm(t *testing.T) {
	t.Parallel()
	program := ir.Program{
		Locals: []ir.Local{{Type: ir.Object}, {Type: ir.Object}},
		Functions: []ir.Function{
			{Name: "run", Returns: ir.Object, Receiver: true, MethodName: "run", Parameters: []int{1}},
		},
		Main: []ir.Statement{ir.Evaluate{Value: ir.ObjectLiteral{Methods: []ir.Method{{Name: "run", Function: 0}}}}},
	}
	method := ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "run", Of: ir.Closure, Method: true}
	bare := ir.CallClosure{Closure: method}
	narrowed := ir.CallClosure{Closure: ir.Defined{Value: method, Message: "undefined where the checker narrowed it away: a call since the narrowing put it back"}}
	if got, want := program.MethodDiscard(narrowed), program.MethodDiscard(bare); got != want || want != ir.DiscardReference {
		t.Errorf("MethodDiscard: narrowed %d, bare %d, want both %d", got, want, ir.DiscardReference)
	}
}

// calleeFieldReaders are the places that take a CallClosure's callee apart as a field read without
// MethodCallee, each with why the narrowed call's Defined is no concern there. Any other assertion of a
// callee to a Property is a pass that would take `if (o.m) o.m()` for a plain function value and call it
// with no receiver, which is what #0bxc4jh moved the narrowed call off of.
var calleeFieldReaders = map[string]string{
	"internal/lower/expression.go:callClosure":      "lowering's own spread stop for a checked callable view, which the native emitter's checked packing needs only for a bare field: a narrowed call packs its arguments as before (emit_functions.go's callThrough)",
	"internal/native/emit_functions.go:callThrough": "chooses the checked callable view's packing for a bare field alone, as lowering's spread stop does, so a narrowed call packs its arguments as it did on the plain road",
	"internal/lower/iterable_views.go:handNextView": "a hand-driven x.next() over an Iterable view, whose callee is a library method that is never optional, so never narrowed",
}

// TestEveryCalleeFieldReadGoesThroughMethodCallee reads every pass that sees a call's callee, and holds
// each assertion of one to a Property to MethodCallee or to calleeFieldReaders, so a pass written later
// can't miss the narrowed call (Lower's condition 3 on #0bxc4jh).
func TestEveryCalleeFieldReadGoesThroughMethodCallee(t *testing.T) {
	t.Parallel()
	// Reads: internal/*/*.go
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	files := token.NewFileSet()
	scanned := 0
	for _, directory := range []string{"ir", "lower", "native", "javascript", "flow", "fresh"} {
		paths, err := filepath.Glob(filepath.Join(root, "internal", directory, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(files, path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			scanned++
			relative, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			for _, declaration := range file.Decls {
				function, isFunction := declaration.(*ast.FuncDecl)
				if !isFunction {
					continue
				}
				ast.Inspect(function, func(node ast.Node) bool {
					assertion, isAssertion := node.(*ast.TypeAssertExpr)
					if !isAssertion || !propertyType(assertion.Type) || !calleeHolder(assertion.X) {
						return true
					}
					key := filepath.ToSlash(relative) + ":" + function.Name.Name
					seen[key] = true
					if _, allowed := calleeFieldReaders[key]; !allowed {
						t.Errorf("%s: %s takes a call's callee apart as a field read without ir.MethodCallee, which a narrowed call (a Defined around the field) passes by as a plain function value; ask MethodCallee, or list it in calleeFieldReaders with why the narrowing is no concern there", files.Position(assertion.Pos()), function.Name.Name)
					}
					return true
				})
			}
		}
	}
	if scanned < 300 {
		t.Fatalf("read %d files under internal: the scan no longer finds the passes", scanned)
	}
	for reader := range calleeFieldReaders {
		if !seen[reader] {
			t.Errorf("stale calleeFieldReaders entry %s: it no longer asserts a callee to a Property; remove it", reader)
		}
	}
}

// propertyType is whether a type assertion's type is the IR's Property, as ir.Property outside the
// package or Property inside it.
func propertyType(expression ast.Expr) bool {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name == "Property"
	case *ast.SelectorExpr:
		identifier, isIdentifier := expression.X.(*ast.Ident)
		return isIdentifier && identifier.Name == "ir" && expression.Sel.Name == "Property"
	}
	return false
}

// calleeHolder is whether an expression holds a call's callee by the names every pass gives it: a
// CallClosure's Closure field, or callClosure's and checkedFunctionTypeCall's closure.
func calleeHolder(expression ast.Expr) bool {
	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		return expression.Sel.Name == "Closure"
	case *ast.Ident:
		return expression.Name == "closure"
	}
	return false
}
