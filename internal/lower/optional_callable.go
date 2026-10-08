package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A callable guard saves the selected function before its arguments run. Saving
// only its receiver is insufficient: an argument can replace the callback field.
func (l *lowering) optionalCallable(node *ast.Node) (ir.Expression, bool, error) {
	written := node.AsCallExpression()
	callee := ast.SkipParentheses(written.Expression)
	if written.QuestionDotToken == nil || callee.Flags&ast.NodeFlagsOptionalChain != 0 || callee.Kind == ast.KindElementAccessExpression || l.libraryMember(callee) {
		return nil, false, nil
	}
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(written.Expression))
	if len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 1 {
		return nil, true, l.notYet(node, "an optional call without one represented callable signature")
	}
	value, err := l.callClosure(node)
	if err != nil {
		return nil, true, err
	}
	call := value.(ir.CallClosure)
	if call.Closure.Type() != ir.Closure {
		return nil, true, l.notYet(node, "an optional call through a value without a closure representation")
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	discarded := at.Parent != nil && at.Parent.Kind == ast.KindExpressionStatement
	var of ir.Type
	if discarded {
		of = ir.Object
	} else {
		if call.Returns == 0 {
			// A void view may hide a function returning a value. Its discarded ABI
			// cannot reproduce an observation of that value without a proof.
			return nil, true, l.notYet(node, "observing an optional void call's runtime result")
		}
		if !l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
			return nil, true, l.notYet(node, "an optional call whose result is narrowed to present")
		}
		of, err = l.typeOf(node)
		if err != nil {
			return nil, true, err
		}
		if !of.IsReference() && !of.IsMaybe() {
			return nil, true, l.notYet(node, "an optional call whose scalar result lacks an undefined representation")
		}
	}
	if receiver, err := l.optionalCallableReceiver(callee); err != nil {
		return nil, true, err
	} else if receiver {
		property, known := call.Closure.(ir.Property)
		if !known || l.result.CheckedFields[property.Name] || l.result.UninitializedFields[property.Name] {
			return nil, true, l.notYet(node, "an optional method requiring a checked field contract")
		}
		call.Optional = true
		if !discarded {
			call.OptionalResult = of
		}
		return call, true, nil
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	read := ir.Read{Local: local, Of: ir.Closure}
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_callable", Type: ir.Closure, Function: function, ExpressionAssigned: true})
	selected := call.Closure
	call.Closure = read
	var present ir.Expression = call
	if discarded {
		present = ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: call}}, Result: ir.Undefined{}}
	}
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: selected}}, Result: ir.Conditional{
		Condition: ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}},
		WhenTrue:  fit(present, of), WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, true, nil
}

// CallClosure's method path resolves a receiver when calling; a plain Property
// read does not materialize a bound callable. Structural callback fields can
// hide literal methods too. Those invocations preserve the selection in the
// optional CallClosure rather than materializing an unbound property value.
func (l *lowering) optionalCallableReceiver(callee *ast.Node) (bool, error) {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false, nil
	}
	needsReceiver := func(member *ast.Symbol) bool {
		if member != nil {
			for _, root := range l.checker.GetRootSymbols(member) {
				for _, declaration := range root.Declarations {
					if declaration.Kind == ast.KindMethodDeclaration {
						return true
					}
				}
			}
		}
		return false
	}
	if needsReceiver(l.memberSymbol(callee)) {
		return true, nil
	}
	access := callee.AsPropertyAccessExpression()
	view := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false, err
	}
	found := false
	var visit ast.Visitor
	visit = func(candidate *ast.Node) bool {
		if found {
			return true
		}
		if candidate.Kind == ast.KindObjectLiteralExpression || candidate.Kind == ast.KindNewExpression {
			shape := l.checker.GetTypeAtLocation(candidate)
			if needsReceiver(l.checker.GetPropertyOfType(shape, access.Name().Text())) && l.iterationShapeFits(shape, view) {
				found = true
				return true
			}
		}
		return candidate.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found, nil
}

// Optional invocation keeps the runtime method entry and receiver together.
// Reading the same member as an ordinary value still loses that representation.
func optionalInvocationCallee(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if node.Parent == nil || node.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := node.Parent.AsCallExpression()
	return call.Expression == node && call.QuestionDotToken != nil
}
