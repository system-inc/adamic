package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only immediate reads can use this private intrinsic carrier. Neither constructor
// nor callable identity escapes, and no general any contract is admitted. The
// scanner consumes one numeric argument; the library already implements that call.
func (l *lowering) scannerStringConstructorCast(node *ast.Node) bool {
	if node.Kind != ast.KindAsExpression || l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsAny == 0 || !l.isLibraryGlobal(ast.SkipParentheses(node.AsAsExpression().Expression), "String") {
		return false
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	property := outer.Parent
	if property == nil || property.Kind != ast.KindPropertyAccessExpression || property.AsPropertyAccessExpression().Expression != outer || property.Name().Text() != "fromCodePoint" || property.AsPropertyAccessExpression().QuestionDotToken != nil {
		return false
	}
	outer = property
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindConditionalExpression && parent.AsConditionalExpression().Condition == outer {
		return true
	}
	if parent.Kind != ast.KindCallExpression || parent.AsCallExpression().Expression != outer || parent.AsCallExpression().QuestionDotToken != nil || len(parent.AsCallExpression().Arguments.Nodes) != 1 {
		return false
	}
	argument := parent.AsCallExpression().Arguments.Nodes[0]
	return argument.Kind != ast.KindSpreadElement && l.checker.GetTypeAtLocation(argument).Flags()&checker.TypeFlagsNumberLike != 0
}

func (l *lowering) scannerStringConstructorRead(node *ast.Node) (ir.Expression, bool, error) {
	property := node
	if node.Kind == ast.KindCallExpression {
		property = ast.SkipParentheses(node.AsCallExpression().Expression)
	}
	if property.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	cast := ast.SkipParentheses(property.AsPropertyAccessExpression().Expression)
	if !l.scannerStringConstructorCast(cast) {
		return nil, false, nil
	}
	// A private carrier is sound only when the closed program cannot replace the
	// intrinsic. Reject all other value uses of String, including aliases and writes.
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return nil, true, err
	}
	var invalid *ast.Node
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if ast.IsIdentifier(part) && l.isLibraryGlobal(part, "String") {
			outer := part
			for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
				outer = outer.Parent
			}
			parent := outer.Parent
			safe := parent != nil && parent.Kind == ast.KindAsExpression && l.scannerStringConstructorCast(parent)
			if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == outer {
				access := parent
				for access.Parent != nil && access.Parent.Kind == ast.KindParenthesizedExpression {
					access = access.Parent
				}
				safe = access.Parent != nil && access.Parent.Kind == ast.KindCallExpression && access.Parent.AsCallExpression().Expression == access
			}
			if parent != nil && parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == outer {
				safe = true
			}
			if !safe && !ast.IsPartOfTypeNode(part) {
				invalid = part
			}
		}
		part.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if invalid != nil {
		return nil, true, l.notYet(invalid, "a scanner String constructor view alongside an escaping or mutable intrinsic")
	}
	function := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "codePoint", Type: ir.Number, Function: function})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "scanner_fromCodePoint", Closure: true, Parameters: []int{local}, Returns: ir.String, Body: []ir.Statement{ir.Return{Value: ir.StringFromCodes{CodePoints: true, Codes: []ir.Expression{ir.Read{Local: local, Of: ir.Number}}}}}})
	parameter := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
	result := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.String, Name: "string"})
	contract := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewCallable, Of: ir.Closure, Name: "(codePoint: number) => string", Parameters: []ir.ViewContractID{parameter}, Result: result, ProducerCertified: true, Functions: []int{function}})
	carrier := ir.ObjectLiteral{Fields: []ir.Field{{Name: "fromCodePoint", Value: ir.MakeClosure{Function: function}}}}
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	l.result.CheckedFields["fromCodePoint"] = true
	l.result.ViewOrigins = append(l.result.ViewOrigins, carrier)
	read := ir.Property{Object: carrier, Name: "fromCodePoint", Of: ir.Closure, View: sourceExpression(property), ViewContract: contract}
	if node.Kind != ast.KindCallExpression {
		return read, true, nil
	}
	argument, err := l.expression(node.AsCallExpression().Arguments.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	return ir.CallClosure{Closure: read, Arguments: []ir.Expression{argument}, Returns: ir.String}, true, nil
}

// The direct expression body has a result established by the intrinsic adapter,
// even though TypeScript infers any through the legacy assertion. No other any
// return, block body or declared signature is inferred from a caller's wishes.
func (l *lowering) scannerStringArrowResult(declaration *ast.Node) bool {
	if declaration.Kind != ast.KindArrowFunction || declaration.Body() == nil {
		return false
	}
	body := ast.SkipParentheses(declaration.Body())
	if body.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(body.AsCallExpression().Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && l.scannerStringConstructorCast(ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression))
}
