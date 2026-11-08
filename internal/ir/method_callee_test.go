package ir_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMethodCalleeSeesTheCheckedField(t *testing.T) {
	t.Parallel()
	method := ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "run", Of: ir.Closure, Method: true}
	field := method
	field.Method = false
	narrowed := ir.Defined{Value: method, Message: "undefined where the checker narrowed it away: a call since the narrowing put it back"}
	failure := ir.StringConstant{Index: 1}
	asserted := ir.Coalesce{Value: method, Panic: failure, Of: ir.Closure}
	for _, row := range []struct {
		name    string
		closure ir.Expression
		method  bool
		check   ir.Expression
	}{
		{"a call through an interface", method, true, nil},
		{"the same call narrowed", narrowed, true, narrowed},
		{"the same call asserted with !", asserted, true, asserted},
		{"a field read as a value", field, false, nil},
		{"a narrowed field read as a value", ir.Defined{Value: field, Message: narrowed.Message}, false, nil},
		{"an asserted field read as a value", ir.Coalesce{Value: field, Panic: failure, Of: ir.Closure}, false, nil},
		{"a field ?? a fallback", ir.Coalesce{Value: method, Fallback: ir.Read{Local: 1, Of: ir.Closure}, Of: ir.Closure}, false, nil},
		{"a narrowed variable", ir.Defined{Value: ir.Read{Local: 0, Of: ir.Closure}, Message: narrowed.Message}, false, nil},
		{"an asserted variable", ir.Coalesce{Value: ir.Read{Local: 0, Of: ir.Closure}, Panic: failure, Of: ir.Closure}, false, nil},
		{"a variable", ir.Read{Local: 0, Of: ir.Closure}, false, nil},
	} {
		property, check, ok := ir.MethodCallee(row.closure)
		if ok != row.method || !reflect.DeepEqual(check, row.check) {
			t.Errorf("%s: MethodCallee = %t with the check %#v, want %t and %#v", row.name, ok, check, row.method, row.check)
			continue
		}
		if ok && property.Name != "run" {
			t.Errorf("%s: MethodCallee read the field %q, want run", row.name, property.Name)
		}
		if ok && !reflect.DeepEqual(ir.MethodCalleeWith(check, property), row.closure) {
			t.Errorf("%s: MethodCalleeWith doesn't give back the callee MethodCallee took apart", row.name)
		}
	}
	// A pass that reads the method off a receiver of its own keeps the call's check around its read.
	held := method
	held.Object = ir.Read{Local: 2, Of: ir.Object}
	if got := ir.MethodCalleeWith(asserted, held); !reflect.DeepEqual(got, ir.Coalesce{Value: held, Panic: failure, Of: ir.Closure}) {
		t.Errorf("MethodCalleeWith(asserted) = %#v, want the assertion around the new read", got)
	}
}

// A narrowed or asserted call through an interface is the same call as a bare one to every analysis that
// asks about the method arm: the class methods it can reach, and what a void call releases of their result.
func TestCheckedMethodCallKeepsTheMethodArm(t *testing.T) {
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
	if want := program.MethodDiscard(bare); want != ir.DiscardReference {
		t.Fatalf("MethodDiscard of the bare call %d, want %d", want, ir.DiscardReference)
	}
	for name, closure := range map[string]ir.Expression{
		"narrowed": ir.Defined{Value: method, Message: "undefined where the checker narrowed it away: a call since the narrowing put it back"},
		"asserted": ir.Coalesce{Value: method, Panic: ir.StringConstant{Index: 1}, Of: ir.Closure},
	} {
		if got := program.MethodDiscard(ir.CallClosure{Closure: closure}); got != ir.DiscardReference {
			t.Errorf("MethodDiscard: %s %d, want the bare call's %d", name, got, ir.DiscardReference)
		}
	}
}

// calleeFieldReaders are the places that take a CallClosure's callee apart as a field read without
// MethodCallee, each with why a checked call's Defined or Coalesce is no concern there. Any other
// assertion of a callee to a Property is a pass that would take `if (o.m) o.m()` or `o.m!()` for a plain
// function value and call it with no receiver, which is what #0bxc4jh moved those calls off of.
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
						t.Errorf("%s: %s takes a call's callee apart as a field read without ir.MethodCallee, which a checked call (a narrowing's Defined or a non-null assertion's Coalesce around the field) passes by as a plain function value; ask MethodCallee, or list it in calleeFieldReaders with why the check is no concern there", files.Position(assertion.Pos()), function.Name.Name)
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
