package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Date's default primitive hint is string, so String, templates and string addition
// use the same intrinsic. Structural objects with user conversions remain refused.
func (l *lowering) dateStringConversion(node *ast.Node, value ir.Expression) (ir.Expression, bool, error) {
	proven := l.checker.GetTypeAtLocation(node)
	if !l.isLibraryType(l.checker.GetNonNullableType(proven), "Date") {
		return nil, false, nil
	}
	return nil, true, l.dateForbidden(node, "Date string conversion (local toString)")
}

func (l *lowering) dateStringAddition(node *ast.Node, left, right ir.Expression) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	dateLeft := right.Type() == ir.String && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(binary.Left)), "Date")
	dateRight := left.Type() == ir.String && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(binary.Right)), "Date")
	if !dateLeft && !dateRight {
		return nil, false, nil
	}
	// Addition evaluates both operands before either ToPrimitive conversion.
	function, reads := l.stringHelper("date_add", []ir.Expression{left, right})
	var text ir.Expression
	var err error
	if dateLeft {
		text, _, err = l.dateStringConversion(binary.Left, reads[0])
		reads[0] = text
		reads[1] = l.spelled(binary.Right, reads[1])
	} else {
		text, _, err = l.dateStringConversion(binary.Right, reads[1])
		reads[1] = text
		reads[0] = l.spelled(binary.Left, reads[0])
	}
	if err != nil {
		return nil, true, err
	}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: ir.Concat{Parts: reads}}}
	return ir.Call{Function: function, Arguments: []ir.Expression{left, right}, Returns: ir.String}, true, nil
}

// A Date-or-null pointer is distinct from a tagged Date-or-null-or-undefined.
func (l *lowering) dateStringRepresentation(proven *checker.Type) (ir.Type, bool) {
	if !l.includesNull(proven) || !l.isLibraryType(l.checker.GetNonNullableType(proven), "Date") {
		return 0, false
	}
	if l.includesUndefined(proven) {
		return ir.Union, true
	}
	return ir.Object, true
}

// This is a Date method call, not an object ToPrimitive implementation.
func (l *lowering) dateStringMethod(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "toString" {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Date") || !l.libraryMember(callee) {
		return nil, false, nil
	}
	if len(call.Arguments.Nodes) != 0 {
		return nil, true, l.notYet(node, "Date.toString with arguments")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	return ir.NodeFSFile{Operation: "date_string", Arguments: []ir.Expression{value}, Of: ir.String}, true, nil
}
