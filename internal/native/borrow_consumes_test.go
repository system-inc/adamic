package native

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Every operation on pure's list (borrow.go, pureKind) is a decision for consumes: a consumer is done
// with its operands when it's evaluated, so it may lend them; a pass-through hands an operand on as
// its own value, so it must not, or its parent keeps a pointer nobody counted. ir.Defined was made
// pure and nobody made that decision, and a narrowed read was lent to a call whose next argument freed
// it (borrow_defined_lent.a). Reading pureKind's list from the source makes the next operation added
// there fail here until it's put in one column or the other.
func TestPassThroughsAreNotConsumers(t *testing.T) {
	passesThrough := map[string]ir.Expression{
		"Read": ir.Read{}, "Conditional": ir.Conditional{}, "Coalesce": ir.Coalesce{}, "Box": ir.Box{},
		"Narrow": ir.Narrow{}, "Unwrap": ir.Unwrap{}, "CheckedCast": ir.CheckedCast{}, "MaybeOf": ir.MaybeOf{},
		"Defined": ir.Defined{}, "Undefined": ir.Undefined{},
		"NumberConstant": ir.NumberConstant{}, "BooleanConstant": ir.BooleanConstant{}, "StringConstant": ir.StringConstant{},
	}
	consumers := map[string]ir.Expression{
		"Unary": ir.Unary{}, "Binary": ir.Binary{}, "NumberToString": ir.NumberToString{}, "BooleanToString": ir.BooleanToString{},
		"Concat": ir.Concat{}, "Length": ir.Length{}, "StringLength": ir.StringLength{}, "CharCodeAt": ir.CharCodeAt{},
		"StringIndex": ir.StringIndex{}, "ArrayIndex": ir.ArrayIndex{}, "Property": ir.Property{}, "MapGet": ir.MapGet{},
		"MapHas": ir.MapHas{}, "MapSize": ir.MapSize{}, "HasOwn": ir.HasOwn{}, "IsUndefined": ir.IsUndefined{},
		"TypeOf": ir.TypeOf{}, "MathCall": ir.MathCall{}, "NumberCall": ir.NumberCall{}, "ToFixed": ir.ToFixed{},
		"NumberFormat": ir.NumberFormat{}, "Trim": ir.Trim{}, "StringCall": ir.StringCall{}, "CodePoints": ir.CodePoints{},
		"ArraySearch": ir.ArraySearch{}, "UnionToString": ir.UnionToString{}, "MaybeToString": ir.MaybeToString{},
		"InstanceOf": ir.InstanceOf{},
	}
	for name, expression := range passesThrough {
		if consumes(expression) {
			t.Errorf("ir.%s hands an operand on as its own value, but consumes says it may lend it", name)
		}
	}
	for name, expression := range consumers {
		if !consumes(expression) {
			t.Errorf("ir.%s is listed as a consumer, but consumes says it isn't", name)
		}
	}
	for _, name := range pureKinds(t) {
		_, through := passesThrough[name]
		_, consumer := consumers[name]
		if !through && !consumer {
			t.Errorf("ir.%s is on pureKind's list but not decided here: does it hand an operand on as its own value (add it to consumes' exclusions and passesThrough) or is it done with them (consumers)?", name)
		}
	}
}

// pureKinds is every type named in pureKind's switch, read from borrow.go.
func pureKinds(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "borrow.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if !ok || function.Name.Name != "pureKind" {
			return true
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if clause, ok := node.(*ast.CaseClause); ok {
				for _, each := range clause.List {
					if selector, ok := each.(*ast.SelectorExpr); ok {
						names = append(names, selector.Sel.Name)
					}
				}
			}
			return true
		})
		return false
	})
	if len(names) < 20 || !slices.Contains(names, "Defined") {
		t.Fatalf("read %d names from pureKind's switch (%v); the test's reading of borrow.go is out of date", len(names), names)
	}
	return names
}
